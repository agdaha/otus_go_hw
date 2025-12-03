package hw04lrucache

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	t.Run("empty list", func(t *testing.T) {
		l := NewList()

		require.Equal(t, 0, l.Len())
		require.Nil(t, l.Front())
		require.Nil(t, l.Back())
	})

	t.Run("complex", func(t *testing.T) {
		l := NewList()

		l.PushFront(10) // [10]
		l.PushBack(20)  // [10, 20]
		l.PushBack(30)  // [10, 20, 30]
		require.Equal(t, 3, l.Len())

		middle := l.Front().Next // 20
		l.Remove(middle)         // [10, 30]
		require.Equal(t, 2, l.Len())

		for i, v := range [...]int{40, 50, 60, 70, 80} {
			if i%2 == 0 {
				l.PushFront(v)
			} else {
				l.PushBack(v)
			}
		} // [80, 60, 40, 10, 30, 50, 70]

		require.Equal(t, 7, l.Len())
		require.Equal(t, 80, l.Front().Value)
		require.Equal(t, 70, l.Back().Value)

		l.MoveToFront(l.Front()) // [80, 60, 40, 10, 30, 50, 70]
		l.MoveToFront(l.Back())  // [70, 80, 60, 40, 10, 30, 50]

		elems := make([]int, 0, l.Len())
		for i := l.Front(); i != nil; i = i.Next {
			elems = append(elems, i.Value.(int))
		}
		require.Equal(t, []int{70, 80, 60, 40, 10, 30, 50}, elems)
	})

	t.Run("front&back equal for single item list", func(t *testing.T) {
		l := NewList()
		node := l.PushFront("10")
		require.Same(t, node, l.Front())
		require.Same(t, node, l.Back())
	})

	t.Run("front&back after 2 PushFront", func(t *testing.T) {
		l := NewList()
		n1 := l.PushFront("10")
		n2 := l.PushFront("20")
		require.Same(t, n2, l.Front())
		require.Same(t, n1, l.Back())
		require.Same(t, l.Back().Prev, l.Front())
		require.Same(t, l.Front().Next, l.Back())
	})

	t.Run("front&back after 2 PushBack", func(t *testing.T) {
		l := NewList()
		n1 := l.PushBack("30")
		n2 := l.PushBack("40")
		require.Same(t, n1, l.Front())
		require.Same(t, n2, l.Back())
		require.Same(t, l.Back().Prev, l.Front())
		require.Same(t, l.Front().Next, l.Back())
	})

	t.Run("remove front", func(t *testing.T) {
		l := NewList()
		n1 := l.PushFront("10")
		n2 := l.PushFront("20")

		l.Remove(n2)
		require.Same(t, n1, l.Front())
		require.Same(t, n1, l.Back())
		require.Nil(t, n1.Prev)
		require.Nil(t, n1.Next)
		require.Equal(t, 1, l.Len())
	})

	t.Run("remove back", func(t *testing.T) {
		l := NewList()
		l.PushFront(1)
		n2 := l.PushFront("30")
		l.Remove(n2)

		n1 := l.Back()
		require.Equal(t, 1, n1.Value)
		require.Same(t, n1, l.Front())
		require.Equal(t, 1, l.Len())
	})
}
