package piccomp

import (
	"bufio"
	"bytes"
	"image"
	"image/color"
	"io"
)

type tag uint8

const (
	Plain tag = iota << 4
	PlainUp
	RunS
	RunL
	DeltaS
	DeltaM
	DeltaUS
	DeltaUM
	Lookup
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

	dx := i.Bounds().Min.X
	dy := i.Bounds().Min.Y
	prior := []byte{0, 0, 0, 0}
	run := 0
	ru := newByteRU(0x80)
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
				if run <= 0x0F {
					_, err = w.Write([]byte{byte(byte(RunS) | byte(run))})
					if err != nil {
						return err
					}
				} else {
					max := 0
					trun := run >> 4
					for trun > 0 {
						max++
						trun >>= 7
					}
					b := []byte{}
					b = append(b, byte(RunL)|byte((run>>uint(max*7))&0x0F))
					for max > 0 {
						max--
						b = append(b, byte(run>>uint(max*7))&0x7F)
						if max > 0 {
							b[len(b)-1] |= 0x80
						}
					}
					_, err = w.Write(b)
					if err != nil {
						return err
					}
				}
				run = 0
			}

			entry := ru.CheckAdd(cur)

			if entry != -1 {
				entry--
				_, err = w.Write([]byte{byte(byte(Lookup) | byte(entry))})
				if err != nil {
					return err
				}
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

			if maxDiff <= 3 && (nDiff >= 1 && nUpDiff >= 1) {
				d := uint16(0)
				for i, v := range cur {
					d <<= 3
					d |= uint16((v - prior[i]) & 0x07)
				}
				_, err = w.Write([]byte{byte(DeltaS) | byte(d>>8), byte(d)})
				if err != nil {
					return err
				}
				prior = cur
				continue
			}

			if maxUpDiff <= 3 && (nDiff >= 1 && nUpDiff >= 1) {
				d := uint16(0)
				for i, v := range cur {
					d <<= 3
					d |= uint16((v - up[i]) & 0x07)
				}
				_, err = w.Write([]byte{byte(DeltaUS) | byte(d>>8), byte(d)})
				if err != nil {
					return err
				}
				prior = cur
				continue
			}

			if maxDiff <= 0xF && (nDiff >= 2 && nUpDiff >= 2) {
				d := uint32(0)
				for i, v := range cur {
					d <<= 5
					d |= uint32((v - prior[i]) & 0x1F)
				}
				_, err = w.Write([]byte{byte(DeltaM) | byte(d>>16), byte(d >> 8), byte(d)})
				if err != nil {
					return err
				}
				prior = cur
				continue
			}

			if maxUpDiff <= 0xF && (nDiff >= 2 && nUpDiff >= 2) {
				d := uint32(0)
				for i, v := range cur {
					d <<= 5
					d |= uint32((v - up[i]) & 0x1F)
				}
				_, err = w.Write([]byte{byte(DeltaUM) | byte(d>>16), byte(d >> 8), byte(d)})
				if err != nil {
					return err
				}
				prior = cur
				continue
			}

			b := []byte{0}
			t := Plain

			if nUpDiff < nDiff {
				prior = up
				t = PlainUp
			}

			if cur[0] != prior[0] {
				b = append(b, cur[0])
				t |= 1 << 3
			}
			if cur[1] != prior[1] {
				b = append(b, cur[1])
				t |= 1 << 2
			}
			if cur[2] != prior[2] {
				b = append(b, cur[2])
				t |= 1 << 1
			}
			if cur[3] != prior[3] {
				b = append(b, cur[3])
				t |= 1 << 0
			}
			b[0] = byte(t)
			_, err = w.Write(b)
			if err != nil {
				return err
			}
			prior = cur
		}
	}
	if run > 0 {
		run--
		if run <= 0xF {
			_, err = w.Write([]byte{byte(byte(RunS) | byte(run))})
			if err != nil {
				return err
			}
		} else {
			max := 0
			trun := run >> 4
			for trun > 0 {
				max++
				trun >>= 7
			}
			b := []byte{}
			b = append(b, byte(RunL)|byte((run>>uint(max*7))&0x0F))
			for max > 0 {
				max--
				b = append(b, byte(run>>uint(max*7))&0x7F)
				if max > 0 {
					b[len(b)-1] |= 0x80
				}
			}
			_, err = w.Write(b)
			if err != nil {
				return err
			}
		}
		run = 0
	}

	return nil
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
