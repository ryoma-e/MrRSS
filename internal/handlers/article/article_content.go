package article

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"MrRSS/internal/feed"
	"MrRSS/internal/handlers/core"
	"MrRSS/internal/models"
	"MrRSS/internal/utils"
)

// HandleGetArticleContent fetches the article content from RSS feed dynamically.
func HandleGetArticleContent(h *core.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	articleIDStr := r.URL.Query().Get("id")
	articleID, err := strconv.ParseInt(articleIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid article ID", http.StatusBadRequest)
		return
	}

	// Get the article directly by ID (more efficient and includes hidden articles)
	article, err := h.DB.GetArticleByID(articleID)
	if err != nil {
		http.Error(w, "Article not found", http.StatusNotFound)
		return
	}

	// Get the feed to fetch fresh content
	feeds, err := h.DB.GetFeeds()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var targetFeed *models.Feed
	for i := range feeds {
		if feeds[i].ID == article.FeedID {
			targetFeed = &feeds[i]
			break
		}
	}

	if targetFeed == nil {
		http.Error(w, "Feed not found", http.StatusNotFound)
		return
	}

	// Parse the feed to get fresh content
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Use the fetcher to parse the feed (handles both regular URLs and custom scripts)
	parsedFeed, err := h.Fetcher.ParseFeedWithScript(ctx, targetFeed.URL, targetFeed.ScriptPath)
	if err != nil {
		log.Printf("Error parsing feed for article content: %v", err)
		http.Error(w, "Failed to fetch article content", http.StatusInternalServerError)
		return
	}

	// Find the article in the feed by URL (use normalized comparison for robustness)
	var content string
	for _, item := range parsedFeed.Items {
		if utils.URLsMatch(item.Link, article.URL) {
			// Use the centralized content extraction logic to ensure consistency
			content = feed.ExtractContent(item)
			// Clean HTML to fix malformed tags that can cause rendering issues
			content = utils.CleanHTML(content)
			break
		}
	}

	json.NewEncoder(w).Encode(map[string]string{
		"content": content,
	})
}
