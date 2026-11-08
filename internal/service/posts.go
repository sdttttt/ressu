package service

import (
	"slices"
	"strings"

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
}

// NewPosts serves the posts of st.
func NewPosts(st *store.Store) *Posts {
	return &Posts{store: st}
}

// PostQuery selects a page of posts. FeedID zero selects all feeds. An
// unknown SortOrder selects the default newest-first order.
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
	return p.store.Post(feedID, guid)
}

// MarkRead marks a post read or unread.
func (p *Posts) MarkRead(feedID int64, guid string, read bool) error {
	if _, err := p.store.UpdatePost(feedID, guid, func(post *store.Post) { post.Read = read }); err != nil {
		return err
	}
	broadcastFeeds(p.store)
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
				return changed, err
			}
			changed += feedChanged
		}
		if changed > 0 {
			broadcastFeeds(p.store)
		}
		return changed, nil
	}

	changed, err := p.store.MarkAllRead(feedID)
	if err != nil {
		return 0, err
	}
	if changed > 0 {
		broadcastFeeds(p.store)
	}
	return changed, nil
}

// ToggleStar stars or unstars a post, and returns it.
func (p *Posts) ToggleStar(feedID int64, guid string) (store.PostSummary, error) {
	post, err := p.store.UpdatePost(feedID, guid, func(post *store.Post) { post.Starred = !post.Starred })
	if err != nil {
		return store.PostSummary{}, err
	}
	return post.ToSummary(), nil
}

func postMatches(post store.Post, needle string) bool {
	return strings.Contains(strings.ToLower(post.Title), needle) ||
		strings.Contains(strings.ToLower(post.Summary), needle) ||
		strings.Contains(strings.ToLower(post.Author), needle)
}

func byPublishedAsc(first, second store.PostSummary) int {
	return first.PublishedAt.Compare(second.PublishedAt)
}
