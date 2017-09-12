package bitarray_test

import (
	"testing"

	"github.com/wagnerbarretto/gomemc3/bitarray"
)

func TestCreation(t *testing.T) {
	b := bitarray.New(100)
	for i := uint(0); i < 100; i++ {
		if ok, _ := b.Get(i); ok {
			t.Errorf("Bit %v should be cleared", i)
		}
	}
}

func TestSet(t *testing.T) {
	b := bitarray.New(100)

	for i := uint(0); i < 100; i += 2 {
		b.Set(i)
	}

	for i := uint(0); i < 100; i++ {
		if ok, _ := b.Get(i); i%2 == 0 && !ok {
			t.Errorf("Bit %v should be set", i)
		} else if i%2 != 0 && ok {
			t.Errorf("Bit %v should be cleared", i)
		}
	}
}

func TestClear(t *testing.T) {
	b := bitarray.New(100)

	for i := uint(0); i < 100; i += 2 {
		b.Set(i)
	}

	for i := uint(0); i < 100; i++ {
		b.Clear(i)
		if ok, _ := b.Get(i); ok {
			t.Errorf("Bit %v should be cleared", i)
		}
	}
}

func TestOutOfRange(t *testing.T) {
	b := bitarray.New(1)
	if err := b.Set(2); err == nil {
		t.Error("Out of range Set should return an error")
	}

	if err := b.Clear(3); err == nil {
		t.Error("Out of range Clear should return an error")
	}

	if _, err := b.Get(4); err == nil {
		t.Error("Out of range Get should return an error")
	}

}

func BenchmarkSet(b *testing.B) {
	uN := uint(b.N)
	ba := bitarray.New(uN)
	b.ResetTimer()
	for i := uint(0); i < uN; i++ {
		ba.Set(i)
	}
}

func BenchmarkGet(b *testing.B) {
	uN := uint(b.N)
	ba := bitarray.New(uN)
	b.ResetTimer()
	for i := uint(0); i < uN; i++ {
		ba.Get(i)
	}
}
