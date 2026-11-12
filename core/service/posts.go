package service

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image/png"
	"log/slog"
	"slices"
	"strings"
	"time"

	appevent "github.com/sdttttt/ressu/core/event"
	"github.com/sdttttt/ressu/core/logging"
	"github.com/sdttttt/ressu/core/media"
	"github.com/sdttttt/ressu/core/store"
)

// defaultLimit bounds a page of posts when the caller asks for none.
const defaultLimit = 100

// maxLimit bounds what a caller may ask for.
const maxLimit = 500

// AggregateFeedID asks for posts from every feed.
const AggregateFeedID = 0

// SortOrder determines whether Posts.List returns newest or oldest first.
type SortOrder string

// SortOrderDescending returns newer posts before older posts.
const SortOrderDescending SortOrder = "desc"

// SortOrderAscending returns older posts before newer posts.
const SortOrderAscending SortOrder = "asc"

// SortField determines which time Posts.List orders posts by: publication
// time or the time Ressu fetched the post.
type SortField string

// SortFieldPublished orders posts by publication time, the default.
const SortFieldPublished SortField = "published"

// SortFieldFetched orders posts by the time Ressu fetched the post.
const SortFieldFetched SortField = "fetched"

// normalize turns unknown and empty orders into the default order.
func (o SortOrder) normalize() SortOrder {
	if o == SortOrderAscending {
		return SortOrderAscending
	}
	return SortOrderDescending
}

// normalize turns unknown and empty fields into the default field.
func (f SortField) normalize() SortField {
	if f == SortFieldFetched {
		return SortFieldFetched
	}
	return SortFieldPublished
}

// Posts is the post service the frontend calls.
type Posts struct {
	store *store.Store
	media *media.Cache
	hub   *appevent.Hub
}

// NewPosts serves the posts of st.
func NewPosts(st *store.Store) *Posts {
	return NewPostsWithMediaAndHub(st, nil, nil)
}

// NewPostsWithHub serves posts and publishes changes through hub.
func NewPostsWithHub(st *store.Store, hub *appevent.Hub) *Posts {
	return NewPostsWithMediaAndHub(st, nil, hub)
}

// NewPostsWithMedia serves posts whose remote images are rewritten to local
// protocol URLs before they cross the binding boundary.
func NewPostsWithMedia(st *store.Store, mediaCache *media.Cache) *Posts {
	return NewPostsWithMediaAndHub(st, mediaCache, nil)
}

// NewPostsWithMediaAndHub serves posts and publishes changes through hub.
func NewPostsWithMediaAndHub(st *store.Store, mediaCache *media.Cache, hub *appevent.Hub) *Posts {
	return &Posts{store: st, media: mediaCache, hub: hub}
}

// broadcastFeeds publishes the complete subscription state.
func (p *Posts) broadcastFeeds() {
	if p.hub != nil {
		p.hub.FeedsChanged(feedsSnapshot(p.store))
	}
}

// ArticleImage is a validated article image loaded from the local cache.
type ArticleImage struct {
	Data      []byte `json:"data"`
	ViewerURL string `json:"viewerUrl"`
}

// LoadImage validates and loads an article image in Go. Remote images are
// loaded only from the local cache.
func (p *Posts) LoadImage(ctx context.Context, source string) (ArticleImage, error) {
	if p.media == nil {
		return ArticleImage{}, errors.New("the image viewer is unavailable")
	}

	entry, err := p.media.Get(ctx, source)
	if err != nil {
		return ArticleImage{}, fmt.Errorf("could not load the image: %w", err)
	}
	data, err := p.media.Read(entry)
	if err != nil {
		return ArticleImage{}, fmt.Errorf("could not load the image: %w", err)
	}
	return ArticleImage{Data: data, ViewerURL: entry.ViewerURL()}, nil
}

