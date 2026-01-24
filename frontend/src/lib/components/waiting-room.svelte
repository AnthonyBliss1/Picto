<script lang="ts">
  import { Button } from "$lib/components/ui/button/index.js";
  import { onMount } from "svelte";
  import { Picto } from "../../../bindings/changeme";
  import { closeSession, type Session } from "$lib/picto-sessions";

  let { session = $bindable<Session>() } = $props<{ session: Session }>();

  async function getCurrentRoom() {
    try {
      session.room = await Picto.GetCurrentroom();
    } catch (error) {
      console.error(error);
      return;
    }
  }

  function establishWSConn() {
    if (session.room) {
      console.log("Establishing Connection to WebSocket...");

      session.websocket = new WebSocket(session.room.url);

      session.websocket.addEventListener("open", (event: Event) => {
        console.log("WebSocket Connection Established: ", event);
        session.connected = true;
      });

      session.websocket.addEventListener("error", (event: Event) => {
        console.log("Websocket Error: ", event);
      });

      session.websocket.addEventListener("close", (event: CloseEvent) => {
        console.log("WebSocket Connection Closed: ", event);
        closeSession(session);
      });
    } else {
      console.error("No room to connect to");
    }
  }

  onMount(async () => {
    console.log("Getting Current Room...");
    await getCurrentRoom();

    establishWSConn();
  });
</script>

<div class="flex min-h-screen flex-col items-center justify-center gap-20">
  <p class="text-2xl">Loading room...</p>
  <Button
    class="h-10 px-4"
    variant="outline"
    onclick={() => {
      closeSession(session);
    }}>Cancel</Button
  >
</div>
