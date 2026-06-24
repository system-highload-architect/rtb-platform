package handler

import (
	"encoding/json"
	"net/http"

	analyticsv1 "rtb-platform/pb/analytics/v1"

	"rtb-platform/services/gateway/internal/ports"
)

type AnalyticsHandler struct {
	analyticsPort ports.AnalyticsPort
}

func NewAnalyticsHandler(analyticsPort ports.AnalyticsPort) *AnalyticsHandler {
	return &AnalyticsHandler{analyticsPort: analyticsPort}
}

func (h *AnalyticsHandler) Report(w http.ResponseWriter, r *http.Request) {
	req := &analyticsv1.ReportRequest{
		StartDate:  r.URL.Query().Get("start_date"),
		EndDate:    r.URL.Query().Get("end_date"),
		Dimensions: []string{"campaign_id", "device_type"},
		Metrics:    []string{"impressions", "clicks", "spend"},
	}
	rows, err := h.analyticsPort.GetReport(r.Context(), req)
	if err != nil {
		http.Error(w, `{"error":"failed to get report"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rows)
}

func (h *AnalyticsHandler) Forecast(w http.ResponseWriter, r *http.Request) {
	// Простой пример: принимаем историю и горизонт из query (в реальности брали бы из тела)
	var hist []float64
	json.Unmarshal([]byte(r.URL.Query().Get("history")), &hist)
	horizon := 7
	resp, err := h.analyticsPort.Forecast(r.Context(), &analyticsv1.ForecastRequest{
		History: hist,
		Horizon: int32(horizon),
	})
	if err != nil {
		http.Error(w, `{"error":"forecast failed"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AnalyticsHandler) FactorAnalysis(w http.ResponseWriter, r *http.Request) {
	// Пример: получаем список профилей пользователей из тела (или заглушка)
	var req analyticsv1.FactorRequest
	// ... заполнение из тела запроса
	resp, err := h.analyticsPort.FactorAnalysis(r.Context(), &req)
	if err != nil {
		http.Error(w, `{"error":"factor analysis failed"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
