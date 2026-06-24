package domain

import (
	"context"
	"rtb-platform/pkg/factoranalysis"
)

type FactorService struct{}

func NewFactorService() *FactorService {
	return &FactorService{}
}

func (s *FactorService) Analyse(ctx context.Context, data [][]float64, nComponents int) (*factoranalysis.PCA, error) {
	return factoranalysis.TrainPCA(data, nComponents)
}
