const CACHE_NAME = 'cityquest-v1'

// Hashed assets are immutable — cache forever.
// index.html and manifest are network-first.
// API, SSE, uploads — NEVER cached.
const PRECACHE_URLS = ['/manifest.json']

self.addEventListener('install', (e) => {
  e.waitUntil(
    caches.open(CACHE_NAME).then((cache) => cache.addAll(PRECACHE_URLS))
  )
  self.skipWaiting()
})

self.addEventListener('activate', (e) => {
  e.waitUntil(
    caches.keys().then((names) =>
      Promise.all(
        names
          .filter((name) => name !== CACHE_NAME)
          .map((name) => caches.delete(name))
      )
    )
  )
  self.clients.claim()
})

self.addEventListener('fetch', (e) => {
  const url = new URL(e.request.url)

  // NEVER cache: API calls, SSE, uploads, WebSocket, admin
  if (
    url.pathname.startsWith('/api/') ||
    url.pathname.startsWith('/uploads/') ||
    url.pathname.startsWith('/healthz') ||
    e.request.headers.get('accept') === 'text/event-stream'
  ) {
    return
  }

  // Hashed assets (/assets/index-abc123.js) — cache-first (immutable)
  if (url.pathname.startsWith('/assets/')) {
    e.respondWith(
      caches.match(e.request).then(
        (cached) =>
          cached ||
          fetch(e.request).then((response) => {
            if (response.ok) {
              const clone = response.clone()
              caches.open(CACHE_NAME).then((cache) => cache.put(e.request, clone))
            }
            return response
          })
      )
    )
    return
  }

  // Navigation (index.html, SPA routes) — network-first, fall back to cache
  // This ensures new deploys are picked up immediately, but offline still works
  if (e.request.mode === 'navigate') {
    e.respondWith(
      fetch(e.request)
        .then((response) => {
          const clone = response.clone()
          caches.open(CACHE_NAME).then((cache) => cache.put(e.request, clone))
          return response
        })
        .catch(() => caches.match(e.request))
    )
    return
  }

  // Static files (icons, fonts, manifest) — stale-while-revalidate
  e.respondWith(
    caches.match(e.request).then((cached) => {
      const fetching = fetch(e.request)
        .then((response) => {
          if (response.ok) {
            const clone = response.clone()
            caches.open(CACHE_NAME).then((cache) => cache.put(e.request, clone))
          }
          return response
        })
        .catch(() => cached)
      return cached || fetching
    })
  )
})
