package timedcache

import (
	"time"
)

func (c *Cache[K, V]) daemon() {
	for {
		now := c.now()
		c.mu.Lock()
		tail := c.evictList.Back()
		if tail != nil {
			ent := tail.Value.(*entry[K, V])
			var sleepDuration time.Duration
			if now.After(ent.expiresAt) {
				sleepDuration = 0
			} else {
				sleepDuration = ent.expiresAt.Sub(now)
			}
			c.mu.Unlock()

			timer := time.NewTimer(sleepDuration)
			select {
			case <-c.stopCh:
				timer.Stop()
				return
			case <-c.wakeCh:
				timer.Stop()
				// Хвост мог измениться, пересчитаем в следующей итерации
			case <-timer.C:
				// Время истекло, удалим просроченные
				c.purgeExpired()
			}
		} else {
			// Список пуст, ждём сигнала или остановки
			c.mu.Unlock()
			select {
			case <-c.stopCh:
				return
			case <-c.wakeCh:
				// появился элемент, продолжаем цикл
			}
		}
	}
}

// purgeExpired каскадно удаляет все элементы из хвоста, у которых истёк срок.
func (c *Cache[K, V]) purgeExpired() {
	now := c.now()
	c.mu.Lock()
	defer c.mu.Unlock()
	for {
		tail := c.evictList.Back()
		if tail == nil {
			break
		}
		ent := tail.Value.(*entry[K, V])
		if now.After(ent.expiresAt) {
			c.removeElement(tail)
			// Отправляем в канал финализации (если задан)
			if c.finalizeCh != nil {
				select {
				case c.finalizeCh <- ent:
				default:
					// Канал переполнен — пропускаем, чтобы не блокировать удаление
				}
			}
		} else {
			break // все оставшиеся ещё живы
		}
	}
}
