package regression

import (
	"errors"
	"math"
)

// LogisticModel — модель бинарной логистической регрессии.
type LogisticModel struct {
	Coefficients []float64
}

// PredictProb возвращает вероятность принадлежности к классу 1 (сигмоида).
func (m *LogisticModel) PredictProb(features []float64) float64 {
	if len(features) != len(m.Coefficients)-1 {
		return 0.5 // fallback
	}
	z := m.Coefficients[0] // intercept
	for i, v := range features {
		z += m.Coefficients[i+1] * v
	}
	return 1.0 / (1.0 + math.Exp(-z))
}

// TrainLogistic обучает модель логистической регрессии градиентным спуском.
// x — матрица признаков, y — целевые значения (0 или 1).
// learningRate — скорость обучения, epochs — количество эпох.
// Возвращает обученную модель.
func TrainLogistic(x [][]float64, y []float64, learningRate float64, epochs int) (*LogisticModel, error) {
	if len(x) == 0 || len(x) != len(y) {
		return nil, errors.New("regression: empty or mismatched data")
	}
	nSamples := len(x)
	if nSamples == 0 {
		return nil, errors.New("regression: no samples")
	}
	nFeatures := len(x[0])
	// Инициализируем коэффициенты (включая intercept)
	coeff := make([]float64, nFeatures+1)
	// Добавим фиктивный признак для интерсепта (столбец единиц)
	augmented := make([][]float64, nSamples)
	for i := range augmented {
		augmented[i] = make([]float64, nFeatures+1)
		augmented[i][0] = 1.0
		copy(augmented[i][1:], x[i])
	}

	for epoch := 0; epoch < epochs; epoch++ {
		grad := make([]float64, nFeatures+1)
		for i := 0; i < nSamples; i++ {
			z := 0.0
			for j := 0; j < nFeatures+1; j++ {
				z += coeff[j] * augmented[i][j]
			}
			pred := 1.0 / (1.0 + math.Exp(-z))
			err := pred - y[i]
			for j := 0; j < nFeatures+1; j++ {
				grad[j] += err * augmented[i][j]
			}
		}
		// Обновляем коэффициенты
		for j := 0; j < nFeatures+1; j++ {
			coeff[j] -= learningRate * grad[j] / float64(nSamples)
		}
	}
	return &LogisticModel{Coefficients: coeff}, nil
}
