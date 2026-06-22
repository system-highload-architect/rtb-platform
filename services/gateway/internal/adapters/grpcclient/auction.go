package grpcclient

import (
	"context"

	"log/slog"

	auctionv1 "rtb-platform/pb/auction/v1"
	"rtb-platform/services/gateway/internal/ports"
)

type auctionAdapter struct {
	client auctionv1.AuctionServiceClient
	logger *slog.Logger
}

func NewAuctionPort(client auctionv1.AuctionServiceClient, logger *slog.Logger) ports.AuctionPort {
	return &auctionAdapter{client: client, logger: logger}
}

func (a *auctionAdapter) Bid(ctx context.Context, req *auctionv1.BidRequest) (*auctionv1.BidResponse, error) {
	resp, err := a.client.Bid(ctx, req)
	if err != nil {
		a.logger.Error("auction bid failed", "error", err)
		return nil, err
	}
	return resp, nil
}
