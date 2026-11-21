// Command playlist-gobbler-aimp watches AIMP's playlist folder and mirrors
// each playlist, converted to m3u8 (or another format), into an output
// folder - typically one synced by the Nextcloud desktop or Android client.
//
//	playlist-gobbler-aimp --src "%APPDATA%\AIMP\PLS" --dst "C:\Users\me\Nextcloud\Playlists"
//
// Runs on Windows, Linux and Android (AIMP for Android + Nextcloud app).
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/sys7mfaul7/gobbler/internal/agent"
	"github.com/sys7mfaul7/gobbler/pkg/redact"
	"github.com/sys7mfaul7/gobbler/pkg/store"
)

var version = "dev"

func main() {
	src := flag.String("src", os.Getenv("GOBBLER_AIMP_PLS"), "AIMP playlist folder (env GOBBLER_AIMP_PLS)")
	dst := flag.String("dst", os.Getenv("GOBBLER_DST"), "output store, e.g. a Nextcloud-synced folder (env GOBBLER_DST)")
	outFmt := flag.String("format", "m3u8", "output playlist format")
	interval := flag.Duration("interval", 10*time.Second, "poll interval")
	once := flag.Bool("once", false, "sync once and exit")
	showVersion := flag.Bool("version", false, "print version")
	flag.Parse()
	if *showVersion {
		fmt.Println("playlist-gobbler-aimp", version)
		return
	}
	if *src == "" || *dst == "" {
		flag.Usage()
		os.Exit(2)
	}
	s, err := store.Open(*src)
	if err != nil {
		log.Fatal(err)
	}
	d, err := store.Open(*dst)
	if err != nil {
		log.Fatal(err)
	}
	a := &agent.Agent{Src: s, Dst: d, OutFormat: *outFmt, Log: func(f string, args ...any) {
		log.Print(redact.String(fmt.Sprintf(f, args...)))
	}}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	if *once {
		names, err := a.SyncOnce(ctx)
		for _, n := range names {
			log.Printf("wrote %s", n)
		}
		if err != nil {
			log.Fatal(redact.String(err.Error()))
		}
		return
	}
	if err := a.Run(ctx, *interval); err != nil && ctx.Err() == nil {
		log.Fatal(err)
	}
}
