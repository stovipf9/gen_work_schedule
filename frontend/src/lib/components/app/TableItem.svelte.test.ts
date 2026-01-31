import { main } from "$wailsjs/go/models";
import { tick, type ComponentProps } from "svelte";
import { beforeEach, describe, expect, test, vi } from "vitest";
import { render } from "vitest-browser-svelte";
import { page } from "vitest/browser";
import TableItem from "./TableItem.svelte";

const mockOnError = vi.fn();
const mockGetColumnNames = vi.fn().mockResolvedValue(
    [{ field: "field_A", label: "列A" } as main.TableColumn]
);
const mockRenderField = vi.fn().mockReturnValue("mocked field")

const defaultProps = {
    title: "Test title",
    records: [{ field_A: "1" }],
    onError: mockOnError,
    getColumnNames: mockGetColumnNames,
    renderField: mockRenderField
} satisfies ComponentProps<typeof TableItem>;

describe("TableItem component", () => {
    beforeEach(() => {
        vi.clearAllMocks();
    });

    describe("Props: タイトル", () => {
        test("与えた文字列が表示されること", async () => {
            render(TableItem, defaultProps);

            const title = page.getByRole("caption");
            await expect.element(title).toBeVisible();
        });
    });

    describe("Props: 列名取得", () => {
        test("初期表示時に処理が呼び出されること", async () => {
            render(TableItem, defaultProps);

            const columnLabel = page.getByText("列A");
            const columnField = page.getByText("field_A");
            expect(mockGetColumnNames).toHaveBeenCalledOnce();
            await expect.element(columnLabel).toBeVisible();
            await expect.element(columnField).not.toBeInTheDocument();
        });

        test("列名が空の場合にはフィールド名が表示されること", async () => {
            const mockGetColumnNamesEmptyLabel = vi.fn().mockResolvedValue(
                [{ field: "field_A", label: "" } as main.TableColumn]
            );
            render(TableItem, {
                ...defaultProps,
                getColumnNames: mockGetColumnNamesEmptyLabel
            });

            const columnField = page.getByText("field_A")
            await expect.element(columnField).toBeVisible();
        });

        test("複数列の場合に表示できること", async () => {
            const mockGetColumnNamesMultiple = vi.fn().mockResolvedValue(
                [{ field: "field_A", label: "列A" },
                { field: "field_B", label: "列B" }]
            );
            render(TableItem, {
                ...defaultProps,
                getColumnNames: mockGetColumnNamesMultiple
            });

            const columnLabelA = page.getByText("列A")
            await expect.element(columnLabelA).toBeVisible();
            const columnLabelB = page.getByText("列B")
            await expect.element(columnLabelB).toBeVisible();
        });

        test("処理失敗時にエラーハンドラが呼び出されること", async () => {
            render(TableItem,
                {
                    ...defaultProps,
                    getColumnNames: vi.fn().mockRejectedValue(
                        "error in getting column names")
                });

            await tick();

            expect(mockOnError).toHaveBeenCalledExactlyOnceWith(
                "カラム名取得エラー: error in getting column names"
            );

            vi.useRealTimers();
        });
    });

    describe("Props: フィールド表示", () => {
        test("レコードに該当列名が存在する場合に表示処理が呼ばれその戻り値が表示されること", async () => {
            render(TableItem, defaultProps);

            await tick();

            const fieldValue = page.getByText("mocked field");
            expect(mockRenderField).toHaveBeenCalledExactlyOnceWith(
                defaultProps.records[0], "field_A"
            );
            await expect.element(fieldValue).toBeVisible();
        });
        test("レコードに該当列名が存在しない場合に表示処理が呼ばれずN/Aが表示されること", async () => {
            render(TableItem, {
                ...defaultProps,
                records: [{ field: "field_B", label: "列B" }]
            });

            await tick();

            const fieldValue = page.getByText("N/A");
            expect(mockRenderField).not.toHaveBeenCalled();
            await expect.element(fieldValue).toBeVisible();
        });
    })
});