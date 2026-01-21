<script lang="ts">
  import { Picto, Room } from "../../../bindings/changeme";
  import { onMount } from "svelte";

  let { isHost = $bindable<boolean>(false) } = $props<{ isHost: boolean }>();
  let room: Room | null = $state(null);

  async function getCurrentRoom() {
    try {
      room = await Picto.GetCurrentroom();
    } catch (error) {
      console.log(error);
      return;
    }
  }

  // If Host Start MDNS (for discovery) and WS (for communication)
  async function startServers() {
    try {
      await Picto.StartServers(); // this starts MDNS AND WS servers
    } catch (error) {
      console.log(error);
      return;
    }
  }

  onMount(() => {
    getCurrentRoom();

    if (isHost) {
      startServers();
    }
  });
</script>

{#if room}
  <div
    class="bg-card border-border mr-3 ml-3 flex translate-y-3 flex-row justify-between rounded-md border px-12 py-2"
  >
    <div class="flex flex-row items-center gap-8">
      <div class="flex flex-col items-center">
        <p>Menu</p>
        <p>[ M ]</p>
      </div>

      <div class="flex flex-col items-center justify-center">
        <p>Clear Canvas</p>
        <p>[ Space ]</p>
      </div>
    </div>

    <div class="flex flex-row items-center gap-20">
      <p class="italic">Host: {room.hostname}</p>
      <p>Users: 100</p>
    </div>
  </div>

  <div
    class="bg-card border-border fixed bottom-3 left-3 z-50 flex items-center justify-end gap-10 rounded-md border px-5 py-2"
  >
    <svg class="h-15 w-15" viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="12" cy="12" r="10" class="fill-white hover:fill-blue-700" />
    </svg>
    <svg class="h-10 w-10" viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="12" cy="12" r="10" class="fill-white hover:fill-blue-700" />
    </svg>
    <svg class="h-5 w-5" viewBox="0 0 24 24" aria-hidden="true">
      <circle cx="12" cy="12" r="10" class="fill-white hover:fill-blue-700" />
    </svg>
  </div>
{/if}
