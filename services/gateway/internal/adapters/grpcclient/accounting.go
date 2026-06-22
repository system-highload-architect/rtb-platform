package grpcclient

import (
	"context"

	"log/slog"

	accountingv1 "rtb-platform/pb/accounting/v1"
	"rtb-platform/services/gateway/internal/ports"
)

type accountingAdapter struct {
	client accountingv1.AccountingServiceClient
	logger *slog.Logger
}

func NewAccountingPort(client accountingv1.AccountingServiceClient, logger *slog.Logger) ports.AccountingPort {
	return &accountingAdapter{client: client, logger: logger}
}

func (a *accountingAdapter) Debit(ctx context.Context, req *accountingv1.DebitRequest) (*accountingv1.DebitResponse, error) {
	resp, err := a.client.Debit(ctx, req)
	if err != nil {
		a.logger.Error("accounting debit failed", "error", err)
		return nil, err
	}
	return resp, nil
}

func (a *accountingAdapter) GetBalance(ctx context.Context, req *accountingv1.GetBalanceRequest) (*accountingv1.GetBalanceResponse, error) {
	resp, err := a.client.GetBalance(ctx, req)
	if err != nil {
		a.logger.Error("accounting get balance failed", "error", err)
		return nil, err
	}
	return resp, nil
}
