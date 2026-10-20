# Frontend de Mías Hamburguesas de Lujo

Aplicación web progresiva (PWA) de alta fidelización, catálogo interactivo y gestión de comandas para Mías Hamburguesas de Lujo (Duitama, Boyacá).

## Arquitectura

- **Framework:** Svelte 5 utilizando el sistema de reactividad por Runes (`$state`, `$derived`, `$effect`, `$props`).
- **Herramienta de Construcción:** Vite.
- **Estilos:** Tailwind CSS con tokens gastronómicos de autor (`cream`, `bone`, `char`, `ember`, `gold`).
- **Navegación Suave:** Lenis en escritorio (desactivado en pantallas táctiles para preservar el scroll nativo a 120 FPS de iOS Safari).
- **Animaciones Físicas:** GSAP y transiciones SVG para transiciones contextuales.
- **Resiliencia Anillo 0:** Service Worker (`public/sw.js`) con almacenamiento en caché Cache-First de activos estáticos e imágenes, permitiendo que el catálogo y el generador de pedidos funcionen de manera autónoma sin conexión a internet.
- **Eventos en Tiempo Real:** Consumo de Server-Sent Events (SSE) vía `/api/events/orders` con multiplexación de eventos de cobro, despacho y sellos.

## Estructura de Directorios

```
frontend/
├── public/
│   ├── _headers             # Encabezados de seguridad y caché para Cloudflare Pages
│   ├── manifest.webmanifest # Manifiesto PWA standalone
│   ├── sw.js                # Service Worker para funcionamiento fuera de línea
│   └── imagenes/            # Activos fotográficos optimizados en formato WebP
├── src/
│   ├── components/          # Componentes de interfaz (Catálogo, Carrito, Cartilla, Staff)
│   ├── lib/
│   │   ├── api.js           # Cliente HTTP desacoplado con soporte JWT
│   │   ├── audio.js         # Síntesis acústica de confirmación con Web Audio API
│   │   └── store.svelte.js  # Estado global reactivo con runes y sincronización local
│   ├── App.svelte           # Componente raíz
│   └── main.js              # Punto de entrada de la aplicación
├── index.html               # Documento base optimizado con metadatos SEO y JSON-LD
└── package.json             # Dependencias y scripts de construcción
```

## Ejecución Local

Instalación de dependencias:
```bash
pnpm install
```

Servidor de desarrollo:
```bash
pnpm run dev
```

Compilación para producción:
```bash
pnpm run build
```

Previsualización del bundle de producción:
```bash
pnpm run preview
```
