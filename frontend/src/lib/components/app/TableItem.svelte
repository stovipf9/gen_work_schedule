<script lang="ts" generics="Row extends object">
    import * as Table from "$lib/components/ui/table";
    import { main } from "$wailsjs/go/models";
    import { onMount } from "svelte";

    interface Props {
        title: string;
        records: Row[];
        onError: (error: string) => void;
        getColumnNames: () => Promise<main.TableColumn[]>;
        renderField: (record: Row, field: keyof Row) => string;
    }

    let { title, records, onError, getColumnNames, renderField }: Props =
        $props();

    let columnNames: main.TableColumn[] = $state([]);
    onMount(() => {
        // 列名を取得
        getColumnNames()
            .then((names) => {
                columnNames = names;
            })
            .catch((error) => {
                onError("カラム名取得エラー: " + error);
            });
    });

    function renderFieldWrapper(record: Row, field: string): string {
        return field in record
            ? renderField(record, field as keyof Row)
            : "N/A";
    }
</script>

<Table.Root>
    <Table.Caption>{title}</Table.Caption>
    <Table.Header>
        <Table.Row>
            {#each columnNames as col}
                <Table.Head>
                    {col.label || col.field}
                </Table.Head>
            {/each}
        </Table.Row>
    </Table.Header>
    <Table.Body>
        {#each records as record}
            <Table.Row>
                {#each columnNames as col}
                    <Table.Cell>
                        {renderFieldWrapper(record, col.field)}
                    </Table.Cell>
                {/each}
            </Table.Row>
        {/each}
    </Table.Body>
</Table.Root>
