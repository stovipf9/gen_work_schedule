import type { Component } from "svelte";

export type Page = {
    id: `${string}-${string}-${string}-${string}-${string}`
    label: string;
    // component: any | null; // TODO: Replace 'any' with more specific type if possible
    component?: Component<any, any, any>;
    isSuccess?: boolean;
};

export type PageGroup = {
    label: string;
    pages: Page[];
};