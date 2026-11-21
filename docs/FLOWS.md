# Flows

## 1. Edit in AIMP → Jellyfin playlist (target v0.2/v0.3 flow)

```mermaid
sequenceDiagram
    autonumber
    actor User
    participant AIMP as AIMP (Windows)
    participant Agent as playlist-gobbler-aimp
    participant NC as Nextcloud Playlists/
    participant Plugin as Jellyfin plugin
    participant Core as gobbler sync
    participant JF as Jellyfin API

    User->>AIMP: add, remove or reorder tracks
    AIMP->>AIMP: save PLS/Road Trip.aimppl4
    loop every poll interval
        Agent->>AIMP: list PLS folder, compare mtimes
    end
    Agent->>Agent: read aimppl4, convert
    Agent->>NC: write Road Trip.m3u8 atomically
    Note over NC: Nextcloud desktop client or WebDAV uploads the file
    Plugin->>Core: scheduled task starts sync
    Core->>NC: WebDAV list and fetch changed playlists
    Core->>JF: fetch library snapshot for the mapped user
    Core->>Core: match tracks
    Core->>JF: create or replace playlist items
    Core-->>Plugin: report with matched and unmatched counts
    Plugin-->>User: task log and dashboard status
```

## 2. Track matching pipeline

```mermaid
flowchart TD
    T["Playlist track"] --> A{"Has server id?<br/>ServerID or Subsonic id= param"}
    A -->|"yes, exists in library"| M1["Match: server-id"]
    A -->|"no"| B["Apply path map<br/>longest prefix, case-insensitive, \\ = /"]
    B --> C{"Mapped path equals<br/>a library item path?"}
    C -->|"yes"| M2["Match: path"]
    C -->|"no"| D{"Artist + title match<br/>(normalised), album agrees,<br/>duration within ±2s?"}
    D -->|"one or more"| M3["Match: metadata<br/>closest duration wins"]
    D -->|"none"| U["Unmatched<br/>reported, never dropped"]
```

## 3. First-run import

```mermaid
sequenceDiagram
    autonumber
    actor Admin
    participant JF as Jellyfin
    participant Core as gobbler
    participant NC as Nextcloud

    Admin->>JF: move music shares one level down and add .ignore
    Note over JF: stops album .m3u files becoming playlists. Track ids change, so do this before the first sync
    Admin->>JF: rescan library
    Admin->>Core: configure path maps and folder to user mapping
    Admin->>Core: gobbler sync --dry-run
    Core->>NC: list all playlists for each mapped user
    Core->>JF: fetch library snapshots
    Core-->>Admin: dry-run report, would create N playlists, M unmatched tracks
    Admin->>Admin: fix path maps until unmatched is acceptable
    Admin->>Core: gobbler sync
    Core->>JF: create playlists, tagged as owned by gobbler
    Core-->>Admin: final report
```

## 4. Conflicts and unmatched tracks

```mermaid
flowchart TD
    S["Sync run"] --> P{"Playlist with same name<br/>exists in Jellyfin?"}
    P -->|"no"| C["Create it, mark as gobbler-owned"]
    P -->|"yes, gobbler-owned"| R["Replace item list<br/>from the store"]
    P -->|"yes, not gobbler-owned"| K["Skip and report conflict<br/>never overwrite user playlists"]
    C --> U{"Unmatched tracks?"}
    R --> U
    U -->|"yes"| W["Keep matched order,<br/>list unmatched in report<br/>with redacted locations"]
    U -->|"no"| OK["Done"]
    W --> OK
```

In two-way mode (planned), gobbler keeps the last-synced snapshot of each
playlist and does a three-way merge of item lists; changes on both sides to the
same position are reported as conflicts instead of guessed.

## 5. Android: AIMP for Android + Nextcloud

```mermaid
sequenceDiagram
    autonumber
    actor User
    participant AIMPA as AIMP for Android
    participant Agent as Android agent (gomobile)
    participant NCA as Nextcloud Android app
    participant NC as Nextcloud server
    participant Core as gobbler on server
    participant JF as Jellyfin

    User->>AIMPA: edit playlist on phone
    AIMPA->>AIMPA: save playlist file in its playlist folder
    Agent->>AIMPA: poll playlist folder
    Agent->>Agent: Convert or SyncFolder into the synced folder
    NCA->>NC: upload changed file (auto-upload or sync)
    Core->>NC: next scheduled sync picks it up
    Core->>JF: update the user's playlist
```

Assumption to verify: AIMP for Android exposes its playlists as files in a
folder the agent can read. If it does not, the Android agent falls back to
importing exported `.m3u8` files.
