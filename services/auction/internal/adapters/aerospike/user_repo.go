package aerospike

import (
	"context"

	"rtb-platform/pkg/breaker"
	"rtb-platform/services/auction/internal/ports"

	as "github.com/aerospike/aerospike-client-go/v7"
)

type userRepo struct {
	client  *as.Client
	ns      string
	set     string
	breaker *breaker.Breaker
}

// NewUserRepo создаёт репозиторий с circuit breaker.
func NewUserRepo(client *as.Client, namespace, set string, cb *breaker.Breaker) ports.UserProfileRepo {
	return &userRepo{
		client:  client,
		ns:      namespace,
		set:     set,
		breaker: cb,
	}
}

func (r *userRepo) Get(ctx context.Context, deviceID string) (*ports.UserProfile, error) {
	var profile *ports.UserProfile
	err := r.breaker.Execute(ctx, func() error {
		key, err := as.NewKey(r.ns, r.set, deviceID)
		if err != nil {
			return err
		}
		rec, err := r.client.Get(nil, key)
		if err != nil {
			return err
		}
		// TODO: маппинг rec.Bins в UserProfile
		_ = rec
		profile = &ports.UserProfile{DeviceID: deviceID}
		return nil
	})
	return profile, err
}
