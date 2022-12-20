package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math/bits"
	"os"

	"github.com/icza/bitio"
	"github.com/pkg/profile"
)

var in = flag.String("in", "/path/to/input", "input file")
var out = flag.String("out", "/path/to/output", "output file")
var k = flag.Int("k", 12, "dictionary size n=2^k")
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

// dictionary is a node of a tree structure pointing to its children depending on the byte.
type dictionary [N]int

// dict[0] is root, and any node pointing to it is considered as not having that child
var dict = make([]dictionary, N+1)

func writeBits(w *bitio.Writer, v int) {
	// Because root is stored in dict array, subtract one more
	k := bits.Len(uint(len(dict) - 1 - 1))
	w.WriteBits(uint64(v), uint8(k))
}

func compress(r io.Reader, w io.Writer, k int, reset bool) {
	bw := bitio.NewWriter(w)
	defer bw.Close()
	br := bufio.NewReader(r)

	// fill in the initial 256 byte dictionary
	// since root is zero, all other indeces are stored increased by one
	cur := 0
	for i := range dict[0] {
		dict[0][i] = i + 1
	}

	// Write out the parameters
	bw.WriteBool(reset)
	bw.WriteBits(uint64(k-1), 5)

	for {
		b, err := br.ReadByte()
		if err != nil {
			break
		}

		// Longest match ends here
		if dict[cur][b] == 0 {
			writeBits(bw, cur-1)
			// Add to dictionary if not full
			if len(dict)-1 < 1<<k {
				dict[cur][b] = len(dict)
				dict = append(dict, dictionary{})
			}

			// Start from the root of tree
			cur = 0
		}
		// Advance position in the tree
		cur = dict[cur][b]

		// If we are reseting, do it now
		if reset && len(dict)-1 == 1<<k {
			dict = dict[:N+1]
			for i := 1; i <= N; i++ {
				dict[i] = dictionary{}
			}
			cur = 0
			// We need to read current byte again as we're starting from scratch
			br.UnreadByte()
		}
	}

	// Write out any remaining data if not empty
	if cur > 0 {
		writeBits(bw, cur-1)
	}
}
