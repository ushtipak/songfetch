package main

import (
	"net/http"
	"crypto/tls"
	"encoding/json"
	"log"
	"sync"
	"strings"
	"os"
	"regexp"
	"fmt"
	"net/url"
	"os/exec"
)

type OCRSpaceResponse struct {
	ParsedResults []struct {
		ParsedText string `json:"ParsedText"`
	} `json:"ParsedResults"`
}

type YtResponse struct {
	Items []struct {
		ID struct {
			VideoID string `json:"videoId"`
		} `json:"id"`
	} `json:"items"`
}

var client = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
var wg sync.WaitGroup

// UPDATE ME XD
func getSongsFromImage(img string) {
	var ocrSpaceResp OCRSpaceResponse
	defer log.Println("all done \\o/")
	defer wg.Wait()

	req, err := http.NewRequest("GET", img, nil)
	if err != nil {
		log.Printf("http.NewRequest FAIL [%v]", err)
	}
	if response, err := client.Do(req); err == nil {
		if response.Body != nil {
			if err := json.NewDecoder(response.Body).Decode(&ocrSpaceResp); err != nil {
				log.Printf("json.NewDecoder FAIL [%v]", err)
			}
			response.Body.Close()
		}
	} else {
		log.Printf("client.Do FAIL [%v]", err)
	}

	log.Println("started")
	songs := strings.Split(ocrSpaceResp.ParsedResults[0].ParsedText, "\n")
	for _, song := range songs {
		if strings.Contains(song, "-") {
			wg.Add(1)
			go fetchSong(song)
		}
	}
}

// UPDATE ME XD
func fetchSong(song string) {
	var ytResp YtResponse
	defer log.Println("finished - " + strings.ToLower(song))
	defer wg.Done()

	// compile regex to clean special characters from song (ease up work for youtube/v3 api :)
	reg, err := regexp.Compile("[^a-zA-Z ]+")
	if err != nil {
		log.Printf("regexp.Compile FAIL [%v]", err)
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?q=%v&maxResults=1&part=snippet&key=AIzaSyCURl1CVR_-gL227h5S8GIhtGXU7kMIFvc", url.QueryEscape(reg.ReplaceAllString(song, ""))), nil)
	if err != nil {
		log.Printf("http.NewRequest FAIL [%v]", err)
	}
	if response, err := client.Do(req); err == nil {
		if response.Body != nil {
			if err := json.NewDecoder(response.Body).Decode(&ytResp); err != nil {
				log.Printf("json.NewDecoder FAIL [%v]", err)
			}
			response.Body.Close()
		}
	} else {
		log.Printf("client.Do FAIL [%v]", err)
	}
	id := ytResp.Items[0].ID.VideoID
	if id != "" {
		cmd := exec.Command("youtube-dl", "-x", "--audio-format", "mp3", fmt.Sprintf("https://www.youtube.com/watch?v=%v", id))
		cmd.Dir = "./temp"
		if _, err := cmd.CombinedOutput(); err != nil {
			log.Printf("exec.Command FAIL [%v]", err)
		}
	}
}

func main() {
	os.Mkdir("./temp", 0777)
	//imageURL := "https://api.ocr.space/parse/imageurl?apikey=e36149dd0488957&url=https://images-cdn.9gag.com/photo/aYwOdrw_700b_v1.jpg"
	imageURL := "http://localhost:8000/output.json"
	getSongsFromImage(imageURL)
}
