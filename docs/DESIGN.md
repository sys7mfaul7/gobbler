# gobbler design

## Goals

1. **AIMP stays the editor.** People build playlists in AIMP (Windows, and
   AIMP for Android). gobbler never asks them to change how they work.
2. **One source of truth per user**, stored like bookmarks, notes and
   passwords already are: a Nextcloud folder (`Playlists/`).
3. **Jellyfin mirrors it**, with real, editable Jellyfin playlists made of
   library items rather than read-only folder imports.
4. **Portable core.** One Go codebase, single static binaries, no cgo:
   Windows and Linux (amd64 + arm64) first-class, Android next, macOS and
   FreeBSD best-effort.
5. **Unix philosophy.** Small commands that do one thing, speak JSON on
   stdin/stdout, and compose in pipelines. Adapters are chosen with flags or
   environment variables, not recompiles.

## Non-goals (for now)

* Two-way sync (Jellyfin → AIMP). Designed for, built later.
* Editing playlists inside Jellyfin or Nextcloud UIs.
* Transcoding, tagging, or library management.

## Core: `gobbler`

`gobbler` is both a Go library (`pkg/...`) and a CLI (`cmd/gobbler`).

| Command | In | Out | Purpose |
|---|---|---|---|
| `gobbler read [--format F] [FILE]` | playlist file / stdin | playlist JSON | decode any supported format |
| `gobbler write --format F` | playlist JSON | playlist file | encode |
| `gobbler convert [--from F] --to F [FILE]` | playlist file | playlist file | read + write |
| `gobbler match --library L [--map FROM=TO]...` | playlist JSON | match report JSON (+ unmatched on stderr) | resolve tracks to server items |
| `gobbler sync --from STORE --to TARGET [--format F]` | – | – | full pipeline (planned) |
| `gobbler formats`, `gobbler version` | – | text | introspection |

Example of the planned all-in-one form:

```sh
gobbler sync --from nextcloud:/Playlists --to jellyfin://cantina.lan --format aimppl4
```

…which is exactly `list store | read | match | upsert target` and can be
spelled out as a pipeline for debugging:

```sh
gobbler read --format aimppl4 Road.aimppl4 | gobbler match --library lib.json > report.json
```

### Configuration and environment

Precedence: flags > environment > config file > defaults.

| Variable | Meaning |
|---|---|
| `GOBBLER_FORMAT` | default input format |
| `GOBBLER_MAP` | `;`-separated `FROM=TO` path rules |
| `GOBBLER_CONFIG` | config file path (planned; TOML/JSON, see DATA-MODEL.md) |
| `GOBBLER_JELLYFIN_URL`, `GOBBLER_JELLYFIN_API_KEY` | Jellyfin target (planned) |
| `GOBBLER_WEBDAV_URL`, `GOBBLER_WEBDAV_USER`, `GOBBLER_WEBDAV_PASSWORD` | Nextcloud store (planned; use an app password) |
| `GOBBLER_AIMP_PLS`, `GOBBLER_DST` | AIMP agent source and destination |

## Adapter interfaces

All adapters are plain Go interfaces in `pkg/` with a small registry so new
ones are a file plus an `init()`.

### Store — where playlist files live (`pkg/store`)

```go
type Store interface {
    List(ctx) ([]Entry, error)                  // name, mtime, size
    Open(ctx, name) (io.ReadCloser, error)
    Create(ctx, name) (io.WriteCloser, error)   // atomic replace
}
```

Selected by URI scheme: `file:` / bare path (done), `nextcloud:` and
`webdav:` (next), `smb:` (later). Local writes are atomic (temp file +
rename).

### Format — how playlists are encoded (`pkg/format`)

```go
type Reader interface { Read(io.Reader) (*playlist.Playlist, error) }
type Writer interface { Write(io.Writer, *playlist.Playlist) error }
```

| Format | Read | Write | Notes |
|---|---|---|---|
| `aimppl4` (AIMP 4.x / 5.x) | ✅ | planned | UTF-16LE, sections, pipe-delimited tracks |
| `m3u8` / `m3u` | ✅ | ✅ | `#EXTM3U`, `#EXTINF`, `#PLAYLIST` |
| `aimppl` (AIMP 3.x and older) | planned | – | older section layout |
| `pls` | planned | planned | |
| `xspf` | planned | planned | |

Text decoding honours UTF-8, UTF-16LE and UTF-16BE BOMs and sniffs BOM-less
UTF-16LE.

#### aimppl4 layout (as understood; verify against real AIMP 5.40 files)

```
#-----SUMMARY-----#
ID={GUID}
Name=Road Trip
...
#-----SETTINGS-----#
Flags=...
FormatMainLine=...
#-----CONTENT-----#
-M:\Music\Artist\Album                  <- group line: '-' + folder
M:\Music\Artist\Album\01.flac|Title|Artist|Album|AlbumArtist|Genre|Year|TrackNo|DiskNo|Composer|Publisher|Bitrate|Channels|SampleRate|DurationMs|Size|BPM|IsActive|Index|PluginReserved|...
```

