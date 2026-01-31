<script lang="ts">
    import * as Alert from "$lib/components/ui/alert";
    import * as Card from "$lib/components/ui/card";
    import { BadgeAlertIcon } from "@lucide/svelte";

    let {
        errorMessage,
    }: {
        errorMessage: string;
    } = $props();
</script>

{#if errorMessage !== ""}
    <Card.Description>
        <Alert.Root variant="destructive">
            <BadgeAlertIcon />
            <Alert.Title role="heading">エラー</Alert.Title>
            <Alert.Description>
                {errorMessage}
            </Alert.Description>
        </Alert.Root>
    </Card.Description>
{/if}
