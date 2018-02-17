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
	"io"
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

// verifyStep checks for errors in execution, updates output accordingly and exists if needed
func verifyStep(err error) {
	if err != nil {
		fmt.Println("FAIL")
		log.Fatal(err)
	} else {
		fmt.Println("done")
	}
}

// prepareForFetch creates output dir and fetches image to source-img
// it returns output dir string
func prepareForFetch(img string) string {
	outputDir = "songfetch" + time.Now().Format("--2006-01-02--15-04-05--") + strings.ToLower(reg.ReplaceAllString(img, "-"))

	fmt.Print("> create output dir ... ")
	err := os.Mkdir(outputDir, 0777)
	verifyStep(err)

	fmt.Print("> retrieve img content ... ")
	resp, err := http.Get(img)
	verifyStep(err)
	defer resp.Body.Close()

	fmt.Print("> create fd ... ")
	f, err := os.Create(outputDir + "/source-img")
	verifyStep(err)
	defer f.Close()

	fmt.Print("> write img to file ... ")
	_, err = io.Copy(f, resp.Body)
	verifyStep(err)

	return outputDir
}

// getSongsFromImage does ocr on fetched img and calls fetchSong on each line of text with target delimiter
func getSongsFromImage(outputDir string) {
	defer fmt.Println("> all done \\o/")
	defer fmt.Println(" done")
	defer wg.Wait()

	fmt.Print("> perform img ocr ... ")
	client := gosseract.NewClient()
	defer client.Close()
	client.SetImage(outputDir + "/source-img")
	text, err := client.Text()
	verifyStep(err)

	fmt.Print("> processing ")
	songs := strings.Split(text, "\n")
	for _, song := range songs {
		if strings.Contains(song, *delimiter) {
			wg.Add(1)
			go fetchSong(song, outputDir)
		}
	}
}

// fetchSong retrieves song youtube video id via yt api and downloads it with local youtube-dl cli tool
func fetchSong(song, outputDir string) {
	defer fmt.Print(".")
	defer wg.Done()

	req, _ := http.NewRequest("GET", fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?q=%v&maxResults=1&part=snippet&key=AIzaSyCURl1CVR_-gL227h5S8GIhtGXU7kMIFvc", url.QueryEscape(reg.ReplaceAllString(song, ""))), nil)
	if resp, err := client.Do(req); err == nil {
		if resp.Body != nil {
			json.NewDecoder(resp.Body).Decode(&ytResp)
			resp.Body.Close()
		}
	}

	if len(ytResp.Items) > 0 {
		id := ytResp.Items[0].ID.VideoID
		if id != "" {
			cmd := exec.Command("youtube-dl", "-x", "--audio-format", "mp3", fmt.Sprintf("https://www.youtube.com/watch?v=%v", id))
			cmd.Dir = outputDir
			cmd.CombinedOutput()
		}
	}
}

func main() {
	flag.Parse()
	outputDir := prepareForFetch(*img)
	getSongsFromImage(outputDir)
}
