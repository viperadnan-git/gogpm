package main

import (
	"context"
	"fmt"

	gogpm "github.com/viperadnan-git/gogpm"

	"github.com/urfave/cli/v3"
)

func deleteAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	cfg := cfgManager.GetConfig()

	input := cmd.StringArg("input")
	restore := cmd.Bool("restore")

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

	itemKey, err := apiClient.ResolveItemKey(ctx, input)
	if err != nil {
		return err
	}

	if restore {
		logger.Info("restoring from trash", "item_key", itemKey)
		if err := apiClient.RestoreFromTrash([]string{itemKey}); err != nil {
			return fmt.Errorf("failed to restore from trash: %w", err)
		}
	} else {
		logger.Info("moving to trash", "item_key", itemKey)
		if err := apiClient.MoveToTrash([]string{itemKey}); err != nil {
			return fmt.Errorf("failed to move to trash: %w", err)
		}
	}

	return nil
}

func archiveAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	cfg := cfgManager.GetConfig()

	input := cmd.StringArg("input")
	unarchive := cmd.Bool("unarchive")

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

	itemKey, err := apiClient.ResolveItemKey(ctx, input)
	if err != nil {
		return err
	}

	isArchived := !unarchive
	if isArchived {
		logger.Info("archiving", "item_key", itemKey)
	} else {
		logger.Info("unarchiving", "item_key", itemKey)
	}

	if err := apiClient.SetArchived([]string{itemKey}, isArchived); err != nil {
		return fmt.Errorf("failed to set archived status: %w", err)
	}

	return nil
}

func favouriteAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	cfg := cfgManager.GetConfig()

	input := cmd.StringArg("input")
	remove := cmd.Bool("remove")

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

	itemKey, err := apiClient.ResolveItemKey(ctx, input)
	if err != nil {
		return err
	}

	isFavourite := !remove
	if isFavourite {
		logger.Info("adding to favourites", "item_key", itemKey)
	} else {
		logger.Info("removing from favourites", "item_key", itemKey)
	}

	if err := apiClient.SetFavourite(itemKey, isFavourite); err != nil {
		return fmt.Errorf("failed to set favourite status: %w", err)
	}

	return nil
}

func captionAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	cfg := cfgManager.GetConfig()

	input := cmd.StringArg("input")
	caption := cmd.StringArg("caption")

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

	itemKey, err := apiClient.ResolveItemKey(ctx, input)
	if err != nil {
		return err
	}

	logger.Info("setting caption", "item_key", itemKey, "caption", caption)

	if err := apiClient.SetCaption(itemKey, caption); err != nil {
		return fmt.Errorf("failed to set caption: %w", err)
	}

	return nil
}
