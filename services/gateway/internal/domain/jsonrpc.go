package domain

import (
	"context"
	"fmt"

	"github.com/goccy/go-json"

	"log/slog"

	accountingv1 "rtb-platform/pb/accounting/v1"
	auctionv1 "rtb-platform/pb/auction/v1"
	authv1 "rtb-platform/pb/auth/v1"
	"rtb-platform/pkg/registry"
	"rtb-platform/services/gateway/internal/ports"

	"google.golang.org/protobuf/proto"
)

type JSONRPCService struct {
	registry *registry.Registry[string, json.RawMessage, proto.Message]
	logger   *slog.Logger
}

func NewJSONRPCService(
	auction ports.AuctionPort,
	accounting ports.AccountingPort,
	analytics ports.AnalyticsPort,
	authPort ports.AuthPort,
	logger *slog.Logger,
) *JSONRPCService {
	s := &JSONRPCService{
		registry: registry.New[string, json.RawMessage, proto.Message](),
		logger:   logger,
	}
	s.registerHandlers(auction, accounting, authPort)
	return s
}

func (s *JSONRPCService) registerHandlers(
	auction ports.AuctionPort,
	accounting ports.AccountingPort,
	authPort ports.AuthPort,
) {
	// auction.bid
	s.registry.Register("auction.bid", func(ctx context.Context, raw json.RawMessage) (proto.Message, error) {
		var req auctionv1.BidRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			return nil, fmt.Errorf("invalid bid request: %w", err)
		}
		return auction.Bid(ctx, &req)
	})

	// accounting.debit
	s.registry.Register("accounting.debit", func(ctx context.Context, raw json.RawMessage) (proto.Message, error) {
		var req accountingv1.DebitRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			return nil, fmt.Errorf("invalid debit request: %w", err)
		}
		return accounting.Debit(ctx, &req)
	})

	// auth.register
	s.registry.Register("auth.register", func(ctx context.Context, raw json.RawMessage) (proto.Message, error) {
		var req authv1.RegisterRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			return nil, fmt.Errorf("invalid register request: %w", err)
		}
		return authPort.Register(ctx, &req)
	})

	// auth.login
	s.registry.Register("auth.login", func(ctx context.Context, raw json.RawMessage) (proto.Message, error) {
		var req authv1.LoginRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			return nil, fmt.Errorf("invalid login request: %w", err)
		}
		return authPort.Login(ctx, &req)
	})
}

func (s *JSONRPCService) Dispatch(ctx context.Context, method string, params json.RawMessage) (proto.Message, error) {
	return s.registry.Dispatch(ctx, method, params)
}
