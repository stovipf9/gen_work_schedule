/// <reference types="vitest/config" />
import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import path from "path";
import { preview } from '@vitest/browser-preview'

export default defineConfig(({ mode }) => {
    // Shared config
    const sharedConfig = {
        plugins: [svelte()],
        resolve: {
            alias: {
                $lib: path.resolve(__dirname, 'src/lib'),
                $wailsjs: path.resolve(__dirname, 'wailsjs'),
                $assets: path.resolve(__dirname, 'src/assets'),
            }
        }
    }

    // Test-specific config
    if (mode === 'test') {
        return {
            ...sharedConfig,
            test: {
                browser: {
                    enabled: true,
                    provider: preview(),
                    // provider: playwright(),
                    api: {
                        host: '0.0.0.0',
                        // port: 63315,
                    },
                    instances: [{ browser: 'chromium' }]
                },
                coverage: {
                    enabled: true,
                    provider: 'istanbul',
                    reporter: ['text', 'html', 'lcov'],
                    include: ['src/lib/components/app/**'],
                },
                setupFiles: ['src/vitest-setup.ts']
            },
        }
    }

    return sharedConfig
})