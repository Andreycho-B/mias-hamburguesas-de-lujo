<script>
  import { onMount, tick } from 'svelte';
  import gsap from 'gsap';
  import Lenis from 'lenis';
  import { ChevronRight, X, ShieldCheck, ArrowLeft, Clock } from 'lucide-svelte';
  import { preloadCriticalImages } from '../lib/imagePreloader.js';

  let { onComplete } = $props();
  let introContainer = $state(null);
  let canvasEl = $state(null);
  let isVisible = $state(true);
  let isInfoModalOpen = $state(false);
  let isModalAnimating = false;

  let infoBtnEl = $state(null);
  let cardEl = $state(null);
  let modalBackdropEl = $state(null);
  let cardScrollEl = $state(null);
  let cardScrollContentEl = $state(null);
  let cardLenis = null;
  let cardLenisRafId = null;

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

    const metaTheme = document.querySelector('meta[name="theme-color"]');
    if (metaTheme) metaTheme.setAttribute('content', '#050507');

    // Precarga y decodificación inmediata en paralelo de todas las imágenes críticas
    preloadCriticalImages();

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
      if (cardLenis) {
        if (windowWidth >= 1024) {
          destroyCardLenis();
        } else {
          cardLenis.resize();
        }
      } else if (isInfoModalOpen && windowWidth < 1024) {
        initCardLenis();
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
      const metaTheme = document.querySelector('meta[name="theme-color"]');
      if (metaTheme) metaTheme.setAttribute('content', '#FAF6EE');
      window.removeEventListener('resize', handleOrientationOrResize);
      if (cleanupGL) cleanupGL();
      if (cleanupRestrictions) cleanupRestrictions();
      destroyCardLenis();
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

        // Ritmo continuo, vivo y fluido (brasa y fuego vivo)
        float speed = 0.20;
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
    const uRes = gl.getUniformLocation(prog, 'u_resolution');
    const uTime = gl.getUniformLocation(prog, 'u_time');

    // -------------------------------------------------------------------------
    // SISTEMA DE PARTÍCULAS DE CENIZA INCANDESCENTE Y CHISPAS DE BRASA (GPU PURO)
    // -------------------------------------------------------------------------
    const particleVsSource = `
      attribute vec4 a_seed; // x, y, size, speedFactor
      attribute vec2 a_life; // lifeOffset, isUpperRegion
      uniform float u_time;
      uniform vec2 u_resolution;
      varying float v_alpha;
      varying vec3 v_color;
      varying float v_flicker;

      void main() {
        // Velocidad sincronizada exactamente con el flujo horizontal del shader de brasa (0.02333 pantallas/segundo)
        float baseSpeed = 0.02333;
        float speedX = baseSpeed * a_seed.w;
        
        // Movimiento horizontal continuo hacia la derecha (+x)
        float posX = fract(a_seed.x + u_time * speedX);
        
        float posY = 0.0;
        float fade = 1.0;

        if (a_life.y > 0.5) {
          // --- Partículas en la parte superior oscura ---
          // Deriva sutil y flotación lenta horizontal en la oscuridad
          float driftY = sin(posX * 3.5 + a_life.x * 6.28 + u_time * 0.4) * 0.022;
          posY = mod(a_seed.y + u_time * 0.007 * a_seed.w, 0.42) + 0.52 + driftY;
          
          // Suave ciclo de respiración en la penumbra
          float cycle = fract(a_life.x + u_time * 0.08);
          fade = sin(cycle * 3.14159) * 0.68;
          
          // Tonalidad: ascuas sutiles que se enfrían (ámbar dorado y chispa cálida tenue)
          if (a_seed.z > 2.5) {
            v_color = vec3(0.96, 0.48, 0.12);
          } else {
            v_color = vec3(1.00, 0.88, 0.65);
          }
        } else {
          // --- Partículas en la zona de brasa y llama ---
          float riseRate = 0.024 * a_seed.w;
          float normY = mod(a_seed.y + u_time * riseRate, 0.58);
          float waveX = (posX * 2.4 - u_time * 0.056);
          float waveY = sin(waveX * 2.8 + a_life.x * 6.28) * 0.032 + sin(u_time * 2.8) * 0.010;
          posY = normY + waveY;
          
          float cycle = fract(a_life.x + u_time * 0.11);
          float burnLife = sin(cycle * 3.14159);
          float altitudeFade = smoothstep(0.55, 0.20, posY);
          fade = burnLife * altitudeFade * 0.92;
          
          if (a_seed.z > 4.4) {
            v_color = vec3(0.98, 0.32, 0.04);
          } else if (a_seed.z > 2.8) {
            v_color = vec3(1.00, 0.65, 0.16);
          } else {
            v_color = vec3(1.00, 0.92, 0.72);
          }
        }

        // Centelleo térmico orgánico
        v_flicker = sin(u_time * 12.0 + a_seed.x * 20.0) * 0.5 + 0.5;
        v_alpha = fade;
        
        gl_Position = vec4(posX * 2.0 - 1.0, posY * 2.0 - 1.0, 0.0, 1.0);
        gl_PointSize = a_seed.z * (0.85 + v_flicker * 0.30);
      }
    `;

    const particleFsSource = `
      precision mediump float;
      varying float v_alpha;
      varying vec3 v_color;
      varying float v_flicker;

      void main() {
        vec2 coord = gl_PointCoord - vec2(0.5);
        float dist = length(coord);
        if (dist > 0.5) discard;

        // Halo suave difuso de brasa ardiente
        float intensity = exp(-dist * dist * 12.0);
        
        // Núcleo térmico brillante en el centro
        float core = smoothstep(0.22, 0.0, dist);
        vec3 col = mix(v_color, vec3(1.0, 0.98, 0.92), core * 0.75);
        
        gl_FragColor = vec4(col * (0.90 + v_flicker * 0.20), v_alpha * intensity);
      }
    `;

    const particleVs = compile(gl.VERTEX_SHADER, particleVsSource);
    const particleFs = compile(gl.FRAGMENT_SHADER, particleFsSource);
    const particleProg = gl.createProgram();
    gl.attachShader(particleProg, particleVs);
    gl.attachShader(particleProg, particleFs);
    gl.linkProgram(particleProg);

    const PARTICLE_COUNT = 78;
    const particleData = new Float32Array(PARTICLE_COUNT * 6);
    for (let i = 0; i < PARTICLE_COUNT; i++) {
      const idx = i * 6;
      particleData[idx] = Math.random(); // X inicial

      if (i >= 60) {
        // ~18 partículas delicadas en la parte superior oscura (0.52 a 0.92)
        particleData[idx + 1] = 0.52 + Math.random() * 0.38;
        particleData[idx + 2] = 1.6 + Math.random() * 1.6;  // tamaño sutil (1.6 a 3.2px)
        particleData[idx + 3] = 0.75 + Math.random() * 0.35; // velocidad horizontal
        particleData[idx + 4] = Math.random();               // life offset
        particleData[idx + 5] = 1.0;                         // 1.0 = zona superior oscura
      } else {
        // Partículas en la zona de brasa y fuego (0.02 a 0.42)
        particleData[idx + 1] = 0.02 + Math.random() * 0.40;
        if (i < 28) {
          particleData[idx + 2] = 1.8 + Math.random() * 1.0;
        } else if (i < 48) {
          particleData[idx + 2] = 2.9 + Math.random() * 1.0;
        } else {
          particleData[idx + 2] = 4.2 + Math.random() * 1.6;
        }
        particleData[idx + 3] = 0.82 + Math.random() * 0.40;
        particleData[idx + 4] = Math.random();
        particleData[idx + 5] = 0.0;                         // 0.0 = zona de llama/brasa
      }
    }

    const particleBuf = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, particleBuf);
    gl.bufferData(gl.ARRAY_BUFFER, particleData, gl.STATIC_DRAW);

    const aParticleSeed = gl.getAttribLocation(particleProg, 'a_seed');
    const aParticleLife = gl.getAttribLocation(particleProg, 'a_life');
    const uParticleRes = gl.getUniformLocation(particleProg, 'u_resolution');
    const uParticleTime = gl.getUniformLocation(particleProg, 'u_time');

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
    let totalPauseTime = 0;
    let pauseStart = 0;
    let wasModalOpen = false;

    function loop(now) {
      if (!gl || !canvas) return;

      if (isInfoModalOpen) {
        if (!wasModalOpen) {
          wasModalOpen = true;
          pauseStart = now;
        }
      } else {
        if (wasModalOpen) {
          wasModalOpen = false;
          totalPauseTime += (now - pauseStart);
        }
        const timeVal = (now - startTime - totalPauseTime) / 1000;

        // 1. Fondo de brasa fluida
        gl.disable(gl.BLEND);
        gl.useProgram(prog);
        gl.bindBuffer(gl.ARRAY_BUFFER, buf);
        gl.vertexAttribPointer(aPos, 2, gl.FLOAT, false, 0, 0);
        gl.enableVertexAttribArray(aPos);
        gl.uniform2f(uRes, canvas.width, canvas.height);
        gl.uniform1f(uTime, timeVal);
        gl.drawArrays(gl.TRIANGLES, 0, 6);

        // 2. Cenizas ardientes y chispas flotando al mismo ritmo hacia la derecha
        gl.enable(gl.BLEND);
        gl.blendFunc(gl.SRC_ALPHA, gl.ONE);
        gl.useProgram(particleProg);
        gl.bindBuffer(gl.ARRAY_BUFFER, particleBuf);
        gl.vertexAttribPointer(aParticleSeed, 4, gl.FLOAT, false, 24, 0);
        gl.enableVertexAttribArray(aParticleSeed);
        gl.vertexAttribPointer(aParticleLife, 2, gl.FLOAT, false, 24, 16);
        gl.enableVertexAttribArray(aParticleLife);
        gl.uniform2f(uParticleRes, canvas.width, canvas.height);
        gl.uniform1f(uParticleTime, timeVal);
        gl.drawArrays(gl.POINTS, 0, PARTICLE_COUNT);
      }
      animFrameId = requestAnimationFrame(loop);
    }
    animFrameId = requestAnimationFrame(loop);

    return () => {
      if (animFrameId) cancelAnimationFrame(animFrameId);
      window.removeEventListener('resize', handleCanvasResize);
      window.removeEventListener('orientationchange', handleCanvasResize);
      if (gl) {
        gl.deleteBuffer(buf);
        gl.deleteBuffer(particleBuf);
        gl.deleteProgram(prog);
        gl.deleteProgram(particleProg);
        gl.deleteShader(vs);
        gl.deleteShader(fs);
        gl.deleteShader(particleVs);
        gl.deleteShader(particleFs);
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

    // Preparar el fondo para el desvanecimiento hacia el Home (beige suave #FAF6EE)
    document.documentElement.classList.remove('theme-intro');
    document.documentElement.classList.add('theme-home');
    document.documentElement.style.backgroundColor = '#FAF6EE';
    document.documentElement.style.colorScheme = 'light';
    const metaTheme = document.getElementById('theme-color-meta') || document.querySelector('meta[name="theme-color"]');
    if (metaTheme) metaTheme.setAttribute('content', '#FAF6EE');
    const metaApple = document.getElementById('apple-status-bar-meta') || document.querySelector('meta[name="apple-mobile-web-app-status-bar-style"]');
    if (metaApple) metaApple.setAttribute('content', 'default');

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

  async function openInfoModal(e) {
    if (e) e.stopPropagation();
    if (isModalAnimating || isInfoModalOpen) return;
    isModalAnimating = true;

    // Feedback táctil del botón (GPU puro)
    if (infoBtnEl) {
      gsap.fromTo(infoBtnEl, 
        { scale: 0.90 }, 
        { scale: 1, duration: 0.25, ease: 'power2.out', force3D: true }
      );
    }

    isInfoModalOpen = true;
    await tick();

    if (!cardEl || !modalBackdropEl) {
      isModalAnimating = false;
      return;
    }

    // Inicializar Lenis Smooth Scroll estilo Awwwards
    initCardLenis();

    let originX = '50%';
    let originY = '50%';

    if (infoBtnEl) {
      const btnRect = infoBtnEl.getBoundingClientRect();
      const cardRect = cardEl.getBoundingClientRect();
      const btnCenterX = btnRect.left + btnRect.width / 2;
      const btnCenterY = btnRect.top + btnRect.height / 2;
      originX = `${Math.round(btnCenterX - cardRect.left)}px`;
      originY = `${Math.round(btnCenterY - cardRect.top)}px`;
    }

    gsap.killTweensOf([modalBackdropEl, cardEl]);

    gsap.set(cardEl, {
      transformOrigin: `${originX} ${originY}`,
      scale: 0.18,
      opacity: 0,
      force3D: true
    });

    gsap.set(modalBackdropEl, {
      opacity: 0,
      force3D: true
    });

    const tl = gsap.timeline({
      onComplete: () => {
        isModalAnimating = false;
        if (cardEl) {
          gsap.set(cardEl, { clearProps: 'transform' });
        }
        if (cardLenis) {
          cardLenis.resize();
        }
      }
    });

    // 1. Fondo oscuro cinematográfico (Opacidad pura por GPU a 60/120 FPS sin tocar el canvas)
    tl.to(modalBackdropEl, {
      opacity: 1,
      duration: 0.28,
      ease: 'power2.out',
      force3D: true
    }, 0);

    // 2. Expansión fluida de la tarjeta estilo Apple Spring (100% GPU compositor, cero repaints)
    tl.to(cardEl, {
      scale: 1,
      opacity: 1,
      duration: 0.30,
      ease: 'power2.out',
      force3D: true
    }, 0);
  }

  function closeInfoModal(e) {
    if (e) e.stopPropagation();
    if (isModalAnimating || !isInfoModalOpen) return;
    if (!cardEl || !modalBackdropEl) {
      isInfoModalOpen = false;
      return;
    }

    isModalAnimating = true;
    destroyCardLenis();

    let originX = '50%';
    let originY = '50%';

    if (infoBtnEl) {
      const btnRect = infoBtnEl.getBoundingClientRect();
      const cardRect = cardEl.getBoundingClientRect();
      const btnCenterX = btnRect.left + btnRect.width / 2;
      const btnCenterY = btnRect.top + btnRect.height / 2;
      originX = `${Math.round(btnCenterX - cardRect.left)}px`;
      originY = `${Math.round(btnCenterY - cardRect.top)}px`;
    }

    gsap.killTweensOf([modalBackdropEl, cardEl]);

    gsap.set(cardEl, {
      transformOrigin: `${originX} ${originY}`,
      force3D: true
    });

    const tl = gsap.timeline({
      onComplete: () => {
        isInfoModalOpen = false;
        isModalAnimating = false;
      }
    });

    // 1. Colapso en espejo de la tarjeta hacia el botón a 60/120 FPS
    tl.to(cardEl, {
      scale: 0.18,
      opacity: 0,
      duration: 0.22,
      ease: 'power2.in',
      force3D: true
    }, 0);

    // 2. Desvanecimiento del fondo oscuro
    tl.to(modalBackdropEl, {
      opacity: 0,
      duration: 0.20,
      ease: 'power2.in',
      force3D: true
    }, 0.02);
  }

  // Acción para fijar el límite de scroll en iOS Safari y eliminar el rebote elástico inferior
  // Acción para fijar el límite de scroll en iOS Safari y eliminar el rebote elástico inferior
  // Evita que la palabra "mias" suba y se vea recortada flotando cuando el usuario desliza más al final.
  let touchVelocityY = 0;
  let touchLastY = 0;
  let touchLastTime = 0;
  let touchStartY = 0;
  let touchCleanupFns = [];

  // Inicialización de Lenis Smooth Scroll optimizado a 120 FPS (Apple ProMotion):
  // Aterrizaje suave con desaceleración progresiva continua (sin rebotes, sin jalón y sin estiramientos)
  function initCardLenis() {
    destroyCardLenis();
    if (!cardScrollEl || !cardScrollContentEl) return;
    if (windowWidth >= 1024) return;

    cardLenis = new Lenis({
      wrapper: cardScrollEl,
      content: cardScrollContentEl,
      eventsTarget: cardScrollEl,
      duration: 1.4,
      easing: (t) => 1 - Math.pow(1 - t, 3.8),
      smoothWheel: true,
      syncTouch: true,
      syncTouchLerp: 0.10,
      touchInertiaExponent: 1.5,
      touchMultiplier: 1.25,
      autoResize: true,
      overscroll: false,
      virtualScroll: (data) => {
        const { event } = data;
        if (event && event.type === 'touchend') {
          const timeSinceLast = performance.now() - touchLastTime;
          // Si el usuario soltó con impulso de deslizamiento:
          if (timeSinceLast < 160 && Math.abs(touchVelocityY) > 0.08) {
            // Distancia de inercia extendida estilo Awwwards
            const inertiaDistance = Math.sign(touchVelocityY) * Math.min(Math.pow(Math.abs(touchVelocityY), 1.30) * 360, 680);
            const current = cardLenis.animatedScroll;
            const target = Math.max(0, Math.min(cardLenis.limit, current + inertiaDistance));
            // Duración extendida para aterrizaje pausado y sin prisa
            const duration = Math.max(1.15, Math.min(2.1, 0.75 + Math.abs(touchVelocityY) * 0.55));

            // Curva Quartic Ease-Out extendida con desaceleración progresiva y suave
            cardLenis.scrollTo(target, {
              duration,
              easing: (t) => 1 - Math.pow(1 - t, 3.8),
              programmatic: true
            });
            return false;
          }
        }
        return true;
      }
    });

    let maxScroll = 0;

    const handleTouchStart = (e) => {
      if (e.touches.length === 1) {
        touchStartY = e.touches[0].clientY;
        touchLastY = touchStartY;
        touchLastTime = performance.now();
        touchVelocityY = 0;
        maxScroll = cardScrollEl.scrollHeight - cardScrollEl.clientHeight;
      }
    };

    const handleTouchMove = (e) => {
      if (e.touches.length !== 1) return;
      const currentY = e.touches[0].clientY;
      const now = performance.now();
      const dt = now - touchLastTime;

      // Muestreo de alta frecuencia calibrado para pantallas ProMotion 120Hz (~8ms)
      if (dt >= 3) {
        touchVelocityY = (touchLastY - currentY) / dt; // px/ms
        touchLastY = currentY;
        touchLastTime = now;
      }

      const currentScroll = cardScrollEl.scrollTop;

      // Bloquear rebote elástico en los límites superior e inferior para anclar la tarjeta nítidamente
      if (currentScroll <= 0 && currentY > touchStartY) {
        if (e.cancelable) e.preventDefault();
        return;
      }

      if (currentScroll >= maxScroll - 1 && currentY < touchStartY) {
        if (e.cancelable) e.preventDefault();
        return;
      }

      touchStartY = currentY;
    };

    cardScrollEl.addEventListener('touchstart', handleTouchStart, { passive: true });
    cardScrollEl.addEventListener('touchmove', handleTouchMove, { passive: false });

    touchCleanupFns = [
      () => cardScrollEl?.removeEventListener('touchstart', handleTouchStart),
      () => cardScrollEl?.removeEventListener('touchmove', handleTouchMove)
    ];

    function lenisLoop(time) {
      if (cardLenis) {
        cardLenis.raf(time);
        cardLenisRafId = requestAnimationFrame(lenisLoop);
      }
    }
    cardLenisRafId = requestAnimationFrame(lenisLoop);
  }

  function destroyCardLenis() {
    if (cardLenis) {
      cardLenis.destroy();
      cardLenis = null;
    }
    if (cardLenisRafId) {
      cancelAnimationFrame(cardLenisRafId);
      cardLenisRafId = null;
    }
    touchCleanupFns.forEach(fn => fn());
    touchCleanupFns = [];
  }
</script>

<svelte:window 
  bind:innerWidth={windowWidth} 
  bind:innerHeight={windowHeight} 
  onkeydown={(e) => { if (e.key === 'Escape' && isInfoModalOpen) closeInfoModal(e); }}
/>

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
        Bienvenido a Mías
      </p>

      <!-- Titular Editorial con tipografía intacta y cápsulas proporcionales al texto que les precede -->
      <h1 class="welcome-title font-display text-[2.35rem] sm:text-5xl lg:text-6xl tracking-tight text-white leading-[1.26] sm:leading-[1.3] select-none">
        <span class="font-semibold text-white">Hamburguesas</span>
        <!-- Cápsula 1: Proporcional a 'Hamburguesas' (la más larga) -->
        <span class="title-capsule-1 inline-block w-[58px] sm:w-[88px] h-6 sm:h-7 rounded-full border border-white/35 bg-white/[0.03] backdrop-blur-[2px] align-middle mx-1.5"></span>
        <br />
        <!-- de lujo: Mantenido intacto con Pinyon Script en rojo escarlata vivo -->
        <span 
          class="title-luxury inline-block align-baseline mr-1.5 pr-0.5 text-[#FF2400] text-[3.2rem] sm:text-6xl lg:text-7xl font-normal drop-shadow-[0_0_20px_rgba(255,36,0,0.45)] drop-shadow-[0_2px_10px_rgba(0,0,0,0.5)]"
          style="font-family: 'Pinyon Script', cursive; font-style: normal;"
        >de lujo</span>
        <!-- Cápsula 2: Proporcional a 'de lujo' (la más compacta) -->
        <span class="title-capsule-2 inline-block w-[28px] sm:w-[40px] h-6 sm:h-7 rounded-full border border-white/35 bg-white/[0.03] backdrop-blur-[2px] align-middle mx-1.5"></span>
        <br />
        <span class="font-semibold text-white">al carbón</span> 
        <!-- Cápsula 3: Proporcional a 'al carbón' (longitud intermedia) -->
        <span class="title-capsule-3 inline-block w-[42px] sm:w-[62px] h-6 sm:h-7 rounded-full border border-white/35 bg-white/[0.03] backdrop-blur-[2px] align-middle mx-1.5"></span>
      </h1>
    </div>

    <!-- 6. CONTROLES INFERIORES: PLANOS Y ALINEADOS EN LA BASE -->
    <div class="welcome-controls relative z-10 flex items-center justify-between pb-3 sm:pb-5">
      <!-- Botón Izquierdo: Vidrio esmerilado suave translúcido con isotipo candado circular -->
      <button 
        bind:this={infoBtnEl}
        type="button"
        onclick={openInfoModal}
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

  <!-- TARJETA COLOR CREMA LUMINOSO CON BORDES IPHONE (MARGEN UNIFORME 12PX EN MÓVIL, HORIZONTAL COMPLETO EN ESCRITORIO) -->
  {#if isInfoModalOpen}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div 
      bind:this={modalBackdropEl}
      data-lenis-prevent
      class="welcome-modal-backdrop fixed inset-0 h-[100dvh] z-[100] flex items-center justify-center bg-[#050507]/94 lg:backdrop-blur-sm cursor-pointer will-change-[opacity]"
      style="touch-action: pan-y;"
      onclick={closeInfoModal}
    >
      <div 
        bind:this={cardEl}
        data-lenis-prevent
        class="welcome-card-modal w-full h-full lg:max-w-[94vw] xl:max-w-[1440px] 2xl:max-w-[1560px] lg:h-[88vh] xl:h-[90vh] bg-[#FAF6EE] shadow-2xl lg:shadow-[0_25px_60px_-15px_rgba(0,0,0,0.6)] relative border border-black/5 rounded-[38px] sm:rounded-[48px] overflow-hidden cursor-default flex flex-col"
        style="container-type: inline-size; touch-action: pan-y;"
        onclick={(e) => e.stopPropagation()}
      >
        <!-- Botón Cerrar Flotante Apple Standard (top-5 right-5 según estándar, 100% accesible y fijo) -->
        <button 
          type="button"
          onclick={closeInfoModal}
          class="absolute top-5 right-5 lg:top-7 lg:right-7 w-10 h-10 rounded-full bg-black/[0.06] hover:bg-black/[0.12] active:scale-95 flex items-center justify-center text-char/80 hover:text-char transition-all cursor-pointer shadow-sm z-30 shrink-0"
          aria-label="Cerrar tarjeta"
        >
          <X class="w-5 h-5 stroke-[2.2]" />
        </button>

        <!-- Contenedor Deslizable Continuo (Vertical en Móvil, Spread Horizontal de 2 Columnas en Escritorio) -->
        <article 
          bind:this={cardScrollEl}
          class="welcome-card-scroll flex-1 min-h-0 w-full overflow-y-auto lg:overflow-hidden text-char"
          data-lenis-prevent
          style="touch-action: pan-y; -webkit-overflow-scrolling: auto; scroll-behavior: auto; overscroll-behavior: none; overscroll-behavior-y: none; scrollbar-width: none;"
        >
          <div 
            bind:this={cardScrollContentEl} 
            class="w-full flex flex-col lg:flex-row lg:h-full"
            style="will-change: transform; transform: translate3d(0, 0, 0); -webkit-backface-visibility: hidden; backface-visibility: hidden;"
          >
            
            <!-- PANEL IZQUIERDO: HERO VISUAL, COMPOSICIÓN ASIMÉTRICA & MANIFIESTO DE AUTOR -->
            <div class="w-full lg:w-[53%] xl:w-[51%] flex flex-col justify-between p-6 sm:p-8 lg:p-8 xl:p-10 lg:pr-8 lg:border-r lg:border-black/[0.08] lg:h-full lg:overflow-y-auto welcome-column-scroll">
              <div>
                <!-- 1. CABECERA EDITORIAL MINIMALISTA -->
                <header class="flex items-center justify-between pr-16 select-none">
                  <span class="text-[10px] sm:text-[11px] font-sans font-semibold tracking-[0.26em] text-char/60 uppercase">
                    MÍAS
                  </span>
                </header>

                <!-- 2. COMPOSICIÓN EDITORIAL DE DOS IMÁGENES ASIMÉTRICAS -->
                <div class="flex items-start justify-between lg:justify-start gap-4 sm:gap-6 lg:gap-7 mt-8 sm:mt-10 lg:mt-5 select-none">
                  <!-- Imagen Principal Izquierda (Salón Interior, más alta, esquinas rectas) -->
                  <figure class="w-[56%] lg:w-[195px] xl:w-[220px] flex-shrink-0 flex flex-col m-0">
                    <div class="w-full aspect-[3/4] overflow-hidden border border-black/10 shadow-sm bg-black/5 rounded-none select-none">
                      <img 
                        src="/imagenes/local/restaurante_mias_desde_dentro.webp" 
                        alt="Salón Restaurante Mías Hamburguesas de Lujo en Duitama" 
                        class="w-full h-full object-cover select-none pointer-events-none"
                        loading="eager"
                        fetchpriority="high"
                        decoding="async"
                      />
                    </div>
                    <figcaption class="mt-2 font-sans text-[8px] sm:text-[9px] uppercase tracking-[0.14em] text-char/50 leading-tight select-none">
                      SALÓN PRINCIPAL / DUITAMA
                    </figcaption>
                  </figure>

                  <!-- Imagen Secundaria Derecha (Fachada Exterior más pequeña tal cual la referencia, esquinas rectas) -->
                  <figure class="w-[38%] lg:w-[140px] xl:w-[155px] flex-shrink-0 flex flex-col m-0">
                    <div class="w-full aspect-[4/5] overflow-hidden border border-black/10 shadow-sm bg-black/5 rounded-none select-none">
                      <img 
                        src="/imagenes/local/restaurante_mias_desde_fuera.webp" 
                        alt="Fachada Exterior Mías Duitama Horno de Brasa" 
                        class="w-full h-full object-cover select-none pointer-events-none"
                        loading="eager"
                        fetchpriority="high"
                        decoding="async"
                      />
                    </div>
                    <figcaption class="mt-2 font-sans text-[8px] sm:text-[9px] uppercase tracking-[0.14em] text-char/50 leading-tight select-none">
                      FACHADA / HORNO DE BRASA
                    </figcaption>
                  </figure>
                </div>
              </div>

              <!-- 3. BLOQUE TIPOGRÁFICO DE AUTOR (REFERENCIA ALEXANDER MCQUEEN) -->
              <section class="mt-6 sm:mt-7 lg:mt-5 lg:mb-1">
                <span class="text-[#FF2400] text-lg font-serif leading-none block mb-1 select-none">*</span>
                <h2 class="font-serif font-normal text-2xl sm:text-3xl lg:text-[2rem] tracking-tight text-char uppercase leading-[1.12] select-none">
                  ALTA COCINA —<br />
                  <span class="font-bold">HAMBURGUESAS</span><br />
                  DE AUTOR.
                </h2>
                <p class="mt-2.5 text-xs sm:text-[13px] text-char/75 leading-relaxed font-sans max-w-lg">
                  Lo íntimo hecho para compartir. Cada creación es elaborada como una pieza de alta cocina: 160 gramos de res seleccionada, pan pretzel fresco y sellado a la brasa viva en nuestro horno artesanal.
                </p>
              </section>
            </div>

            <!-- PANEL DERECHO: HORARIO, FIDELIDAD, PRIVACIDAD & MONOLITO MIAS -->
            <div class="w-full lg:w-[47%] xl:w-[49%] flex flex-col justify-start lg:justify-between px-6 pt-6 pb-0 sm:px-8 sm:pt-8 sm:pb-0 lg:p-8 xl:p-10 lg:pl-8 lg:h-full lg:overflow-y-auto welcome-column-scroll">
              <div class="space-y-6 lg:space-y-7 lg:pt-2 lg:pr-14">
                <!-- 4. HORARIO DE ATENCIÓN Y DOMICILIOS EN DUITAMA -->
                <section class="pt-6 lg:pt-0 border-t lg:border-t-0 border-black/[0.08]">
                  <p class="text-[9px] sm:text-[10px] font-mono tracking-[0.20em] uppercase text-char/45 font-medium mb-1">
                    HORARIO DE ATENCIÓN
                  </p>
                  <p class="font-serif text-base sm:text-lg font-bold text-char">
                    Todos los días • 3:00 PM a 9:30 PM
                  </p>
                  <p class="mt-1 text-xs text-char/70 leading-relaxed max-w-md font-sans">
                    Servicio continuo en mesa en salón gourmet y despachos a domicilio en la ciudad de Duitama.
                  </p>
                  <p class="mt-1.5 text-[11px] font-sans text-char/55">
                    Línea directa y reservas: <span class="font-medium text-char/80">321 252 0881</span>
                  </p>
                </section>

                <!-- 5. PASAPORTE DE SELLOS (NUMERACIÓN 10 Y 1 CON MÁXIMA CLARIDAD Y CONTRASTE) -->
                <section class="pt-6 border-t border-black/[0.08]">
                  <p class="text-[9px] sm:text-[10px] font-mono tracking-[0.20em] uppercase text-char/45 font-medium mb-1">
                    PASAPORTE DE SELLOS
                  </p>
                  <p class="font-sans text-base sm:text-lg font-bold text-char tracking-tight">
                    <span class="font-display font-extrabold text-char text-[1.12em] tracking-normal inline-block" style="font-feature-settings: 'lnum' 1, 'tnum' 1;">10</span> Sellos = <span class="font-display font-extrabold text-char text-[1.12em] tracking-normal inline-block" style="font-feature-settings: 'lnum' 1, 'tnum' 1;">1</span> Original Gratis
                  </p>
                  <p class="mt-1 text-xs text-char/70 leading-relaxed max-w-md font-sans">
                    Acreditación por cada factura física de consumo en el restaurante. Beneficio personal e intransferible.
                  </p>
                </section>

                <!-- 6. POLÍTICA DE PRIVACIDAD & TRATAMIENTO DE DATOS (LEY 1581) -->
                <section class="pt-6 border-t border-black/[0.08]">
                  <p class="text-[9px] sm:text-[10px] font-mono tracking-[0.20em] uppercase text-char/45 font-medium mb-1">
                    POLÍTICA DE PRIVACIDAD & TRATAMIENTO DE DATOS (LEY 1581)
                  </p>
                  <p class="mt-1 text-[11px] font-sans text-char/65 leading-relaxed max-w-lg select-text">
                    En estricto cumplimiento de la Ley Estatutaria 1581 de 2012 (Habeas Data) de la República de Colombia, Mías garantiza que los datos personales suministrados (nombre, línea telefónica, dirección de entrega y comprobantes de facturación) son tratados bajo rigurosa confidencialidad con el fin exclusivo de procesar pedidos gastronómicos y validar sellos de fidelización en Duitama. Mías no comparte ni comercializa información con terceros. El titular puede ejercer en todo momento sus derechos de consulta, actualización y supresión comunicándose a nuestros canales oficiales.
                  </p>
                </section>
              </div>

              <!-- 7. TITULAR GIGANTE MONOLÍTICO "mias" AL PIE -->
              <footer class="welcome-card-footer w-full mt-6 lg:mt-0 lg:pt-8 pointer-events-none select-none flex items-start lg:items-end justify-center lg:justify-start leading-none">
                <span 
                  class="welcome-card-mias font-flourish font-black text-[#FF2400] tracking-[-0.04em] lg:tracking-[-0.035em] uppercase block text-center lg:text-left w-max whitespace-nowrap select-none"
                  style="font-family: 'Playfair Display', serif; font-style: normal; font-synthesis: none; line-height: 0.82; will-change: transform; transform: translate3d(0, 0, 0); -webkit-backface-visibility: hidden; backface-visibility: hidden;"
                >
                  mias
                </span>
              </footer>
            </div>
          </div>
        </article>
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

  /* CONTENEDOR PRINCIPAL: PROMOCIÓN GPU A NIVEL DE COMPOSITOR PARA 60/120 FPS */
  .intro-screen-container {
    will-change: transform;
    transform: translate3d(0, 0, 0);
    -webkit-backface-visibility: hidden;
    backface-visibility: hidden;
  }

  /* CONTENEDOR Y BACKDROP DEL MODAL: RESPETA LA ISLA/NOTCH Y CONTROLES SUPERIORES DE IPHONE PWA */
  .welcome-modal-backdrop {
    padding-left: 12px;
    padding-right: 12px;
    padding-bottom: 12px;
    padding-top: max(14px, calc(env(safe-area-inset-top, 0px) + 6px));
  }

  @media (min-width: 640px) {
    .welcome-modal-backdrop {
      padding: 1.25rem; /* sm:p-5 */
    }
  }

  @media (min-width: 768px) {
    .welcome-modal-backdrop {
      padding: 1.5rem; /* md:p-6 */
    }
  }

  @media (min-width: 1024px) {
    .welcome-modal-backdrop {
      padding: 1.75rem; /* lg:p-7 */
    }
  }

  @media (min-width: 1280px) {
    .welcome-modal-backdrop {
      padding: 2.25rem; /* xl:p-9 */
    }
  }

  /* TARJETA EDITORIAL APPLE STANDARD: CURVATURA SQUIRCLE Y DESLIZAMIENTO TÁCTIL NATIVO */
  .welcome-card-modal {
    border-radius: 38px;
    -webkit-border-radius: 38px;
    overflow: hidden;
    isolation: isolate;
    will-change: transform, opacity;
    transform: translateZ(0);
    -webkit-backface-visibility: hidden;
    backface-visibility: hidden;
    contain: layout paint;
  }

  @media (min-width: 640px) {
    .welcome-card-modal {
      border-radius: 48px;
      -webkit-border-radius: 48px;
    }
  }

  /* DESPLAZAMIENTO FLUIDO NATIVO SIN BARRAS VISIBLES (MOMENTUM GPU A 120 FPS PROMOTION) */
  .welcome-card-scroll {
    touch-action: pan-y;
    -webkit-overflow-scrolling: auto !important;
    scroll-behavior: auto !important;
    overscroll-behavior: none !important;
    overscroll-behavior-y: none !important;
    scrollbar-width: none;
    -ms-overflow-style: none;
    transform: translate3d(0, 0, 0);
    -webkit-backface-visibility: hidden;
    backface-visibility: hidden;
    contain: layout paint;
  }
  .welcome-column-scroll {
    touch-action: pan-y;
    -webkit-overflow-scrolling: touch;
    scroll-behavior: auto;
    overscroll-behavior: none;
    scrollbar-width: none;
    -ms-overflow-style: none;
  }
  .welcome-card-scroll::-webkit-scrollbar,
  .welcome-column-scroll::-webkit-scrollbar {
    display: none;
  }

  /* TITULAR MONOLÍTICO "MIAS" AL PIE DE LA TARJETA EN MOBILE: RECORTADO EN LA BASE SEGÚN REFERENCIA */
  .welcome-card-footer {
    padding-bottom: 0;
    margin-bottom: 0;
    margin-left: -1.5rem;
    margin-right: -1.5rem;
    width: calc(100% + 3rem);
    font-size: clamp(6.6rem, 36.5vw, 11rem);
    height: 0.70em;
    overflow: hidden;
    will-change: transform;
    transform: translate3d(0, 0, 0);
  }

  .welcome-card-mias {
    font-size: 1em;
    letter-spacing: -0.04em;
    display: block;
    line-height: 0.82;
    transform: translate3d(0, 0, 0);
    -webkit-backface-visibility: hidden;
    backface-visibility: hidden;
  }

  @media (min-width: 640px) and (max-width: 1023px) {
    .welcome-card-footer {
      margin-bottom: 0;
      margin-left: -2rem;
      margin-right: -2rem;
      width: calc(100% + 4rem);
      font-size: clamp(6.6rem, 36.5vw, 11rem);
      height: 0.70em;
    }
  }

  @media (min-width: 1024px) {
    .welcome-card-footer {
      margin-bottom: 0;
      margin-left: 0;
      margin-right: 0;
      width: 100%;
      height: auto;
      padding-bottom: 0;
      overflow: hidden;
      font-size: inherit;
    }

    .welcome-card-mias {
      font-size: min(32cqw, 14vw, 15vh);
      letter-spacing: -0.035em;
      transform: none;
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

    .welcome-title :global(.title-capsule-1) {
      width: 38px !important;
      height: 14px !important;
    }

    .welcome-title :global(.title-capsule-3) {
      width: 28px !important;
      height: 14px !important;
    }

    .welcome-title :global(.title-capsule-2) {
      width: 18px !important;
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

  .welcome-column-scroll {
    touch-action: pan-y;
    -webkit-overflow-scrolling: touch;
    overscroll-behavior: contain;
    scrollbar-width: thin;
    scrollbar-color: rgba(0, 0, 0, 0.15) transparent;
  }

  .welcome-column-scroll::-webkit-scrollbar {
    width: 4px;
  }

  .welcome-column-scroll::-webkit-scrollbar-thumb {
    background-color: rgba(0, 0, 0, 0.15);
    border-radius: 4px;
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

