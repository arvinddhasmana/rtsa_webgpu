// CLASSIFICATION: UNCLASSIFIED
// src/gpu/coverage.ts — Logic for sensor coverage rendering pass
//
// Manages the coverage storage buffer and draw commands.
// Coverage data is updated from the CPU based on ingestion events and Spatial Alerts.

import { COVERAGE_RECORD_BYTES, GPUBuffers, MAX_COVERAGE_RECORDS } from "./buffers";
import { AllPipelines } from "./pipelines";

export interface CoverageRecord {
  centerLon:    number;
  centerLat:    number;
  rangeNm:      number;
  bearingStart: number;
  bearingEnd:   number;
  recordType:   number; // 0 = Sector, 1 = Gap Polygon
  alertLevel:   number; // 0 = Normal, 1 = Warning, 2 = Critical
}

export class CoverageManager {
  private device: GPUDevice;
  private buffers: GPUBuffers;
  private pipelines: AllPipelines;

  private recordCount: number = 0;
  private cpuBuffer: ArrayBuffer;
  private f32View: Float32Array;
  private u32View: Uint32Array;
  private recordsBindGroup: GPUBindGroup;

  constructor(device: GPUDevice, buffers: GPUBuffers, pipelines: AllPipelines) {
    this.device = device;
    this.buffers = buffers;
    this.pipelines = pipelines;

    this.cpuBuffer = new ArrayBuffer(MAX_COVERAGE_RECORDS * COVERAGE_RECORD_BYTES);
    this.f32View = new Float32Array(this.cpuBuffer);
    this.u32View = new Uint32Array(this.cpuBuffer);

    // Pre-allocate bind group once in constructor (webgpu_guidelines.md — no per-frame GPU object creation)
    this.recordsBindGroup = this.device.createBindGroup({
      label: "coverage-records-bind-group",
      layout: this.pipelines.render.coverage.getBindGroupLayout(1),
      entries: [
        {
          binding: 0,
          resource: { buffer: this.buffers.coverageStorage }
        }
      ]
    });
  }

  /**
   * Reset the record count for the current frame.
   */
  reset(): void {
    this.recordCount = 0;
  }

  /**
   * Add a coverage record (sector or gap) to be rendered this frame.
   */
  addRecord(rec: CoverageRecord): void {
    if (this.recordCount >= MAX_COVERAGE_RECORDS) return;

    const offset = this.recordCount * (COVERAGE_RECORD_BYTES / 4);

    this.f32View[offset + 0] = rec.centerLon;
    this.f32View[offset + 1] = rec.centerLat;
    this.f32View[offset + 2] = rec.rangeNm;
    this.f32View[offset + 3] = rec.bearingStart;
    this.f32View[offset + 4] = rec.bearingEnd;

    this.u32View[offset + 5] = rec.recordType;
    this.u32View[offset + 6] = rec.alertLevel;
    this.u32View[offset + 7] = 0; // Padding

    this.recordCount++;
  }

  /**
   * Upload record data to the GPU.
   */
  upload(): void {
    if (this.recordCount === 0) return;

    this.device.queue.writeBuffer(
      this.buffers.coverageStorage,
      0,
      this.cpuBuffer,
      0,
      this.recordCount * COVERAGE_RECORD_BYTES
    );
  }

  /**
   * Record draw commands for the coverage pass.
   */
  draw(passEncoder: GPURenderPassEncoder, uniformBindGroup: GPUBindGroup): void {
    if (this.recordCount === 0) return;

    passEncoder.setPipeline(this.pipelines.render.coverage);
    passEncoder.setBindGroup(0, uniformBindGroup);
    passEncoder.setBindGroup(1, this.recordsBindGroup);
    passEncoder.draw(6, this.recordCount, 0, 0);
  }
}
