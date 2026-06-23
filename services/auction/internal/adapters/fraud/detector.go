package fraud

import (
	"sync"

	"rtb-platform/services/auction/internal/ports"
)

// inmemDetector — простейший детектор мошенничества на основе чёрного списка IP и deviceID в памяти.
type inmemDetector struct {
	mu        sync.RWMutex
	blacklist map[string]bool
}

// NewInmemDetector создаёт детектор с начальным списком подозрительных ключей.
func NewInmemDetector(initial []string) ports.FraudDetector {
	d := &inmemDetector{blacklist: make(map[string]bool, len(initial))}
	for _, key := range initial {
		d.blacklist[key] = true
	}
	return d
}

func (d *inmemDetector) IsSuspicious(ip, deviceID string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.blacklist[ip] || d.blacklist[deviceID]
}
