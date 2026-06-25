package server

import (
	"context"
	"crypto/sha256"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"

	authv1 "rtb-platform/pb/auth/v1"

	"rtb-platform/services/auth/internal/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type AuthServer struct {
	authv1.UnimplementedAuthServiceServer
	store      domain.UserStore
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	logger     *slog.Logger
}

func NewAuthServer(store domain.UserStore, jwtSecret string, accessTTL, refreshTTL time.Duration, logger *slog.Logger) *AuthServer {
	return &AuthServer{
		store:      store,
		jwtSecret:  []byte(jwtSecret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		logger:     logger,
	}
}

func (s *AuthServer) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password required")
	}
	hash := sha256.Sum256([]byte(req.Password))
	user := domain.User{
		ID:           fmt.Sprintf("user-%d", time.Now().UnixNano()),
		Email:        req.Email,
		PasswordHash: fmt.Sprintf("%x", hash),
		Role:         req.Role,
		CreatedAt:    time.Now(),
	}
	if err := s.store.Create(user); err != nil {
		if err == domain.ErrUserExists {
			return &authv1.RegisterResponse{Error: "user already exists"}, nil
		}
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}
	return &authv1.RegisterResponse{UserId: user.ID}, nil
}

func (s *AuthServer) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	user, ok := s.store.GetByEmail(req.Email)
	if !ok {
		return &authv1.LoginResponse{Error: "invalid credentials"}, nil
	}
	hash := sha256.Sum256([]byte(req.Password))
	if fmt.Sprintf("%x", hash) != user.PasswordHash {
		return &authv1.LoginResponse{Error: "invalid credentials"}, nil
	}
	accessToken, err := s.generateToken(user, s.accessTTL)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token: %v", err)
	}
	refreshToken, err := s.generateToken(user, s.refreshTTL)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to generate token: %v", err)
	}
	return &authv1.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, nil
}

func (s *AuthServer) Validate(ctx context.Context, req *authv1.ValidateRequest) (*authv1.ValidateResponse, error) {
	token, err := jwt.Parse(req.AccessToken, func(t *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})
	if err != nil || !token.Valid {
		return &authv1.ValidateResponse{Valid: false}, nil
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return &authv1.ValidateResponse{Valid: false}, nil
	}
	userID, _ := claims["user_id"].(string)
	role, _ := claims["role"].(string)
	exp, _ := claims["exp"].(float64)
	expTime := time.Unix(int64(exp), 0)
	return &authv1.ValidateResponse{
		Valid:     true,
		UserId:    userID,
		Role:      role,
		ExpiresAt: timestamppb.New(expTime),
	}, nil
}

func (s *AuthServer) Refresh(ctx context.Context, req *authv1.RefreshRequest) (*authv1.RefreshResponse, error) {
	// упрощённо – просто валидируем refresh и выдаём новые токены
	validResp, err := s.Validate(ctx, &authv1.ValidateRequest{AccessToken: req.RefreshToken})
	if err != nil || !validResp.Valid {
		return &authv1.RefreshResponse{Error: "invalid refresh token"}, nil
	}
	user, ok := s.store.GetByEmail(validResp.UserId) // в нашем хранилище ID = email, но это временно
	if !ok {
		return &authv1.RefreshResponse{Error: "user not found"}, nil
	}
	accessToken, _ := s.generateToken(user, s.accessTTL)
	refreshToken, _ := s.generateToken(user, s.refreshTTL)
	return &authv1.RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, nil
}

func (s *AuthServer) generateToken(user domain.User, ttl time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(ttl).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}
