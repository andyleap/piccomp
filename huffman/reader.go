package huffman

import (
	"errors"
	"io"
)

var (
	// ErrInvalidClass is returned when the class is invalid.
	ErrInvalidClass = errors.New("invalid class")
)

// Reader reads a huffman encoded stream.
type Reader struct {
	r     *bitReader
	codes map[byte]*huffTree
}

// NewReader returns a new Reader reading from r.
func NewReader(r io.ByteReader) *Reader {
	hr := &Reader{
		r:     newBitReader(r),
		codes: map[byte]*huffTree{},
	}
	hr.readCodes()
	return hr
}

// readCodes reads the huffman codes from the stream.
func (hr *Reader) readCodes() error {
	for {
		c, err := hr.r.readBits(8)
		if err != nil {
			return err
		}
		if c == 255 {
			break
		}
		tree, err := readTree(hr.r)
		if err != nil {
			return err
		}
		hr.codes[byte(c)] = tree
	}
	return nil
}

// Read reads the next token of a specific class from the stream.
func (hr *Reader) Read(class byte) (byte, error) {
	tree := hr.codes[class]
	if tree == nil {
		return 0, ErrInvalidClass
	}
	for {
		if tree.left == nil {
			return tree.value, nil
		}
		b, err := hr.r.readBit()
		if err != nil {
			return 0, err
		}
		if b == 0 {
			tree = tree.left
		} else {
			tree = tree.right
		}
	}
}
