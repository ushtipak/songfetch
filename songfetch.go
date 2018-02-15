package main

import (
	"net/http"
	"crypto/tls"
	"log"
	"os"
	"runtime"
	"strings"
	"encoding/json"
	"fmt"
	"os/exec"
	"net/url"
	"regexp"
)

var logDebug, logError *log.Logger

// setupLogs sets debug and error logs
func setupLogs() {
	logDebug = log.New(os.Stderr, "DEBUG ", log.Ldate|log.Ltime)
	logError = log.New(os.Stderr, "ERROR ", log.Ldate|log.Ltime)
}

// getTextFromImage uses OCRSpace API to get text from given image.
// It removes special chars from retrieved data and returns string slice of hopefully proper songs :)
func getTextFromImage(imageURL string) []string {
	pc, _, _, _ := runtime.Caller(0)
	funcName := strings.Split(runtime.FuncForPC(pc).Name(), ".")[1]
	logDebug.Printf("%v | Enter -- url [%v]", funcName, imageURL)
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	// compile regex to clean special characters from song (ease up work for youtube/v3 api :)
	reg, err := regexp.Compile("[^a-zA-Z ]+")
	if err != nil {
		logError.Printf("%v | regexp.Compile FAIL [%v]", funcName, err)
	}

	type OCRSpaceResponse struct {
		ParsedResults []struct {
			ParsedText string `json:"ParsedText"`
		} `json:"ParsedResults"`
	}
	var resp OCRSpaceResponse

	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		logError.Printf("%v | http.NewRequest FAIL [%v]", funcName, err)
	}
	if response, err := client.Do(req); err == nil {
		if response.Body != nil {
			if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
				logError.Printf("%v | json.NewDecoder FAIL [%v]", funcName, err)
			}
			response.Body.Close()
		}
	} else {
		logError.Printf("%v | client.Do FAIL [%v]", funcName, err)
	}

	var songs []string
	for _, val := range strings.Split(fmt.Sprintf("%v", resp.ParsedResults), "\n") {
		if strings.Contains(val, "-") {
			songs = append(songs, reg.ReplaceAllString(val, ""))
		}
	}
	logDebug.Printf("%v | Return -- songs [%v ... ]", funcName, songs[0][:20])
	return songs
}

// getYouTubeLink uses YouTube Data API to get IDs of videos from given songs.
// It doesn't handle missing data from API response, cause, yeah, happy path all the way :)
// It returns string slice of most prominent IDs
func getYouTubeLinks(songs []string) []string {
	pc, _, _, _ := runtime.Caller(0)
	funcName := strings.Split(runtime.FuncForPC(pc).Name(), ".")[1]
	logDebug.Printf("%v | Enter -- songs [%v ... ]", funcName, songs[0][:20])
	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}

	type YouTubeResponse struct {
		Items []struct {
			ID struct {
				VideoID string `json:"videoId"`
			} `json:"id"`
		} `json:"items"`
	}
	var resp YouTubeResponse

	var ids []string
	for _, song := range songs {
		var youTubeAPIURL = fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?q=%v&maxResults=1&part=snippet&key=AIzaSyCURl1CVR_-gL227h5S8GIhtGXU7kMIFvc", url.QueryEscape(song))
		logDebug.Printf("%v | url [%v]", funcName, youTubeAPIURL)

			req, err := http.NewRequest("GET", youTubeAPIURL, nil)
			if err != nil {
				logError.Printf("%v | http.NewRequest FAIL [%v]", funcName, err)
			}
			if response, err := client.Do(req); err == nil {
				if response.Body != nil {
					if err := json.NewDecoder(response.Body).Decode(&resp); err != nil {
						logError.Printf("%v | json.NewDecoder FAIL [%v]", funcName, err)
					}
					response.Body.Close()
				}
			} else {
				logError.Printf("%v | client.Do FAIL [%v]", funcName, err)
			}
			ids = append(ids, fmt.Sprintf("%v", resp.Items[0].ID.VideoID))
	}

	logDebug.Printf("%v | Return -- ids [%v ...]", funcName, ids)
	return ids
}

// fetchSong uses local youtube-dl app to download video and extract audio from given IDs
func fetchSong(ids []string) {
	pc, _, _, _ := runtime.Caller(0)
	funcName := strings.Split(runtime.FuncForPC(pc).Name(), ".")[1]
	logDebug.Printf("%v | Enter -- ids [%v]", funcName, ids)

	for _, id := range ids {
		var youTubeVideoURL= fmt.Sprintf("https://www.youtube.com/watch?v=%v", id)
		logDebug.Printf("%v | url [%v]", funcName, youTubeVideoURL)
		if _, err := exec.Command("youtube-dl", "-x", "--audio-format", "mp3", youTubeVideoURL).CombinedOutput(); err != nil {
			logError.Printf("%v | exec.Command FAIL [%v]", funcName, err)
		}
	}
}

func main() {
	setupLogs()
	imageURL := "https://api.ocr.space/parse/imageurl?apikey=e36149dd0488957&url=https://images-cdn.9gag.com/photo/aYwOdrw_700b_v1.jpg"
	songs := getTextFromImage(imageURL)
	ids := getYouTubeLinks(songs)
	fetchSong(ids)
}
