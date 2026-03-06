// CLASSIFICATION: UNCLASSIFIED
// src/App.tsx — Root application component

import { createSignal, onMount, Show } from "solid-js";
import { checkCapabilities, type Capabilities } from "./services/capabilities";
import { renderDegradedNotice } from "./components/DegradedNotice";
import { allocateSAB } from "./services/sab";

export default function App() {
  const [caps, setCaps] = createSignal<Capabilities | null>(null);
  const [ready, setReady] = createSignal(false);

  onMount(async () => {
    const detected = await checkCapabilities();
    setCaps(detected);

    const allOk =
      detected.webgpu &&
      detected.sharedArrayBuffer &&
      detected.offscreenCanvas &&
      detected.webtransport;

    if (!allOk) {
      return;
    }

    // Allocate the SharedArrayBuffer ring buffer
    allocateSAB();
    setReady(true);
  });

  return (
    <Show
      when={caps() !== null}
      fallback={<div style={{ padding: "2rem" }}>Initialising…</div>}
    >
      <Show
        when={
          caps()!.webgpu &&
          caps()!.sharedArrayBuffer &&
          caps()!.offscreenCanvas &&
          caps()!.webtransport
        }
        fallback={renderDegradedNotice(caps()!)}
      >
        <Show
          when={ready()}
          fallback={<div style={{ padding: "2rem" }}>Loading…</div>}
        >
          <canvas id="gpu-canvas" style={{ width: "100%", height: "100%" }} />
        </Show>
      </Show>
    </Show>
  );
}
