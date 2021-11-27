package piccomp

import (
	"bufio"
	"image"
	"image/color"
	"io"
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

	c := color.NRGBA{}
	run := 0
	ru := newByteRU(0x80)
	for y := 0; y < con.Height; y++ {
		for x := 0; x < con.Width; x++ {
			if run > 0 {
				run--
				i.Set(x, y, c)
				continue
			}
			t, err := br.ReadByte()
			if err != nil {
				return nil, err
			}
			switch tag(t & 0xF0) {
			case DeltaS:
				b, err := br.ReadByte()
				if err != nil {
					return nil, err
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
					return nil, err
				}
				d |= int32(b) << 8
				b, err = br.ReadByte()
				if err != nil {
					return nil, err
				}
				d |= int32(b)
				c.R += uint8((d << 12) >> 27)
				c.G += uint8((d << 17) >> 27)
				c.B += uint8((d << 22) >> 27)
				c.A += uint8((d << 27) >> 27)
				ru.CheckAdd([]byte{c.R, c.G, c.B, c.A})
			case Run:
				run = int(t & 0x0F)
			case RunL:
				n := int(t & 0x0F)
				run = 0
				for n > 0 {
					b, err := br.ReadByte()
					if err != nil {
						return nil, err
					}
					run <<= 8
					run += int(b)
					n--
				}
			case Plain:
				if t&(1<<3) != 0 {
					c.R, err = br.ReadByte()
					if err != nil {
						return nil, err
					}
				}
				if t&(1<<2) != 0 {
					c.G, err = br.ReadByte()
					if err != nil {
						return nil, err
					}
				}
				if t&(1<<1) != 0 {
					c.B, err = br.ReadByte()
					if err != nil {
						return nil, err
					}
				}
				if t&(1<<0) != 0 {
					c.A, err = br.ReadByte()
					if err != nil {
						return nil, err
					}
				}
				ru.CheckAdd([]byte{c.R, c.G, c.B, c.A})
			default:
				if t&(1<<7) != 1<<7 {
					return nil, image.ErrFormat
				}
				ba := ru.Get(int(t&0x7F) + 1)
				c = color.NRGBA{ba[0], ba[1], ba[2], ba[3]}
			}

			i.Set(x, y, c)
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
