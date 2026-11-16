package main

import (
	"os"
	"reflect"
	"testing"
)

// The job index gives the rows a full read gives, and a second read parses
// no snapshot it has already summarised.
func TestJobIndexKeepsRowsAndSkipsIndexedSnapshots(t *testing.T) {
	f := tableFixture(t)
	reads := 0
	tool := f.tool0()
	tool.ReadSnapshot = func(path string) ([]Row, error) { reads++; return readRows(path) }

	full, err := tool.runRows() // no index yet: reads every snapshot
	if err != nil || reads == 0 {
		t.Fatalf("first read: %v, %d reads", err, reads)
	}
	if _, err := os.Stat(tool.jobIndexPath()); err != nil {
		t.Fatalf("index not written: %v", err)
	}
	reads = 0
	again, err := tool.runRows()
	if err != nil || reads != 0 {
		t.Fatalf("second read: %v, %d snapshot reads, want 0", err, reads)
	}
	if !reflect.DeepEqual(full, again) {
		t.Errorf("indexed rows differ:\n%v\n%v", full, again)
	}
	if err := os.Remove(tool.jobIndexPath()); err != nil { // rebuildable
		t.Fatal(err)
	}
	rebuilt, err := tool.runRows()
	if err != nil || !reflect.DeepEqual(full, rebuilt) {
		t.Errorf("rebuilt rows differ: %v\n%v\n%v", err, full, rebuilt)
	}
}
