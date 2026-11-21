# Architecture

gobbler is one portable engine with small adapters around it, plus thin
satellites that adapt it to specific apps. Everything that can be written once
lives in the Go core; satellites stay as small as their host platform allows.

## Component diagram

```mermaid
flowchart TB
    subgraph Clients["Editing clients"]
        AIMPW["AIMP 5.x on Windows"]
        AIMPA["AIMP for Android"]
    end

    subgraph Agents["playlist-gobbler-aimp agents (MIT, Go)"]
        AGW["Windows / Linux agent<br/>cmd/playlist-gobbler-aimp"]
        AGA["Android agent<br/>mobile/ gomobile AAR"]
    end

    subgraph NC["Nextcloud"]
        NCF[("Per-user Playlists/ folders<br/>WebDAV")]
        NCAPP["playlist-gobbler.nextcloud app<br/>(PHP, MIT, optional)"]
    end

    subgraph Core["gobbler core (Go, MIT)"]
        CLI["gobbler CLI<br/>read · write · convert · match · sync"]
        subgraph Adapters["Adapter interfaces"]
            STORE["Store<br/>file · webdav/nextcloud · smb"]
            FORMAT["Format<br/>aimppl4 · m3u/m3u8 · pls · xspf"]
            MATCH["Matcher<br/>server id → path map → metadata"]
            TARGET["Target<br/>Jellyfin · Subsonic"]
        end
        REDACT["redact<br/>strips tokens from logs"]
    end

    subgraph JF["Jellyfin 12.4"]
        JFP["playlist-gobbler.jellyfin plugin<br/>(C#, GPL-3.0)"]
        JFPL[("Jellyfin playlists")]
        JFLIB[("Music library")]
    end

    AIMPW -->|".aimppl4 in PLS folder"| AGW
    AIMPA -->|"playlist folder"| AGA
    AGW -->|"Nextcloud desktop sync or WebDAV"| NCF
    AGA -->|"Nextcloud Android app or WebDAV"| NCF
    NCAPP -.->|"settings, change hints"| NCF
    JFP -->|"scheduled task runs"| CLI
    CLI --> STORE
    CLI --> FORMAT
    CLI --> MATCH
    CLI --> TARGET
    CLI --> REDACT
    STORE -->|"WebDAV"| NCF
    TARGET -->|"REST API"| JFPL
    TARGET -->|"library snapshot"| JFLIB
```

## Packages

```mermaid
flowchart LR
    cmd_gobbler["cmd/gobbler"] --> format["pkg/format"]
    cmd_gobbler --> match["pkg/match"]
    cmd_gobbler --> redact["pkg/redact"]
    cmd_aimp["cmd/playlist-gobbler-aimp"] --> agent["internal/agent"]
    mobile["mobile (gomobile)"] --> agent
    mobile --> format
    agent --> format
    agent --> store["pkg/store"]
    format --> playlist["pkg/playlist"]
    match --> playlist
    match --> target["pkg/target"]
```

| Package | Responsibility |
|---|---|
| `pkg/playlist` | Format-neutral `Playlist` / `Track` model |
| `pkg/format` | Format adapters + registry (by name or extension), text decoding |
| `pkg/store` | Store interface + registry by URI scheme; local folder store |
| `pkg/target` | Target interface (media server) |
| `pkg/match` | Matching strategies and path-map table |
| `pkg/redact` | Credential scrubbing for anything logged |
| `internal/agent` | Polling folder mirror used by the AIMP agent and Android binding |
| `mobile` | gomobile-friendly facade (`Convert`, `SyncFolder`) |
| `cmd/gobbler` | CLI |
| `cmd/playlist-gobbler-aimp` | AIMP folder agent |
| `playlist-gobbler.nextcloud/` | Nextcloud app skeleton |

## Repositories and licenses

| Repo | Contents | License |
|---|---|---|
| `sys7mfaul7/gobbler` | core, CLI, AIMP agent, Android binding, Nextcloud app skeleton, docs | MIT |
| `sys7mfaul7/playlist-gobbler.jellyfin` | Jellyfin plugin | GPL-3.0, because it links Jellyfin's GPL libraries (same as the official plugin template) |

The Jellyfin plugin calls `gobbler` as a separate program, so the MIT core
never links GPL code.

## Platforms

| Tier | Platforms | CI |
|---|---|---|
| First-class | windows/amd64, windows/arm64, linux/amd64, linux/arm64 | test + build |
| Mobile | android/arm64 (static CLI + gomobile AAR) | build |
| Best-effort | darwin/amd64, darwin/arm64, freebsd/amd64 | build, non-blocking |
