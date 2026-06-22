package valuation

import "rtb-platform/pkg/regression"

// ImpressionValue оценивает ценность одного показа (в условных единицах).
// Использует логистические модели для pCTR и pCVR.
type ImpressionValue struct {
	pCTRModel *regression.LogisticModel
	pCVRModel *regression.LogisticModel
	// ConversionValue — стоимость конверсии в fixedpoint (заполняется при создании)
	ConversionValue float64
}

// NewImpressionValue создаёт оценщик ценности показа.
// pCTR и pCVR — предобученные коэффициенты логистической регрессии.
func NewImpressionValue(pCTRCoeff, pCVRCoeff []float64, convValue float64) *ImpressionValue {
	return &ImpressionValue{
		pCTRModel:       &regression.LogisticModel{Coefficients: pCTRCoeff},
		pCVRModel:       &regression.LogisticModel{Coefficients: pCVRCoeff},
		ConversionValue: convValue,
	}
}

// Value вычисляет ценность показа для пары пользователь/креатив.
// userFeatures — признаки пользователя, adFeatures — признаки рекламного креатива.
func (iv *ImpressionValue) Value(userFeatures, adFeatures []float64) float64 {
	pCTR := iv.pCTRModel.PredictProb(append(userFeatures, adFeatures...))
	pCVR := iv.pCVRModel.PredictProb(append(userFeatures, adFeatures...))
	return pCTR * pCVR * iv.ConversionValue
}
