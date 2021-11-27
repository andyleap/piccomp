package huffman

import "io"

// bitWriter is a writer that writes bits to an underlying io.writer.
type bitWriter struct {
	w   io.Writer
	buf byte
	n   int
}

// newBitWriter returns a new bitWriter that writes to w.
func newBitWriter(w io.Writer) *bitWriter {
	return &bitWriter{w: w}
}

// writeBit writes a single bit to the underlying io.Writer.
func (bw *bitWriter) writeBit(bit bool) error {
	if bw.n == 8 {
		if err := bw.flush(); err != nil {
			return err
		}
	}
	if bit {
		bw.buf |= 1 << uint(7-bw.n)
	}
	bw.n++
	return nil
}

// writeBits writes n bits from the given int to the underlying io.Writer.
func (bw *bitWriter) writeBits(x int, n uint) error {
	for i := uint(0); i < n; i++ {
		if err := bw.writeBit((x>>(n-1-i))&1 == 1); err != nil {
			return err
		}
	}
	return nil
}

// flush flushes the current byte to the underlying io.Writer.
func (bw *bitWriter) flush() error {
	if bw.n == 0 {
		return nil
	}
	if _, err := bw.w.Write([]byte{bw.buf}); err != nil {
		return err
	}
	bw.buf = 0
	bw.n = 0
	return nil
}
