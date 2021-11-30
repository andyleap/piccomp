package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"text/tabwriter"
)

type Results struct {
	Sizes map[string]int
}

// Read results from a file and return a Results struct.
func readResults(file string) (Results, error) {
	f, err := os.Open(file)
	if err != nil {
		return Results{}, err
	}
	defer f.Close()
	var r Results
	err = json.NewDecoder(f).Decode(&r)
	return r, err
}

// Read 2 files of json encoded results from command line arguments and compare them.
func main() {
	rs := []Results{}
	if len(os.Args) <= 1 {
		log.Fatal("Usage: compare <file1> <fileN>...")
	}
	for _, file := range os.Args[1:] {
		r, err := readResults(file)
		if err != nil {
			log.Fatal(err)
		}
		rs = append(rs, r)
	}
	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 1, ' ', 0)
	defer tw.Flush()
	// Print header
	fmt.Fprintf(tw, "Size")
	for i := range rs {
		if i > 0 {
			fmt.Fprintf(tw, "\t%s\t%%", os.Args[i+1])
		} else {
			fmt.Fprintf(tw, "\t%s", os.Args[i+1])
		}
	}
	fmt.Fprintf(tw, "\n")

	allNames := []string{}
	exists := map[string]struct{}{}
	for _, r := range rs {
		for name := range r.Sizes {
			if _, ok := exists[name]; !ok {
				allNames = append(allNames, name)
				exists[name] = struct{}{}
			}
		}
	}
	sort.Strings(allNames)

	for _, name := range allNames {
		sizes := []int{}
		for _, r := range rs {
			sizes = append(sizes, r.Sizes[name])
		}

		fmt.Fprintf(tw, "%s", name)
		for i, size := range sizes {
			if i > 0 {
				fmt.Fprintf(tw, "\t%d\t%6.2f", size/1024, float64(size)/float64(sizes[0])*100)
			} else {
				fmt.Fprintf(tw, "\t%d", size/1024)
			}
		}
		fmt.Fprintf(tw, "\n")
	}

}
