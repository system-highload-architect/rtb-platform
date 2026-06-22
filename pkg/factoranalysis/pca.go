package factoranalysis

import (
	"errors"
	"math"

	"rtb-platform/pkg/statistics"
)

// PCA реализует метод главных компонент.
type PCA struct {
	// Компоненты: каждая строка — собственный вектор (направление компоненты).
	Components [][]float64
	// Объяснённая дисперсия для каждой компоненты.
	ExplainedVariance []float64
	// Среднее значение признаков (для центрирования).
	Mean []float64
}

// TrainPCA обучает PCA на матрице данных X (строки — образцы, столбцы — признаки).
// nComponents — желаемое число главных компонент (если 0, то равно числу признаков).
func TrainPCA(X [][]float64, nComponents int) (*PCA, error) {
	if len(X) == 0 || len(X[0]) == 0 {
		return nil, errors.New("factoranalysis: empty data")
	}
	nSamples := len(X)
	nFeatures := len(X[0])

	// 1. Центрирование данных: вычитаем среднее для каждого признака
	mean := make([]float64, nFeatures)
	for j := 0; j < nFeatures; j++ {
		col := make([]float64, nSamples)
		for i := 0; i < nSamples; i++ {
			col[i] = X[i][j]
		}
		mean[j] = statistics.Mean(col)
	}

	centered := make([][]float64, nSamples)
	for i := 0; i < nSamples; i++ {
		centered[i] = make([]float64, nFeatures)
		for j := 0; j < nFeatures; j++ {
			centered[i][j] = X[i][j] - mean[j]
		}
	}

	// 2. Ковариационная матрица (nFeatures x nFeatures)
	cov := make([][]float64, nFeatures)
	for i := 0; i < nFeatures; i++ {
		cov[i] = make([]float64, nFeatures)
		for j := 0; j < nFeatures; j++ {
			col1 := make([]float64, nSamples)
			col2 := make([]float64, nSamples)
			for k := 0; k < nSamples; k++ {
				col1[k] = centered[k][i]
				col2[k] = centered[k][j]
			}
			cov[i][j] = statistics.Covariance(col1, col2)
		}
	}

	// 3. Поиск собственных векторов и значений степенным методом
	if nComponents <= 0 || nComponents > nFeatures {
		nComponents = nFeatures
	}
	components := make([][]float64, nComponents)
	explained := make([]float64, nComponents)

	// Копируем ковариационную матрицу (будем вычитать вклад найденных компонент)
	residual := copyMatrix(cov)

	for comp := 0; comp < nComponents; comp++ {
		// Начальный случайный вектор
		eigVec := randomVector(nFeatures)
		// Степенной метод
		for iter := 0; iter < 100; iter++ {
			newVec := matVecMul(residual, eigVec)
			norm := vectorNorm(newVec)
			if norm < 1e-10 {
				break
			}
			for j := range newVec {
				newVec[j] /= norm
			}
			// Проверка сходимости (косинусное расстояние)
			if cosineDist(eigVec, newVec) < 1e-6 {
				eigVec = newVec
				break
			}
			eigVec = newVec
		}
		// Вычисляем соответствующее собственное значение: v' * cov * v
		temp := matVecMul(cov, eigVec)
		eigVal := dot(eigVec, temp)
		components[comp] = eigVec
		explained[comp] = eigVal

		// Вычитаем вклад найденной компоненты из остаточной матрицы
		// residual = residual - eigVal * (eigVec ⊗ eigVec)
		for i := 0; i < nFeatures; i++ {
			for j := 0; j < nFeatures; j++ {
				residual[i][j] -= eigVal * eigVec[i] * eigVec[j]
			}
		}
	}

	return &PCA{
		Components:        components,
		ExplainedVariance: explained,
		Mean:              mean,
	}, nil
}

// Transform проецирует данные X в пространство главных компонент.
// Возвращает матрицу размера nSamples x nComponents.
func (p *PCA) Transform(X [][]float64) [][]float64 {
	nSamples := len(X)
	if nSamples == 0 {
		return nil
	}
	nComponents := len(p.Components)
	result := make([][]float64, nSamples)
	for i := 0; i < nSamples; i++ {
		result[i] = make([]float64, nComponents)
		for j := 0; j < nComponents; j++ {
			sum := 0.0
			for k := 0; k < len(X[i]); k++ {
				sum += (X[i][k] - p.Mean[k]) * p.Components[j][k]
			}
			result[i][j] = sum
		}
	}
	return result
}

// InverseTransform восстанавливает исходные признаки из сжатого представления.
func (p *PCA) InverseTransform(Z [][]float64) [][]float64 {
	nSamples := len(Z)
	if nSamples == 0 {
		return nil
	}
	nFeatures := len(p.Mean)
	result := make([][]float64, nSamples)
	for i := 0; i < nSamples; i++ {
		result[i] = make([]float64, nFeatures)
		// Добавляем среднее
		copy(result[i], p.Mean)
		// Прибавляем вклад каждой компоненты
		for j := 0; j < len(p.Components); j++ {
			for k := 0; k < nFeatures; k++ {
				result[i][k] += Z[i][j] * p.Components[j][k]
			}
		}
	}
	return result
}

// ExplainedVarianceRatio возвращает долю объяснённой дисперсии для каждой компоненты.
func (p *PCA) ExplainedVarianceRatio() []float64 {
	total := 0.0
	for _, v := range p.ExplainedVariance {
		total += v
	}
	if total == 0 {
		return make([]float64, len(p.ExplainedVariance))
	}
	ratio := make([]float64, len(p.ExplainedVariance))
	for i, v := range p.ExplainedVariance {
		ratio[i] = v / total
	}
	return ratio
}

// ──────────── Вспомогательные функции ────────────

func copyMatrix(m [][]float64) [][]float64 {
	cp := make([][]float64, len(m))
	for i := range m {
		cp[i] = make([]float64, len(m[i]))
		copy(cp[i], m[i])
	}
	return cp
}

func matVecMul(A [][]float64, x []float64) []float64 {
	res := make([]float64, len(A))
	for i := range A {
		for j, v := range x {
			res[i] += A[i][j] * v
		}
	}
	return res
}

func vectorNorm(v []float64) float64 {
	sum := 0.0
	for _, val := range v {
		sum += val * val
	}
	return math.Sqrt(sum)
}

func dot(a, b []float64) float64 {
	res := 0.0
	for i := range a {
		res += a[i] * b[i]
	}
	return res
}

func cosineDist(a, b []float64) float64 {
	ab := dot(a, b)
	na := vectorNorm(a)
	nb := vectorNorm(b)
	if na == 0 || nb == 0 {
		return 1.0
	}
	return 1.0 - ab/(na*nb)
}

func randomVector(dim int) []float64 {
	v := make([]float64, dim)
	// Инициализируем детерминированно для воспроизводимости (на самом деле нужно seed)
	// В продакшене лучше использовать math/rand с сидом.
	for i := range v {
		v[i] = float64(i%10) / 10.0 // временная заглушка
	}
	return v
}
