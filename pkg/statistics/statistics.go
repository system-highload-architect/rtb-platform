package statistics

import (
	"math"
	"sort"
)

// Mean возвращает среднее арифметическое.
func Mean(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	return sum / float64(len(values))
}

// Variance вычисляет несмещённую выборочную дисперсию (деление на n-1).
func Variance(values []float64) float64 {
	if len(values) < 2 {
		return 0
	}
	m := Mean(values)
	sumSq := 0.0
	for _, v := range values {
		diff := v - m
		sumSq += diff * diff
	}
	return sumSq / float64(len(values)-1)
}

// StdDev — квадратный корень из дисперсии.
func StdDev(values []float64) float64 {
	return math.Sqrt(Variance(values))
}

// Covariance вычисляет ковариацию между двумя выборками (несмещённая, n-1).
func Covariance(x, y []float64) float64 {
	if len(x) != len(y) || len(x) < 2 {
		return 0
	}
	mx, my := Mean(x), Mean(y)
	sum := 0.0
	for i := range x {
		sum += (x[i] - mx) * (y[i] - my)
	}
	return sum / float64(len(x)-1)
}

// Correlation возвращает коэффициент корреляции Пирсона.
func Correlation(x, y []float64) float64 {
	if len(x) < 2 {
		return 0
	}
	stdX, stdY := StdDev(x), StdDev(y)
	if stdX == 0 || stdY == 0 {
		return 0
	}
	return Covariance(x, y) / (stdX * stdY)
}

// Percentile вычисляет p-й перцентиль (p от 0 до 100) методом линейной интерполяции.
// Использует алгоритм, близкий к numpy.percentile.
func Percentile(values []float64, p float64) float64 {
	if len(values) == 0 {
		return 0
	}
	if p <= 0 {
		return min(values)
	}
	if p >= 100 {
		return max(values)
	}
	sorted := make([]float64, len(values))
	copy(sorted, values)
	sort.Float64s(sorted)
	// Индекс как в numpy: (n-1)*p/100, затем интерполяция.
	n := float64(len(sorted))
	idx := (n - 1) * p / 100.0
	lo := int(math.Floor(idx))
	hi := int(math.Ceil(idx))
	if lo == hi {
		return sorted[lo]
	}
	frac := idx - float64(lo)
	return sorted[lo]*(1-frac) + sorted[hi]*frac
}

// Median возвращает 50-й перцентиль.
func Median(values []float64) float64 {
	return Percentile(values, 50)
}

func min(vals []float64) float64 {
	m := vals[0]
	for _, v := range vals[1:] {
		if v < m {
			m = v
		}
	}
	return m
}

func max(vals []float64) float64 {
	m := vals[0]
	for _, v := range vals[1:] {
		if v > m {
			m = v
		}
	}
	return m
}
