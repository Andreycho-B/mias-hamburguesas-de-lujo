<script>
  import { onMount } from 'svelte';
  import gsap from 'gsap';
  import { ChevronRight } from 'lucide-svelte';

  let { onComplete } = $props();
  let introContainer = $state(null);
  let isVisible = $state(true);

  onMount(() => {
    // Animación de entrada suave y cinematográfica
    const tl = gsap.timeline();

    tl.fromTo('.welcome-grid', 
      { opacity: 0 }, 
      { opacity: 1, duration: 1.2, ease: 'power2.out' }
    )
    .fromTo('.welcome-sidebar', 
      { opacity: 0, x: -10 }, 
      { opacity: 0.4, x: 0, duration: 0.8, ease: 'power2.out' }, 
      '-=0.8'
    )
    .fromTo('.welcome-kicker', 
      { opacity: 0, y: 15 }, 
      { opacity: 0.8, y: 0, duration: 0.7, ease: 'power2.out' }, 
      '-=0.6'
    )
    .fromTo('.welcome-title', 
      { opacity: 0, y: 20 }, 
      { opacity: 1, y: 0, duration: 0.9, ease: 'power3.out' }, 
      '-=0.5'
    )
    .fromTo('.welcome-controls', 
      { opacity: 0, y: 15 }, 
      { opacity: 1, y: 0, duration: 0.8, ease: 'power2.out' }, 
      '-=0.4'
    );
  });

  function handleAdvance(e) {
    if (e) e.stopPropagation();
    if (!isVisible) return;

    gsap.to(introContainer, {
      opacity: 0,
      scale: 1.02,
      duration: 0.5,
      ease: 'power2.inOut',
      onComplete: () => {
        isVisible = false;
        onComplete();
      }
    });
  }
</script>

