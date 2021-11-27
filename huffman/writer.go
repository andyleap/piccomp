package huffman

import (
	"io"
	"math/bits"
	"sort"
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

func (w *Writer) buildCodes() map[byte]map[byte]string {
	counts := w.buildCounts()
	classCodes := map[byte]map[byte]string{}
	for class, cnts := range counts {
		tree := buildTree(cnts)
		codes := map[byte]string{}
		tree.buildCodes("", codes)
		classCodes[class] = codes
	}
	return classCodes
}

func (w *Writer) Dump(wr io.Writer) error {
	codes := w.buildCodes()
	bw := newBitWriter(wr)
	for c, codes := range codes {
		maxlength := 0
		for _, code := range codes {
			if len(code) > maxlength {
				maxlength = len(code)
			}
		}
		bits.Len8(uint8(maxlength))
		bw.writeBits(int(c), 8)
		bw.writeBits(int(maxlength), 4)
		for i := 0; i <= 255; i++ {
			bw.writeBits(len(codes[byte(i)]), uint(maxlength))
		}
	}
	bw.writeBits(int(255), 8)
	for _, t := range w.tokens {
		code := codes[t.class][t.value]
		for _, c := range code {
			err := bw.writeBit(c == '1')
			if err != nil {
				return err
			}
		}
	}
	bw.flush()
	return nil
}

// huffTree is a node in a huffman tree
type huffTree struct {
	left  *huffTree
	right *huffTree
	value byte

	full bool
}

// buildTree builds a huffman tree from value counts
func buildTree(counts map[byte]int) *huffTree {
	keys := []byte{}
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return buildSubTree(keys, counts)
}

// buildSubTree builds a huffman tree node from a sorted list of keys and counts
func buildSubTree(keys []byte, counts map[byte]int) *huffTree {
	if len(keys) == 1 {
		return &huffTree{
			value: keys[0],
		}
	}
	if len(keys) == 0 {
		return nil
	}
	totalWeight := 0
	for _, k := range keys {
		totalWeight += counts[k]
	}
	// find the point that splits the weight
	split := 0
	splitWeight := 0
	for i, k := range keys {
		splitWeight += counts[k]
		if splitWeight >= totalWeight/2 {
			split = i + 1
			break
		}
	}
	if split == len(keys) {
		split--
	}
	left := buildSubTree(keys[:split], counts)
	right := buildSubTree(keys[split:], counts)
	return &huffTree{
		left:  left,
		right: right,
	}
}

// buildCodes builds a map of codes for each value of a huffTree
func (t *huffTree) buildCodes(prefix string, codes map[byte]string) {
	if t.left == nil && t.right == nil {
		codes[t.value] = prefix
		return
	}
	if t.left != nil {
		t.left.buildCodes(prefix+"0", codes)
	}
	if t.right != nil {
		t.right.buildCodes(prefix+"1", codes)
	}
}
