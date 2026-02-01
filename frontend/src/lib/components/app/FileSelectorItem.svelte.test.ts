import { FolderIcon } from "@lucide/svelte";
import type { ComponentProps } from "svelte";
import { beforeEach, describe, expect, test, vi } from "vitest";
import { render } from "vitest-browser-svelte";
import { page, userEvent } from "vitest/browser";
import FileSelectorItem from "./FileSelectorItem.svelte";

const mockSelectHandler = vi.fn();
const mockOpenHandler = vi.fn();

const defaultProps = {
    header: "Test header",
    placeholder: "Select a file...",
    path: "/initial/path",
    Icon: FolderIcon,
    onClickSelect: mockSelectHandler,
    onClickOpen: mockOpenHandler,
} satisfies ComponentProps<typeof FileSelectorItem>;

describe("FileSelectorItem Component", () => {

    beforeEach(() => {
        vi.clearAllMocks();
    })

    describe("Props: ヘッダー", () => {
        test("与えた文字列が表示されること", async () => {
            render(FileSelectorItem, defaultProps);

            const header = page.getByRole("heading", { name: defaultProps.header });
            await expect.element(header).toBeInTheDocument();
        });
    });

    // describe("Props: プレースホルダー", () => {
    // });

    describe("Props: パス", () => {
        test("与えた文字列が空の場合プレースホルダテキストが表示されること", async () => {
            render(FileSelectorItem, { ...defaultProps, path: "" });

            const pathtext = page.getByText(defaultProps.placeholder);
            await expect.element(pathtext).toBeInTheDocument();
        });
        test("与えた空でない文字列が表示され、プレースホルダテキストが表示されないこと", async () => {
            render(FileSelectorItem, defaultProps);

            const pathtext = page.getByText(defaultProps.path);
            await expect.element(pathtext).toBeInTheDocument();

            const placeholderText = page.getByText(defaultProps.placeholder);
            await expect.element(placeholderText).not.toBeInTheDocument();
        });

        test("与えた文字列が空である場合に[開く]ボタンが非アクティブであること", async () => {
            render(FileSelectorItem, { ...defaultProps, path: "" });

            const openButton = page.getByRole("button", { name: "開く" });
            await expect.element(openButton).toBeDisabled();
        });

        test("与えた文字列が空でない場合に[開く]ボタンがアクティブであること", async () => {
            render(FileSelectorItem, defaultProps);

            const openButton = page.getByRole("button", { name: "開く" });
            await expect.element(openButton).toBeEnabled();
        })
    });

    describe("Props: アイコン", () => {
        test("与えたアイコンが表示されること", async () => {
            render(FileSelectorItem, defaultProps);

            const icon = page.getByTestId("file-selector-icon");
            await expect.element(icon).toHaveAttribute("aria-hidden", "true");
        });
    });

    describe("button: ファイル/フォルダ選択", () => {
        test("クリック時にハンドラが実行されること", async () => {
            render(FileSelectorItem, defaultProps);

            const selectButton = page.getByRole("button", { name: "ファイル/フォルダ選択" });
            await selectButton.click();

            expect(mockSelectHandler).toHaveBeenCalledOnce();
        });

    });

    describe("button: 開く", () => {
        test("クリック時にハンドラが実行されること", async () => {
            render(FileSelectorItem, defaultProps);

            const openButton = page.getByRole("button", { name: "開く" });
            await openButton.click();

            expect(mockOpenHandler).toHaveBeenCalledOnce();
        });

    });

    describe("ホットキー", () => {
        test("Shift+Enterでファイル/フォルダ選択ハンドラが実行されること", async () => {
            render(FileSelectorItem, defaultProps);

            await userEvent.keyboard('{Shift>}{Enter/}');

            expect(mockSelectHandler).toHaveBeenCalledOnce();
        });
    });
})