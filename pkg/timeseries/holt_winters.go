package timeseries

import "errors"

// HoltWintersParams задаёт параметры модели Хольт-Уинтерс.
type HoltWintersParams struct {
	Alpha  float64 // сглаживание уровня
	Beta   float64 // сглаживание тренда
	Gamma  float64 // сглаживание сезонности
	Period int     // длина сезона (например, 24 для часов в сутках)
}

// HoltWintersForecast строит прогноз на следующий горизонт,
// используя тройное экспоненциальное сглаживание (аддитивная сезонность).
// data — исторические значения.
func HoltWintersForecast(data []float64, horizon int, params HoltWintersParams) ([]float64, error) {
	if len(data) < 2*params.Period {
		return nil, errors.New("timeseries: insufficient data for seasonality")
	}
	if params.Period <= 0 {
		return nil, errors.New("timeseries: period must be positive")
	}

	// Инициализация начальных уровней, тренда и сезонных компонент
	level := make([]float64, len(data))
	trend := make([]float64, len(data))
	seasonal := make([]float64, len(data)+params.Period) // индекс + период

	// Инициализация сезонности по первым двум периодам
	for i := 0; i < params.Period; i++ {
		seasonal[i] = data[i] - average(data[:params.Period])
	}
	// Начальный уровень и тренд
	level[params.Period-1] = data[params.Period-1] - seasonal[params.Period-1]
	trend[params.Period-1] = 0

	// Обновление параметров по мере поступления данных
	for t := params.Period; t < len(data); t++ {
		level[t] = params.Alpha*(data[t]-seasonal[t-params.Period]) + (1-params.Alpha)*(level[t-1]+trend[t-1])
		trend[t] = params.Beta*(level[t]-level[t-1]) + (1-params.Beta)*trend[t-1]
		seasonal[t] = params.Gamma*(data[t]-level[t]) + (1-params.Gamma)*seasonal[t-params.Period]
	}

	// Прогноз
	forecast := make([]float64, horizon)
	lastLevel := level[len(data)-1]
	lastTrend := trend[len(data)-1]
	for h := 0; h < horizon; h++ {
		// Определяем сезонный индекс для будущего шага
		sIdx := (len(data) - params.Period) + h%params.Period
		if sIdx >= len(data) {
			sIdx = len(data) - params.Period + (h % params.Period)
		}
		forecast[h] = lastLevel + float64(h+1)*lastTrend + seasonal[sIdx]
	}
	return forecast, nil
}

func average(vals []float64) float64 {
	if len(vals) == 0 {
		return 0
	}
	sum := 0.0
	for _, v := range vals {
		sum += v
	}
	return sum / float64(len(vals))
}
