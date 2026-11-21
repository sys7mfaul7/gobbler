<?php

declare(strict_types=1);

// SPDX-License-Identifier: MIT

namespace OCA\PlaylistGobbler\AppInfo;

use OCP\AppFramework\App;
use OCP\AppFramework\Bootstrap\IBootContext;
use OCP\AppFramework\Bootstrap\IBootstrap;
use OCP\AppFramework\Bootstrap\IRegistrationContext;

class Application extends App implements IBootstrap {
    public const APP_ID = 'playlistgobbler';

    public function __construct(array $urlParams = []) {
        parent::__construct(self::APP_ID, $urlParams);
    }

    public function register(IRegistrationContext $context): void {
        // Future: file-change listeners on each user's Playlists folder.
    }

    public function boot(IBootContext $context): void {
    }
}
