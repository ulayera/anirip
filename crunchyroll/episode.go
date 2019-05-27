package crunchyroll /* import "s32x.com/anirip/crunchyroll" */

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"s32x.com/anirip/common"
)

var (
	formats = map[string]string{
		"android": "107",
		"360":     "106",
		"480":     "106",
		"720":     "106",
		"1080":    "108",
		"default": "0",
	}
	qualities = map[string]string{
		"android": "71",
		"360":     "60",
		"480":     "61",
		"720":     "62",
		"1080":    "80",
		"default": "0",
	}
)

// Episode holds all episode metadata needed for downloading
type Episode struct {
	ID          int
	SubtitleID  int
	Title       string
	Description string
	Number      float64
	Quality     string
	Path        string
	URL         string
	Filename    string
	StreamURL   string
}

// GetEpisodeInfo retrieves and populates the metadata on the Episode
func (e *Episode) GetEpisodeInfo(client *common.HTTPClient, quality string) error {
	e.Quality = quality // Sets the quality to the passed quality string

	// Gets the HTML of the episode page
	// client.Header.Add("Referer", "http://www.crunchyroll.com/"+strings.Split(e.Path, "/")[1])
	resp, err := client.Get(e.URL, nil)
	if err != nil {
		return common.NewError("There was an error requesting the episode doc", err)
	}

	// Creates the document that will be used to scrape for episode metadata
	doc, err := goquery.NewDocumentFromResponse(resp)
	if err != nil {
		return common.NewError("There was an error reading the episode doc", err)
	}

	// Request querystring
	queryString := url.Values{
		"req":                     {"RpcApiVideoPlayer_GetStandardConfig"},
		"media_id":                {strconv.Itoa(e.ID)},
		"video_format":            {getMapping(e.Quality, formats)},
		"video_quality":           {getMapping(e.Quality, qualities)},
		"auto_play":               {"1"},
		"aff":                     {"crunchyroll-website"},
		"show_pop_out_controls":   {"1"},
		"pop_out_disable_message": {""},
		"click_through":           {"0"},
	}.Encode()

	// Request body
	reqBody := bytes.NewBufferString(url.Values{"current_page": {e.URL}}.Encode())

	// Request header
	header := http.Header{}
	header.Add("Host", "www.crunchyroll.com")
	header.Add("Origin", "http://static.ak.crunchyroll.com")
	header.Add("Content-Type", "application/x-www-form-urlencoded")
	header.Set("Referer", "http://static.ak.crunchyroll.com/versioned_assets/StandardVideoPlayer.f3770232.swf")
	header.Add("X-Requested-With", "ShockwaveFlash/22.0.0.192")
	resp, err = client.Post("http://www.crunchyroll.com/xml/?"+queryString, header, reqBody)
	if err != nil {
		return common.NewError("There was an error retrieving the manifest", err)
	}

	// Gets the xml string from the received xml response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return common.NewError("There was an error reading the xml response", err)
	}

	// Checks for an unsupported region first
	// TODO Use REGEX to extract xml
	xmlString := string(body)
	if strings.Contains(xmlString, "<code>") && strings.Contains(xmlString, "</code>") {
		if strings.SplitN(strings.SplitN(xmlString, "<code>", 2)[1], "</code>", 2)[0] == "4" {
			return common.NewError("This video is not available in your region", err)
		}
	}

	// Same type of xml parsing to get the file
	// TODO Use REGEX to extract efile
	eFile := ""
	if strings.Contains(xmlString, "<file>") && strings.Contains(xmlString, "</file>") {
		eFile = strings.SplitN(strings.SplitN(xmlString, "<file>", 2)[1], "</file>", 2)[0]
	} else {
		return common.NewError("No hosts were found for the episode", err)
	}

	e.Title = strings.Replace(strings.Replace(doc.Find("#showmedia_about_name").First().Text(), "“", "", -1), "”", "", -1)
	e.Filename = common.CleanFilename(e.Filename)
	e.StreamURL = strings.Replace(eFile, "amp;", "", -1)
	return nil
}

// Download downloads entire episode to our temp directory
func (e *Episode) Download(vp *common.VideoProcessor, testOnly bool) error {
	streamURL := e.StreamURL
	if testOnly {
		streamURL = "https://dl.v.vrv.co/evs/f567bc31ec99545c1b550ee212e0cee2/assets/yfnvwi9qd7b0jtu_,1757557.mp4,1757559.mp4,1757555.mp4,1757553.mp4,1757551.mp4,.urlset/master.m3u8?Policy=eyJTdGF0ZW1lbnQiOlt7IlJlc291cmNlIjoiaHR0cCo6Ly9kbC52LnZydi5jby9ldnMvZjU2N2JjMzFlYzk5NTQ1YzFiNTUwZWUyMTJlMGNlZTIvYXNzZXRzL3lmbnZ3aTlxZDdiMGp0dV8sMTc1NzU1Ny5tcDQsMTc1NzU1OS5tcDQsMTc1NzU1NS5tcDQsMTc1NzU1My5tcDQsMTc1NzU1MS5tcDQsLnVybHNldC9tYXN0ZXIubTN1OCIsIkNvbmRpdGlvbiI6eyJEYXRlTGVzc1RoYW4iOnsiQVdTOkVwb2NoVGltZSI6MTU1OTA4OTgyOX19fV19&Signature=cf2gng5c-eG84gIydJ1VAzNcVjEi~dOua8Fa-0X3aCxCp1x7QC4QUC5aRmqI9Ea1cBpY2Hn4hLv17rxhoTrg2ZnrD6bWY~lUibj~08XdhRLXugKEvy9N6AraF35lvFb-X4kVis5EDGnVxp-n3StmotRgRA44ReHWYT-wFovJcUX3QvOszH9rWNByalP9Edmb3NavKY~KYWiUIaRf~qJmMNBFuS3Lc6Y0uP6zvQEgGWK2fPImDEkzuX0~9mxaTcnWtvxpLBwwQEoBoAd5Wamw-3K~nI5kGu-rkiyjQlmrAnR3MaQNcoYhycz78L7wgJg6m1FnxwvFEHYVX52u9-EvnQ__&Key-Pair-Id=DLVR"
	}
	return vp.DumpHLS(streamURL)
}

// GetFilename returns the Episodes filename
func (e *Episode) GetFilename() string {
	return e.Filename
}

// getMapping out what the format or resolution of the video should be based on
// crunchyroll xml
func getMapping(quality string, m map[string]string) string {
	a := strings.ToLower(quality)
	for k, v := range m {
		if strings.Contains(a, k) {
			return v
		}
	}
	return "0"
}
