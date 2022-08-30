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
	"sync"

	"golang.org/x/exp/slices"

	"github.com/cheggaaa/pb/v3"
	"os/exec"
)

type Caption struct {
	From string `json:"from"`
	To   string `json:"to"`
	Text string `json:"text"`
}

type Captions []*Caption

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

// 内部APIから取得したデータから字幕情報を抽出する
func extractCaptions(transcript ResTranscriptAPI) Captions {
	var rawCaptions Captions
	path := outputDirPath+"/rawCaptions.json"
	if checkFileExist(path) {
		readBytes, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}

		json.Unmarshal(readBytes, &rawCaptions)
		return rawCaptions
	}

	actions := transcript.Actions

	idx := slices.IndexFunc(actions, func(action Action) bool {
		return action.UpdateEngagementPanelAction.TargetId == "engagement-panel-transcript"
	})
	if idx == -1 {
		panic(errors.New("no target id"))
	}

	cueGroups := transcript.Actions[idx].UpdateEngagementPanelAction.Content.TranscriptRenderer.Body.TranscriptBodyRenderer.CueGroups

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
				fmt.Printf("no simpleText id, cueRenderer: %v, last caption: %v\n", cueRenderer, *rawCaptions[len(rawCaptions)-1])
				continue
			}

			rawCaptions = append(rawCaptions, &Caption{
				From: ms2likeISOFormat(start_ms),
				To:   ms2likeISOFormat(end_ms),
				Text: strings.Trim(text, " "), // ここでトリム
			})
		}
	}

	file, _ := json.MarshalIndent(rawCaptions, "", " ")
	_ = ioutil.WriteFile(path, file, 0644)

	return rawCaptions
}

// 抽出した字幕データを整形する
func formatCaptions(rawCaptions Captions, videoId string) Captions {
	var formattedCaps Captions

	path := outputDirPath+"/formattedCaptions.json"
	if checkFileExist(path) {
		readBytes, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}

		json.Unmarshal(readBytes, &formattedCaps)
		return formattedCaps
	}

	videoDuration_mili := int(fetchVideoLen(videoId))

	var formattedWords []string
	for i, c := range rawCaptions {
		// フィルタリング
		if c.Text == "[Music]" {
			continue
		}

		// 字幕テキストの処理
		idx := strings.Index(c.Text, ".")
		newText := c.Text
		if idx == len(c.Text)-1 { // 末尾がピリオド
			newText = c.Text[:len(c.Text)-1]
		} else if idx > 0 && (string(c.Text[idx-1]) == " " || string(c.Text[idx+1]) == " ") { // "aa. aa" or "aa .aa"のケース
			newText = c.Text[0:idx] + c.Text[idx+1:]
		}

		caption := &Caption{
			From: c.From,
			To:   ms2likeISOFormat(videoDuration_mili),
			Text: newText,
		}

		if len(rawCaptions)-1 == i {
			caption.To = ms2likeISOFormat(videoDuration_mili)
		} else {
			caption.To = rawCaptions[i+1].From
		}

		formattedWords = append(formattedWords, strings.Split(newText, " ")...)
		formattedCaps = append(formattedCaps, caption)
	}

	formattedText := strings.ToLower(strings.Join(formattedWords, " "))
	_ = ioutil.WriteFile(outputDirPath+"/"+escapedPuncTxtName, []byte(formattedText), 0644)

	file, _ := json.MarshalIndent(formattedCaps, "", " ")
	_ = ioutil.WriteFile(path, file, 0644)

	return formattedCaps
}

func createDict(captions Captions) WordDict {
	var dict WordDict
	path := outputDirPath + "/dict.json"

	if checkFileExist(path) {
		readBytes, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}

		json.Unmarshal(readBytes, &dict)
		return dict
	}

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
	_ = ioutil.WriteFile(path, file, 0644)

	return dict
}

