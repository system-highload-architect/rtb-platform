package server

import (
	"context"
	"log/slog"

	accountingv1 "rtb-platform/pb/accounting/v1"
	commonv1 "rtb-platform/pb/common/v1"
	"rtb-platform/pkg/fixedpoint"
	"rtb-platform/services/accounting/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AccountingServer struct {
	accountingv1.UnimplementedAccountingServiceServer
	svc    *domain.Service
	logger *slog.Logger
}

func NewAccountingServer(svc *domain.Service, logger *slog.Logger) *AccountingServer {
	return &AccountingServer{svc: svc, logger: logger}
}

func (s *AccountingServer) Debit(ctx context.Context, req *accountingv1.DebitRequest) (*accountingv1.DebitResponse, error) {
	amount := fixedpoint.Money(req.Amount.Amount)
	err := s.svc.Debit(ctx, req.CampaignId, amount, req.BidId)
	if err != nil {
		if err == domain.ErrInsufficientFunds {
			return &accountingv1.DebitResponse{
				Success: false,
				Error:   "insufficient funds",
			}, nil
		}
		if err == domain.ErrCampaignNotFound {
			return nil, status.Error(codes.NotFound, "campaign not found")
		}
		s.logger.Error("debit failed", "error", err)
		return nil, status.Error(codes.Internal, "internal error")
	}

	bal, _ := s.svc.GetBalance(ctx, req.CampaignId)
	return &accountingv1.DebitResponse{
		Success:          true,
		RemainingBalance: &commonv1.Money{Amount: int64(bal), Scale: 2},
	}, nil
}

func (s *AccountingServer) GetBalance(ctx context.Context, req *accountingv1.GetBalanceRequest) (*accountingv1.GetBalanceResponse, error) {
	bal, err := s.svc.GetBalance(ctx, req.CampaignId)
	if err != nil {
		return nil, status.Error(codes.NotFound, "campaign not found")
	}
	return &accountingv1.GetBalanceResponse{
		Balance: &commonv1.Money{Amount: int64(bal), Scale: 2},
	}, nil
}
