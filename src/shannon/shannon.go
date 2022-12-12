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
var l = flag.Int("l", 8, "word length is l")

func main() {
	flag.Parse()

	if *l < 2 || *l > 16 {
		fmt.Printf("Error: l must be in [2; 16]")
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

	compress(fi, fo, *l)
}

// i not finish yet
func compress(r io.Reader, w io.Writer, l int) {
	bw := bitio.NewWriter(w)
	defer bw.Close()
	br := bitio.NewReader(r)

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