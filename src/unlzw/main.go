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

func main() {
	flag.Parse()

	err := run()
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}

func run() error {
	fi, err := os.Open(*in)
	if err != nil {
		return err
	}
	defer fi.Close()

	br := bitio.NewReader(fi)
	reset, err := br.ReadBool()
	if err != nil {
		return err
	}
	k, _ := br.ReadBits(5)
	k++
	if k < 8 {
		return fmt.Errorf("k must be in [8; 32]")
	}

	fo, err := os.Create(*out)
	if err != nil {
		return err
	}
	defer fo.Close()
	w := bufio.NewWriter(fo)
	defer w.Flush()

	uncompress(br, w, int(k), reset)

	return nil
}

type row struct {
	parent int
	suffix byte
}

func uncompress(br *bitio.Reader, w io.Writer, k int, reset bool) {
	dict := make([]row, 256)
	for i := range dict {
		dict[i] = row{parent: -1, suffix: byte(i)}
	}

	// Buffer for reverse-writing data
	buf := make([]byte, 1)

	// Previously read record and last position of write in buffer
	prev, pos := -1, 0

	// First byte of previously added record
	var pfirst byte

	for {
		ub, err := br.ReadBits(uint8(k))
		if err != nil {
			break
		}
		b := int(ub)

		if b > len(dict) {
			fmt.Println(b, len(dict))
			panic("invalid data")
		}
		if b == len(dict) {
			// Special case: this record is (b) (b)_1 which last+prev first byte
			buf[len(buf)-1] = pfirst
		} else {
			pos = len(buf)
			// Traverse down the parent chain, filling buffer from end
			for b >= 0 {
				pos--
				buf[pos] = dict[b].suffix
				b = dict[b].parent
			}

			pfirst = buf[pos]
		}
		debuf := false
		// Only save dictionary records if we have the previous record
		if prev >= 0 {
			//fmt.Println("prev ok")
			if len(dict) < 1<<k {
				dict = append(dict, row{parent: prev, suffix: buf[pos]})
			}
			//fmt.Println(len(dict), " < ", 1<<k)
			if len(dict)+1 == 1<<k && reset {
				//fmt.Println("reseting")
				dict = dict[:256]
				debuf = true
			}
		}
		//fmt.Fprint(w, "|")
		w.Write(buf[pos:])
		prev = int(ub)
		//fmt.Println("record", ub, "len:", len(buf)-pos, "dict:", len(dict), "prev:", prev)
		//fmt.Println("record", ub, "len:", len(buf)-pos)
		buf = append(buf, 0)
		if debuf {
			pos = 0
			buf = buf[:1]
			prev = -1
		}
	}
}