func groupBySentence(puncRestoredText string, dict WordDict) Sentences {
	var wordsBySentence WordDict
	var sentences Sentences

	path := outputDirPath + "/captions_en_by_sentence.json"

	if checkFileExist(path) {
		readBytes, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}

		json.Unmarshal(readBytes, &sentences)
		return sentences
	}

	restoredWords := strings.Split(puncRestoredText, " ")
	for i, rw := range restoredWords {
		dictWord := dict[0].Word
		timestamp := dict[0].Timestamp

		if strings.Index(strings.ToLower(rw), strings.ToLower(dictWord)) != -1 {
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
			panic(errors.New(fmt.Sprintf("strings.ToLower(rw): %v, strings.ToLower(dictWord): %v, timestamp: %v", strings.ToLower(rw), strings.ToLower(dictWord), timestamp)))
		}

		dict = dict[1:]
	}

	file, _ := json.MarshalIndent(sentences, "", " ")
	_ = ioutil.WriteFile(path, file, 0644)

	return sentences
}

func translateSentences(sentences Sentences) Sentences {
	var jpSentences Sentences

	path := outputDirPath + "/captions_ja_by_sentence.json"

	if checkFileExist(path) {
		readBytes, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}

		json.Unmarshal(readBytes, &jpSentences)
		return jpSentences
	}

	bar := pb.Simple.Start(len(sentences))
	bar.SetMaxWidth(80)

	var wg sync.WaitGroup
	var mu sync.Mutex
	jpSentences = make(Sentences, len(sentences))
	semaphore := make(chan struct{}, 10)
	for i, s := range sentences {
		semaphore <- struct{}{}
		wg.Add(1)
		go func(i int, s Sentence) {
			defer func() {
				<-semaphore
				bar.Increment()
				wg.Done()
			}()
			translatedText := translate(s.Sentence).Text
			jpSentence := Sentence{
				Sentence: translatedText,
				From:     s.From,
				To:       s.To,
			}
			mu.Lock()
			jpSentences[i] = jpSentence
			mu.Unlock()
		}(i, s)
	}
	wg.Wait()
	bar.Finish()

	file, _ := json.MarshalIndent(jpSentences, "", " ")
	_ = ioutil.WriteFile(path, file, 0644)

	return jpSentences
}

func createSrt(jpSentences Sentences) {
	srt := ""

	path := outputDirPath + "/captions_ja.srt"
	if checkFileExist(path) {
		return
	}

	for i, js := range jpSentences {
		jpText := js.Sentence
		from := js.From
		to := js.To

		srt += fmt.Sprintf("%v\n%v --> %v\n%v\n\n", i+1, strings.Replace(from, ".", ",", 1), strings.Replace(to, ".", ",", 1), jpText)
	}
	_ = ioutil.WriteFile(path, []byte(srt), 0644)
}

func repunc() string {
	if !checkFileExist(puncRestoredTextFilePath) {
		err := exec.Command("python3", "repunc_by_nemo.py", videoId, escapedPuncTxtName, restoredPuncTxtName).Run()
		if err != nil {
			panic(err)
		}
	}

	readBytes, err := ioutil.ReadFile(puncRestoredTextFilePath)
	if err != nil {
		panic(err)
	}
	return string(readBytes)
}

var outputDirPath string
var videoId string

var escapedPuncTxtName string
var restoredPuncTxtName string

var puncRestoredTextFilePath string

func init() {
	flag.Parse()
	videoId = flag.Args()[0]

	crrDir, _ := os.Getwd()
	outputDirPath = crrDir + "/captions/" + videoId

	if err := os.MkdirAll(outputDirPath, 0777); err != nil {
		panic(err)
	}

	escapedPuncTxtName = "formatted_captions.txt"
	restoredPuncTxtName = "textPuncEscapedAndRestored.txt"
	puncRestoredTextFilePath = outputDirPath + "/" + restoredPuncTxtName
}

func main() {
	fmt.Println("Step: 1/8")
	fetchedTranscript := fetchTranscription(generateTranscriptParams(videoId, generateLangParams("en", "asr", "")))

	fmt.Println("Step: 2/8")
	rawCaptions := extractCaptions(fetchedTranscript)

	fmt.Println("Step: 3/8")
	captions := formatCaptions(rawCaptions, videoId)

	fmt.Println("Step: 4/8")
	dict := createDict(captions)

	fmt.Println("Step: 5/8")
	puncRestoredText := repunc()

	fmt.Println("Step: 6/8")
	sentences := groupBySentence(puncRestoredText, dict)

	os.Exit(1)

	fmt.Println("Step: 7/8")
	jpSentences := translateSentences(sentences)

	fmt.Println("Step: 8/8")
	createSrt(jpSentences)
}
