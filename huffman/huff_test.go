package huffman

import (
	"bytes"
	"testing"
)

func TestHuffman(t *testing.T) {
	hw := NewWriter()
	hw.Write(0, 1)
	hw.Write(0, 2)
	hw.Write(0, 3)
	hw.Write(0, 4)
	hw.Write(0, 1)
	hw.Write(0, 1)
	hw.Write(0, 1)
	buf := &bytes.Buffer{}
	err := hw.Dump(buf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestLoop(t *testing.T) {
	data := []byte{
		0, 1, 2, 3, 4, 1, 1, 1, 140, 0, 12, 141,
	}
	hw := NewWriter()
	for _, v := range data {
		hw.Write(0, v)
	}
	buf := &bytes.Buffer{}
	err := hw.Dump(buf)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(buf.Bytes())

	bufr := bytes.NewReader(buf.Bytes())
	hr := NewReader(bufr)
	for v := range data {
		r, err := hr.Read(0)
		if err != nil {
			t.Fatal(err)
		}
		if r != data[v] {
			t.Fatalf("%d != %d", r, data[v])
		}
	}
}
