package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/icza/bitio"
)

var in = flag.String("in", "/path/to/input", "input file")
var out = flag.String("out", "/path/to/output", "output file")
var k = flag.Int("k", 12, "dictionary size N=2^k")
var reset = flag.Bool("r", false, "reset dictionary when full")

func main() {
	flag.Parse()

	if *k < 8 || *k > 32 {
		fmt.Printf("Error: k must be in [8; 32]")
	}

	fi, err := os.Open(*in)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer fi.Close()

	fo, err := os.Create(*out)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer fo.Close()

	compress(fi, fo, *k, *reset)
}

type dictionary struct {
	num int
	sub [256]*dictionary
}

func compress(r io.Reader, w io.Writer, k int, reset bool) {
	bw := bitio.NewWriter(w)
	defer bw.Close()
	br := bitio.NewReader(r)

	var dict dictionary
	for i := range dict.sub {
		dict.sub[i] = &dictionary{num: i}
	}
	cur := &dict
	dsize := len(dict.sub)

	bw.WriteBool(reset)
	bw.WriteBits(uint64(k-1), 5)

	for {
		b, err := br.ReadByte()
		if err != nil {
			break
		}

		// Longest match ends here
		if cur.sub[b] == nil {
			bw.WriteBits(uint64(cur.num), uint8(k))
			if dsize < 1<<k {
				cur.sub[b] = &dictionary{num: dsize}
				dsize++
			} else if reset {
				panic("not implemented")
			}
			// Start from empty
			cur = dict.sub[b]
		} else {
			cur = cur.sub[b]
		}
	}

	// Write out any remaining data if not empty
	if cur != &dict {
		bw.WriteBits(uint64(cur.num), uint8(k))
	}
}
