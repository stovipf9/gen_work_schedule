import { cmd } from "$wailsjs/go/models";
import { LogError } from "$wailsjs/runtime/runtime";
import { tick } from "svelte";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { render } from "vitest-browser-svelte";
import { page } from "vitest/browser";
import App from "./App.svelte";
import { paths } from "./paths.svelte";
import { mockWailsAppFn } from "./test-utils";

describe("App component", () => {
    let mockLoadCachedPaths = vi.fn();
    beforeEach(() => {
        mockWailsAppFn("SetPaths");
        mockWailsAppFn("GetOrderColumnNames").mockResolvedValue([]);
        mockWailsAppFn("LoadOrders").mockRejectedValue("error in loading order");
        mockWailsAppFn("GetRosterColumnNames").mockResolvedValue([]);
        mockWailsAppFn("SelectDirectory").mockResolvedValue("/mock/dir");
        mockLoadCachedPaths = mockWailsAppFn("LoadCachedPaths")
            .mockResolvedValue({ [cmd.PathKey.ORDER_XLSX]: "/mock/path" });
        vi.mock('$wailsjs/runtime/runtime', () => ({
            LogError: vi.fn(),
            Quit: vi.fn(),
        }));
    });
    afterEach(() => {
        vi.clearAllMocks();
    })

    describe("マウント時", () => {
        test("キャッシュ済みパス読み込み処理が呼び出されること", async () => {
            render(App);

            expect(mockLoadCachedPaths).toHaveBeenCalledOnce();
        });
        test("キャッシュ済みパス読み込み処理が成功した際に\
            パスがアップデートされること", async () => {
            await paths.update({ [cmd.PathKey.ORDER_XLSX]: "/mock/old_path" });
            render(App);

            await tick();
            expect(mockLoadCachedPaths).toHaveBeenCalledOnce();
            expect(paths.value[cmd.PathKey.ORDER_XLSX]).toEqual("/mock/path");
        });
        test("キャッシュ済みパス読み込み処理が失敗した際に\
            エラーロギング処理が呼び出されること", async () => {
            const rejectedValue = "error in mock";
            const mockLoadCachedPathsReject = mockWailsAppFn("LoadCachedPaths")
                .mockRejectedValue(rejectedValue);
            render(App);

            await tick();
            expect(mockLoadCachedPathsReject).toHaveBeenCalledOnce();
            expect(LogError).toHaveBeenCalledExactlyOnceWith(
                "Failed to load cached paths: " + rejectedValue);
        })
    });

    // describe("ページ群", () => {

    // });

    describe("選択ページ", () => {
        test("初期選択ページは発注情報であること", async () => {
            render(App);

            await expect.element(page.getByTitle("発注情報")).toBeVisible();
        });
        test("ページネーション前後に合わせて変化すること", async () => {
            render(App);

            await page.getByRole("button", { name: "Next" }).click();

            await expect.element(page.getByTitle("発注情報")).not.toBeInTheDocument();
            await expect.element(page.getByTitle("名簿")).toBeVisible();

            await page.getByRole("button", { name: "Previous" }).click();

            await expect.element(page.getByTitle("発注情報")).toBeVisible();
            await expect.element(page.getByTitle("名簿")).not.toBeInTheDocument();
        });
        test("サイドバーでの選択に合わせて変化すること", async () => {
            render(App);

            await page.getByRole("button", { name: "勤務表テンプレート" }).click();

            await expect.element(page.getByTitle("勤務表テンプレート")).toBeVisible();
        });
    });

    describe("ページ操作完了状況", () => {
        test("サイドバーと連携すること", async () => {
            mockWailsAppFn("SelectFile").mockResolvedValue("/mock/path");
            render(App);

            await page.getByRole("button", { name: "勤務表テンプレート" }).click();
            const selectButton = page.getByRole("button", { name: "ファイル/フォルダ選択" });
            await expect.element(selectButton).toHaveLength(1);
            await selectButton.click();

            const checkbox = page.getByTestId("success-checkbox-勤務表テンプレート");
            await expect.element(checkbox).toBeVisible();
            await expect.element(checkbox).toBeChecked();

            for (const title of ["発注情報", "名簿", "出力"]) {
                const checkbox = page.getByTestId(`success-checkbox-${title}`);
                await expect.element(checkbox).not.toBeInTheDocument();
            }
        });
    });

    describe("入力ページ", () => {
        test("未完了の入力ページが存在する場合、出力ボタンが非活性であること", async () => {
            mockWailsAppFn("LoadCachedPaths").mockResolvedValue(
                {
                    [cmd.PathKey.ORDER_XLSX]: "/mock/order_path",
                    [cmd.PathKey.NAMES_CSV]: "/mock/names_path",
                    [cmd.PathKey.TEMPLATE_XLSX]: "/mock/template_path"
                },
            );

            mockWailsAppFn("LoadOrders").mockResolvedValue([]);
            mockWailsAppFn("LoadRosterRecords").mockResolvedValue([]);

            render(App);

            await page.getByRole("button", { name: "発注情報" }).click();
            await page.getByRole("button", { name: "名簿" }).click();
            // await page.getByRole("button", { name: "勤務表テンプレート" }).click();

            await page.getByRole("button", { name: "出力" }).click();
            const selectButton = page.getByRole("button", { name: "ファイル/フォルダ選択" });
            await expect.element(selectButton).toHaveLength(1);
            await selectButton.click();

            await expect.element(page.getByLabelText("勤務表出力")).toBeDisabled();
        });
        test("入力ページが全て完了かつ出力フォルダが非空の場合、\
            出力ボタンが活性であること", async () => {
            mockWailsAppFn("LoadCachedPaths").mockResolvedValue(
                {
                    [cmd.PathKey.ORDER_XLSX]: "/mock/order_path",
                    [cmd.PathKey.NAMES_CSV]: "/mock/names_path",
                    [cmd.PathKey.TEMPLATE_XLSX]: "/mock/template_path"
                },
            );

            mockWailsAppFn("LoadOrders").mockResolvedValue([]);
            mockWailsAppFn("LoadRosterRecords").mockResolvedValue([]);

            render(App);

            await page.getByRole("button", { name: "発注情報" }).click();
            await page.getByRole("button", { name: "名簿" }).click();
            await page.getByRole("button", { name: "勤務表テンプレート" }).click();

            await page.getByRole("button", { name: "出力" }).click();
            const selectButton = page.getByRole("button", { name: "ファイル/フォルダ選択" });
            await expect.element(selectButton).toHaveLength(1);
            await selectButton.click();

            await expect.element(page.getByLabelText("勤務表出力")).not.toBeDisabled();
        });
    });
})