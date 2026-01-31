import { cmd } from "$wailsjs/go/models";
import { type ComponentProps } from "svelte";
import { afterEach, beforeEach, describe, expect, test, vi } from "vitest";
import { render } from "vitest-browser-svelte";
import { page, userEvent } from "vitest/browser";
import GenerateCard from "./GenerateCard.svelte";
import { paths } from "./paths.svelte";
import { mockWailsAppFn } from "./test-utils";

const defaultProps = {
    title: "Test title",
    inputPages: [
        {
            id: crypto.randomUUID(),
            label: "発注情報",
            isSuccess: true
        },
        {
            id: crypto.randomUUID(),
            label: "名簿",
            isSuccess: true
        },
        {
            id: crypto.randomUUID(),
            label: "勤務表テンプレート",
            isSuccess: true
        },
    ],
} satisfies ComponentProps<typeof GenerateCard>;

describe("GenerateCard component", () => {
    let mockGenerate = vi.fn();
    beforeEach(async () => {
        mockWailsAppFn("SetPaths");
        mockGenerate = mockWailsAppFn("Generate").mockResolvedValue({});
        await paths.update({ [cmd.PathKey.OUTPUT_DIR]: "/mock/path" });
    });
    afterEach(() => {
        vi.clearAllMocks();
    });

    describe("Props: タイトル", () => {
        test("与えた文字列が表示されること", async () => {
            render(GenerateCard, defaultProps);

            const title = page.getByTitle(defaultProps.title);
            await expect.element(title).toBeVisible();
        });
    });

    describe("button: 出力", () => {
        const genButtonText = '勤務表出力 Ctrl Enter'
        test("左クリック時に生成処理が呼び出されること", async () => {
            render(GenerateCard, defaultProps);

            const generateButton = page.getByText(genButtonText);
            await generateButton.click();

            expect(mockGenerate).toHaveBeenCalledExactlyOnceWith(paths.value);
        });
        test("生成処理の結果成功のみの場合、\
            出力成功の文字列と出力成功したファイル名が表示され、\
            失敗ログは表示されないこと", async () => {
            mockGenerate.mockResolvedValue({ "1.xlsx": undefined });
            render(GenerateCard, defaultProps);

            const generateButton = page.getByText(genButtonText);
            await generateButton.click();

            const completeMessage = page.getByText("出力成功!");
            await expect.element(completeMessage).toBeVisible();
            const successfulLog = page.getByText("1.xlsxを作成しました。");
            await expect.element(successfulLog).toBeVisible();
            const failureLog = page.getByText("失敗");
            await expect.element(failureLog).not.toBeInTheDocument();
        });
        test("生成処理の結果成功と失敗が混じっている場合\
            出力成功の文字列が表示されず、\
            成功したファイル名は成功したことが表示され、\
            失敗したファイル名が失敗したことが表示されること", async () => {
            mockGenerate.mockResolvedValue({
                "1.xlsx": undefined,
                "2.xlsx": new Error("error occurred in generating 2.xlsx")
            });
            render(GenerateCard, defaultProps);

            const generateButton = page.getByText(genButtonText);
            await generateButton.click();

            const completeMessage = page.getByText("出力成功!")
            await expect.element(completeMessage).not.toBeInTheDocument();
            const successfulLog = page.getByText("1.xlsxを作成しました。");
            await expect.element(successfulLog).toBeVisible();
            const failureLog = page.getByText(new RegExp("2.xlsxの出力に失敗しました.*occurred.*"))
            await expect.element(failureLog).toBeVisible();
        });
        test("生成処理が致命的失敗の場合エラーメッセージが表示されること", async () => {
            mockGenerate.mockRejectedValue("fatal error in generating schedules");
            render(GenerateCard, defaultProps);

            const generateButton = page.getByText(genButtonText);
            await generateButton.click();

            const errorMessage = page.getByText("fatal error in generating schedules");
            await expect.element(errorMessage).toBeVisible();
        });
        test("生成中の場合ボタンが非アクティブかつスピナー表示かつ\
            勤務表出力中の文字列になること", async () => {
            render(GenerateCard, defaultProps);

            // 生成処理を長引かせ一度実行
            mockGenerate.mockImplementation(() =>
                new Promise(rslv => setTimeout(() => rslv([]), 100))
            );
            const generateButton = page.getByText(genButtonText);
            await generateButton.click();

            await expect.element(generateButton).not.toBeInTheDocument();
            const generatingButton = page.getByText("勤務表出力中");
            await expect.element(generatingButton).toBeDisabled();
            const spinner = page.getByRole("status");
            await expect.element(spinner).toBeVisible();
        });
        test("入力未完の場合に入力完了を促すツールチップが表示され\
            かつボタンが非活性であること", async () => {
            render(GenerateCard, {
                ...defaultProps,
                inputPages: [{ id: crypto.randomUUID(), label: "label_A", isSuccess: false }]
            });

            const generateButton = page.getByText(genButtonText);
            await generateButton.hover();

            const tooltip = page.getByRole("tooltip");
            await expect.element(tooltip).toContainHTML("全入力を完了してください");
            await expect.element(generateButton).toBeDisabled();
        });
        test("出力フォルダのパスが空の場合に入力完了を促すツールチップが\
            表示され、かつボタンが非活性であること", async () => {
            await paths.update({ [cmd.PathKey.OUTPUT_DIR]: "" });
            render(GenerateCard, defaultProps);

            const generateButton = page.getByText(genButtonText);
            await generateButton.hover();

            const tooltip = page.getByRole("tooltip");
            await expect.element(tooltip).toContainHTML("全入力を完了してください");
            await expect.element(generateButton).toBeDisabled();
        });
        test("入力完了かつ出力フォルダのパスが空でない場合に\
            入力完了を促すツールチップが表示されず\
            かつボタンが活性であること", async () => {
            render(GenerateCard, defaultProps);

            const generateButton = page.getByText(genButtonText);
            await generateButton.hover();

            const tooltip = page.getByRole("tooltip");
            await expect.element(tooltip).not.toBeInTheDocument();
            await expect.element(generateButton).not.toBeDisabled();
        });
    });

    describe("button: 選択", () => {
        let mockSelectDirectory = vi.fn();
        beforeEach(async () => {
            mockSelectDirectory = mockWailsAppFn("SelectDirectory").mockResolvedValue("/mock/new_path");
            await paths.update({ [cmd.PathKey.OUTPUT_DIR]: "/mock/old_path" });
        });

        test("左クリック時にフォルダ選択処理が呼び出されること", async () => {
            render(GenerateCard, defaultProps);

            const selectButton = page.getByRole("button", { name: "ファイル/フォルダ選択" });
            await selectButton.click();

            expect(mockSelectDirectory).toHaveBeenCalledOnce();
        });
        test("フォルダ選択処理成功時に出力フォルダパスが更新されること", async () => {
            render(GenerateCard, defaultProps);

            const selectButton = page.getByRole("button", { name: "ファイル/フォルダ選択" });
            await selectButton.click();

            expect(paths.value[cmd.PathKey.OUTPUT_DIR]).toEqual("/mock/new_path");
        });
        test("フォルダ選択処理失敗時に出力フォルダパスは更新されずエラーメッセージが表示されること", async () => {
            mockSelectDirectory.mockRejectedValue("error in selecting directory");
            render(GenerateCard, defaultProps);

            const selectButton = page.getByRole("button", { name: "ファイル/フォルダ選択" });
            await selectButton.click();

            const errorMessage = page.getByText("error in selecting directory");
            await expect.element(errorMessage).toBeVisible();
        });
    });
    describe("button: 開く", () => {
        let mockOpenFile = vi.fn();
        beforeEach(async () => {
            mockOpenFile = mockWailsAppFn("OpenFile").mockRejectedValue("error in opening file");
            await paths.update({ [cmd.PathKey.OUTPUT_DIR]: "/mock/path" });
        });

        test("左クリック時にオープン処理が呼ばれること", async () => {
            render(GenerateCard, defaultProps);

            const openButton = page.getByRole("button", { name: "開く" });
            await openButton.click();

            expect(mockOpenFile).toHaveBeenCalledExactlyOnceWith("/mock/path");
        });
        test("オープン処理失敗時にエラーメッセージが表示されること", async () => {
            render(GenerateCard, defaultProps);

            const openButton = page.getByRole("button", { name: "開く" });
            await openButton.click();

            const errorMessage = page.getByText("error in opening file");
            await expect.element(errorMessage).toBeVisible();
        });
    });

    describe("ホットキー: Ctrl+Enter", () => {
        let mockGenerate = vi.fn();
        beforeEach(async () => {
            mockGenerate = mockWailsAppFn("Generate").mockResolvedValue({});
            await paths.update({ [cmd.PathKey.OUTPUT_DIR]: "/mock/path" });
        });
        test("入力完了かつ生成中でない場合に生成処理が呼び出されること", async () => {
            render(GenerateCard, {
                ...defaultProps,
                inputPages: [{ id: crypto.randomUUID(), label: "label_A", isSuccess: true }]
            });

            await userEvent.keyboard('{Control>}{Enter/}');

            expect(mockGenerate).toHaveBeenCalledOnce();
        });
        test("入力未完かつ生成中でない場合に生成処理が呼び出されないこと", async () => {
            render(GenerateCard, {
                ...defaultProps,
                inputPages: [{ id: crypto.randomUUID(), label: "label_A", isSuccess: false }]
            });

            await userEvent.keyboard('{Control>}{Enter/}');

            expect(mockGenerate).not.toHaveBeenCalledOnce();
        });
        test("入力完了かつ生成中の場合に生成処理が呼び出されないこと", async () => {
            render(GenerateCard, {
                ...defaultProps,
                inputPages: [{ id: crypto.randomUUID(), label: "label_A", isSuccess: true }]
            });

            // 生成処理を長引かせ2回連打
            mockGenerate.mockImplementation(() =>
                new Promise(rslv => setTimeout(() => rslv([]), 100))
            );
            await userEvent.keyboard('{Control>}{Enter/}');
            await userEvent.keyboard('{Control>}{Enter/}');

            expect(mockGenerate).toHaveBeenCalledOnce();
        });
        test("パスが空かつ入力完了かつ生成中でない場合に生成処理が呼び出されること", async () => {
            await paths.update({ [cmd.PathKey.OUTPUT_DIR]: "" });
            render(GenerateCard, {
                ...defaultProps,
                inputPages: [{ id: crypto.randomUUID(), label: "label_A", isSuccess: true }]
            });

            await userEvent.keyboard('{Control>}{Enter/}');

            expect(mockGenerate).not.toHaveBeenCalledOnce();
        });
    });


});