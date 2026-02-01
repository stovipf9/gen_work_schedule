<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import * as Item from "$lib/components/ui/item";
    import * as Kbd from "$lib/components/ui/kbd";
    import type { Component } from "svelte";
    import type { ClassValue } from "svelte/elements";

    interface Props {
        class?: ClassValue;
        label: string;
        onclick: () => void;
        disabled: boolean;
        ChevronIcon: Component;
        ArrowBigIcon: Component;
    }
    let {
        class: className,
        label,
        onclick,
        disabled,
        ChevronIcon,
        ArrowBigIcon,
    }: Props = $props();
</script>

<Item.Root class={className} data-testid="pagination-item-top">
    <Item.Content>
        <Button variant="secondary" {onclick} {disabled} aria-label={label}>
            <ChevronIcon />
            <Kbd.Root>
                <ArrowBigIcon />
            </Kbd.Root>
        </Button>
    </Item.Content>
</Item.Root>
