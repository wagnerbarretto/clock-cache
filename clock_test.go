package clock_test

import (
	"sync"
	"testing"

	clock "github.com/wagnerbarretto/clock-cache"
)

// Ensure the cache can store new entries
func TestPut(t *testing.T) {
	c := clock.New(10)
	for i := 0; i < 10; i++ {
		c.Put(i, string(i))
	}
	for i := 0; i < 10; i++ {
		res, ok := c.Get(i)
		if !ok {
			t.Errorf("Get on key %v returned false, expected true", i)
		}
		if res != string(i) {
			t.Errorf("Get for key %v = %s; Expected %s", i, res, string(i))
		}
	}

	// re-insert the same keys with diferent values
	for i := 0; i < 10; i++ {
		c.Put(i, string(i*2))
	}
	for i := 0; i < 10; i++ {
		res, ok := c.Get(i)
		if !ok {
			t.Errorf("Get on key %v returned false, expected true", i)
		}
		if res != string(i*2) {
			t.Errorf("Get for key %v = %s; Expected %s", i, res, string(i*2))
		}
	}
}

// Ensure the cache can delete entries
func TestDelete(t *testing.T) {
	c := clock.New(10)
	for i := 0; i < 10; i++ {
		c.Put(i, string(i))
	}
	for i := 0; i < 10; i++ {
		c.Delete(i)
		if v, ok := c.Get(i); ok {
			t.Errorf("Key %v not deleted. Returned value %s", i, v)
		}
	}
}

// Ensure evictions correctness with a few full clock cycles
func TestEviction(t *testing.T) {
	// put 5 items in a cache with maxEntries = 3
	c := clock.New(3)
	for i := 0; i < 4; i++ {
		c.Put(i, i)
	}

	if _, ok := c.Get(0); ok {
		t.Error("Key 0 should have been evicted")
	}

	// access key 1 so it is not the next victim
	c.Get(1)

	c.Put(4, 4)
	if _, ok := c.Get(2); ok {
		t.Error("Key 2 should have been evicted")
	}
	// key 1 must still be on th cache
	if _, ok := c.Get(1); !ok {
		t.Error("Key 1 should not have been evicted")
	}
}

// Ensure the onEvic function is properly called
func TestOnEvict(t *testing.T) {
	ch := make(chan int, 5)
	onEvict := func(k, v interface{}) {
		ch <- k.(int)
	}
	c := clock.New(5, clock.OnEvict(onEvict))

	for i := 0; i < 10; i++ {
		c.Put(i, i)
	}

	evicted := make(map[int]bool)
	for i := 0; i < 5; i++ {
		evicted[<-ch] = true
	}

	for i := 0; i < 5; i++ {
		if _, ok := evicted[i]; !ok {
			t.Errorf("onEvict not called for key %v", i)
		}
	}

}

// Ensure the cache behave correctly under concurrent access
func TestConcurrency(t *testing.T) {
	c := clock.New(100)
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c.Put(i, i)
		}(i)
	}
	wg.Wait()

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if _, ok := c.Get(i); !ok {
				t.Fail()
			}
		}(i)
	}
	wg.Wait()
	if t.Failed() {
		t.Fatal("Failed to read the values from 0 to 99")
	}

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c.Put(i+100, i)
			if _, ok := c.Get(i + 50); !ok {
				t.Fail()
			}
		}(i)
	}
	wg.Wait()
	if t.Failed() {
		t.Fatal("Failed to read the values from 50 to 99 whlie executing other inserts")
	}
}

func BenchmarkMapSet(b *testing.B) {
	m := make(map[int]int)
	for i := 0; i < b.N; i++ {
		m[i] = i
	}
}

func BenchmarkPutTenth(b *testing.B) {
	size := uint(b.N / 10)
	if size < 1 {
		size++
	}
	c := clock.New(size)
	for i := 0; i < b.N; i++ {
		c.Put(i, i)
	}
}

func BenchmarkPutQuarter(b *testing.B) {
	size := uint(b.N / 4)
	if size < 1 {
		size++
	}
	c := clock.New(size)
	for i := 0; i < b.N; i++ {
		c.Put(i, i)
	}
}

func BenchmarkPutHalf(b *testing.B) {
	size := uint(b.N / 2)
	if size < 1 {
		size++
	}
	c := clock.New(size)
	for i := 0; i < b.N; i++ {
		c.Put(i, i)
	}
}

func BenchmarkPutNoEviction(b *testing.B) {
	size := uint(b.N / 2)
	if size < 1 {
		size++
	}
	c := clock.New(size)
	for i := 0; i < b.N; i++ {
		c.Put(i, i)
	}
}

func BenchmarkMapGet(b *testing.B) {
	b.StopTimer()
	m := make(map[interface{}]interface{})
	for i := 0; i < b.N; i++ {
		m[i] = i
	}

	b.StartTimer()
	hitCount := 0
	for i := 0; i < b.N; i++ {
		if _, ok := m[i]; ok {
			hitCount++
		}
	}
}

func BenchmarkGet(b *testing.B) {
	b.StopTimer()
	size := uint(b.N / 2)
	if size < 1 {
		size++
	}
	c := clock.New(size)
	for i := uint(0); i < size; i++ {
		c.Put(i, i)
	}

	b.StartTimer()
	hitCount := 0
	for i := 0; i < b.N; i++ {
		if _, ok := c.Get(i); ok {
			hitCount++
		}
	}

}
