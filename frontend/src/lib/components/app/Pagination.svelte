<script lang="ts">
    import {
        ArrowBigDownIcon,
        ArrowBigUpIcon,
        ChevronDownIcon,
        ChevronUpIcon,
    } from "@lucide/svelte";
    import { shortcut, type ShortcutEventDetail } from "@svelte-put/shortcut";
    import type { Snippet } from "svelte";
    import PaginationItem from "./PaginationItem.svelte";
    import type { Page } from "./types";

    interface Props {
        pages: Page[];
        selected: Page;
        selectPage: (page: Page) => void;
        children: Snippet;
    }

    let { pages, selected, selectPage, children }: Props = $props();

    let selectedPageIndex = $derived(
        pages.findIndex((page) => {
            return selected.id === page.id;
        }),
    );
    let isPrevDisabled = $derived(selectedPageIndex <= 0);
    let isNextDisabled = $derived(selectedPageIndex >= pages.length - 1);

    function onClickPrevious() {
        if (!isPrevDisabled) {
            selectPage(pages[selectedPageIndex - 1]);
        }
    }
    function onClickNext() {
        if (!isNextDisabled) {
            selectPage(pages[selectedPageIndex + 1]);
        }
    }
</script>

<svelte:window
    use:shortcut={{
        trigger: [
            {
                key: "ArrowUp",
                callback: (_: ShortcutEventDetail) => onClickPrevious(),
                preventDefault: true,
            },
            {
                key: "ArrowDown",
                callback: (_: ShortcutEventDetail) => onClickNext(),
                preventDefault: true,
            },
        ],
    }}
/>
<PaginationItem
    onclick={onClickPrevious}
    label={"Previous"}
    disabled={isPrevDisabled}
    ChevronIcon={ChevronUpIcon}
    ArrowBigIcon={ArrowBigUpIcon}
/>
{@render children()}
<PaginationItem
    class="mt-auto"
    label={"Next"}
    onclick={onClickNext}
    disabled={isNextDisabled}
    ChevronIcon={ChevronDownIcon}
    ArrowBigIcon={ArrowBigDownIcon}
/>
