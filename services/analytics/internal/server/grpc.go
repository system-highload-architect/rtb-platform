package server

import (
	"context"
	"log/slog"
	"time"

	analyticsv1 "rtb-platform/pb/analytics/v1"

	"rtb-platform/pkg/timeseries"

	"rtb-platform/services/analytics/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AnalyticsServer struct {
	analyticsv1.UnimplementedAnalyticsServiceServer
	reportSvc   *domain.ReportService
	forecastSvc *domain.ForecastService
	factorSvc   *domain.FactorService
	logger      *slog.Logger
}

func NewAnalyticsServer(
	reportSvc *domain.ReportService,
	forecastSvc *domain.ForecastService,
	factorSvc *domain.FactorService,
	logger *slog.Logger,
) *AnalyticsServer {
	return &AnalyticsServer{
		reportSvc:   reportSvc,
		forecastSvc: forecastSvc,
		factorSvc:   factorSvc,
		logger:      logger,
	}
}

func (s *AnalyticsServer) GetReport(req *analyticsv1.ReportRequest, stream analyticsv1.AnalyticsService_GetReportServer) error {
	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid start_date: %v", err)
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "invalid end_date: %v", err)
	}
	end = end.Add(24 * time.Hour)

	rows, err := s.reportSvc.GenerateReport(stream.Context(), start, end, req.Dimensions, req.Metrics)
	if err != nil {
		return status.Errorf(codes.Internal, "failed to generate report: %v", err)
	}

	for _, row := range rows {
		pbRow := &analyticsv1.ReportRow{
			DimensionValues: row.DimensionValues,
			MetricValues:    make(map[string]float64),
		}
		for k, v := range row.MetricValues {
			pbRow.MetricValues[k] = v
		}
		if err := stream.Send(pbRow); err != nil {
			return err
		}
	}
	return nil
}

func (s *AnalyticsServer) Forecast(ctx context.Context, req *analyticsv1.ForecastRequest) (*analyticsv1.ForecastResponse, error) {
	if req.Alpha == 0 {
		req.Alpha = 0.5
	}
	if req.Beta == 0 {
		req.Beta = 0.3
	}
	if req.Gamma == 0 {
		req.Gamma = 0.2
	}
	if req.Period == 0 {
		req.Period = 4 // для 15 значений хватит
	}
	forecast, err := s.forecastSvc.Forecast(ctx, req.History, int(req.Horizon), timeseries.HoltWintersParams{
		Alpha:  req.Alpha,
		Beta:   req.Beta,
		Gamma:  req.Gamma,
		Period: int(req.Period),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "forecast error: %v", err)
	}
	return &analyticsv1.ForecastResponse{Forecast: forecast}, nil
}

func (s *AnalyticsServer) FactorAnalysis(ctx context.Context, req *analyticsv1.FactorRequest) (*analyticsv1.FactorResponse, error) {
	var data [][]float64
	for _, u := range req.Users {
		data = append(data, u.Features)
	}
	pca, err := s.factorSvc.Analyse(ctx, data, int(req.NComponents))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "pca error: %v", err)
	}
	return &analyticsv1.FactorResponse{
		ExplainedVarianceRatio: pca.ExplainedVarianceRatio(),
	}, nil
}
