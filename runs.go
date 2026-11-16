package main

// Jobs are the snapshots that were written, one job row each, newest
// first: {kind: job, id: job:STAMP, stamp, time, rows, changes, added,
// changed, removed}, where the counts are the diff from the snapshot
// before it (the first run's rows are all added; a shadow counts as
// added). A run that found no change wrote no snapshot, so it has no row.
// Runs are derived from the snapshot files, never stored in them: the
// counts are kept in a small index beside the tool's cache (jobindex.go),
// so a read parses only the snapshots the index has not seen.

// runRows are the job rows of the inventory directory's snapshots.
func (t *Tool) runRows() ([]Row, error) {
	stamps, err := t.snapshots()
	if err != nil {
		return nil, err
	}
	idx := t.loadJobIndex()
	// Snapshots are read one at a time and only two are held (the one
	// diffed and the one before it): every snapshot is a full inventory,
	// and a first fill must not take every core or all the memory.
	var prev []Row
	prevAt := -1
	dirty := 0
	rows := make([]Row, 0, len(stamps))
	read := func(i int) ([]Row, error) {
		if prevAt == i {
			return prev, nil
		}
		return t.readSnapshotRows(t.snapshotPath(stamps[i]))
	}
	var prevKey fileKey
	for i, stamp := range stamps {
		key, err := statKey(t.snapshotPath(stamp))
		if err != nil {
			return nil, err
		}
		want := jobEntry{File: key, Before: prevKey}
		prevKey = key
		if e, ok := idx.Jobs[stamp]; ok && e.File == want.File && e.Before == want.Before {
			rows = append(rows, jobRow(stamp, e))
			continue
		}
		cur, err := t.readSnapshotRows(t.snapshotPath(stamp))
		if err != nil {
			return nil, err
		}
		var before []Row
		if i > 0 {
			if before, err = read(i - 1); err != nil {
				return nil, err
			}
		}
		changes := Diff(before, cur, false)
		counts := map[string]int{}
		for _, c := range changes {
			verb := c.Verb
			if verb == "shadow" {
				verb = "added"
			}
			counts[verb]++
		}
		want.Rows, want.Changes = len(cur), len(changes)
		want.Added, want.Changed, want.Removed = counts["added"], counts["changed"], counts["removed"]
		idx.Jobs[stamp] = want
		rows = append(rows, jobRow(stamp, want))
		prev, prevAt = cur, i
		if dirty++; dirty%100 == 0 {
			t.saveJobIndex(idx, stamps) // a long first fill resumes where it stopped
		}
	}
	if dirty > 0 || len(idx.Jobs) != len(stamps) {
		t.saveJobIndex(idx, stamps)
	}
	for i, j := 0, len(rows)-1; i < j; i, j = i+1, j-1 {
		rows[i], rows[j] = rows[j], rows[i]
	}
	return rows, nil
}

func jobRow(stamp string, e jobEntry) Row {
	r := newRow("job", stamp)
	r["stamp"], r["time"] = stamp, stampText(stamp)
	r["rows"], r["changes"] = e.Rows, e.Changes
	r["added"], r["changed"], r["removed"] = e.Added, e.Changed, e.Removed
	return r
}
