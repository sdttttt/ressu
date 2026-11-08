package service

import (
	"context"
	"errors"
	"log/slog"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/sdttttt/ressu/internal/icon"
	"github.com/sdttttt/ressu/internal/rss"
	"github.com/sdttttt/ressu/internal/store"
)

// maxConcurrent bounds how many feeds are fetched at once.
const maxConcurrent = 4

// iconTTL bounds how long a resolved feed icon is reused.
const iconTTL = 7 * 24 * time.Hour

// ErrRefreshing reports a refresh that is already running.
var ErrRefreshing = errors.New("a refresh is already running")

// SyncSummary is what a refresh of every feed did.
type SyncSummary struct {
	Updated int `json:"updated"`
	Failed  int `json:"failed"`
	New     int `json:"new"`
}

// FeedPreview is the subscribable metadata of a fetched feed.
type FeedPreview struct {
	Title       string `json:"title"`
	SiteURL     string `json:"siteUrl"`
	Description string `json:"description"`
	IconURL     string `json:"iconUrl"`
	Language    string `json:"language"`
}

// Feeds is the subscription service the frontend calls.
type Feeds struct {
	store *store.Store

	mu           sync.Mutex
	refreshing   bool
	iconMu       sync.Mutex
	iconRequests map[int64]*iconRequest
}

// iconRequest tracks one in-flight icon resolution so repeated calls share it.
type iconRequest struct {
	done chan struct{}
	icon store.Icon
	err  error
}

// NewFeeds serves the subscriptions of st.
func NewFeeds(st *store.Store) *Feeds {
	return &Feeds{store: st, iconRequests: make(map[int64]*iconRequest)}
}

// List returns every subscription, with its unread count.
func (f *Feeds) List() FeedsSnapshot {
	return feedsSnapshot(f.store)
}

// feedsSnapshot combines the subscriptions and their total unread count.
func feedsSnapshot(st *store.Store) FeedsSnapshot {
	return FeedsSnapshot{Feeds: st.Feeds(), Unread: st.Unread()}
}

// Icon returns the local icon of a feed, fetching it when the cache is absent,
// stale, or based on a changed feed icon address.
func (f *Feeds) Icon(ctx context.Context, id int64) (store.Icon, error) {
	if err := ctx.Err(); err != nil {
		return store.Icon{}, err
	}

	feed, err := f.store.Feed(id)
	if err != nil {
		slog.Warn("icon feed lookup failed",
			"operation", "feeds.icon", "feedId", id, "error", err)
		return store.Icon{}, err
	}
	cached, err := f.store.Icon(id)
	if err == nil && iconFresh(cached, feed.IconURL) {
		return cached, nil
	}

	return f.fetchIcon(ctx, feed)
}

// fetchIcon resolves one feed's icon once, sharing a single network request
// with concurrent calls for the same subscription.
func (f *Feeds) fetchIcon(ctx context.Context, feed store.Feed) (store.Icon, error) {
	f.iconMu.Lock()
	if request, ok := f.iconRequests[feed.ID]; ok {
		f.iconMu.Unlock()
		<-request.done
		return request.icon, request.err
	}

	request := &iconRequest{done: make(chan struct{})}
	f.iconRequests[feed.ID] = request
	f.iconMu.Unlock()

	image, err := icon.Fetch(ctx, icon.Source{
		FeedURL: feed.URL,
		SiteURL: feed.SiteURL,
		IconURL: feed.IconURL,
	})
	if err != nil {
		if ctx.Err() == nil {
			slog.Warn("icon fetch failed",
				"operation", "feeds.icon", "feedId", feed.ID, "error", err)
		}
	} else {
		request.icon = store.Icon{
			Meta: store.IconMeta{
				SourceURL:   image.SourceURL,
				ContentType: image.ContentType,
				FetchedAt:   time.Now(),
			},
			Data: image.Data,
		}
		if err = f.store.PutIcon(feed.ID, request.icon); err != nil {
			slog.Error("icon cache write failed",
				"operation", "feeds.icon", "feedId", feed.ID, "error", err)
		}
	}
	request.err = err

	f.iconMu.Lock()
	delete(f.iconRequests, feed.ID)
	f.iconMu.Unlock()
	close(request.done)

	return request.icon, request.err
}

// iconFresh reports whether a cached icon is complete, fresh, and still tied
// to the feed's explicit icon address.
func iconFresh(icon store.Icon, feedIconURL string) bool {
	return len(icon.Data) > 0 && icon.Meta.FetchedAt.Add(iconTTL).After(time.Now()) &&
		(feedIconURL == "" || icon.Meta.SourceURL == feedIconURL)
}

