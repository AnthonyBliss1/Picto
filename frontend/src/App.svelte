<script lang="ts">
  import { Button } from "$lib/components/ui/button/index.js";
  import MDNSTable from "$lib/components/mdns-table.svelte";
  import CanvasRoom from "$lib/components/canvas-room.svelte";
  import { Picto } from "../bindings/changeme";

  type RoomChoice = "create_room" | "join_room" | null;

  let roomChoice: RoomChoice = $state(null);
  let hasRoom: boolean = $state(false);
  let isHost: boolean = $state(false);

  async function onCreateRoom() {
    try {
      isHost = true;
      roomChoice = "create_room";

      await Picto.SetCurrentRoom(null, isHost);
      hasRoom = true;
    } catch (error) {
      console.log(error);
    }
  }
</script>

{#if !roomChoice}
  <div class="flex min-h-screen -translate-y-15 flex-col items-center justify-center gap-10">
    <img src="/Picto-Svelte.svg" width="200" height="150" alt="Picto-Svelte logo" />
    <h1 class="text-4xl">Welcome to Picto</h1>

    <div class="mt-8 flex flex-row gap-10">
      <Button variant="outline" class="p-6 text-lg" onclick={onCreateRoom}>Create Room</Button>
      <Button
        variant="outline"
        class="p-6 text-lg"
        onclick={() => {
          roomChoice = "join_room";
        }}>Join Room</Button
      >
    </div>
  </div>
{:else if roomChoice && hasRoom}
  <CanvasRoom bind:isHost />
{:else if roomChoice === "join_room" && !hasRoom}
  <MDNSTable bind:roomChoice />
{/if}
