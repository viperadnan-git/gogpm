package gpm

import (
	"fmt"
	"sync"

	"github.com/viperadnan-git/go-gpm/internal/core"
)

// ApiConfig holds configuration for the Google Photos API client
type ApiConfig = core.ApiConfig

// GooglePhotosAPI is the main API client for Google Photos operations
type GooglePhotosAPI struct {
	*core.Api
	uploadMu sync.Mutex // Serializes upload batches
}

// NewGooglePhotosAPI creates a new Google Photos API client
func NewGooglePhotosAPI(cfg ApiConfig) (*GooglePhotosAPI, error) {
	coreApi, err := core.NewApi(cfg)
	if err != nil {
		return nil, err
	}
	return &GooglePhotosAPI{Api: coreApi}, nil
}

// DownloadThumbnail downloads a thumbnail to the specified output path
// Returns the final output path
func (g *GooglePhotosAPI) DownloadThumbnail(mediaKey string, width, height int, forceJpeg, noOverlay bool, outputPath string) (string, error) {
	body, err := g.GetThumbnail(mediaKey, width, height, forceJpeg, noOverlay)
	if err != nil {
		return "", err
	}
	defer body.Close()

	filename := mediaKey + ".jpg"
	return DownloadFromReader(body, outputPath, filename)
}

// DownloadMedia downloads a media item to the specified output path
// Returns the final output path
func (g *GooglePhotosAPI) DownloadMedia(mediaKey string, outputPath string) (string, error) {
	downloadURL, _, err := g.GetDownloadUrl(mediaKey)
	if err != nil {
		return "", err
	}
	if downloadURL == "" {
		return "", fmt.Errorf("no download URL available")
	}
	return DownloadFile(downloadURL, outputPath)
}
