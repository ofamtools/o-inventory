package main

import "ofam.dev/ochazuke"

// Namespace is the singular namespace o-inventory declares (lexicon plan): its
// operations are inventory.<noun>.<verb> and a host mounts its catalog under it.
const Namespace = "inventory"

func init() {
	if err := ochazuke.DeclareNamespace(Namespace); err != nil {
		panic(err)
	}
}
