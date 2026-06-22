package registry

import (
	"context"
	"fmt"
)

// Handler — функция-обработчик для конкретного ключа.
type Handler[Req, Resp any] func(ctx context.Context, req Req) (Resp, error)

// Registry хранит отображение ключ → обработчик.
// K — тип ключа, должен быть comparable.
// Регистрация не потокобезопасна: предполагается, что все Register выполняются
// до начала вызовов Dispatch (например, на этапе инициализации сервиса).
type Registry[K comparable, Req, Resp any] struct {
	handlers map[K]Handler[Req, Resp]
}

// New создаёт новый пустой реестр.
func New[K comparable, Req, Resp any]() *Registry[K, Req, Resp] {
	return &Registry[K, Req, Resp]{
		handlers: make(map[K]Handler[Req, Resp]),
	}
}

// Register добавляет обработчик с указанным ключом.
// Если ключ уже зарегистрирован, заменяет его.
func (r *Registry[K, Req, Resp]) Register(key K, h Handler[Req, Resp]) {
	r.handlers[key] = h
}

// Dispatch ищет обработчик по ключу и вызывает его.
// Возвращает ErrNotFound, если ключ отсутствует.
func (r *Registry[K, Req, Resp]) Dispatch(ctx context.Context, key K, req Req) (Resp, error) {
	h, ok := r.handlers[key]
	if !ok {
		var zero Resp
		return zero, fmt.Errorf("registry: handler %v: %w", key, ErrNotFound)
	}
	return h(ctx, req)
}

// Exists проверяет, зарегистрирован ли ключ.
func (r *Registry[K, Req, Resp]) Exists(key K) bool {
	_, ok := r.handlers[key]
	return ok
}

// ErrNotFound возвращается при отсутствии обработчика.
var ErrNotFound = fmt.Errorf("handler not found")
