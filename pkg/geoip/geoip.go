package geoip

import (
	"errors"
	"net"
	"sync"

	"github.com/oschwald/geoip2-golang"
)

// GeoDB представляет загруженную в память базу GeoIP.
type GeoDB struct {
	mu sync.RWMutex
	db *geoip2.Reader
}

// Result содержит информацию о геолокации IP-адреса.
type Result struct {
	Country string
	City    string
	Lat     float64
	Lng     float64
}

// New открывает базу GeoLite2 City по указанному пути.
func New(path string) (*GeoDB, error) {
	db, err := geoip2.Open(path)
	if err != nil {
		return nil, err
	}
	return &GeoDB{db: db}, nil
}

// Lookup выполняет поиск геоданных по строковому представлению IP.
// Возвращает ошибку, если IP некорректен.
func (g *GeoDB) Lookup(ipStr string) (Result, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return Result{}, errors.New("invalid IP address")
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	record, err := g.db.City(ip)
	if err != nil {
		return Result{}, err
	}
	return Result{
		Country: record.Country.Names["en"],
		City:    record.City.Names["en"],
		Lat:     record.Location.Latitude,
		Lng:     record.Location.Longitude,
	}, nil
}

// Close освобождает ресурсы базы.
func (g *GeoDB) Close() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.db != nil {
		return g.db.Close()
	}
	return nil
}
