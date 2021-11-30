package main

import (
	"encoding/json"
	"fmt"
	"image"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"sync"

	"github.com/andyleap/piccomp"

	_ "image/png"
)

type Results struct {
	Sizes map[string]int
}

func main() {
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	wg := sync.WaitGroup{}
	in := make(chan struct {
		k string
		v int
	})
	data := map[string]int{}
	count := 0
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			return nil
		}
		wg.Add(1)
		count++
		go func(path string) {
			defer wg.Done()
			i, err := toBuffer(path)
			if err != nil {
				log.Println(err)
			}

			piccompLen, err := measure(i, piccomp.Save)
			if err != nil {
				log.Println(err)
			}

			name, err := filepath.Rel(dir, path)
			if err != nil {
				log.Println(err)
			}

			in <- struct {
				k string
				v int
			}{name, piccompLen}
		}(path)
		return nil
	})
	go func() {
		defer close(in)
		wg.Wait()
	}()
	done := 0

	numLen := len(fmt.Sprintf("%v", count))
	fmtString := fmt.Sprintf("\r%%%dd/%%%dd", numLen, numLen)

	for d := range in {
		done++
		fmt.Fprintf(os.Stderr, fmtString, done, count)
		data[d.k] = d.v
	}
	fmt.Fprintf(os.Stderr, "\n")
	results := Results{
		Sizes: data,
	}
	buf, err := json.Marshal(results)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(buf))
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
