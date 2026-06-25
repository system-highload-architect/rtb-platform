package ports

import (
	"context"

	accountingv1 "rtb-platform/pb/accounting/v1"
	analyticsv1 "rtb-platform/pb/analytics/v1"
	auctionv1 "rtb-platform/pb/auction/v1"
	authv1 "rtb-platform/pb/auth/v1"
)

type AuctionPort interface {
	Bid(ctx context.Context, req *auctionv1.BidRequest) (*auctionv1.BidResponse, error)
}

type AccountingPort interface {
	Debit(ctx context.Context, req *accountingv1.DebitRequest) (*accountingv1.DebitResponse, error)
	GetBalance(ctx context.Context, req *accountingv1.GetBalanceRequest) (*accountingv1.GetBalanceResponse, error)
}

type AnalyticsPort interface {
	GetReport(ctx context.Context, req *analyticsv1.ReportRequest) ([]*analyticsv1.ReportRow, error)
	ExportExcel(ctx context.Context, req *analyticsv1.ReportRequest) ([]byte, error)
	Forecast(ctx context.Context, req *analyticsv1.ForecastRequest) (*analyticsv1.ForecastResponse, error)
	FactorAnalysis(ctx context.Context, req *analyticsv1.FactorRequest) (*analyticsv1.FactorResponse, error)
}

type AuthPort interface {
	Validate(ctx context.Context, token string) (*authv1.ValidateResponse, error)
	Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error)
	Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error)
}
