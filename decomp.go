package piccomp

import (
	"bufio"
	"image"
	"image/color"
	"io"

	"github.com/andyleap/piccomp/huffman"
)

var magic = "piccomp"

func init() {
	image.RegisterFormat("piccomp", magic, Decode, DecodeConfig)
}

func readMagic(r io.Reader) error {
	m := make([]byte, len(magic))
	_, err := io.ReadFull(r, m)
	if err != nil {
		return err
	}
	if string(m) != magic {
		return image.ErrFormat
	}
	return nil
}

// Decode reads a piccomp image from r and returns it as an image.Image.
func Decode(r io.Reader) (image.Image, error) {
	br := bufio.NewReader(r)
	con, err := DecodeConfig(br)
	if err != nil {
		return nil, err
	}
	i := image.NewNRGBA(image.Rect(0, 0, con.Width, con.Height))

	hr := huffman.NewReader(br)

	c := color.NRGBA{}
	run := 0
	ru := newByteRU(0x100)
	for y := 0; y < con.Height; y++ {
		for x := 0; x < con.Width; x++ {
			if run > 0 {
				run--
				i.Set(x, y, c)
				continue
			}
			t, err := hr.Read(ClassTag)
			if err != nil {
				return nil, err
			}
			switch tag(t) {
			case DeltaUp:
				c = color.NRGBA{}
				if y > 0 {
					c = i.At(x, y-1).(color.NRGBA)
				}
				fallthrough
			case Delta:
				dR, err := hr.Read(ClassDelta)
				if err != nil {
					return nil, err
				}
				c.R += dR
				dG, err := hr.Read(ClassDelta + 1)
				if err != nil {
					return nil, err
				}
				c.G += dG
				dB, err := hr.Read(ClassDelta + 2)
				if err != nil {
					return nil, err
				}
				c.B += dB
				dA, err := hr.Read(ClassDelta + 3)
				if err != nil {
					return nil, err
				}
				c.A += dA
				ru.CheckAdd([]byte{c.R, c.G, c.B, c.A})
			case Lookup:
				entry, err := hr.Read(ClassLookup)
				if err != nil {
					return nil, err
				}
				b := ru.Get(int(entry) + 1)
				c.R = b[0]
				c.G = b[1]
				c.B = b[2]
				c.A = b[3]
			case Run:
				l, err := hr.Read(ClassBCount)
				if err != nil {
					return nil, err
				}
				run = 0
				for i := 0; i < int(l); i++ {
					b, err := hr.Read(ClassRun + byte(i))
					if err != nil {
						return nil, err
					}
					run <<= 8
					run |= int(b)
				}
			default:
				return nil, image.ErrFormat
			}

			i.Set(x, y, c)
			/*
				t, err := br.ReadByte()
				if err != nil {
					return i, err
				}
				switch tag(t & 0xF0) {
				case DeltaS:
					b, err := br.ReadByte()
					if err != nil {
						return i, err
					}
					d := (int16(t&0xF) << 8) | int16(b)
					c.R += uint8((d << 4) >> 13)
					c.G += uint8((d << 7) >> 13)
					c.B += uint8((d << 10) >> 13)
					c.A += uint8((d << 13) >> 13)
					ru.CheckAdd([]byte{c.R, c.G, c.B, c.A})
				case DeltaM:
					d := int32(t&0xF) << 16
					b, err := br.ReadByte()
					if err != nil {
						return i, err
					}
					d |= int32(b) << 8
					b, err = br.ReadByte()
					if err != nil {
						return i, err
					}
					d |= int32(b)
					c.R += uint8((d << 12) >> 27)
					c.G += uint8((d << 17) >> 27)
					c.B += uint8((d << 22) >> 27)
					c.A += uint8((d << 27) >> 27)
					ru.CheckAdd([]byte{c.R, c.G, c.B, c.A})
				case DeltaUS:
					c = color.NRGBA{}
					if y > 0 {
						c = i.At(x, y-1).(color.NRGBA)
					}
					b, err := br.ReadByte()
					if err != nil {
						return i, err
					}
					d := (int16(t&0xF) << 8) | int16(b)
					c.R += uint8((d << 4) >> 13)
					c.G += uint8((d << 7) >> 13)
					c.B += uint8((d << 10) >> 13)
					c.A += uint8((d << 13) >> 13)
					ru.CheckAdd([]byte{c.R, c.G, c.B, c.A})
				case DeltaUM:
					c = color.NRGBA{}
					if y > 0 {
						c = i.At(x, y-1).(color.NRGBA)
					}
					d := int32(t&0xF) << 16
					b, err := br.ReadByte()
					if err != nil {
						return i, err
					}
					d |= int32(b) << 8
					b, err = br.ReadByte()
					if err != nil {
						return i, err
					}
					d |= int32(b)
					c.R += uint8((d << 12) >> 27)
					c.G += uint8((d << 17) >> 27)
					c.B += uint8((d << 22) >> 27)
					c.A += uint8((d << 27) >> 27)
					ru.CheckAdd([]byte{c.R, c.G, c.B, c.A})
				case RunS:
					run = int(t & 0x0F)
				case RunL:
					run = int(t & 0x0F)
					for {
						b, err := br.ReadByte()
						if err != nil {
							return i, err
						}
						run <<= 7
						run |= int(b) & 0x7F
						if b&0x80 == 0 {
							break
						}
					}
				case PlainUp:
					c = color.NRGBA{}
					if y > 0 {
						c = i.At(x, y-1).(color.NRGBA)
					}
					fallthrough
				case Plain:
					if t&(1<<3) != 0 {
						c.R, err = br.ReadByte()
						if err != nil {
							return i, err
						}
					}
					if t&(1<<2) != 0 {
						c.G, err = br.ReadByte()
						if err != nil {
							return i, err
						}
					}
					if t&(1<<1) != 0 {
						c.B, err = br.ReadByte()
						if err != nil {
							return i, err
						}
					}
					if t&(1<<0) != 0 {
						c.A, err = br.ReadByte()
						if err != nil {
							return i, err
						}
					}
					ru.CheckAdd([]byte{c.R, c.G, c.B, c.A})
				default:
					if t&(1<<7) != 1<<7 {
						return i, image.ErrFormat
					}
					ba := ru.Get(int(t&0x7F) + 1)
					c = color.NRGBA{ba[0], ba[1], ba[2], ba[3]}
				}
			*/

		}
	}

	return i, nil
}

// DecodeConfig returns the color model and dimensions of a piccomp image without decoding the entire image.
func DecodeConfig(r io.Reader) (image.Config, error) {
	err := readMagic(r)
	if err != nil {
		return image.Config{}, err
	}
	header := make([]byte, 4)
	_, err = io.ReadFull(r, header)
	if err != nil {
		return image.Config{}, err
	}
	c := image.Config{
		ColorModel: color.NRGBAModel,
		Width:      int(header[0])<<8 | int(header[1]),
		Height:     int(header[2])<<8 | int(header[3]),
	}

	return c, nil
}
