package timedcache

import "time"

// Option настраивает Cache.
type Option[K comparable, V any] func(*Cache[K, V])

// WithFinalizer устанавливает функцию, которая вызывается при истечении элемента.
// Вызов происходит в отдельной горутине из пула финализаторов.
func WithFinalizer[K comparable, V any](fn func(key K, value V)) Option[K, V] {
	return func(c *Cache[K, V]) {
		c.finalizer = fn
	}
}

// WithFinalizerWorkers задаёт количество горутин-финализаторов (по умолчанию 4).
func WithFinalizerWorkers[K comparable, V any](n int) Option[K, V] {
	return func(c *Cache[K, V]) {
		if n < 1 {
			n = 1
		}
		c.finalizerWorkers = n
	}
}

// WithFinalizerBuffer задаёт размер буфера канала финализации (по умолчанию 256).
func WithFinalizerBuffer[K comparable, V any](size int) Option[K, V] {
	return func(c *Cache[K, V]) {
		if size < 1 {
			size = 1
		}
		c.finalizerBuf = size
	}
}

// WithNowFunc позволяет подменить источник времени (для тестов).
func WithNowFunc[K comparable, V any](fn func() time.Time) Option[K, V] {
	return func(c *Cache[K, V]) {
		c.now = fn
	}
}
