<script lang="ts">
    import * as ButtonGroup from "$lib/components/ui/button-group";
    import Button from "$lib/components/ui/button/button.svelte";
    import * as Item from "$lib/components/ui/item";
    import * as Kbd from "$lib/components/ui/kbd";
    import { FolderOpenIcon, SquareArrowOutUpRight } from "@lucide/svelte";
    import { shortcut, type ShortcutEventDetail } from "@svelte-put/shortcut";
    import type { Component } from "svelte";

    interface Props {
        header: string;
        placeholder: string;
        path?: string;
        Icon?: Component;
        onClickSelect?: () => void;
        onClickOpen?: () => void;
    }

    let { header, placeholder, path, Icon, onClickSelect, onClickOpen }: Props =
        $props();
</script>

<svelte:window
    use:shortcut={{
        trigger: {
            key: "Enter",
            modifier: "shift",
            callback: (_: ShortcutEventDetail) => onClickSelect?.(),
            preventDefault: true,
        },
    }}
/>
<Item.Root>
    <Item.Header role="heading">{header}</Item.Header>
    <Item.Media variant="icon">
        <Icon aria-hidden="true" data-testid="file-selector-icon" />
    </Item.Media>
    <Item.Content>
        <Item.Description>
            {path ? path : placeholder}
        </Item.Description>
    </Item.Content>
    <Item.Actions>
        <ButtonGroup.Root>
            <Button onclick={onClickSelect}>
                <FolderOpenIcon />
                ファイル/フォルダ選択
                <Kbd.Group>
                    <Kbd.Root>Shift</Kbd.Root>
                    <Kbd.Root>Enter</Kbd.Root>
                </Kbd.Group>
            </Button>
            <ButtonGroup.Separator />
            <Button disabled={!path} onclick={onClickOpen}>
                <SquareArrowOutUpRight />
                開く
            </Button>
        </ButtonGroup.Root>
    </Item.Actions>
</Item.Root>
