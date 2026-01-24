<script lang="ts">
  import { Button } from "$lib/components/ui/button/index.js";
  import MDNSTable from "$lib/components/mdns-table.svelte";
  import WaitingRoom from "$lib/components/waiting-room.svelte";
  import CanvasRoom from "$lib/components/canvas-room.svelte";
  import { Picto } from "../bindings/changeme";
  import type { Session } from "$lib/picto-sessions";

  type Phase = "home" | "mdns" | "waiting" | "canvas";

  let session = $state<Session>({
    roomChoice: null,
    hasRoom: false,
    connected: false,
    isHost: false,
    websocket: null,
    room: null,
  });

  let phase = $derived.by((): Phase => {
    if (session.roomChoice === "join_room" && !session.hasRoom) return "mdns";
    if (session.hasRoom && !session.connected) return "waiting";
    if (session.hasRoom && session.connected) return "canvas";
    return "home";
  });

  async function onCreateRoom() {
    try {
      session.isHost = true;
      session.roomChoice = "create_room";

      const roomSet = await Picto.SetCurrentRoom(null, session.isHost);
      if (!roomSet) {
        throw new Error("Failed to set room");
      }
      const startedServers = await Picto.StartServers();
      if (!startedServers) {
        throw new Error("Failed to start Servers");
      }

      session.hasRoom = true;
    } catch (error) {
      console.error(error);
    }
  }
</script>

{#if phase === "home"}
  <div class="flex min-h-screen -translate-y-15 flex-col items-center justify-center gap-10">
    <img src="/Picto-Svelte.svg" width="200" height="150" alt="Picto-Svelte logo" />
    <h1 class="text-4xl">Welcome to Picto</h1>

    <div class="mt-8 flex flex-row gap-10">
      <Button variant="outline" class="p-6 text-lg" onclick={onCreateRoom}>Create Room</Button>
      <Button
        variant="outline"
        class="p-6 text-lg"
        onclick={() => {
          session.roomChoice = "join_room";
        }}>Join Room</Button
      >
    </div>
  </div>
{:else if phase === "mdns"}
  <MDNSTable bind:session />
{:else if phase === "waiting"}
  <WaitingRoom bind:session />
{:else if phase === "canvas"}
  <CanvasRoom {session} />
{/if}
