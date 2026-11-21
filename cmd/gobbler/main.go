// Command gobbler reads, converts, matches and (soon) syncs playlists.
//
// Every subcommand does one thing and speaks JSON on stdin/stdout so they can
// be chained:
//
//	gobbler read  --format aimppl4 Road.aimppl4 | gobbler match --library lib.json --map 'M:\Music=/data/music' | ...
//	gobbler read  Road.aimppl4 | gobbler write --format m3u8 > Road.m3u8
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sys7mfaul7/gobbler/pkg/format"
	"github.com/sys7mfaul7/gobbler/pkg/match"
	"github.com/sys7mfaul7/gobbler/pkg/playlist"
	"github.com/sys7mfaul7/gobbler/pkg/redact"
	"github.com/sys7mfaul7/gobbler/pkg/target"
)

// version is set at build time with -ldflags "-X main.version=...".
var version = "dev"

const usage = `gobbler - portable playlist converter and sync engine

Usage:
  gobbler read    [--format NAME] [FILE]        playlist file (or stdin) -> JSON
  gobbler write   --format NAME                 JSON on stdin -> playlist file on stdout
  gobbler convert [--from NAME] --to NAME [FILE] read + write in one step
  gobbler match   --library FILE [--map FROM=TO ...] [--tolerance 2s]
                                                JSON playlist on stdin -> match report JSON
  gobbler sync    --from STORE --to TARGET      (not implemented yet)
  gobbler formats                               list format adapters
  gobbler version

Environment:
  GOBBLER_FORMAT   default input format for read/convert
  GOBBLER_MAP      ';'-separated FROM=TO path rules for match
`

func main() {
	if err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "gobbler:", redact.String(err.Error()))
		var ue usageError
		if errors.As(err, &ue) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}

type usageError struct{ msg string }

func (u usageError) Error() string { return u.msg }

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return usageError{"missing subcommand"}
	}
	cmd, args := args[0], args[1:]
	switch cmd {
	case "read":
		return cmdRead(args, stdin, stdout)
	case "write":
		return cmdWrite(args, stdin, stdout)
	case "convert":
		return cmdConvert(args, stdin, stdout)
	case "match":
		return cmdMatch(args, stdin, stdout, stderr)
	case "sync":
		return errors.New("sync is not implemented yet; see docs/DESIGN.md")
	case "formats":
		for _, n := range format.Names() {
			f, _ := format.Lookup(n)
			fmt.Fprintf(stdout, "%-8s read=%-5v write=%-5v %s\n", n, f.Reader != nil, f.Writer != nil, strings.Join(f.Extensions, " "))
		}
		return nil
	case "version", "--version", "-v":
		fmt.Fprintln(stdout, "gobbler", version)
		return nil
	case "help", "-h", "--help":
		fmt.Fprint(stdout, usage)
		return nil
	}
	fmt.Fprint(stderr, usage)
	return usageError{"unknown subcommand " + cmd}
}

func newFlags(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs
}

// readPlaylist reads from the file named in fs.Arg(0), or stdin when absent.
func readPlaylist(name, file string, stdin io.Reader) (*playlist.Playlist, error) {
	if name == "" {
		name = os.Getenv("GOBBLER_FORMAT")
	}
	var f *format.Format
	var err error
	switch {
	case name != "":
		f, err = format.Lookup(name)
	case file != "":
		f, err = format.ForPath(file)
	default:
		return nil, usageError{"--format is required when reading stdin"}
	}
	if err != nil {
		return nil, err
	}
	if f.Reader == nil {
		return nil, fmt.Errorf("format %s cannot be read", f.Name)
	}
	in := stdin
	if file != "" && file != "-" {
		fh, err := os.Open(file)
		if err != nil {
			return nil, err
		}
		defer fh.Close()
		in = fh
	}
	return f.Reader.Read(in)
}

func writeJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}

