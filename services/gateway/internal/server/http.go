package server

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/goccy/go-json"

	analyticsv1 "rtb-platform/pb/analytics/v1"

	"rtb-platform/pkg/zerocopy"

	"rtb-platform/services/gateway/internal/domain"
	"rtb-platform/services/gateway/internal/handler"
	"rtb-platform/services/gateway/internal/middleware"
	"rtb-platform/services/gateway/internal/ports"
)

// HTTPServer оборачивает стандартный http.Server с middleware и зависимостями.
type HTTPServer struct {
	server           *http.Server
	logger           *slog.Logger
	jsonRPCService   *domain.JSONRPCService
	analyticsPort    ports.AnalyticsPort       // для экспорта Excel
	analyticsREST    *handler.AnalyticsHandler // для REST-аналитики
	authMiddleware   *middleware.AuthMiddleware
	appsecMiddleware *middleware.AppsecMiddleware
	rateLimitMW      *middleware.RateLimitMiddleware
	idempotentMW     *middleware.IdempotentMiddleware
	port             int
	readTimeout      time.Duration
	writeTimeout     time.Duration
	idleTimeout      time.Duration
	mux              *http.ServeMux
}

// Option — функциональная опция для конструирования HTTPServer.
type Option func(*HTTPServer)

// WithPort задаёт порт прослушивания.
func WithPort(port int) Option {
	return func(s *HTTPServer) { s.port = port }
}

// WithLogger задаёт структурированный логгер.
func WithLogger(l *slog.Logger) Option {
	return func(s *HTTPServer) { s.logger = l }
}

// WithJSONRPCService задаёт доменный сервис для обработки JSON-RPC методов.
func WithJSONRPCService(svc *domain.JSONRPCService) Option {
	return func(s *HTTPServer) { s.jsonRPCService = svc }
}

// WithAnalyticsHandler задаёт порт аналитики для экспорта Excel и отчётов.
func WithAnalyticsHandler(p ports.AnalyticsPort) Option {
	return func(s *HTTPServer) { s.analyticsPort = p }
}

// WithAnalyticsRESTHandler задаёт обработчик REST-аналитики.
func WithAnalyticsRESTHandler(ah *handler.AnalyticsHandler) Option {
	return func(s *HTTPServer) { s.analyticsREST = ah }
}

// WithAuthMiddleware подключает middleware аутентификации.
func WithAuthMiddleware(am *middleware.AuthMiddleware) Option {
	return func(s *HTTPServer) { s.authMiddleware = am }
}

// WithAppsecMiddleware подключает middleware безопасности.
func WithAppsecMiddleware(am *middleware.AppsecMiddleware) Option {
	return func(s *HTTPServer) { s.appsecMiddleware = am }
}

// WithRateLimitMiddleware подключает middleware ограничения частоты.
func WithRateLimitMiddleware(rlm *middleware.RateLimitMiddleware) Option {
	return func(s *HTTPServer) { s.rateLimitMW = rlm }
}

// WithIdempotentMiddleware подключает middleware идемпотентности.
func WithIdempotentMiddleware(im *middleware.IdempotentMiddleware) Option {
	return func(s *HTTPServer) { s.idempotentMW = im }
}

// WithReadTimeout задаёт таймаут чтения запроса.
func WithReadTimeout(d time.Duration) Option {
	return func(s *HTTPServer) { s.readTimeout = d }
}

// WithWriteTimeout задаёт таймаут записи ответа.
func WithWriteTimeout(d time.Duration) Option {
	return func(s *HTTPServer) { s.writeTimeout = d }
}

// WithIdleTimeout задаёт таймаут бездействия соединения.
func WithIdleTimeout(d time.Duration) Option {
	return func(s *HTTPServer) { s.idleTimeout = d }
}

// NewHTTPServer создаёт экземпляр сервера с применёнными опциями.
func NewHTTPServer(opts ...Option) *HTTPServer {
	s := &HTTPServer{}
	for _, opt := range opts {
		opt(s)
	}

	mux := http.NewServeMux()
	s.mux = mux

	// 1. Регистрация маршрутов
	// JSON-RPC (публичный)
	mux.HandleFunc("/rpc", s.handleRPC)

	// Аналитические маршруты с аутентификацией (если заданы)
	if s.analyticsREST != nil && s.authMiddleware != nil {
		auth := s.authMiddleware.Handler
		mux.Handle("/api/report", auth(http.HandlerFunc(s.analyticsREST.Report)))
		mux.Handle("/api/forecast", auth(http.HandlerFunc(s.analyticsREST.Forecast)))
		mux.Handle("/api/factor-analysis", auth(http.HandlerFunc(s.analyticsREST.FactorAnalysis)))
	}

	// Экспорт Excel (публичный или может быть также защищён)
	mux.HandleFunc("/export/report", s.handleExportReport)

	// 2. Построение цепочки middleware (порядок важен)
	var handler http.Handler = mux
	handler = corsMiddleware(handler) // CORS всегда снаружи
	if s.appsecMiddleware != nil {
		handler = s.appsecMiddleware.Handler(handler)
	}
	if s.rateLimitMW != nil {
		handler = s.rateLimitMW.Handler(handler)
	}
	if s.idempotentMW != nil {
		handler = s.idempotentMW.Handler(handler)
	}

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      handler,
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
		IdleTimeout:  s.idleTimeout,
	}
	return s
}

