<script lang="ts">
  import { closeSession, type Session } from "$lib/picto-sessions";
  import { Message, Picto, Point, Room } from "../../../bindings/changeme";
  import { onDestroy, onMount } from "svelte";

  let { session = $bindable<Session>() } = $props<{ session: Session }>();

  const Colors = {
    white: { name: "white", hex: "#FFFFFF" },
    red: { name: "red", hex: "#B91C1C" },
    green: { name: "green", hex: "#15803D" },
  };

  let room: Room | null = $state(null);
  let numClients: number = $state(0);

  let canvasEl: HTMLCanvasElement | null = $state(null);
  let remoteCanvas: HTMLCanvasElement | null = $state(null);

  let ctx: CanvasRenderingContext2D | null = $state(null);
  let remoteCtx: CanvasRenderingContext2D | null = $state(null);

  let cleanupCanvas: (() => void) | undefined;
  let didInit: boolean = $state(false);

  let isDrawing = $state(false);

  let color = $state(Colors.white);

  let lineWidth = $state(10);

  let isEraser = $state(false);

  let pendingPoints: Point[] = [];
  let lastSentPoint: Point | null = null;
  let lastPoint: Point | null = null;

  let flushScheduled = false;
  const FLUSH_INTERVAL_MS = 16;
  let lastFlushAt = 0;

  const onKeyDown = (event: KeyboardEvent) => {
    if (event.key === "m") {
      console.log("M Key Pressed!");
      closeSession(session);
    }

    if (event.key === " ") {
      console.log("Space Key Pressed!");

      const msg: Message = {
        action: "clear",
        phase: "",
        points: [],
        strokeWidth: lineWidth,
        color: color.hex,
        numClients: 0,
      };

      wsSend(msg);
    }
  };

  async function getCurrentRoom() {
    try {
      room = await Picto.GetCurrentroom();
    } catch (error) {
      console.error(error);
      return;
    }
  }

  session.websocket.addEventListener("message", (event: MessageEvent) => {
    console.log("Message Received: ", event);

    const message = JSON.parse(event.data) as Message;

    if (!message) {
      console.error("Failed to parse JSON, incorrect shape");
    }

    switch (message.action) {
      case "draw":
        handleRemoteDrawing(message);
        break;

      case "clear":
        clearCanvas();
        break;

      case "new-connection":
        numClients = message.numClients;
        break;

      case "closed-connection":
        numClients = message.numClients;
        break;

      case "server-shutdown":
        closeSession(session);
        break;

      default:
        return;
    }
  });

  function handleRemoteDrawing(message: Message) {
    if (!remoteCtx) return;

    applyRemoteStrokeStyle(message);

    const pts = message.points;
    if (!pts || pts.length === 0) return;

    const phase = message.phase ?? "move";

    if (phase === "start") {
      // anchor
      lastPoint = pts[0];

      // draw a dot so taps appear
      remoteCtx.beginPath();
      remoteCtx.moveTo(pts[0].x, pts[0].y);
      remoteCtx.lineTo(pts[0].x + 0.01, pts[0].y + 0.01);
      remoteCtx.stroke();
      remoteCtx.closePath();
      return;
    }

    if (phase === "move") {
      if (!lastPoint) lastPoint = pts[0];

      remoteCtx.beginPath();
      remoteCtx.moveTo(lastPoint.x, lastPoint.y);

      for (const p of pts) remoteCtx.lineTo(p.x, p.y);

      remoteCtx.stroke();
      remoteCtx.closePath();

      // update anchor to last point in this batch
      lastPoint = pts[pts.length - 1];
      return;
    }

    if (phase === "end") {
      // draw final segment if there is an anchor
      if (lastPoint) {
        remoteCtx.beginPath();
        remoteCtx.moveTo(lastPoint.x, lastPoint.y);
        for (const p of pts) remoteCtx.lineTo(p.x, p.y);
        remoteCtx.stroke();
        remoteCtx.closePath();
      }

      // clear anchor for next line
      lastPoint = null;
      return;
    }
  }

  function applyRemoteStrokeStyle(message: Message) {
    if (!remoteCtx) return;

    remoteCtx.lineJoin = "round";
    remoteCtx.lineCap = "round";
    remoteCtx.lineWidth = message.strokeWidth;
    remoteCtx.strokeStyle = message.color;
  }

  function applyStrokeStyle(ctx: CanvasRenderingContext2D | null) {
    if (!ctx) return;
    ctx.lineJoin = "round";
    ctx.lineCap = "round";
    ctx.lineWidth = lineWidth;
    ctx.strokeStyle = color.hex;
    ctx.globalCompositeOperation = isEraser ? "destination-out" : "source-over";
  }

  function wsSend(msg: Message) {
    if (!session.websocket || session.websocket.readyState !== WebSocket.OPEN) return;
    session.websocket.send(JSON.stringify(msg));
  }

  function queuePoint(p: Point) {
    pendingPoints.push(p);
    scheduleFlush();
  }

  function scheduleFlush() {
    if (flushScheduled) return;
    flushScheduled = true;

    requestAnimationFrame(() => {
      flushScheduled = false;
      flushMoveBatch();
    });
  }

  function flushMoveBatch(force = false) {
    const now = performance.now();
    if (!force && now - lastFlushAt < FLUSH_INTERVAL_MS) {
      scheduleFlush();
      return;
    }

    if (pendingPoints.length === 0) return;

    const pts = pendingPoints;
    pendingPoints = [];
    lastFlushAt = now;
    lastSentPoint = pts[pts.length - 1];

    const msg: Message = {
      action: "draw",
      phase: "move",
      points: pts,
      strokeWidth: lineWidth,
      color: color.hex,
      numClients: 0,
    };

    wsSend(msg);
  }

  function getPoint(e: PointerEvent) {
    if (!canvasEl) return { x: 0, y: 0 };
    const rect = canvasEl.getBoundingClientRect();
    return { x: e.clientX - rect.left, y: e.clientY - rect.top };
  }

  function startDraw(e: PointerEvent) {
    if (!ctx || !canvasEl) {
      console.error("No canvasEL or ctx...");
      return;
    }

    canvasEl.setPointerCapture(e.pointerId);
    isDrawing = true;
    console.log("Drawing...");

    pendingPoints = [];
    lastSentPoint = null;

    const p = getPoint(e);
    lastPoint = p;

    ctx.beginPath();
    ctx.moveTo(p.x, p.y);

    const msg: Message = {
      action: "draw",
      phase: "start",
      points: [p],
      strokeWidth: lineWidth,
      color: color.hex,
      numClients: 0,
    };

    wsSend(msg);
  }

  function moveDraw(e: PointerEvent) {
    if (!ctx || !isDrawing) return;

    const p = getPoint(e);
    ctx.lineTo(p.x, p.y);
    ctx.stroke();

    queuePoint(p);
  }

  function endDraw(e: PointerEvent) {
    if (!ctx || !canvasEl) return;

    isDrawing = false;

    try {
      canvasEl.releasePointerCapture(e.pointerId);
    } catch {}

    flushMoveBatch(true);

    const p = getPoint(e);

    const msg: Message = {
      action: "draw",
      phase: "end",
      points: [p],
      strokeWidth: lineWidth,
      color: color.hex,
      numClients: 0,
    };

    wsSend(msg);

    pendingPoints = [];
    lastSentPoint = null;
    lastPoint = null;

    ctx.closePath();
  }

  function clearCanvas() {
    if (!ctx || !canvasEl || !remoteCtx || !remoteCanvas) return;
    const { width, height } = canvasEl.getBoundingClientRect();
    ctx.clearRect(0, 0, width, height);
    remoteCtx.clearRect(0, 0, width, height);
  }

  function setupCanvas() {
    if (!canvasEl || !remoteCanvas) {
      console.error("No canvases... can't setupCanvas");
      return;
    }

    ctx = canvasEl.getContext("2d");
    remoteCtx = remoteCanvas.getContext("2d");

    if (!ctx || !remoteCtx) return;

    didInit = true;

    const resize = () => {
      if (!canvasEl || !ctx || !remoteCanvas || !remoteCtx) return;

      const rect = canvasEl.getBoundingClientRect();
      const dpr = window.devicePixelRatio || 1;

      // Set drawing buffer size (physical pixels)
      canvasEl.width = Math.max(1, Math.floor(rect.width * dpr));
      canvasEl.height = Math.max(1, Math.floor(rect.height * dpr));

      remoteCanvas.width = Math.max(1, Math.floor(rect.width * dpr));
      remoteCanvas.height = Math.max(1, Math.floor(rect.height * dpr));

      // Draw using CSS pixel coordinates
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      remoteCtx.setTransform(dpr, 0, 0, dpr, 0, 0);

      applyStrokeStyle(ctx);
    };

    const ro = new ResizeObserver(resize);
    ro.observe(canvasEl);

    resize();

    return () => ro.disconnect();
  }

  $effect(() => {
    if (!canvasEl || didInit) return;

    console.log("Running cleanup...");

    cleanupCanvas?.();
    cleanupCanvas = setupCanvas();
  });

  $effect(() => {
    applyStrokeStyle(ctx);
  });

  onMount(() => {
    getCurrentRoom();
    document.addEventListener("keydown", onKeyDown);
  });

  onDestroy(() => {
    document.removeEventListener("keydown", onKeyDown);
  });
