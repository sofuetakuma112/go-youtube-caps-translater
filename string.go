package main

import (
	"strings"
	"unicode/utf8"
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

func splitSentenceIfLong(sentence string) []string {
	// 一文の文字数が55文字を超えるかどうかチェック
	if utf8.RuneCountInString(sentence) <= 55 {
		return []string{sentence}
	}

	// 「、」で分割
	parts := strings.Split(sentence, "、")

	// 最適な分割点を見つける（できるだけ均等に分割）
	bestIndex := -1
	for i := range parts {
		// 前半の長さを計算
		length := utf8.RuneCountInString(strings.Join(parts[:i+1], "、"))
		if length > 55 {
			break
		}
		bestIndex = i
	}

	// 分割点が見つからない場合は、そのまま返す
	if bestIndex == -1 {
		return []string{sentence}
	}

	// 文を適切な場所で2行に分割
	before := strings.Join(parts[:bestIndex+1], "、") + "、"
	after := strings.Join(parts[bestIndex+1:], "、")
	return []string{before, after}
}
