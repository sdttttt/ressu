package store

import (
	"encoding/json"

	"github.com/sdttttt/ressu/internal/sanitize"
	"github.com/sdttttt/ressu/internal/textmetrics"
	bolt "go.etcd.io/bbolt"
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
	if version < 2 {
		if err := sanitizeStoredPosts(tx); err != nil {
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
	return put(settings, []byte(schemaVersionKey), currentSchemaVersion)
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

// sanitizeStoredPosts rewrites HTML saved before content was cleaned on parse.
func sanitizeStoredPosts(tx *bolt.Tx) error {
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
			cleaned := sanitize.HTML(post.Content, post.URL)
			if cleaned == post.Content {
				return nil
			}
			post.Content = cleaned
			return put(posts, key, post)
		})
	})
}
