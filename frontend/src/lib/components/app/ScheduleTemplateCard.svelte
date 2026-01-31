<script lang="ts">
    import * as Card from "$lib/components/ui/card";
    import * as Item from "$lib/components/ui/item";
    import { OpenFile, SelectFile } from "$wailsjs/go/main/WailsApp";
    import { cmd } from "$wailsjs/go/models";
    import { TableIcon } from "@lucide/svelte";
    import CardDescriptionOnError from "./CardDescriptionOnError.svelte";
    import FileSelectorItem from "./FileSelectorItem.svelte";
    import { paths } from "./paths.svelte";

    const EXTENSION = "xlsx";
    const PathKey = cmd.PathKey.TEMPLATE_XLSX;

    interface Props {
        title: string;
        setCardCompleted: (result: boolean) => void;
    }
    let { title, setCardCompleted }: Props = $props();

    let errorMessage = $state("");

    $effect(() => {
        setCardCompleted(!!paths.value?.[PathKey]);
    });

    function onClickSelect() {
        SelectFile(EXTENSION)
            .then((newPath) => {
                paths.update({ [PathKey]: newPath });
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

<Card.Root>
    <Card.Header>
        <Card.Title {title}>{title}</Card.Title>
        <CardDescriptionOnError {errorMessage} />
    </Card.Header>
    <Card.Content>
        <Item.Group>
            <FileSelectorItem
                header={title + "ファイル"}
                placeholder={`${title}(${EXTENSION})を選択。記入済みも可。`}
                path={paths.value[cmd.PathKey.TEMPLATE_XLSX]}
                Icon={TableIcon}
                {onClickSelect}
                {onClickOpen}
            />
        </Item.Group>
    </Card.Content>
</Card.Root>
