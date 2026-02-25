// CLASSIFICATION: UNCLASSIFIED
// public/sw.js — Service Worker for offline capability

const CACHE_NAME = "rtsa-cop-v1";
const STATIC_ASSETS = [
  "/",
  "/index.html",
];

self.addEventListener("install", (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => cache.addAll(STATIC_ASSETS))
  );
  self.skipWaiting();
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches.keys().then((names) =>
      Promise.all(
        names
          .filter((name) => name !== CACHE_NAME)
          .map((name) => caches.delete(name))
      )
    )
  );
  self.clients.claim();
});

self.addEventListener("fetch", (event) => {
  const url = new URL(event.request.url);

  // Do NOT cache gRPC-Web requests — they must go to the backend
  if (url.pathname.startsWith("/rtsa.")) {
    return;
  }

  event.respondWith(
    caches.match(event.request).then(
      (cached) => cached ?? fetch(event.request)
    )
  );
});
