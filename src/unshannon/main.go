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



func main() {
	flag.Parse()

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

	uncompress(fi, fo)
}



func uncompress(r io.Reader, w io.Writer) {
	br := bitio.NewReader(r)
	bw := bitio.NewWriter(w)
	defer bw.Close()



	b, err := br.ReadBits(32)
		if err != nil {
			fmt.Println("Error: ", err)
			return
		}
	wordCount := b + 1
	
	b, err = br.ReadBits(5)
		if err != nil {
			fmt.Println("Error: ", err)
			return
		}
	wordLen := b + 1

	b, err = br.ReadBits(16)
		if err != nil {
			fmt.Println("Error: ", err)
			return
		}
	dictLen := b + 1

	b, err = br.ReadBits(5)
		if err != nil {
			fmt.Println("Error: ", err)
			return
		}
	leftover := b



	var dictionary map[string][]bool = make(map[string][]bool)
	
	for i := 0; i < int(dictLen); i++ {

		word := make([]bool, wordLen)
		for j := range word {
			bit, err := br.ReadBool()
			if err != nil {
				break
			}

			word[j] = bit
		}
		
		b, err = br.ReadBits(5)
		if err != nil {
			fmt.Println("Error: ", err)
			return
		}
		codeLen := b + 1

		code := ""
		for j := 0; j < int(codeLen); j++ {
			bit, err := br.ReadBool()
			if err != nil {
				break
			}

			if bit {
				code += "1"
			} else {
				code += "0"
			}
		}

		dictionary[code] = word
	}
	
	leftoverWord := make([]bool, leftover)
	for i := range leftoverWord {
		bit, err := br.ReadBool()
		if err != nil {
			break
		}

		leftoverWord[i] = bit
	}

	c := ""
	for wordCount > 0 {
		bit, err := br.ReadBool()
		if err != nil {
			break
		}

		if bit {
			c += "1"
		} else {
			c += "0"
		}

		if word, ok := dictionary[c]; ok {
			wordCount --
			for i := range word {
				bw.WriteBool(word[i])
			}
			c = ""
		} 
	}

	if leftover > 0 {
		for i := range leftoverWord {
			bw.WriteBool(leftoverWord[i])
		}
	}
}