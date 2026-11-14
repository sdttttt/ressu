package app

import (
	"github.com/sdttttt/ressu/core/internal/buildinfo"
	"github.com/sdttttt/ressu/core/internal/event"
	"github.com/sdttttt/ressu/core/internal/logging"
	"github.com/sdttttt/ressu/core/internal/media"
	"github.com/sdttttt/ressu/core/internal/service"
	"github.com/sdttttt/ressu/core/internal/status"
	"github.com/sdttttt/ressu/core/internal/store"
)

// The aliases and variables below re-export the internal core packages so
// that core/app is the only import surface available to frontends. They keep
// the original names except where the internal packages collide
// (OpenStore/OpenMedia/OpenLog, StatusFeedsSnapshot).

// Store types and preferences.
type (
	// Prefs are the persisted user preferences.
	Prefs = store.Prefs
	// Feed is one subscription with its metadata and unread count.
	Feed = store.Feed
	// Post is one stored article.
	Post = store.Post
	// PostSummary is the list representation of a post.
	PostSummary = store.PostSummary
	// Icon is a cached feed icon.
	Icon = store.Icon
)

// Store preference and theme constants.
const (
	FontSizeSmall  = store.FontSizeSmall
	FontSizeMedium = store.FontSizeMedium
	FontSizeLarge  = store.FontSizeLarge

	ThemeSystem = store.ThemeSystem
	ThemeLight  = store.ThemeLight
	ThemeDark   = store.ThemeDark
	ThemeSepia  = store.ThemeSepia

	PostSortPublished = store.PostSortPublished
	PostSortFetched   = store.PostSortFetched

	OpenLinksBrowser = store.OpenLinksBrowser
	OpenLinksWindow  = store.OpenLinksWindow

	ResourceAccessWebview = store.ResourceAccessWebview
	ResourceAccessGoProxy = store.ResourceAccessGoProxy
)

var (
	// ErrNotFound is returned when a stored record does not exist.
	ErrNotFound = store.ErrNotFound
	// DefaultPrefs are the preferences of a new install.
	DefaultPrefs = store.DefaultPrefs
	// OpenStore opens the database in dir.
	OpenStore = store.Open
)

// Service types for feeds, posts, and settings.
type (
	// FeedPreview is the validation result for a subscription URL.
	FeedPreview = service.FeedPreview
	// FeedsSnapshot is the subscription list with the total unread count.
	FeedsSnapshot = service.FeedsSnapshot
	// ForegroundRefresher triggers stale refreshes when the app returns to
	// the foreground.
	ForegroundRefresher = service.ForegroundRefresher
	// ImportSummary reports the result of an OPML import.
	ImportSummary = service.ImportSummary
	// PostPage is one page of the post list.
	PostPage = service.PostPage
	// PostQuery filters and pages the post list.
	PostQuery = service.PostQuery
	// SnapshotCaptureRequest is one scroll-position capture of the reader
	// page.
	SnapshotCaptureRequest = service.SnapshotCaptureRequest
	// SnapshotTile is one stitched snapshot tile.
	SnapshotTile = service.SnapshotTile
	// SortField is a post list sort column.
	SortField = service.SortField
	// StorageInfo reports the database and media cache sizes.
	StorageInfo = service.StorageInfo
	// SyncSummary reports the result of a refresh batch.
	SyncSummary = service.SyncSummary
)

// Service constructors and errors.
var (
	// ErrRefreshing reports a refresh already in progress.
	ErrRefreshing = service.ErrRefreshing
	// SortOrderAscending sorts oldest first.
	SortOrderAscending = service.SortOrderAscending
	// SortOrderDescending sorts newest first.
	SortOrderDescending = service.SortOrderDescending

	NewFeeds                = service.NewFeeds
	NewFeedsWithHub         = service.NewFeedsWithHub
	NewPosts                = service.NewPosts
	NewPostsWithMediaAndHub = service.NewPostsWithMediaAndHub
	NewSettings             = service.NewSettings
	NewSettingsWithMedia    = service.NewSettingsWithMedia
	NewStatus               = service.NewStatus
	NewForegroundRefresher  = service.NewForegroundRefresher
)

const (
	// AggregateFeedID addresses the virtual feed that aggregates all posts.
	AggregateFeedID = service.AggregateFeedID
)

// Status mirror types.
type (
	// Model is the in-memory mirror of the service state.
	Model = status.Model
	// PostSelection identifies the selected post.
	PostSelection = status.PostSelection
	// StatusFeedsSnapshot is the subscription list kept by the status mirror.
	StatusFeedsSnapshot = status.FeedsSnapshot
	// Snapshot is a consistent read of the status mirror.
	Snapshot = status.Snapshot
)

// NewStatusModel creates the status mirror model.
var NewStatusModel = status.New

// Event types and constructor.
type (
	// Event is the unified broadcast envelope.
	Event = event.Event
	// Hub publishes events serially to its sink.
	Hub = event.Hub
)

// NewHub creates the event hub on top of the status mirror.
var NewHub = event.NewHub

// Media cache.
const (
	// MediaScheme is the local protocol used by rewritten article images.
	MediaScheme = media.Scheme
)

var (
	// ContentType sniffs an image content type.
	ContentType = media.ContentType
	// IsAVIF reports whether data is an AVIF image.
	IsAVIF = media.IsAVIF
	// OpenMedia opens the article image cache under dir.
	OpenMedia = media.Open
)

// MediaOptions configure the article image cache.
type MediaOptions = media.Options

// Logging.
var (
	// OpenLog opens the optional log file in the data directory.
	OpenLog = logging.Open
)

// LogOptions configure the log file.
type LogOptions = logging.Options

// Date returns the build date recorded in the running binary.
var Date = buildinfo.Date
