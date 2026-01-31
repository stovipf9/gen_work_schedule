<script lang="ts">
    import {
        GetRosterColumnNames,
        LoadRosterRecords,
    } from "$wailsjs/go/main/WailsApp";
    import { cmd } from "$wailsjs/go/models";
    import BaseTableCard from "./BaseTableCard.svelte";

    interface Props {
        title: string;
        setCardCompleted: (result: boolean) => void;
    }

    let { title, setCardCompleted }: Props = $props();

    // フィールドのレンダリング関数
    function renderField(
        record: Record<string, string>,
        field: string,
    ): string {
        return record[field];
    }

    const WrappedLoadRosterRecords = (
        path: string,
    ): Promise<
        {
            familyName: string;
            fullName: string;
        }[]
    > =>
        LoadRosterRecords(path).then((r) =>
            Object.entries(r).map(([familyName, fullName]) => ({
                familyName,
                fullName,
            })),
        );
</script>

<BaseTableCard
    {title}
    pathKey={cmd.PathKey.NAMES_CSV}
    {setCardCompleted}
    tableHandlers={{
        getColumnNames: GetRosterColumnNames,
        loadData: WrappedLoadRosterRecords,
        renderField: renderField,
    }}
/>
