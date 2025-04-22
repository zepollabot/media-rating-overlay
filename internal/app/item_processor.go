package core

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	config "github.com/zepollabot/media-rating-overlay/internal/config/model"
	media "github.com/zepollabot/media-rating-overlay/internal/media-service"
	model "github.com/zepollabot/media-rating-overlay/internal/model"
)

type PosterGenerator interface {
	ApplyLogos(ctx context.Context, filePath string, config *config.Library, item model.Item) (string, error)
}

// ItemEligibilityChecker determines if an item should be processed
type ItemEligibilityChecker interface {
	IsEligible(item *model.Item) bool
}

// RatingBuilder fetches and builds ratings for an item
type RatingBuilder interface {
	BuildRatings(ctx context.Context, item *model.Item) error
}

type SemaphoreWeighted interface {
	Acquire(ctx context.Context, n int64) error
	Release(n int64)
	TryAcquire(n int64) bool
}

// ItemProcessor handles the processing of library items and posters
type ItemProcessor struct {
	logger             *zap.Logger
	ctx                context.Context
	workSemaphore      SemaphoreWeighted
	posterService      media.PosterService
	posterGenerator    PosterGenerator
	eligibilityChecker ItemEligibilityChecker
	ratingBuilder      RatingBuilder
}

// NewItemProcessor creates a new item processor
func NewItemProcessor(
	logger *zap.Logger,
	ctx context.Context,
	workSemaphore SemaphoreWeighted,
	posterGenerator PosterGenerator,
	eligibilityChecker ItemEligibilityChecker,
	ratingBuilder RatingBuilder,
) *ItemProcessor {
	return &ItemProcessor{
		logger:             logger,
		ctx:                ctx,
		workSemaphore:      workSemaphore,
		posterGenerator:    posterGenerator,
		eligibilityChecker: eligibilityChecker,
		ratingBuilder:      ratingBuilder,
	}
}

func (ip *ItemProcessor) SetPosterService(posterService media.PosterService) {
	ip.posterService = posterService
}

// ProcessItems processes library items in parallel
func (ip *ItemProcessor) ProcessItems(items []model.Item, configLib *config.Library) error {
	var notProcessed int
	var postersWithErrors int

	if err := ip.validateDependencies(); err != nil {
		return err
	}

	resultChan := make(chan model.PosterResult, len(items))
	processingDone := make(chan struct{})

	// Create a new errgroup with the context
	eg, ctx := errgroup.WithContext(ip.ctx)

	// Start result collection in a separate goroutine
	go func() {
		defer close(processingDone)
		processed := 0
		for result := range resultChan {
			processed++
			status := "completed"
			if result.Err != nil {
				status = "failed"
				postersWithErrors++
			}

			ip.logger.Info("Item processed",
				zap.String("Progress", fmt.Sprintf("%d/%d", processed, len(items))),
				zap.String("Title", result.Title),
				zap.String("Status", status),
				zap.String("Original Poster Disk Position", result.OriginalPosterDiskPosition),
				zap.String("Overlay Poster Disk Position", result.OverlayPosterDiskPosition),
				zap.Error(result.Err),
			)
		}
	}()

	// Process items in parallel
	for i, item := range items {
		if !ip.eligibilityChecker.IsEligible(&item) {
			notProcessed++
			continue
		}

		currentItem := item
		currentIndex := i

		eg.Go(func() error {
			if err := ip.workSemaphore.Acquire(ctx, 1); err != nil {
				return fmt.Errorf("error acquiring semaphore: %w", err)
			}
			defer ip.workSemaphore.Release(1)

			result := ip.ProcessItem(currentItem, currentIndex, ip.posterService, configLib)

			select {
			case resultChan <- result:
			case <-ctx.Done():
				return ctx.Err()
			}

			return nil
		})
	}

	// Wait for all processing goroutines to complete
	if err := eg.Wait(); err != nil {
		close(resultChan)
		return fmt.Errorf("error during parallel processing: %w", err)
	}

	// Close the result channel after all processing is done
	close(resultChan)

	// Wait for result collection to complete
	<-processingDone

	ip.logger.Info("Processing Report",
		zap.String("Library Name", configLib.Name),
		zap.Int("Total Items Found", len(items)),
		zap.Int("Processed Items", len(items)-notProcessed),
		zap.Int("Ineligible Items", notProcessed),
		zap.Int("Posters With Errors", postersWithErrors),
	)

	return nil
}

func (ip *ItemProcessor) validateDependencies() error {
	if ip.posterService == nil {
		ip.logger.Error("Poster service not set")
		return fmt.Errorf("poster service not set")
	}

	return nil
}

// ProcessItem processes the poster of a single item
func (ip *ItemProcessor) ProcessItem(item model.Item, index int, posterService media.PosterService, configLib *config.Library) model.PosterResult {
	ip.logger.Debug("Starting item processing",
		zap.Int("Index", index),
		zap.String("Item ID", item.ID),
		zap.String("Item Title", item.Title),
	)

	// Check for context cancellation before starting
	if err := ip.ctx.Err(); err != nil {
		return model.PosterResult{
			Title: item.Title,
			Err:   err,
		}
	}

	if err := ip.ratingBuilder.BuildRatings(ip.ctx, &item); err != nil {
		return model.PosterResult{
			Title: item.Title,
			Err:   err,
		}
	}

	ip.logger.Debug("Ratings",
		zap.String("Item ID", item.ID),
		zap.Any("Ratings", item.Ratings),
	)

	// First ensure the poster exists
	if err := posterService.EnsurePosterExists(ip.ctx, item, configLib); err != nil {
		return model.PosterResult{
			Title: item.Title,
			Err:   err,
		}
	}

	// Then get its position
	posterDiskPosition, err := posterService.GetPosterDiskPosition(ip.ctx, item, configLib)
	if err != nil {
		return model.PosterResult{
			Title: item.Title,
			Err:   err,
		}
	}

	// Check for context cancellation after GetPosterDiskPosition
	if ctxErr := ip.ctx.Err(); ctxErr != nil {
		return model.PosterResult{
			Title:                      item.Title,
			OriginalPosterDiskPosition: posterDiskPosition,
			Err:                        ctxErr,
		}
	}

	result := model.PosterResult{
		Title:                      item.Title,
		OriginalPosterDiskPosition: posterDiskPosition,
	}

	newPosterDiskPosition, err := ip.posterGenerator.ApplyLogos(ip.ctx, posterDiskPosition, configLib, item)
	if err != nil {
		return model.PosterResult{
			Title: item.Title,
			Err:   err,
		}
	}

	// Check for context cancellation after ApplyLogos
	if ctxErr := ip.ctx.Err(); ctxErr != nil {
		return model.PosterResult{
			Title:                      item.Title,
			OriginalPosterDiskPosition: posterDiskPosition,
			Err:                        ctxErr,
		}
	}

	result.OverlayPosterDiskPosition = newPosterDiskPosition

	ip.logger.Debug("Poster processing completed",
		zap.Int("Index", index),
		zap.String("Item ID", item.ID),
		zap.String("Item Title", item.Title),
	)

	return result
}
