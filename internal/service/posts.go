package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"

	"github.com/egoist/mygo"

	appevent "github.com/sdttttt/ressu/internal/event"
	"github.com/sdttttt/ressu/internal/gui"
	"github.com/sdttttt/ressu/internal/media"
	"github.com/sdttttt/ressu/internal/store"
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

// normalize turns unknown and empty orders into the default order.
func (o SortOrder) normalize() SortOrder {
	if o == SortOrderAscending {
		return SortOrderAscending
	}
	return SortOrderDescending
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

// OpenImage validates and loads an article image in Go, then shows it in a
// child window. Remote images are opened only from the local cache; AVIF
// falls back to a child WebView because the native image decoder cannot read
// it.
func (p *Posts) OpenImage(ctx context.Context, source string, alt string) error {
	if p.media == nil {
		return errors.New("the image viewer is unavailable")
	}

	entry, err := p.media.Get(ctx, source)
	if err != nil {
		return fmt.Errorf("could not load the image: %w", err)
	}
	data, err := p.media.Read(entry)
	if err != nil {
		return fmt.Errorf("could not load the image: %w", err)
	}
	return gui.OpenImageViewer(mygo.CallerWindow(ctx), data, entry.ViewerURL(), alt)
}

// PostQuery selects a page of posts. FeedID zero selects all feeds. Search
// matches titles only. An unknown SortOrder selects the default newest-first
// order.
type PostQuery struct {
	FeedID      int64     `json:"feedId"`
	UnreadOnly  bool      `json:"unreadOnly"`
	StarredOnly bool      `json:"starredOnly"`
	Search      string    `json:"search"`
	Offset      int       `json:"offset"`
	Limit       int       `json:"limit"`
	SortOrder   SortOrder `json:"sortOrder"`
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
	var posts []store.Post
	var err error
	if query.FeedID == AggregateFeedID {
		posts, err = p.store.AllPosts()
	} else {
		posts, err = p.store.Posts(query.FeedID)
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
	if query.SortOrder.normalize() == SortOrderAscending {
		slices.SortStableFunc(matches, byPublishedAsc)
	}

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

// Get returns one post with its body.
func (p *Posts) Get(feedID int64, guid string) (store.Post, error) {
	post, err := p.store.Post(feedID, guid)
	if err != nil {
		slog.Warn("post lookup failed",
			"operation", "posts.get", "feedId", feedID, "error", err)
		return store.Post{}, err
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
	p.broadcastFeeds()
	return nil
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
	return post.ToSummary(), nil
}

// CaptureSnapshot captures the visible page for one scroll position.
func (p *Posts) CaptureSnapshot(ctx context.Context, request SnapshotCaptureRequest) ([]byte, error) {
	if err := request.validate(); err != nil {
		return nil, err
	}
	if request.tileCount() > maxSnapshotTiles {
		return nil, fmt.Errorf("the article is too long for one snapshot")
	}
	var data []byte
	var err error
	if page := mygo.CallerPage(ctx); page != nil && page.WebView() != nil {
		data, err = page.WebView().CapturePage()
	} else {
		window := mygo.CallerWindow(ctx)
		if window == nil || window.IsDestroyed() {
			err = fmt.Errorf("the snapshot window is no longer available")
		} else {
			data, err = window.CapturePage()
		}
	}
	if err != nil {
		slog.Warn("snapshot capture failed",
			"operation", "posts.captureSnapshot",
			"error", err,
		)
		return nil, err
	}
	return data, nil
}

// StitchSnapshot joins visible-page captures into one complete PNG.
func (p *Posts) StitchSnapshot(ctx context.Context, tiles []SnapshotTile) ([]byte, error) {
	data, err := stitchSnapshot(tiles, geometryOf(tiles))
	if err != nil {
		slog.Warn("snapshot stitch failed", "operation", "posts.stitchSnapshot", "error", err)
	}
	return data, err
}

// SaveSnapshot shows a save dialog and writes a stitched snapshot to the
// chosen path, which it returns. A cancelled dialog returns no path.
func (p *Posts) SaveSnapshot(ctx context.Context, fileName string, data []byte) (string, error) {
	if err := validateSnapshotImage(data); err != nil {
		return "", err
	}

	path, err := mygo.Dialog.Save(mygo.SaveDialogOptions{
		Parent:      mygo.CallerWindow(ctx),
		Title:       "保存文章快照",
		DefaultPath: snapshotFileName(fileName),
		Filters:     []mygo.FileFilter{{Name: "PNG 图片", Extensions: []string{"png"}}},
	})
	if err != nil {
		slog.Warn("snapshot save picker failed", "operation", "posts.saveSnapshot", "error", err)
		return "", err
	}
	if path == "" {
		return "", nil
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		slog.Error("snapshot write failed", "operation", "posts.saveSnapshot", "error", err)
		return "", fmt.Errorf("could not save the snapshot: %w", err)
	}
	return path, nil
}

// CopySnapshot puts a stitched snapshot on the clipboard.
func (p *Posts) CopySnapshot(ctx context.Context, data []byte) error {
	if err := validateSnapshotImage(data); err != nil {
		slog.Warn("invalid snapshot image", "operation", "posts.copySnapshot", "error", err)
		return err
	}
	if err := mygo.Clipboard.WriteImage(data); err != nil {
		slog.Warn("snapshot copy failed", "operation", "posts.copySnapshot", "error", err)
		return err
	}
	return nil
}

func postMatches(post store.Post, needle string) bool {
	return strings.Contains(strings.ToLower(post.Title), needle)
}

func byPublishedAsc(first, second store.PostSummary) int {
	return first.PublishedAt.Compare(second.PublishedAt)
}
