<script lang="ts">
    import * as Alert from "$lib/components/ui/alert";
    import { Button } from "$lib/components/ui/button";
    import * as Card from "$lib/components/ui/card";
    import * as Item from "$lib/components/ui/item";
    import * as Kbd from "$lib/components/ui/kbd";
    import { Spinner } from "$lib/components/ui/spinner";
    import * as Tooltip from "$lib/components/ui/tooltip";
    import {
        Generate,
        OpenFile,
        SelectDirectory,
    } from "$wailsjs/go/main/WailsApp";
    import { cmd } from "$wailsjs/go/models";
    import { BadgeCheckIcon, FileOutputIcon, FolderIcon } from "@lucide/svelte";
    import { shortcut, type ShortcutEventDetail } from "@svelte-put/shortcut";
    import CardDescriptionOnError from "./CardDescriptionOnError.svelte";
    import FileSelectorItem from "./FileSelectorItem.svelte";
    import { paths } from "./paths.svelte";
    import type { Page } from "./types";

    interface Props {
        title: string;
        inputPages: Page[];
    }

    let { title, inputPages }: Props = $props();

    const PathKey = cmd.PathKey.OUTPUT_DIR;

    let isLoading = $state(false);
    let generationLogs: Record<string, Error> = $state({});
    let errorMessage = $state("");

    let isReady = $derived(
        inputPages.every((v) => v.isSuccess) && paths.value[PathKey],
    );
    let disableGenerate = $derived(!isReady || isLoading);
    let successfulLogs = $derived(
        Object.entries(generationLogs)
            .filter(([, error]) => !error)
            .map(([key]) => `${key}を作成しました。`)
            .sort(),
    );
    let failureLogs = $derived(
        Object.entries(generationLogs)
            .filter(([, error]) => error)
            .map(
                ([key, error]) =>
                    `${key}の出力に失敗しました: ${error.message}`,
            )
            .sort(),
    );

    function onClickGenerate() {
        isLoading = true;
        Generate(paths.value as Record<cmd.PathKey, string>)
            .then((results) => {
                generationLogs = results;
            })
            .catch((fatalError) => {
                errorMessage = fatalError;
            })
            .finally(() => {
                isLoading = false;
            });
    }

    function onClickSelect() {
        SelectDirectory()
            .then((new_path) => {
                paths.update({ [PathKey]: new_path });
            })
            .catch((error) => {
                errorMessage = "選択エラー: " + error;
            });
    }

    function onClickOpen() {
        OpenFile(paths.value[PathKey] || "").catch((error) => {
            errorMessage = "ファイルオープンエラー: " + error;
        });
    }
</script>

<svelte:window
    use:shortcut={{
        trigger: {
            key: "Enter",
            modifier: "ctrl",
            callback: (_: ShortcutEventDetail) => {
                if (!disableGenerate) onClickGenerate();
            },
            preventDefault: true,
        },
    }}
/>
<Card.Root>
    <Card.Header>
        <Card.Title {title}>{title}</Card.Title>
        {#if successfulLogs.length > 0 && failureLogs.length === 0}
            <Card.Description>
                <Alert.Root
                    class="bg-green-100 border-green-600 text-green-800"
                >
                    <BadgeCheckIcon />
                    <Alert.Title>出力成功!</Alert.Title>
                </Alert.Root>
            </Card.Description>
        {/if}
        <CardDescriptionOnError {errorMessage} />
    </Card.Header>
    <Card.Content>
        <Item.Group>
            <FileSelectorItem
                header="出力フォルダ"
                placeholder="出力先フォルダを選択"
                path={paths.value[cmd.PathKey.OUTPUT_DIR]}
                Icon={FolderIcon}
                {onClickSelect}
                {onClickOpen}
            />

            <Tooltip.Provider>
                <Tooltip.Root>
                    <Tooltip.Trigger>
                        <Item.Root>
                            <Item.Content>
                                <Button
                                    variant="destructive"
                                    disabled={disableGenerate}
                                    onclick={onClickGenerate}
                                    aria-label={"勤務表出力"}
                                >
                                    {#if isLoading}
                                        <Spinner role="status" />勤務表出力中…
                                    {:else}
                                        <FileOutputIcon />
                                        勤務表出力
                                        <Kbd.Group>
                                            <Kbd.Root>Ctrl</Kbd.Root>
                                            <Kbd.Root>Enter</Kbd.Root>
                                        </Kbd.Group>
                                    {/if}
                                </Button>
                            </Item.Content>
                        </Item.Root>
                    </Tooltip.Trigger>
                    {#if !isReady}
                        <Tooltip.Content role="tooltip">
                            全入力を完了してください。
                        </Tooltip.Content>
                    {/if}
                </Tooltip.Root>
            </Tooltip.Provider>

            {#if successfulLogs.length > 0}
                <Alert.Root
                    class="bg-green-100 border-green-600 text-green-800"
                >
                    {#each successfulLogs as logs}
                        <BadgeCheckIcon />
                        <Alert.Title>
                            {logs}
                        </Alert.Title>
                    {/each}
                </Alert.Root>
            {/if}
            {#if failureLogs.length > 0}
                <Alert.Root class="bg-red-100 border-red-600 text-red-800">
                    {#each failureLogs as logs}
                        <Alert.Title>
                            <span class="font-bold">{logs}</span>
                        </Alert.Title>
                    {/each}
                </Alert.Root>
            {/if}
        </Item.Group>
    </Card.Content>
</Card.Root>
