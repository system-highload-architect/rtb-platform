package server

import (
	"context"
	"log/slog"
	"strconv"

	accountingv1 "rtb-platform/pb/accounting/v1"
	auctionv1 "rtb-platform/pb/auction/v1"
	commonv1 "rtb-platform/pb/common/v1"

	"rtb-platform/pkg/device"
	"rtb-platform/pkg/experiment"
	"rtb-platform/pkg/geoip"
	"rtb-platform/pkg/idempotent"
	"rtb-platform/pkg/sampler"
	"rtb-platform/pkg/timedcache"

	"rtb-platform/services/auction/internal/domain"
	"rtb-platform/services/auction/internal/domain/scoring"
	"rtb-platform/services/auction/internal/ports"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuctionServer struct {
	auctionv1.UnimplementedAuctionServiceServer
	userRepo         ports.UserProfileRepo
	campaignRepo     ports.CampaignRepo
	fraud            ports.FraudDetector
	geoResolver      domain.GeoResolver
	scorer           scoring.Scorer
	altScorer        scoring.Scorer
	publisher        ports.EventPublisher
	idempotent       *idempotent.Store
	log              *slog.Logger
	geoipDB          *geoip.GeoDB
	sampler          *sampler.Sampler
	experiments      *experiment.Experiments
	accountingClient ports.AccountingPort
	auctionCache     *timedcache.Cache[string, []*auctionv1.BidRequest]
}

func NewAuctionServer(
	userRepo ports.UserProfileRepo,
	campaignRepo ports.CampaignRepo,
	fraud ports.FraudDetector,
	geoResolver domain.GeoResolver,
	scorer scoring.Scorer,
	altScorer scoring.Scorer,
	publisher ports.EventPublisher,
	idemp *idempotent.Store,
	log *slog.Logger,
	geoipDB *geoip.GeoDB,
	sampler *sampler.Sampler,
	experiments *experiment.Experiments,
	accountingPort ports.AccountingPort,
) *AuctionServer {
	return &AuctionServer{
		userRepo:         userRepo,
		campaignRepo:     campaignRepo,
		fraud:            fraud,
		geoResolver:      geoResolver,
		scorer:           scorer,
		altScorer:        altScorer,
		publisher:        publisher,
		idempotent:       idemp,
		log:              log,
		geoipDB:          geoipDB,
		sampler:          sampler,
		experiments:      experiments,
		accountingClient: accountingPort,
	}
}

func (s *AuctionServer) Bid(ctx context.Context, req *auctionv1.BidRequest) (*auctionv1.BidResponse, error) {
	// Асинхронный режим: сразу кладём в кэш и отвечаем "accepted"
	if s.auctionCache != nil {
		if req.IdempotencyKey != "" {
			if !s.idempotent.Check(req.IdempotencyKey) {
				return nil, status.Error(codes.AlreadyExists, "duplicate request")
			}
		}
		// fraud check ДО кэширования
		if s.fraud.IsSuspicious(req.Ip, req.DeviceId) {
			return &auctionv1.BidResponse{Error: "fraud"}, nil
		}
		key := "global"
		existing, _ := s.auctionCache.Get(key)
		existing = append(existing, req)
		s.auctionCache.Set(key, existing)
		return &auctionv1.BidResponse{Error: "accepted"}, nil
	}

	// Синхронный режим (без кэша) – существующая логика
	if req.IdempotencyKey != "" {
		if !s.idempotent.Check(req.IdempotencyKey) {
			return nil, status.Error(codes.AlreadyExists, "duplicate request")
		}
	}

	// Профиль пользователя
	userProf, err := s.userRepo.Get(ctx, req.DeviceId)
	if err != nil {
		s.log.Error("failed to get user profile", "error", err)
		return nil, status.Errorf(codes.Internal, "user profile unavailable")
	}

	// Определение координат
	lat, lng := userProf.Lat, userProf.Lng
	if lat == 0 && lng == 0 && req.Ip != "" && s.geoipDB != nil {
		if geoResult, err := s.geoipDB.Lookup(req.Ip); err == nil {
			lat = geoResult.Lat
			lng = geoResult.Lng
		}
	}

	// Тип устройства
	deviceType := device.UnknownDevice
	if req.UserAgent != "" {
		info := device.Parse(req.UserAgent)
		deviceType = info.Type
	}

	userFeats := scoring.UserFeatures{
		Features:   userProf.Features,
		Lat:        lat,
		Lng:        lng,
		DeviceType: deviceType,
	}

	campaigns, err := s.campaignRepo.GetActive(ctx)
	if err != nil {
		s.log.Error("failed to get campaigns", "error", err)
		return nil, status.Errorf(codes.Internal, "campaigns unavailable")
	}

	effectiveScorer := s.scorer
	if s.experiments != nil && s.experiments.IsInExperiment(req.DeviceId, "new_scoring_v2") {
		if s.altScorer != nil {
			effectiveScorer = s.altScorer
		}
	}

	result, err := domain.RunAuction(ctx, campaigns, userFeats, s.geoResolver, effectiveScorer)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return &auctionv1.BidResponse{Error: "no suitable campaign"}, nil
	}

	if s.accountingClient != nil {
		debitReq := &accountingv1.DebitRequest{
			CampaignId: strconv.FormatUint(uint64(result.CampaignID), 10),
			Amount:     &commonv1.Money{Amount: result.BidPrice, Scale: 2},
			BidId:      req.IdempotencyKey,
		}
		if _, err := s.accountingClient.Debit(ctx, debitReq); err != nil {
			s.log.Error("failed to debit campaign", "error", err)
		}
	}

	if s.sampler == nil || s.sampler.Sample() {
		event := domain.BidEvent{
			BidID:      req.IdempotencyKey,
			CampaignID: result.CampaignID,
			DeviceID:   req.DeviceId,
			PriceCents: result.BidPrice,
			Win:        true,
		}
		if err := s.publisher.PublishBidEvent(ctx, event); err != nil {
			s.log.Error("failed to publish bid event", "error", err)
		}
	}

	return &auctionv1.BidResponse{
		CampaignId:  strconv.FormatUint(uint64(result.CampaignID), 10),
		BidPrice:    &commonv1.Money{Amount: result.BidPrice, Scale: 2},
		CreativeUrl: result.CreativeURL,
	}, nil
}

