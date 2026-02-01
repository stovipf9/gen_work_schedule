// execute pnpm knip at frontend directory.
import type { KnipConfig } from "knip";

export default {
    ignoreDependencies: ["tailwindcss", "autoprefixer"],
    postcss: {
        entry: ["postcss.config.js"]
    },
    vitest: {
        project: "vite.config.ts"
    }
} satisfies KnipConfig;