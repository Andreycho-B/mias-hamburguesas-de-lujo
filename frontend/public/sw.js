// Service Worker Ultraligero para Mías Hamburguesas de Lujo
// Optimizado para arranque instantáneo de PWA en iOS Safari y Android
const CACHE_NAME = 'mias-static-v4';
const IMAGE_CACHE = 'mias-images-v4';

// App Shell crítico y activos esenciales para renderizado y navegación instantánea
const CRITICAL_SHELL = [
  '/',
  '/index.html',
  '/manifest.webmanifest',
  '/apple-touch-icon.png',
  '/imagenes/marca/isotipo.webp',
  '/imagenes/marca/logo_en_blanco.webp',
  '/imagenes/local/restaurante_mias_desde_dentro.webp',
  '/imagenes/local/restaurante_mias_desde_fuera.webp',
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
      return cache.addAll(CRITICAL_SHELL).catch((err) => {
        console.warn('[SW] Precache no bloqueante:', err);
      });
    })
  );
  self.skipWaiting();
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((keys) => {
      return Promise.all(
        keys
          .filter((key) => key !== CACHE_NAME && key !== IMAGE_CACHE)
          .map((key) => caches.delete(key))
      );
    })
  );
  self.clients.claim();
});

self.addEventListener('fetch', (event) => {
  if (event.request.method !== 'GET') return;
  if (!event.request.url.startsWith('http')) return;

  const url = new URL(event.request.url);

  // 1. Omitir endpoints dinámicos de API y módulos de desarrollo en vivo (Vite)
  // para que vayan directo por red sin latencia de Service Worker
  if (
    url.pathname.startsWith('/api/') ||
    url.pathname.startsWith('/@vite/') ||
    url.pathname.startsWith('/@fs/') ||
    url.pathname.startsWith('/@id/') ||
    url.pathname.startsWith('/src/') ||
    url.pathname.includes('node_modules') ||
    url.searchParams.has('t') ||
    url.searchParams.has('import')
  ) {
    return;
  }

  // 2. Navegación HTML (PWA Start URL): Network-First con fallback a index.html en caché
  if (event.request.mode === 'navigate') {
    event.respondWith(
      fetch(event.request)
        .then((networkResponse) => {
          if (networkResponse && networkResponse.status === 200) {
            const copy = networkResponse.clone();
            caches.open(CACHE_NAME).then((cache) => cache.put(event.request, copy));
          }
          return networkResponse;
        })
        .catch(() => {
          return caches.match('/index.html') || caches.match('/');
        })
    );
    return;
  }

  // 3. Imágenes: Cache-First con guardado dinámico on-demand
  // Las imágenes se almacenan en segundo plano a medida que el usuario navega,
  // eliminando la saturación de conexiones durante la primera carga.
  const isImage =
    url.pathname.startsWith('/imagenes/') ||
    url.pathname.match(/\.(webp|png|jpg|jpeg|svg|gif|ico)$/i);

  if (isImage) {
    event.respondWith(
      caches.match(event.request).then((cachedResponse) => {
        if (cachedResponse) {
          return cachedResponse;
        }

        return fetch(event.request)
          .then((networkResponse) => {
            if (networkResponse && networkResponse.status === 200) {
              const copy = networkResponse.clone();
              caches.open(IMAGE_CACHE).then((cache) => cache.put(event.request, copy));
            }
            return networkResponse;
          })
          .catch(() => {
            return caches.match('/imagenes/marca/isotipo.webp');
          });
      })
    );
    return;
  }

  // 4. Recursos estáticos generales (CSS/JS compilados, fuentes)
  event.respondWith(
    caches.match(event.request).then((cachedResponse) => {
      if (cachedResponse) {
        return cachedResponse;
      }
      return fetch(event.request);
    })
  );
});
