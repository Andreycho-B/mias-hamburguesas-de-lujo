/**
 * Preloader ultra-rápido de imágenes críticas con decodificación GPU asíncrona
 * Carga de inmediato todas las imágenes esenciales en paralelo para disponibilidad instantánea.
 */
export const MODAL_IMAGES = [
  '/imagenes/local/restaurante_mias_desde_dentro.webp',
  '/imagenes/local/restaurante_mias_desde_fuera.webp',
  '/imagenes/marca/isotipo.webp',
  '/imagenes/marca/logo_en_blanco.webp'
];

export const CATALOG_IMAGES = [
  '/imagenes/hamburguesas/Orignal.webp',
  '/imagenes/hamburguesas/Crispy.webp',
  '/imagenes/hamburguesas/Golden.webp',
  '/imagenes/hamburguesas/Criolla.webp',
  '/imagenes/hamburguesas/Brava.webp',
  '/imagenes/hamburguesas/Suprema.webp',
  '/imagenes/de_la_casa/salchi_papa.png',
  '/imagenes/perros/perro.png',
  '/imagenes/bebidas/cocacola.webp',
  '/imagenes/bebidas/bretana-vidrio.png',
  '/imagenes/historia/candado_branding_historia.webp',
  '/imagenes/historia/horno_a_carbon.webp'
];

let preloadPromise = null;

function loadImage(src) {
  return new Promise((resolve) => {
    const img = new Image();
    img.src = src;
    if ('decode' in img) {
      img.decode().then(resolve).catch(resolve);
    } else {
      img.onload = resolve;
      img.onerror = resolve;
    }
  });
}

/**
 * Precarga y decodificación inmediata en paralelo para renderizado instantáneo
 */
export function preloadCriticalImages() {
  if (typeof window === 'undefined') return Promise.resolve();
  if (preloadPromise) return preloadPromise;

  // Precargar primero el modal de inmediato y en paralelo
  const modalPromise = Promise.allSettled(MODAL_IMAGES.map(loadImage));

  // Paralelamente iniciar catálogo completo
  const catalogPromise = Promise.allSettled(CATALOG_IMAGES.map(loadImage));

  preloadPromise = Promise.all([modalPromise, catalogPromise]);
  return preloadPromise;
}
