package core

import (
	"MrRSS/internal/models"
	"context"
	"log"
	"strconv"
	"time"
)

// StartBackgroundScheduler starts the background scheduler for auto-updates and cleanup.
func (h *Handler) StartBackgroundScheduler(ctx context.Context) {
	// Run initial cleanup only if auto_cleanup is enabled
	go func() {
		autoCleanup, _ := h.DB.GetSetting("auto_cleanup_enabled")
		if autoCleanup == "true" {
			log.Println("Running initial article cleanup...")
			count, err := h.DB.CleanupOldArticles()
			if err != nil {
				log.Printf("Error during initial cleanup: %v", err)
			} else {
				log.Printf("Initial cleanup: removed %d old articles", count)
			}
		}
	}()

	// Check refresh mode
	refreshMode, _ := h.DB.GetSetting("refresh_mode")
	
	if refreshMode == "intelligent" {
		// Use intelligent refresh mode with per-feed intervals
		h.startIntelligentScheduler(ctx)
	} else {
		// Use fixed interval mode (default)
		h.startFixedScheduler(ctx)
	}
}

// startFixedScheduler uses a fixed interval for all feeds
func (h *Handler) startFixedScheduler(ctx context.Context) {
	for {
		intervalStr, _ := h.DB.GetSetting("update_interval")
		interval := 30
		if i, err := strconv.Atoi(intervalStr); err == nil && i > 0 {
			interval = i
		}

		log.Printf("Next auto-update in %d minutes (fixed mode)", interval)

		select {
		case <-ctx.Done():
			log.Println("Stopping background scheduler")
			return
		case <-time.After(time.Duration(interval) * time.Minute):
			h.Fetcher.FetchAll(ctx)
			h.runCleanup()
		}
	}
}

// startIntelligentScheduler uses per-feed intervals with staggered refresh
func (h *Handler) startIntelligentScheduler(ctx context.Context) {
	log.Println("Starting intelligent refresh scheduler")
	
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping intelligent scheduler")
			return
		case <-ticker.C:
			// Check each feed to see if it needs refresh
			go h.refreshFeedsIntelligently(ctx)
		}
	}
}

// refreshFeedsIntelligently checks and refreshes feeds based on their individual intervals
func (h *Handler) refreshFeedsIntelligently(ctx context.Context) {
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		log.Printf("Error getting feeds for intelligent refresh: %v", err)
		return
	}

	// Use intelligent refresh calculator
	calculator := h.Fetcher.GetIntelligentRefreshCalculator()
	
	for i, feed := range feeds {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return
		default:
		}

		// Determine refresh interval for this feed
		var refreshInterval time.Duration
		if feed.RefreshInterval > 0 {
			// Use per-feed custom interval
			refreshInterval = time.Duration(feed.RefreshInterval) * time.Minute
		} else {
			// Calculate intelligent interval
			refreshInterval = calculator.CalculateInterval(feed)
		}

		// Check if feed needs refresh based on last_updated time
		timeSinceUpdate := time.Since(feed.LastUpdated)
		if timeSinceUpdate >= refreshInterval {
			// Apply staggered delay to avoid thundering herd
			staggerDelay := h.Fetcher.GetStaggeredDelay(feed.ID, len(feeds))
			
			// Schedule feed refresh with stagger
			go func(f models.Feed, delay time.Duration, interval time.Duration) {
				time.Sleep(delay)
				select {
				case <-ctx.Done():
					return
				default:
					log.Printf("Intelligently refreshing feed %s (interval: %v)", f.Title, interval)
					h.Fetcher.FetchSingleFeed(ctx, f)
				}
			}(feed, staggerDelay, refreshInterval)
		}
		
		// Small delay between checking feeds to avoid CPU spikes
		if i < len(feeds)-1 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	// Run cleanup after refresh cycle
	h.runCleanup()
}

// runCleanup runs the cleanup routine if enabled
func (h *Handler) runCleanup() {
	autoCleanup, _ := h.DB.GetSetting("auto_cleanup_enabled")
	if autoCleanup == "true" {
		count, err := h.DB.CleanupOldArticles()
		if err != nil {
			log.Printf("Error during automatic cleanup: %v", err)
		} else if count > 0 {
			log.Printf("Automatic cleanup: removed %d old articles", count)
		}
	}
}
