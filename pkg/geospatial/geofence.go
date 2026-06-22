package geospatial

// PointInPolygon проверяет, лежит ли точка внутри полигона (алгоритм луча).
// polygon — замкнутый список вершин (последняя точка соединяется с первой).
func PointInPolygon(pt Point, polygon []Point) bool {
	n := len(polygon)
	if n < 3 {
		return false
	}
	inside := false
	j := n - 1
	for i := 0; i < n; i++ {
		if (polygon[i].Lng > pt.Lng) != (polygon[j].Lng > pt.Lng) &&
			pt.Lat < (polygon[j].Lat-polygon[i].Lat)*(pt.Lng-polygon[i].Lng)/(polygon[j].Lng-polygon[i].Lng)+polygon[i].Lat {
			inside = !inside
		}
		j = i
	}
	return inside
}
