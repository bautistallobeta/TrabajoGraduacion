package cache

import (
	"sync"
	"time"
)

type item[T any] struct {
	valor      T
	expiracion time.Time
}

// Cachein-memory genérico, thread-safe, con TTL.
type Cache[T any] struct {
	mu    sync.RWMutex
	items map[string]item[T]
	ttl   time.Duration
}

// crea un caché con el TTL indicado.
func NewCache[T any](ttl time.Duration) *Cache[T] {
	return &Cache[T]{
		items: make(map[string]item[T]),
		ttl:   ttl,
	}
}

// devuelve el valor asociado a la clave y true si existe y no expiró.
// Si no existe o expiró, devuelve el zero value de T y false.
func (c *Cache[T]) Dame(clave string) (T, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	it, existe := c.items[clave]
	if !existe || time.Now().After(it.expiracion) {
		var zero T
		return zero, false
	}
	return it.valor, true
}

func (c *Cache[T]) Guardar(clave string, valor T) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[clave] = item[T]{
		valor:      valor,
		expiracion: time.Now().Add(c.ttl),
	}
}

func (c *Cache[T]) Borrar(clave string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.items, clave)
}

// Limpia el caché entero
func (c *Cache[T]) Limpiar() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[string]item[T])
}
