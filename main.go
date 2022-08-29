package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"

	"os/exec"
)

type Caption struct {
	From string `json:"from"`
	To   string `json:"to"`
	Text string `json:"text"`
}

type Captions []*Caption

func (captions Captions) Where(fn func(*Caption) bool) (result Captions) {
	for _, c := range captions {
		if fn(c) {
			result = append(result, c)
		}
	}
	return result
}

// 内部APIから取得した字幕データを整形する
func formatCaptions(transcript ResTranscriptAPI, videoId string) Captions {
	actions := transcript.Actions

	idx := slices.IndexFunc(actions, func(action Action) bool {
		return action.UpdateEngagementPanelAction.TargetId == "engagement-panel-transcript"
	})
	if idx == -1 {
		panic(errors.New("no target id"))
	}

	cueGroups := transcript.Actions[idx].UpdateEngagementPanelAction.Content.TranscriptRenderer.Body.TranscriptBodyRenderer.CueGroups

	var captions Captions

	for _, cueGroup := range cueGroups {
		for _, cue := range cueGroup.TranscriptCueGroupRenderer.Cues {
			cueRenderer := cue.TranscriptCueRenderer

			start_ms, err := strconv.Atoi(cueRenderer.StartOffsetMs)
			if err != nil {
				panic(err)
			}

			duration_ms, err := strconv.Atoi(cueRenderer.DurationMs)
			if err != nil {
				panic(err)
			}

			end_ms := start_ms + duration_ms

			text := cueRenderer.Cue.SimpleText
			if text == "" {
				panic(errors.New("no simpleText id"))
			}

			captions = append(captions, &Caption{
				From: ms2likeISOFormat(start_ms),
				To:   ms2likeISOFormat(end_ms),
				Text: strings.Trim(text, " "), // ここでトリム
			})
		}
	}

	videDuration_mili := int(fetchVideoLen(videoId))

	var formattedCaps Captions
	for i, v := range captions {
		caption := &Caption{
			From: v.From,
			To:   ms2likeISOFormat(videDuration_mili),
			Text: v.Text,
		}

		if len(captions)-1 == i {
			caption.To = ms2likeISOFormat(videDuration_mili)
		} else {
			caption.To = captions[i+1].From
		}

		formattedCaps = append(formattedCaps, caption)
	}

	var originalWords []string
	for _, fc := range formattedCaps {
		originalWords = append(originalWords, fc.Text)
	}

	originalText := strings.Join(originalWords, " ")
	_ = ioutil.WriteFile(outputDirPath+"/original_captions.text", []byte(originalText), 0644)

	noMusicCaps := formattedCaps.Where(func(c *Caption) bool {
		return c.Text != "[Music]"
	})

	var formattedWords []string
	var removedPeriodCaps Captions
	for _, c := range noMusicCaps {
		idx := strings.Index(c.Text, ".")

		newText := c.Text
		if idx == len(c.Text) { // 末尾がピリオド
			newText = c.Text[:len(c.Text)-1]
		} else if idx > 0 && (string(c.Text[idx-1]) == " " || string(c.Text[idx+1]) == " ") { // "aa. aa" or "aa .aa"のケース
			newText = c.Text[0:idx] + c.Text[idx+1:]
		}

		formattedWords = append(formattedWords, strings.Split(newText, " ")...)

		removedPeriodCaps = append(removedPeriodCaps, &Caption{
			From: c.From,
			To:   c.To,
			Text: newText,
		})
	}

	formattedText := strings.ToLower(strings.Join(formattedWords, " "))
	_ = ioutil.WriteFile(outputDirPath+"/"+escapedPuncTxtName, []byte(formattedText), 0644)

	return removedPeriodCaps
}

type WordWithTimeStamp struct {
	Word      string  `json:"word"`
	Timestamp float64 `json:"timestamp"`
}

type WordDict []*WordWithTimeStamp

type WordGroup []WordDict

type Sentence struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Sentence string `json:"sentence"`
}

type Sentences []Sentence

func createDict(captions Captions) WordDict {
	var dict WordDict

	for _, c := range captions {
		words := strings.Split(c.Text, " ")
		countOfWords := len(words)

		from_float := likeIso2Float(c.From)
		to_float := likeIso2Float(c.To)

		lenOfTalk := to_float - from_float

		var secOfBetWords float64 = 0
		if countOfWords != 1 {
			secOfBetWords = lenOfTalk / float64(countOfWords-1)
		}

		for i, w := range words {
			dict = append(dict, &WordWithTimeStamp{
				Word:      w,
				Timestamp: from_float + float64(i)*secOfBetWords,
			})
		}
	}

	file, _ := json.MarshalIndent(dict, "", " ")
	_ = ioutil.WriteFile(outputDirPath+"/dict.json", file, 0644)

	return dict
}

func readPuncRestoredText(filePath string) string {
	readBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	return string(readBytes)
}

