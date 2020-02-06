package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/gookit/color"
	"github.com/victorgama/tlvp/cmd"
	"github.com/victorgama/tlvp/printer"
	"github.com/victorgama/tlvp/tlv"
)

var help = flag.Bool("help", false, "Displays help")
var describe = flag.Bool("describe", false, "Shows tags description")
var noColor = flag.Bool("no-color", false, "Disables color output")
var forceColor = flag.Bool("force-color", false, "Forces color output")
var isPDOL = flag.Bool("pdol", false, "Reads input as a PDOL")

func main() {
	flag.Parse()

	if *help {
		printHelp()
		return
	}

	if *forceColor {
		color.ForceOpenColor()
	}

	if *noColor {
		color.Disable()
	}

	var (
		data []byte
		err error
		reader io.Reader
	)

	if cmd.FromStdin() {
		data, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			color.Error.Tips("tlvp found an error processing input: %s", err)
			os.Exit(1)
		}
		reader = bytes.NewReader(data)
	}

	if flag.NArg() == 1 {
		filePath := flag.Arg(0)
		_, err := os.Stat(filePath)
		if os.IsNotExist(err) {
			color.Error.Tips("tlvp could not find a file named \"%s\". Please ensure the path is correct and try again.", filePath)
			os.Exit(1)
		}
		if err != nil {
			color.Error.Tips("tlvp found an error opening \"%s\": %s", filePath, err)
			os.Exit(1)
		}
		data, err = ioutil.ReadFile(filePath)
		if err != nil {
			color.Error.Tips("tlvp found an error processing %s: %s", filePath, err)
			os.Exit(1)
		}
		reader = bytes.NewReader(data)
	}

	if reader == nil {
		printHelp()
		os.Exit(1)
	}

	parser, err := tlv.NewParser(reader, *isPDOL)
	if err != nil {
		color.Error.Tips("Error parsing: %s", err)
		os.Exit(1)
	}


	entries, err := parser.Parse()
	if err != nil {
		color.Error.Tips("Error parsing: %s", err)
		os.Exit(1)
	}

	printer.Print(entries, *describe)
}

func printHelp() {
	fmt.Printf(`tlvp is a TLV parser for EMV data

Usage:

	tlvp [--describe] [--force-color|--no-color] [--help] [input]

	--describe:
		Outputs extra descriptions of tags

	--force-color:
		Forces color formatting

	--help:
		Displays this message

	--no-color:
		Disables color formatting

	--pdol:
		Reads data as a PDOL.

	[input]:
		Path to a file to be parsed. Standard input can also be used by piping 
		data into tlvp.

Bug reports:
	Please file an issue: https://github.com/victorgama/tlvp/issues/new

License:
	MIT. Copyright (C) 2020 - Victor Gama (hey@vito.io)
`)
}
