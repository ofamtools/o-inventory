package main

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"ofam.dev/ochazuke"
	ocli "ofam.dev/ochazuke/cli"
	"ofam.dev/oden"
)

// cmdVerb is a CLI command that is also a verb of the main noun: one table,
// one handler. The verb builds the command's own words from its parameters
// (a boolean option when true, a valued one when given, the free words last)
// and runs the command against a copy of the tool whose stdout is captured,
// so the command and the verb cannot differ. The verb names are the
// lexicon's action words (a command named for a noun, such as snapshot, has
// the action that makes one); the command words keep working.
type cmdVerb struct {
	Verb, Command string
	Prefix        []string // words the command takes first (hook install)
	Summary       string
	History       string
	Bools, Valued []string
	Words         string // what the free words are; "" for none
}

var cmdVerbs = []cmdVerb{
	{Verb: "record", Command: "snapshot", Summary: "collect every source and write a snapshot when a compared field differs (command snapshot)", History: ochazuke.HistoryChange, Bools: []string{"commit", "quiet"}},
	{Verb: "diff", Command: "diff", Summary: "the changes between snapshots A and B (stamps, latest or paths; default the previous and the latest)", History: ochazuke.HistoryFetch, Bools: []string{"jsonl", "all"}, Words: "snapshot stamps or paths: A [B]"},
	{Verb: "merge", Command: "changelog", Summary: "merge the snapshots' entries into the hub CHANGELOG.md; repo prints one repo's; write-repos also writes each repo's, never committed (command changelog)", History: ochazuke.HistoryChange, Bools: []string{"write-repos", "commit", "quiet"}, Valued: []string{"repo"}},
	{Verb: "check", Command: "check", Summary: "drift: tools without a catalog, stale or missing hosts, missing project trees, repos without the hook; fails when anything is wrong", History: ochazuke.HistoryFetch, Bools: []string{"jsonl"}},
	{Verb: "export", Command: "repos", Summary: "the family repos config (repos.json, seeded from the hub's o links on first use) as JSON (command repos)", History: ochazuke.HistoryFetch},
	{Verb: "install", Command: "hook", Prefix: []string{"install"}, Summary: "install the post-commit hook in a repo, or in every family repo with all (command hook install)", History: ochazuke.HistoryChange, Bools: []string{"all"}, Words: "the repo to install the hook in"},
	{Verb: "inspect", Command: "errors", Summary: "every error site of a Go module with its function, verb, conditions and a heuristic class; Markdown, or jsonl and md write files (command errors)", History: ochazuke.HistoryFetch, Valued: []string{"jsonl", "md"}, Words: "the module root, then name patterns: ROOT [PATTERN...]"},
}

// wordsArg is the free-words parameter: a list, or one string of
// space-separated words.
func wordsArg(a ochazuke.Arguments) []string {
	switch v := a["words"].(type) {
	case string:
		return strings.Fields(v)
	case []string:
		return v
	case []any:
		var out []string
		for _, w := range v {
			out = append(out, fmt.Sprint(w))
		}
		return out
	}
	return nil
}

// argv is the command line a verb's arguments stand for.
func (v cmdVerb) argv(a ochazuke.Arguments) []string {
	out := append([]string(nil), v.Prefix...)
	for _, b := range v.Bools {
		if on, _ := a[b].(bool); on {
			out = append(out, "--"+b)
		}
	}
	for _, name := range v.Valued {
		if s := ochazuke.StringArg(a, name, ""); s != "" {
			out = append(out, "--"+name+"="+s)
		}
	}
	return append(out, wordsArg(a)...)
}

// runCommand runs the named command of this tool with its stdout captured.
func (t *Tool) runCommand(ctx context.Context, name string, argv []string) (ocli.Lines, error) {
	var buf bytes.Buffer
	tt := *t
	tt.Stdout = &buf
	var run func(context.Context, []string) error
	for _, c := range findCommands(tt.commands(), name) {
		run = c.Run
	}
	if run == nil {
		return nil, fmt.Errorf("no command %q", name)
	}
	err := run(ctx, argv)
	var out ocli.Lines
	if text := strings.TrimRight(buf.String(), "\n"); text != "" {
		for _, line := range strings.Split(text, "\n") {
			out = append(out, line)
		}
	}
	if err != nil && len(out) > 0 {
		err = fmt.Errorf("%s\n%w", strings.Join(lineStrings(out), "\n"), err)
	}
	return out, err
}

func findCommands(cs []oden.Command, name string) []oden.Command {
	var out []oden.Command
	for _, c := range cs {
		if c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

func lineStrings(l ocli.Lines) []string {
	out := make([]string, len(l))
	for i, x := range l {
		out[i] = fmt.Sprint(x)
	}
	return out
}

// defineCommandVerbs gives the main noun a verb for each command in cmdVerbs.
func (t *Tool) defineCommandVerbs(r *ochazuke.Runtime) {
	inv := r.Define("inventory")
	for _, v := range cmdVerbs {
		inv.Verb(v.Verb, func(c ochazuke.Ctx) ochazuke.Result {
			out, err := t.runCommand(context.Background(), v.Command, v.argv(c.Args))
			if err != nil {
				return ochazuke.Failed(err)
			}
			return ochazuke.Done(out)
		}, v.Summary)
		setHistory(r, v.History, "inventory."+v.Verb)
		op := r.Operations.At(ochazuke.Key("inventory." + v.Verb))
		op.Parameters = map[string]ochazuke.Parameter{}
		for _, b := range v.Bools {
			op.Parameters[b] = ochazuke.Parameter{Type: "boolean", Description: "the command's --" + b}
		}
		for _, n := range v.Valued {
			op.Parameters[n] = ochazuke.Parameter{Type: "string", Description: "the command's --" + n}
		}
		if v.Words != "" {
			op.Parameters["words"] = ochazuke.Parameter{Type: "string", Description: v.Words}
		}
		r.Operations.Set(op)
	}
}
