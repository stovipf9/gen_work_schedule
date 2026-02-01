
import { tick, type ComponentProps } from "svelte";
import { test, describe, vi, expect, beforeEach, afterEach } from "vitest";
import OrderCard from "./OrderCard.svelte";
import { render } from "vitest-browser-svelte";
import { page } from "vitest/browser";
import { mockWailsAppFn } from "./test-utils";
import { paths } from "./paths.svelte";
import { cmd } from "$wailsjs/go/models";

const mockSetCardCompleted = vi.fn();

const defaultProps = {
    title: "Test title",
    setCardCompleted: mockSetCardCompleted
} satisfies ComponentProps<typeof OrderCard>;

describe("OrderCard component", () => {
    let mockLoadOrders = vi.fn();
    let mockGetOrderColumnNames = vi.fn();
    beforeEach(() => {
        mockLoadOrders = mockWailsAppFn("LoadOrders").mockRejectedValue({});
        mockGetOrderColumnNames = mockWailsAppFn("GetOrderColumnNames").mockRejectedValue("error in getting column names");
        mockWailsAppFn("SetPaths");
    });
    afterEach(() => {
        paths.update({ [cmd.PathKey.ORDER_XLSX]: "" });
        vi.clearAllMocks();
    });

    describe("Props: タイトル", () => {
        test("与えた文字列が表示されること", async () => {
            render(OrderCard, defaultProps);

            const title = page.getByText(defaultProps.title, { exact: true });
            await expect.element(title).toBeVisible();
        });
    });

    describe("Props: 完了処理", () => {
        beforeEach(async () => {
            paths.update({ [cmd.PathKey.ORDER_XLSX]: "/mock/path" });
        })
        test("データロードが完了した場合に呼び出されること", async () => {
            render(OrderCard, defaultProps);

            await tick();
            expect(mockLoadOrders).toHaveBeenCalled();
            expect(mockSetCardCompleted).toHaveBeenCalled();
        })
    });

    describe("function: フィールドのレンダリング", () => {
        beforeEach(async () => {
            await paths.update({ [cmd.PathKey.ORDER_XLSX]: "/mock/path" });
        })
        test.each(
            [
                "startDate",
                "endDate"
            ])("フィールドが%sである値はYYYY-MM-DDで表示されること", async (dateField: string) => {
                mockLoadOrders.mockResolvedValue(
                    [{ [dateField]: "2025-09-30T12:34:56Z" } as Partial<cmd.Order>]);
                mockGetOrderColumnNames.mockResolvedValue([{ field: dateField, label: "日 (テスト)" }]);
                render(OrderCard, defaultProps);

                const dateValue = page.getByText("2025-09-30", { exact: true });
                await expect.element(dateValue).toBeVisible();
            });
        test.each(
            Object.keys(new cmd.Order({}))
                .filter(key => !["startDate", "endDate"].includes(key))
        )("フィールド%sの値はそのまま表示されること", async (field: string) => {
            mockLoadOrders.mockResolvedValue(
                [{ [field]: `${field}_X` }]
            );
            mockGetOrderColumnNames.mockResolvedValue(
                [{ field: field, label: `${field} (テスト)` }]);
            render(OrderCard, defaultProps);

            const value = page.getByText(`${field}_X`, { exact: true });
            await expect.element(value).toBeVisible();
        })
    });
})