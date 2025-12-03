package main

import (
	"fmt"
	"gpcli/src"
	"log/slog"
	"os"
	"strings"
	"sync"

	"github.com/urfave/cli/v2"
)

var logger *slog.Logger
var currentLogLevel slog.Level

// parseLogLevel converts a string log level to slog.Level
func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// initLogger initializes the global logger with the specified level
func initLogger(level slog.Level) {
	opts := &slog.HandlerOptions{
		Level: level,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger = slog.New(handler)
	slog.SetDefault(logger)
}

// initQuietLogger initializes a logger that only shows errors
func initQuietLogger() {
	opts := &slog.HandlerOptions{
		Level: slog.LevelError,
	}
	handler := slog.NewTextHandler(os.Stdout, opts)
	logger = slog.New(handler)
	slog.SetDefault(logger)
	currentLogLevel = slog.LevelError
}

func runCLI() {
	app := &cli.App{
		Name:                   "gpcli",
		Usage:                  "Google Photos unofficial CLI client",
		Version:                src.Version,
		UseShortOptionHandling: true,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Path to config file (default: ./gpcli.config)",
			},
			&cli.StringFlag{
				Name:    "log-level",
				Aliases: []string{"l"},
				Value:   "info",
				Usage:   "Set log level: debug, info, warn, error",
			},
			&cli.BoolFlag{
				Name:    "quiet",
				Aliases: []string{"q"},
				Usage:   "Suppress all log output (overrides --log-level)",
			},
			&cli.StringFlag{
				Name:  "auth",
				Usage: "Authentication string (overrides config file)",
			},
		},
		Before: func(c *cli.Context) error {
			// Initialize logger - quiet mode overrides log level
			if c.Bool("quiet") {
				initQuietLogger()
			} else {
				currentLogLevel = parseLogLevel(c.String("log-level"))
				initLogger(currentLogLevel)
			}

			// Set config path from global flag before any command runs
			if configPath := c.String("config"); configPath != "" {
				src.ConfigPath = configPath
			}

			// Set auth override from flag
			if auth := c.String("auth"); auth != "" {
				src.AuthOverride = auth
			}
			return nil
		},
		Commands: []*cli.Command{
			{
				Name:      "upload",
				Usage:     "Upload a file or directory to Google Photos",
				ArgsUsage: "[flags] <filepath>",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "recursive",
						Aliases: []string{"r"},
						Usage:   "Include subdirectories",
					},
					&cli.IntFlag{
						Name:    "threads",
						Aliases: []string{"t"},
						Value:   3,
						Usage:   "Number of upload threads",
					},
					&cli.BoolFlag{
						Name:    "force",
						Aliases: []string{"f"},
						Usage:   "Force upload even if file exists",
					},
					&cli.BoolFlag{
						Name:    "delete",
						Aliases: []string{"d"},
						Usage:   "Delete from host after upload",
					},
					&cli.BoolFlag{
						Name:    "disable-filter",
						Aliases: []string{"df"},
						Usage:   "Disable file type filtering",
					},
				},
				Action: uploadAction,
			},
			{
				Name:   "auth",
				Usage:  "Manage Google Photos authentication",
				Action: authInfoAction,
				Subcommands: []*cli.Command{
					{
						Name:      "add",
						Usage:     "Add a new authentication",
						ArgsUsage: "<auth-string>",
						Action:    credentialsAddAction,
					},
					{
						Name:      "remove",
						Aliases:   []string{"rm"},
						Usage:     "Remove an authentication by email",
						ArgsUsage: "<email>",
						Action:    credentialsRemoveAction,
					},
					{
						Name:    "list",
						Aliases: []string{"ls"},
						Usage:   "List all authentications",
						Action:  credentialsListAction,
					},
					{
						Name:      "set",
						Aliases:   []string{"select"},
						Usage:     "Set active authentication (supports partial matching)",
						ArgsUsage: "<email>",
						Action:    credentialsSetAction,
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		slog.Error("command failed", "error", err)
		os.Exit(1)
	}
}

func uploadAction(c *cli.Context) error {
	if c.NArg() < 1 {
		return fmt.Errorf("filepath required")
	}

	filePath := c.Args().First()

	// Validate that filepath exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("file or directory does not exist: %s", filePath)
	}

	// Load backend config
	err := src.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Override config with CLI flags
	threads := c.Int("threads")
	src.AppConfig.Recursive = c.Bool("recursive")
	src.AppConfig.UploadThreads = threads
	src.AppConfig.ForceUpload = c.Bool("force")
	src.AppConfig.DeleteFromHost = c.Bool("delete")
	src.AppConfig.DisableUnsupportedFilesFilter = c.Bool("disable-filter")

	// Log configuration at start
	logger.Info("starting upload",
		"path", filePath,
		"threads", threads,
		"recursive", src.AppConfig.Recursive,
		"force", src.AppConfig.ForceUpload,
		"delete", src.AppConfig.DeleteFromHost,
		"disable-filter", src.AppConfig.DisableUnsupportedFilesFilter,
	)

	// Track results
	var mu sync.Mutex
	var totalFiles int
	var uploaded int
	var existing int
	var failed int
	done := make(chan struct{})

	// Create CLI app with event callback
	eventCallback := func(event string, data any) {
		mu.Lock()
		defer mu.Unlock()

		switch event {
		case "uploadStart":
			if start, ok := data.(src.UploadBatchStart); ok {
				totalFiles = start.Total
				logger.Info("upload batch started", "total", totalFiles)
			}
		case "ThreadStatus":
			if status, ok := data.(src.ThreadStatus); ok {
				logger.Debug("worker status",
					"worker_id", status.WorkerID,
					"status", status.Status,
					"file", status.FileName,
				)
			}
		case "FileStatus":
			if result, ok := data.(src.FileUploadResult); ok {
				if result.IsError {
					failed++
					logger.Error("upload failed",
						"path", result.Path,
						"error", result.Error,
					)
				} else if result.IsExisting {
					existing++
					logger.Info("already exists",
						"path", result.Path,
						"media_key", result.MediaKey,
					)
				} else {
					uploaded++
					logger.Info("upload success",
						"path", result.Path,
						"media_key", result.MediaKey,
					)
				}
			}
		case "uploadStop":
			close(done)
		}
	}

	app := src.NewGooglePhotosCLI(eventCallback, currentLogLevel)
	uploadManager := src.NewUploadManager(app)

	// Run upload in background
	go func() {
		uploadManager.Upload(app, []string{filePath})
	}()

	// Wait for upload to complete
	<-done

	// Print summary
	logger.Info("upload complete",
		"total", totalFiles,
		"succeeded", uploaded+existing,
		"failed", failed,
		"uploaded", uploaded,
		"existing", existing,
	)

	return nil
}

