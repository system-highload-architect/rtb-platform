package server

import (
	"context"
	"log/slog"
	"strconv"

	auctionv1 "rtb-platform/pb/auction/v1"
	commonv1 "rtb-platform/pb/common/v1"

	"rtb-platform/pkg/device"
	"rtb-platform/pkg/experiment"
	"rtb-platform/pkg/geoip"
	"rtb-platform/pkg/idempotent"
	"rtb-platform/pkg/sampler"

	"rtb-platform/services/auction/internal/domain"
	"rtb-platform/services/auction/internal/domain/scoring"
	"rtb-platform/services/auction/internal/ports"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuctionServer struct {
	auctionv1.UnimplementedAuctionServiceServer
	userRepo     ports.UserProfileRepo
	campaignRepo ports.CampaignRepo
	fraud        ports.FraudDetector
	geoResolver  domain.GeoResolver
	scorer       scoring.Scorer
	altScorer    scoring.Scorer
	publisher    ports.EventPublisher
	idempotent   *idempotent.Store
	log          *slog.Logger
	geoipDB      *geoip.GeoDB
	sampler      *sampler.Sampler
	experiments  *experiment.Experiments
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
) *AuctionServer {
	return &AuctionServer{
		userRepo:     userRepo,
		campaignRepo: campaignRepo,
		fraud:        fraud,
		geoResolver:  geoResolver,
		scorer:       scorer,
		altScorer:    altScorer,
		publisher:    publisher,
		idempotent:   idemp,
		log:          log,
		geoipDB:      geoipDB,
		sampler:      sampler,
		experiments:  experiments,
	}
}

func (s *AuctionServer) Bid(ctx context.Context, req *auctionv1.BidRequest) (*auctionv1.BidResponse, error) {
	// Идемпотентность
	if req.IdempotencyKey != "" {
		if !s.idempotent.Check(req.IdempotencyKey) {
			return nil, status.Error(codes.AlreadyExists, "duplicate request")
		}
	}

	// Фрод
	if s.fraud.IsSuspicious(req.Ip, req.DeviceId) {
		return &auctionv1.BidResponse{Error: "fraud"}, nil
	}

	// Получаем профиль пользователя
	userProf, err := s.userRepo.Get(ctx, req.DeviceId)
	if err != nil {
		s.log.Error("failed to get user profile", "error", err)
		return nil, status.Errorf(codes.Internal, "user profile unavailable")
	}

	// Определение координат (если нет в профиле, используем GeoIP)
	lat, lng := userProf.Lat, userProf.Lng
	if lat == 0 && lng == 0 && req.Ip != "" && s.geoipDB != nil {
		if geoResult, err := s.geoipDB.Lookup(req.Ip); err == nil {
			lat = geoResult.Lat
			lng = geoResult.Lng
		}
	}

	// Определение типа устройства
	deviceType := device.UnknownDevice
	if req.UserAgent != "" {
		info := device.Parse(req.UserAgent)
		deviceType = info.Type
	}

	// Формируем признаки пользователя
	userFeats := scoring.UserFeatures{
		Features:   userProf.Features,
		Lat:        lat,
		Lng:        lng,
		DeviceType: deviceType,
	}

	// Получаем список кампаний
	campaigns, err := s.campaignRepo.GetActive(ctx)
	if err != nil {
		s.log.Error("failed to get campaigns", "error", err)
		return nil, status.Errorf(codes.Internal, "campaigns unavailable")
	}

	// Выбор скорера в зависимости от эксперимента
	effectiveScorer := s.scorer
	if s.experiments != nil && s.experiments.IsInExperiment(req.DeviceId, "new_scoring_v2") {
		if s.altScorer != nil {
			effectiveScorer = s.altScorer
		}
	}

	// Запуск аукциона
	result, err := domain.RunAuction(ctx, campaigns, userFeats, s.geoResolver, effectiveScorer)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return &auctionv1.BidResponse{Error: "no suitable campaign"}, nil
	}

	// Публикация события с сэмплированием
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
