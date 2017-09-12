// Package bitarray is a bare bones implementation of a fixed sized bit array.
// It achieves space efficiency by using bit shift operations on byte values.
package bitarray

import "errors"

// BitArray is an abstract representation of an array of binary values.
type BitArray interface {
	Set(uint) error
	Clear(uint) error
	Get(uint) (bool, error)
}

type bitArray struct {
	array []byte
	size  uint
}

const pageSize = 8

// New creates a new BitArray with a given size
func New(size uint) BitArray {
	s := size / pageSize
	if size%pageSize > 0 {
		s++
	}
	return &bitArray{array: make([]byte, s), size: size}
}

// Set the k'th bit in the array
func (b *bitArray) Set(k uint) error {
	if k > b.size {
		return errors.New("Index out of range")
	}
	b.array[k/pageSize] |= 1 << (k % pageSize)
	return nil
}

// Clear the k'th bit in the array
func (b *bitArray) Clear(k uint) error {
	if k > b.size {
		return errors.New("Index out of range")
	}
	b.array[k/pageSize] &= ^(1 << (k % pageSize))
	return nil
}

// Get the value of the k'th bit in the array
func (b *bitArray) Get(k uint) (bool, error) {
	if k > b.size {
		return false, errors.New("Index out of range")
	}
	return b.array[k/pageSize]&(1<<(k%pageSize)) != 0, nil
}
