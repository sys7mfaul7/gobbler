<?php

declare(strict_types=1);

namespace OCA\PlaylistGobbler\Tests\Unit;

use OCA\PlaylistGobbler\Controller\StatusController;
use PHPUnit\Framework\TestCase;

class StatusControllerTest extends TestCase {
    public function testStatus(): void {
        $s = StatusController::status();
        $this->assertSame('playlistgobbler', $s['app']);
        $this->assertSame('scaffold', $s['status']);
    }
}
