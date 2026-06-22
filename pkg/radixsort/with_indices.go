package radixsort

// SortInt64WithIndices сортирует ставки и параллельно переставляет индексы.
// data — ставки (будут отсортированы по возрастанию),
// indices — соответствующие индексы кампаний (переставляются синхронно).
func SortInt64WithIndices(data []int64, indices []int) {
	if len(data) != len(indices) || len(data) < 2 {
		return
	}
	n := len(data)
	// Преобразуем int64 -> uint64
	src := make([]uint64, n)
	dst := make([]uint64, n)
	for i, v := range data {
		src[i] = uint64(v) ^ (1 << 63)
	}

	srcIdx := make([]int, n)
	dstIdx := make([]int, n)
	copy(srcIdx, indices)

	for shift := 0; shift < 64; shift += 8 {
		var count [256]int
		for _, v := range src {
			b := (v >> shift) & 0xFF
			count[b]++
		}
		for i := 1; i < 256; i++ {
			count[i] += count[i-1]
		}
		// Проходим с конца для стабильности
		for i := n - 1; i >= 0; i-- {
			b := (src[i] >> shift) & 0xFF
			count[b]--
			dst[count[b]] = src[i]
			dstIdx[count[b]] = srcIdx[i]
		}
		src, dst = dst, src
		srcIdx, dstIdx = dstIdx, srcIdx
	}

	for i, v := range src {
		data[i] = int64(v ^ (1 << 63))
	}
	copy(indices, srcIdx)
}
