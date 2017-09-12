package gomemc3_test

import (
	"sync"
	"testing"

	"github.com/wagnerbarretto/gomemc3"
)

// Ensure the cache can store new entries
func TestPut(t *testing.T) {
	c := gomemc3.New(0, nil)
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
	c := gomemc3.New(0, nil)
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
	c := gomemc3.New(3, nil)
	for i := 0; i < 5; i++ {
		c.Put(i, i)
	}

	if _, ok := c.Get(0); ok {
		t.Error("Key 0 should have been evicted")
	}
	if _, ok := c.Get(1); ok {
		t.Error("Key 1 should have been evicted")
	}

	// access key 2 so it is not the next victim
	c.Get(2)

	c.Put(5, 5)
	if _, ok := c.Get(3); ok {
		t.Error("Key 3 should have been evicted")
	}
	// key 2 must still be on th cache
	if _, ok := c.Get(2); !ok {
		t.Error("Key 2 should not have been evicted")
	}

	for i := 6; i < 9; i++ {
		c.Put(i, i)
	}

	// key 6 must have been evicted since key 2 has R = true
	if _, ok := c.Get(6); ok {
		t.Error("Key 6 should have been evicted")
	}

	// with the next Put, key 2 should finally be evicted
	c.Put(9, 9)
	if _, ok := c.Get(2); ok {
		t.Error("Key 2 should have been evicted")
	}
}

// Ensure the onEvic function is properly called
func TestOnEvict(t *testing.T) {
	ch := make(chan int, 5)
	onEvict := func(k, v interface{}) {
		ch <- k.(int)
	}
	c := gomemc3.New(5, onEvict)

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
	c := gomemc3.New(100, nil)
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
	c := gomemc3.New(b.N/10, nil)
	for i := 0; i < b.N; i++ {
		c.Put(i, i)
	}
}

func BenchmarkPutQuarter(b *testing.B) {
	c := gomemc3.New(b.N/4, nil)
	for i := 0; i < b.N; i++ {
		c.Put(i, i)
	}
}

func BenchmarkPutHalf(b *testing.B) {
	c := gomemc3.New(b.N/2, nil)
	for i := 0; i < b.N; i++ {
		c.Put(i, i)
	}
}

func BenchmarkPutNoEviction(b *testing.B) {
	c := gomemc3.New(b.N, nil)
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
	c := gomemc3.New(b.N/2, nil)
	for i := 0; i < b.N; i++ {
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
