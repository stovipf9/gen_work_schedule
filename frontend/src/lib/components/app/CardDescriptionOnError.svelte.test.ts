import { describe, expect, test } from 'vitest';
import { render } from 'vitest-browser-svelte';
import { page } from 'vitest/browser';
import CardDescriptionOnError from './CardDescriptionOnError.svelte';

describe('CardDescriptionOnError Component', () => {
    describe("Props: エラーメッセージ", () => {
        test('空文字でない場合にアラートが表示されること', async () => {
            const errorMessage = "error message";

            render(CardDescriptionOnError, { errorMessage: errorMessage });

            const alert = page.getByRole("alert");
            await expect.element(alert).toBeInTheDocument();

            const title = alert.getByRole("heading", { name: "エラー" });
            await expect.element(title).toBeInTheDocument();

            await expect.element(alert).toHaveTextContent(errorMessage);
        });

        test('空である場合にアラートが表示されないこと', async () => {
            const errorMessage = "";

            render(CardDescriptionOnError, { errorMessage: errorMessage });

            const alert = page.getByRole("alert");
            await expect.element(alert).not.toBeInTheDocument();
        });
    });

});