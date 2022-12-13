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
	
	b, err := br.ReadBits(5)
		if err != nil {
			fmt.Println("Error: ", err)
			return
		}
	wordLen := b + 1
	b = 0

	b, err = br.ReadBits(6)
		if err != nil {
			fmt.Println("Error: ", err)
			return
		}
	dictLen := b + 1
	b = 0

	b, err = br.ReadBits(5)
		if err != nil {
			fmt.Println("Error: ", err)
			return
		}
	leftover := b
	b = 0
	//fmt.Println(wordLen, dictLen, leftover)

	var dictionary map[string]string = make(map[string]string)
	
	for i := 0; i < int(dictLen); i++ {

		word := ""
		for j := 0; j < int(wordLen); j++ {
			bit, err := br.ReadBool()
			if err != nil {
				break
			}

			if bit {
				word += "1"
			} else {
				word += "0"
			}
		}
		
		codeLen, errcl := br.ReadBits(5)
		if errcl != nil {
			fmt.Println("Error: ", errcl)
			return
		}
		codeLen++
		//println(codeLen)

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
		//println(word, codeLen, code)
	}
 	//fmt.Print(dictionary)
	
	leftoverWord := ""
	for i := 0; i < int(leftover); i++ {
		bit, err := br.ReadBool()
		if err != nil {
			break
		}

		if bit {
			leftoverWord += "1"
		} else {
			leftoverWord += "0"
		}
	}

	c := ""
	for {
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
			for _, rune := range word {
				bw.WriteBool(rune == '1')
			}
			c = ""
		} 
	}

	if leftover > 0 {
		for _, rune := range leftoverWord {
			bw.WriteBool(rune == '1')
		}
	}
}