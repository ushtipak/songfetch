package main

import (
	"net/http"
	"crypto/tls"
	"log"
	"sync"
	"strings"
	"regexp"
	"fmt"
	"net/url"
	"flag"
	"time"
	"github.com/otiai10/gosseract"
	"os"
	"encoding/json"
	"os/exec"
)

var (
	delimiter = flag.String("delimiter", "-", "char that separates artist / song")
	img       = flag.String("image", "https://images-cdn.9gag.com/photo/aYwOdrw_700b_v1.jpg", "playlist image url")
	outputDir = ""
	reg, _    = regexp.Compile("[^a-zA-Z ]+")
	ytResp    = YtResponse{}
)

type YtResponse struct {
	Items []struct {
		ID struct {
			VideoID string `json:"videoId"`
		} `json:"id"`
	} `json:"items"`
}

var client = &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
var wg sync.WaitGroup

func checkError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// UPDATE ME XD
func getSongsFromImage(img string) {
	defer log.Println("all done \\o/")
	defer wg.Wait()

	client := gosseract.NewClient()
	defer client.Close()
	client.SetImage("./test.png")
	text, err := client.Text()
	checkError(err)

	err = os.Mkdir(outputDir, 0777)
	checkError(err)

	songs := strings.Split(text, "\n")
	for _, song := range songs {
		log.Printf("=> [%v]", song)
		if strings.Contains(song, *delimiter) {
			log.Printf("=> [%v] : YES", song)
			wg.Add(1)
			go fetchSong(song)
		}
	}
}

// UPDATE ME XD
func fetchSong(song string) {
	defer log.Println(strings.ToLower(song))
	defer wg.Done()

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
		cmd.Dir = outputDir
		if _, err := cmd.CombinedOutput(); err != nil {
			log.Printf("exec.Command FAIL [%v]", err)
		}
	}
}

func main() {
	flag.Parse()
	outputDir = "songfetch" + time.Now().Format("--2006-01-02--15-04-05--") + strings.ToLower(reg.ReplaceAllString(*img, "-"))
	getSongsFromImage(*img)
}
