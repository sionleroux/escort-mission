package main

import (
	"fmt"
	"strings"
	"text/scanner"
)

var stuff = `
1
00:00:00,000 --> 00:00:02,500
Welcome to the Example Subtitle File!

2
00:00:03,000 --> 00:00:06,000
This is a demonstration of SRT subtitles.
`

func main() {
	var s scanner.Scanner
	s.Init(strings.NewReader(stuff))

	for tok := s.Scan(); tok != scanner.EOF; tok = s.Scan() {
		fmt.Println(s.TokenText())
		// Check if it's a number followed by a number? Or not colon?
	}
}
