package mongodb

import (
	"context"
	"sync"
	"time"

	"rtb-platform/pkg/breaker"
	"rtb-platform/services/auction/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoRepo struct {
	client     *mongo.Client
	collection *mongo.Collection
	cache      []domain.Campaign
	mu         sync.RWMutex
	breaker    *breaker.Breaker
}

func NewMongoRepo(ctx context.Context, uri, dbName string, breaker *breaker.Breaker) (*mongoRepo, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	repo := &mongoRepo{
		client:     client,
		collection: client.Database(dbName).Collection("campaigns"),
		breaker:    breaker,
	}
	if err := repo.refreshCache(ctx); err != nil {
		return nil, err
	}
	return repo, nil
}

func (r *mongoRepo) GetActive(ctx context.Context) ([]domain.Campaign, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]domain.Campaign, len(r.cache))
	copy(result, r.cache)
	return result, nil
}

func (r *mongoRepo) refreshCache(ctx context.Context) error {
	var campaigns []domain.Campaign
	err := r.breaker.Execute(ctx, func() error {
		cursor, err := r.collection.Find(ctx, bson.M{"active": true})
		if err != nil {
			return err
		}
		defer cursor.Close(ctx)
		return cursor.All(ctx, &campaigns)
	})
	if err != nil {
		return err
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.cache = campaigns
	return nil
}

// Метод для периодического обновления (можно вызывать из main в горутине)
func (r *mongoRepo) StartAutoRefresh(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				r.refreshCache(ctx)
			}
		}
	}()
}

// Update обновляет кампании в MongoDB и кэше.
func (r *mongoRepo) Update(campaigns []domain.Campaign) {
	ctx := context.Background()
	for _, c := range campaigns {
		_ = r.breaker.Execute(ctx, func() error {
			_, err := r.collection.UpdateOne(ctx,
				bson.M{"id": c.ID},
				bson.M{"$set": bson.M{
					"id":            c.ID,
					"bid_cents":     c.BidCents,
					"daily_budget":  c.DailyBudget,
					"creative_url":  c.CreativeURL,
					"billboard_lat": c.BillboardLat,
					"billboard_lng": c.BillboardLng,
					"active":        true,
				}},
				options.Update().SetUpsert(true),
			)
			return err
		})
	}
	r.refreshCache(ctx) // обновить кэш после вставки
}

func (r *mongoRepo) Close() {
	r.client.Disconnect(context.Background())
}

func (r *mongoRepo) Stop() {
	// ничего не делаем, закрытие происходит при вызове Close из main
}
