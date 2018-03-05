package main

import (
	"net/http"
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
	"./xperimental"
)

var (
	delimiter      = flag.String("delimiter", "-", "char that separates artist / song")
	discardStrings = flag.String("discard-str", "", "comma separated list of chars to discard during ocr")
	ocrOnly        = flag.Bool("gimme-fuel-gimme-fire-gimme-that-which-i-desire", false, "only perform ocr")
	img            = flag.String("img", "https://images-cdn.9gag.com/photo/aYwOdrw_700b_v1.jpg", "playlist image url")
	multiLine      = flag.String("multi-line", "", "listed fields that represent artist / song")
	outputDir      = ""
	reg, _         = regexp.Compile("[^a-zA-Z ]+")
	wg             = sync.WaitGroup{}
	ytResp         = YtResponse{}
)

type YtResponse struct {
	Items []struct {
		ID struct {
			VideoID string `json:"videoId"`
		} `json:"id"`
	} `json:"items"`
}

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
	err := os.Mkdir(outputDir, 0755)
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

	scannedLines := strings.Split(text, "\n")
	if *ocrOnly {
		for _, line := range scannedLines {
			fmt.Println(line)
		}
		os.Exit(0)
	}

	fmt.Print("> processing ")
	if *multiLine == "" {
		for _, song := range scannedLines {
			if strings.Contains(song, *delimiter) && shouldNotDiscard(song) {
				wg.Add(1)
				go fetchSong(song, outputDir)
			}
		}
	} else {
		for _, song := range xperimental.GetSongsFromMultipleLines(scannedLines, *multiLine) {
			wg.Add(1)
			go fetchSong(song, outputDir)
		}
	}
}

// fetchSong retrieves song youtube video id via yt api and downloads it with local youtube-dl cli tool
func fetchSong(name, outputDir string) {
	defer fmt.Print(".")
	defer wg.Done()

	resp, err := http.Get(fmt.Sprintf("https://www.googleapis.com/youtube/v3/search?q=%v&maxResults=1&part=snippet&key=AIzaSyCURl1CVR_-gL227h5S8GIhtGXU7kMIFvc", url.QueryEscape(reg.ReplaceAllString(name, ""))))
	if err == nil && resp.Body != nil {
		json.NewDecoder(resp.Body).Decode(&ytResp)
		resp.Body.Close()
	}

	if len(ytResp.Items) > 0 {
		cmd := exec.Command("youtube-dl", "-x", "--audio-format", "mp3", fmt.Sprintf("https://www.youtube.com/watch?v=%v", ytResp.Items[0].ID.VideoID))
		cmd.Dir = outputDir
		cmd.CombinedOutput()
	}
}

// shouldNotDiscard checks if discard-str is provided
// it returns bool if song is valid to process
func shouldNotDiscard(song string) bool {
	if *discardStrings == "" {
		return true
	}
	discardString := strings.Split(*discardStrings, ",")
	for _, stringToDiscard := range discardString {
		if strings.Contains(song, stringToDiscard) {
			return false
		}
	}
	return true
}

func main() {
	flag.Parse()
	outputDir := prepareForFetch(*img)
	getSongsFromImage(outputDir)
}
