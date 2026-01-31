<script lang="ts">
    import * as Empty from "$lib/components/ui/empty";
    import { type Component } from "svelte";

    interface Props {
        title: string;
        Icon?: Component;
        imgSrcWhenEmpty?: string;
    }

    let { title, Icon, imgSrcWhenEmpty }: Props = $props();
</script>

<Empty.Root>
    <Empty.Header>
        <Empty.Media variant="icon">
            <Icon aria-hidden="true" data-testid="empty-item-icon" />
        </Empty.Media>
        <Empty.Title>
            {title}が読み込まれていません。
        </Empty.Title>
        {#if imgSrcWhenEmpty}
            <Empty.Description>例:</Empty.Description>
        {/if}
    </Empty.Header>
    {#if imgSrcWhenEmpty}
        <img src={imgSrcWhenEmpty} alt={imgSrcWhenEmpty} />
    {/if}
</Empty.Root>
