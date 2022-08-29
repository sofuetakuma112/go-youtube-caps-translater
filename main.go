package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/joho/godotenv"
	"golang.org/x/exp/slices"

	"os/exec"
	"regexp"
)

type ReqClient struct {
	Client Client `json:"client"`
}

type Client struct {
	Hl            string `json:"hl"`
	Gl            string `json:"gl"`
	ClientName    string `json:"clientName"`
	ClientVersion string `json:"clientVersion"`
}

type ReqBody struct {
	Context ReqClient `json:"context"`
	Params  string    `json:"params"`
}

type ResponseContext struct {
	VisitorData                     string                          `json:"visitorData"`
	ServiceTrackingParams           ServiceTrackingParams           `json:"serviceTrackingParams"`
	MainAppWebResponseContext       MainAppWebResponseContext       `json:"mainAppWebResponseContext"`
	WebResponseContextExtensionData WebResponseContextExtensionData `json:"webResponseContextExtensionData"`
}

type ServiceTrackingParams []struct {
	Service string `json:"service"`
	Params  Params `json:"params"`
}

type Params []struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type MainAppWebResponseContext struct {
	LoggedOut bool `json:"loggedOut"`
}

type WebResponseContextExtensionData struct {
	HasDecorated bool `json:"hasDecorated"`
}

type Action struct {
	ClickTrackingParams         string                      `json:"clickTrackingParams"`
	UpdateEngagementPanelAction UpdateEngagementPanelAction `json:"updateEngagementPanelAction"`
}

type Actions []Action

type UpdateEngagementPanelAction struct {
	TargetId string  `json:"targetId"`
	Content  Content `json:"content"`
}

type Content struct {
	TranscriptRenderer TranscriptRenderer `json:"transcriptRenderer"`
}

type Body struct {
	TranscriptBodyRenderer TranscriptBodyRenderer `json:"transcriptBodyRenderer"`
}

type TranscriptBodyRenderer struct {
	CueGroups CueGroups `json:"cueGroups"`
}

type CueGroups []struct {
	TranscriptCueGroupRenderer TranscriptCueGroupRenderer `json:"transcriptCueGroupRenderer"`
}

type TranscriptCueGroupRenderer struct {
	FormattedStartOffset FormattedStartOffset `json:"formattedStartOffset"`
	Cues                 Cues                 `json:"cues"`
}

type FormattedStartOffset struct {
	SimpleText string `json:"simpleText"`
}

type Cues []struct {
	TranscriptCueRenderer TranscriptCueRenderer `json:"transcriptCueRenderer"`
}

type TranscriptCueRenderer struct {
	Cue           Cue    `json:"cue"`
	StartOffsetMs string `json:"startOffsetMs"`
	DurationMs    string `json:"durationMs"`
}

type Cue struct {
	SimpleText string `json:"simpleText"`
}

type Footer struct {
	TranscriptFooterRenderer `json:"transcriptFooterRenderer"`
}

type TranscriptFooterRenderer struct {
	LanguageMenu LanguageMenu `json:"languageMenu"`
}

type LanguageMenu struct {
	SortFilterSubMenuRenderer SortFilterSubMenuRenderer `json:"sortFilterSubMenuRenderer"`
}

type SortFilterSubMenuRenderer struct {
	SubMenuItems   SubMenuItems `json:"subMenuItems"`
	TrackingParams string       `json:"trackingParams"`
}

type SubMenuItems []struct {
	Title        string       `json:"title"`
	Selected     bool         `json:"selected"`
	Continuation Continuation `json:"continuation"`
}

type Continuation struct {
	ReloadContinuationData ReloadContinuationData `json:"reloadContinuationData"`
}

type ReloadContinuationData struct {
	Continuation        string `json:"continuation"`
	ClickTrackingParams string `json:"clickTrackingParams"`
}

type TranscriptRenderer struct {
	Body           Body   `json:"body"`
	Footer         Footer `json:"footer"`
	TrackingParams string `json:"trackingParams"`
}

type ResTranscriptAPI struct {
	ResponseContext ResponseContext `json:"responseContext"`
	Actions         Actions         `json:"actions"`
	TrackingParams  string          `json:"trackingParams"`
}

type Caption struct {
	From string `json:"from"`
	To   string `json:"to"`
	Text string `json:"text"`
}

type VideoListResponse struct {
	Kind     string      `json:"kind"`
	Etag     string      `json:"etag"`
	Items    []VideoItem `json:"items"`
	PageInfo PageInfo    `json:"pageInfo"`
}

type PageInfo struct {
	TotalResults   int `json:"totalResults"`
	ResultsPerPage int `json:"resultsPerPage"`
}

type VideoItem struct {
	Kind           string          `json:"kind"`
	Etag           string          `json:"etag"`
	Id             string          `json:"id"`
	ContentDetails VideoItemDetail `json:"contentDetails"`
}

type VideoItemDetail struct {
	Duration        string      `json:"duration"`
	Dimension       string      `json:"dimension"`
	Definition      string      `json:"definition"`
	Caption         string      `json:"caption"`
	LicensedContent bool        `json:"licensedContent"`
	ContentRating   interface{} `json:"contentRating"`
	Projection      string      `json:"projection"`
}