{#if isVisible}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div 
    bind:this={introContainer}
    onclick={handleAdvance}
    class="fixed inset-0 z-50 flex flex-col justify-between select-none cursor-pointer overflow-hidden p-6 sm:p-10 font-sans text-white"
    style="
      background-color: #0c0b0e;
      background-image: 
        radial-gradient(ellipse 95% 85% at 85% 90%, #FF5A1F 0%, #D8310E 25%, #7D1507 50%, rgba(20, 6, 8, 0.5) 75%, transparent 100%),
        radial-gradient(ellipse 60% 50% at 95% 95%, #FFA53B 0%, #FF6619 40%, transparent 80%),
        linear-gradient(135deg, #09090B 0%, #0E0D12 40%, #1A0907 70%, #3D0E07 100%);
    "
  >
    <!-- 1. RETÍCULA ARQUITECTÓNICA CON ESTRELLA CÓNCAVA CENTRAL -->
    <div class="welcome-grid absolute inset-0 pointer-events-none">
      <!-- Líneas de Cuadrantes y Estrella Cóncava Vectorial -->
      <svg class="w-full h-full" xmlns="http://www.w3.org/2000/svg" preserveAspectRatio="none">
        <defs>
          <filter id="subtle-glow" x="-20%" y="-20%" width="140%" height="140%">
            <feGaussianBlur stdDeviation="1" result="blur" />
            <feMerge>
              <feMergeNode in="blur" />
              <feMergeNode in="SourceGraphic" />
            </feMerge>
          </filter>
        </defs>

        <!-- Retícula responsive: Centro de la estrella al 50% horizontal y 36% vertical -->
        <!-- Línea Vertical Superior -->
        <line x1="50%" y1="0%" x2="50%" y2="calc(36% - 18px)" stroke="rgba(255,255,255,0.10)" stroke-width="1" />
        <!-- Línea Vertical Inferior -->
        <line x1="50%" y1="calc(36% + 18px)" x2="50%" y2="100%" stroke="rgba(255,255,255,0.08)" stroke-width="1" />
        <!-- Línea Horizontal Izquierda -->
        <line x1="0%" y1="36%" x2="calc(50% - 18px)" y2="36%" stroke="rgba(255,255,255,0.10)" stroke-width="1" />
        <!-- Línea Horizontal Derecha -->
        <line x1="calc(50% + 18px)" y1="36%" x2="100%" y2="36%" stroke="rgba(255,255,255,0.08)" stroke-width="1" />

        <!-- Estrella Cóncava de 4 puntas en la Intersección (cx=50%, cy=36%, radio=18px) -->
        <g transform="translate(0, 0)">
          <!-- Trayecto de la estrella de autor: 4 arcos cóncavos hacia el centro -->
          <!-- Usamos un elemento centrado en coordenadas relativas mediante SVG anidado -->
          <svg x="50%" y="36%" overflow="visible">
            <path 
              d="M 0 -18 
                 Q 0 0 18 0 
                 Q 0 0 0 18 
                 Q 0 0 -18 0 
                 Q 0 0 0 -18 Z" 
              fill="none" 
              stroke="rgba(255,255,255,0.22)" 
              stroke-width="1.2"
              filter="url(#subtle-glow)"
            />
            <!-- Centro estelar sutil -->
            <circle cx="0" cy="0" r="1.5" fill="rgba(255,255,255,0.4)" />
          </svg>
        </g>
      </svg>
    </div>

    <!-- 2. MARGEN IZQUIERDO: COORDENADAS Y TIPOGRAFÍA VERTICAL -->
    <div class="welcome-sidebar absolute left-4 sm:left-6 top-16 sm:top-20 flex flex-col gap-6 text-[10px] font-mono tracking-[0.25em] text-white/40 pointer-events-none select-none uppercase">
      <span class="[writing-mode:vertical-rl] rotate-180">1260</span>
      <span class="[writing-mode:vertical-rl] rotate-180">MÍAS</span>
    </div>

    <!-- 3. ZONA SUPERIOR VACÍA PARA RESPIRACIÓN DE RETÍCULA -->
    <div class="relative z-10 pt-4"></div>

    <!-- 4. ZONA INFERIOR: TITULAR CON CÁPSULAS CONTORNEADAS EMBEBIDAS -->
    <div class="relative z-10 max-w-lg mb-8 sm:mb-12">
      <!-- Subtítulo Kicker -->
      <p class="welcome-kicker font-sans text-xs sm:text-sm font-normal text-white/70 tracking-wider mb-3">
        Bienvenido
      </p>

      <!-- Titular Editorial con Cápsulas de Autor -->
      <h1 class="welcome-title font-sans text-3xl sm:text-5xl font-normal tracking-tight text-white leading-[1.35] select-none">
        Hamburguesas 
        <span class="inline-block w-8 sm:w-11 h-4 sm:h-5 rounded-full border border-white/40 align-middle mx-1 bg-white/[0.03] backdrop-blur-[1px]"></span>
        <br />
        de lujo 
        <span class="inline-block w-8 sm:w-11 h-4 sm:h-5 rounded-full border border-white/40 align-middle mx-1 bg-white/[0.03] backdrop-blur-[1px]"></span> 
        al fuego
        <br />
        del carbón 
        <span class="inline-block w-8 sm:w-11 h-4 sm:h-5 rounded-full border border-white/40 align-middle mx-1 bg-white/[0.03] backdrop-blur-[1px]"></span>
      </h1>
    </div>

    <!-- 5. CONTROLES INFERIORES: VIDRIO ESMERILADO (IZQ) Y CÍRCULO BLANCO DE AVANCE (DER) -->
    <div class="welcome-controls relative z-10 flex items-center justify-between pt-2">
      <!-- Botón Izquierdo: Vidrio esmerilado con Isotipo del Candado de Corazón -->
      <div 
        class="w-12 h-12 rounded-full bg-white/10 backdrop-blur-md border border-white/20 flex items-center justify-center text-white/90 shadow-lg hover:bg-white/15 transition-all"
        title="Mías — Lo Íntimo Hecho para Compartir"
      >
        <img 
          src="/imagenes/marca/isotipo.webp" 
          alt="Mías Candado de Corazón" 
          class="w-5 h-5 object-contain brightness-0 invert opacity-80"
        />
      </div>

      <!-- Botón Derecho: Círculo blanco de alto contraste con flecha de avance -->
      <button 
        type="button"
        onclick={handleAdvance}
        class="w-14 h-14 rounded-full bg-white text-black flex items-center justify-center shadow-[0_10px_30px_rgba(0,0,0,0.5)] hover:scale-105 active:scale-95 transition-all cursor-pointer group"
        aria-label="Ingresar al restaurante"
      >
        <ChevronRight class="w-6 h-6 stroke-[2.2] text-black group-hover:translate-x-0.5 transition-transform" />
      </button>
    </div>
  </div>
{/if}
