<script>
  import { onMount } from 'svelte';
  import gsap from 'gsap';
  import { ChevronRight, X } from 'lucide-svelte';

  let { onComplete } = $props();
  let introContainer = $state(null);
  let canvasEl = $state(null);
  let isVisible = $state(true);
  let isInfoModalOpen = $state(false);

  let windowWidth = $state(typeof window !== 'undefined' ? window.innerWidth : 1200);
  let windowHeight = $state(typeof window !== 'undefined' ? window.innerHeight : 800);

  let isShortLandscape = $derived(windowHeight < 560 && windowWidth > windowHeight);
  let isDesktop = $derived(windowWidth >= 1024 && windowHeight >= 650);
  let cyRatio = $derived(isShortLandscape ? 0.19 : isDesktop ? 0.22 : 0.28);
  let cx = $derived(windowWidth * 0.5);
  let cy = $derived(windowHeight * cyRatio);
  const r = 44;
  let bottomReach = $derived(
    isShortLandscape 
      ? Math.min(windowHeight * 0.18, 65) 
      : isDesktop 
        ? Math.min(windowHeight * 0.22, 170) 
        : Math.min(windowHeight * 0.28, 240)
  );

  let pathBase1 = $state(null);
  let pathSparkCore1 = $state(null);
  let pathSparkGlow1 = $state(null);

  let pathBase2 = $state(null);
  let pathSparkCore2 = $state(null);
  let pathSparkGlow2 = $state(null);

  let pathBase3 = $state(null);
  let pathSparkCore3 = $state(null);
  let pathSparkGlow3 = $state(null);

  let pathBase4 = $state(null);
  let pathSparkCore4 = $state(null);
  let pathSparkGlow4 = $state(null);

  // 1. Estrella central: Asteroide cerrado continuo de 4 puntas perfectamente simétrico (sellado final uniforme)
  let astroid_d = $derived(`M ${cx} ${cy - r} A ${r} ${r} 0 0 1 ${cx - r} ${cy} A ${r} ${r} 0 0 1 ${cx} ${cy + r} A ${r} ${r} 0 0 1 ${cx + r} ${cy} A ${r} ${r} 0 0 1 ${cx} ${cy - r} Z`);

  // 2. Cuatro trayectorias cinemáticas continuas: convergen desde los bordes y trazan con sus estelas cada cuadrante de la estrella
  let path1_d = $derived(`M ${cx} 0 L ${cx} ${cy - r} A ${r} ${r} 0 0 1 ${cx - r} ${cy}`);
  let path2_d = $derived(`M 0 ${cy} L ${cx - r} ${cy} A ${r} ${r} 0 0 1 ${cx} ${cy + r}`);
  let path3_d = $derived(`M ${cx} ${cy + r + bottomReach} L ${cx} ${cy + r} A ${r} ${r} 0 0 1 ${cx + r} ${cy}`);
  let path4_d = $derived(`M ${windowWidth} ${cy} L ${cx + r} ${cy} A ${r} ${r} 0 0 1 ${cx} ${cy - r}`);

  let animFrameId = null;
  let cleanupGL = null;
  let cleanupRestrictions = null;

  onMount(() => {
    const origHtmlOverflow = document.documentElement.style.overflow;
    const origBodyOverflow = document.body.style.overflow;
    document.documentElement.style.overflow = 'hidden';
    document.body.style.overflow = 'hidden';

    let isIntroDone = false;

    // Liberar trazos base una vez que finalice la animación para que SVG dibuje el 100% continuo sin importar giros o cambios de tamaño
    function releaseBasePaths() {
      const bases = [pathBase1, pathBase2, pathBase3, pathBase4];
      bases.forEach(b => {
        if (b) {
          gsap.set(b, {
            clearProps: 'strokeDasharray,strokeDashoffset',
            opacity: 1
          });
        }
      });
    }

    // Asegurar que las estelas de chispas queden totalmente invisibles tras el intro
    function cleanupSparks() {
      const allSparks = [
        pathSparkCore1, pathSparkGlow1,
        pathSparkCore2, pathSparkGlow2,
        pathSparkCore3, pathSparkGlow3,
        pathSparkCore4, pathSparkGlow4
      ];
      allSparks.forEach(s => {
        if (s) gsap.set(s, { opacity: 0 });
      });
    }

    // Animación cinemática continua y ultrarrápida: aceleración por hardware GPU a 60/120 FPS
    const tl = gsap.timeline({ 
      defaults: { ease: 'power3.out' },
      onComplete: () => {
        isIntroDone = true;
        releaseBasePaths();
        cleanupSparks();
      }
    });

    // 1. Fondo WebGL de brasa: emerge con suavidad y fluidez sin filtros CSS que bloqueen el pipeline
    tl.fromTo('.welcome-bg', 
      { opacity: 0, y: 35, scale: 1.03 }, 
      { opacity: 1, y: 0, scale: 1, duration: 1.2, ease: 'power3.out' },
      0
    );

    // 2. Convergencia uniforme, sincronizada y fluida de las 4 líneas con estelas que trazan la estrella
    const sparkLength = 48;
    const animDelay = 0.18;
    const animDuration = 0.88;
    const paths = [
      { base: pathBase1, sparks: [pathSparkCore1, pathSparkGlow1] },
      { base: pathBase2, sparks: [pathSparkCore2, pathSparkGlow2] },
      { base: pathBase3, sparks: [pathSparkCore3, pathSparkGlow3] },
      { base: pathBase4, sparks: [pathSparkCore4, pathSparkGlow4] }
    ];

    paths.forEach(({ base, sparks }) => {
      if (!base) return;
      const len = base.getTotalLength();
      const safeLen = Math.ceil(len) + 40;

      // Trazo arquitectónico base (oculto en el gap con offset seguro para evitar residuos subpíxel)
      gsap.set(base, {
        strokeDasharray: `${safeLen} ${safeLen * 2}`,
        strokeDashoffset: safeLen + 5,
        opacity: 0
      });

      // Estelas incandescentes duales (núcleo blanco + halo térmico ámbar)
      sparks.forEach(spark => {
        if (!spark) return;
        gsap.set(spark, {
          strokeDasharray: `${sparkLength} ${len * 2}`,
          strokeDashoffset: sparkLength,
          opacity: 0
        });
      });

      // Animación continua y simultánea del trazo base hacia cada punta y a través del arco
      tl.set(base, { opacity: 1 }, animDelay);
      tl.to(base, {
        strokeDashoffset: 0,
        duration: animDuration,
        ease: 'power3.inOut'
      }, animDelay);

      // Desplazamiento sincronizado de las estelas incandescentes que recorren la línea y dibujan los pétalos de la estrella
      sparks.forEach(spark => {
        if (!spark) return;
        tl.to(spark, {
          strokeDashoffset: -(len - sparkLength),
          duration: animDuration,
          ease: 'power3.inOut'
        }, animDelay);

        // Encendido nítido y disolución suave al cerrar los cuadrantes
        tl.fromTo(spark, 
          { opacity: 0 }, 
          { opacity: 1, duration: 0.15, ease: 'power2.out' }, 
          animDelay
        );
        tl.to(spark, {
          opacity: 0,
          duration: 0.22,
          ease: 'power2.out'
        }, animDelay + animDuration - 0.15);
      });
    });

    // Liberar trazos base justo al completarse la convergencia
    tl.call(() => {
      releaseBasePaths();
      cleanupSparks();
    }, null, animDelay + animDuration + 0.05);

    // 3. Sello hermético, nítido y luminoso de la estrella simétrica al completarse el dibujo de los 4 arcos
    tl.fromTo('.welcome-astroid-stroke', 
      { opacity: 0, scale: 0.98 }, 
      { opacity: 1, scale: 1.0, duration: 0.35, ease: 'power2.out' }, 
      animDelay + animDuration - 0.10
    );

    tl.fromTo('.welcome-astroid-fill', 
      { opacity: 0 }, 
      { opacity: 1, duration: 0.35, ease: 'power2.out' }, 
      animDelay + animDuration - 0.08
    );

    tl.fromTo('.welcome-reticle-group', 
      { scale: 1.025 }, 
      { scale: 1.0, duration: 0.40, ease: 'power2.out' }, 
      animDelay + animDuration - 0.05
    );

    // 4. Isotipo de Candado: florece en el centro exacto con autoridad y pop cinemático
    tl.fromTo('.welcome-isotipo', 
      { opacity: 0, scale: 0.15 }, 
      { opacity: 0.85, scale: 1, duration: 0.52, ease: 'back.out(1.8)' }, 
      animDelay + animDuration - 0.02
    );

    // 5. Textos y controles editoriales
    tl.fromTo('.welcome-sidebar', 
      { opacity: 0, y: 15 }, 
      { opacity: 0.45, y: 0, duration: 0.6, ease: 'power2.out' }, 
      1.16
    )
    .fromTo('.welcome-kicker', 
      { opacity: 0, y: 18 }, 
      { opacity: 0.80, y: 0, duration: 0.6, ease: 'power3.out' }, 
      1.22
    )
    .fromTo('.welcome-title', 
      { opacity: 0, y: 22 }, 
      { opacity: 1, y: 0, duration: 0.7, ease: 'power3.out' }, 
      1.30
    )
    .fromTo('.welcome-controls', 
      { opacity: 0, y: 16 }, 
      { opacity: 1, y: 0, duration: 0.6, ease: 'power2.out' }, 
      1.40
    );

    // Inicialización del shader WebGL de brasa y fuego líquido continuo
    if (canvasEl) {
      cleanupGL = initEmberShader(canvasEl);
    }

    // Reajuste inmediato en rotación de pantalla o redimensionado
    const handleOrientationOrResize = () => {
      windowWidth = window.innerWidth;
      windowHeight = window.innerHeight;
      if (isIntroDone) {
        releaseBasePaths();
        cleanupSparks();
      }
    };

    window.addEventListener('resize', handleOrientationOrResize);
    window.addEventListener('orientationchange', () => {
      handleOrientationOrResize();
      setTimeout(handleOrientationOrResize, 80);
      setTimeout(handleOrientationOrResize, 250);
    });

    // Bloquear zoom, pellizco y deslizamiento EXCLUSIVAMENTE en el contenedor de bienvenida
    if (introContainer) {
      const preventTouchMove = (e) => {
        if (!isInfoModalOpen) {
          e.preventDefault();
        }
      };

      const preventGesture = (e) => {
        e.preventDefault();
      };

      const preventWheel = (e) => {
        if (!isInfoModalOpen) {
          e.preventDefault();
        }
      };

      introContainer.addEventListener('touchmove', preventTouchMove, { passive: false });
      introContainer.addEventListener('gesturestart', preventGesture, { passive: false });
      introContainer.addEventListener('gesturechange', preventGesture, { passive: false });
      introContainer.addEventListener('wheel', preventWheel, { passive: false });

      cleanupRestrictions = () => {
        if (introContainer) {
          introContainer.removeEventListener('touchmove', preventTouchMove);
          introContainer.removeEventListener('gesturestart', preventGesture);
          introContainer.removeEventListener('gesturechange', preventGesture);
          introContainer.removeEventListener('wheel', preventWheel);
        }
      };
    }

    return () => {
      document.documentElement.style.overflow = origHtmlOverflow;
      document.body.style.overflow = origBodyOverflow;
      window.removeEventListener('resize', handleOrientationOrResize);
      if (cleanupGL) cleanupGL();
      if (cleanupRestrictions) cleanupRestrictions();
    };
  });

  function initEmberShader(canvas) {
    const gl = canvas.getContext('webgl', { alpha: false, antialias: false, powerPreference: 'high-performance' });
    if (!gl) return null;

    const vsSource = `
      attribute vec2 position;
      void main() {
        gl_Position = vec4(position, 0.0, 1.0);
      }
    `;

    const fsSource = `
      precision mediump float;
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

        if (t < 0.12) return mix(c0, c1, t / 0.12);
        if (t < 0.28) return mix(c1, c2, (t - 0.12) / 0.16);
        if (t < 0.46) return mix(c2, c3, (t - 0.28) / 0.18);
        if (t < 0.64) return mix(c3, c4, (t - 0.46) / 0.18);
        if (t < 0.78) return mix(c4, c5, (t - 0.64) / 0.14);
        if (t < 0.90) return mix(c5, c6, (t - 0.78) / 0.12);
        return mix(c6, c7, (t - 0.90) / 0.10);
      }

      void main() {
        vec2 uv = gl_FragCoord.xy / u_resolution.xy;

        // Early-out: la zona superior es negro puro #050507 (ahorra fillrate crítico)
        if (uv.y > 0.62) {
          gl_FragColor = vec4(0.020, 0.020, 0.027, 1.0);
          return;
        }

        // Ritmo lento, uniforme y sedoso (brasa y fuego lento)
        float speed = 0.12;
        float t = u_time * speed;

        // Flujo horizontal continuo y uniforme hacia la derecha (+x)
        float flow = uv.x * 2.4 - (t * 0.28);

        // Onda sinuosa orgánica de brasa que fluye a la derecha
        float w1 = flow * 2.8;
        float crest = 0.17 
                    + 0.12 * sin(w1) 
                    + 0.05 * sin(w1 * 2.0 - 0.8) 
                    + 0.02 * cos(w1 * 3.0 + 0.4);

        // Micro-ondulaciones térmicas fluidas
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

        // Delimitación espacial y desvanecimiento
        float topFade = smoothstep(0.58, 0.18, uv.y);
        float rightPresence = smoothstep(0.20, 0.80, uv.x);
        float bottomBase = exp(-uv.y * 3.2) * smoothstep(0.15, 0.85, uv.x) * 0.35;
        float cornerEmber = exp(-((uv.x * 2.5) * (uv.x * 2.5) + (uv.y * 3.0) * (uv.y * 3.0))) * 0.18;

        float totalHeat = clamp((flameHeat * topFade * rightPresence) + (bottomBase * topFade) + (cornerEmber * 0.30), 0.0, 1.0);

        // Dithering analógico de alta velocidad (elimina banding sin sobrecargar el compositor con SVG)
        float dither = (fract(sin(dot(gl_FragCoord.xy, vec2(12.9898, 78.233))) * 43758.5453) - 0.5) * (1.8 / 255.0);
        vec3 col = clamp(emberPalette(totalHeat) + dither, 0.0, 1.0);

        gl_FragColor = vec4(col, 1.0);
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
      const dpr = Math.min(window.devicePixelRatio || 1, 1.0);
      const w = canvas.parentElement ? canvas.parentElement.clientWidth : window.innerWidth;
      const h = canvas.parentElement ? canvas.parentElement.clientHeight : window.innerHeight;
      canvas.width = Math.round(w * dpr);
      canvas.height = Math.round(h * dpr);
      gl.viewport(0, 0, canvas.width, canvas.height);
    }
    resize();

    const handleCanvasResize = () => {
      resize();
      setTimeout(resize, 80);
      setTimeout(resize, 250);
    };

    window.addEventListener('resize', handleCanvasResize);
    window.addEventListener('orientationchange', handleCanvasResize);

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
      window.removeEventListener('resize', handleCanvasResize);
      window.removeEventListener('orientationchange', handleCanvasResize);
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

    if (cleanupRestrictions) {
      cleanupRestrictions();
      cleanupRestrictions = null;
    }

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
        document.documentElement.style.overflow = '';
        document.body.style.overflow = '';
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

<svelte:window bind:innerWidth={windowWidth} bind:innerHeight={windowHeight} />

{#if isVisible}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div 
    bind:this={introContainer}
    onclick={handleAdvance}
    class="intro-screen-container fixed inset-0 h-[100dvh] z-50 flex flex-col justify-between select-none cursor-pointer overflow-hidden p-6 sm:p-10 font-sans text-white bg-[#050507] touch-none overscroll-none"
    style="touch-action: none; overscroll-behavior: none; -webkit-touch-callout: none; -webkit-user-select: none; user-select: none;"
  >
    <!-- FONDO DE BRASA Y FUEGO LENTO AL CARBÓN (100% GPU WebGL - Cero cortes, Infinito) -->
    <div class="welcome-bg absolute inset-0 overflow-hidden pointer-events-none z-0 bg-[#050507]">
      <!-- Canvas de flujo de brasa y magma hacia la derecha -->
      <canvas bind:this={canvasEl} class="absolute inset-0 w-full h-full pointer-events-none z-0"></canvas>
    </div>

    <!-- 1. RETÍCULA ARQUITECTÓNICA VECTORIAL CONTINUA CON ARCOS TANGENCIALES C1 E ISOTIPO -->
    <div class="welcome-grid absolute inset-0 z-10 pointer-events-none">
      <svg 
        class="absolute inset-0 w-full h-full pointer-events-none overflow-visible"
        xmlns="http://www.w3.org/2000/svg"
      >
        <defs>
          <!-- Degradado lineal para la línea inferior: 100% nítido en el arco y la parte alta, desvaneciéndose suavemente en la cola inferior -->
          <linearGradient id="bottomLineGrad" x1="0" y1="100%" x2="0" y2="0%">
            <stop offset="0%" stop-color="#ffffff" stop-opacity="0" />
            <stop offset="25%" stop-color="#ffffff" stop-opacity="0.05" />
            <stop offset="45%" stop-color="#ffffff" stop-opacity="0.15" />
            <stop offset="100%" stop-color="#ffffff" stop-opacity="0.15" />
          </linearGradient>
        </defs>

        <!-- Grupo con origen exacto en el centro del asteroide para el pulso de cierre -->
        <g class="welcome-reticle-group" style="transform-origin: {cx}px {cy}px;">
          <!-- 1. Trazo continuo, simétrico y cerrado de la estrella asteroide (sin cortes ni asimetrías) -->
          <path 
            class="welcome-astroid-stroke"
            d={astroid_d}
            fill="none"
            stroke="rgba(255, 255, 255, 0.18)"
            stroke-width="1"
            vector-effect="non-scaling-stroke"
            opacity="0"
          />
          <!-- Fondo interior del asteroide al cerrarse -->
          <path 
            class="welcome-astroid-fill"
            d={astroid_d}
            fill="rgba(255,255,255,0.022)"
            stroke="none"
            opacity="0"
          />

          <!-- LÍNEA 1: Superior -> Converge directamente a la punta superior -->
          <path 
            bind:this={pathBase1}
            d={path1_d}
            fill="none"
            stroke="rgba(255, 255, 255, 0.15)"
            stroke-width="1"
            stroke-linecap="butt"
            vector-effect="non-scaling-stroke"
            opacity="0"
          />
          <path 
            bind:this={pathSparkGlow1}
            d={path1_d}
            fill="none"
            stroke="#FF7A00"
            stroke-width="4.5"
            stroke-linecap="round"
            vector-effect="non-scaling-stroke"
            opacity="0"
          />
          <path 
            bind:this={pathSparkCore1}
            d={path1_d}
            fill="none"
            stroke="#FFF8EE"
            stroke-width="1.8"
            stroke-linecap="round"
            vector-effect="non-scaling-stroke"
            opacity="0"
          />

          <!-- LÍNEA 2: Izquierda -> Converge directamente a la punta izquierda -->
          <path 
            bind:this={pathBase2}
            d={path2_d}
            fill="none"
            stroke="rgba(255, 255, 255, 0.15)"
            stroke-width="1"
            stroke-linecap="butt"
            vector-effect="non-scaling-stroke"
            opacity="0"
          />
          <path 
            bind:this={pathSparkGlow2}
            d={path2_d}
            fill="none"
            stroke="#FF7A00"
            stroke-width="4.5"
            stroke-linecap="round"
            vector-effect="non-scaling-stroke"
            opacity="0"
          />
          <path 
            bind:this={pathSparkCore2}
            d={path2_d}
            fill="none"
            stroke="#FFF8EE"
            stroke-width="1.8"
            stroke-linecap="round"
            vector-effect="non-scaling-stroke"
            opacity="0"
          />

          <!-- LÍNEA 3: Inferior -> Converge directamente a la punta inferior -->
          <path 
            bind:this={pathBase3}
            d={path3_d}
            fill="none"
            stroke="url(#bottomLineGrad)"
            stroke-width="1"
            stroke-linecap="butt"
            vector-effect="non-scaling-stroke"
            opacity="0"
          />
          <path 
            bind:this={pathSparkGlow3}
            d={path3_d}
            fill="none"
            stroke="#FF7A00"
            stroke-width="4.5"
            stroke-linecap="round"
            vector-effect="non-scaling-stroke"
            opacity="0"
          />
          <path 
            bind:this={pathSparkCore3}
            d={path3_d}
            fill="none"
            stroke="#FFF8EE"
            stroke-width="1.8"
            stroke-linecap="round"
            vector-effect="non-scaling-stroke"
            opacity="0"
          />

          <!-- LÍNEA 4: Derecha -> Converge directamente a la punta derecha -->
          <path 
            bind:this={pathBase4}
            d={path4_d}
            fill="none"
            stroke="rgba(255, 255, 255, 0.15)"
            stroke-width="1"
            stroke-linecap="butt"
            vector-effect="non-scaling-stroke"
            opacity="0"
          />
          <path 
            bind:this={pathSparkGlow4}
            d={path4_d}
            fill="none"
            stroke="#FF7A00"
            stroke-width="4.5"
            stroke-linecap="round"
            vector-effect="non-scaling-stroke"
            opacity="0"
          />
          <path 
            bind:this={pathSparkCore4}
            d={path4_d}
            fill="none"
            stroke="#FFF8EE"
            stroke-width="1.8"
            stroke-linecap="round"
            vector-effect="non-scaling-stroke"
            opacity="0"
          />
        </g>
      </svg>

      <!-- Contenedor concéntrico que garantiza el centrado geométrico perfecto del isotipo en el centro de la estrella tanto en escritorio como en móvil -->
      <div 
        class="absolute -translate-x-1/2 -translate-y-1/2 pointer-events-none flex items-center justify-center w-[48px] h-[48px] z-20"
        style="left: {cx}px; top: {cy}px;"
      >
        <div 
          class="welcome-isotipo w-full h-full bg-[#FF2400] opacity-0 select-none drop-shadow-[0_0_18px_rgba(255,36,0,0.60)] drop-shadow-[0_2px_10px_rgba(0,0,0,0.5)]"
          style="
            mask-image: url('/imagenes/marca/isotipo.webp');
            -webkit-mask-image: url('/imagenes/marca/isotipo.webp');
            mask-size: contain;
            -webkit-mask-size: contain;
            mask-repeat: no-repeat;
            -webkit-mask-repeat: no-repeat;
            mask-position: center;
            -webkit-mask-position: center;
          "
          role="img"
          aria-label="Mías Isotipo"
        ></div>
      </div>
    </div>

    <!-- 2. MARGEN IZQUIERDO: COORDENADAS VERTICALES ("2026" Y "MÍAS") FLANQUEANDO EL EJE HORIZONTAL -->
    <div 
      class="welcome-sidebar absolute left-4 sm:left-6 z-10 flex flex-col items-center pointer-events-none select-none text-[9px] font-mono tracking-[0.28em] text-white/45 uppercase"
      style="top: {cy}px; transform: translateY(-50%);"
    >
      <span class="[writing-mode:vertical-rl] rotate-180 mb-3.5">2026</span>
      <span class="[writing-mode:vertical-rl] rotate-180 mt-3.5">MÍAS</span>
    </div>

    <!-- 4. ESPACIO SUPERIOR DE RESPIRACIÓN EDITORIAL -->
    <div class="welcome-top-spacer flex-1 min-h-[10vh]"></div>

    <!-- 5. BLOQUE DE TITULAR: ELEVADO A LA MITAD DEL CUADRANTE INFERIOR -->
    <div class="welcome-title-wrapper relative z-10 w-full max-w-xl mb-14 sm:mb-20">
      <!-- Subtítulo Kicker Bienvenido elevado más arriba con tipografía font-display font-light -->
      <p class="welcome-kicker font-display font-light text-xl sm:text-2xl text-white/80 mb-10 sm:mb-14 tracking-tight">
        Bienvenido
      </p>

      <!-- Titular Editorial con tipografía intacta y cápsulas estilizadas idénticas a la referencia -->
      <h1 class="welcome-title font-display text-[2.35rem] sm:text-5xl lg:text-6xl tracking-tight text-white leading-[1.26] sm:leading-[1.3] select-none">
        <span class="font-semibold text-white">Hamburguesas</span>
        <!-- Cápsula 1: Alargada y estilizada -->
        <span class="title-capsule-1 inline-block w-14 sm:w-16 h-6 sm:h-7 rounded-full border border-white/35 bg-white/[0.03] backdrop-blur-[2px] align-middle mx-1.5"></span>
        <br />
        <!-- de lujo: Mantenido intacto con Pinyon Script en rojo escarlata vivo -->
        <span 
          class="title-luxury inline-block align-baseline mr-1.5 pr-0.5 text-[#FF2400] text-[3.2rem] sm:text-6xl lg:text-7xl font-normal drop-shadow-[0_0_20px_rgba(255,36,0,0.45)] drop-shadow-[0_2px_10px_rgba(0,0,0,0.5)]"
          style="font-family: 'Pinyon Script', cursive; font-style: normal;"
        >de lujo</span>
        <!-- Cápsula 2: Alargada como en la referencia -->
        <span class="title-capsule-2 inline-block w-12 sm:w-14 h-6 sm:h-7 rounded-full border border-white/35 bg-white/[0.03] backdrop-blur-[2px] align-middle mx-1.5"></span> 
        <span class="font-light text-white/80">a fuego</span>
        <br />
        <span class="font-semibold text-white">del carbón</span> 
        <!-- Cápsula 3: Alargada -->
        <span class="title-capsule-3 inline-block w-14 sm:w-16 h-6 sm:h-7 rounded-full border border-white/35 bg-white/[0.03] backdrop-blur-[2px] align-middle mx-1.5"></span>
      </h1>
    </div>

    <!-- 6. CONTROLES INFERIORES: PLANOS Y ALINEADOS EN LA BASE -->
    <div class="welcome-controls relative z-10 flex items-center justify-between pb-3 sm:pb-5">
      <!-- Botón Izquierdo: Vidrio esmerilado suave translúcido con isotipo candado circular -->
      <button 
        type="button"
        onclick={toggleInfoModal}
        class="welcome-info-btn w-14 h-14 rounded-full border-0 outline-none bg-white/15 backdrop-blur-xl flex items-center justify-center text-white/90 hover:bg-white/25 active:scale-95 transition-all cursor-pointer group shadow-lg shadow-black/15"
        title="Información, Horarios y Términos Mías"
        aria-label="Ver información y términos"
      >
        <img 
          src="/imagenes/marca/isotipo.webp" 
          alt="Mías Candado de Corazón" 
          class="w-6 h-6 object-contain brightness-0 invert opacity-90 group-hover:scale-110 transition-transform"
        />
      </button>

      <!-- Botón Derecho: Contenedor con 3 contornos concéntricos blancos organizados y botón de avance en formato píldora -->
      <div class="welcome-advance-wrapper relative flex items-center justify-center w-24 sm:w-28 h-14">
        <!-- 3 contornos concéntricos blancos que emanan de manera organizada -->
        <div class="ripple-ring ripple-ring-1"></div>
        <div class="ripple-ring ripple-ring-2"></div>
        <div class="ripple-ring ripple-ring-3"></div>

        <!-- Botón de avance blanco de alto contraste en formato píldora -->
        <button 
          type="button"
          onclick={handleAdvance}
          class="welcome-advance-btn relative z-10 w-full h-full rounded-full bg-white text-black border border-black/20 flex items-center justify-center hover:scale-105 active:scale-95 transition-all cursor-pointer group"
          aria-label="Ingresar al restaurante"
        >
          <ChevronRight class="w-6 h-6 stroke-[2.2] text-black group-hover:translate-x-1 transition-transform" />
        </button>
      </div>
    </div>
  </div>

  <!-- TARJETA VACÍA COLOR CREMA CÁLIDO CON BORDES IPHONE OCUPANDO TODA LA PANTALLA (CON MARGEN PERIFÉRICO) -->
  {#if isInfoModalOpen}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div 
      class="fixed inset-0 z-50 flex items-center justify-center p-2.5 sm:p-4 md:p-5 bg-black/75 backdrop-blur-md animate-in fade-in duration-200"
      onclick={(e) => { e.stopPropagation(); isInfoModalOpen = false; }}
    >
      <div 
        class="w-full h-full bg-[#FAF6ED] rounded-[38px] sm:rounded-[48px] shadow-[0_25px_60px_-15px_rgba(0,0,0,0.6)] relative overflow-hidden flex flex-col items-center justify-center transition-all border border-black/5"
        onclick={(e) => e.stopPropagation()}
      >
        <!-- Botón Cerrar Minimalista -->
        <button 
          type="button"
          onclick={() => isInfoModalOpen = false}
          class="absolute top-5 right-5 sm:top-7 sm:right-7 w-11 h-11 rounded-full bg-black/[0.06] hover:bg-black/[0.12] active:scale-95 flex items-center justify-center text-char/70 hover:text-char transition-all cursor-pointer shadow-sm"
          aria-label="Cerrar tarjeta"
        >
          <X class="w-5 h-5 stroke-[2.2]" />
        </button>
      </div>
    </div>
  {/if}
{/if}

<style>
  /* Tres contornos concéntricos blancos que emanan organizadamente del botón de llamada a la acción */
  .ripple-ring {
    position: absolute;
    inset: 0;
    border-radius: 9999px;
    pointer-events: none;
    border: 1px solid rgba(255, 255, 255, 0.70);
    box-shadow: 0 0 10px rgba(255, 255, 255, 0.15);
    will-change: transform, opacity;
    animation: ripple-emanate 3s cubic-bezier(0.15, 0.65, 0.35, 1) infinite;
  }

  .ripple-ring-1 {
    animation-delay: 0s;
  }

  .ripple-ring-2 {
    animation-delay: -1s;
  }

  .ripple-ring-3 {
    animation-delay: -2s;
  }

  @keyframes ripple-emanate {
    0% {
      transform: scale(1);
      opacity: 0.85;
      border-color: rgba(255, 255, 255, 0.80);
      box-shadow: 0 0 8px rgba(255, 255, 255, 0.20);
    }
    40% {
      opacity: 0.40;
      border-color: rgba(255, 255, 255, 0.45);
      box-shadow: 0 0 12px rgba(255, 255, 255, 0.10);
    }
    100% {
      transform: scale(1.50);
      opacity: 0;
      border-color: rgba(255, 255, 255, 0);
      box-shadow: none;
    }
  }

  /* REGLAS RESPONSIVAS PARA DISPOSITIVOS MÓVILES EN APISADO (LANDSCAPE) Y PANTALLAS BAJAS */
  @media (max-height: 560px) {
    .intro-screen-container {
      padding: 8px 16px !important;
      padding-top: max(8px, env(safe-area-inset-top)) !important;
      padding-bottom: max(8px, env(safe-area-inset-bottom)) !important;
      padding-left: max(16px, env(safe-area-inset-left)) !important;
      padding-right: max(16px, env(safe-area-inset-right)) !important;
    }

    .welcome-top-spacer {
      min-height: 0px !important;
      flex: 0 0 0px !important;
    }

    .welcome-title-wrapper {
      margin-bottom: 6px !important;
      max-width: 92% !important;
    }

    .welcome-kicker {
      font-size: 0.82rem !important;
      margin-bottom: 3px !important;
    }

    .welcome-title {
      font-size: 1.35rem !important;
      line-height: 1.15 !important;
    }

    .welcome-title :global(.title-capsule-1),
    .welcome-title :global(.title-capsule-3) {
      width: 32px !important;
      height: 14px !important;
    }

    .welcome-title :global(.title-capsule-2) {
      width: 24px !important;
      height: 14px !important;
    }

    .welcome-title :global(.title-luxury) {
      font-size: 1.95rem !important;
    }

    .welcome-sidebar {
      font-size: 7px !important;
      letter-spacing: 0.18em !important;
      opacity: 0.25 !important;
    }

    .welcome-controls {
      padding-bottom: 2px !important;
    }

    .welcome-info-btn {
      width: 40px !important;
      height: 40px !important;
    }

    .welcome-info-btn img {
      width: 17px !important;
      height: 17px !important;
    }

    .welcome-advance-wrapper {
      width: 64px !important;
      height: 40px !important;
    }

    .welcome-advance-btn :global(svg) {
      width: 18px !important;
      height: 18px !important;
    }
  }

  @media (max-height: 420px) {
    .welcome-sidebar {
      display: none !important;
    }
    .welcome-title {
      font-size: 1.22rem !important;
      line-height: 1.12 !important;
    }
    .welcome-title :global(.title-luxury) {
      font-size: 1.75rem !important;
    }
  }
</style>

