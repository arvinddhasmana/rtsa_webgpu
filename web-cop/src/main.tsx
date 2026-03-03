// CLASSIFICATION: UNCLASSIFIED
// src/main.tsx

import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import "./index.css";

const rootElement = document.getElementById("root");
if (!rootElement) {
  throw new Error("Root element not found");
}

// Register Service Worker for offline mode
if ("serviceWorker" in navigator) {
  window.addEventListener("load", () => {
    navigator.serviceWorker
      .register("/sw.js")
      .then((registration) => {
        console.log("SW registered: ", registration);
      })
      .catch((registrationError) => {
        console.log("SW registration failed: ", registrationError);
      });
  });
}

// Expose Zustand stores on window for Playwright E2E testing and devtools.
// UNCLASSIFIED artifacts only — these stores contain display-level state, no secrets.
import("./stores/trackStore").then(({ useTrackStore }) => {
  (window as unknown as Record<string, unknown>)["__RTSA_TRACK_STORE__"] =
    useTrackStore;
});
import("./stores/alertStore").then(({ useAlertStore }) => {
  (window as unknown as Record<string, unknown>)["__RTSA_ALERT_STORE__"] =
    useAlertStore;
});
import("./stores/sensorHealthStore").then(({ useSensorHealthStore }) => {
  (window as unknown as Record<string, unknown>)[
    "__RTSA_SENSOR_HEALTH_STORE__"
  ] = useSensorHealthStore;
});

ReactDOM.createRoot(rootElement).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
