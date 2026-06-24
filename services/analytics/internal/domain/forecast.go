package domain

import (
	"context"
	"rtb-platform/pkg/timeseries"
)

type ForecastService struct{}

func NewForecastService() *ForecastService {
	return &ForecastService{}
}

func (s *ForecastService) Forecast(ctx context.Context, history []float64, horizon int, params timeseries.HoltWintersParams) ([]float64, error) {
	return timeseries.HoltWintersForecast(history, horizon, params)
}
