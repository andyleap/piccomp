package main

import (
	"image"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	"image/png"

	"github.com/andyleap/piccomp"

	_ "github.com/lmittmann/ppm"
)

func main() {

	files, err := ioutil.ReadDir("../testdata")
	if err != nil {
		log.Fatal(err)
	}
	wg := sync.WaitGroup{}
	for _, f := range files {
		if !strings.HasSuffix(f.Name(), ".ppm") {
			continue
		}
		f := f
		wg.Add(1)
		go func() {
			defer wg.Done()
			i, err := toBuffer("../testdata/" + f.Name())
			if err != nil {
				log.Println(err)
			}

			pngLen, err := measure(i, png.Encode)
			if err != nil {
				log.Println(err)
			}
			piccompLen, err := measure(i, piccomp.Save)
			if err != nil {
				log.Println(err)
			}

			log.Println(f.Name(), f.Size(), pngLen, float64(pngLen)/float64(f.Size()), piccompLen, float64(piccompLen)/float64(f.Size()))
		}()
	}
	wg.Wait()
}

func toBuffer(from string) (image.Image, error) {
	r, err := os.Open(from)
	if err != nil {
		return nil, err
	}
	defer r.Close()
	i, _, err := image.Decode(r)
	if err != nil {
		log.Println(err)
	}
	return i, err
}

func measure(i image.Image, fn func(w io.Writer, i image.Image) error) (int, error) {
	var m measureWriter
	err := fn(&m, i)
	return int(m), err
}

type measureWriter int

func (m *measureWriter) Write(p []byte) (int, error) {
	*m += measureWriter(len(p))
	return len(p), nil
}
