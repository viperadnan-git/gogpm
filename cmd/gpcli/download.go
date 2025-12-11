package main

import (
	"context"
	"fmt"

	gogpm "github.com/viperadnan-git/gogpm"

	"github.com/urfave/cli/v3"
)

func downloadAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	cfg := cfgManager.GetConfig()

	input := cmd.StringArg("input")
	urlOnly := cmd.Bool("url")
	outputPath := cmd.String("output")

	authData := getAuthData(cfg)
	if authData == "" {
		return fmt.Errorf("no authentication configured. Use 'gpcli auth add' to add credentials")
	}

	apiClient, err := gogpm.NewGooglePhotosAPI(gogpm.ApiConfig{
		AuthData: authData,
		Proxy:    cfg.Proxy,
	})
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	mediaKey, err := apiClient.ResolveMediaKey(ctx, input)
	if err != nil {
		return err
	}

	if !urlOnly {
		logger.Info("fetching download URL", "media_key", mediaKey)
	}

	downloadURL, isEdited, err := apiClient.GetDownloadUrl(mediaKey)
	if err != nil {
		return fmt.Errorf("failed to get download URL: %w", err)
	}

	if downloadURL == "" {
		return fmt.Errorf("no download URL available")
	}

	// If --url flag is set, just print the URL and exit
	if urlOnly {
		fmt.Println(downloadURL)
		return nil
	}

	// Download the file
	logger.Info("downloading", "is_edited", isEdited)
	savedPath, err := gogpm.DownloadFile(downloadURL, outputPath)
	if err != nil {
		return err
	}
	logger.Info("download complete", "path", savedPath)
	return nil
}

func thumbnailAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	cfg := cfgManager.GetConfig()

	input := cmd.StringArg("input")
	outputPath := cmd.String("output")
	width := int(cmd.Int("width"))
	height := int(cmd.Int("height"))
	forceJpeg := cmd.Bool("jpeg")
	noOverlay := !cmd.Bool("overlay")

	authData := getAuthData(cfg)
	if authData == "" {
		return fmt.Errorf("no authentication configured. Use 'gpcli auth add' to add credentials")
	}

	apiClient, err := gogpm.NewGooglePhotosAPI(gogpm.ApiConfig{
		AuthData: authData,
		Proxy:    cfg.Proxy,
	})
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	mediaKey, err := apiClient.ResolveMediaKey(ctx, input)
	if err != nil {
		return err
	}

	logger.Info("downloading thumbnail", "media_key", mediaKey)

	savedPath, err := apiClient.DownloadThumbnail(mediaKey, width, height, forceJpeg, noOverlay, outputPath)
	if err != nil {
		return err
	}
	logger.Info("thumbnail downloaded", "path", savedPath)
	return nil
}
