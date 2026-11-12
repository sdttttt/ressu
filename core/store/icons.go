package store

import bolt "go.etcd.io/bbolt"

// readIcon loads the raw image and metadata of a feed.
func readIcon(tx *bolt.Tx, feedID int64) ([]byte, IconMeta, error) {
	feeds := tx.Bucket([]byte(iconsBucket))
	if feeds == nil {
		return nil, IconMeta{}, ErrNotFound
	}
	feed := feeds.Bucket(feedKey(feedID))
	if feed == nil {
		return nil, IconMeta{}, ErrNotFound
	}

	data := feed.Get([]byte(iconDataKey))
	if data == nil {
		return nil, IconMeta{}, ErrNotFound
	}

	var meta IconMeta
	found, err := get(feed, []byte(iconMetaKey), &meta)
	if err != nil {
		return nil, IconMeta{}, err
	}
	if !found {
		return nil, IconMeta{}, ErrNotFound
	}
	return data, meta, nil
}

// deleteIcon removes a feed's cached image and metadata.
func deleteIcon(tx *bolt.Tx, feedID int64) error {
	feeds := tx.Bucket([]byte(iconsBucket))
	if feeds == nil {
		return nil
	}
	feed := feeds.Bucket(feedKey(feedID))
	if feed == nil {
		return nil
	}
	return feeds.DeleteBucket(feedKey(feedID))
}
