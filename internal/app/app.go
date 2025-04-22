package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/semaphore"

	configService "github.com/zepollabot/media-rating-overlay/internal/config"
	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	env "github.com/zepollabot/media-rating-overlay/internal/environment"
	mediaModel "github.com/zepollabot/media-rating-overlay/internal/media-service/model"
	ratingModel "github.com/zepollabot/media-rating-overlay/internal/rating-service/model"
)

const defaultLogFilePath = "logs/media-rating-overlay.log"

// App represents the main application with all its dependencies
type App struct {
	config                 *config.Config
	Logger                 *zap.Logger
	ctx                    context.Context
	cancel                 context.CancelFunc
	workSemaphore          *semaphore.Weighted
	libraryProcessor       *LibraryProcessor
	mediaServices          []mediaModel.MediaService
	ratingPlatformServices []ratingModel.RatingService
	shutdownChan           chan struct{}
	doneChan               chan struct{}
}

// NewApp creates a new instance of the application
func NewApp() (*App, error) {
	// Get current environment
	env := env.GetEnvironment()

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())

	// Load configuration
	appConfig, err := configService.LoadConfig(env)
	if err != nil {
		cancel()
		return nil, err
	}

	// TODO: Validate configuration

	// Logger initialization
	var logger *zap.Logger
	if appConfig.Logger.LogFilePath != "" {
		// Use logger configuration from config file
		logger, err = NewLoggerFromConfig(&appConfig.Logger)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("error initializing logger from config: %w", err)
		}
	} else {
		// Use default logger configuration
		logger, err = NewLoggerWithDefaults(defaultLogFilePath)
		if err != nil {
			cancel()
			return nil, fmt.Errorf("error initializing logger with defaults: %w", err)
		}
	}
	logger.Info("Application initialization", zap.String("environment", env))

	// System resource configuration
	maxThreads := configService.ConfigureSystemResources(appConfig.Performance, logger)
	workSemaphore := semaphore.NewWeighted(int64(maxThreads))

	// Initialize services
	serviceInitializer := NewServiceInitializer(logger, appConfig, ctx, workSemaphore)
	if err := serviceInitializer.InitializeServices(); err != nil {
		cancel()
		return nil, fmt.Errorf("error initializing services: %w", err)
	}

	// Create the App instance
	app := &App{
		config:                 appConfig,
		Logger:                 logger,
		ctx:                    ctx,
		cancel:                 cancel,
		workSemaphore:          workSemaphore,
		mediaServices:          serviceInitializer.GetMediaServices(),
		ratingPlatformServices: serviceInitializer.GetRatingPlatformServices(),
		libraryProcessor:       serviceInitializer.GetLibraryProcessor(),
		shutdownChan:           make(chan struct{}),
		doneChan:               make(chan struct{}),
	}

	return app, nil
}

// Run executes the main application flow
func (a *App) Run() error {
	start := time.Now()
	a.Logger.Info("Starting application")

	// Setup signal handling for graceful shutdown
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	// Goroutine to handle termination signals
	go func() {
		sig := <-signalChan
		a.Logger.Info("Termination signal received", zap.String("signal", sig.String()))
		a.cancel() // Just cancel the context, don't call Shutdown
	}()

	// Create a context with timeout for the entire processing
	ctx, cancel := context.WithTimeout(a.ctx, a.config.Performance.LibraryProcessingTimeout)
	defer cancel()

	// Process each media service
	for _, mediaService := range a.mediaServices {
		// Check for context cancellation before processing each service
		if err := ctx.Err(); err != nil {
			close(a.doneChan) // Signal completion
			return fmt.Errorf("context cancelled before processing media service %s: %w", mediaService.Name, err)
		}

		a.Logger.Info("Processing media service", zap.String("media_service", mediaService.Name))

		// Retrieve libraries from the media service
		serviceLibraries, err := mediaService.LibraryService.GetLibraries(ctx)
		if err != nil {
			a.Logger.Error("error retrieving libraries",
				zap.String("media_service", mediaService.Name),
				zap.Error(err),
			)
			continue
		}

		// Create service context for the current media service
		serviceCtx := ServiceContext{
			LibrariesService: mediaService.LibraryService,
			ItemsService:     mediaService.ItemService,
			PostersService:   mediaService.PosterService,
		}

		// Process each library
		for _, configLibrary := range mediaService.Libraries {
			// Check for context cancellation before processing each library
			if err := ctx.Err(); err != nil {
				close(a.doneChan) // Signal completion
				return fmt.Errorf("context cancelled before processing library %s: %w", configLibrary.Name, err)
			}

			if err := a.libraryProcessor.ProcessLibrary(ctx, &configLibrary, &serviceLibraries, serviceCtx); err != nil {
				if errors.Is(err, context.DeadlineExceeded) {
					a.Logger.Error("Error processing library: processing timed out",
						zap.String("library", configLibrary.Name),
						zap.Error(err),
						zap.String("suggestion", fmt.Sprintf("Try increasing 'processor.library_processor.default_timeout' in your config. Current timeout for this library was: %s. Overall application timeout is: %s", a.libraryProcessor.defaultTimeout, a.config.Performance.LibraryProcessingTimeout)),
					)
				} else {
					a.Logger.Error("Error processing library",
						zap.String("library", configLibrary.Name),
						zap.Error(err),
					)
				}
				continue
			}
		}
	}

	a.Logger.Info("Operation completed")
	a.Logger.Info("Execution time", zap.Duration("duration", time.Since(start)))

	close(a.doneChan) // Signal completion
	return nil
}

// Shutdown performs cleanup and graceful shutdown of the application
func (a *App) Shutdown() {
	a.Logger.Info("Shutting down application")

	// Cancel the main context
	a.cancel()

	// Wait for completion or timeout
	select {
	case <-a.doneChan:
		a.Logger.Info("All processing completed")
	case <-time.After(5 * time.Second):
		a.Logger.Warn("Timeout waiting for processing to complete")
	}

	a.Logger.Info("Application terminated")
}
