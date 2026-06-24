package grpcclient

import (
	"context"
	"fmt"
	"io"

	"log/slog"

	analyticsv1 "rtb-platform/pb/analytics/v1"

	"rtb-platform/services/gateway/internal/ports"

	"github.com/xuri/excelize/v2"
)

type analyticsAdapter struct {
	client analyticsv1.AnalyticsServiceClient
	logger *slog.Logger
}

func NewAnalyticsPort(client analyticsv1.AnalyticsServiceClient, logger *slog.Logger) ports.AnalyticsPort {
	return &analyticsAdapter{client: client, logger: logger}
}

func (a *analyticsAdapter) GetReport(ctx context.Context, req *analyticsv1.ReportRequest) ([]*analyticsv1.ReportRow, error) {
	stream, err := a.client.GetReport(ctx, req)
	if err != nil {
		a.logger.Error("analytics get report failed", "error", err)
		return nil, err
	}
	var rows []*analyticsv1.ReportRow
	for {
		row, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			a.logger.Error("analytics recv row failed", "error", err)
			return nil, err
		}
		rows = append(rows, row)
	}
	return rows, nil
}

func (a *analyticsAdapter) ExportExcel(ctx context.Context, req *analyticsv1.ReportRequest) ([]byte, error) {
	rows, err := a.GetReport(ctx, req)
	if err != nil {
		return nil, err
	}
	f := excelize.NewFile()
	sheet := "Report"
	f.SetSheetName("Sheet1", sheet)
	// Заголовки из первого ряда метрик и измерений
	if len(rows) > 0 {
		col := 1
		for dim := range rows[0].DimensionValues {
			f.SetCellValue(sheet, cellName(1, col), dim)
			col++
		}
		for met := range rows[0].MetricValues {
			f.SetCellValue(sheet, cellName(1, col), met)
			col++
		}
	}
	for i, row := range rows {
		col := 1
		for _, v := range row.DimensionValues {
			f.SetCellValue(sheet, cellName(i+2, col), v)
			col++
		}
		for _, v := range row.MetricValues {
			f.SetCellValue(sheet, cellName(i+2, col), v)
			col++
		}
	}
	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func cellName(row, col int) string {
	colName, _ := excelize.ColumnNumberToName(col)
	return colName + fmt.Sprint(row)
}
func (a *analyticsAdapter) Forecast(ctx context.Context, req *analyticsv1.ForecastRequest) (*analyticsv1.ForecastResponse, error) {
	resp, err := a.client.Forecast(ctx, req)
	if err != nil {
		a.logger.Error("analytics forecast failed", "error", err)
		return nil, err
	}
	return resp, nil
}

func (a *analyticsAdapter) FactorAnalysis(ctx context.Context, req *analyticsv1.FactorRequest) (*analyticsv1.FactorResponse, error) {
	resp, err := a.client.FactorAnalysis(ctx, req)
	if err != nil {
		a.logger.Error("analytics factor analysis failed", "error", err)
		return nil, err
	}
	return resp, nil
}