// Test fetches and parses a feed without subscribing to it.
func (f *Feeds) Test(ctx context.Context, feedURL string) (FeedPreview, error) {
	feedURL = strings.TrimSpace(feedURL)
	if err := rss.ValidateURL(feedURL); err != nil {
		return FeedPreview{}, err
	}
	if f.duplicateFeed(feedURL) {
		return FeedPreview{}, store.ErrDuplicate
	}

	result, err := rss.Fetch(ctx, feedURL, "", "")
	if err != nil {
		slog.Warn("feed test failed",
			"operation", "feeds.test",
			"feedHost", feedHost(feedURL),
			"error", redactedFeedError(feedURL, err),
		)
		return FeedPreview{}, err
	}

	return FeedPreview{
		Title:       titleOf(result.Feed, feedURL),
		SiteURL:     result.Feed.Meta.SiteURL,
		Description: result.Feed.Meta.Description,
		IconURL:     result.Feed.Meta.IconURL,
		Language:    result.Feed.Meta.Language,
	}, nil
}

// duplicateFeed reports whether a feed is already stored by address.
func (f *Feeds) duplicateFeed(feedURL string) bool {
	for _, exist := range f.store.Feeds() {
		if exist.URL == feedURL {
			return true
		}
	}
	return false
}

// Add subscribes to a feed: it fetches the address once, stores the channel
// and its posts. A non-empty title replaces the feed's channel title.
func (f *Feeds) Add(ctx context.Context, title, feedURL string) (store.Feed, error) {
	feedURL = strings.TrimSpace(feedURL)
	if err := rss.ValidateURL(feedURL); err != nil {
		return store.Feed{}, err
	}
	if f.duplicateFeed(feedURL) {
		return store.Feed{}, store.ErrDuplicate
	}

	result, err := rss.Fetch(ctx, feedURL, "", "")
	if err != nil {
		slog.Warn("feed add failed",
			"operation", "feeds.add", "feedHost", feedHost(feedURL),
			"error", redactedFeedError(feedURL, err))
		return store.Feed{}, err
	}

	now := time.Now()
	title = strings.TrimSpace(title)
	feed, err := f.store.AddFeed(store.Feed{
		URL:           feedURL,
		Title:         firstNonEmpty(title, titleOf(result.Feed, feedURL)),
		SiteURL:       result.Feed.Meta.SiteURL,
		Description:   result.Feed.Meta.Description,
		IconURL:       result.Feed.Meta.IconURL,
		Language:      result.Feed.Meta.Language,
		ETag:          result.ETag,
		LastModified:  result.LastModified,
		LastFetchedAt: &now,
	})
	if err != nil {
		slog.Error("feed add save failed",
			"operation", "feeds.add", "error", err)
		return store.Feed{}, err
	}

	added, err := f.store.MergePosts(feed.ID, toPosts(result.Feed.Items))
	if err != nil {
		slog.Error("stored feed posts failed",
			"operation", "feeds.add", "feedId", feed.ID, "error", err)
		return store.Feed{}, err
	}

	broadcastFeeds(f.store)
	_ = PostsChangedEvent.Broadcast(PostsChanged{FeedID: feed.ID, New: added})
	return f.store.Feed(feed.ID)
}

// Remove unsubscribes from a feed and deletes its posts.
func (f *Feeds) Remove(id int64) error {
	if err := f.store.RemoveFeed(id); err != nil {
		slog.Warn("feed remove failed",
			"operation", "feeds.remove", "feedId", id, "error", err)
		return err
	}
	broadcastFeeds(f.store)
	_ = PostsChangedEvent.Broadcast(PostsChanged{FeedID: id})
	return nil
}

// Refresh fetches one feed and stores the posts it does not have yet.
func (f *Feeds) Refresh(ctx context.Context, id int64) (store.Feed, error) {
	if _, err := f.refreshOne(ctx, id); err != nil {
		return store.Feed{}, err
	}
	return f.store.Feed(id)
}

// RefreshAll fetches every feed, four at a time, and reports each one through
// the sync:state event.
func (f *Feeds) RefreshAll(ctx context.Context) (SyncSummary, error) {
	f.mu.Lock()
	if f.refreshing {
		f.mu.Unlock()
		return SyncSummary{}, ErrRefreshing
	}
	f.refreshing = true
	f.mu.Unlock()
	feeds := f.store.Feeds()
	total := len(feeds)
	slog.Info("refresh all started",
		"operation", "feeds.refreshAll",
		"total", total,
	)
	defer func() {
		f.mu.Lock()
		f.refreshing = false
		f.mu.Unlock()
	}()

	_ = SyncStateEvent.Broadcast(SyncState{Active: total > 0, Total: total})
	if total == 0 {
		return SyncSummary{}, nil
	}

	var (
		wg      sync.WaitGroup
		mu      sync.Mutex
		summary SyncSummary
		done    int
	)
	sem := make(chan struct{}, maxConcurrent)

	for _, feed := range feeds {
		wg.Add(1)
		go func(feed store.Feed) {
			defer wg.Done()

			sem <- struct{}{}
			defer func() { <-sem }()

			added, err := f.refreshOne(ctx, feed.ID)

			mu.Lock()
			done++
			state := SyncState{Active: true, Done: done, Total: total, FeedID: feed.ID, Feed: feed.Title}
			if err != nil {
				summary.Failed++
				state.Error = err.Error()
			} else {
				summary.Updated++
				summary.New += added
			}
			mu.Unlock()

			_ = SyncStateEvent.Broadcast(state)
		}(feed)
	}

	wg.Wait()
	_ = SyncStateEvent.Broadcast(SyncState{Total: total, Done: total})
	slog.Info("refresh all finished",
		"operation", "feeds.refreshAll",
		"updated", summary.Updated,
		"failed", summary.Failed,
		"new", summary.New,
	)
	return summary, nil
}

