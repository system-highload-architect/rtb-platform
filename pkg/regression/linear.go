package regression

import (
	"errors"
	"math"
)

// LinearModel содержит обученные коэффициенты линейной регрессии.
// Первый коэффициент (индекс 0) — это intercept (свободный член).
type LinearModel struct {
	Coefficients []float64
}

// Predict вычисляет предсказанное значение по вектору признаков.
// features должен быть длины len(model.Coefficients)-1.
// Аллокаций не делает.
func (m *LinearModel) Predict(features []float64) float64 {
	if len(features) != len(m.Coefficients)-1 {
		return 0 // или паника, но в продакшене лучше ошибка
	}
	result := m.Coefficients[0] // intercept
	for i, v := range features {
		result += m.Coefficients[i+1] * v
	}
	return result
}

// TrainLinear обучает линейную регрессию методом наименьших квадратов
// через нормальное уравнение (X'X)^-1 X'y.
// x — матрица признаков (каждый внутренний слайс — один образец),
// y — целевые значения. Добавляет столбец единиц для intercept.
func TrainLinear(x [][]float64, y []float64) (*LinearModel, error) {
	if len(x) == 0 || len(x) != len(y) {
		return nil, errors.New("regression: empty or mismatched data")
	}
	nSamples := len(x)
	if nSamples == 0 {
		return nil, errors.New("regression: no samples")
	}
	nFeatures := len(x[0])
	// Добавим intercept как дополнительный признак = 1
	augmented := make([][]float64, nSamples)
	for i := range augmented {
		augmented[i] = make([]float64, nFeatures+1)
		augmented[i][0] = 1.0
		copy(augmented[i][1:], x[i])
	}

	// X'X и X'y
	xtx := make([][]float64, nFeatures+1)
	xty := make([]float64, nFeatures+1)
	for i := 0; i < nFeatures+1; i++ {
		xtx[i] = make([]float64, nFeatures+1)
		for k := 0; k < nSamples; k++ {
			xty[i] += augmented[k][i] * y[k]
			for j := 0; j < nFeatures+1; j++ {
				xtx[i][j] += augmented[k][i] * augmented[k][j]
			}
		}
	}

	// Решаем систему xtx * coeff = xty через обратную матрицу
	// (для малых размерностей допустимо, но для продакшена лучше использовать QR-разложение)
	inv, err := invertMatrix(xtx)
	if err != nil {
		// fallback: градиентный спуск (не реализован, пока ошибка)
		return nil, errors.New("regression: matrix inversion failed, try gradient descent")
	}

	coeff := make([]float64, nFeatures+1)
	for i := 0; i < nFeatures+1; i++ {
		for j := 0; j < nFeatures+1; j++ {
			coeff[i] += inv[i][j] * xty[j]
		}
	}

	return &LinearModel{Coefficients: coeff}, nil
}

// invertMatrix обращает квадратную матрицу методом Гаусса-Жордана.
// Возвращает ошибку, если матрица вырождена.
func invertMatrix(m [][]float64) ([][]float64, error) {
	n := len(m)
	// Создаем расширенную матрицу [m | I]
	a := make([][]float64, n)
	for i := range a {
		a[i] = make([]float64, 2*n)
		copy(a[i], m[i])
		a[i][n+i] = 1.0
	}

	for i := 0; i < n; i++ {
		// Ищем главный элемент
		pivot := i
		for j := i + 1; j < n; j++ {
			if math.Abs(a[j][i]) > math.Abs(a[pivot][i]) {
				pivot = j
			}
		}
		if math.Abs(a[pivot][i]) < 1e-10 {
			return nil, errors.New("regression: singular matrix")
		}
		// Меняем строки
		a[i], a[pivot] = a[pivot], a[i]

		// Нормируем строку
		div := a[i][i]
		for j := 0; j < 2*n; j++ {
			a[i][j] /= div
		}

		// Вычитаем из остальных строк
		for k := 0; k < n; k++ {
			if k == i {
				continue
			}
			factor := a[k][i]
			for j := 0; j < 2*n; j++ {
				a[k][j] -= factor * a[i][j]
			}
		}
	}

	// Извлекаем обратную матрицу
	inv := make([][]float64, n)
	for i := range inv {
		inv[i] = make([]float64, n)
		copy(inv[i], a[i][n:])
	}
	return inv, nil
}
