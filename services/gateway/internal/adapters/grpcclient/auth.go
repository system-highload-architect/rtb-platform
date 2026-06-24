package grpcclient

import (
	"context"
	"log/slog"

	authv1 "rtb-platform/pb/auth/v1"

	"rtb-platform/services/gateway/internal/ports"
)

type authAdapter struct {
	client authv1.AuthServiceClient
	logger *slog.Logger
}

func NewAuthPort(client authv1.AuthServiceClient, logger *slog.Logger) ports.AuthPort {
	return &authAdapter{client: client, logger: logger}
}

func (a *authAdapter) Validate(ctx context.Context, token string) (*authv1.ValidateResponse, error) {
	resp, err := a.client.Validate(ctx, &authv1.ValidateRequest{AccessToken: token})
	if err != nil {
		a.logger.Error("auth validate failed", "error", err)
		return nil, err
	}
	return resp, nil
}
