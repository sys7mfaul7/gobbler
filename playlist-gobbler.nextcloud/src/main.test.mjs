import { test } from 'node:test'
import assert from 'node:assert/strict'
import { appId } from './main.js'

test('app id matches appinfo', () => {
  assert.equal(appId, 'playlistgobbler')
})