The reader is lenient: missing trailing fields are empty, unknown sections
are skipped, bad numbers become 0. Format knowledge comes from reading
[ApexWeed/aimppl-copy](https://github.com/ApexWeed/aimppl-copy) (MIT) and
[vkostyanetsky/AIMPPlaylistsConvertor](https://github.com/vkostyanetsky/AIMPPlaylistsConvertor)
(Unlicense); no code was copied. No AGPL/GPL code is used in this repo.

### Target — the media server (`pkg/target`)

```go
type Target interface {
    Library(ctx, user) ([]Item, error)
    Playlists(ctx, user) ([]RemotePlaylist, error)
    Upsert(ctx, user, name, itemIDs []string) (RemotePlaylist, error)
}
```

Jellyfin first (REST API, `Authorization: MediaBrowser Token="…"`), then
Subsonic/OpenSubsonic (which also covers Navidrome and Jellyfin via Subfin).

### Matcher — tracks to server items (`pkg/match`)

Strategies in order, first hit wins:

1. **Server ID** — `Track.ServerID`, or the `id=` of a Subsonic stream URL
   (AIMP's Subsonic plugin stores these).
2. **Path map** — rewrite the client path through a prefix table
   (`M:\Music` → `/data/music`, longest prefix wins, case-insensitive,
   `\` = `/`), then exact match against server item paths.
3. **Metadata** — normalised artist + title, album when both sides have it,
   duration within **±2 s** (configurable); closest duration wins.
4. **Unmatched** — reported on stderr and in the JSON report, never dropped
   silently.

## Users and folders

Each user's store folder maps to one Jellyfin user:

```
nextcloud:/shawn/Playlists   -> jellyfin user "shawn"
nextcloud:/scott/Playlists   -> jellyfin user "scott"
nextcloud:/Shared/Playlists  -> public playlists (owner configurable)
```

## Sync semantics

* **v1: one-way, store → Jellyfin.** The store is authoritative. gobbler
  owns only playlists it created (tracked by name + a marker in the
  playlist overview) and replaces their contents on change. Playlists it did
  not create are never touched.
* **Dry run by default** in the Jellyfin plugin; reports what would change.
* **Later: two-way.** Keep the last-synced snapshot per playlist; three-way
  merge item lists; conflicts are reported, not guessed.

## Satellites

| Name | Role | Home |
|---|---|---|
| `playlist-gobbler.jellyfin` | Thin C# plugin: settings page + scheduled task that runs/embeds `gobbler sync`. Must be C#; kept minimal. | separate repo, GPL-3.0 |
| `playlist-gobbler.nextcloud` | v1: none needed (WebDAV store). Later: Nextcloud app for per-user settings, browsing and change notifications. | `playlist-gobbler.nextcloud/` here; may split out |
| `playlist-gobbler-aimp` | Agent that watches AIMP's `PLS` folder and mirrors playlists into the store. Polling, cgo-free, runs on Windows, Linux, Android. Windows tray UI later; native AIMP plugin possibly later still. | `cmd/playlist-gobbler-aimp` here |
| Android agent | AIMP for Android + Nextcloud Android app: the agent syncs AIMP's Android playlist folder into the Nextcloud-synced folder. | `mobile/` (gomobile) here; app later |

## Security

* AIMP Subsonic stream URLs carry `u=`, `t=`, `s=` (and sometimes `p=`).
  **Never log them.** All diagnostic output goes through `pkg/redact`, which
  strips secret query parameters and URL userinfo.
* Secrets come from env or a config file with `0600` permissions, never from
  the repository, and preferably not from command-line flags.
* Jellyfin API key and Nextcloud **app password** (not the login password).
* The local store refuses paths that escape its root.

## Roadmap

| Milestone | Contents |
|---|---|
| **0.1** (now) | Go module, `aimppl4` reader, `m3u/m3u8` reader+writer, matcher, local store, CLI `read/write/convert/match`, AIMP folder agent, Jellyfin plugin scaffold, Nextcloud app scaffold, CI |
| 0.2 | WebDAV/Nextcloud store, Jellyfin target, `gobbler sync` one-way, config file, dry-run report |
| 0.3 | Jellyfin plugin runs sync on schedule; per-user mapping UI |
| 0.4 | AIMP 3.x/older formats, `pls`, `xspf`, `aimppl4` writer |
| 0.5 | Windows tray agent (Windows + Linux first-class, amd64 + arm64) |
| 0.6 | **Android**: agent for AIMP for Android + Nextcloud Android app, via a `GOOS=android GOARCH=arm64` static build (works today) or a `gomobile bind` AAR for a proper app |
| 1.0 | Two-way sync with conflict reporting; Subsonic target |

### Platform priority

1. Windows amd64/arm64, Linux amd64/arm64 — first-class, tested in CI.
2. Android arm64 — built in CI (static CLI + AAR); app later.
3. macOS, FreeBSD — best-effort CI builds, not release-blocking.
