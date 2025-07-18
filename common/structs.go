package common

type Collector[T comparable] struct {
	items []T
}

func (c *Collector[T]) AddIfNotZero(item T) {
	var zero T
	if item != zero {
		c.items = append(c.items, item)
	}
}

func (c *Collector[T]) Add(item T) {
	c.items = append(c.items, item)
}

func (c *Collector[T]) AddMany(items []T) {
	for _, item := range items {
		c.Add(item)
	}
}

func (c *Collector[T]) Items() []T {
	return c.items
}

func (c *Collector[T]) HasAny() bool {
	return len(c.items) > 0
}
