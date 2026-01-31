<script lang="ts">
    import { Button } from "$lib/components/ui/button";
    import { Checkbox } from "$lib/components/ui/checkbox";
    import * as Kbd from "$lib/components/ui/kbd";
    import * as Sidebar from "$lib/components/ui/sidebar";
    import { Quit } from "$wailsjs/runtime/runtime";
    import type { Page, PageGroup } from "./types";

    interface Props {
        pageGroups: Record<string, PageGroup>;
        selected: Page;
        onSelected: (page: Page) => void;
    }
    let { pageGroups, selected, onSelected }: Props = $props();
</script>

<Sidebar.Root collapsible="none" class="h-screen w-48 min-w-48">
    <Sidebar.Content>
        {#each Object.values(pageGroups) as pageGroup (pageGroup.label)}
            <Sidebar.Group>
                <Sidebar.GroupLabel>{pageGroup.label}</Sidebar.GroupLabel>
                <Sidebar.GroupContent>
                    <Sidebar.Menu>
                        {#each pageGroup.pages as page}
                            <Sidebar.MenuItem>
                                <Sidebar.MenuButton
                                    onclick={() => onSelected(page)}
                                    isActive={selected.id === page.id}
                                    class="
                                    data-[active=true]:bg-foreground
                                    data-[active=true]:text-background
                                    "
                                >
                                    {page.label}
                                </Sidebar.MenuButton>
                                <Sidebar.MenuBadge>
                                    {#if page.isSuccess}
                                        <Checkbox
                                            checked={page.isSuccess}
                                            data-testid={`success-checkbox-${page.label}`}
                                        />
                                    {/if}
                                </Sidebar.MenuBadge>
                            </Sidebar.MenuItem>
                        {/each}
                    </Sidebar.Menu>
                </Sidebar.GroupContent>
            </Sidebar.Group>
            <Sidebar.Separator decorative={true} />
        {/each}
    </Sidebar.Content>
    <Sidebar.Footer>
        <Sidebar.Menu>
            <Sidebar.MenuItem>
                <Button onclick={Quit}>
                    終了
                    <Kbd.Group>
                        <Kbd.Root>Alt</Kbd.Root>
                        <Kbd.Root>F4</Kbd.Root>
                    </Kbd.Group>
                </Button>
            </Sidebar.MenuItem>
        </Sidebar.Menu>
    </Sidebar.Footer>
</Sidebar.Root>
