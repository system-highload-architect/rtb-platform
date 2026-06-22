package ports

import (
	"context"

	accountingv1 "rtb-platform/pb/accounting/v1"
	analyticsv1 "rtb-platform/pb/analytics/v1"
	auctionv1 "rtb-platform/pb/auction/v1"
)

// AuctionPort – вызов аукциона.
type AuctionPort interface {
	Bid(ctx context.Context, req *auctionv1.BidRequest) (*auctionv1.BidResponse, error)
}

// AccountingPort – работа с бюджетом.
type AccountingPort interface {
	Debit(ctx context.Context, req *accountingv1.DebitRequest) (*accountingv1.DebitResponse, error)
	GetBalance(ctx context.Context, req *accountingv1.GetBalanceRequest) (*accountingv1.GetBalanceResponse, error)
}

// AnalyticsPort – данные для дашборда.
type AnalyticsPort interface {
	GetReport(ctx context.Context, req *analyticsv1.ReportRequest) ([]*analyticsv1.ReportRow, error)
	// ExportExcel возвращает бинарные данные xlsx-файла.
	ExportExcel(ctx context.Context, req *analyticsv1.ReportRequest) ([]byte, error)
}
