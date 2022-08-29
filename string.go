package main

import (
	"strings"
)

func capitalizeFirstChar(text string) string {
	return strings.ToUpper(text[:1]) + text[1:]
}

func escapeDot(word string) string {
	return strings.Replace(word, ".", "[dot]", -1)
}

// TODO: "~~~[dot]."への対応
func unescapeDot(word string) string {
	return strings.Replace(word, "[dot]", ".", -1)
}