func groupBySentence(puncRestoredText string, dict WordDict) Sentences {
	var wordsBySentence WordDict
	var sentences Sentences
	restoredWords := strings.Split(puncRestoredText, " ")
	for i, rw := range restoredWords {
		dictWord := dict[0].Word
		timestamp := dict[0].Timestamp

		if strings.Index(rw, dictWord) != -1 || strings.Index(rw, capitalizeFirstChar(dictWord)) != -1 {
			indexOfLastChar := len(rw) - 1

			hasPunc := false
			for _, punc := range []string{".", "?"} {
				if strings.Index(rw, punc) == indexOfLastChar { // 末尾文字が句読点
					hasPunc = true
				}
			}
			// 次の単語が文章の先頭に来る単語なら、現在の単語を文章の末尾単語とする
			isLastWord := false
			for _, firstWord := range []string{"It"} {
				if len(restoredWords)-1 != i && firstWord == restoredWords[i+1] { // 次の単語が文章の先頭に来る単語の場合
					isLastWord = true
				}
			}

			wordsBySentence = append(wordsBySentence, &WordWithTimeStamp{
				Word:      rw,
				Timestamp: timestamp,
			})

			if hasPunc || isLastWord { // 直前でappendしたWordWithTimeStampのWordに文末記号が含まれていた
				// wordsBySentenceを{ from, to, sentence }の形状に変換する
				var words []string
				for _, w := range wordsBySentence {
					words = append(words, w.Word)
				}
				sentence := Sentence{
					From:     ms2likeISOFormat(int(wordsBySentence[0].Timestamp * 1000))[3:],
					To:       ms2likeISOFormat(int(wordsBySentence[len(wordsBySentence)-1].Timestamp * 1000))[3:],
					Sentence: unescapeDot(strings.Join(words, " ")),
				}

				// FIXME: 句読点以外に対応する必要があるかも
				if isLastWord && string(sentence.Sentence[len(sentence.Sentence)-1]) != "." { // 次の単語が先頭単語でかつ現在の文章の末尾に句読点が存在しない
					sentence.Sentence += "."
				}

				sentences = append(sentences, sentence)
				wordsBySentence = nil
			}
		} else {
			panic(errors.New(fmt.Sprintf("rw: %v, dictWord: %v, timestamp: %v", rw, dictWord, timestamp)))
		}

		dict = dict[1:]
	}

	file, _ := json.MarshalIndent(sentences, "", " ")
	_ = ioutil.WriteFile(outputDirPath+"/captions_en_by_sentence.json", file, 0644)

	return sentences
}

func translateSentences(sentences Sentences) Sentences {
	var jpSentences Sentences
	for _, s := range sentences {
		translatedText := translate(s.Sentence).Text

		jpSentence := Sentence{
			Sentence: translatedText,
			From:     s.From,
			To:       s.To,
		}

		// if i != 0 {
		// 	jpSentence.From = sentences[i-1].From
		// }

		jpSentences = append(jpSentences, jpSentence)
	}
	file, _ := json.MarshalIndent(jpSentences, "", " ")
	_ = ioutil.WriteFile(outputDirPath+"/captions_ja_by_sentence.json", file, 0644)

	return jpSentences
}

func createSrt(jpSentences Sentences) {
	srt := ""
	for i, js := range jpSentences {
		jpText := js.Sentence
		from := js.From
		to := js.To

		srt += fmt.Sprintf("%v\n%v --> %v\n%v\n\n", i+1, strings.Replace(from, ".", ",", 1), strings.Replace(to, ".", ",", 1), jpText)
	}
	_ = ioutil.WriteFile(outputDirPath+"/captions_ja.srt", []byte(srt), 0644)
}

func repunc(puncRestoredTextFilePath string) {
	_, err := os.Stat(puncRestoredTextFilePath)
	if os.IsNotExist(err) {
		err := exec.Command("python3", "repunc_by_nemo.py", videoId, escapedPuncTxtName, restoredPuncTxtName).Run()
		if err != nil {
			panic(err)
		}
	}
}

var outputDirPath string
var videoId string
var escapedPuncTxtName string
var restoredPuncTxtName string

func init() {
	flag.Parse()
	videoId = flag.Args()[0]

	crrDir, _ := os.Getwd()
	outputDirPath = crrDir + "/captions/" + videoId

	if err := os.MkdirAll(outputDirPath, 0777); err != nil {
		panic(err)
	}
}

func main() {
	escapedPuncTxtName = "formatted_captions.txt"
	restoredPuncTxtName = "textPuncEscapedAndRestored.txt"

	fetchedCaps := fetchTranscription(generateTranscriptParams(videoId, generateLangParams("en", "asr", "")))
	captions := formatCaptions(fetchedCaps, videoId)
	dict := createDict(captions)

	puncRestoredTextFilePath := outputDirPath + "/" + restoredPuncTxtName

	repunc(puncRestoredTextFilePath)
	puncRestoredText := readPuncRestoredText(puncRestoredTextFilePath)
	sentences := groupBySentence(puncRestoredText, dict)
	jpSentences := translateSentences(sentences)
	createSrt(jpSentences)
}
