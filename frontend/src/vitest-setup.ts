import { beforeEach, vi } from 'vitest';

declare global {
    interface Window {
        go: {};
    }
};

beforeEach(() => {
    window.go = {
        main: {
            WailsApp: new Proxy(
                {} as Record<string, any>,
                {
                    get(target: Record<string, any>, fnName: string | symbol): any {
                        if (!target[fnName as string]) {
                            target[fnName as string] = vi.fn(() => {
                                throw new Error(
                                    `[🚨 未モック]\n` +
                                    `Call mockWailsAppFn('${String(fnName)}')`
                                );
                            });
                        }
                        return target[fnName as string];
                    },
                    set(target: Record<string, any>, fnName: string | symbol, value: any): boolean {
                        target[fnName as string] = value;
                        return true;
                    }
                }
            )
        }
    };
});
