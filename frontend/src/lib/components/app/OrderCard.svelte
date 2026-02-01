<script lang="ts">
    import { GetOrderColumnNames, LoadOrders } from "$wailsjs/go/main/WailsApp";
    import { cmd } from "$wailsjs/go/models";
    import BaseTableCard from "./BaseTableCard.svelte";

    interface Props {
        title: string;
        setCardCompleted: (result: boolean) => void;
    }
    let { title, setCardCompleted }: Props = $props();

    // フィールドのレンダリング関数
    function renderField(record: cmd.Order, field: keyof cmd.Order): string {
        // 日時については
        // ISO8601の文字列YYYY-MM-DDTHH:mm:ssZを
        // YYYY-MM-DDに
        if (field === "startDate" || field === "endDate") {
            return new Date(record[field] as string)
                .toISOString()
                .split("T")[0];
        }
        // それ以外のフィールドはそのまま返す
        return String(record[field]);
    }
</script>

<BaseTableCard
    {title}
    pathKey={cmd.PathKey.ORDER_XLSX}
    {setCardCompleted}
    tableHandlers={{
        getColumnNames: GetOrderColumnNames,
        loadData: LoadOrders,
        renderField: renderField,
    }}
/>