// PostQuery selects a page of posts. FeedID zero selects all feeds. Search
// matches titles only. An unknown SortOrder selects the default newest-first
// order, and an unknown SortField selects publication time.
type PostQuery struct {
	FeedID      int64     `json:"feedId"`
	UnreadOnly  bool      `json:"unreadOnly"`
	StarredOnly bool      `json:"starredOnly"`
	Search      string    `json:"search"`
	Offset      int       `json:"offset"`
	Limit       int       `json:"limit"`
	SortOrder   SortOrder `json:"sortOrder"`
	SortBy      SortField `json:"sortBy"`
}

// PostPage is a page of posts, with how many match the query and how many of
// the feed are unread.
type PostPage struct {
	Posts  []store.PostSummary `json:"posts"`
	Total  int                 `json:"total"`
	Unread int                 `json:"unread"`
}

// List returns the posts of a feed that match the query, newest first unless
// the query asks for oldest first.
func (p *Posts) List(query PostQuery) (PostPage, error) {
	var posts []store.PostSkim
	var err error
	if query.FeedID == AggregateFeedID {
		posts, err = p.store.AllPostSkims()
	} else {
		posts, err = p.store.PostSkims(query.FeedID)
	}
	if err != nil {
		slog.Warn("post list failed", "operation", "posts.list", "error", err)
		return PostPage{}, err
	}

	needle := strings.ToLower(strings.TrimSpace(query.Search))
	matches := make([]store.PostSummary, 0, len(posts))
	unread := 0
	for _, post := range posts {
		if !post.Read {
			unread++
		}
		if query.UnreadOnly && post.Read {
			continue
		}
		if query.StarredOnly && !post.Starred {
			continue
		}
		if needle != "" && !postMatches(post, needle) {
			continue
		}
		matches = append(matches, post.ToSummary())
	}
	sortPosts(matches, query)

	total := len(matches)
	offset := max(query.Offset, 0)
	limit := query.Limit
	if limit <= 0 {
		limit = defaultLimit
	}
	limit = min(limit, maxLimit)

	if offset >= total {
		return PostPage{Posts: []store.PostSummary{}, Total: total, Unread: unread}, nil
	}
	end := min(offset+limit, total)
	return PostPage{Posts: matches[offset:end], Total: total, Unread: unread}, nil
}

// Get returns one post with its body. When the preferences ask for it, an
// unread post is marked read before it is returned; a failed update does not
// block reading.
func (p *Posts) Get(feedID int64, guid string) (store.Post, error) {
	post, err := p.store.Post(feedID, guid)
	if err != nil {
		slog.Warn("post lookup failed",
			"operation", "posts.get", "feedId", feedID, "error", err)
		return store.Post{}, err
	}
	if !post.Read && p.store.Prefs().MarkReadOnOpen {
		updated, err := p.store.UpdatePost(feedID, guid, func(post *store.Post) { post.Read = true })
		if err != nil {
			slog.Warn("post read state update failed",
				"operation", "posts.get", "feedId", feedID, "error", err)
		} else {
			post = updated
			if p.hub != nil {
				p.hub.PostUpdated(post.ToSummary())
			}
			p.broadcastFeeds()
		}
	}
	if p.media != nil && p.store.Prefs().RemoteResources == store.ResourceAccessGoProxy {
		post.Content = p.media.RewriteHTML(post.Content, post.URL)
	}
	return post, nil
}

// MarkRead marks a post read or unread.
func (p *Posts) MarkRead(feedID int64, guid string, read bool) error {
	if _, err := p.store.UpdatePost(feedID, guid, func(post *store.Post) { post.Read = read }); err != nil {
		slog.Warn("post read state update failed",
			"operation", "posts.markRead", "feedId", feedID, "error", err)
		return err
	}
	if p.hub != nil {
		if post, err := p.store.Post(feedID, guid); err == nil {
			p.hub.PostUpdated(post.ToSummary())
		}
	}
	p.broadcastFeeds()
	return nil
}