// corsMiddleware простая реализация CORS.
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Idempotency-Key")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// Handle регистрирует дополнительный обработчик (например, /metrics).
func (s *HTTPServer) Handle(pattern string, handler http.Handler) {
	s.mux.Handle(pattern, handler)
}

// ListenAndServe запускает сервер.
func (s *HTTPServer) ListenAndServe() error {
	return s.server.ListenAndServe()
}

// Shutdown корректно останавливает сервер.
func (s *HTTPServer) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// ─── Обработчики ────────────────────────────────────────────

func (s *HTTPServer) handleRPC(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	s.logger.Debug("rpc request", "remote_addr", r.RemoteAddr)

	// 1. Читаем тело в буфер из пула
	bufPtr := zerocopy.GetBytes()
	buf := *bufPtr
	defer zerocopy.PutBytes(bufPtr)

	for {
		if len(buf) == cap(buf) {
			newBuf := make([]byte, len(buf), 2*cap(buf))
			copy(newBuf, buf)
			buf = newBuf
		}
		n, err := r.Body.Read(buf[len(buf):cap(buf)])
		buf = buf[:len(buf)+n]
		if err == io.EOF {
			break
		}
		if err != nil {
			http.Error(w, "cannot read body", http.StatusBadRequest)
			return
		}
	}
	*bufPtr = buf

	// 2. Zero‑copy парсинг полей JSON
	methodBytes, ok := zerocopy.GetJSONField(buf, "method")
	if !ok {
		s.writeJSONRPCErrorZeroCopy(w, -32600, "Invalid Request: missing method", nil)
		return
	}
	method := zerocopy.BytesToString(methodBytes[1 : len(methodBytes)-1]) // убираем кавычки

	paramsBytes, _ := zerocopy.GetJSONField(buf, "params")
	var params json.RawMessage
	if paramsBytes != nil {
		params = json.RawMessage(paramsBytes)
	}

	idBytes, ok := zerocopy.GetJSONField(buf, "id")
	var id interface{}
	if ok {
		id = json.RawMessage(idBytes)
	}

	// 3. Вызов доменного диспетчера
	result, err := s.jsonRPCService.Dispatch(r.Context(), method, params)
	if err != nil {
		s.logger.Error("dispatch error", "method", method, "error", err)
		s.writeJSONRPCErrorZeroCopy(w, -32603, "Internal error", id)
		return
	}

	// 4. Сборка ответа
	respBufPtr := zerocopy.GetBytes()
	respBuf := *respBufPtr
	defer zerocopy.PutBytes(respBufPtr)

	respBuf = append(respBuf, `{"jsonrpc":"2.0","result":`...)
	resultJSON, err := json.Marshal(result)
	if err != nil {
		s.logger.Error("marshal result error", "error", err)
		s.writeJSONRPCErrorZeroCopy(w, -32603, "Internal error", id)
		return
	}
	respBuf = append(respBuf, resultJSON...)
	respBuf = append(respBuf, `,"id":`...)
	if id != nil {
		idJSON, _ := json.Marshal(id)
		respBuf = append(respBuf, idJSON...)
	} else {
		respBuf = append(respBuf, "null"...)
	}
	respBuf = append(respBuf, '}')

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(respBuf)
	*respBufPtr = respBuf[:0]
}

func (s *HTTPServer) handleExportReport(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	req := &analyticsv1.ReportRequest{
		StartDate:  r.URL.Query().Get("start_date"),
		EndDate:    r.URL.Query().Get("end_date"),
		Dimensions: []string{"campaign_id"},
		Metrics:    []string{"impressions", "clicks"},
	}
	data, err := s.analyticsPort.ExportExcel(r.Context(), req)
	if err != nil {
		s.logger.Error("export excel error", "error", err)
		http.Error(w, "failed to generate report", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	w.Header().Set("Content-Disposition", "attachment; filename=report.xlsx")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

// writeJSONRPCErrorZeroCopy формирует JSON-RPC ошибку без лишних аллокаций.
func (s *HTTPServer) writeJSONRPCErrorZeroCopy(w http.ResponseWriter, code int, message string, id interface{}) {
	bufPtr := zerocopy.GetBytes()
	buf := *bufPtr
	defer zerocopy.PutBytes(bufPtr)

	buf = append(buf, `{"jsonrpc":"2.0","error":{"code":`...)
	buf = zerocopy.AppendJSONInt(buf, int64(code))
	buf = append(buf, `,"message":`...)
	buf = zerocopy.AppendJSONString(buf, message)
	buf = append(buf, `},"id":`...)
	if id != nil {
		idJSON, _ := json.Marshal(id)
		buf = append(buf, idJSON...)
	} else {
		buf = append(buf, "null"...)
	}
	buf = append(buf, '}')

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(buf)
}
