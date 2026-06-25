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
func (a *authAdapter) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	resp, err := a.client.Register(ctx, req)
	if err != nil {
		a.logger.Error("auth register failed", "error", err)
		return nil, err
	}
	return resp, nil
}

func (a *authAdapter) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	resp, err := a.client.Login(ctx, req)
	if err != nil {
		a.logger.Error("auth login failed", "error", err)
		return nil, err
	}
	return resp, nil
}
