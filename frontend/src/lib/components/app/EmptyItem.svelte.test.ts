
import { FolderIcon } from "@lucide/svelte";
import type { ComponentProps } from "svelte";
import { describe, expect, test } from "vitest";
import { render } from "vitest-browser-svelte";
import { page } from "vitest/browser";
import EmptyItem from "./EmptyItem.svelte";

const defaultProps = {
    title: "Test title",
    Icon: FolderIcon,
    imgSrcWhenEmpty: "/mock/empty.svg",
} satisfies ComponentProps<typeof EmptyItem>;

describe("EmptyItem Component", () => {
    describe("Props: タイトル", () => {
        test("与えた文字列が表示されること", async () => {
            render(EmptyItem, defaultProps);

            const title = page.getByText(`${defaultProps.title}が読み込まれていません。`);
            await expect.element(title).toBeInTheDocument();
        })
    });

    describe("Props: アイコン", () => {
        test("与えたアイコンが表示されること", async () => {
            render(EmptyItem, defaultProps);

            const icon = page.getByTestId("empty-item-icon");
            await expect.element(icon).toBeInTheDocument();
        });
    });

    describe("Props: 画像ソース", () => {
        test("空の場合、例や画像は表示されないこと", async () => {
            render(EmptyItem, { ...defaultProps, imgSrcWhenEmpty: "" });

            const example = page.getByText("例:");
            const img = page.getByRole("img");
            await expect.element(example).not.toBeInTheDocument();
            await expect.element(img).not.toBeInTheDocument();
        });

        test("空でない場合、例や画像が表示されること", async () => {
            render(EmptyItem, defaultProps);

            const example = page.getByText("例:");
            const img = page.getByRole("img");
            await expect.element(example).toBeInTheDocument();
            await expect.element(img).toBeVisible();
            await expect.element(img).toHaveAttribute("alt", defaultProps.imgSrcWhenEmpty);
        })
    });
});