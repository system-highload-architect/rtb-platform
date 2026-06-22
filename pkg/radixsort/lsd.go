package radixsort

// SortInt64 сортирует слайс int64 по возрастанию (in-place, LSD radix sort).
// Сложность O(N), без аллокаций (используется один временный слайс того же размера).
func SortInt64(data []int64) {
	n := len(data)
	if n < 2 {
		return
	}
	// Преобразуем int64 в uint64 с инвертированием знакового бита,
	// чтобы отрицательные числа шли перед положительными и сортировались корректно.
	src := make([]uint64, n)
	dst := make([]uint64, n)
	for i, v := range data {
		src[i] = uint64(v) ^ (1 << 63)
	}

	// 8 проходов по 8 бит (байтам), начиная с младшего (LSD)
	for shift := 0; shift < 64; shift += 8 {
		var count [256]int
		for _, v := range src {
			b := (v >> shift) & 0xFF
			count[b]++
		}
		// Префиксные суммы
		for i := 1; i < 256; i++ {
			count[i] += count[i-1]
		}
		// Заполняем выходной массив (стабильно)
		for i := n - 1; i >= 0; i-- {
			b := (src[i] >> shift) & 0xFF
			count[b]--
			dst[count[b]] = src[i]
		}
		// Меняем местами src и dst для следующего прохода
		src, dst = dst, src
	}

	// Восстанавливаем исходные int64
	for i, v := range src {
		data[i] = int64(v ^ (1 << 63))
	}
}