func loadConfig() error {
	return src.LoadConfig()
}

func authInfoAction(c *cli.Context) error {
	// Check if --auth flag is set
	if src.AuthOverride != "" {
		params, err := src.ParseAuthString(src.AuthOverride)
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

	configManager := &src.ConfigManager{}
	config := configManager.GetConfig()

	if config.Selected == "" {
		return fmt.Errorf("no active authentication. Use 'gpcli auth set <email>' or --auth flag")
	}

	fmt.Println("Current authentication:")
	fmt.Printf("  Email: %s\n", config.Selected)
	return nil
}

func credentialsAddAction(c *cli.Context) error {
	if c.NArg() < 1 {
		return fmt.Errorf("auth-string required")
	}

	if err := loadConfig(); err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	authString := c.Args().First()
	configManager := &src.ConfigManager{}

	if err := configManager.AddCredentials(authString); err != nil {
		return fmt.Errorf("error adding authentication: %w", err)
	}

	slog.Info("authentication added successfully")
	return nil
}

func credentialsRemoveAction(c *cli.Context) error {
	if c.NArg() < 1 {
		return fmt.Errorf("email required")
	}

	if err := loadConfig(); err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	email := c.Args().First()
	configManager := &src.ConfigManager{}

	if err := configManager.RemoveCredentials(email); err != nil {
		return fmt.Errorf("error removing authentication: %w", err)
	}

	slog.Info("authentication removed", "email", email)
	return nil
}

func credentialsListAction(c *cli.Context) error {
	if err := loadConfig(); err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	configManager := &src.ConfigManager{}
	config := configManager.GetConfig()

	if len(config.Credentials) == 0 {
		slog.Info("no authentication found")
		return nil
	}

	fmt.Println("Accounts:")
	for i, cred := range config.Credentials {
		params, err := src.ParseAuthString(cred)
		if err != nil {
			fmt.Printf("  %d. [Invalid]\n", i+1)
			continue
		}
		email := params.Get("Email")
		marker := " "
		if email == config.Selected {
			marker = "*"
		}
		fmt.Printf("  %s %s\n", marker, email)
	}

	if config.Selected != "" {
		fmt.Printf("\n* = active\n")
	}
	fmt.Printf("\nUse 'gpcli auth set <email>' to change active authentication\n")

	return nil
}

func credentialsSetAction(c *cli.Context) error {
	if c.NArg() < 1 {
		return fmt.Errorf("email required")
	}

	if err := loadConfig(); err != nil {
		return fmt.Errorf("error loading config: %w", err)
	}

	query := c.Args().First()
	configManager := &src.ConfigManager{}
	config := configManager.GetConfig()

	// Try to find exact match first
	var matchedEmail string
	for _, cred := range config.Credentials {
		params, err := src.ParseAuthString(cred)
		if err != nil {
			continue
		}
		email := params.Get("Email")
		if email == query {
			matchedEmail = email
			break
		}
	}

	// If no exact match, try fuzzy matching (substring match)
	if matchedEmail == "" {
		var candidates []string
		for _, cred := range config.Credentials {
			params, err := src.ParseAuthString(cred)
			if err != nil {
				continue
			}
			email := params.Get("Email")
			if containsSubstring(email, query) {
				candidates = append(candidates, email)
			}
		}

		if len(candidates) == 0 {
			return fmt.Errorf("no authentication found matching '%s'", query)
		} else if len(candidates) == 1 {
			matchedEmail = candidates[0]
		} else {
			slog.Error("multiple accounts match query", "query", query, "candidates", candidates)
			return fmt.Errorf("please be more specific")
		}
	}

	configManager.SetSelected(matchedEmail)
	slog.Info("active account set", "email", matchedEmail)

	return nil
}

func containsSubstring(str, substr string) bool {
	strLower := strings.ToLower(str)
	substrLower := strings.ToLower(substr)
	return strings.Contains(strLower, substrLower)
}
