// CLASSIFICATION: UNCLASSIFIED
// src/hooks/useClassification.ts

import { useEffect, useRef, useState } from "react";
import { useAlertStore } from "../stores/alertStore";
import { useTrackStore } from "../stores/trackStore";
import { ClassificationLevel } from "../types/common";
import { getHighestClassification } from "../utils/classification";

const CLASSIFICATION_CEILING = (import.meta as { env?: Record<string, string> })
  .env?.["VITE_CLASSIFICATION_CEILING"] as ClassificationLevel | undefined;

/**
 * Compute the effective classification by reading store state directly.
 * Called OUTSIDE React render paths to avoid creating a reactive dependency
 * on the full tracks Map (which changes every streaming frame).
 */
function computeClassification(): ClassificationLevel {
  const levels: ClassificationLevel[] = [];
  for (const track of useTrackStore.getState().tracks.values()) {
    levels.push(track.classification);
  }
  for (const alert of useAlertStore.getState().alerts.values()) {
    levels.push(alert.classification);
  }
  if (CLASSIFICATION_CEILING) levels.push(CLASSIFICATION_CEILING);
  return getHighestClassification(levels);
}

/**
 * useClassification — computes the highest classification level currently displayed.
 *
 * PERFORMANCE: Uses store.subscribe() + 2-second throttle + RAF instead of a
 * React selector on the full tracks Map.  This prevents MainLayout from
 * re-rendering on every streaming frame (which caused UI unresponsiveness and
 * caused native <select> dropdowns to close instantly after opening).
 *
 * @returns { effectiveClassification }
 */
export function useClassification(): {
  effectiveClassification: ClassificationLevel;
} {
  const [effectiveClassification, setEffective] = useState<ClassificationLevel>(
    () => computeClassification(),
  );
  const rafIdRef = useRef<number | null>(null);
  // -2001 so the very first store update (with real track data) always passes.
  const lastUpdateRef = useRef<number>(-2001);

  useEffect(() => {
    const scheduleRecompute = () => {
      // Throttle: at most 1 recompute per 2 seconds.
      // Classification changes only when track/alert classification fields change —
      // extremely rare compared to position updates streaming at 60 fps.
      const now = Date.now();
      if (now - lastUpdateRef.current < 2000) return;
      lastUpdateRef.current = now;

      if (rafIdRef.current !== null) return; // already scheduled
      rafIdRef.current = requestAnimationFrame(() => {
        rafIdRef.current = null;
        const next = computeClassification();
        setEffective((prev) => (prev === next ? prev : next));
      });
    };

    const unsub1 = useTrackStore.subscribe(scheduleRecompute);
    const unsub2 = useAlertStore.subscribe(scheduleRecompute);

    return () => {
      if (rafIdRef.current !== null) {
        cancelAnimationFrame(rafIdRef.current);
        rafIdRef.current = null;
      }
      unsub1();
      unsub2();
    };
  }, []);

  return { effectiveClassification };
}
