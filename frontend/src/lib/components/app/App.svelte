<script lang="ts">
	import * as Sidebar from "$lib/components/ui/sidebar";
	import { LoadCachedPaths } from "$wailsjs/go/main/WailsApp";
	import { LogError } from "$wailsjs/runtime/runtime";
	import { onMount, type Component } from "svelte";
	import { fade } from "svelte/transition";
	import AppSidebar from "./AppSidebar.svelte";
	import GenerateCard from "./GenerateCard.svelte";
	import OrderCard from "./OrderCard.svelte";
	import Pagination from "./Pagination.svelte";
	import { paths } from "./paths.svelte";
	import RosterCard from "./RosterCard.svelte";
	import ScheduleTemplateCard from "./ScheduleTemplateCard.svelte";
	import type { Page, PageGroup } from "./types";

	const createPage = (
		label: string,
		component?: Component<any, any, any>,
	): Page => ({
		id: crypto.randomUUID(),
		label: label,
		component: component,
	});

	let pageGroups: Record<string, PageGroup> = $state({
		inputs: {
			label: "入力",
			pages: [
				{
					label: "発注情報",
					component: OrderCard,
				},
				{
					label: "名簿",
					component: RosterCard,
				},
				{
					label: "勤務表テンプレート",
					component: ScheduleTemplateCard,
				},
			].map(({ label, component }) => createPage(label, component)),
		},
		outputs: {
			label: "出力",
			pages: [createPage("出力", GenerateCard)],
		},
	});
	let selected: Page = $state(pageGroups["inputs"].pages[0]);

	onMount(() => {
		LoadCachedPaths()
			.then((loaded_paths) => {
				paths.update(loaded_paths);
			})
			.catch((error) => {
				LogError("Failed to load cached paths: " + error);
			});
	});

	const AllPages = $derived(
		Object.values(pageGroups).flatMap((pageGroup) => {
			return Object.values(pageGroup.pages);
		}),
	);

	function selectPage(page: Page) {
		selected = page;
	}

	function setCardCompleted(result: boolean) {
		selected.isSuccess = result;
	}
</script>

<Sidebar.Provider>
	<AppSidebar {pageGroups} {selected} onSelected={selectPage} />
	<main class="flex flex-1 flex-col h-screen">
		<Pagination pages={AllPages} {selected} {selectPage}>
			<div style="overflow: hidden;">
				{#if selected.component}
					{#key selected.component}
						<div
							transition:fade={{ duration: 200 }}
							style="position: absolute;"
						>
							<selected.component
								title={selected.label}
								{setCardCompleted}
								inputPages={pageGroups["inputs"].pages}
							/>
						</div>
					{/key}
				{/if}
			</div>
		</Pagination>
	</main>
</Sidebar.Provider>
