package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// An older snapshot spells a tool row's state source "state"; it reads as
// "lists" and the next write is "lists" (a data migration, not an alias).
func TestOlderSnapshotStateKeyReadsAsLists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "20260101T000000Z.jsonl")
	old := `{"catalog":true,"id":"tool:o-alpha","kind":"tool","state":"skill.state","version":"0.1.0"}` + "\n" +
		`{"id":"project:a","kind":"project","status":"active"}` + "\n"
	if err := os.WriteFile(path, []byte(old), 0o644); err != nil {
		t.Fatal(err)
	}
	rows, err := readRows(path)
	if err != nil {
		t.Fatal(err)
	}
	byid := byID(rows)
	tool := byid["tool:o-alpha"]
	if tool.Str("lists") != "skill.state" || tool["state"] != nil || byid["project:a"].Str("status") != "active" {
		t.Fatalf("read back %v / %v", tool, byid["project:a"])
	}
	out, err := encodeRows(rows)
	if err != nil || !strings.Contains(string(out), `"lists":"skill.state"`) || strings.Contains(string(out), `"state"`) {
		t.Fatalf("written back %s (%v)", out, err)
	}
}

// daily status is a prop of the daily job, read as a value; the daily verbs
// that change something stay verbs.
func TestDailyStatusIsAProp(t *testing.T) {
	f := newFixture(t)
	out := f.mustRun("daily", "get", "--prop", "status")
	if !strings.Contains(out, "plist:") {
		t.Fatalf("daily status prop: %s", out)
	}
	if _, err := f.run("daily", "install", "--help"); err != nil {
		t.Fatalf("daily install is a verb: %v", err)
	}
}
