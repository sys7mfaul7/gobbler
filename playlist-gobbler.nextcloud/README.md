# playlist-gobbler.nextcloud

Nextcloud app (`playlistgobbler`) for the [gobbler](../README.md) project.

**Status: scaffold.** In v0.1 the Nextcloud side needs no app at all:
gobbler talks to each user's `Playlists/` folder over WebDAV. This app is the
future home of per-user settings, a playlist browser, and change
notifications to the Jellyfin plugin.

Today it ships one endpoint, `GET /ocs/v2.php/apps/playlistgobbler/api/status`.

## Build & test

```sh
composer install
composer lint && composer test
npm run build && npm test
```

Toolchain: PHP ≥ 8.1 (tested 8.4), Composer 2, Node ≥ 20. See
`../scripts/setup-nextcloud.sh`.

License: MIT.
