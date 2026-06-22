package geospatial

import "math"

// Point — координаты в десятичных градусах (широта, долгота).
type Point struct {
	Lat, Lng float64
}

// HaversineDistance вычисляет расстояние в метрах между двумя точками.
// Используется сферическая модель Земли (радиус 6371000 м).
func HaversineDistance(a, b Point) float64 {
	const R = 6371000
	lat1 := toRad(a.Lat)
	lng1 := toRad(a.Lng)
	lat2 := toRad(b.Lat)
	lng2 := toRad(b.Lng)

	dLat := lat2 - lat1
	dLng := lng2 - lng1
	sinDLat := math.Sin(dLat / 2)
	sinDLng := math.Sin(dLng / 2)
	aH := sinDLat*sinDLat + math.Cos(lat1)*math.Cos(lat2)*sinDLng*sinDLng
	c := 2 * math.Atan2(math.Sqrt(aH), math.Sqrt(1-aH))
	return R * c
}

// toRad переводит градусы в радианы.
func toRad(deg float64) float64 {
	return deg * math.Pi / 180
}
