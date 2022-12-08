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
	buf := make([]byte, 1)
	prev, pos := -1, 0
	var pfirst byte

	for {
		ub, err := br.ReadBits(uint8(k))
		if err != nil {
			break
		}
		b := int(ub)

		if b > len(dict) {
			panic("invalid data")
		}
		if b == len(dict) {
			buf[len(buf)-1] = pfirst
		} else {
			pos = len(buf)
			for b >= 0 {
				pos--
				buf[pos] = dict[b].suffix
				b = dict[b].parent
			}

			pfirst = buf[pos]
		}
		if prev >= 0 {
			if len(dict) < 1<<k {
				dict = append(dict, row{parent: prev, suffix: buf[pos]})
			} else if reset {
				panic("not implemented")
			}
		}
		//fmt.Fprint(w, "|")
		w.Write(buf[pos:])
		buf = append(buf, 0)
		prev = int(ub)
	}
}
