package timedcache

import (
	"container/list"
	"sync"
	"time"
)

// entry — внутренний элемент двусвязного списка.
type entry[K comparable, V any] struct {
	key       K
	value     V
	expiresAt time.Time
}

// Cache — потокобезопасный кэш с фиксированным TTL и упорядоченным списком по expiresAt.
// Хвост списка (Back) содержит элемент с ближайшим сроком истечения.
type Cache[K comparable, V any] struct {
	mu               sync.Mutex
	ttl              time.Duration
	items            map[K]*list.Element
	evictList        *list.List
	finalizer        func(key K, value V)
	finalizerWorkers int
	finalizerBuf     int
	finalizeCh       chan *entry[K, V]
	now              func() time.Time
	stopCh           chan struct{}
	stopped          bool
	wakeCh           chan struct{} // сигнал демону: появился новый элемент
}

// New создаёт кэш с заданным TTL.
// Параметр ttl должен быть положительным.
func New[K comparable, V any](ttl time.Duration, opts ...Option[K, V]) *Cache[K, V] {
	c := &Cache[K, V]{
		ttl:              ttl,
		items:            make(map[K]*list.Element),
		evictList:        list.New(),
		finalizerWorkers: 4,
		finalizerBuf:     256,
		now:              time.Now,
		stopCh:           make(chan struct{}),
		wakeCh:           make(chan struct{}, 1), // буфер 1, чтобы не блокировать отправителя
	}
	for _, opt := range opts {
		opt(c)
	}
	if c.finalizer != nil {
		c.finalizeCh = make(chan *entry[K, V], c.finalizerBuf)
		for i := 0; i < c.finalizerWorkers; i++ {
			go c.finalizeWorker()
		}
	}
	go c.daemon()
	return c
}

// Get возвращает значение по ключу, если оно не просрочено.
// При успешном доступе обновляет expiresAt и перемещает элемент в голову.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	elem, ok := c.items[key]
	if !ok {
		var zero V
		return zero, false
	}
	ent := elem.Value.(*entry[K, V])
	now := c.now()
	if now.After(ent.expiresAt) {
		// Элемент уже просрочен, но ещё не удалён демоном. Удаляем сейчас.
		c.removeElement(elem)
		var zero V
		return zero, false
	}
	// Обновляем срок и перемещаем в голову (новый expiresAt максимальный)
	ent.expiresAt = now.Add(c.ttl)
	c.evictList.MoveToFront(elem)
	return ent.value, true
}

// Set добавляет или обновляет значение. Элемент помещается в голову списка.
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := c.now()
	if elem, ok := c.items[key]; ok {
		// Обновление существующего
		ent := elem.Value.(*entry[K, V])
		ent.value = value
		ent.expiresAt = now.Add(c.ttl)
		c.evictList.MoveToFront(elem)
		return
	}
	// Новый элемент
	ent := &entry[K, V]{key: key, value: value, expiresAt: now.Add(c.ttl)}
	elem := c.evictList.PushFront(ent)
	c.items[key] = elem

	if c.evictList.Len() == 1 {
		c.wakeUp()
	}
}

// Extend продлевает время жизни элемента, если он существует и не просрочен.
func (c *Cache[K, V]) Extend(key K) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	elem, ok := c.items[key]
	if !ok {
		return false
	}
	ent, ok := elem.Value.(*entry[K, V])
	if !ok {
		return false
	}
	now := c.now()
	if now.After(ent.expiresAt) {
		c.removeElement(elem)
		return false
	}
	ent.expiresAt = now.Add(c.ttl)
	c.evictList.MoveToFront(elem)
	return true
}

// Delete удаляет элемент по ключу, если он существует.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if elem, ok := c.items[key]; ok {
		c.removeElement(elem)
	}
}

// Stop останавливает демона и закрывает канал финализации.
// После вызова Stop кэш больше нельзя использовать.
func (c *Cache[K, V]) Stop() {
	c.mu.Lock()
	if c.stopped {
		c.mu.Unlock()
		return
	}
	c.stopped = true
	close(c.stopCh)
	if c.finalizeCh != nil {
		close(c.finalizeCh)
	}
	c.mu.Unlock()
}

// removeElement удаляет элемент из списка и карты (должен вызываться под мьютексом).
func (c *Cache[K, V]) removeElement(elem *list.Element) {
	ent := elem.Value.(*entry[K, V])
	delete(c.items, ent.key)
	c.evictList.Remove(elem)
}

func (c *Cache[K, V]) wakeUp() {
	select {
	case c.wakeCh <- struct{}{}:
	default:
		// уже есть сигнал, демон и так проснётся
	}
}

// finalizeWorker обрабатывает элементы, требующие финализации.
func (c *Cache[K, V]) finalizeWorker() {
	for ent := range c.finalizeCh {
		if c.finalizer != nil {
			c.finalizer(ent.key, ent.value)
		}
	}
}

// Values возвращает срез всех значений, находящихся в кэше.
// Порядок не гарантирован.
func (c *Cache[K, V]) Values() []V {
	c.mu.Lock()
	defer c.mu.Unlock()
	values := make([]V, 0, len(c.items))
	for _, elem := range c.items {
		ent, ok := elem.Value.(*entry[K, V])
		if ok {
			values = append(values, ent.value)
		}
	}
	return values
}
