# Contributing

Thanks for helping! gobbler aims to stay small, portable and boring.

## Ground rules

* **Unix philosophy.** One command or adapter does one job. New features are
  usually a new adapter (Store, Format, Target) or a new subcommand, not a
  flag on an existing one.
* **Portable.** No cgo in the core. Code must build for every target in
  `make cross` (Windows/Linux amd64+arm64, android/arm64).
* **No secrets in logs or the repo.** Anything printed goes through
  `pkg/redact`. Tests use synthetic fixtures only. Never commit real
  playlists containing stream URLs, API keys or app passwords.
* **License hygiene.** The core is MIT. Do not copy code from GPL/AGPL
  projects; reading them for format knowledge is fine, but write your own
  implementation and say where the knowledge came from.

## Development

```sh
scripts/setup-go.sh          # or install Go >= the version in go.mod
make vet test                # gofmt check, go vet, unit tests
make cross                   # cross-compile first-class platforms
make nextcloud-test          # Nextcloud app (needs scripts/setup-nextcloud.sh)
make mobile-aar              # Android AAR (needs scripts/setup-android.sh)
```

The Jellyfin plugin lives in
[sys7mfaul7/playlist-gobbler.jellyfin](https://github.com/sys7mfaul7/playlist-gobbler.jellyfin)
(GPL-3.0); `make jellyfin` builds it when checked out next to this repo.

## Adding a format

1. Add `pkg/format/<name>.go` implementing `Reader` and/or `Writer`.
2. Register it in `init()` with its file extensions.
3. Add a synthetic fixture to `pkg/format/testdata/` and tests.
4. Update the format table in `docs/DESIGN.md`.

## Pull requests

* Keep PRs focused; include tests.
* `make vet test` must pass. CI runs tests on Linux and Windows and builds
  every platform.
* Describe any change to on-disk formats or sync semantics in `docs/`.
