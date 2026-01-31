
import { cmd } from "$wailsjs/go/models";
import { tick, type ComponentProps } from "svelte";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { render } from "vitest-browser-svelte";
import { page } from "vitest/browser";
import { paths } from "./paths.svelte";
import RosterCard from "./RosterCard.svelte";
import { mockWailsAppFn } from "./test-utils";

const mockSetCardCompleted = vi.fn();

const defaultProps = {
    title: "Test title",
    setCardCompleted: mockSetCardCompleted
} satisfies ComponentProps<typeof RosterCard>

describe("RosterCard component", () => {
    let mockLoadRosterRecords = vi.fn();
    beforeEach(() => {
        mockLoadRosterRecords = mockWailsAppFn("LoadRosterRecords").mockResolvedValue({ 山田: "山田 太郎" });
        mockWailsAppFn("SetPaths");
        mockWailsAppFn("GetRosterColumnNames").mockResolvedValue([{ field: "fullName", label: "姓名 (テスト)" }]);
    });
    afterEach(async () => {
        await paths.update({ [cmd.PathKey.NAMES_CSV]: "" });
        vi.clearAllMocks();
    })

    describe("Props: タイトル", () => {
        test("与えた文字列が表示されること", async () => {
            render(RosterCard, defaultProps);

            const title = page.getByText(defaultProps.title, { exact: true });
            await expect.element(title).toBeVisible();
        }
        );
    });

    describe("Props: 完了処理", () => {
        beforeEach(async () => {
            await paths.update({ [cmd.PathKey.NAMES_CSV]: "/mock/path" });
        })
        test("データロードが完了した場合に呼び出されること", async () => {
            render(RosterCard, defaultProps);

            await tick();
            expect(mockLoadRosterRecords).toHaveBeenCalled();
            expect(mockSetCardCompleted).toHaveBeenCalled();
        });
    });

    describe("function: フィールドのレンダリング", () => {
        test("ロードされたデータの値がそのまま表示されること", async () => {
            mockWailsAppFn("SetPaths");
            await paths.update({ [cmd.PathKey.NAMES_CSV]: "/mock/path" });
            render(RosterCard, defaultProps);

            const recordValue = page.getByText("山田 太郎");
            await expect.element(recordValue).toBeVisible();
        })
    })
})