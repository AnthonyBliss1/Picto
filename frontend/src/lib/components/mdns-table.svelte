<script lang="ts">
  import * as Empty from "$lib/components/ui/empty/index.js";
  import { Button } from "$lib/components/ui/button/index.js";
  import { TriangleAlert } from "lucide-svelte";
  import { onMount } from "svelte";
  import Refresh from "@lucide/svelte/icons/refresh-ccw";
  import { Picto } from "../../../bindings/changeme";
  import { Room } from "../../../bindings/changeme/models";
  import { closeSession, type Session } from "$lib/picto-sessions";

  let { session = $bindable<Session>() } = $props<{ session: Session }>();

  let availRooms: Room[] = $state([]);

  async function getRooms() {
    try {
      console.log("Searching for rooms...");

      await Picto.MDNSLookup();
      availRooms = await Picto.GetAvailableRooms();
    } catch (error) {
      console.error(error);
      return;
    }
  }

  async function onJoinRoom(room: Room) {
    try {
      const ok = await Picto.SetCurrentRoom(room, false);

      if (ok) {
        session.hasRoom = true;
      }
    } catch (error) {
      console.error(error);
    }
  }

  onMount(() => {
    getRooms();
  });
</script>

<div class="bg-background grid min-h-screen place-items-center p-6">
  <div class="border-border w-full max-w-[420px] rounded-md border p-8">
    <div class="relative">
      <h1 class="text-center text-2xl font-semibold underline underline-offset-8">
        Select a room...
      </h1>

      <Button
        class="absolute top-1/2 right-0 -translate-y-1/2"
        size="icon"
        variant="ghost"
        onclick={getRooms}
      >
        <Refresh class="h-5 w-5" />
      </Button>
    </div>

    <div class="mt-6">
      {#if availRooms.length === 0}
        <Empty.Root class="border-border rounded-xl border border-dashed p-6">
          <Empty.Header>
            <Empty.Media variant="icon">
              <TriangleAlert />
            </Empty.Media>
            <Empty.Title>No Rooms Found</Empty.Title>
            <Empty.Description>Get some friends... to start a room!</Empty.Description>
          </Empty.Header>
        </Empty.Root>
      {:else}
        <div class="space-y-3">
          {#each availRooms as room (room)}
            <div
              class="border-border flex items-center justify-between rounded-2xl border px-4 py-3"
            >
              <div class="min-w-0">
                <div class="truncate font-medium">{room.hostname}</div>
              </div>

              <Button class="h-10 px-4" variant="outline" onclick={void onJoinRoom(room)}
                >Join</Button
              >
            </div>
          {/each}
        </div>
      {/if}
    </div>
  </div>
  <Button
    class="p-5 text-lg"
    variant="outline"
    onclick={() => {
      closeSession(session);
    }}>Back</Button
  >
</div>
