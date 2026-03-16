// CLASSIFICATION: UNCLASSIFIED
// src/components/search/SearchOverlay.tsx
//
// Ctrl+K search overlay allowing operators to search for tracks by ID or type.
// Results fetched via gRPC QueryService.QueryTracks, clicking a result selects
// the track and opens the detail panel.
// Reference: docs/implementation/v4/phase3_ui_interaction.md §3 U3-6

import { For, Show, createSignal, onCleanup, onMount } from "solid-js";
import { searchTracks } from "../../services/query";
import type { TrackDetail } from "../../signals/track";
import {
  setSelectedTrack,
  setTrackDetail,
  setTrackDetailError,
  setTrackDetailLoading,
} from "../../signals/track";
import { searchOpen, setSearchOpen } from "../../signals/viewport";

/** SearchOverlay activated by Ctrl+K. Never destructure props. */
export function SearchOverlay() {
  const [query, setQuery] = createSignal("");
  const [results, setResults] = createSignal<TrackDetail[]>([]);
  const [searching, setSearching] = createSignal(false);
  const [searchError, setSearchError] = createSignal<string | null>(null);
  let searchTimeout: ReturnType<typeof setTimeout> | null = null;
  let inputRef: HTMLInputElement | undefined;

  // Keyboard shortcut: Ctrl+K opens, Escape closes
  function handleKeyDown(e: KeyboardEvent) {
    if ((e.ctrlKey || e.metaKey) && e.key === "k") {
      e.preventDefault();
      setSearchOpen(true);
      setTimeout(() => inputRef?.focus(), 50);
    }
    if (e.key === "Escape" && searchOpen()) {
      close();
    }
  }

  onMount(() => {
    window.addEventListener("keydown", handleKeyDown);
  });

  onCleanup(() => {
    window.removeEventListener("keydown", handleKeyDown);
    if (searchTimeout) clearTimeout(searchTimeout);
  });

  function close() {
    setSearchOpen(false);
    setQuery("");
    setResults([]);
    setSearchError(null);
  }

  function handleInput(e: Event) {
    const value = (e.currentTarget as HTMLInputElement).value;
    setQuery(value);

    if (searchTimeout) clearTimeout(searchTimeout);

    if (value.trim().length < 2) {
      setResults([]);
      return;
    }

    searchTimeout = setTimeout(async () => {
      setSearching(true);
      setSearchError(null);
      try {
        const found = await searchTracks(value.trim());
        setResults(found);
      } catch (_err) {
        setSearchError("Search failed — check connection");
        setResults([]);
      } finally {
        setSearching(false);
      }
    }, 300);
  }

  function selectResult(track: TrackDetail) {
    setSelectedTrack({ trackIdHash: 0, x: 0, y: 0, source: "search" });
    setTrackDetailLoading(false);
    setTrackDetailError(null);
    setTrackDetail(track);
    close();
  }

  return (
    <Show when={searchOpen()}>
      {/* Backdrop */}
      <div
        data-testid="search-overlay"
        style={{
          position: "fixed",
          inset: "0",
          background: "rgba(0,0,0,0.5)",
          "z-index": "600",
          display: "flex",
          "align-items": "flex-start",
          "justify-content": "center",
          "padding-top": "8vh",
        }}
        onClick={close}
        aria-modal="true"
        role="dialog"
        aria-label="Track search"
      >
        <div
          style={{
            background: "#0d1424",
            border: "1px solid #2d3f56",
            "border-radius": "8px",
            width: "36rem",
            "max-width": "90vw",
            overflow: "hidden",
          }}
          onClick={(e) => e.stopPropagation()}
        >
          {/* Search input */}
          <div
            style={{ padding: "0.75rem", "border-bottom": "1px solid #1e2a3a" }}
          >
            <input
              ref={inputRef}
              type="text"
              placeholder="Search track ID…"
              value={query()}
              onInput={handleInput}
              style={{
                width: "100%",
                background: "transparent",
                border: "none",
                outline: "none",
                color: "#e2e8f0",
                "font-size": "0.9rem",
                "font-family": "monospace",
                "box-sizing": "border-box",
              }}
              aria-label="Track search input"
            />
          </div>

          {/* Results */}
          <div style={{ "max-height": "50vh", "overflow-y": "auto" }}>
            <Show when={searching()}>
              <div
                style={{
                  padding: "0.75rem",
                  color: "#94a3b8",
                  "font-size": "0.8rem",
                }}
              >
                Searching…
              </div>
            </Show>

            <Show when={searchError() !== null}>
              <div
                style={{
                  padding: "0.75rem",
                  color: "#ef4444",
                  "font-size": "0.8rem",
                }}
                role="alert"
              >
                {searchError()}
              </div>
            </Show>

            <Show
              when={
                !searching() &&
                results().length === 0 &&
                query().trim().length >= 2 &&
                !searchError()
              }
            >
              <div
                style={{
                  padding: "0.75rem",
                  color: "#64748b",
                  "font-size": "0.8rem",
                }}
              >
                No tracks found
              </div>
            </Show>

            <For each={results()}>
              {(track) => (
                <button
                  onClick={() => selectResult(track)}
                  style={{
                    display: "block",
                    width: "100%",
                    background: "none",
                    border: "none",
                    "border-bottom": "1px solid #1e2a3a",
                    color: "#e2e8f0",
                    padding: "0.6rem 0.75rem",
                    "text-align": "left",
                    cursor: "pointer",
                    "font-family": "monospace",
                    "font-size": "0.8rem",
                  }}
                  aria-label={`Select track ${track.trackId}`}
                >
                  <div
                    style={{ "font-weight": "bold", "margin-bottom": "0.2rem" }}
                  >
                    {track.trackId}
                  </div>
                  <div style={{ color: "#94a3b8", "font-size": "0.7rem" }}>
                    {track.entityType} · {track.hostileClass} · {track.status}
                  </div>
                </button>
              )}
            </For>
          </div>

          {/* Footer hint */}
          <div
            style={{
              padding: "0.4rem 0.75rem",
              "border-top": "1px solid #1e2a3a",
              "font-size": "0.65rem",
              color: "#64748b",
            }}
          >
            Press Esc to close
          </div>
        </div>
      </div>
    </Show>
  );
}