</script>

{#if room}
  <div class="bg-background relative h-full w-full">
    <!--> Canvas Object to handle drawing <--->
    <div class="absolute inset-0">
      <canvas
        class="pointer-events-none absolute inset-0 z-0 h-full w-full"
        bind:this={remoteCanvas}
      ></canvas>
      <canvas
        class="absolute inset-0 z-10 h-full w-full bg-transparent"
        bind:this={canvasEl}
        onpointerdown={startDraw}
        onpointermove={moveDraw}
        onpointerup={endDraw}
        onpointercancel={endDraw}
        onpointerleave={endDraw}
      ></canvas>
    </div>

    <!--> Top Toolbar <--->
    <div
      class="bg-card border-border relative z-50 mr-3 ml-3 flex translate-y-3 flex-row justify-between rounded-md border px-12 py-2"
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
        <p class="truncate italic">Host: {room.hostname}</p>
        <p>Users: {numClients}</p>
      </div>
    </div>

    <!--> Bottom Draw Size Selector <--->
    <div class="fixed bottom-3 left-3 z-50 flex flex-row justify-between gap-10">
      <div
        class="bg-card border-border flex items-center justify-end gap-10 rounded-md border px-5 py-2"
      >
        <svg
          class="h-15 w-15"
          viewBox="0 0 24 24"
          aria-hidden="true"
          onclick={() => {
            lineWidth = 15;
          }}
        >
          <circle
            cx="12"
            cy="12"
            r="10"
            class={`hover:fill-blue-700 ${lineWidth === 15 ? `fill-blue-700` : `fill-white`}`}
          />
        </svg>
        <svg
          class="h-10 w-10"
          viewBox="0 0 24 24"
          aria-hidden="true"
          onclick={() => {
            lineWidth = 10;
          }}
        >
          <circle
            cx="12"
            cy="12"
            r="10"
            class={`hover:fill-blue-700 ${lineWidth === 10 ? `fill-blue-700` : `fill-white`}`}
          />
        </svg>
        <svg
          class="h-5 w-5"
          viewBox="0 0 24 24"
          aria-hidden="true"
          onclick={() => {
            lineWidth = 5;
          }}
        >
          <circle
            cx="12"
            cy="12"
            r="10"
            class={`hover:fill-blue-700 ${lineWidth === 5 ? `fill-blue-700` : `fill-white`}`}
          />
        </svg>
      </div>
      <div
        class="bg-card border-border flex items-center justify-end gap-10 rounded-md border px-5 py-2"
      >
        <div
          class={`h-10 w-10 rounded-md border-3 ${color.name === "white" ? `border-blue-700` : `border-white`} bg-white hover:border-blue-700`}
          aria-hidden="true"
          onclick={() => {
            color = Colors.white;
          }}
        ></div>
        <div
          class={`h-10 w-10 rounded-md border-3 ${color.name === "red" ? `border-blue-700` : `border-red-700`} bg-red-700 hover:border-blue-700`}
          aria-hidden="true"
          onclick={() => {
            color = Colors.red;
          }}
        ></div>
        <div
          class={`h-10 w-10 rounded-md border-3 ${color.name === "green" ? `border-blue-700` : `border-green-700`} bg-green-700 hover:border-blue-700`}
          aria-hidden="true"
          onclick={() => {
            color = Colors.green;
          }}
        ></div>
      </div>
    </div>
  </div>
{/if}