func generateLangParams(lang, subType, subVariant string) string {
	arr := []uint8{0x0a, uint8(utf8.RuneCountInString(subType))}
	arr = append(arr, []byte(subType)...)
	arr = append(arr, 0x12, uint8(utf8.RuneCountInString(lang)))
	arr = append(arr, []byte(lang)...)
	arr = append(arr, 0x1a, uint8(utf8.RuneCountInString(subVariant)))
	arr = append(arr, []byte(subVariant)...)

	return url.QueryEscape(b64.StdEncoding.EncodeToString(arr))
}

func fetchTranscription(params string) ResTranscriptAPI {
	reqBody := &ReqBody{
		Context: ReqClient{
			Client: Client{
				Hl:            "en",
				Gl:            "US",
				ClientName:    "WEB",
				ClientVersion: "2.20210101",
			},
		},
		Params: params,
	}

	e, err := json.Marshal(reqBody)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequest(
		"POST",
		"https://www.youtube.com/youtubei/v1/get_transcript?key=AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8",
		bytes.NewBuffer(e),
	)
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json") // Content-Type 設定

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()

	body, _ := ioutil.ReadAll(res.Body)

	var fetchedCaps ResTranscriptAPI
	json.Unmarshal(body, &fetchedCaps)

	return fetchedCaps
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
				Text: strings.Trim(text, " "),
			})
		}
	}

	videDuration_mili := int(fetchVideoLen(videoId))

	var result Captions
	for i, v := range captions {
		if len(captions)-1 == i {
			result = append(result, &Caption{
				From: v.From,
				To:   ms2likeISOFormat(videDuration_mili),
				Text: v.Text,
			})
		} else {
			result = append(result, &Caption{
				From: v.From,
				To:   captions[i+1].From,
				Text: v.Text,
			})
		}
	}

	return result.Where(func(c *Caption) bool {
		return c.Text != "[Music]"
	})
}

func fetchVideoLen(videoId string) float64 {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	// .envの SAMPLE_MESSAGEを取得して、messageに代入します。
	apiKey := os.Getenv("YOUTUBE_DATA_API_KEY")

	// 動画の長さを取得する
	url := fmt.Sprintf("https://www.googleapis.com/youtube/v3/videos?id=%s&key=%s&part=contentDetails", videoId, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()

	byteArray, _ := ioutil.ReadAll(resp.Body)

	var videoDetail VideoListResponse
	json.Unmarshal(byteArray, &videoDetail)

	duration_str := videoDetail.Items[0].ContentDetails.Duration

	duration_sec, err := ParseDuration(duration_str)
	if err != nil {
		panic(err)
	}

	return duration_sec * 1000
}

// ParseDuration parses an ISO 8601 string representing a duration,
// and returns the resultant golang time.Duration instance.
func ParseDuration(isoDuration string) (float64, error) {
	re := regexp.MustCompile(`^P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)D)?T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:.\d+)?)S)?$`)
	matches := re.FindStringSubmatch(isoDuration)
	if matches == nil {
		return 0, errors.New("input string is of incorrect format")
	}

	seconds := 0.0

	//skipping years and months

	//days
	if matches[3] != "" {
		f, err := strconv.ParseFloat(matches[3], 32)
		if err != nil {
			return 0, err
		}

		seconds += (f * 24 * 60 * 60)
	}
	//hours
	if matches[4] != "" {
		f, err := strconv.ParseFloat(matches[4], 32)
		if err != nil {
			return 0, err
		}

		seconds += (f * 60 * 60)
	}
	//minutes
	if matches[5] != "" {
		f, err := strconv.ParseFloat(matches[5], 32)
		if err != nil {
			return 0, err
		}

		seconds += (f * 60)
	}
	//seconds & milliseconds
	if matches[6] != "" {
		f, err := strconv.ParseFloat(matches[6], 32)
		if err != nil {
			return 0, err
		}

		seconds += f
	}

	return seconds, nil
}

// FormatDuration returns an ISO 8601 duration string.
func FormatDuration(dur time.Duration) string {
	return "PT" + strings.ToUpper(dur.Truncate(time.Millisecond).String())
}

func ms2likeISOFormat(ms int) string {
	nano := ms * 1000000

	t := time.Date(1970, time.January, 1, 0, 0, 0, nano, time.UTC)
	format := "2006-01-02T15:04:05.999Z"
	iso := t.UTC().Format(format)

	if len([]rune(iso)) != len([]rune(format)) {
		idx := strings.Index(iso, ".")
		if idx == -1 {
			// 2006-01-02T15:04:05Z
			iso = iso[:len([]rune(iso))-1] + ".000Z"
		} else {
			// 2006-01-02T15:04:05.0Z
			// 2006-01-02T15:04:05.00Z
			// 2006-01-02T15:04:05.000Z
			mili_str := iso[idx+1 : len([]rune(iso))-1]
			for {
				if len(mili_str) == 3 {
					break
				}
				mili_str = mili_str + "0"
			}
			iso = iso[:idx] + "." + mili_str + "Z"
		}
	}

	trimmedIso := iso[8 : len([]rune(iso))-1]
	day_str := trimmedIso[0:2]
	day, _ := strconv.Atoi(day_str)
	dayStartFromZero := fmt.Sprintf("%02d", day-1)
	isoOnlyTime := trimmedIso[3:]
	return dayStartFromZero + ":" + isoOnlyTime
}

func generateTranscriptParams(videoId, langParams string) string {
	if langParams == "" {
		arr := []uint8{0x0a, 0x0b}
		arr = append(arr, []byte(videoId)...)
		return url.QueryEscape(b64.StdEncoding.EncodeToString(arr))
	} else {
		arr := []uint8{0x0a, 0x0b}
		arr = append(arr, []byte(videoId)...)
		arr = append(arr, 0x12, uint8(utf8.RuneCountInString(langParams)))
		arr = append(arr, []byte(langParams)...)
		return url.QueryEscape(b64.StdEncoding.EncodeToString(arr))
	}
}

func likeIso2Float(likeIso string) float64 {
	splitted := strings.Split(likeIso, ":")
	days, err := strconv.ParseFloat(splitted[0], 64)
	if err != nil {
		panic(err)
	}

	hours, err := strconv.ParseFloat(splitted[1], 64)
	if err != nil {
		panic(err)
	}

	minutes, err := strconv.ParseFloat(splitted[2], 64)
	if err != nil {
		panic(err)
	}

	seconds, err := strconv.ParseFloat(splitted[3], 64)
	if err != nil {
		panic(err)
	}

	return days*24*60*60 + hours*60*60 + minutes*60 + seconds
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

func escapeDot(word string) string {
	return strings.Replace(word, ".", "[dot]", -1)
}

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
				Word:      escapeDot(w),
				Timestamp: from_float + float64(i)*secOfBetWords,
			})
		}
	}

	file, _ := json.MarshalIndent(dict, "", " ")
	_ = ioutil.WriteFile(outputDirPath+"/dict.json", file, 0644)

	return dict
}

