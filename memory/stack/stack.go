package stack

type Stack[T any] struct {
	items []T
}

func (stack *Stack[T]) ToSlice() []T {
	return stack.items
}

func (stack *Stack[T]) Set(values []T) {
	stack.items = values
}

func (stack *Stack[T]) IsEmpty() bool {
	return len(stack.items) == 0
}

func (stack *Stack[T]) Peek() T {
	n := stack.Len()
	if n <= 0 {
		var defaultValue T
		return defaultValue
	}

	return stack.items[n-1]
}

func (stack *Stack[T]) Reverse() *Stack[T] {
	for i, j := 0, len(stack.items)-1; i < j; i, j = i+1, j-1 {
		stack.items[i], stack.items[j] = stack.items[j], stack.items[i]
	}

	return stack
}

func (stack *Stack[T]) Push(value T) {
	stack.items = append(stack.items, value)
}

func (stack *Stack[T]) PushLeft(value T) {
	stack.items = append([]T{value}, stack.items...)
}

func (stack *Stack[T]) Pop() T {
	n := stack.Len()
	if n <= 0 {
		var defaultValue T
		return defaultValue
	}

	p := stack.Peek()
	stack.items = stack.items[:n-1]

	return p
}

func (stack *Stack[T]) PopLeft() T {
	var item T

	n := len(stack.items)
	if n <= 0 {
		return item
	}

	item = stack.items[0]
	stack.items = stack.items[1:]

	return item
}

func (stack *Stack[T]) Len() int {
	return len(stack.items)
}
