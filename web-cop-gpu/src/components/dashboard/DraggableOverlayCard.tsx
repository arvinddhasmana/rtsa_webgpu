// CLASSIFICATION: UNCLASSIFIED
// src/components/dashboard/DraggableOverlayCard.tsx — Draggable, minimisable glassmorphic overlay card
//
// Reference: docs/business/usecases/UC017_sensor_health_monitoring.md

import { createEffect, createSignal, JSX, onCleanup, Show } from "solid-js";

export interface DraggableOverlayCardProps {
  title: string;
  icon?: JSX.Element;
  initialX?: number;
  initialY?: number;
  position?: { x: number; y: number };
  width?: string;
  minWidth?: string;
  maxHeight?: string;
  children: JSX.Element;
  zIndex?: number;
  accentColor?: string;
  /** Unique key for CSS scrollbar scoping */
  scrollKey?: string;
  onPositionChange?: (pos: { x: number; y: number }) => void;
  onClose?: () => void;
  constrainToParent?: boolean;
}

/**
 * Floating, draggable, minimisable glassmorphic overlay card.
 * Intended for use over a map or canvas area.
 * Never destructure props — breaks SolidJS reactivity.
 */
export function DraggableOverlayCard(
  props: DraggableOverlayCardProps,
): JSX.Element {
  const [minimized, setMinimized] = createSignal(false);
  const [pos, setPos] = createSignal({
    x: props.initialX ?? 16,
    y: props.initialY ?? 16,
  });
  const [dragging, setDragging] = createSignal(false);
  const [dragOffset, setDragOffset] = createSignal({ dx: 0, dy: 0 });
  let rootEl: HTMLDivElement | undefined;

  const color = () => props.accentColor ?? "rgba(59,130,246,0.55)";
  const scrollKey = props.scrollKey ?? "overlay";

  function onMouseDown(e: MouseEvent) {
    const target = e.target as HTMLElement;
    if (target.tagName === "BUTTON" || target.closest("button")) return;
    setDragging(true);
    setDragOffset({ dx: e.clientX - pos().x, dy: e.clientY - pos().y });
    e.preventDefault();
  }

  // Sync externally provided position (used by auto-arrange) without disrupting active drag
  createEffect(() => {
    const external = props.position;
    if (external && !dragging()) {
      setPos(external);
    }
  });

  createEffect(() => {
    const handleMouseMove = (e: MouseEvent) => {
      if (!dragging()) return;

      let next = {
        x: e.clientX - dragOffset().dx,
        y: e.clientY - dragOffset().dy,
      };

      if (props.constrainToParent && rootEl?.parentElement) {
        const parent = rootEl.parentElement;
        const parentW = parent.clientWidth;
        const parentH = parent.clientHeight;
        const cardW = rootEl.offsetWidth;
        const cardH = rootEl.offsetHeight;
        const pad = 8;
        next = {
          x: Math.min(Math.max(next.x, pad), Math.max(pad, parentW - cardW - pad)),
          y: Math.min(Math.max(next.y, pad), Math.max(pad, parentH - cardH - pad)),
        };
      }

      setPos(next);
      props.onPositionChange?.(next);
    };

    const handleMouseUp = () => setDragging(false);

    window.addEventListener("mousemove", handleMouseMove);
    window.addEventListener("mouseup", handleMouseUp);

    onCleanup(() => {
      window.removeEventListener("mousemove", handleMouseMove);
      window.removeEventListener("mouseup", handleMouseUp);
    });
  });

  return (
    <div
      data-testid={`overlay-card-${props.title.toLowerCase().replace(/\s+/g, "-")}`}
      ref={rootEl}
      style={{
        position: "absolute",
        left: `${pos().x}px`,
        top: `${pos().y}px`,
        width: props.width ?? "220px",
        "min-width": props.minWidth ?? "180px",
        "z-index": props.zIndex ?? 10,
        "user-select": dragging() ? "none" : "auto",
        "pointer-events": "auto",
      }}
    >
      <div
        style={{
          background:
            "linear-gradient(180deg, rgba(8, 14, 26, 0.92), rgba(6, 11, 22, 0.9))",
          "backdrop-filter": "blur(28px)",
          "-webkit-backdrop-filter": "blur(28px)",
          border: `1px solid ${color()}`,
          "border-radius": "10px",
          "box-shadow": `0 14px 48px rgba(0,0,0,0.68), inset 0 1px 0 rgba(255,255,255,0.07)`,
          overflow: "hidden",
          transition: "box-shadow 0.2s ease",
        }}
      >
        {/* ── Header / drag handle ── */}
        <div
          onMouseDown={onMouseDown}
          style={{
            display: "flex",
            "align-items": "center",
            gap: "6px",
            padding: "7px 8px 7px 10px",
            background:
              "linear-gradient(90deg, rgba(59,130,246,0.09), rgba(255,255,255,0.025))",
            "border-bottom": minimized()
              ? "none"
              : "1px solid rgba(255,255,255,0.07)",
            cursor: dragging() ? "grabbing" : "grab",
            "border-radius": minimized() ? "10px" : "10px 10px 0 0",
          }}
        >
          <Show when={props.icon}>
            <div
              style={{ color: "#64748b", "line-height": 0, "flex-shrink": 0 }}
            >
              {props.icon}
            </div>
          </Show>

          <span
            style={{
              flex: 1,
              "font-size": "clamp(0.6rem, 0.55rem + 0.12vw, 0.72rem)",
              "font-weight": "700",
              "text-transform": "uppercase",
              "letter-spacing": "0.1em",
              color: "#cbd5e1",
              "font-family": "monospace",
              "white-space": "nowrap",
              overflow: "hidden",
              "text-overflow": "ellipsis",
            }}
          >
            {props.title}
          </span>

          {/* Drag indicator */}
          <div
            style={{
              display: "flex",
              "flex-direction": "column",
              gap: "2px",
              "flex-shrink": 0,
              opacity: 0.3,
              "pointer-events": "none",
            }}
          >
            <div
              style={{ width: "12px", height: "1px", background: "#94a3b8" }}
            />
            <div
              style={{ width: "12px", height: "1px", background: "#94a3b8" }}
            />
            <div
              style={{ width: "12px", height: "1px", background: "#94a3b8" }}
            />
          </div>

          <div style={{ display: "flex", "align-items": "center", gap: "6px" }}>
            {/* Minimize / restore button */}
            <button
              data-testid={`overlay-minimize-${props.title.toLowerCase().replace(/\s+/g, "-")}`}
              onClick={() => setMinimized((m) => !m)}
              title={minimized() ? "Restore panel" : "Minimize panel"}
              style={{
                background: minimized()
                  ? "rgba(59,130,246,0.15)"
                  : "rgba(255,255,255,0.04)",
                border: "1px solid rgba(255,255,255,0.1)",
                "border-radius": "5px",
                cursor: "pointer",
                color: minimized() ? "#60a5fa" : "#64748b",
                padding: "2px 6px",
                "font-size": "0.65rem",
                "line-height": 1.2,
                display: "flex",
                "align-items": "center",
                "flex-shrink": 0,
                transition: "all 0.15s ease",
              }}
            >
              {minimized() ? "▲" : "▼"}
            </button>

            <Show when={props.onClose}>
              <button
                data-testid={`overlay-close-${props.title.toLowerCase().replace(/\s+/g, "-")}`}
                title="Close panel"
                onClick={(e) => {
                  e.stopPropagation();
                  props.onClose?.();
                }}
                style={{
                  background: "rgba(255,255,255,0.04)",
                  border: "1px solid rgba(255,255,255,0.12)",
                  "border-radius": "5px",
                  cursor: "pointer",
                  color: "#f87171",
                  padding: "2px 6px",
                  "font-size": "0.65rem",
                  "line-height": 1.2,
                  display: "flex",
                  "align-items": "center",
                  "flex-shrink": 0,
                  transition: "all 0.15s ease",
                }}
              >
                ✕
              </button>
            </Show>
          </div>
        </div>

        {/* ── Body ── */}
        <Show when={!minimized()}>
          <div
            class={`overlay-scroll-${scrollKey}`}
            style={{
              "max-height": props.maxHeight ?? "280px",
              "overflow-y": "auto",
              "overflow-x": "hidden",
              padding: "8px",
            }}
          >
            {props.children}
          </div>
        </Show>
      </div>

      <style>{`
        .overlay-scroll-${scrollKey}::-webkit-scrollbar { width: 4px; }
        .overlay-scroll-${scrollKey}::-webkit-scrollbar-track { background: rgba(255,255,255,0.02); border-radius: 2px; }
        .overlay-scroll-${scrollKey}::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.12); border-radius: 2px; }
        .overlay-scroll-${scrollKey}::-webkit-scrollbar-thumb:hover { background: rgba(255,255,255,0.2); }
      `}</style>
    </div>
  );
}
