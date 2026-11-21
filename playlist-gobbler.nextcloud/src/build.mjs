// Placeholder build: copies src/main.js to js/ so the app layout is complete.
// Replace with @nextcloud/vite-config once there is a real UI.
import { mkdirSync, copyFileSync } from 'node:fs'
mkdirSync(new URL('../js/', import.meta.url), { recursive: true })
copyFileSync(new URL('./main.js', import.meta.url), new URL('../js/playlistgobbler-main.js', import.meta.url))
console.log('built js/playlistgobbler-main.js')
