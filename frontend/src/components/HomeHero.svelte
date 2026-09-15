<script>
  import { store } from '../lib/store.svelte.js';
  import { ArrowRight, User, Receipt, Flame, Clock, Bike } from 'lucide-svelte';

  let { selectedCategory = 'hamburguesas', onSelectCategory } = $props();

  const activeStatuses = ['PENDING', 'CONFIRMED', 'PREPARING', 'IN_DELIVERY', 'READY'];
  let latestActiveOrder = $derived(
    store.activeOrders?.find((o) => activeStatuses.includes(o.status)) || null
  );
  let hasAnyOrders = $derived(store.activeOrders && store.activeOrders.length > 0);

  const defaultCategoryDetails = {
    hamburguesas: 'Pan pretzel artesanal y 160g de res al carbón',
    perros: 'Pan brioche artesanal y chorizo santarrosano',
    de_la_casa: 'Salchipapas, choripapas y desgranados',
    adicionales: 'Papas nativas y tocineta crujiente',
    bebidas: 'Línea en vidrio, refrescos y agua con gas',
    temporada: 'Especiales de temporada'
  };

  const defaultCategories = [
    { slug: 'hamburguesas', name: 'Hamburguesas', detail: 'Pan pretzel artesanal y 160g de res al carbón' },
    { slug: 'perros', name: 'Perros Calientes', detail: 'Pan brioche artesanal y chorizo santarrosano' },
    { slug: 'de_la_casa', name: 'De la Casa', detail: 'Salchipapas, choripapas y desgranados' },
    { slug: 'adicionales', name: 'Adicionales', detail: 'Papas nativas y tocineta crujiente' },
    { slug: 'bebidas', name: 'Bebidas', detail: 'Línea en vidrio, refrescos y agua con gas' }
  ];

  // Categorías sincronizadas dinámicamente con la base de datos
  let categories = $derived(
    store.categories && store.categories.length > 0
      ? store.categories.map((c) => ({
          slug: c.slug,
          name: c.name,
          detail: c.metadata?.detail || defaultCategoryDetails[c.slug] || `Especialidades de ${c.name}`
        }))
      : defaultCategories
  );

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
  
  <!-- RETÍCULA SUPERIOR: LOGO CENTRAL "MÍAS" EN BLANCO Y BADGE DE PERFIL / SELLOS -->
  <header class="relative grid grid-cols-[1fr_auto_1fr] items-center pb-6 md:pb-10 pt-1 md:pt-4">
    
    <!-- Espacio izquierdo equilibrado sin isotipo para centrado simétrico del logotipo -->
    <div class="flex items-center justify-start"></div>

    <!-- Logo Central "Mías" en Blanco Noble para fondo carbón -->
    <div class="flex items-center justify-center px-2 z-10">
      <img 
        src="/imagenes/marca/logo_en_blanco.webp" 
        alt="Mías Hamburguesas de Lujo" 
        class="h-14 sm:h-16 md:h-20 lg:h-24 w-auto max-w-[195px] sm:max-w-[220px] md:max-w-[280px] lg:max-w-[320px] object-contain select-none pointer-events-none transition-all"
      />
    </div>

    <!-- Badge Derecho: Contador de Sellos y Avatar de Usuario (Estilo mate sutil) -->
    <div class="flex items-center justify-end z-20">
      <button
        type="button"
        onclick={openAuth}
        class="gold-border-card flex items-center gap-2 pl-3.5 sm:pl-4 pr-1.5 py-1.5 md:py-2 rounded-full active:scale-95 transition-all cursor-pointer group"
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
      <h1 class="font-display text-4xl sm:text-5xl md:text-6xl lg:text-[4rem] font-normal text-bone tracking-tight leading-[1.28] sm:leading-[1.24] md:leading-[1.18]">
        <span class="gold-text-pinyon text-5xl sm:text-6xl md:text-7xl lg:text-[4.75rem] capitalize mr-1 sm:mr-1.5">¡{firstName}!</span> Qué gusto<br class="hidden sm:inline" />
        <span class="font-normal text-bone">tenerte de vuelta</span>
      </h1>
    {:else}
      <!-- Usuario No Registrado: Saludo cálido y natural -->
      <h1 class="font-display text-4xl sm:text-5xl md:text-6xl lg:text-[4rem] font-normal text-bone tracking-tight leading-[1.28] sm:leading-[1.24] md:leading-[1.18]">
        <span class="gold-text-pinyon text-5xl sm:text-6xl md:text-7xl lg:text-[4.75rem] mr-1 sm:mr-1.5">¡Hola!</span> Qué gusto<br class="hidden sm:inline" />
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
        class="category-pill gold-border-card w-full rounded-full flex items-center justify-between pl-7 sm:pl-8 md:pl-9 pr-3 md:pr-4 py-4 sm:py-5 min-h-[76px] sm:min-h-[82px] md:min-h-[88px] transition-all duration-200 group cursor-pointer active:scale-[0.985] text-left {i === 4 ? 'md:col-span-2 md:max-w-md md:mx-auto' : ''}"
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

        <!-- Botón Circular con Chevron Mórfico a Flecha y Borde Dorado Biselado -->
        <div class="category-arrow-circle w-14 h-14 sm:w-16 sm:h-16 rounded-full flex items-center justify-center shrink-0 {isActive ? 'is-active' : ''}">
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

