package huffman

import "sort"

// huffTree is a node in a huffman tree
type huffTree struct {
	left  *huffTree
	right *huffTree
	value byte

	weight int
	full   bool
}

// buildTree builds a huffman tree from value counts
func buildTree(counts map[byte]int) *huffTree {
	keys := []byte{}
	for k := range counts {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	nodes := []*huffTree{}
	for i := 0; i < len(keys); i++ {
		nodes = append(nodes, &huffTree{value: keys[i], weight: counts[keys[i]]})
	}

	for len(nodes) > 1 {
		sort.Slice(nodes, func(i, j int) bool { return nodes[i].weight < nodes[j].weight })
		nodes = append(nodes, &huffTree{left: nodes[0], right: nodes[1], weight: nodes[0].weight + nodes[1].weight})
		nodes = nodes[2:]
	}

	return nodes[0]
}

func (ht *huffTree) write(bw *bitWriter) error {
	if ht.left == nil && ht.right == nil {
		err := bw.writeBit(false)
		if err != nil {
			return err
		}
		return bw.writeBits(int(ht.value), 8)
	}
	err := bw.writeBit(true)
	if err != nil {
		return err
	}
	err = ht.left.write(bw)
	if err != nil {
		return err
	}
	return ht.right.write(bw)
}

func readTree(br *bitReader) (*huffTree, error) {
	t, err := br.readBit()
	if err != nil {
		return nil, err
	}
	if t == 0 {
		v, err := br.readBits(8)
		if err != nil {
			return nil, err
		}
		return &huffTree{value: byte(v)}, nil
	}
	left, err := readTree(br)
	if err != nil {
		return nil, err
	}
	right, err := readTree(br)
	if err != nil {
		return nil, err
	}
	return &huffTree{left: left, right: right}, nil
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
