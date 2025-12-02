package hw04lrucache

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCache(t *testing.T) {
	t.Run("empty cache", func(t *testing.T) {
		c := NewCache(10)

		_, ok := c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("bbb")
		require.False(t, ok)
	})

	t.Run("simple", func(t *testing.T) {
		c := NewCache(5)

		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		val, ok := c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 100, val)

		val, ok = c.Get("bbb")
		require.True(t, ok)
		require.Equal(t, 200, val)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")
		require.False(t, ok)
		require.Nil(t, val)
	})

	t.Run("purge logic", func(t *testing.T) {
		c := NewCache(3)
		wasInCache := c.Set("aaa", 100)
		require.False(t, wasInCache)

		wasInCache = c.Set("bbb", 200)
		require.False(t, wasInCache)

		wasInCache = c.Set("aaa", 300)
		require.True(t, wasInCache)

		wasInCache = c.Set("ccc", 400)
		require.False(t, wasInCache)

		wasInCache = c.Set("ddd", 200)
		require.False(t, wasInCache)

		val, ok := c.Get("bbb")
		require.False(t, ok)
		require.Equal(t, nil, val)

		val, ok = c.Get("aaa")
		require.True(t, ok)
		require.Equal(t, 300, val)

		val, ok = c.Get("ccc")
		require.True(t, ok)
		require.Equal(t, 400, val)

		c.Clear()
		_, ok = c.Get("aaa")
		require.False(t, ok)

		_, ok = c.Get("ccc")
		require.False(t, ok)

		_, ok = c.Get("ddd")
		require.False(t, ok)
	})
}

func TestCacheMultithreading(t *testing.T) {
	c := NewCache(10)
	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			c.Set(Key(strconv.Itoa(i)), i)
		}
	}()

	go func() {
		defer wg.Done()
		for i := 0; i < 1_000_000; i++ {
			r := rand.Intn(1_000_000)
			k := Key(strconv.Itoa(r))
			v, ok := c.Get(k)
			if ok {
				require.Equal(t, r, v.(int))
			}
		}
	}()

	wg.Wait()
	_, ok := c.Get("999999")
	require.True(t, ok)

	_, ok = c.Get("999998")
	require.True(t, ok)

	_, ok = c.Get("999997")
	require.True(t, ok)

	_, ok = c.Get("999996")
	require.True(t, ok)

	_, ok = c.Get("999995")
	require.True(t, ok)

	_, ok = c.Get("9999989")
	require.False(t, ok)
}
