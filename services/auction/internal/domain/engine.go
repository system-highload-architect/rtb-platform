package domain

import (
	"context"

	"rtb-platform/pkg/geospatial"
	"rtb-platform/pkg/statistics"

	"rtb-platform/services/auction/internal/domain/scoring"
)

func RunAuction(
	ctx context.Context,
	campaigns []Campaign,
	userFeatures scoring.UserFeatures,
	geoResolver GeoResolver,
	scorer scoring.Scorer,
) (*BidResponse, error) {
	if len(campaigns) == 0 {
		return nil, nil
	}

	type candidate struct {
		campaign     Campaign
		effectiveBid int64
		score        float64
	}
	candidates := make([]candidate, 0, len(campaigns))
	userPoint := geospatial.Point{Lat: userFeatures.Lat, Lng: userFeatures.Lng}

	for _, camp := range campaigns {
		var distance float64 = 1e9
		if geoResolver != nil {
			if point, ok := geoResolver.GetBillboardLocation(camp.ID); ok {
				distance = geospatial.HaversineDistance(userPoint, point)
			}
		}
		campData := scoring.CampaignScoringData{
			ID:           camp.ID,
			BidCents:     camp.BidCents,
			BillboardLat: camp.BillboardLat,
			BillboardLng: camp.BillboardLng,
			CreativeURL:  camp.CreativeURL,
		}
		score, optBid := scorer.Score(campData, userFeatures, distance)
		candidates = append(candidates, candidate{
			campaign:     camp,
			effectiveBid: optBid,
			score:        score,
		})
	}

	// Сортировка
	bids := make([]int64, len(candidates))
	indices := make([]int, len(candidates))
	for i, c := range candidates {
		bids[i] = c.effectiveBid
		indices[i] = i
	}
	SortCandidatesByEffectiveBid(bids, indices)
	winnerIdx := indices[len(indices)-1]
	winner := candidates[winnerIdx]

	// Статистики для мониторинга (при желании можно залогировать или отправить в метрики)
	if len(bids) > 0 {
		bidsFloat := make([]float64, len(bids))
		for i, v := range bids {
			bidsFloat[i] = float64(v)
		}
		_ = statistics.Median(bidsFloat)
		_ = statistics.Percentile(bidsFloat, 0.9)
	}

	return &BidResponse{
		CampaignID:  winner.campaign.ID,
		BidPrice:    winner.campaign.BidCents,
		CreativeURL: winner.campaign.CreativeURL,
	}, nil
}
