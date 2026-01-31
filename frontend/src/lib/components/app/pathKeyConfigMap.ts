import { cmd } from "$wailsjs/go/models";
import { FileSpreadsheetIcon, SheetIcon } from "@lucide/svelte";
import { type Component } from "svelte";
import orderExample from "$assets/order_example.png";
import rosterExample from "$assets/roster_example.png";

export const pathKeyConfigMap: Partial<Record<cmd.PathKey, {
    extension: string;
    Icon: Component;
    imgSrcWhenEmpty: string;
}>> = {
    [cmd.PathKey.ORDER_XLSX]: {
        extension: "xlsx",
        Icon: FileSpreadsheetIcon,
        imgSrcWhenEmpty: orderExample,
    },
    [cmd.PathKey.NAMES_CSV]: {
        extension: "csv",
        Icon: SheetIcon,
        imgSrcWhenEmpty: rosterExample,
    }
};