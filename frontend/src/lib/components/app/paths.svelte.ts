import { SetPaths } from "$wailsjs/go/main/WailsApp";
import type { cmd } from "$wailsjs/go/models";

type Paths = Record<cmd.PathKey, string>;

function createPaths() {
    let paths = $state<Partial<Paths>>({});

    return {
        get value() {
            return paths;
        },
        async update(updates: Partial<Paths>): Promise<void> {
            paths = { ...paths, ...updates };
            await SetPaths(paths as Paths);
        }
    }
}

export const paths = createPaths();
