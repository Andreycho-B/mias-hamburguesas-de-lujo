// Service Worker para Mías.8 (Nivel 0: Resiliencia de Menú y Despacho Offline)
const CACHE_NAME = 'mias-static-v1';
const ASSETS_TO_CACHE = [
  '/',
  '/index.html',
  '/manifest.webmanifest',
  '/imagenes/marca/isotipo.webp',
  '/imagenes/marca/logo_en_blanco.webp',
  '/imagenes/marca/logo_en_negro.webp',
  '/imagenes/marca/logo_en_rojo.webp',
  '/imagenes/hamburguesas/Orignal.webp',
  '/imagenes/hamburguesas/Crispy.webp',
  '/imagenes/hamburguesas/Golden.webp',
  '/imagenes/hamburguesas/Criolla.webp',
  '/imagenes/hamburguesas/Brava.webp',
  '/imagenes/hamburguesas/Suprema.webp'
];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => {
      return cache.addAll(ASSETS_TO_CACHE);
    })
  );
  self.skipWaiting();
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((keys) => {
      return Promise.all(
        keys.filter((key) => key !== CACHE_NAME).map((key) => caches.delete(key))
      );
    })
  );
  self.clients.claim();
});

self.addEventListener('fetch', (event) => {
  // Solo cachear peticiones GET
  if (event.request.method !== 'GET') return;

  event.respondWith(
    caches.match(event.request).then((cachedResponse) => {
      if (cachedResponse) {
        return cachedResponse;
      }
      return fetch(event.request).then((networkResponse) => {
        return networkResponse;
      }).catch(() => {
        // Fallback offline
        if (event.request.mode === 'navigate') {
          return caches.match('/index.html');
        }
      });
    })
  );
});