<!-- ═══════════════════════════════════════════════════════════════════════ -->
<!-- SECCIONES EDITORIALES DE MARCA (DISEÑO LIMPIO, SIN TARJETAS NI CAJAS)    -->
<!-- ═══════════════════════════════════════════════════════════════════════ -->
<section class="w-full mt-20 sm:mt-28 md:mt-36 border-t hairline-dark bg-[#08080a] text-bone select-none">
  
  <!-- 1. CAPÍTULO: EL HORNO RAYO Y LA COCCIÓN AL CARBÓN -->
  <article class="max-w-5xl mx-auto px-5 sm:px-8 md:px-12 py-16 sm:py-24 md:py-32">
    <div class="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-10 sm:mb-14">
      <div class="max-w-xl">
        <span class="font-mono text-[11px] sm:text-xs text-gold uppercase tracking-[0.22em] block mb-3">
          Mías • Fundamentos del Fuego
        </span>
        <h2 class="font-flourish italic text-3xl sm:text-4xl md:text-5xl lg:text-6xl text-bone leading-[1.12] tracking-tight">
          El Horno Rayo y la Cocción al Carbón
        </h2>
      </div>
      <p class="font-sans text-sm sm:text-base text-ash max-w-md leading-relaxed font-normal">
        Carne de res premium de 160 gramos sobre brasa viva. Sellado rápido que conserva los jugos y aporta un ahumado natural sin recurrir a prensas mecánicas.
      </p>
    </div>

    <!-- Fotografía Editorial a ancho completo sin marcos -->
    <div class="relative w-full aspect-[16/10] sm:aspect-[21/9] overflow-hidden my-8 md:my-12">
      <img 
        src="/imagenes/historia/horno_a_carbon.webp" 
        alt="Horno Rayo al carbón en Mías Hamburguesas" 
        class="w-full h-full object-cover object-center filter contrast-[1.05] brightness-95"
        loading="lazy"
      />
      <div class="absolute inset-0 bg-gradient-to-t from-[#08080a] via-transparent to-transparent opacity-60"></div>
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-8 sm:gap-12 pt-4 sm:pt-8 text-sm sm:text-base text-ash leading-relaxed">
      <p>
        El horno «Rayo», fabricado en hierro escarlata con remates de latón dorado, trabaja a temperatura continua con carbón vegetal seleccionado. Este método carameliza la superficie exterior de la carne en segundos, sellando la grasa intramuscular para garantizar terneza y una humedad uniforme en cada porción.
      </p>
      <p>
        Cada fórmula de cocción se apega a la receta original: sazón al momento del contacto con el hierro, tiempo cronometrado de volteo y reposo breve antes del montaje para estabilizar la temperatura interna sin alterar la fibra cárnica.
      </p>
    </div>
  </article>

  <!-- LÍNEA DIVISORIA SUIZA -->
  <div class="w-full max-w-5xl mx-auto px-5 sm:px-8 md:px-12">
    <div class="w-full border-t hairline-dark"></div>
  </div>

  <!-- 2. CAPÍTULO: EL PAN ARTESANAL (PRETZEL & BRIOCHE) -->
  <article class="max-w-5xl mx-auto px-5 sm:px-8 md:px-12 py-16 sm:py-24 md:py-32">
    <div class="flex flex-col md:flex-row md:items-end justify-between gap-6 mb-10 sm:mb-14">
      <div class="max-w-xl">
        <span class="font-mono text-[11px] sm:text-xs text-gold uppercase tracking-[0.22em] block mb-3">
          Oficio Panadero
        </span>
        <h2 class="font-flourish italic text-3xl sm:text-4xl md:text-5xl lg:text-6xl text-bone leading-[1.12] tracking-tight">
          Dos Masas, Dos Propósitos
        </h2>
      </div>
      <p class="font-sans text-sm sm:text-base text-ash max-w-md leading-relaxed font-normal">
        El pan no es acompañamiento: define la arquitectura de cada bocado. Formulaciones independientes para hamburguesas y perros calientes.
      </p>
    </div>

    <!-- Galería Editorial Limpia (Sin cards, ritmo fotográfico natural) -->
    <div class="grid grid-cols-1 sm:grid-cols-2 gap-8 sm:gap-12 my-8 md:my-12">
      <div>
        <div class="relative w-full aspect-[4/3] overflow-hidden">
          <img 
            src="/imagenes/historia/pinselando_pan.webp" 
            alt="Pincelado y horneado de pan pretzel" 
            class="w-full h-full object-cover object-center filter contrast-[1.04]"
            loading="lazy"
          />
        </div>
        <div class="mt-5 space-y-2">
          <h3 class="font-serif text-lg sm:text-xl font-bold text-bone">
            Pan Pretzel • Hamburguesas
          </h3>
          <p class="font-sans text-xs sm:text-sm text-ash leading-relaxed">
            Miga estructurada de fermentación lenta con baño tradicional y barniz dorado de mantequilla antes del horno. Su textura sostiene los 160g de carne y los jugos sin desmoronarse, aportando un matiz salino y tostado característico.
          </p>
        </div>
      </div>

      <div>
        <div class="relative w-full aspect-[4/3] overflow-hidden">
          <img 
            src="/imagenes/historia/armando_hamburguesa.webp" 
            alt="Montaje artesanal de hamburguesas" 
            class="w-full h-full object-cover object-center filter contrast-[1.04]"
            loading="lazy"
          />
        </div>
        <div class="mt-5 space-y-2">
          <h3 class="font-serif text-lg sm:text-xl font-bold text-bone">
            Pan Brioche • Perros Calientes
          </h3>
          <p class="font-sans text-xs sm:text-sm text-ash leading-relaxed">
            Masa esponjosa enriquecida con mantequilla de vaca y dorada a la plancha al momento del servicio. Desarrollada para fundirse con el chorizo santarrosano, la salchicha americana y las salsas de la casa.
          </p>
        </div>
      </div>
    </div>
  </article>

  <!-- LÍNEA DIVISORIA SUIZA -->
  <div class="w-full max-w-5xl mx-auto px-5 sm:px-8 md:px-12">
    <div class="w-full border-t hairline-dark"></div>
  </div>

  <!-- 3. CAPÍTULO: LA EXPERIENCIA ÍNTIMA Y EL CANDADO DE AMOR -->
  <article class="max-w-5xl mx-auto px-5 sm:px-8 md:px-12 py-16 sm:py-24 md:py-32">
    <div class="grid grid-cols-1 md:grid-cols-12 gap-10 md:gap-16 items-center">
      <div class="md:col-span-5 order-2 md:order-1 space-y-5 sm:space-y-6">
        <span class="font-mono text-[11px] sm:text-xs text-gold uppercase tracking-[0.22em] block">
          Tradición & Espacio
        </span>
        <h2 class="font-flourish italic text-3xl sm:text-4xl md:text-5xl text-bone leading-[1.12] tracking-tight">
          «Lo Íntimo Hecho para Compartir»
        </h2>
        <p class="font-sans text-sm sm:text-base text-ash leading-relaxed">
          El salón de Mías fue proyectado como un comedor boutique: terciopelo escarlata, luz ámbar cálida filtrada por candelabros de cristal y un ambiente sereno para comer sin urgencia.
        </p>
        <p class="font-sans text-sm sm:text-base text-ash leading-relaxed">
          Inspirados en la tradición de los puentes europeos, en la reja del restaurante los comensales dejan candados de latón para sellar pactos, celebraciones y momentos especiales. Cada candado que descansa en la reja forma parte del patrimonio vivo de la casa.
        </p>
      </div>

      <div class="md:col-span-7 order-1 md:order-2 space-y-4">
        <div class="relative w-full aspect-[16/11] overflow-hidden">
          <img 
            src="/imagenes/local/restaurante_mias_desde_dentro.webp" 
            alt="Interior boutique de Mías Hamburguesas de Lujo" 
            class="w-full h-full object-cover object-center filter contrast-[1.05]"
            loading="lazy"
          />
        </div>
        <div class="flex items-center gap-3 text-xs font-mono text-ash/80 pt-1">
          <img 
            src="/imagenes/historia/candado_branding_historia.webp" 
            alt="Detalle candado de amor" 
            class="w-8 h-8 object-cover rounded-full border hairline-dark shrink-0"
          />
          <span>Salón principal • Terciopelo escarlata y candados de latón grabados por comensales</span>
        </div>
      </div>
    </div>
  </article>

  <!-- LÍNEA DIVISORIA SUIZA -->
  <div class="w-full max-w-5xl mx-auto px-5 sm:px-8 md:px-12">
    <div class="w-full border-t hairline-dark"></div>
  </div>

  <!-- 4. CAPÍTULO: HORARIOS DE ATENCIÓN Y SERVICIO -->
  <article class="max-w-5xl mx-auto px-5 sm:px-8 md:px-12 pt-16 sm:pt-24 md:pt-32 pb-28 sm:pb-36">
    <div class="max-w-2xl mb-10 sm:mb-14">
      <span class="font-mono text-[11px] sm:text-xs text-gold uppercase tracking-[0.22em] block mb-3">
        Servicio en Sala y Domicilios
      </span>
      <h2 class="font-flourish italic text-3xl sm:text-4xl md:text-5xl lg:text-6xl text-bone leading-[1.12] tracking-tight">
        Horarios de Atención
      </h2>
      <p class="font-sans text-sm sm:text-base text-ash mt-3 leading-relaxed">
        Servicio continuo de cocina al carbón, atención en sala y despacho directo de domicilios.
      </p>
    </div>

    <!-- Lista editorial tipográfica limpia (sin cajas, sin tarjetas) -->
    <div class="border-t border-b hairline-dark divide-y hairline-dark my-8">
      <div class="py-6 sm:py-7 flex flex-col sm:flex-row sm:items-baseline justify-between gap-2 sm:gap-6">
        <div>
          <h3 class="font-serif text-lg sm:text-xl font-bold text-bone">
            Lunes a Viernes
          </h3>
          <p class="font-sans text-xs sm:text-sm text-ash mt-0.5">
            Servicio en salón y despacho de domicilios
          </p>
        </div>
        <div class="font-mono text-base sm:text-lg font-bold text-gold shrink-0">
          3:00 PM – 9:00 PM
        </div>
      </div>

      <div class="py-6 sm:py-7 flex flex-col sm:flex-row sm:items-baseline justify-between gap-2 sm:gap-6">
        <div>
          <h3 class="font-serif text-lg sm:text-xl font-bold text-bone">
            Sábados y Domingos
          </h3>
          <p class="font-sans text-xs sm:text-sm text-ash mt-0.5">
            Jornada extendida de fin de semana
          </p>
        </div>
        <div class="font-mono text-base sm:text-lg font-bold text-gold shrink-0">
          3:00 PM – 10:00 PM
        </div>
      </div>

      <div class="py-6 sm:py-7 flex flex-col sm:flex-row sm:items-baseline justify-between gap-2 sm:gap-6">
        <div>
          <h3 class="font-serif text-lg sm:text-xl font-bold text-bone">
            Línea Directa de Domicilios
          </h3>
          <p class="font-sans text-xs sm:text-sm text-ash mt-0.5">
            Despacho coordinado vía WhatsApp
          </p>
        </div>
        <div class="font-mono text-base sm:text-lg font-bold text-bone shrink-0">
          +57 321 252 0881
        </div>
      </div>
    </div>

    <!-- Pie del capítulo con anclaje de retorno superior -->
    <div class="mt-12 sm:mt-16 flex flex-col sm:flex-row items-center justify-between gap-6 pt-4 text-xs font-mono text-ash">
      <div class="flex items-center gap-2.5">
        <span class="w-1.5 h-1.5 rounded-full bg-gold"></span>
        <span>Mías Hamburguesas de Lujo • Carta y Fidelización</span>
      </div>

      <button
        type="button"
        onclick={() => {
          if (typeof window !== 'undefined') {
            window.scrollTo({ top: 0, behavior: 'smooth' });
          }
        }}
        class="text-gold hover:text-bone transition-colors uppercase tracking-widest cursor-pointer py-2"
      >
        ↑ Volver a la Carta
      </button>
    </div>
  </article>
