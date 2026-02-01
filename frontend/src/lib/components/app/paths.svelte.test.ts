import { cmd } from "$wailsjs/go/models";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { paths } from "./paths.svelte";
import { mockWailsAppFn } from "./test-utils";

describe("ストア: paths", () => {
    let mockSetPaths = vi.fn();
    beforeEach(() => {
        mockSetPaths = mockWailsAppFn("SetPaths");
    });
    afterEach(() => {
        vi.clearAllMocks();
    });

    test("値を更新・取得できること", async () => {
        await paths.update({ [cmd.PathKey.ORDER_XLSX]: "/mock/order_path" });

        let value = paths.value;

        expect(value).toEqual({ [cmd.PathKey.ORDER_XLSX]: "/mock/order_path" });
    });

    test("更新時にSetPathsが呼ばれること", async () => {
        expect(mockSetPaths).not.toHaveBeenCalled();

        await paths.update({ [cmd.PathKey.NAMES_CSV]: "/mock/names_path" });

        expect(mockSetPaths).toHaveBeenCalledExactlyOnceWith(paths.value);
    })
});