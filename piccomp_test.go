package piccomp

import (
	"bytes"
	"image"
	"io"
	"log"
	"os"
	"testing"

	"image/png"

	_ "github.com/lmittmann/ppm"
)

func TestSave(t *testing.T) {
	err := trip(t, "test.png", "test.piccomp", Save)
	if err != nil {
		t.Error(err)
	}
}

func TestLoop(t *testing.T) {

	err := trip(t, "test.png", "test.piccomp", Save)
	if err != nil {
		t.Error(err)
	}
	err = trip(t, "test.piccomp", "test.piccomp.png", png.Encode)
	if err != nil {
		t.Error(err)
	}
}

func trip(t *testing.T, from, to string, encode func(io.Writer, image.Image) error) error {
	r, err := os.Open(from)
	if err != nil {
		return err
	}
	defer r.Close()
	i, _, err := image.Decode(r)
	if err != nil {
		log.Println(err)
	}
	w, err := os.Create(to)
	if err != nil {
		return err
	}
	defer w.Close()
	return encode(w, i)
}

func toBuffer(t *testing.T, from string, encode func(io.Writer, image.Image) error) ([]byte, error) {
	r, err := os.Open(from)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	i, _, err := image.Decode(r)
	if err != nil {
		log.Println(err)
	}
	w := &bytes.Buffer{}
	err = encode(w, i)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}
