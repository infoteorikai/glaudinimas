package main

import (
	"bufio"
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

const N = 256

type dictionary struct {
	num int
	sub [N]*dictionary
}

func (d *dictionary) reset() {
	for i := range d.sub {
		d.sub[i] = &dictionary{num: i}
	}
}

func compress(r io.Reader, w io.Writer, k int, reset bool) {
	bw := bitio.NewWriter(w)
	defer bw.Close()
	br := bufio.NewReader(r)

	var dict dictionary
	dict.reset()
	cur := &dict
	dsize := N

	// Write out the parameters
	bw.WriteBool(reset)
	bw.WriteBits(uint64(k-1), 5)

	ln := 0

	for {
		b, err := br.ReadByte()
		if err != nil {
			break
		}

		// Longest match ends here
		if cur.sub[b] == nil {
			bw.WriteBits(uint64(cur.num), uint8(k))
			//fmt.Println("record ", cur.num, " len: ", ln)
			ln = 0

			// Add to dictionary if not full
			if dsize < 1<<k {
				cur.sub[b] = &dictionary{num: dsize}
				dsize++
			}

			// If we are reseting, do it now
			if dsize == 1<<k && reset {
				dict.reset()
				//fmt.Println("resetting")

				dsize = N
				ln = 0
				br.UnreadByte()
				cur = &dict
				continue

				// Save this new byte from the tree root
				//dict.sub[b] = &dictionary{num: N}
				//dsize = N + 1
				//ln = 1
			}

			// } else if reset {
			// 	dict.reset()
			// 	dsize = len(dict.sub)
			// 	dict.sub[b] = &dictionary{num: dsize}
			// 	dsize++
			// }

			// Start from the root of tree
			cur = &dict
		}
		ln++
		// Advance position in the tree
		cur = cur.sub[b]
	}

	// Write out any remaining data if not empty
	if cur != &dict {
		bw.WriteBits(uint64(cur.num), uint8(k))
	}
}
