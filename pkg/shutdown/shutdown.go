package shutdown

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Closer — функция, которую нужно вызвать для корректного завершения компонента.
type Closer func(ctx context.Context) error

// Manager управляет последовательным завершением компонентов.
type Manager struct {
	mu      sync.Mutex
	closers []namedCloser
	logger  Logger
	timeout time.Duration // общий таймаут на все компоненты
}

// Logger — минимальный интерфейс для логирования.
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

type namedCloser struct {
	name     string
	priority int
	fn       Closer
	timeout  time.Duration // индивидуальный таймаут
}

// defaultLogger использует стандартный log, если не задан другой.
type defaultLogger struct{}

func (l defaultLogger) Info(msg string, keysAndValues ...interface{}) {
	log.Printf("[INFO] "+msg, keysAndValues...)
}
func (l defaultLogger) Error(msg string, keysAndValues ...interface{}) {
	log.Printf("[ERROR] "+msg, keysAndValues...)
}

// NewManager создаёт менеджер с общим таймаутом totalTimeout.
// Если totalTimeout == 0, используется 30 секунд.
func NewManager(totalTimeout time.Duration) *Manager {
	if totalTimeout <= 0 {
		totalTimeout = 30 * time.Second
	}
	return &Manager{
		timeout: totalTimeout,
		logger:  defaultLogger{},
	}
}

// SetLogger задаёт пользовательский логгер.
func (m *Manager) SetLogger(l Logger) {
	m.logger = l
}

// Add регистрирует функцию закрытия с указанным именем, приоритетом (меньше — раньше) и индивидуальным таймаутом.
// Если timeout == 0, используется общий таймаут менеджера.
func (m *Manager) Add(name string, priority int, fn Closer, timeout time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closers = append(m.closers, namedCloser{
		name:     name,
		priority: priority,
		fn:       fn,
		timeout:  timeout,
	})
}

// Shutdown запускает последовательное завершение всех зарегистрированных компонентов.
// Сначала сортирует по приоритету (меньше — раньше). Каждый компонент завершается в отдельном контексте с таймаутом.
// Возвращает ошибку, если какой-то компонент не успел завершиться за отведённое время.
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	closers := make([]namedCloser, len(m.closers)) // исправлено m.mclosers → m.closers
	copy(closers, m.closers)
	m.mu.Unlock()

	sortByPriority(closers)

	var errs []error
	for _, c := range closers {
		timeout := c.timeout
		if timeout <= 0 {
			timeout = m.timeout
		}
		ctx, cancel := context.WithTimeout(ctx, timeout)
		m.logger.Info("shutting down", "name", c.name, "timeout", timeout)
		err := c.fn(ctx)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", c.name, err))
			m.logger.Error("shutdown error", "name", c.name, "error", err)
		} else {
			m.logger.Info("shutdown complete", "name", c.name)
		}
		cancel()
	}
	if len(errs) > 0 {
		return fmt.Errorf("shutdown errors: %v", errs)
	}
	return nil
}

// Wait блокируется до получения SIGINT или SIGTERM, затем вызывает Shutdown.
// Общий таймаут для всего процесса завершения равен m.timeout.
func (m *Manager) Wait() {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	m.logger.Info("received signal, starting shutdown", "signal", sig.String())

	ctx, cancel := context.WithTimeout(context.Background(), m.timeout)
	defer cancel()

	if err := m.Shutdown(ctx); err != nil {
		m.logger.Error("shutdown failed", "error", err)
		os.Exit(1)
	}
	m.logger.Info("shutdown completed successfully")
}

// sortByPriority сортирует слайс по возрастанию priority.
func sortByPriority(closers []namedCloser) {
	// Сортировка вставками, так как N обычно < 10
	for i := 1; i < len(closers); i++ {
		j := i
		for j > 0 && closers[j].priority < closers[j-1].priority {
			closers[j], closers[j-1] = closers[j-1], closers[j]
			j--
		}
	}
}
