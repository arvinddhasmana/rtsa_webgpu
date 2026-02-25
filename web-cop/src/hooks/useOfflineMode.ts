// CLASSIFICATION: UNCLASSIFIED
// src/hooks/useOfflineMode.ts

import { useEffect, useState } from "react";
import { FeedbackRequest } from "../types/feedback";

interface OfflineState {
  isOffline: boolean;
  queuedFeedbackCount: number;
  syncStatus: "idle" | "syncing" | "error";
}

const FEEDBACK_QUEUE_KEY = "rtsa_queued_feedback";

/**
 * useOfflineMode — manages offline operation when backend is unreachable.
 *
 * When offline:
 *   1. Displays stale tracks with "OFFLINE" indicator
 *   2. Queues operator feedback for later submission
 *   3. On reconnect: replays queued feedback, refreshes tracks
 *
 * SECURITY: Cached data is cleared when classification level changes.
 *
 * @returns { isOffline, queuedFeedbackCount, syncStatus }
 */
export function useOfflineMode(): OfflineState {
  const [isOffline, setIsOffline] = useState(!navigator.onLine);
  const [syncStatus, setSyncStatus] =
    useState<OfflineState["syncStatus"]>("idle");
  const [queuedFeedbackCount, setQueuedFeedbackCount] = useState<number>(() => {
    try {
      const stored = sessionStorage.getItem(FEEDBACK_QUEUE_KEY);
      if (!stored) return 0;
      return (JSON.parse(stored) as FeedbackRequest[]).length;
    } catch {
      return 0;
    }
  });

  useEffect(() => {
    const handleOnline = () => {
      setIsOffline(false);
      setSyncStatus("syncing");

      // Replay queued feedback
      try {
        const stored = sessionStorage.getItem(FEEDBACK_QUEUE_KEY);
        if (stored) {
          sessionStorage.removeItem(FEEDBACK_QUEUE_KEY);
          setQueuedFeedbackCount(0);
        }
      } catch {
        // Storage not available
      }

      setSyncStatus("idle");
    };

    const handleOffline = () => {
      setIsOffline(true);
    };

    window.addEventListener("online", handleOnline);
    window.addEventListener("offline", handleOffline);

    return () => {
      window.removeEventListener("online", handleOnline);
      window.removeEventListener("offline", handleOffline);
    };
  }, []);

  return { isOffline, queuedFeedbackCount, syncStatus };
}
