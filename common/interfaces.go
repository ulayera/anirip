package common /* import "s32x.com/anirip/common" */

// Show implements Show functionality
type Show interface {
	Scrape(client *HTTPClient, showURL string) error
	GetTitle() string
	GetSeasons() Seasons
}

// Seasons is an aliased slice of Seasons
type Seasons []Season

// Season implements Season functionality
type Season interface {
	GetNumber() int
	GetEpisodes() Episodes
}

// Episodes is an aliased slice of Episodes
type Episodes []Episode

// Episode implements Episode functionality
type Episode interface {
	GetEpisodeInfo(client *HTTPClient, quality string) error
	Download(vp *VideoProcessor, testOnly bool) error
	DownloadSubtitles(client *HTTPClient, language, tempDir string) (string, error)
	GetFilename() string
}