// refreshOne fetches a feed and returns how many posts were new. A feed that
// did not change keeps its posts; a feed that failed keeps its error, which the
// page shows.
func (f *Feeds) refreshOne(ctx context.Context, id int64) (int, error) {
	feed, err := f.store.Feed(id)
	if err != nil {
		slog.Warn("feed lookup failed",
			"operation", "feeds.refresh", "feedId", id, "error", err)
		return 0, err
	}

	started := time.Now()
	result, err := rss.Fetch(ctx, feed.URL, feed.ETag, feed.LastModified)
	if errors.Is(err, rss.ErrNotModified) {
		feed.LastFetchedAt = &started
		feed.LastError = ""
		return 0, f.store.UpdateFeed(feed)
	}
	if err != nil {
		feed.LastError = err.Error()
		if updateErr := f.store.UpdateFeed(feed); updateErr != nil {
			slog.Error("feed failure state update failed",
				"operation", "feeds.refresh", "feedId", feed.ID, "error", updateErr)
			return 0, updateErr
		}
		broadcastFeeds(f.store)
		slog.Warn("feed refresh failed",
			"operation", "feeds.refresh",
			"feedId", feed.ID,
			"feedHost", feedHost(feed.URL),
			"elapsedMs", time.Since(started).Milliseconds(),
			"error", redactedFeedError(feed.URL, err),
		)
		return 0, err
	}

	added, err := f.store.MergePosts(id, toPosts(result.Feed.Items))
	if err != nil {
		slog.Error("feed merge failed",
			"operation", "feeds.refresh", "feedId", feed.ID, "error", err)
		return 0, err
	}

	// MergePosts refreshed the unread count, so read the feed again.
	feed, err = f.store.Feed(id)
	if err != nil {
		return 0, err
	}

	feed.Title = titleOf(result.Feed, feed.URL)
	feed.SiteURL = firstNonEmpty(result.Feed.Meta.SiteURL, feed.SiteURL)
	feed.Description = firstNonEmpty(result.Feed.Meta.Description, feed.Description)
	feed.IconURL = firstNonEmpty(result.Feed.Meta.IconURL, feed.IconURL)
	feed.Language = firstNonEmpty(result.Feed.Meta.Language, feed.Language)
	if result.ETag != "" {
		feed.ETag = result.ETag
	}
	if result.LastModified != "" {
		feed.LastModified = result.LastModified
	}
	feed.LastFetchedAt = &started
	feed.LastError = ""
	if err := f.store.UpdateFeed(feed); err != nil {
		slog.Error("feed update failed",
			"operation", "feeds.refresh", "feedId", feed.ID, "error", err)
		return 0, err
	}

	broadcastFeeds(f.store)
	_ = PostsChangedEvent.Broadcast(PostsChanged{FeedID: id, New: added})
	return added, nil
}

func toPosts(items []rss.Item) []store.Post {
	posts := make([]store.Post, 0, len(items))
	for _, item := range items {
		posts = append(posts, store.Post{
			GUID:        item.GUID,
			Title:       item.Title,
			URL:         item.URL,
			Author:      item.Author,
			Summary:     item.Summary,
			Content:     item.Content,
			WordCount:   item.WordCount,
			PublishedAt: item.PublishedAt,
		})
	}
	return posts
}

func titleOf(feed *rss.Feed, feedURL string) string {
	if feed.Meta.Title != "" {
		return feed.Meta.Title
	}
	return feedURL
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}

// feedHost returns only the host of an address, keeping post and query data
// out of logs.
func feedHost(feedURL string) string {
	parsed, err := url.Parse(feedURL)
	if err != nil {
		return ""
	}
	return parsed.Host
}

// redactedFeedError replaces the complete feed address with its host while
// keeping the diagnostic part of the error message.
func redactedFeedError(feedURL string, err error) string {
	detail := err.Error()
	if host := feedHost(feedURL); host != "" {
		return strings.ReplaceAll(detail, feedURL, host)
	}
	return detail
}
