package xperimental

import "strings"

// GetSongsFromMultipleLines parses --multi-line argument and matches lines returned from ocr
// It returns slice of potential artist / song pairs strings
func GetSongsFromMultipleLines(scannedLanes []string, multiLine string) []string {
	mapping := strings.Split(multiLine, ",")
	artistStr, songStr := mapping[0], mapping[1]

	var artists, songs []string
	for _, line := range scannedLanes {
		if strings.Contains(line, artistStr) {
			artists = append(artists, strings.Split(line, artistStr)[1])
		}
		if strings.Contains(line, songStr) {
			songs = append(songs, strings.Split(line, songStr)[1])
		}
	}

	var artistSongCombos []string
	for _, artist := range artists {
		for _, song := range songs {
			artistSongCombos = append(artistSongCombos, artist+song)
		}
	}
	return artistSongCombos
}
