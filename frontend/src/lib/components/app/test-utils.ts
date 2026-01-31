import { vi } from 'vitest';

// 再帰ネスト作成（型安全）
function ensureNestedPath(obj: any, path: string[]): void {
    for (const key of path.slice(0, -1)) {
        obj[key] ??= {};
        obj = obj[key];
    }
}

// 汎用モック（Proxy と連携）
function mockWailsPath(path: string[], impl = vi.fn()) {
    const [pkg, app, ...rest] = path;
    if (!pkg || !app) throw new Error('path must be ["pkg", "app", "fn"]');

    // vitest-setup.ts の Proxy 環境前提
    ensureNestedPath(window.go, [pkg, app]);
    (window.go as any)[pkg][app][path[2]!] = impl;

    return impl;
}

export const mockWailsAppFn = (fn: string, impl = vi.fn()) =>
    mockWailsPath(['main', 'WailsApp', fn], impl);
