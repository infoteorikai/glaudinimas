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
	sub [N]int
}

var dicts = make([]dictionary, N+1)

func compress(r io.Reader, w io.Writer, k int, reset bool) {
	bw := bitio.NewWriter(w)
	defer bw.Close()
	br := bufio.NewReader(r)

	cur := 0
	for i := range dicts[0].sub {
		dicts[0].sub[i] = i + 1
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
		if dicts[cur].sub[b] == 0 {
			bw.WriteBits(uint64(cur-1), uint8(k))

			// Add to dictionary if not full
			if len(dicts)-1 < 1<<k {
				dicts[cur].sub[b] = len(dicts)
				dicts = append(dicts, dictionary{})
			}

			// Start from the root of tree
			cur = 0
		}
		// Advance position in the tree
		cur = dicts[cur].sub[b]

		// If we are reseting, do it now
		if len(dicts)-1 == 1<<k && reset {
			dicts = dicts[:N+1]
			for i := 1; i <= N; i++ {
				dicts[i] = dictionary{}
			}
			cur = 0
			// We need to read current byte again as we're starting from scratch
			br.UnreadByte()
		}
	}

	// Write out any remaining data if not empty
	if cur != 0 {
		bw.WriteBits(uint64(cur-1), uint8(k))
	}
}
