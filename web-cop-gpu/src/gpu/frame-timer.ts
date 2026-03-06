// CLASSIFICATION: UNCLASSIFIED
// src/gpu/frame-timer.ts — GPU timestamp queries + JS frame timing
//
// Provides per-frame GPU timing via the WebGPU timestamp-query feature
// and JS wall-clock timing for Main Thread CPU measurement.
//
// Target metrics (docs/architecture/component_design.md §6):
//   Total frame time  ≤ 8 ms
//   SAB read + upload ≤ 2 ms
//   Compute passes    ≤ 1 ms
//   Render passes     ≤ 4 ms
//   Main thread CPU   < 20 %
//
// Reference: docs/implementation/v4/phase4_hardening_cutover.md H4-1

/** Labels for each timestamp-query slot (in order). */
export const PASS_LABELS = [
  "frame_start",
  "after_sab_upload",
  "after_interpolation",
  "after_culling",
  "after_background",
  "after_trails",
  "after_icons",
  "after_halos",
  "after_labels",
  "after_pick",
  "frame_end",
] as const;

export type PassLabel = (typeof PASS_LABELS)[number];
const NUM_TIMESTAMPS = PASS_LABELS.length;

/** Per-frame timing results (all values in milliseconds). */
export interface FrameTimings {
  /** Total GPU frame duration (frame_start → frame_end). */
  totalFrameMs: number;
  /** SAB read + writeBuffer duration. */
  sabUploadMs: number;
  /** Both compute passes combined. */
  computeMs: number;
  /** All render passes combined (background through pick). */
  renderMs: number;
  /** JS main-thread wall-clock time for buildFrame(). */
  mainThreadMs: number;
}

/** Rolling window size for averaging frame timings. */
const WINDOW = 60;

/**
 * FrameTimer — wraps GPU timestamp-query resolve + JS perf timing.
 *
 * Usage:
 *   const timer = new FrameTimer(device);
 *   // Each frame:
 *   const { querySet, resolveBuffer } = timer.beginFrame(encoder);
 *   // ... encode passes, writing timestamps via encoder.writeTimestamp()
 *   timer.endFrame(encoder, resolveBuffer);
 *   device.queue.submit([encoder.finish()]);
 *   await timer.resolveAsync();  // triggers readback on the next idle tick
 */
export class FrameTimer {
  private readonly supported: boolean;

  private querySet: GPUQuerySet | null = null;
  private resolveBuffer: GPUBuffer | null = null;
  private readbackBuffer: GPUBuffer | null = null;

  /** JS wall-clock start for the current frame. */
  private jsStart = 0;

  /** Rolling average of frame timings. */
  private window: FrameTimings[] = [];
  private windowIndex = 0;

  /** Smoothed average (updated after each frame). */
  smoothed: FrameTimings = {
    totalFrameMs: 0,
    sabUploadMs: 0,
    computeMs: 0,
    renderMs: 0,
    mainThreadMs: 0,
  };

  constructor(device: GPUDevice) {
    this.supported =
      device.features.has("timestamp-query") &&
      device.features.has("timestamp-query-inside-passes");

    if (this.supported) {
      this.querySet = device.createQuerySet({
        type: "timestamp",
        count: NUM_TIMESTAMPS,
      });

      // Buffer to resolve GPU timestamps into (one u64 per slot).
      this.resolveBuffer = device.createBuffer({
        label: "timestamp-resolve",
        size: NUM_TIMESTAMPS * 8,
        usage: GPUBufferUsage.QUERY_RESOLVE | GPUBufferUsage.COPY_SRC,
      });

      // Buffer for CPU readback.
      this.readbackBuffer = device.createBuffer({
        label: "timestamp-readback",
        size: NUM_TIMESTAMPS * 8,
        usage: GPUBufferUsage.COPY_DST | GPUBufferUsage.MAP_READ,
      });
    }
  }

  /** Return true if the device supports timestamp queries. */
  get isSupported(): boolean {
    return this.supported;
  }

  /** Returns the GPUQuerySet (null if unsupported). */
  get gpuQuerySet(): GPUQuerySet | null {
    return this.querySet;
  }

  /** Call at the very start of frame JS work to begin wall-clock measurement. */
  markJsStart(): void {
    this.jsStart = performance.now();
  }

  /** Call after device.queue.submit(). Returns JS wall-clock elapsed ms. */
  markJsEnd(): number {
    return performance.now() - this.jsStart;
  }

  /**
   * Resolve GPU timestamps into resolveBuffer and schedule a COPY to the
   * readback buffer. Must be called before encoder.finish().
   */
  resolveTimestamps(encoder: GPUCommandEncoder): void {
    if (!this.supported || !this.querySet || !this.resolveBuffer || !this.readbackBuffer) return;

    encoder.resolveQuerySet(this.querySet, 0, NUM_TIMESTAMPS, this.resolveBuffer, 0);
    encoder.copyBufferToBuffer(
      this.resolveBuffer, 0,
      this.readbackBuffer, 0,
      NUM_TIMESTAMPS * 8,
    );
  }

  /**
   * Asynchronously read back the resolved timestamps and update smoothed averages.
   * Call this after device.queue.submit() — it maps the readback buffer on the next
   * available tick without blocking the render loop.
   *
   * @param mainThreadMs - JS wall-clock time (from markJsEnd()) to include in the average.
   */
  async readbackAsync(mainThreadMs: number): Promise<void> {
    if (!this.supported || !this.readbackBuffer) {
      this.addToWindow({ totalFrameMs: mainThreadMs, sabUploadMs: 0, computeMs: 0, renderMs: 0, mainThreadMs });
      return;
    }

    const buf = this.readbackBuffer;
    // Only attempt map if buffer is not already mapped.
    await buf.mapAsync(GPUMapMode.READ).catch(() => { /* device lost */ });

    const timestamps = new BigUint64Array(buf.getMappedRange());
    const ns = (a: PassLabel, b: PassLabel): number => {
      const ia = PASS_LABELS.indexOf(a);
      const ib = PASS_LABELS.indexOf(b);
      const delta = timestamps[ib] - timestamps[ia];
      // Clamp negative values (GPU clock wrap or unsupported path)
      return delta < 0n ? 0 : Number(delta) / 1e6; // ns → ms
    };

    const timings: FrameTimings = {
      totalFrameMs:  ns("frame_start", "frame_end"),
      sabUploadMs:   ns("frame_start", "after_sab_upload"),
      computeMs:     ns("after_sab_upload", "after_culling"),
      renderMs:      ns("after_culling", "frame_end"),
      mainThreadMs,
    };

    buf.unmap();
    this.addToWindow(timings);
  }

  private addToWindow(t: FrameTimings): void {
    if (this.window.length < WINDOW) {
      this.window.push(t);
    } else {
      this.window[this.windowIndex % WINDOW] = t;
    }
    this.windowIndex++;

    const count = this.window.length;
    this.smoothed = {
      totalFrameMs:  this.window.reduce((s, f) => s + f.totalFrameMs,  0) / count,
      sabUploadMs:   this.window.reduce((s, f) => s + f.sabUploadMs,   0) / count,
      computeMs:     this.window.reduce((s, f) => s + f.computeMs,     0) / count,
      renderMs:      this.window.reduce((s, f) => s + f.renderMs,      0) / count,
      mainThreadMs:  this.window.reduce((s, f) => s + f.mainThreadMs,  0) / count,
    };
  }

  /** Destroy GPU resources. Call on device loss. */
  destroy(): void {
    this.querySet?.destroy();
    this.resolveBuffer?.destroy();
    this.readbackBuffer?.destroy();
  }
}
