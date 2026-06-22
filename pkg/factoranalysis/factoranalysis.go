package factoranalysis

// FactorAnalysis — интерфейс для разных методов факторного анализа.
type FactorAnalysis interface {
	Transform(X [][]float64) [][]float64
	InverseTransform(Z [][]float64) [][]float64
}
