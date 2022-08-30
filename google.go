package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type ResGoogleTranslate struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

func translate(text string) ResGoogleTranslate {
	url := fmt.Sprintf("https://script.google.com/macros/s/AKfycbwHvOCeufro86JCbI8pZh_XdDXahWLv8tvmqhC_jfYkEXMtm00N6o-pzU5D0bTvGZLfDA/exec?text=%v&source=en&target=ja", url.QueryEscape(text))

	var res ResGoogleTranslate

	for i := 0; i < 5; i++ {
		resp, err := http.Get(url)
		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()

		byteArray, _ := ioutil.ReadAll(resp.Body)

		json.Unmarshal(byteArray, &res)

		if res.Code == 200 {
			break
		}
	}

	if res.Code != 200 {
		panic(errors.New("翻訳に失敗"))
	}

	return res
}
