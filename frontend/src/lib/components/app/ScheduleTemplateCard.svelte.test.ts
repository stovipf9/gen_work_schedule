
import { cmd } from "$wailsjs/go/models";
import { tick, type ComponentProps } from "svelte";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { render } from "vitest-browser-svelte";
import { page } from "vitest/browser";
import ScheduleTemplateCard from "./ScheduleTemplateCard.svelte";
import { paths } from "./paths.svelte";
import { mockWailsAppFn } from "./test-utils";

const mockSetCardCompleted = vi.fn();

const defaultProps = {
    title: "Test情報",
    setCardCompleted: mockSetCardCompleted
} satisfies ComponentProps<typeof ScheduleTemplateCard>;

describe("ScheduleTemplateCard component", () => {
    let mockSelectFile = vi.fn();
    let locateSelectButton = () => page.getByRole("button", { name: "ファイル/フォルダ選択" });
    beforeEach(() => {
        mockWailsAppFn("SetPaths");
        mockSelectFile = mockWailsAppFn("SelectFile").mockResolvedValue("");
    });

    afterEach(() => {
        vi.clearAllMocks();
    });

    describe("Props: タイトル", () => {
        test("与えた文字列が表示されること", async () => {
            render(ScheduleTemplateCard, defaultProps);

            const title = page.getByTitle(defaultProps.title);
            const selectorTitle = page.getByRole("heading", { name: `${defaultProps.title}ファイル` })
            const placeholder = page.getByText(`${defaultProps.title}(xlsx)を選択。記入済みも可。`);
            await expect.element(title).toBeVisible();
            await expect.element(selectorTitle).toBeVisible();
            await expect.element(placeholder).toBeVisible();
        });
    });

    describe("Props: 完了処理", () => {
        test("初期 (パスが空) の場合にはfalseで呼ばれること", async () => {
            render(ScheduleTemplateCard, defaultProps);

            expect(mockSetCardCompleted).toHaveBeenCalledExactlyOnceWith(false);
        });

        test("パスが非空となった場合にはtrueで呼ばれること", async () => {
            render(ScheduleTemplateCard, defaultProps);

            paths.update({ [cmd.PathKey.TEMPLATE_XLSX]: "/mock/template_path" });
            await tick();
            expect(mockSetCardCompleted).toHaveBeenLastCalledWith(true);
        });
    })

    describe("function: ファイル選択ハンドラ", () => {
        test("選択ボタン押下でファイル選択処理が呼ばれること", async () => {
            render(ScheduleTemplateCard, defaultProps);

            const selectButton = locateSelectButton();
            await selectButton.click();

            expect(mockSelectFile).toHaveBeenCalledWith("xlsx");
        });
        test("ファイル選択処理が成功した際にパスが更新されること", async () => {
            mockSelectFile.mockResolvedValue("/mock/template_path.xlsx");

            render(ScheduleTemplateCard, defaultProps);

            const selectButton = locateSelectButton();
            await selectButton.click();

            const path = paths.value[cmd.PathKey.TEMPLATE_XLSX];
            expect(path).toBe("/mock/template_path.xlsx");
        });
        test("ファイル選択処理が失敗した際にエラーメッセージが表示されること", async () => {
            mockSelectFile.mockRejectedValue("error in file selection");

            render(ScheduleTemplateCard, defaultProps);

            const selectButton = locateSelectButton();
            await selectButton.click();

            const errorMessage = page.getByText("error in file selection");
            await expect.element(errorMessage).toBeVisible();
        })
    });

    describe("function: 開く", () => {
        let mockOpenFile = vi.fn();

        beforeEach(() => {
            mockOpenFile = mockWailsAppFn("OpenFile").mockRejectedValue("");
        });
        test("開くボタンクリック時にファイルオープン処理が呼ばれること", async () => {
            paths.update({ [cmd.PathKey.TEMPLATE_XLSX]: "/mock/template_path.xlsx" });

            render(ScheduleTemplateCard, defaultProps);

            const openButton = page.getByRole("button", { name: "開く" });
            await openButton.click();

            expect(mockOpenFile).toHaveBeenCalledExactlyOnceWith("/mock/template_path.xlsx");
        });
        test("ファイルオープン処理失敗時にエラーメッセージが表示されること", async () => {
            paths.update({ [cmd.PathKey.TEMPLATE_XLSX]: "/mock/template_path.xlsx" });
            mockOpenFile.mockRejectedValue("error in opening file");
            render(ScheduleTemplateCard, defaultProps);

            const openButton = page.getByRole("button", { name: "開く" });
            await openButton.click();

            const errorMessage = page.getByText("error in opening file");
            await expect.element(errorMessage).toBeVisible();
        });
    });
});