</section>

<!-- ═══════════════════════════════════════════════════════════════════════ -->
<!-- BOTÓN FLOTANTE "MIS PEDIDOS" EN PANTALLA HOME                        -->
<!-- ═══════════════════════════════════════════════════════════════════════ -->
<div 
  class="fixed bottom-5 inset-x-4 max-w-sm mx-auto z-40 pointer-events-auto select-none"
  style="padding-bottom: max(0px, env(safe-area-inset-bottom, 0px));"
>
  <button
    type="button"
    onclick={() => store.currentView = 'orders'}
    class="w-full min-h-[62px] sm:min-h-[66px] rounded-full bg-char-card/95 backdrop-blur-md border hairline-dark hover:border-gold/30 p-2.5 pl-4 pr-5 flex items-center justify-between shadow-lux active:scale-[0.98] transition-all cursor-pointer group"
    aria-label="Ver el estado de mis pedidos"
  >
    <div class="flex items-center gap-3.5 min-w-0">
      {#if latestActiveOrder}
        <!-- Pedido activo: Horno y estado editorial -->
        <div class="w-11 h-11 rounded-full bg-ember/15 border border-ember/30 flex items-center justify-center text-ember-light shrink-0">
          <Flame class="w-5 h-5" />
        </div>
        <div class="text-left min-w-0">
          <span class="font-mono text-[11px] font-bold text-gold tracking-wide block leading-tight">
            Comanda #{latestActiveOrder.code || 'ORD'}
          </span>
          <span class="font-sans font-bold text-sm text-bone truncate block mt-0.5">
            {#if latestActiveOrder.status === 'PREPARING'}
              En el horno al carbón
            {:else if latestActiveOrder.status === 'IN_DELIVERY'}
              En camino a tu dirección
            {:else if latestActiveOrder.status === 'READY'}
              Listo para retirar
            {:else}
              Comanda en preparación
            {/if}
          </span>
        </div>
      {:else}
        <!-- Sin pedido activo: Acceso al historial -->
        <div class="w-11 h-11 rounded-full bg-gold/10 border border-gold/20 flex items-center justify-center text-gold shrink-0">
          <Receipt class="w-5 h-5 stroke-[2]" />
        </div>
        <div class="text-left min-w-0">
          <span class="font-sans font-bold text-sm sm:text-base text-bone block leading-tight">
            Historial de Pedidos
          </span>
          <span class="font-sans text-xs text-ash truncate block mt-0.5">
            {hasAnyOrders ? 'Revisa tus comandas y pagos' : 'Consulta el estado de tu pedido'}
          </span>
        </div>
      {/if}
    </div>

    <div class="flex items-center gap-1.5 shrink-0">
      {#if latestActiveOrder}
        <span class="text-xs font-sans font-bold text-white bg-ember hover:bg-ember-light px-3.5 py-1.5 rounded-full transition-colors">
          Ver estado
        </span>
      {:else}
        <div class="w-9 h-9 rounded-full bg-char border hairline-dark flex items-center justify-center text-ash group-hover:text-bone group-hover:border-gold/30 group-hover:translate-x-0.5 transition-all">
          <ArrowRight class="w-4 h-4 stroke-[2]" />
        </div>
      {/if}
    </div>
  </button>
</div>

<style>
  /* MICRO-INTERACCIÓN: CHEVRON MÓRFICO A FLECHA CON BORDE DORADO BISELADO */
  .category-arrow-circle {
    background: 
      linear-gradient(var(--arrow-circle-bg, #161619), var(--arrow-circle-bg, #161619)) padding-box,
      conic-gradient(
        from var(--gold-angle),
        #7A5212 0deg,
        #C99E47 45deg,
        #FFFDF0 90deg,
        #F0CA76 120deg,
        #C99E47 160deg,
        #7A5212 210deg,
        #4A3108 260deg,
        #8A601A 310deg,
        #7A5212 360deg
      ) border-box;
    border: 1.5px solid transparent !important;
    animation: gold-sheen-rotate 9s linear infinite;
    transform: translateZ(0);
    -webkit-backface-visibility: hidden;
    backface-visibility: hidden;
    will-change: transform;
    transition: transform 0.28s cubic-bezier(0.34, 1.4, 0.64, 1), 
                box-shadow 0.25s cubic-bezier(0.16, 1, 0.3, 1),
                color 0.24s ease;
  }

  .category-arrow-circle:not(.is-active) {
    --arrow-circle-bg: #161619;
    color: #FF2400;
    box-shadow: 
      0 2px 6px rgba(0, 0, 0, 0.35),
      0 6px 18px -4px rgba(197, 160, 89, 0.22),
      inset 0 1px 1px rgba(255, 245, 207, 0.16);
  }

  .category-arrow-circle.is-active {
    --arrow-circle-bg: #FF2400;
    color: #ffffff;
    box-shadow: 
      0 2px 8px rgba(0, 0, 0, 0.35),
      0 8px 24px -4px rgba(255, 36, 0, 0.40),
      0 0 16px rgba(201, 158, 71, 0.28),
      inset 0 1px 1px rgba(255, 245, 207, 0.35);
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

    /* El círculo crece sutilmente y se tiñe de rojo vivo con destello dorado */
    .category-pill:hover .category-arrow-circle:not(.is-active) {
      --arrow-circle-bg: #FF2400;
      color: #ffffff !important;
      transform: scale(1.06);
      box-shadow: 
        0 4px 14px rgba(0, 0, 0, 0.5),
        0 10px 28px -4px rgba(255, 36, 0, 0.40),
        0 0 16px rgba(201, 158, 71, 0.25),
        inset 0 1px 1px rgba(255, 245, 207, 0.35);
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
