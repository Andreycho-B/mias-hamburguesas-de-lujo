<script>
  import { store } from '../lib/store.svelte.js';
  import { ArrowRight, User } from 'lucide-svelte';

  let { selectedCategory = 'hamburguesas', onSelectCategory } = $props();

  // Categorías solicitadas en orden intuitivo con tipografía clara y sin sobrecargas
  const categories = [
    { 
      slug: 'hamburguesas', 
      name: 'Hamburguesas', 
      detail: 'Pan pretzel artesanal y 160g de res al carbón' 
    },
    { 
      slug: 'perros', 
      name: 'Perros Calientes', 
      detail: 'Pan brioche artesanal y chorizo santarrosano' 
    },
    { 
      slug: 'de_la_casa', 
      name: 'De la Casa', 
      detail: 'Salchipapas, choripapas y desgranados' 
    },
    { 
      slug: 'adicionales', 
      name: 'Adicionales', 
      detail: 'Papas nativas y tocineta crujiente' 
    },
    { 
      slug: 'bebidas', 
      name: 'Bebidas', 
      detail: 'Línea en vidrio, refrescos y agua con gas' 
    }
  ];

  // Extraer el primer nombre del usuario
  let firstName = $derived.by(() => {
    if (!store.user || !store.user.full_name) return '';
    const parts = store.user.full_name.trim().split(' ');
    return parts[0] || '';
  });

  let touchStartX = 0;
  let touchStartY = 0;
  let isScrollDragging = false;

  function onTouchStart(e) {
    if (e.touches && e.touches[0]) {
      touchStartX = e.touches[0].clientX;
      touchStartY = e.touches[0].clientY;
      isScrollDragging = false;
    }
  }

  function onTouchMove(e) {
    if (!isScrollDragging && e.touches && e.touches[0]) {
      const dx = Math.abs(e.touches[0].clientX - touchStartX);
      const dy = Math.abs(e.touches[0].clientY - touchStartY);
      // Si el desplazamiento supera los 6px, es un gesto de scroll con el dedo
      if (dx > 6 || dy > 6) {
        isScrollDragging = true;
      }
    }
  }

  function handleCategoryClick(e, slug) {
    if (isScrollDragging) {
      if (e) {
        e.preventDefault();
        e.stopPropagation();
      }
      return;
    }
    store.currentView = slug;
    if (onSelectCategory) {
      onSelectCategory(slug);
    }
  }

  function openAuth() {
    store.isAuthModalOpen = true;
  }
</script>

<section 
  class="relative w-full max-w-md md:max-w-4xl lg:max-w-5xl xl:max-w-6xl mx-auto px-5 sm:px-8 md:px-10 lg:px-12 pb-20 md:pb-28 select-none"
  style="padding-top: max(1.25rem, calc(env(safe-area-inset-top, 0px) + 0.85rem)); padding-left: max(1.25rem, env(safe-area-inset-left, 0px)); padding-right: max(1.25rem, env(safe-area-inset-right, 0px));"
