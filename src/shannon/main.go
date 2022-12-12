package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"

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

type pair struct {
	word uint64
	freq float64
}

type byFreq []pair

func compress(r io.Reader, w io.Writer, l int) {
	bw := bitio.NewWriter(w)
	defer bw.Close()
	br := bitio.NewReader(r)
	wordCount := 0
	var frequencies map[uint64]float64 = make(map[uint64]float64)


	// reading and storing all words
	for {
		b, err := br.ReadBits(uint8(l))
		if err != nil {
			break
		}

		wordCount ++
		_, ok := frequencies[b]
		
		if ok {
			frequencies[b] ++
		} else {
			frequencies[b] = 1
		}
		
		b = 0
	}


	var freqSorted []pair

	// saving and sorting frequencies
	for w, f := range frequencies {
		frequencies[w] = f/float64(wordCount)
		freqSorted = append(freqSorted, pair{word: w, freq: frequencies[w]})
	}
	sort.Sort(byFreq(freqSorted))
}



func (f byFreq) Len() int {
    return len(f)
}
func (f byFreq) Swap(i, j int) {
    f[i], f[j] = f[j], f[i]
}
func (f byFreq) Less(i, j int) bool {
    return f[i].freq < f[j].freq
}