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
	var dictionary map[uint64]string = make(map[uint64]string)

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

		dictionary[t.word] = code
	} 



	// writing to file
	// i could make these call a function but i dont wanna

	// word length-1 		5
	wordLenBitlen := uint8(math.Ceil(math.Log2(float64(l))))
	bw.WriteBits(0, uint8(5-wordLenBitlen))
	bw.WriteBits(uint64(l-1), wordLenBitlen)

	// dictionary len-1 	6
	dictLen := len(dictionary)
	wordCountBitlen := uint8(math.Ceil(math.Log2(float64(dictLen))))
	bw.WriteBits(0, uint8(6-wordCountBitlen))
	bw.WriteBits(uint64(dictLen-1), wordCountBitlen)

	// leftover length	 	5
	leftoverBitlen := uint8(math.Ceil(math.Log2(float64(leftover+1))))
	bw.WriteBits(0, uint8(5-leftoverBitlen))
	bw.WriteBits(leftover, leftoverBitlen)

	for w, c := range dictionary {

		// word			wordlen
		wordBitlen := int(math.Ceil(math.Log2(float64(w+1))))
		bw.WriteBits(0, uint8(l-wordBitlen))
		bw.WriteBits(w, uint8(wordBitlen))

		// codelen-1	5
		fmt.Println(len(c), freqSummed[w].len)
		codeLen := uint8(freqSummed[w].len)
		codeLenBitlen := uint8(math.Ceil(math.Log2(float64(codeLen))))
		bw.WriteBits(0, uint8(5-codeLenBitlen))
		bw.WriteBits(uint64(codeLen-1), codeLenBitlen)

		// code			codelen
		for _, rune := range c {
			if rune == 1 {
				bw.WriteBool(true)
			} else {
				bw.WriteBool(false)
			}
		}
	}
	
	// leftover 
	leftoverWordBitlen := uint64(math.Ceil(math.Log2(float64(leftoverWord+1))))
	bw.WriteBits(0, uint8(leftover-leftoverWordBitlen))
	bw.WriteBits(leftoverWord, uint8(leftoverWordBitlen))

	// encoded bits
	rf.Seek(0,0)
	for i := 0; i < int(wordCount); i++ {
		b, err := br.ReadBits(uint8(l))
		if err != nil {
			break
		}

		for _, rune := range dictionary[b] {
			if rune == 1 {
				bw.WriteBool(true)
			} else {
				bw.WriteBool(false)
			}
		}
		
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
