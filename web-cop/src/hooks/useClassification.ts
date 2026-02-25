// CLASSIFICATION: UNCLASSIFIED
// src/hooks/useClassification.ts

import { useMemo } from "react";
import { useTrackStore } from "../stores/trackStore";
import { useAlertStore } from "../stores/alertStore";
import { getHighestClassification } from "../utils/classification";
import { ClassificationLevel } from "../types/common";

/**
 * useClassification — computes the highest classification level currently displayed.
 *
 * Scans all visible tracks and alerts to determine the banner level.
 *
 * @returns { effectiveClassification }
 */
export function useClassification(): { effectiveClassification: ClassificationLevel } {
  const tracks = useTrackStore((s) => s.tracks);
  const alerts = useAlertStore((s) => s.alerts);

  const effectiveClassification = useMemo<ClassificationLevel>(() => {
    const levels: ClassificationLevel[] = [];

    for (const track of tracks.values()) {
      levels.push(track.classification);
    }
    for (const alert of alerts.values()) {
      levels.push(alert.classification);
    }

    const ceiling =
      (import.meta as { env?: Record<string, string> }).env?.["VITE_CLASSIFICATION_CEILING"] as ClassificationLevel | undefined;
    if (ceiling) levels.push(ceiling);

    return getHighestClassification(levels);
  }, [tracks, alerts]);

  return { effectiveClassification };
}
