package valuation

import "rtb-platform/pkg/regression"

// LTVModel предсказывает пожизненную ценность пользователя.
// Использует линейную регрессию.
type LTVModel struct {
	model *regression.LinearModel
}

// NewLTVModel создаёт модель из предобученных коэффициентов.
// Первый коэффициент — intercept.
func NewLTVModel(coefficients []float64) *LTVModel {
	return &LTVModel{
		model: &regression.LinearModel{Coefficients: coefficients},
	}
}

// Predict возвращает прогноз LTV для вектора признаков пользователя.
func (m *LTVModel) Predict(features []float64) float64 {
	return m.model.Predict(features)
}
