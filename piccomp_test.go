package piccomp

import (
	"bytes"
	"image"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"testing"

	"image/png"

	_ "github.com/lmittmann/ppm"
)

func TestSave(t *testing.T) {
	err := trip(t, "artificial.png", "artificial.piccomp", Save)
	if err != nil {
		t.Error(err)
	}
}

func TestLoop(t *testing.T) {

	err := trip(t, "artificial.png", "artificial.piccomp", Save)
	if err != nil {
		t.Error(err)
	}
	err = trip(t, "artificial.piccomp", "artificial.piccomp.png", png.Encode)
	if err != nil {
		t.Error(err)
	}
	t.Error("ugh")
}

func TestAll(t *testing.T) {
	files, err := ioutil.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		f := f
		if !strings.HasSuffix(f.Name(), ".ppm") {
			continue
		}
		t.Run(f.Name(), func(t *testing.T) {
			t.Parallel()

			pngBuf, err := toBuffer(t, "testdata/"+f.Name(), png.Encode)
			if err != nil {
				t.Error(err)
			}
			piccompBuf, err := toBuffer(t, "testdata/"+f.Name(), Save)
			if err != nil {
				t.Error(err)
			}

			t.Log(f.Size(), len(pngBuf), len(piccompBuf))
		})
	}
	t.Error("test")
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