func cmdRead(args []string, stdin io.Reader, stdout io.Writer) error {
	fs := newFlags("read")
	name := fs.String("format", "", "input format")
	if err := fs.Parse(args); err != nil {
		return usageError{err.Error()}
	}
	p, err := readPlaylist(*name, fs.Arg(0), stdin)
	if err != nil {
		return err
	}
	return writeJSON(stdout, p)
}

func writePlaylist(name string, p *playlist.Playlist, stdout io.Writer) error {
	f, err := format.Lookup(name)
	if err != nil {
		return err
	}
	if f.Writer == nil {
		return fmt.Errorf("format %s cannot be written yet", f.Name)
	}
	return f.Writer.Write(stdout, p)
}

func cmdWrite(args []string, stdin io.Reader, stdout io.Writer) error {
	fs := newFlags("write")
	name := fs.String("format", "", "output format")
	if err := fs.Parse(args); err != nil {
		return usageError{err.Error()}
	}
	if *name == "" {
		return usageError{"--format is required"}
	}
	var p playlist.Playlist
	if err := json.NewDecoder(stdin).Decode(&p); err != nil {
		return fmt.Errorf("decoding playlist JSON: %w", err)
	}
	return writePlaylist(*name, &p, stdout)
}

func cmdConvert(args []string, stdin io.Reader, stdout io.Writer) error {
	fs := newFlags("convert")
	from := fs.String("from", "", "input format")
	to := fs.String("to", "", "output format")
	if err := fs.Parse(args); err != nil {
		return usageError{err.Error()}
	}
	if *to == "" {
		return usageError{"--to is required"}
	}
	p, err := readPlaylist(*from, fs.Arg(0), stdin)
	if err != nil {
		return err
	}
	return writePlaylist(*to, p, stdout)
}

type multiFlag []string

func (m *multiFlag) String() string     { return strings.Join(*m, ";") }
func (m *multiFlag) Set(s string) error { *m = append(*m, s); return nil }

func cmdMatch(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := newFlags("match")
	lib := fs.String("library", "", "JSON file with an array of library items")
	tol := fs.Duration("tolerance", 0, "duration tolerance for metadata matches (default 2s)")
	var maps multiFlag
	fs.Var(&maps, "map", "path rule FROM=TO (repeatable)")
	if err := fs.Parse(args); err != nil {
		return usageError{err.Error()}
	}
	if *lib == "" {
		return usageError{"--library is required"}
	}
	if env := os.Getenv("GOBBLER_MAP"); env != "" {
		maps = append(strings.Split(env, ";"), maps...)
	}
	var pm match.PathMap
	for _, s := range maps {
		r, ok := match.ParseRule(s)
		if !ok {
			return usageError{"bad --map rule " + s}
		}
		pm = append(pm, r)
	}
	b, err := os.ReadFile(*lib)
	if err != nil {
		return err
	}
	var items []target.Item
	if err := json.Unmarshal(b, &items); err != nil {
		return fmt.Errorf("decoding library JSON: %w", err)
	}
	var p playlist.Playlist
	if err := json.NewDecoder(stdin).Decode(&p); err != nil {
		return fmt.Errorf("decoding playlist JSON: %w", err)
	}
	m := match.New(items, pm)
	m.Tolerance = *tol
	res := m.Match(&p)
	unmatched := 0
	for i := range res {
		// Never echo credential-bearing stream URLs.
		res[i].Track.Location = redact.String(res[i].Track.Location)
		if res[i].Method == match.Unmatched {
			unmatched++
			fmt.Fprintf(stderr, "unmatched #%d: %s - %s (%s)\n", res[i].Index+1, res[i].Track.Artist, res[i].Track.Title, res[i].Track.Location)
		}
	}
	fmt.Fprintf(stderr, "%d/%d tracks matched\n", len(res)-unmatched, len(res))
	return writeJSON(stdout, map[string]any{"name": p.Name, "results": res})
}
