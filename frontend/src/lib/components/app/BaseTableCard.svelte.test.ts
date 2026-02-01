import { cmd } from "$wailsjs/go/models";
import { tick, type ComponentProps } from "svelte";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { render } from "vitest-browser-svelte";
import { page } from "vitest/browser";
import BaseTableCard from "./BaseTableCard.svelte";
import { pathKeyConfigMap } from "./pathKeyConfigMap";
import { paths } from "./paths.svelte";
import { mockWailsAppFn } from "./test-utils";

const mockSetCardCompleted = vi.fn();
const mockGetColumnNames = vi.fn().mockResolvedValue([{ field: "field_A", label: "列A" }]);
const mockLoadData = vi.fn().mockResolvedValue([{ field_A: "値A" }]);
const mockRenderField = vi.fn();

const defaultProps = {
    title: "Test Table",
    pathKey: cmd.PathKey.ORDER_XLSX,
    setCardCompleted: mockSetCardCompleted,
    tableHandlers: {
        getColumnNames: mockGetColumnNames,
        loadData: mockLoadData,
        renderField: mockRenderField
    }
} satisfies ComponentProps<typeof BaseTableCard>;

describe("BaseTableCard component", () => {
    beforeEach(() => {
        mockWailsAppFn("SetPaths");
    });

    afterEach(async () => {
        vi.clearAllMocks();
        await paths.update({ [defaultProps.pathKey]: "" });
    })

    describe("Props: タイトル", () => {
        test("与えた文字列が表示されること (空パス時)", async () => {
            render(BaseTableCard, defaultProps);

            const title = page.getByTitle(defaultProps.title);
            await expect.element(title).toBeVisible();
            const fileSelectorHeader = page.getByRole("heading", { name: defaultProps.title });
            await expect.element(fileSelectorHeader).toBeVisible();
            const placeholder = page.getByText(new RegExp(`${defaultProps.title}ファイル.*を選択`));
            await expect.element(placeholder).toBeVisible();
            const previewHeader = page.getByText(`${defaultProps.title}プレビュー`);
            await expect.element(previewHeader).toBeVisible();
            const emptyItemTitle = page.getByText(`${defaultProps.title}が読み込まれていません`);
            await expect.element(emptyItemTitle).toBeVisible();
        });

        test("与えた文字列が表示されること (非空パス時)", async () => {
            paths.update({ [cmd.PathKey.ORDER_XLSX]: "/mock/order.xlsx" });
            render(BaseTableCard, defaultProps);

            const tableTitle = page.getByRole("caption");
            await expect.element(tableTitle).toBeVisible();
        });
    });

    describe("Props: パスキー", () => {
        test("与えた値に対応するパスが表示されること", async () => {
            const pathKey = cmd.PathKey.TEMPLATE_XLSX;
            paths.update({ [pathKey]: "/mock/given_path" });
            render(BaseTableCard, { ...defaultProps, pathKey: pathKey });

            const path = page.getByText("/mock/given_path");
            await expect.element(path).toBeVisible();
        });
        test("与えた値に対応する拡張子が表示されること", async () => {
            render(BaseTableCard, defaultProps);

            const extension = page.getByText("xlsx");
            await expect.element(extension).toBeVisible();
        });
        test("与えた値がサポート外であれば拡張子は表示されないこと", async () => {
            render(BaseTableCard, { ...defaultProps, pathKey: cmd.PathKey.TEMPLATE_XLSX });

            const invisibleExtension = page.getByText("xlsx");
            await expect.element(invisibleExtension).not.toBeInTheDocument();
        });
    });

    describe("Props: 完了処理", () => {
        beforeEach(async () => {
            await paths.update({ [defaultProps.pathKey]: "/mock/path" });
        });

        test("データロードが成功した場合には真で呼ばれること", async () => {
            render(BaseTableCard, defaultProps);

            await tick();
            expect(mockLoadData).toHaveBeenCalledOnce();
            expect(mockLoadData).toHaveResolvedTimes(1);
            expect(mockSetCardCompleted).toHaveBeenCalledExactlyOnceWith(true);
        });

        test("データロードが失敗した場合には偽で呼ばれること", async () => {
            const mockLoadDataReject = vi.fn().mockRejectedValue("error in loading data");
            render(BaseTableCard, {
                ...defaultProps, tableHandlers: {
                    ...defaultProps.tableHandlers,
                    loadData: mockLoadDataReject
                }
            });

            expect(mockLoadDataReject).toHaveBeenCalledOnce();
            await expect(mockLoadDataReject).rejects.toThrow();
            expect(mockSetCardCompleted).toHaveBeenCalledExactlyOnceWith(false);
        });
    });

    describe("Props: 列名取得", () => {
        test("パスキーに対するパスが非空、ロードされたデータが1件以上の時に処理が呼び出されること", async () => {
            render(BaseTableCard, defaultProps);

            paths.update({ [defaultProps.pathKey]: "/mock/path" });

            await tick();
            expect(mockGetColumnNames).toHaveBeenCalled();
        })
    });

    describe("Props: データロード処理", () => {
        test("初期表示時にパスが渡され呼び出されること", async () => {
            paths.update({ [defaultProps.pathKey]: "/mock/path" });
            render(BaseTableCard, defaultProps);

            await tick();
            expect(mockLoadData).toHaveBeenCalledExactlyOnceWith("/mock/path");
        });
        test("データロード失敗時にエラーメッセージが表示されること", async () => {
            await paths.update({ [defaultProps.pathKey]: "/mock/path" });
            const mockLoadDataReject = vi.fn().mockRejectedValue("error in loading data");
            render(BaseTableCard, {
                ...defaultProps,
                tableHandlers: {
                    ...defaultProps.tableHandlers,
                    loadData: mockLoadDataReject
                }
            });

            const errorMessage = page.getByText("error in loading data");
            await expect.element(errorMessage).toBeVisible();
        })
    });

    describe("Props: フィールド表示処理", () => {
        test("パスキーに対するパスが非空の時に処理が呼び出されること", async () => {
            render(BaseTableCard, defaultProps);

            await paths.update({ [defaultProps.pathKey]: "/mock/path" });
            await tick();

            expect(mockRenderField).toHaveBeenCalled();
        })
    });

    describe("Events: 選択ボタン左クリック", () => {
        let mockSelectFile = vi.fn();
        beforeEach(() => {
            mockSelectFile = mockWailsAppFn("SelectFile").mockResolvedValue("/mock/new_path");
        });

        test("ファイル選択処理がパスキーに応じた拡張子で呼び出されること", async () => {
            const extension = pathKeyConfigMap[defaultProps.pathKey]?.extension;
            render(BaseTableCard, defaultProps);

            const selectButton = page.getByRole("button", { name: "ファイル/フォルダ選択" });
            await selectButton.click();

            expect(mockSelectFile).toHaveBeenCalledExactlyOnceWith(extension);
        });
        test("ファイル選択処理成功時にパスが更新されること", async () => {
            await paths.update({ [defaultProps.pathKey]: "/mock/old_path" });
            render(BaseTableCard, defaultProps);

            const selectButton = page.getByRole("button", { name: "ファイル/フォルダ選択" });
            await selectButton.click();

            expect(paths.value[defaultProps.pathKey]).toEqual("/mock/new_path");
        });
        test("ファイル選択処理失敗時にエラーメッセージが表示されること", async () => {
            mockWailsAppFn("SelectFile").mockRejectedValue("error in selecting file");
            render(BaseTableCard, defaultProps);

            const selectButton = page.getByRole("button", { name: "ファイル/フォルダ選択" });
            await selectButton.click();

            const errorMessage = page.getByText("error in selecting file");
            await expect.element(errorMessage).toBeVisible();
        });
    });
    describe("Events: 開くボタンクリック", () => {
        let mockOpenFile = vi.fn();
        beforeEach(async () => {
            mockOpenFile = mockWailsAppFn("OpenFile").mockRejectedValue("");
            await paths.update({ [defaultProps.pathKey]: "/mock/path" });
        });

        test("開く処理がパスキーに応じたパスで呼び出されること", async () => {
            render(BaseTableCard, defaultProps);

            const openButton = page.getByRole("button", { name: "開く" });
            await openButton.click();

            expect(mockOpenFile).toHaveBeenCalledExactlyOnceWith("/mock/path");
        });
        test("開く処理失敗時にエラーメッセージが表示されること", async () => {
            mockWailsAppFn("OpenFile").mockRejectedValue("error in opening file");
            render(BaseTableCard, defaultProps);

            const openButton = page.getByRole("button", { name: "開く" });
            await openButton.click();

            const errorMessage = page.getByText("error in opening file");
            await expect.element(errorMessage).toBeVisible();

            page.getByRole("stat")
        });
    });

    describe("ローディング中", () => {
        beforeEach(async () => {
            await paths.update({ [defaultProps.pathKey]: "/mock/path" });
        });
        test("ロード中でないときスピナーが表示されないこと", async () => {
            render(BaseTableCard, defaultProps);

            // 一度処理が終わるのを待つ
            await tick();
            expect(mockLoadData).toHaveBeenCalled();

            const spinner = page.getByRole("status");
            await expect.element(spinner).not.toBeInTheDocument();
        });
        test("ロード中にスピナーが表示されること", async () => {
            render(BaseTableCard, defaultProps);

            // 一度処理が終わるのを待つ
            await tick();
            expect(mockLoadData).toHaveBeenCalled();

            // ロードを長引かせパス更新をトリガー
            mockLoadData.mockImplementation(() =>
                new Promise(r => setTimeout(() => r([]), 100))
            );
            await paths.update({ [defaultProps.pathKey]: "/mock/new_path" });

            // ロード中にスピナーが表示されている
            const spinner = page.getByRole("status");
            await expect.element(spinner).toBeVisible();
        });
        test("ロード成功後にスピナーが表示されないこと", async () => {
            render(BaseTableCard, defaultProps);

            // 一度処理が終わるのを待つ
            await tick();
            expect(mockLoadData).toHaveBeenCalled();

            // ロードを長引かせパス更新をトリガー
            mockLoadData.mockImplementation(() =>
                new Promise(r => setTimeout(() => r([]), 100))
            );
            await paths.update({ [defaultProps.pathKey]: "/mock/new_path" });

            // ロード中にスピナーが表示されている
            const spinner = page.getByRole("status");
            await expect.element(spinner).toBeVisible();

            // 待てばスピナーが表示されなくなる
            await expect.element(spinner).not.toBeInTheDocument();
        });
        test("ロード失敗後にスピナーが表示されないこと", async () => {
            render(BaseTableCard, defaultProps);

            // 一度処理が終わるのを待つ
            await tick();
            expect(mockLoadData).toHaveBeenCalled();

            // ロードを長引かせパス更新をトリガー
            mockLoadData.mockImplementation(() =>
                new Promise((_, reject) =>
                    setTimeout(() => reject(new Error("error in loading data")), 100))
            );
            await paths.update({ [defaultProps.pathKey]: "/mock/new_path" });

            // ロード中にスピナーが表示されている
            const spinner = page.getByRole("status");
            await expect.element(spinner).toBeVisible();

            // 待てばスピナーが表示されなくなる
            await expect.element(spinner).not.toBeInTheDocument();
        });
    })
})
