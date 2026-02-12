package resettablepoll

type Resettable[T any] interface {
	Reset()
}

type Poll[T Resettable[T]] struct {
	items []T
}

func New[T Resettable[T]]() *Poll[T] {
	return &Poll[T]{
		items: make([]T, 0),
	}
}

func (p *Poll[T]) Get() Resettable[T] {
	lastIdx := len(p.items) - 1

	if lastIdx < 0 {
		return nil
	}
	lastItem := p.items[lastIdx]
	p.items = p.items[:lastIdx]

	return lastItem
}

func (p *Poll[T]) Put(item T) {
	item.Reset()

	p.items = append(p.items, item)
}
