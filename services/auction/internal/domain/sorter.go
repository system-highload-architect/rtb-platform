package domain

import "rtb-platform/pkg/radixsort"

// SortCandidatesByEffectiveBid сортирует срез эффективных ставок и синхронно переставляет индексы.
// После сортировки первый элемент соответствует минимальной ставке, последний – максимальной.
func SortCandidatesByEffectiveBid(bids []int64, indices []int) {
	radixsort.SortInt64WithIndices(bids, indices)
}
