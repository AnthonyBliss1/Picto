<script lang="ts">
  import * as Empty from "$lib/components/ui/empty/index.js";
  import { TriangleAlert } from "lucide-svelte";
  import { onMount } from "svelte";

  import { Picto } from "../../../bindings/changeme";
  import { Room } from "../../../bindings/changeme/models";

  let availRooms: Room[] = $state([]);

  async function getRooms() {
    try {
      for (;;) {
        console.log("Searching for rooms...");

        await Picto.MDNSLookup();
        availRooms = await Picto.GetAvailableRooms();
      }
    } catch (error) {
      console.log(error);
      return;
    }
  }

  onMount(() => {
    getRooms();
  });
</script>

<div class="flex min-h-screen -translate-y-12 flex-col items-center justify-center gap-15">
  <h1 class="text-2xl">Select a room...</h1>

  <div class="flex flex-col">
    {#if availRooms.length === 0}
      <Empty.Root class="border border-dashed">
        <Empty.Header>
          <Empty.Media variant="icon">
            <TriangleAlert />
          </Empty.Media>
          <Empty.Title>No Rooms Found</Empty.Title>
          <Empty.Description>Get some friends... to start a room!</Empty.Description>
        </Empty.Header>
      </Empty.Root>
    {:else if availRooms.length > 0}
      {#each availRooms as room (room)}
        <div>{room.HostName}</div>
      {/each}
    {/if}
  </div>
</div>
