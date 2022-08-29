package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"github.com/joho/godotenv"
)

func fetchVideoLen(videoId string) float64 {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

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