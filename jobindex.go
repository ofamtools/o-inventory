package main

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// The job index keeps each snapshot's job counts so a read does not parse
// the whole history again. It is derived data: deleting it only costs the
// time to rebuild it, and it lives in the cache directory
// ($XDG_CACHE_HOME/o-inventory/jobs.json, else ~/.cache/o-inventory), not
// in the hub's committed inventory directory.

// fileKey identifies a snapshot file's content: snapshots are written once
// under a new name, so size and mtime change only if one is rewritten.
type fileKey struct {
	Size  int64 `json:"size"`
	Mtime int64 `json:"mtime"` // UnixNano
}

// jobEntry is one job's counts and the files they were derived from: the
// snapshot and the one before it (the diff's two sides).
type jobEntry struct {
	File    fileKey `json:"file"`
	Before  fileKey `json:"before"`
	Rows    int     `json:"rows"`
	Changes int     `json:"changes"`
	Added   int     `json:"added"`
	Changed int     `json:"changed"`
	Removed int     `json:"removed"`
}

type jobIndex struct {
	Jobs map[string]jobEntry `json:"jobs"`
}

func statKey(path string) (fileKey, error) {
	info, err := os.Stat(path)
	if err != nil {
		return fileKey{}, err
	}
	return fileKey{Size: info.Size(), Mtime: info.ModTime().UnixNano()}, nil
}

// CacheDir is $XDG_CACHE_HOME/o-inventory, else ~/.cache/o-inventory.
func (t *Tool) CacheDir() string {
	base := t.getenv("XDG_CACHE_HOME")
	if base == "" {
		base = filepath.Join(t.Home, ".cache")
	}
	return filepath.Join(base, "o-inventory")
}

func (t *Tool) jobIndexPath() string { return filepath.Join(t.CacheDir(), "jobs.json") }

// loadJobIndex reads the index; a missing or unreadable file is an empty
// index (it is rebuilt from the snapshots).
func (t *Tool) loadJobIndex() *jobIndex {
	idx := &jobIndex{}
	if b, err := os.ReadFile(t.jobIndexPath()); err == nil {
		if json.Unmarshal(b, idx) != nil {
			idx = &jobIndex{}
		}
	}
	if idx.Jobs == nil {
		idx.Jobs = map[string]jobEntry{}
	}
	return idx
}

// saveJobIndex writes the index, keeping only the stamps that still exist.
// A write that fails costs only the next read's time.
func (t *Tool) saveJobIndex(idx *jobIndex, stamps []string) {
	keep := make(map[string]jobEntry, len(stamps))
	for _, s := range stamps {
		if e, ok := idx.Jobs[s]; ok {
			keep[s] = e
		}
	}
	idx.Jobs = keep
	if b, err := json.Marshal(idx); err == nil {
		_ = writeFile(t.jobIndexPath(), b)
	}
}

// readSnapshotRows is readRows unless a test counts the reads.
func (t *Tool) readSnapshotRows(path string) ([]Row, error) {
	if t.ReadSnapshot != nil {
		return t.ReadSnapshot(path)
	}
	return readRows(path)
}
