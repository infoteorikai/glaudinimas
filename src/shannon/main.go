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
		
		// if word is in map - increment, if not - add it
		if ok {
			frequencies[b] ++
		} else {
			frequencies[b] = 1
		}
		
		b = 0
	}

	// saving the leftover tail
	var leftoverWord uint64
	var err1 error
	if leftover > 0 {
		leftoverWord, err1 = br.ReadBits(uint8(l))
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
	var dictionary map[uint64]int64 = make(map[uint64]int64)

	for _, t := range freqSummed {
		code := ""
		fraction := t.freq
		// taking the first len bits of binary sumfrequency representation
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
			break
		} else {
			dictionary[t.word] = c 
		}
	} 



	// writing to file
	// i could make these call a function but i dont wanna

	// word length-1 		5
	wordLenBitlen := uint8(math.Ceil(math.Log2(float64(l))))
	bw.WriteBits(0, uint8(5-wordLenBitlen))
	bw.WriteBits(uint64(l-1), wordLenBitlen)

	// word count-1 		6
	wordCountBitlen := uint8(math.Ceil(math.Log2(float64(wordCount))))
	bw.WriteBits(0, uint8(6-wordCountBitlen))
	bw.WriteBits(uint64(wordCount-1), wordCountBitlen)

	// leftover length	 	5
	leftoverBitlen := uint8(math.Ceil(math.Log2(float64(leftover+1))))
	bw.WriteBits(0, uint8(5-leftoverBitlen))
	bw.WriteBits(leftover, leftoverBitlen)

	ind := 0
	for w, c := range dictionary {
		// word			wordlen
		bw.WriteBits(w, uint8(l))
		// codelen-1	5
		codeLen := uint8(freqSummed[w].len)
		codeLenBitlen := uint8(math.Ceil(math.Log2(float64(codeLen))))
		bw.WriteBits(0, uint8(5-codeLenBitlen))
		bw.WriteBits(uint64(codeLen-1), codeLenBitlen)
		// code			codelen
		bw.WriteBits(uint64(c), codeLen)
		ind++
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

		bw.WriteBits(uint64(dictionary[b]), uint8(freqSummed[b].len))
		
		b = 0
	}
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