// ToggleRead flips one post's read state and returns its new summary.
func (p *Posts) ToggleRead(feedID int64, guid string) (store.PostSummary, error) {
	post, err := p.store.UpdatePost(feedID, guid, func(post *store.Post) { post.Read = !post.Read })
	if err != nil {
		slog.Warn("post read state update failed",
			"operation", "posts.toggleRead", "feedId", feedID, "error", err)
		return store.PostSummary{}, err
	}
	if p.hub != nil {
		p.hub.PostUpdated(post.ToSummary())
	}
	p.broadcastFeeds()
	return post.ToSummary(), nil
}

// MarkAllRead marks every post of a feed read, and returns how many changed.
// The aggregate id marks every stored post.
func (p *Posts) MarkAllRead(feedID int64) (int, error) {
	if feedID == AggregateFeedID {
		changed := 0
		for _, feed := range p.store.Feeds() {
			feedChanged, err := p.store.MarkAllRead(feed.ID)
			if err != nil {
				slog.Error("all posts read update failed",
					"operation", "posts.markAllRead", "feedId", feed.ID, "error", err)
				return changed, err
			}
			changed += feedChanged
		}
		if changed > 0 {
			p.broadcastFeeds()
		}
		return changed, nil
	}

	changed, err := p.store.MarkAllRead(feedID)
	if err != nil {
		slog.Warn("all posts read update failed",
			"operation", "posts.markAllRead", "feedId", feedID, "error", err)
		return 0, err
	}
	if changed > 0 {
		p.broadcastFeeds()
	}
	return changed, nil
}

// ToggleStar stars or unstars a post, and returns it.
func (p *Posts) ToggleStar(feedID int64, guid string) (store.PostSummary, error) {
	post, err := p.store.UpdatePost(feedID, guid, func(post *store.Post) { post.Starred = !post.Starred })
	if err != nil {
		slog.Warn("post star update failed",
			"operation", "posts.toggleStar", "feedId", feedID, "error", err)
		return store.PostSummary{}, err
	}
	summary := post.ToSummary()
	if p.hub != nil {
		p.hub.PostUpdated(summary)
	}
	return summary, nil
}

// StitchSnapshot joins visible-page captures into one complete PNG.
func (p *Posts) StitchSnapshot(ctx context.Context, tiles []SnapshotTile) ([]byte, error) {
	started := time.Now()
	data, err := stitchSnapshot(tiles, geometryOf(tiles))
	if err != nil {
		slog.Warn("snapshot stitch failed", "operation", "posts.stitchSnapshot", "error", err)
		return data, err
	}
	args := []any{
		"operation", "posts.stitchSnapshot",
		"tiles", len(tiles),
		"bytes", len(data),
		"elapsedMs", time.Since(started).Milliseconds(),
	}
	if config, configErr := png.DecodeConfig(bytes.NewReader(data)); configErr == nil {
		args = append(args, "width", config.Width, "height", config.Height)
	}
	logging.Debug("snapshot stitched", args...)
	return data, nil
}

// sortPosts orders the list by the query's field and direction. Both are
// explicit: the store returns newest-published first, which only matches the
// published descending default.
func sortPosts(posts []store.PostSummary, query PostQuery) {
	var byTime func(first, second store.PostSummary) int
	if query.SortBy.normalize() == SortFieldFetched {
		byTime = func(first, second store.PostSummary) int {
			return first.FetchedAt.Compare(second.FetchedAt)
		}
	} else {
		byTime = func(first, second store.PostSummary) int {
			return first.PublishedAt.Compare(second.PublishedAt)
		}
	}
	if query.SortOrder.normalize() == SortOrderDescending {
		slices.SortStableFunc(posts, func(first, second store.PostSummary) int {
			return -byTime(first, second)
		})
		return
	}
	slices.SortStableFunc(posts, byTime)
}

func postMatches(post store.PostSkim, needle string) bool {
	return strings.Contains(strings.ToLower(post.Title), needle)
}
