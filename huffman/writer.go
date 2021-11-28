package huffman

import (
	"io"
)

// Writer writes huffman encoded data
type Writer struct {
	tokens []token
}

type token struct {
	class byte
	value byte
}

func NewWriter() *Writer {
	return &Writer{}
}

func (w *Writer) Write(c byte, v byte) {
	w.tokens = append(w.tokens, token{c, v})
}

// buildCounts builds a map of counts for each value of each class
func (w *Writer) buildCounts() map[byte]map[byte]int {
	counts := map[byte]map[byte]int{}
	for _, t := range w.tokens {
		if _, ok := counts[t.class]; !ok {
			counts[t.class] = map[byte]int{}
		}
		counts[t.class][t.value]++
	}
	return counts
}

func (w *Writer) buildTrees() map[byte]*huffTree {
	counts := w.buildCounts()
	trees := map[byte]*huffTree{}
	for class, values := range counts {
		trees[class] = buildTree(values)
	}
	return trees
}

func (w *Writer) buildCodes(trees map[byte]*huffTree) map[byte]map[byte]string {
	classCodes := map[byte]map[byte]string{}
	for class, tree := range trees {
		codes := map[byte]string{}
		tree.buildCodes("", codes)
		classCodes[class] = codes
	}
	return classCodes
}

func (w *Writer) Dump(wr io.Writer) error {
	trees := w.buildTrees()
	codes := w.buildCodes(trees)
	bw := newBitWriter(wr)
	for c, tree := range trees {
		err := bw.writeBits(int(c), 8)
		if err != nil {
			return err
		}
		err = tree.write(bw)
		if err != nil {
			return err
		}
	}
	err := bw.writeBits(int(255), 8)
	if err != nil {
		return err
	}
	for _, t := range w.tokens {
		code := codes[t.class][t.value]
		for _, c := range code {
			err = bw.writeBit(c == '1')
			if err != nil {
				return err
			}
		}
	}
	bw.flush()
	return nil
}
