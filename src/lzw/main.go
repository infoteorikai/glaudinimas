package main

import (
	"flag"
	"fmt"
	"os"
)

var outName = flag.String("out", "-", "output file, use - for stdout")

func check(err error) {
	if err != nil {
		fmt.Fprintln(os.Stdout, "Error:", err)
		os.Exit(1)
	}
}

func main() {
	flag.Parse()

	f := os.Stdout
	var err error
	if *outName != "-" {
		f, err = os.Create(*outName)
		check(err)
	}

	fmt.Fprintln(f, "suglaudinta")
}