func (s *AuctionServer) SetAuctionCache(c *timedcache.Cache[string, []*auctionv1.BidRequest]) {
	s.auctionCache = c
}

func (s *AuctionServer) ProcessBids(ctx context.Context, key string, requests []*auctionv1.BidRequest) {
	for _, req := range requests {
		s.processBid(ctx, req)
	}
}

// внутренний метод, аналогичен синхронному Bid, но без возврата ответа
func (s *AuctionServer) processBid(ctx context.Context, req *auctionv1.BidRequest) {
	// Идемпотентность
	if req.IdempotencyKey != "" {
		if !s.idempotent.Check(req.IdempotencyKey) {
			return
		}
	}

	// Фрод
	if s.fraud.IsSuspicious(req.Ip, req.DeviceId) {
		return
	}

	// Профиль пользователя
	userProf, err := s.userRepo.Get(ctx, req.DeviceId)
	if err != nil {
		s.log.Error("failed to get user profile", "error", err)
		return
	}

	// Определение координат (GeoIP)
	lat, lng := userProf.Lat, userProf.Lng
	if lat == 0 && lng == 0 && req.Ip != "" && s.geoipDB != nil {
		if geoResult, err := s.geoipDB.Lookup(req.Ip); err == nil {
			lat = geoResult.Lat
			lng = geoResult.Lng
		}
	}

	// Тип устройства
	deviceType := device.UnknownDevice
	if req.UserAgent != "" {
		info := device.Parse(req.UserAgent)
		deviceType = info.Type
	}

	userFeats := scoring.UserFeatures{
		Features:   userProf.Features,
		Lat:        lat,
		Lng:        lng,
		DeviceType: deviceType,
	}

	// Кампании
	campaigns, err := s.campaignRepo.GetActive(ctx)
	if err != nil {
		s.log.Error("failed to get campaigns", "error", err)
		return
	}

	// Выбор скорера с экспериментом
	effectiveScorer := s.scorer
	if s.experiments != nil && s.experiments.IsInExperiment(req.DeviceId, "new_scoring_v2") {
		if s.altScorer != nil {
			effectiveScorer = s.altScorer
		}
	}

	// Аукцион
	result, err := domain.RunAuction(ctx, campaigns, userFeats, s.geoResolver, effectiveScorer)
	if err != nil {
		s.log.Error("auction failed", "error", err)
		return
	}
	if result == nil {
		return
	}

	// Списание бюджета
	if s.accountingClient != nil {
		debitReq := &accountingv1.DebitRequest{
			CampaignId: strconv.FormatUint(uint64(result.CampaignID), 10),
			Amount:     &commonv1.Money{Amount: result.BidPrice, Scale: 2},
			BidId:      req.IdempotencyKey,
		}
		if _, err := s.accountingClient.Debit(ctx, debitReq); err != nil {
			s.log.Error("failed to debit campaign", "error", err)
		}
	}

	// Публикация события
	if s.sampler == nil || s.sampler.Sample() {
		event := domain.BidEvent{
			BidID:      req.IdempotencyKey,
			CampaignID: result.CampaignID,
			DeviceID:   req.DeviceId,
			PriceCents: result.BidPrice,
			Win:        true,
		}
		if err := s.publisher.PublishBidEvent(ctx, event); err != nil {
			s.log.Error("failed to publish bid event", "error", err)
		}
	}
}
