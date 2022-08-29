package main

import (
	"bytes"
	b64 "encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"unicode/utf8"
)

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