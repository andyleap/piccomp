package huffman

import "io"

type bitReader struct {
	r   io.ByteReader
	buf byte
	n   int
}

// newBitReader returns a new bitReader that reads from r.
func newBitReader(r io.ByteReader) *bitReader {
	return &bitReader{r: r}
}

// readBit reads the next bit from the input source.
func (br *bitReader) readBit() (byte, error) {
	if br.n == 0 {
		var err error
		br.buf, err = br.r.ReadByte()
		if err != nil {
			return 0, err
		}
		br.n = 8
	}
	br.n--
	return byte(br.buf>>br.n) & 1, nil
}

// readBits reads n bits from the input source.
func (br *bitReader) readBits(n int) (int, error) {
	var r int
	for i := 0; i < n; i++ {
		b, err := br.readBit()
		if err != nil {
			return 0, err
		}
		r = r<<1 | int(b)
	}
	return r, nil
}
