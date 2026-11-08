package main

import (
	"encoding/json"
	"testing"

	"ofam.dev/ochazuke"
	ocli "ofam.dev/ochazuke/cli"
)

func TestCatalogDeclaresSingularNamespace(t *testing.T) {
	f := newFixture(t)
	var doc ocli.CatalogDoc
	if err := json.Unmarshal([]byte(f.mustRun("catalog")), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.Namespace != "inventory" || ochazuke.DefaultNamespace != "inventory" {
		t.Fatalf("catalog namespace %q, default %q; want inventory", doc.Namespace, ochazuke.DefaultNamespace)
	}
	if got, e := ocli.CatalogNamespace(doc, "o-inventory"); e != nil || got != "inventory" {
		t.Fatal(got, e)
	}
}
