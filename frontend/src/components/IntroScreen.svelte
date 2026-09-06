<script>
  import { onMount } from 'svelte';
  import gsap from 'gsap';
  import { ChevronRight, X, Clock, ShieldCheck, Flame, FileText } from 'lucide-svelte';

  let { onComplete } = $props();
  let introContainer = $state(null);
  let canvasEl = $state(null);
  let isVisible = $state(true);
  let isInfoModalOpen = $state(false);

  let animFrameId = null;
  let cleanupGL = null;

  onMount(() => {
    // Animación de entrada suave y cinematográfica
    const tl = gsap.timeline();

    tl.fromTo('.welcome-grid', 
      { opacity: 0 }, 
      { opacity: 1, duration: 1.2, ease: 'power2.out' }
    )
    .fromTo('.welcome-sidebar', 
      { opacity: 0, x: -10 }, 
      { opacity: 0.45, x: 0, duration: 0.8, ease: 'power2.out' }, 
      '-=0.8'
    )
    .fromTo('.welcome-kicker', 
      { opacity: 0, y: 15 }, 
      { opacity: 0.85, y: 0, duration: 0.7, ease: 'power2.out' }, 
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

    // Inicialización del shader WebGL de brasa y fuego líquido continuo
    if (canvasEl) {
      cleanupGL = initEmberShader(canvasEl);
    }

    return () => {
      if (cleanupGL) cleanupGL();
    };
  });

  function initEmberShader(canvas) {
    const gl = canvas.getContext('webgl', { alpha: false, antialias: false });
    if (!gl) return null;

    const vsSource = `
      attribute vec2 position;
      void main() {
        gl_Position = vec4(position, 0.0, 1.0);
      }
    `;

    const fsSource = `
      precision highp float;
      uniform vec2 u_resolution;
      uniform float u_time;

      // Paleta artesanal de brasa y magma al carbón: cero tonos rosas
      vec3 emberPalette(float t) {
        t = clamp(t, 0.0, 1.0);
        
        vec3 c0 = vec3(0.020, 0.020, 0.027); // #050507 negro obsidiana
        vec3 c1 = vec3(0.180, 0.024, 0.006); // #2e0601 carbón incandescente oscuro
        vec3 c2 = vec3(0.460, 0.060, 0.010); // #750f02 rojo brasa profundo
        vec3 c3 = vec3(0.780, 0.160, 0.014); // #c72904 bermellón vivo
        vec3 c4 = vec3(0.970, 0.400, 0.030); // #f76608 naranja fuego ardiente
        vec3 c5 = vec3(1.000, 0.700, 0.160); // #ffb329 ámbar dorado
        vec3 c6 = vec3(1.000, 0.900, 0.600); // #ffe699 crema incandescente cálida
        vec3 c7 = vec3(1.000, 0.980, 0.920); // #fffcf0 núcleo blanco térmico

        if (t < 0.12) {
          return mix(c0, c1, t / 0.12);
        } else if (t < 0.28) {
          return mix(c1, c2, (t - 0.12) / 0.16);
        } else if (t < 0.46) {
          return mix(c2, c3, (t - 0.28) / 0.18);
        } else if (t < 0.64) {
          return mix(c3, c4, (t - 0.46) / 0.18);
        } else if (t < 0.78) {
          return mix(c4, c5, (t - 0.64) / 0.14);
        } else if (t < 0.90) {
          return mix(c5, c6, (t - 0.78) / 0.12);
        } else {
          return mix(c6, c7, (t - 0.90) / 0.10);
        }
      }

      void main() {
        vec2 uv = gl_FragCoord.xy / u_resolution.xy;

        // Ritmo lento, uniforme y sedoso (brasa y fuego lento)
        float speed = 0.12;
        float t = u_time * speed;

        // Flujo horizontal continuo y uniforme hacia la derecha de la pantalla (+x)
        float flow = uv.x * 2.4 - (t * 0.28);

        // Onda sinuosa orgánica de brasa que fluye a la derecha
        float w1 = flow * 2.8;
        float crest = 0.17 
                    + 0.12 * sin(w1) 
                    + 0.05 * sin(w1 * 2.0 - 0.8) 
                    + 0.02 * cos(w1 * 3.0 + 0.4);

        // Micro-ondulaciones térmicas fluidas (100% armónicas e infinitas sin cortes)
        float micro = sin(flow * 6.5 + uv.y * 3.5) * 0.015 
                    + cos(flow * 11.0 - uv.y * 5.0) * 0.008;
        crest += micro;

        float dy = uv.y - crest;

        // Perfil térmico asimétrico sedoso
        float heat = 0.0;
        if (dy > 0.0) {
          float dUp = dy / 0.13;
          heat = exp(-dUp * dUp * 2.6);
        } else {
          float dDown = -dy / 0.34;
          heat = exp(-dDown * dDown * 1.5);
        }

        // Núcleo incandescente térmico sobre la cresta
        float coreDist = abs(dy) / 0.055;
        float coreGlow = exp(-coreDist * coreDist * 2.8);

        // Capa armónica profunda de magma fluido
        float w2 = (uv.x * 2.4 - t * 0.20) * 1.8;
        float crest2 = 0.11 + 0.09 * sin(w2 + 1.2) + 0.03 * cos(w2 * 2.0);
        float dy2 = uv.y - crest2;
        float heat2 = exp(-(dy2 * dy2) / 0.05) * 0.35;

        float flameHeat = heat * 0.70 + coreGlow * 0.46 + heat2 * 0.30;

        // Delimitación espacial: "solo en la parte donde esta actualmente"
        // 1. Desvanecimiento superior: la mitad superior queda en negro puro (#050507) para retícula, estrella y títulos
        float topFade = smoothstep(0.58, 0.18, uv.y);

        // 2. Presencia hacia el flanco derecho: garantiza que 'Hamburguesas', 'de lujo' y 'del carbón'
        // mantengan legibilidad absoluta con alto contraste
        float rightPresence = smoothstep(0.20, 0.80, uv.x);

        // 3. Fondo de brasa cálida en la base inferior
        float bottomBase = exp(-uv.y * 3.2) * smoothstep(0.15, 0.85, uv.x) * 0.35;

        // 4. Resplandor oscuro en la esquina inferior izquierda: carbón profundo, CERO tonos rosas
        float cornerEmber = exp(-((uv.x * 2.5) * (uv.x * 2.5) + (uv.y * 3.0) * (uv.y * 3.0))) * 0.18;

        float totalHeat = (flameHeat * topFade * rightPresence) + (bottomBase * topFade) + (cornerEmber * 0.30);
        totalHeat = clamp(totalHeat, 0.0, 1.0);

        gl_FragColor = vec4(emberPalette(totalHeat), 1.0);
      }
    `;

    function compile(type, src) {
      const s = gl.createShader(type);
      gl.shaderSource(s, src);
      gl.compileShader(s);
      return s;
    }

    const vs = compile(gl.VERTEX_SHADER, vsSource);
    const fs = compile(gl.FRAGMENT_SHADER, fsSource);
    const prog = gl.createProgram();
    gl.attachShader(prog, vs);
    gl.attachShader(prog, fs);
    gl.linkProgram(prog);

    const buf = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, buf);
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([
      -1, -1, 1, -1, -1, 1,
      -1, 1, 1, -1, 1, 1
    ]), gl.STATIC_DRAW);

    const aPos = gl.getAttribLocation(prog, 'position');
    gl.enableVertexAttribArray(aPos);
    gl.vertexAttribPointer(aPos, 2, gl.FLOAT, false, 0, 0);

    const uRes = gl.getUniformLocation(prog, 'u_resolution');
    const uTime = gl.getUniformLocation(prog, 'u_time');

    function resize() {
      if (!canvas) return;
      const dpr = Math.min(window.devicePixelRatio || 1, 2);
      const w = canvas.parentElement ? canvas.parentElement.clientWidth : window.innerWidth;
      const h = canvas.parentElement ? canvas.parentElement.clientHeight : window.innerHeight;
      canvas.width = w * dpr;
      canvas.height = h * dpr;
      gl.viewport(0, 0, canvas.width, canvas.height);
    }
    resize();
    window.addEventListener('resize', resize);

    const startTime = performance.now();
    function loop() {
      if (!gl || !canvas) return;
      gl.useProgram(prog);
      gl.uniform2f(uRes, canvas.width, canvas.height);
      gl.uniform1f(uTime, (performance.now() - startTime) / 1000);
      gl.drawArrays(gl.TRIANGLES, 0, 6);
      animFrameId = requestAnimationFrame(loop);
    }
    animFrameId = requestAnimationFrame(loop);

    return () => {
      if (animFrameId) cancelAnimationFrame(animFrameId);
      window.removeEventListener('resize', resize);
      if (gl) {
        gl.deleteBuffer(buf);
        gl.deleteProgram(prog);
        gl.deleteShader(vs);
        gl.deleteShader(fs);
      }
    };
  }

  function handleAdvance(e) {
    if (e) e.stopPropagation();
    if (isInfoModalOpen) return;
    if (!isVisible) return;

    if (cleanupGL) {
      cleanupGL();
      cleanupGL = null;
    }

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

  function toggleInfoModal(e) {
    if (e) e.stopPropagation();
    isInfoModalOpen = !isInfoModalOpen;
  }
</script>

{#if isVisible}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div 
    bind:this={introContainer}
    onclick={handleAdvance}
    class="fixed inset-0 z-50 flex flex-col justify-between select-none cursor-pointer overflow-hidden p-6 sm:p-10 font-sans text-white bg-[#050507]"
  >
    <!-- FONDO DE BRASA Y FUEGO LENTO AL CARBÓN (100% GPU WebGL - Cero cortes, Infinito) -->
    <div class="absolute inset-0 overflow-hidden pointer-events-none z-0 bg-[#050507]">
      <!-- Canvas de flujo de brasa y magma hacia la derecha -->
      <canvas bind:this={canvasEl} class="absolute inset-0 w-full h-full pointer-events-none z-0"></canvas>

      <!-- Grano sutil analógico / Textura de ruido cinemática Awwwards (elimina banding y da acabado sedoso) -->
      <svg class="pointer-events-none absolute inset-0 w-full h-full opacity-[0.045] mix-blend-overlay z-[2]" xmlns="http://www.w3.org/2000/svg">
        <filter id="aurora-grain">
          <feTurbulence type="fractalNoise" baseFrequency="0.8" numOctaves="3" stitchTiles="stitch" />
          <feColorMatrix type="saturate" values="0" />
        </filter>
        <rect width="100%" height="100%" filter="url(#aurora-grain)" />
      </svg>
    </div>

    <!-- 1. RETÍCULA ARQUITECTÓNICA CON UNIÓN CURVA (FLARED FILLET) E ISOTIPO -->
    <div class="welcome-grid absolute inset-0 z-10 pointer-events-none">
      <!-- Línea Vertical Superior (desde el borde superior hasta la punta superior de la estrella) -->
      <div 
        class="absolute w-[1px] bg-white/[0.12] -translate-x-1/2"
        style="left: 50%; top: 0; height: calc(28% - 44px);"
      ></div>

      <!-- Línea Vertical Inferior (se desvanece gradualmente hacia la zona media) -->
      <div 
        class="absolute w-[1px] -translate-x-1/2 pointer-events-none"
        style="
          left: 50%; 
          top: calc(28% + 44px); 
          height: 24vh; 
          background: linear-gradient(180deg, rgba(255, 255, 255, 0.12) 0%, rgba(255, 255, 255, 0.04) 55%, transparent 100%);
        "
      ></div>

      <!-- Línea Horizontal Izquierda (desde el borde izquierdo hasta la punta de la estrella) -->
      <div 
        class="absolute h-[1px] bg-white/[0.12] -translate-y-1/2"
        style="top: 28%; left: 0; width: calc(50% - 44px);"
      ></div>

      <!-- Línea Horizontal Derecha (desde la punta derecha de la estrella hacia el borde derecho) -->
      <div 
        class="absolute h-[1px] bg-white/[0.08] -translate-y-1/2"
        style="top: 28%; left: calc(50% + 44px); right: 0;"
      ></div>

      <!-- Unión Curva Flared (Estrella Astroid cóncava R=44px, 88x88px) con el Isotipo Mías en su centro -->
      <div 
        class="absolute -translate-x-1/2 -translate-y-1/2 pointer-events-none flex items-center justify-center"
        style="left: 50%; top: 28%; width: 88px; height: 88px;"
      >
        <svg 
          viewBox="0 0 88 88" 
          class="w-full h-full absolute inset-0 overflow-visible" 
          fill="none" 
          xmlns="http://www.w3.org/2000/svg"
        >
          <!-- Curvas cóncavas de empalme continuo (R=44px) que forman la estrella de cuatro puntas -->
          <path 
            d="M 44 0 A 44 44 0 0 1 0 44 A 44 44 0 0 1 44 88 A 44 44 0 0 1 88 44 A 44 44 0 0 1 44 0 Z" 
            stroke="rgba(255,255,255,0.18)" 
            stroke-width="1"
            fill="rgba(255,255,255,0.015)"
          />
        </svg>

        <!-- Isotipo del Candado de Corazón Mías calibrado con holgura fina -->
        <img 
          src="/imagenes/marca/isotipo.webp" 
          alt="Mías" 
          class="w-[48px] h-[48px] object-contain opacity-95 brightness-0 invert select-none drop-shadow-[0_2px_10px_rgba(0,0,0,0.5)]"
        />
      </div>
    </div>

    <!-- 2. MARGEN IZQUIERDO: COORDENADAS VERTICALES ("2026" Y "MÍAS") FLANQUEANDO EL EJE HORIZONTAL -->
    <div 
      class="welcome-sidebar absolute left-4 sm:left-6 z-10 flex flex-col items-center pointer-events-none select-none text-[9px] font-mono tracking-[0.28em] text-white/45 uppercase"
      style="top: 28%; transform: translateY(-50%);"
    >
      <span class="[writing-mode:vertical-rl] rotate-180 mb-3.5">2026</span>
      <span class="[writing-mode:vertical-rl] rotate-180 mt-3.5">MÍAS</span>
    </div>

    <!-- 4. ESPACIO SUPERIOR DE RESPIRACIÓN EDITORIAL -->
    <div class="flex-1 min-h-[10vh]"></div>

    <!-- 5. BLOQUE DE TITULAR: ELEVADO A LA MITAD DEL CUADRANTE INFERIOR -->
    <div class="relative z-10 w-full max-w-xl mb-14 sm:mb-20">
      <!-- Subtítulo Kicker Bienvenido -->
      <p class="welcome-kicker font-display text-sm sm:text-base font-normal text-white/85 tracking-[0.2em] uppercase mb-2 sm:mb-3">
        Bienvenido
      </p>

      <!-- Titular Editorial con tipografía intacta y cápsulas estilizadas idénticas a la referencia -->
      <h1 class="welcome-title font-display text-[2.35rem] sm:text-5xl lg:text-6xl tracking-tight text-white leading-[1.26] sm:leading-[1.3] select-none">
        <span class="font-semibold text-white">Hamburguesas</span>
        <!-- Cápsula 1: Alargada y estilizada -->
        <span class="inline-block w-14 sm:w-16 h-6 sm:h-7 rounded-full border border-white/35 bg-white/[0.03] backdrop-blur-[2px] align-middle mx-1.5"></span>
        <br />
        <!-- de lujo: Mantenido intacto con Pinyon Script -->
        <span 
          class="inline-block align-baseline mr-1.5 pr-0.5 text-[#FFF5EA] text-[3.2rem] sm:text-6xl lg:text-7xl font-normal drop-shadow-[0_2px_12px_rgba(0,0,0,0.35)]"
          style="font-family: 'Pinyon Script', cursive; font-style: normal;"
        >de lujo</span>
        <!-- Cápsula 2: Alargada como en la referencia -->
        <span class="inline-block w-12 sm:w-14 h-6 sm:h-7 rounded-full border border-white/35 bg-white/[0.03] backdrop-blur-[2px] align-middle mx-1.5"></span> 
        <span class="font-light text-white/80">al fuego</span>
        <br />
        <span class="font-semibold text-white">del carbón</span> 
        <!-- Cápsula 3: Alargada -->
        <span class="inline-block w-14 sm:w-16 h-6 sm:h-7 rounded-full border border-white/35 bg-white/[0.03] backdrop-blur-[2px] align-middle mx-1.5"></span>
      </h1>
    </div>

    <!-- 6. CONTROLES INFERIORES: PLANOS Y ALINEADOS EN LA BASE -->
    <div class="welcome-controls relative z-10 flex items-center justify-between pb-3 sm:pb-5">
      <!-- Botón Izquierdo: Vidrio esmerilado suave translúcido con isotipo candado -->
      <button 
        type="button"
        onclick={toggleInfoModal}
        class="w-14 h-14 rounded-full border-0 outline-none bg-white/15 backdrop-blur-xl flex items-center justify-center text-white/90 hover:bg-white/25 active:scale-95 transition-all cursor-pointer group shadow-lg shadow-black/15"
        title="Información, Horarios y Términos Mías"
        aria-label="Ver información y términos"
      >
        <img 
          src="/imagenes/marca/isotipo.webp" 
          alt="Mías Candado de Corazón" 
          class="w-6 h-6 object-contain brightness-0 invert opacity-90 group-hover:scale-110 transition-transform"
        />
      </button>

      <!-- Botón Derecho: Contenedor con halo de luz cálida pulsante y círculo de avance -->
      <div class="relative flex items-center justify-center">
        <!-- 1. Halo expansivo de respiración cálida (Aura exterior difusa de brasa) -->
        <div class="advance-halo-aura pointer-events-none absolute -inset-3 rounded-full"></div>

        <!-- 2. Anillo de pulso beacon sutil (onda de llamada a la acción) -->
        <div class="advance-halo-ring pointer-events-none absolute -inset-1.5 rounded-full"></div>

        <!-- 3. Botón de avance blanco de alto contraste -->
        <button 
          type="button"
          onclick={handleAdvance}
          class="relative z-10 w-14 h-14 rounded-full bg-white text-black flex items-center justify-center hover:scale-105 active:scale-95 transition-all cursor-pointer group shadow-[0_4px_24px_rgba(0,0,0,0.40)]"
          aria-label="Ingresar al restaurante"
        >
          <ChevronRight class="w-6 h-6 stroke-[2.2] text-black group-hover:translate-x-0.5 transition-transform" />
        </button>
      </div>
    </div>
  </div>

  <!-- 6. MODAL REFINADO DE INFORMACIÓN, HORARIOS Y TÉRMINOS (SE ABRE DESDE EL BOTÓN IZQUIERDO) -->
  {#if isInfoModalOpen}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div 
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-md p-4 animate-in fade-in duration-200"
      onclick={(e) => { e.stopPropagation(); isInfoModalOpen = false; }}
    >
      <div 
        class="bg-[#121114] border border-white/10 text-white w-full max-w-md rounded-3xl p-6 sm:p-8 shadow-2xl space-y-6 relative"
        onclick={(e) => e.stopPropagation()}
      >
        <!-- Botón Cerrar -->
        <button 
          type="button"
          onclick={() => isInfoModalOpen = false}
          class="absolute top-5 right-5 p-2 rounded-full hover:bg-white/10 text-white/60 hover:text-white transition-colors"
          aria-label="Cerrar modal de información"
        >
          <X class="w-5 h-5" />
        </button>

        <!-- Cabecera de la Marca -->
        <div class="flex items-center gap-3 border-b border-white/10 pb-4">
          <img 
            src="/imagenes/marca/isotipo.webp" 
            alt="Mías" 
            class="w-10 h-10 object-contain brightness-0 invert"
          />
          <div>
            <h3 class="font-serif text-xl font-bold tracking-wide">Mías Hamburguesas de Lujo</h3>
            <p class="font-sans text-xs text-white/50">Lo Íntimo Hecho para Compartir</p>
          </div>
        </div>

        <!-- Secciones de Información Clave -->
        <div class="space-y-4 text-xs font-sans leading-relaxed text-white/80">
          <div class="flex items-start gap-3 p-3 rounded-2xl bg-white/[0.04]">
            <Clock class="w-4 h-4 text-[#FF9E2C] shrink-0 mt-0.5" />
            <div>
              <p class="font-semibold text-white">Horarios Oficiales de Atención</p>
              <p class="text-white/60 mt-0.5">Sábado a Jueves: 3:00 PM – 9:30 PM<br />Viernes: Cerrado por producción artesanal.</p>
            </div>
          </div>

          <div class="flex items-start gap-3 p-3 rounded-2xl bg-white/[0.04]">
            <Flame class="w-4 h-4 text-[#FF5A1F] shrink-0 mt-0.5" />
            <div>
              <p class="font-semibold text-white">Nuestra Artesanía</p>
              <p class="text-white/60 mt-0.5">Cocción exclusiva al fuego en Horno «Rayo», 160g de carne de res premium sobre Pan Pretzel artesanal y perros calientes en Pan Brioche.</p>
            </div>
          </div>

          <div class="flex items-start gap-3 p-3 rounded-2xl bg-white/[0.04]">
            <ShieldCheck class="w-4 h-4 text-[#4ADE80] shrink-0 mt-0.5" />
            <div>
              <p class="font-semibold text-white">Tratamiento de Datos y Fidelización</p>
              <p class="text-white/60 mt-0.5">Cumplimiento estricto de la Ley 1581 de 2012 (Habeas Data). Programa de sellos por factura física (10 sellos = 1 Original gratis).</p>
            </div>
          </div>
        </div>

        <!-- Botón de acción para entrar -->
        <button 
          type="button"
          onclick={() => { isInfoModalOpen = false; handleAdvance(); }}
          class="w-full py-3.5 rounded-full bg-white text-black font-semibold text-xs uppercase tracking-wider hover:bg-white/90 transition-all shadow-lg"
        >
          Entrar al Restaurante
        </button>
      </div>
    </div>
  {/if}
{/if}

<style>
  /* 1. Aura cálida pulsante (resplandor de brasa suave y continuo) */
  .advance-halo-aura {
    background: radial-gradient(
      circle at center,
      rgba(255, 200, 80, 0.95) 0%,
      rgba(255, 125, 25, 0.75) 45%,
      rgba(230, 50, 10, 0.35) 70%,
      transparent 90%
    );
    filter: blur(12px);
    will-change: transform, opacity;
    animation: advance-aura-pulse 2.4s ease-in-out infinite alternate;
  }

  @keyframes advance-aura-pulse {
    0% {
      transform: scale(0.90);
      opacity: 0.50;
    }
    50% {
      transform: scale(1.42);
      opacity: 1.0;
    }
    100% {
      transform: scale(1.15);
      opacity: 0.75;
    }
  }

  /* 2. Anillo térmico de expansión suave (beacon micro-interactivo de brasa) */
  .advance-halo-ring {
    border: 1.5px solid rgba(255, 210, 90, 0.90);
    will-change: transform, opacity;
    animation: advance-beacon-wave 2.4s cubic-bezier(0.15, 0.85, 0.25, 1) infinite;
  }

  @keyframes advance-beacon-wave {
    0% {
      transform: scale(0.95);
      opacity: 0.90;
      border-color: rgba(255, 220, 110, 0.95);
      box-shadow: 0 0 10px rgba(255, 180, 40, 0.6);
    }
    60% {
      transform: scale(1.68);
      opacity: 0.25;
      border-color: rgba(255, 130, 25, 0.50);
      box-shadow: 0 0 18px rgba(255, 110, 15, 0.3);
    }
    100% {
      transform: scale(1.90);
      opacity: 0;
      border-color: transparent;
      box-shadow: none;
    }
  }
</style>

