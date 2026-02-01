import type { ComponentProps } from "svelte";
import { test, describe, vi, expect, afterEach } from "vitest";
import PaginationItem from "./PaginationItem.svelte";
import { ArrowBigUpIcon, ChevronUpIcon } from "@lucide/svelte";
import { render } from "vitest-browser-svelte";
import { page } from "vitest/browser";

const mockOnClick = vi.fn();

const defaultProps = {
    class: "mt-0",
    label: "Previous",
    onclick: mockOnClick,
    disabled: false,
    ChevronIcon: ChevronUpIcon,
    ArrowBigIcon: ArrowBigUpIcon
} satisfies ComponentProps<typeof PaginationItem>;

describe("PaginationItem component", () => {
    afterEach(() => {
        vi.clearAllMocks();
    });

    describe("Props: class", () => {
        test("トップレベルに与えたクラスが適用されること", async () => {
            render(PaginationItem, { ...defaultProps, class: "mytestclass" });

            const topElem = page.getByTestId("pagination-item-top");

            await expect.element(topElem).toHaveClass("mytestclass");
        });
    });

    describe("Props: ラベル", () => {
        test("与えたラベルを持つ要素が存在すること", async () => {
            render(PaginationItem, defaultProps);

            const labeledElement = page.getByLabelText(defaultProps.label);

            await expect.element(labeledElement).toBeVisible();
        })
    })

    describe("Props: 押下時ハンドラ", () => {
        test("ボタン左クリック時に実行されること", async () => {
            render(PaginationItem, defaultProps);

            const button = page.getByRole("button");
            await button.click();

            expect(mockOnClick).toHaveBeenCalledOnce();
        });
    });

    describe("Props: 活性", () => {
        test("falseのときにボタンが活性であること", async () => {
            render(PaginationItem, defaultProps);

            const button = page.getByRole("button");

            await expect.element(button).not.toBeDisabled();
        });
        test("trueのときにボタンが非活性であること", async () => {
            render(PaginationItem, { ...defaultProps, disabled: true });

            const button = page.getByRole("button");

            await expect.element(button).toBeDisabled();
        });
    });

})