>
  
  <!-- RETÍCULA SUPERIOR: BOTÓN MENÚ (=), LOGO CENTRAL "MÍAS" Y BADGE DE PERFIL / SELLOS -->
  <header class="relative flex items-center justify-between pb-6 md:pb-10 pt-1 md:pt-4">
    
    <!-- 1. Botón Izquierdo Circular (Estilo card mate con hairline sutil) -->
    <button
      type="button"
      onclick={() => store.toggleSidebar()}
      class="w-14 h-14 sm:w-16 sm:h-16 rounded-full bg-char-card border border-white/10 flex flex-col items-center justify-center gap-1.5 text-bone hover:bg-char-card-hover active:scale-95 transition-all cursor-pointer z-20 shrink-0"
      aria-label="Abrir menú de navegación"
    >
      <span class="w-5 h-[2.5px] bg-bone rounded-full"></span>
      <span class="w-5 h-[2.5px] bg-bone rounded-full"></span>
    </button>

    <!-- 2. Logo Central "Mías" en Blanco Noble para fondo carbón -->
    <div class="flex items-center justify-center flex-1 px-3 z-10">
      <img 
        src="/imagenes/marca/logo_en_blanco.webp" 
        alt="Mías Hamburguesas de Lujo" 
        class="h-14 sm:h-16 md:h-20 lg:h-24 w-auto max-w-[195px] sm:max-w-[220px] md:max-w-[280px] lg:max-w-[320px] object-contain select-none pointer-events-none transition-all"
      />
    </div>

    <!-- 3. Badge Derecho: Contador de Sellos y Avatar de Usuario (Estilo card mate) -->
    <button
      type="button"
      onclick={openAuth}
      class="flex items-center gap-2 pl-3.5 sm:pl-4 pr-1.5 py-1.5 md:py-2 rounded-full bg-char-card border border-white/10 hover:border-white/20 active:scale-95 transition-all cursor-pointer z-20 shrink-0 group"
      aria-label="Perfil y cartilla de sellos"
      title={store.user ? `${store.user.full_name} (${store.stampBalance} sellos)` : 'Iniciar sesión / Miembros'}
    >
      <!-- Contador de Sellos / Puntos con estrella en oro pulido facetado con brillo especular -->
      <div class="flex items-center gap-1.5">
        <span class="font-mono text-xs sm:text-sm font-bold text-bone tracking-tight">
          {store.stampBalance}
        </span>
        
        <!-- Estrella Minimalista de Cuatro Puntas en Oro Pulido -->
        <svg 
          viewBox="0 0 24 24" 
          class="w-4 h-4 sm:w-4.5 sm:h-4.5 shrink-0 overflow-visible" 
          aria-hidden="true"
        >
          <defs>
            <!-- Gradiente Oro Pulido Minimalista -->
            <linearGradient id="minimalGoldStar" x1="0%" y1="0%" x2="100%" y2="100%">
              <stop offset="0%" stop-color="#FFFDF5" />
              <stop offset="28%" stop-color="#FDE893" />
              <stop offset="65%" stop-color="#D4A745" />
              <stop offset="100%" stop-color="#996E20" />
            </linearGradient>

            <filter id="goldGleamShadow" x="-25%" y="-25%" width="150%" height="150%">
              <feDropShadow dx="0" dy="0.75" stdDeviation="0.6" flood-color="#5B3703" flood-opacity="0.25" />
            </filter>
          </defs>

          <!-- Silueta pura y continua de 4 puntas -->
          <path 
            d="M 12 2.5 C 12 7.8 16.2 12 21.5 12 C 16.2 12 12 16.2 12 21.5 C 12 16.2 7.8 12 2.5 12 C 7.8 12 12 7.8 12 2.5 Z" 
            fill="url(#minimalGoldStar)" 
            filter="url(#goldGleamShadow)"
          />
        </svg>
      </div>

      <!-- Avatar Circular del Usuario -->
      <div class="w-10 h-10 sm:w-11 sm:h-11 rounded-full bg-char text-bone flex items-center justify-center overflow-hidden border border-white/10 group-hover:scale-105 transition-transform">
        {#if store.user}
          <span class="font-display font-bold text-xs uppercase tracking-wider text-gold">
            {firstName.substring(0, 2)}
          </span>
        {:else}
          <User class="w-4 h-4 text-bone/90" />
        {/if}
      </div>
    </button>
  </header>

  <!-- BLOQUE DE SALUDO PRINCIPAL: ADAPTABLE A ESCRITORIO -->
  <div class="mt-8 mb-6 md:mt-12 md:mb-10 text-left md:text-center max-w-3xl md:mx-auto">
    <p class="font-display text-sm md:text-base font-medium tracking-normal md:tracking-wide text-ash mb-1 md:mb-2">
      Lo íntimo hecho para compartir
    </p>
    {#if store.user && firstName}
      <!-- Usuario Registrado: "¡Carlos! Qué gusto tenerte de vuelta" -->
      <h1 class="font-display text-4xl sm:text-5xl md:text-6xl lg:text-[4rem] font-normal text-bone tracking-tight leading-[1.14] md:leading-[1.10]">
        <span class="font-extrabold text-[#FF2400] capitalize">¡{firstName}!</span> Qué gusto<br class="hidden sm:inline" />
        <span class="font-normal text-bone">tenerte de vuelta</span>
      </h1>
    {:else}
      <!-- Usuario No Registrado: Saludo cálido y natural -->
      <h1 class="font-display text-4xl sm:text-5xl md:text-6xl lg:text-[4rem] font-normal text-bone tracking-tight leading-[1.14] md:leading-[1.10]">
        <span class="font-extrabold text-[#FF2400]">¡Hola!</span> Qué gusto<br class="hidden sm:inline" />
        <span class="font-normal text-bone">tenerte por aquí</span>
      </h1>
    {/if}
  </div>

  <!-- SECCIÓN "CARTA": CENTRADA EN TIPOGRAFÍA ITÁLICA EDITORIAL (BONE) -->
  <div class="text-center mt-14 sm:mt-16 md:mt-20 lg:mt-24 mb-7 sm:mb-8 md:mb-12 select-none">
    <h2 class="font-flourish italic text-3xl sm:text-4xl md:text-5xl lg:text-6xl font-normal tracking-tight text-bone">
      Carta
    </h2>
  </div>

  <!-- PÍLDORAS DE CATEGORÍAS (RESPONSIVE GRID: 1 COLUMNA EN MÓVIL, 2 COLUMNAS EN TABLET / ESCRITORIO) -->
  <div class="grid grid-cols-1 md:grid-cols-2 gap-4 sm:gap-5 md:gap-6 max-w-md md:max-w-4xl mx-auto">
    {#each categories as cat, i (cat.slug)}
      {@const isActive = selectedCategory === cat.slug}
      <button
        type="button"
        ontouchstart={onTouchStart}
        ontouchmove={onTouchMove}
        onclick={(e) => handleCategoryClick(e, cat.slug)}
        class="category-pill w-full rounded-full bg-char-card border border-white/10 hover:border-white/20 flex items-center justify-between pl-7 sm:pl-8 md:pl-9 pr-3 md:pr-4 py-4 sm:py-5 min-h-[76px] sm:min-h-[82px] md:min-h-[88px] transition-all duration-200 group cursor-pointer active:scale-[0.985] text-left {i === 4 ? 'md:col-span-2 md:max-w-md md:mx-auto' : ''}"
        style="touch-action: pan-y; -webkit-tap-highlight-color: transparent;"
        aria-label="Ver categoría {cat.name}"
      >
        <!-- Nombre de la categoría e información sutil -->
        <div class="flex-1 pr-3 min-w-0">
          <span class="font-display font-extrabold text-xl sm:text-2xl text-bone tracking-tight block truncate group-hover:translate-x-0.5 transition-transform">
            {cat.name}
          </span>
          <span class="font-display text-xs sm:text-sm text-ash block truncate font-normal mt-0.5">
            {cat.detail}
          </span>
        </div>

        <!-- Botón Circular con Chevron Mórfico a Flecha -->
        <div class="category-arrow-circle w-14 h-14 sm:w-16 sm:h-16 rounded-full {isActive ? 'bg-[#FF2400] text-white is-active' : 'bg-char text-[#FF2400]'} border border-white/10 flex items-center justify-center shrink-0">
          <span class="category-arrow-wrapper inline-flex items-center justify-center">
            <svg 
              viewBox="0 0 24 24" 
              class="w-6 h-6 overflow-visible stroke-[2.2]" 
              fill="none" 
              stroke="currentColor" 
              stroke-linecap="round" 
              stroke-linejoin="round"
              aria-hidden="true"
            >
              <!-- Tallo central que brota fluidamente -->
              <line 
                x1="5.5" 
                y1="12" 
                x2="16" 
                y2="12" 
                class="morphic-stem" 
              />
              <!-- Chevron angular centrado en reposo -->
              <path 
                d="M 11.5 7 L 16.5 12 L 11.5 17" 
                class="morphic-chevron" 
              />
            </svg>
          </span>
        </div>
      </button>
    {/each}
  </div>
</section>

<!-- SECCIÓN EDITORIAL A PANTALLA COMPLETA (CONTINUACIÓN DIRECTA DE LA CARTA) -->
<section 
  class="relative w-full min-h-[100dvh] flex flex-col justify-between items-center py-10 sm:py-16 px-4 sm:px-6 select-none overflow-hidden rounded-none bg-[#0D0204]"
>
  <!-- 1. IMAGEN PEGADA A LOS LATERALES (ANCHO TOTAL 100% SIN HUECOS NEGROS) -->
  <div class="absolute inset-0 w-full h-full p-0 z-0">
    <img 
      src="/imagenes/historia/horno_a_carbon.webp" 
      alt="Horno al Carbón Rayo" 
      class="w-full h-full min-w-full object-cover object-center filter contrast-[1.04] brightness-100 select-none pointer-events-none"
      style="mask-image: linear-gradient(180deg, transparent 0%, black 14%, black 86%, transparent 100%); -webkit-mask-image: linear-gradient(180deg, transparent 0%, black 14%, black 86%, transparent 100%);"
      loading="eager"
    />
  </div>

  <!-- 2. VELO OSCURO PERIMETRAL QUE DIFUMINA CUALQUIER LÍNEA DE CORTE EN LOS 4 COSTADOS -->
  <!-- Fundido superior -->
  <div class="absolute top-0 left-0 right-0 h-36 bg-gradient-to-b from-[#0D0204] via-[#0D0204]/80 to-transparent pointer-events-none z-10"></div>
  <!-- Fundido inferior -->
  <div class="absolute bottom-0 left-0 right-0 h-44 bg-gradient-to-t from-[#0D0204] via-[#0D0204]/90 to-transparent pointer-events-none z-10"></div>
  <!-- Velo lateral izquierdo y derecho suavizado con menor sombra -->
  <div 
    class="absolute inset-0 pointer-events-none z-10"
    style="background: linear-gradient(90deg, rgba(13,2,4,0.55) 0%, rgba(13,2,4,0.20) 10%, transparent 28%, transparent 72%, rgba(13,2,4,0.20) 90%, rgba(13,2,4,0.55) 100%);"
  ></div>

  <!-- Resplandor sutil de brasa detrás del horno -->
  <div class="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[320px] sm:w-[500px] h-[320px] sm:h-[500px] bg-[#FF2400]/12 rounded-full blur-3xl pointer-events-none z-10"></div>

  <!-- 3. TÍTULO EN LA ZONA SUPERIOR (ENCIMA DE LA IMAGEN EN TIPOGRAFÍA DE "CARTA") -->
  <div class="relative z-20 text-center px-4 max-w-xl mx-auto pt-6 sm:pt-10 md:pt-12">
    <h2 class="font-flourish italic text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-normal tracking-tight text-[#F7F4EC] leading-[1.12] drop-shadow-[0_4px_24px_rgba(0,0,0,0.95)]">
      El Arte<br />
      <span class="font-flourish italic">de la Brasa</span>
    </h2>
  </div>

  <!-- Espacio inferior limpio -->
  <div class="relative z-20 pb-12 sm:pb-16"></div>
</section>

<style>
  /* MICRO-INTERACCIÓN: CHEVRON MÓRFICO A FLECHA */
  .category-arrow-circle {
    transition: transform 0.28s cubic-bezier(0.34, 1.4, 0.64, 1), 
                background-color 0.24s ease, 
                border-color 0.24s ease, 
                color 0.24s ease;
    will-change: transform;
  }

  .category-arrow-wrapper {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    transition: transform 0.30s cubic-bezier(0.16, 1, 0.3, 1);
    will-change: transform;
  }

  /* Chevron en reposo: perfectamente centrado en el círculo */
  .morphic-chevron {
    transform: translateX(-2.5px);
    transition: transform 0.30s cubic-bezier(0.16, 1, 0.3, 1);
    will-change: transform;
  }

  /* Tallo central en reposo: contraído y oculto */
  .morphic-stem {
    stroke-dasharray: 12;
    stroke-dashoffset: 12;
    opacity: 0;
    transition: stroke-dashoffset 0.28s cubic-bezier(0.16, 1, 0.3, 1),
                opacity 0.20s ease;
    will-change: stroke-dashoffset, opacity;
  }

  /* Estado Activo / Seleccionado: la flecha permanece completamente desplegada */
  .category-arrow-circle.is-active .morphic-stem {
    stroke-dashoffset: 0;
    opacity: 1;
  }
  .category-arrow-circle.is-active .morphic-chevron {
    transform: translateX(1.5px);
  }
  .category-arrow-circle.is-active .category-arrow-wrapper {
    transform: translateX(2px);
  }

  /* Micro-interacción en hover exclusiva para punteros de escritorio */
  @media (hover: hover) and (pointer: fine) {
    .category-pill:hover {
      border-color: rgba(0, 0, 0, 0.50);
    }

    /* El círculo crece sutilmente y se tiñe de rojo vivo */
    .category-pill:hover .category-arrow-circle:not(.is-active) {
      background-color: #FF2400 !important;
      color: #ffffff !important;
      transform: scale(1.06);
    }

    /* Brota el tallo horizontal convirtiendo el chevron en flecha */
    .category-pill:hover .morphic-stem {
      stroke-dashoffset: 0;
      opacity: 1;
    }

    /* El chevron avanza con el impulso del tallo */
    .category-pill:hover .morphic-chevron {
      transform: translateX(2px);
    }

    .category-pill:hover .category-arrow-wrapper {
      transform: translateX(3px);
    }
  }

  /* Feedback táctil o click al presionar (móvil y escritorio) */
  .category-pill:active .category-arrow-circle {
    transform: scale(0.93);
  }

  .category-pill:active .morphic-stem {
    stroke-dashoffset: 0;
    opacity: 1;
  }

  .category-pill:active .morphic-chevron {
    transform: translateX(3.5px);
  }

  .category-pill:active .category-arrow-wrapper {
    transform: translateX(4px);
  }
</style>
