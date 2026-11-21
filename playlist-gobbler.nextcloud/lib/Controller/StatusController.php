<?php

declare(strict_types=1);

// SPDX-License-Identifier: MIT

namespace OCA\PlaylistGobbler\Controller;

use OCA\PlaylistGobbler\AppInfo\Application;
use OCP\AppFramework\Http\Attribute\NoAdminRequired;
use OCP\AppFramework\Http\JSONResponse;
use OCP\AppFramework\OCSController;
use OCP\IRequest;

/**
 * Minimal health endpoint so clients (AIMP agent, Jellyfin plugin) can check
 * the app is installed. Returns no user data.
 */
class StatusController extends OCSController {
    public function __construct(IRequest $request) {
        parent::__construct(Application::APP_ID, $request);
    }

    #[NoAdminRequired]
    public function index(): JSONResponse {
        return new JSONResponse(self::status());
    }

    /** @return array{app: string, version: string, status: string} */
    public static function status(): array {
        return ['app' => Application::APP_ID, 'version' => '0.1.0', 'status' => 'scaffold'];
    }
}
