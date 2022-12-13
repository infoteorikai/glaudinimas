package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/icza/bitio"
	"github.com/pkg/profile"
)

var in = flag.String("in", "/path/to/input", "input file")
var out = flag.String("out", "/path/to/output", "output file")
var k = flag.Int("k", 12, "dictionary size N=2^k")
var reset = flag.Bool("r", false, "reset dictionary when full")

func main() {
	p := flag.Bool("p", false, "enable profiling")
	flag.Parse()
	if *p {
		defer profile.Start().Stop()
	}

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

const N = 256

type dictionary struct {
	num int
	sub [N]*dictionary
}

func (d *dictionary) reset() {
	for i := range d.sub {
		for j := range d.sub[i].sub {
			d.sub[i].sub[j] = nil
		}
	}
}

func compress(r io.Reader, w io.Writer, k int, reset bool) {
	bw := bitio.NewWriter(w)
	defer bw.Close()
	br := bufio.NewReader(r)

	var dict dictionary
	for i := range dict.sub {
		dict.sub[i] = &dictionary{num: i}
	}
	cur := &dict
	dsize := N

	// Write out the parameters
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

			// Add to dictionary if not full
			if dsize < 1<<k {
				cur.sub[b] = &dictionary{num: dsize}
				dsize++
			}

			// Start from the root of tree
			cur = &dict
		}
		// Advance position in the tree
		cur = cur.sub[b]

		// If we are reseting, do it now
		if dsize == 1<<k && reset {
			dict.reset()
			dsize = N
			cur = &dict
			// We need to read current byte again as we're starting from scratch
			br.UnreadByte()
		}
	}

	// Write out any remaining data if not empty
	if cur != &dict {
		bw.WriteBits(uint64(cur.num), uint8(k))
	}
}
