package store

import (
	"maps"

	bolt "go.etcd.io/bbolt"
)

// State a tombstone keeps for a post the retention limit cut.
const (
	tombstoneRead    = 1 << iota
	tombstoneStarred = 1 << iota
)

// readTombstones returns the posts the retention limit cut from a feed,
// keyed by GUID with the state they had. A cut post is no longer stored, so
// this is the only place its state survives.
func readTombstones(tx *bolt.Tx, feedID int64) (map[string]byte, error) {
	states := map[string]byte{}
	parent := tx.Bucket([]byte(tombstonesBucket))
	if parent == nil {
		return states, nil
	}
	feed := parent.Bucket(feedKey(feedID))
	if feed == nil {
		return states, nil
	}
	err := feed.ForEach(func(guid, state []byte) error {
		if len(state) == 1 {
			states[string(guid)] = state[0]
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return states, nil
}

// writeTombstones makes cut the posts remembered for a feed, each with the
// state it has now. previous are the posts remembered before the merge, which
// covers every post the feed listed then: a post the feed no longer lists is
// forgotten, so the record cannot grow without end. Nothing is written when
// both are the same.
func writeTombstones(tx *bolt.Tx, feedID int64, cut []Post, previous map[string]byte) error {
	next := tombstoneStates(cut)
	if maps.Equal(next, previous) {
		return nil
	}
	parent, err := tx.CreateBucketIfNotExists([]byte(tombstonesBucket))
	if err != nil {
		return err
	}
	if err := deleteTombstones(tx, feedID); err != nil {
		return err
	}
	if len(next) == 0 {
		return nil
	}
	feed, err := parent.CreateBucketIfNotExists(feedKey(feedID))
	if err != nil {
		return err
	}
	for guid, state := range next {
		if err := feed.Put([]byte(guid), []byte{state}); err != nil {
			return err
		}
	}
	return nil
}

// deleteTombstones forgets the posts cut from a feed. Removing a subscription
// must do this, because the next subscription may reuse its id.
func deleteTombstones(tx *bolt.Tx, feedID int64) error {
	parent := tx.Bucket([]byte(tombstonesBucket))
	if parent == nil || parent.Bucket(feedKey(feedID)) == nil {
		return nil
	}
	return parent.DeleteBucket(feedKey(feedID))
}

// tombstoneStates returns the state of posts keyed by GUID.
func tombstoneStates(posts []Post) map[string]byte {
	states := make(map[string]byte, len(posts))
	for _, p := range posts {
		var state byte
		if p.Read {
			state |= tombstoneRead
		}
		if p.Starred {
			state |= tombstoneStarred
		}
		states[p.GUID] = state
	}
	return states
}
