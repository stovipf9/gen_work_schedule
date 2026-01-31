import { createRawSnippet, type ComponentProps } from "svelte";
import { afterEach, describe, expect, test, vi } from "vitest";
import { render } from "vitest-browser-svelte";
import { page, userEvent } from "vitest/browser";
import Pagination from "./Pagination.svelte";

const mockSelectPage = vi.fn();
const children = createRawSnippet(() => ({
    render: () => '<div><p data-testid="child">Test content</p></div>',
    setup: () => { }
}));
const pages = [
    {
        id: crypto.randomUUID(),
        label: "発注情報",
    },
    {
        id: crypto.randomUUID(),
        label: "名簿",
    },
    {
        id: crypto.randomUUID(),
        label: "勤務表テンプレート",
    },
]

const defaultProps = {
    pages: pages,
    selected: pages[0],
    selectPage: mockSelectPage,
    children: children
} satisfies ComponentProps<typeof Pagination>;

describe("Pagination component", () => {
    afterEach(() => {
        vi.clearAllMocks();
    });

    // describe("Props: ページ", () => {
    //     test("")
    // })
    // describe("Props: ページ切り替えハンドラ", () => {
    // });

    describe("Props: 子スニペット", () => {
        test("与えたスニペットが描画されること", async () => {
            render(Pagination, defaultProps);

            const child = page.getByTestId("child");

            await expect.element(child).toBeVisible();
        });
    });

    describe("Button: 前", () => {
        const getPrevButton = () => page.getByRole("button", { name: "Previous" });
        test("選択ページが先頭の場合、非活性であること", async () => {
            render(Pagination, defaultProps);

            const prevButton = getPrevButton();

            await expect.element(prevButton).toBeDisabled();
        });
        test("選択ページが2番目の場合、活性であること", async () => {
            render(Pagination, { ...defaultProps, selected: pages[1] });

            const prevButton = getPrevButton();

            await expect.element(prevButton).not.toBeDisabled();
        });
        test("左クリック時に前ページ参照で呼び出されること", async () => {
            render(Pagination, { ...defaultProps, selected: pages[1] });

            const prevButton = getPrevButton();
            await prevButton.click();

            expect(mockSelectPage).toHaveBeenCalledExactlyOnceWith(pages[0]);
        });
    });

    describe("Button: 次", () => {
        const getNextButton = () => page.getByRole("button", { name: "Next" });
        test("選択ページが末尾の場合、非活性であること", async () => {
            render(Pagination,
                {
                    ...defaultProps,
                    selected: defaultProps.pages[pages.length - 1]
                });

            const nextButton = getNextButton();

            await expect.element(nextButton).toBeDisabled();
        });
        test("選択ページが末尾以外の場合、活性であること", async () => {
            render(Pagination, defaultProps);

            const nextButton = getNextButton();

            await expect.element(nextButton).not.toBeDisabled();
        });
        test("左クリック時に次ページ参照で呼び出されること", async () => {
            render(Pagination, defaultProps);

            const nextButton = getNextButton();
            await nextButton.click();

            expect(mockSelectPage).toHaveBeenCalledExactlyOnceWith(pages[1]);
        });
    });

    describe("ホットキー: 前", () => {
        test("選択ページが先頭の場合、上キー押下でページ選択ハンドラが呼び出されないこと", async () => {
            render(Pagination, defaultProps);

            await userEvent.keyboard("{ArrowUp}");

            expect(mockSelectPage).not.toHaveBeenCalled();
        });
        test("選択ページが2番目の場合、上キー押下でページ選択ハンドラが前ページ参照で呼び出されること", async () => {
            render(Pagination,
                {
                    ...defaultProps,
                    selected: pages[1]
                }
            );

            await userEvent.keyboard("{ArrowUp}");

            expect(mockSelectPage).toHaveBeenCalledExactlyOnceWith(
                pages[0]
            );
        });
    });

    describe("ホットキー: 次", () => {
        test("選択ページが末尾の場合、\
            下キー押下でページ選択ハンドラが呼び出されないこと", async () => {
            render(Pagination,
                {
                    ...defaultProps,
                    selected: pages[pages.length - 1]
                }
            );

            await userEvent.keyboard("{ArrowDown}");

            expect(mockSelectPage).not.toHaveBeenCalled();
        });
        test("選択ページが末尾以外の場合、\
            下キー押下でページ選択ハンドラが次ページ参照で呼び出されること", async () => {
            render(Pagination, defaultProps);

            await userEvent.keyboard("{ArrowDown}");

            expect(mockSelectPage).toHaveBeenCalledExactlyOnceWith(pages[1]);
        })
    })
})