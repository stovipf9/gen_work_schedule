<script lang="ts" generics="Row extends object">
    import * as Card from "$lib/components/ui/card";
    import * as Item from "$lib/components/ui/item";
    import { Spinner } from "$lib/components/ui/spinner";
    import { OpenFile, SelectFile } from "$wailsjs/go/main/WailsApp";
    import { cmd, main } from "$wailsjs/go/models";
    import CardDescriptionOnError from "./CardDescriptionOnError.svelte";
    import EmptyItem from "./EmptyItem.svelte";
    import FileSelectorItem from "./FileSelectorItem.svelte";
    import TableItem from "./TableItem.svelte";
    import { pathKeyConfigMap } from "./pathKeyConfigMap";
    import { paths } from "./paths.svelte";

    interface Props {
        title: string;
        pathKey: cmd.PathKey;
        setCardCompleted: (result: boolean) => void;
        tableHandlers: {
            getColumnNames: () => Promise<main.TableColumn[]>;
            loadData: (path: string) => Promise<Row[]>;
            renderField: (record: Row, field: keyof Row) => string;
        };
    }

    let { title, pathKey, setCardCompleted, tableHandlers }: Props = $props();

    let errorMessage: string = $state("");
    let isLoading = $state(false);
    let records: Row[] = $state([]);

    let path = $derived(paths.value[pathKey] || "");
    let config = $derived(pathKeyConfigMap[pathKey]);
    let extension = $derived(config?.extension ?? "");
    let Icon = $derived(config?.Icon);
    let imgSrcWhenEmpty = $derived(config?.imgSrcWhenEmpty);

    $effect(() => {
        if (!path) return;

        isLoading = true;
        tableHandlers
            .loadData(path)
            .then((data) => {
                setCardCompleted(true);
                records = data;
            })
            .catch((error) => {
                errorMessage = "発注情報取得エラー: " + error;
                setCardCompleted(false);
                records = [];
            })
            .finally(() => {
                isLoading = false;
            });
    });

    function onClickSelect() {
        SelectFile(extension)
            .then((newPath) => {
                paths.update({ [pathKey]: newPath });
            })
            .catch((error) => {
                errorMessage = "選択エラー: " + error;
            });
    }

    function onClickOpen() {
        OpenFile(path).catch((error) => {
            errorMessage = "ファイルオープンエラー: " + error;
        });
    }
</script>

<Card.Root>
    <Card.Header>
        <Card.Title {title}>{title}</Card.Title>
        <CardDescriptionOnError {errorMessage}></CardDescriptionOnError>
    </Card.Header>
    <Card.Content>
        <Item.Group>
            <FileSelectorItem
                header={title + "ファイル"}
                placeholder={`${title}ファイル(${extension})を選択`}
                {path}
                {onClickSelect}
                {onClickOpen}
                {Icon}
            />
            <Item.Root>
                <Item.Header>{title}プレビュー</Item.Header>
                <Item.Content>
                    {#if !path}
                        <EmptyItem {title} {Icon} {imgSrcWhenEmpty} />
                    {:else if isLoading}
                        <Spinner role="status" />
                    {:else}
                        <TableItem
                            {title}
                            {records}
                            onError={(error) => (errorMessage = error)}
                            getColumnNames={tableHandlers.getColumnNames}
                            renderField={tableHandlers.renderField}
                        />
                    {/if}
                </Item.Content>
            </Item.Root>
        </Item.Group>
    </Card.Content>
</Card.Root>
