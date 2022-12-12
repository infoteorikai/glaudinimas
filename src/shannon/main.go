package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"math"
	"strconv"

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

	fileStat, err := fi.Stat()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fileSize := uint64(fileStat.Size())

	fo, err := os.Create(*out)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer fo.Close()

	compress(fi, fo, *l, fileSize)
}



type pair struct {
	word uint64
	freq float64
}

type tuple struct {
	word uint64
	freq float64
	len int
}

type byFreq []pair



func compress(r io.Reader, w io.Writer, l int, fileSize uint64) {
	bw := bitio.NewWriter(w)
	defer bw.Close()
	br := bitio.NewReader(r)
	wordCount := fileSize / uint64(l)
	leftover := fileSize % uint64(l)
	var frequencies map[uint64]float64 = make(map[uint64]float64)



	// reading and storing all words
	for i := 0; i < int(wordCount); i++ {
		b, err := br.ReadBits(uint8(l))
		if err != nil {
			break
		}

		_, ok := frequencies[b]
		
		if ok {
			frequencies[b] ++
		} else {
			frequencies[b] = 1
		}
		
		b = 0
	}

	// saving the leftover tail
	if leftover > 0 {
		leftoverWord, err := br.ReadBits(uint8(l))
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
	}



	var freqSorted []pair
	var freqSummed []tuple

	// saving, sorting and summing frequencies
	for w, f := range frequencies {
		frequencies[w] = f/float64(wordCount)
		freqSorted = append(freqSorted, pair{word: w, freq: frequencies[w]})
	}
	
	sort.Sort(byFreq(freqSorted))

	freqSum := 0.0
	for _, p := range freqSorted {
		length := int(math.Ceil(-math.Log2(p.freq)))
		freqSummed = append(freqSummed, tuple{word: p.word, freq: freqSum, len: length})
		freqSum += p.freq
	}


	
	// creating dictionary
	var dictionary map[uint64]int64 = make(map[uint64]int64)

	for _, t := range freqSummed {
		code := ""
		fraction := t.freq
		for i := 0; i < t.len; i++ {
			fraction *= 2
			if fraction < 0 {
				code += "0"
			} else {
				code += "1"
				fraction -= 1
			}
		}

		if c, err := strconv.ParseInt(code, 2, 64); err != nil {
			fmt.Println(err)
		} else {
			dictionary[t.word] = c 
		}
	} 



	// writing to file
	// 
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