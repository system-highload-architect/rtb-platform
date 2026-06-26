package mongodb

import (
	"context"
	"rtb-platform/pkg/breaker"
	"rtb-platform/pkg/timedcache"
	"rtb-platform/services/auction/internal/domain"
	"rtb-platform/services/auction/internal/ports"
	"time"
)

type cachedRepo struct {
	cache   *timedcache.Cache[uint32, domain.Campaign]
	breaker *breaker.Breaker
}

func NewCachedCampaignRepo(ttl time.Duration, cb *breaker.Breaker) ports.CampaignRepo {
	c := timedcache.New[uint32, domain.Campaign](ttl)
	return &cachedRepo{cache: c, breaker: cb}
}

func (r *cachedRepo) GetActive(ctx context.Context) ([]domain.Campaign, error) {
	return r.cache.Values(), nil
}

func (r *cachedRepo) Update(campaigns []domain.Campaign) {
	for _, c := range campaigns {
		r.cache.Set(c.ID, c)
	}
}

func (r *cachedRepo) Stop() {
	r.cache.Stop()
}
