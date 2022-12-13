package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"math"

	"github.com/icza/bitio"
)

var in = flag.String("in", "/path/to/input", "input file")
var out = flag.String("out", "/path/to/output", "output file")
var l = flag.Int("l", 8, "word length")



func main() {
	flag.Parse()

	if *l < 1 || *l > 32 {
		fmt.Printf("Error: l must be in [1; 32]")
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

	compress(fi, fi, fo, *l, fileSize)
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



func compress(rf *os.File, r io.Reader, w io.Writer, l int, fileSize uint64) {
	br := bitio.NewReader(r)
	bw := bitio.NewWriter(w)
	defer bw.Close()
	wordCount := fileSize*8 / uint64(l)
	leftover := fileSize*8 % uint64(l)
	var frequencies map[uint64]float64 = make(map[uint64]float64)



	// reading and storing all words
	for i := 0; i < int(wordCount); i++ {
		b, err := br.ReadBits(uint8(l))
		if err != nil {
			break
		}

		frequencies[b] ++
	}

	// saving the leftover tail
	var leftoverWord uint64
	var err1 error
	if leftover > 0 {
		leftoverWord, err1 = br.ReadBits(uint8(leftover))
		if err1 != nil {
			fmt.Println("Error:", err1)
			return
		}
	}



	// array of sorted pairs (word, its frequecy)
	var freqSorted []pair
	// array of sorted summed tuples (word, sum of earlier freq, its length)
	var freqSummed map[uint64]tuple = make(map[uint64]tuple)

	// saving, sorting and summing frequencies
	for w, f := range frequencies {
		frequencies[w] = f/float64(wordCount)
		freqSorted = append(freqSorted, pair{word: w, freq: frequencies[w]})
	}
	
	sort.Sort(byFreq(freqSorted))

	freqSum := 0.0
	for _, p := range freqSorted {
		length := int(math.Ceil(-math.Log2(p.freq)))
		freqSummed[p.word] = tuple{word: p.word, freq: freqSum, len: length}
		freqSum += p.freq
	}


	
	// creating dictionary
	var dictionary map[uint64][]bool = make(map[uint64][]bool)

	for _, t := range freqSummed {
		code := make([]bool, t.len)
		fraction := t.freq
		// taking the first len bits of sumfrequency
		for i := range code {
			fraction *= 2
			code[i] = (fraction >= 1)
			fraction -= math.Floor(fraction)
		}

		dictionary[t.word] = code
	} 



	// writing to file

	// word count -1		32
	bw.WriteBits(uint64(wordCount-1), 32)

	// word length-1 		5
	bw.WriteBits(uint64(l-1), 5)

	// dictionary len-1 	6
	bw.WriteBits(uint64(len(dictionary)-1), 6)
	
	// leftover length	 	5
	bw.WriteBits(leftover, 5)

	for w, c := range dictionary {

		// word			wordlen
		bw.WriteBits(w, uint8(l))

		// codelen-1	5
		bw.WriteBits(uint64(len(c)-1), 5)

		// code			codelen
		for i := range c {
			bw.WriteBool(c[i])
		}

	}
	
	// leftover 
	bw.WriteBits(leftoverWord, uint8(leftover))

	// encoded bits
	rf.Seek(0,0)
	for i := 0; i < int(wordCount); i++ {
		b, err := br.ReadBits(uint8(l))
		if err != nil {
			break
		}

		for _, c := range dictionary[b] {
			bw.WriteBool(c)
		}
	}
}


func (f byFreq) Len() int {
    return len(f)
}
func (f byFreq) Swap(i, j int) {
    f[i], f[j] = f[j], f[i]
}
func (f byFreq) Less(i, j int) bool {
    return f[i].freq > f[j].freq
}
