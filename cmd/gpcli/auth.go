package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/urfave/cli/v3"
)

func authInfoAction(ctx context.Context, cmd *cli.Command) error {
	// Check if --auth flag is set
	if authOverride != "" {
		params, err := ParseAuthString(authOverride)
		if err != nil {
			return fmt.Errorf("invalid auth string: %w", err)
		}
		fmt.Println("Current authentication (from --auth flag):")
		fmt.Printf("  Email: %s\n", params.Get("Email"))
		return nil
	}

	// Load from config
	if err := loadConfig(); err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	config := cfgManager.GetConfig()

	// Show current authentication
	if config.Selected != "" {
		fmt.Printf("Current authentication: %s\n", config.Selected)
	} else {
		fmt.Println("No active authentication")
	}

	// List all available accounts
	if len(config.Credentials) == 0 {
		fmt.Println("\nNo accounts configured. Use 'gpcli auth add <auth-string>' to add one.")
		return nil
	}

	fmt.Println("\nAvailable accounts:")
	for i, cred := range config.Credentials {
		params, err := ParseAuthString(cred)
		if err != nil {
			fmt.Printf("  %d. [Invalid]\n", i+1)
			continue
		}
		email := params.Get("Email")
		marker := ""
		if email == config.Selected {
			marker = " *"
		}
		fmt.Printf("  %d. %s%s\n", i+1, email, marker)
	}

	fmt.Println("\nUse 'gpcli auth set <number|email>' to change active authentication")

	return nil
}

func credentialsAddAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	authString := strings.TrimSpace(cmd.StringArg("auth-string"))

	email, err := cfgManager.AddCredentials(authString)
	if err != nil {
		return fmt.Errorf("invalid credentials: %w", err)
	}

	slog.Info("authentication added successfully", "email", email)
	return nil
}

func credentialsRemoveAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	arg := cmd.StringArg("identifier")
	config := cfgManager.GetConfig()

	email, err := resolveEmailFromArg(arg, config.Credentials)
	if err != nil {
		return err
	}

	if err := cfgManager.RemoveCredentials(email); err != nil {
		return fmt.Errorf("error removing authentication: %w", err)
	}

	slog.Info("authentication removed", "email", email)
	return nil
}

func credentialsSetAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	arg := cmd.StringArg("identifier")
	config := cfgManager.GetConfig()

	email, err := resolveEmailFromArg(arg, config.Credentials)
	if err != nil {
		return err
	}

	cfgManager.SetSelected(email)
	slog.Info("active account set", "email", email)

	return nil
}

func authFileAction(ctx context.Context, cmd *cli.Command) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	fmt.Println(cfgManager.GetConfigPath())
	return nil
}
