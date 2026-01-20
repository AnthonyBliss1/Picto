<script lang="ts">
  import { Button } from "$lib/components/ui/button/index.js";
  import MDNSTable from "$lib/components/mdns-table.svelte";
  import CanvasRoom from "$lib/components/canvas-room.svelte";
  import { Picto } from "../bindings/changeme";

  type RoomChoice = "create_room" | "join_room" | null;

  let roomChoice: RoomChoice = $state(null);

  async function onCreateRoom() {
    try {
      roomChoice = "create_room";
      await Picto.SetIsHost(true);

      await Picto.SetCurrentRoom(null, true);
    } catch (error) {
      console.log(error);
    }
  }
</script>

{#if !roomChoice}
  <div class="flex min-h-screen -translate-y-12 flex-col items-center justify-center gap-10">
    <img src="/Picto-Svelte.svg" width="200" height="150" alt="Picto-Svelte logo" />
    <h1 class="text-5xl">Welcome to Picto</h1>

    <div class="flex flex-row gap-10">
      <Button variant="outline" class="p-7 text-2xl" onclick={onCreateRoom}>Create Room</Button>
      <Button
        variant="outline"
        class="p-7 text-2xl"
        onclick={() => {
          roomChoice = "join_room";
        }}>Join Room</Button
      >
    </div>
  </div>
{:else if roomChoice === "create_room"}
  <CanvasRoom />
{:else if roomChoice === "join_room"}
  <MDNSTable />
{/if}