var outputDirPath string
var videoId string

func init() {
	flag.Parse()
	videoId = flag.Args()[0]

	crrDir, _ := os.Getwd()
	outputDirPath = crrDir + "/captions/" + videoId

	if err := os.MkdirAll(outputDirPath, 0777); err != nil {
		panic(err)
	}
}

func createEscapedText(captions Captions) string {
	var captionTexts []string
	for _, c := range captions {
		captionTexts = append(captionTexts, c.Text)
	}

	mayWords := strings.Split(strings.Join(captionTexts, " "), " ")
	words := []string{}
	for _, w := range mayWords {
		if w != "" {
			words = append(words, escapeDot(w))
		}
	}
	escapedText := strings.Join(words, " ")
	_ = ioutil.WriteFile(outputDirPath+"/textPuncEscaped.txt", []byte(escapedText), 0644)

	return escapedText
}

func readPuncRestoredText(filePath string) string {
	readBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		panic(err)
	}
	return string(readBytes)
}

func capitalizeFirstChar(text string) string {
	return strings.ToUpper(text[:1]) + text[1:]
}

// TODO: "~~~[dot]."への対応
func unescapeDot(word string) string {
	return strings.Replace(word, "[dot]", ".", -1)
}

func createCapsBySentence(puncRestoredText string, dict WordDict) Sentences {
	var wordsBySentence WordDict
	var sentences Sentences
	restoredWords := strings.Split(puncRestoredText, " ")
	for _, rw := range restoredWords {
		dictWord := dict[0].Word
		timestamp := dict[0].Timestamp

		if strings.Index(rw, dictWord) != -1 || strings.Index(rw, capitalizeFirstChar(dictWord)) != -1 {
			indexOfLastChar := len(rw) - 1
			hasPunc := false
			for _, punc := range []string{".", "?"} {
				if strings.Index(rw, punc) == indexOfLastChar {
					hasPunc = true
				}
			}

			wordsBySentence = append(wordsBySentence, &WordWithTimeStamp{
				Word:      rw,
				Timestamp: timestamp,
			})

			if hasPunc { // 直前でappendしたWordWithTimeStampのWordに文末記号が含まれていた
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

func main() {
	fetchedCaps := fetchTranscription(generateTranscriptParams(videoId, generateLangParams("en", "asr", "")))
	captions := formatCaptions(fetchedCaps, videoId)
	dict := createDict(captions)
	createEscapedText(captions)

	puncRestoredTextFilePath := outputDirPath + "/textPuncEscapedAndRestored.txt"
	_, err := os.Stat(puncRestoredTextFilePath)
	if os.IsNotExist(err) {
		err := exec.Command("python3", "repunc.py", videoId).Run()
		if err != nil {
			panic(err)
		}
	}

	puncRestoredText := readPuncRestoredText(puncRestoredTextFilePath)

	sentences := createCapsBySentence(puncRestoredText, dict)
}
