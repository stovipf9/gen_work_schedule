import type { ComponentProps } from "svelte";
import { test, describe, expect, vi } from "vitest";
import AppSidebar from "./AppSidebar.svelte";
import { render } from "vitest-browser-svelte";
import { page } from "vitest/browser";
import { Quit } from "$wailsjs/runtime/runtime";
import type { Page } from "./types";

const mockOnSelected = vi.fn();
const inputsPages: Page[] = [
    {
        id: crypto.randomUUID(),
        label: "発注情報B",
        isSuccess: true
    },
    {
        id: crypto.randomUUID(),
        label: "名簿B",
        isSuccess: false
    },
    {
        id: crypto.randomUUID(),
        label: "勤務表テンプレートB",
    },
]
const defaultProps = {
    pageGroups: {
        inputs: {
            label: "入力A",
            pages: inputsPages,
        },
        outputs: {
            label: "出力A",
            pages: [{ id: crypto.randomUUID(), label: "出力B", isSuccess: true }],
        },
    },
    selected: inputsPages[0],
    onSelected: mockOnSelected
} satisfies ComponentProps<typeof AppSidebar>;

describe("AppSidebar component", () => {
    describe("Props: ページ群", () => {
        test("ページ群の全てのラベルが表示されること", async () => {
            render(AppSidebar, defaultProps);

            for (const pageGroup of Object.values(defaultProps.pageGroups)) {
                const pageGroupLabel = page.getByText(pageGroup.label);
                await expect.element(pageGroupLabel).toBeVisible();
            }
        });
        test("ページの全てのラベルが表示されること", async () => {
            render(AppSidebar, defaultProps);

            for (const pageGroup of Object.values(defaultProps.pageGroups)) {
                for (const eachPage of pageGroup.pages) {
                    const pageLabel = page.getByText(eachPage.label);
                    await expect.element(pageLabel).toBeVisible();
                }
            }
        });
        test("処理完了しているページに対応する項目にのみチェックが付いていること", async () => {
            render(AppSidebar, defaultProps);

            for (const pageGroup of Object.values(defaultProps.pageGroups)) {
                for (const eachPage of pageGroup.pages) {
                    const checkbox = page.getByTestId(`success-checkbox-${eachPage.label}`);
                    if (eachPage.isSuccess) {
                        await expect.element(checkbox).toBeVisible();
                        await expect.element(checkbox).toBeChecked();
                    } else {
                        await expect.element(checkbox).not.toBeInTheDocument();
                    }
                }
            }
        })
    });

    describe("Props: 選択中のページ", () => {
        test("選択中のページに対応する項目が強調されること", async () => {
            render(AppSidebar, defaultProps);

            for (const pageGroup of Object.values(defaultProps.pageGroups)) {
                for (const eachPage of pageGroup.pages) {
                    const button = page.getByRole("button", { name: eachPage.label });

                    if (eachPage.id === defaultProps.selected.id) {
                        await expect.element(button).toHaveAttribute("data-active", "true");
                    } else {
                        await expect.element(button).toHaveAttribute("data-active", "false");
                    }
                }
            }
        })
    });

    describe("Props: 選択コールバック", () => {
        test("各項目を左クリックした際に呼び出されること", async () => {
            render(AppSidebar, defaultProps);

            for (const pageGroup of Object.values(defaultProps.pageGroups)) {
                for (const eachPage of pageGroup.pages) {
                    const button = page.getByRole("button", { name: eachPage.label });
                    await button.click();

                    expect(mockOnSelected).toHaveBeenLastCalledWith(eachPage);
                }
            }
        })
    });

    describe("button: 終了", () => {
        test("左クリック時に終了処理が走ること", async () => {
            vi.mock('$wailsjs/runtime/runtime', () => ({
                Quit: vi.fn()
            }));
            render(AppSidebar, defaultProps);

            const quitButton = page.getByRole("button", { name: "終了" });
            await quitButton.click();

            expect(Quit).toHaveBeenCalledOnce();
        })
    })
})