package main

import (
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"text/tabwriter"

	"image/png"

	"github.com/andyleap/piccomp"

	_ "github.com/lmittmann/ppm"
)

func main() {

	files, err := ioutil.ReadDir(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	wg := sync.WaitGroup{}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprint(tw, "name\traw (kb)\tpng (kb)\tpng %\tpiccomp (kb)\tpiccomp %\n")
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		f := f
		wg.Add(1)

		go func() {
			defer wg.Done()
			i, err := toBuffer(filepath.Join(os.Args[1], f.Name()))
			if err != nil {
				log.Println(err)
			}

			raw := i.Bounds().Dx() * i.Bounds().Dy() * 4

			pngLen, err := measure(i, png.Encode)
			if err != nil {
				log.Println(err)
			}
			piccompLen, err := measure(i, piccomp.Save)
			if err != nil {
				log.Println(err)
			}

			fmt.Fprintf(tw, "%s\t%v\t%v\t%.2f\t%v\t%.2f\n", f.Name(), raw/1024, pngLen/1024, float64(pngLen)/float64(raw)*100, piccompLen/1024, float64(piccompLen)/float64(raw)*100)
		}()
	}
	wg.Wait()
	tw.Flush()
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
