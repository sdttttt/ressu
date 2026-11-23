package store

import (
	"encoding/json"
	"log/slog"

	bolt "go.etcd.io/bbolt"

	"github.com/sdttttt/ressu/core/internal/sanitize"
	"github.com/sdttttt/ressu/core/internal/textmetrics"
)

// migrate upgrades stored data to the current schema version.
func migrate(tx *bolt.Tx) error {
	settings := tx.Bucket([]byte(settingsBucket))
	if settings == nil {
		return nil
	}

	var version int
	found, err := get(settings, []byte(schemaVersionKey), &version)
	if err != nil {
		return err
	}
	if found && version >= currentSchemaVersion {
		return nil
	}
	if found && version < currentSchemaVersion {
		slog.Info("database schema migrating",
			"operation", "store.migrate",
			"from", version,
			"to", currentSchemaVersion,
		)
	}
	if version < 2 {
		prefs := DefaultPrefs()
		if _, err := get(settings, []byte(prefsKey), &prefs); err != nil {
			return err
		}
		if err := sanitizeStoredPosts(tx, prefs.StripTrackers); err != nil {
			return err
		}
	}
	if version < 3 {
		if _, err := tx.CreateBucketIfNotExists([]byte(iconsBucket)); err != nil {
			return err
		}
	}
	if version < 4 {
		if err := backfillPostWordCounts(tx); err != nil {
			return err
		}
	}
	if version < 5 {
		if err := initFeedOrder(tx); err != nil {
			return err
		}
	}
	return put(settings, []byte(schemaVersionKey), currentSchemaVersion)
}

// initFeedOrder seeds the subscription order of databases saved before the
// order was persisted. The id order of the feeds bucket is the order the
// sidebar always showed, so it becomes the initial order.
func initFeedOrder(tx *bolt.Tx) error {
	settings := tx.Bucket([]byte(settingsBucket))
	var stored []int64
	found, err := get(settings, []byte(feedOrderKey), &stored)
	if err != nil {
		return err
	}
	if found {
		return nil
	}

	feeds := tx.Bucket([]byte(feedsBucket))
	if feeds == nil {
		return nil
	}
	order := make([]int64, 0, 64)
	if err := feeds.ForEach(func(_, value []byte) error {
		var f Feed
		if err := json.Unmarshal(value, &f); err != nil {
			return err
		}
		order = append(order, f.ID)
		return nil
	}); err != nil {
		return err
	}
	return put(settings, []byte(feedOrderKey), order)
}

// backfillPostWordCounts counts the readable text of posts saved before the
// metric was added.
func backfillPostWordCounts(tx *bolt.Tx) error {
	parent := tx.Bucket([]byte(postsBucket))
	if parent == nil {
		return nil
	}
	return parent.ForEach(func(key, value []byte) error {
		if value != nil {
			return nil
		}
		posts := parent.Bucket(key)
		if posts == nil {
			return nil
		}
		return posts.ForEach(func(key, value []byte) error {
			if value == nil {
				return nil
			}
			var post Post
			if err := json.Unmarshal(value, &post); err != nil {
				return err
			}
			post.WordCount = textmetrics.Count(post.Content)
			if post.WordCount == 0 {
				post.WordCount = textmetrics.Count(post.Summary)
			}
			return put(posts, key, post)
		})
	})
}

// sanitizeStoredPosts rewrites HTML saved before content was cleaned on
// parse. strip also removes known tracking parameters from outbound links.
func sanitizeStoredPosts(tx *bolt.Tx, strip bool) error {
	parent := tx.Bucket([]byte(postsBucket))
	if parent == nil {
		return nil
	}
	return parent.ForEach(func(key, value []byte) error {
		if value != nil {
			return nil
		}
		posts := parent.Bucket(key)
		if posts == nil {
			return nil
		}
		return posts.ForEach(func(key, value []byte) error {
			if value == nil {
				return nil
			}
			var post Post
			if err := json.Unmarshal(value, &post); err != nil {
				return err
			}
			cleaned := sanitize.HTML(post.Content, post.URL, strip)
			if cleaned == post.Content {
				return nil
			}
			post.Content = cleaned
			return put(posts, key, post)
		})
	})
}
