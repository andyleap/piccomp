package piccomp

import (
	"bufio"
	"bytes"
	"image"
	"image/color"
	"io"
	"math/bits"

	"github.com/andyleap/piccomp/huffman"
)

type tag uint8

const (
	Delta tag = iota
	DeltaUp
	Lookup
	Run
)

const (
	ClassTag byte = iota
	ClassBCount
	ClassLookup
	ClassDelta // R
	_          // G
	_          // B
	_          // A
	ClassRun
)

func Save(w io.Writer, i image.Image) error {
	bw := bufio.NewWriter(w)
	defer bw.Flush()
	w = bw
	_, err := w.Write([]byte(magic))
	if err != nil {
		return err
	}
	wid := i.Bounds().Dx()
	hgt := i.Bounds().Dy()
	header := []byte{byte(wid >> 8), byte(wid), byte(hgt >> 8), byte(hgt)}
	_, err = w.Write(header)
	if err != nil {
		return err
	}
	m := color.NRGBAModel

	hw := huffman.NewWriter()
	dx := i.Bounds().Min.X
	dy := i.Bounds().Min.Y
	prior := []byte{0, 0, 0, 0}
	run := 0
	ru := newByteRU(0x100)
	for y := 0; y < hgt; y++ {
		for x := 0; x < wid; x++ {
			c := m.Convert(i.At(x+dx, y+dy)).(color.NRGBA)
			cur := []byte{c.R, c.G, c.B, c.A}
			if bytes.Equal(cur, prior) {
				run++
				continue
			}

			if run > 0 {
				run--

				hw.Write(ClassTag, byte(Run))
				b := (bits.Len(uint(run)) + 7) / 8
				hw.Write(ClassBCount, byte(b))
				for i := 0; i < b; i++ {
					hw.Write(ClassRun+byte(i), byte(run>>((b-i-1)*8)))
				}
				run = 0
			}

			entry := ru.CheckAdd(cur)

			if entry != -1 {
				entry--
				hw.Write(ClassTag, byte(Lookup))
				hw.Write(ClassLookup, byte(entry))
				prior = cur
				continue
			}

			up := []byte{0, 0, 0, 0}
			if y > 0 {
				c := m.Convert(i.At(x+dx, y+dy-1)).(color.NRGBA)
				up = []byte{c.R, c.G, c.B, c.A}
			}

			maxDiff := 0
			nDiff := 0
			for i, v := range cur {
				if v != prior[i] {
					nDiff++
				}
				d := int(v) - int(prior[i])
				if d < 0 {
					d = -d - 1
				}
				if d > maxDiff {
					maxDiff = d
				}
			}

			maxUpDiff := 0
			nUpDiff := 0
			for i, v := range cur {
				if v != up[i] {
					nUpDiff++
				}
				d := int(v) - int(up[i])
				if d < 0 {
					d = -d - 1
				}
				if d > maxUpDiff {
					maxUpDiff = d
				}
			}

			t := Delta

			if nUpDiff < nDiff {
				prior = up
				t = DeltaUp
			} else if nUpDiff == nDiff && maxUpDiff < maxDiff {
				prior = up
				t = DeltaUp
			}

			hw.Write(ClassTag, byte(t))

			for i, v := range cur {
				d := v - prior[i]
				hw.Write(ClassDelta+byte(i), byte(d))
			}
			prior = cur
		}
	}
	if run > 0 {
		hw.Write(ClassTag, byte(Run))
		b := (bits.Len(uint(run)) + 7) / 8
		hw.Write(ClassRun, byte(b))
		for i := 0; i < b; i++ {
			hw.Write(ClassRun, byte(run>>(i*8)))
		}
	}

	return hw.Dump(w)
}

type byteRU struct {
	entries [][]byte
}

// newByteRU returns a new byteRU with the given capacity.
func newByteRU(capacity int) *byteRU {
	return &byteRU{
		entries: make([][]byte, capacity),
	}
}

// CheckAdd returns the old index of the entry (or -1 if it didn't exist)
// if it exists, it gets shifted to the head
// if it doesn't exist, it gets added to the head and the oldest removed
func (b *byteRU) CheckAdd(entry []byte) int {
	idx := -1
	for i, e := range b.entries {
		if bytes.Equal(e, entry) {
			idx = i
			break
		}
	}
	if idx == -1 {
		copy(b.entries[1:], b.entries)
		b.entries[0] = entry
		return idx
	}
	if idx == 0 {
		return idx
	}
	copy(b.entries[1:idx], b.entries)
	b.entries[0] = entry
	return idx
}

// Get returns the entry at the given index, moving it to the head
func (b *byteRU) Get(idx int) []byte {
	if idx == 0 {
		return b.entries[0]
	}
	tmp := b.entries[idx]
	copy(b.entries[1:idx], b.entries[0:])
	b.entries[0] = tmp
	return tmp
}
