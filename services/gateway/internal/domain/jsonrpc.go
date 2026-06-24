package domain

import (
	"context"
	"fmt"

	"github.com/goccy/go-json"

	"log/slog"

	accountingv1 "rtb-platform/pb/accounting/v1"
	auctionv1 "rtb-platform/pb/auction/v1"
	"rtb-platform/pkg/registry"
	"rtb-platform/services/gateway/internal/ports"

	"google.golang.org/protobuf/proto"
)

type JSONRPCService struct {
	registry *registry.Registry[string, json.RawMessage, proto.Message] // <-- типизированный ответ
	logger   *slog.Logger
}

func NewJSONRPCService(
	auction ports.AuctionPort,
	accounting ports.AccountingPort,
	analytics ports.AnalyticsPort,
	logger *slog.Logger,
) *JSONRPCService {
	s := &JSONRPCService{
		registry: registry.New[string, json.RawMessage, proto.Message](),
		logger:   logger,
	}
	s.registerHandlers(auction, accounting)
	return s
}

func (s *JSONRPCService) registerHandlers(auction ports.AuctionPort, accounting ports.AccountingPort) {
	s.registry.Register("auction.bid", func(ctx context.Context, raw json.RawMessage) (proto.Message, error) {
		var req auctionv1.BidRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			return nil, fmt.Errorf("invalid bid request: %w", err)
		}
		resp, err := auction.Bid(ctx, &req)
		if err != nil {
			return nil, err
		}
		return resp, nil // *BidResponse реализует proto.Message
	})

	s.registry.Register("accounting.debit", func(ctx context.Context, raw json.RawMessage) (proto.Message, error) {
		var req accountingv1.DebitRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			return nil, fmt.Errorf("invalid debit request: %w", err)
		}
		resp, err := accounting.Debit(ctx, &req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	})

	s.registry.Register("accounting.getBalance", func(ctx context.Context, raw json.RawMessage) (proto.Message, error) {
		var req accountingv1.GetBalanceRequest
		if err := json.Unmarshal(raw, &req); err != nil {
			return nil, fmt.Errorf("invalid getBalance request: %w", err)
		}
		resp, err := accounting.GetBalance(ctx, &req)
		if err != nil {
			return nil, err
		}
		return resp, nil
	})
}

func (s *JSONRPCService) Dispatch(ctx context.Context, method string, params json.RawMessage) (proto.Message, error) {
	return s.registry.Dispatch(ctx, method, params)
}
