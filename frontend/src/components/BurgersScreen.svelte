<script>
  import { tick } from 'svelte';
  import gsap from 'gsap';
  import { store } from '../lib/store.svelte.js';
  import { 
    ChevronLeft, 
    ChevronRight, 
    Heart, 
    Check, 
    Sparkles, 
    ArrowLeft,
    X,
    ShoppingBag,
    Plus,
    Minus,
    Trash2
  } from 'lucide-svelte';

  let { onBack } = $props();

  let isClosing = $state(false);
  let isEntering = $state(true);
  let showIngredients = $state(false);

  let windowWidth = $state(typeof window !== 'undefined' ? window.innerWidth : 1200);
  let windowHeight = $state(typeof window !== 'undefined' ? window.innerHeight : 800);
  let isStandalone = $state(
    typeof window !== 'undefined' && (
      window.matchMedia('(display-mode: standalone)').matches ||
      window.navigator.standalone === true
    )
  );

  let ingredientsBtnEl = $state(null);
  let cardEl = $state(null);
  let modalBackdropEl = $state(null);
  let isModalAnimating = false;

  let orderBtnEl = $state(null);

  async function openIngredientsModal(e) {
    if (e) e.stopPropagation();
    if (isModalAnimating || showIngredients) return;
    isModalAnimating = true;

    // Feedback táctil del botón (GPU)
    if (ingredientsBtnEl) {
      gsap.fromTo(ingredientsBtnEl, 
        { scale: 0.94 }, 
        { scale: 1, duration: 0.25, ease: 'power2.out', force3D: true }
      );
    }

    showIngredients = true;
    await tick();

    if (!cardEl || !modalBackdropEl) {
      isModalAnimating = false;
      return;
    }

    let originX = '50%';
    let originY = '50%';

    if (ingredientsBtnEl) {
      const btnRect = ingredientsBtnEl.getBoundingClientRect();
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
      }
    });

    // 1. Fondo oscuro cinematográfico (Misma animación que inicio)
    tl.to(modalBackdropEl, {
      opacity: 1,
      duration: 0.28,
      ease: 'power2.out',
      force3D: true
    }, 0);

    // 2. Expansión fluida de la tarjeta estilo Apple Spring (Misma animación que inicio)
    tl.to(cardEl, {
      scale: 1,
      opacity: 1,
      duration: 0.30,
      ease: 'power2.out',
      force3D: true
    }, 0);
  }

  function closeIngredientsModal(e) {
    if (e) e.stopPropagation();
    if (isModalAnimating || !showIngredients) return;
    if (!cardEl || !modalBackdropEl) {
      showIngredients = false;
      return;
    }

    isModalAnimating = true;

    let originX = '50%';
    let originY = '50%';

    if (ingredientsBtnEl) {
      const btnRect = ingredientsBtnEl.getBoundingClientRect();
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
        showIngredients = false;
        isModalAnimating = false;
      }
    });

    // Colapso en espejo de la tarjeta hacia el botón (Misma animación que inicio)
    tl.to(cardEl, {
      scale: 0.18,
      opacity: 0,
      duration: 0.22,
      ease: 'power2.in',
      force3D: true
    }, 0);

    // Desvanecimiento del fondo oscuro
    tl.to(modalBackdropEl, {
      opacity: 0,
      duration: 0.20,
      ease: 'power2.in',
      force3D: true
    }, 0.02);
  }

  function openOrderModal(e) {
    if (e) e.stopPropagation();
    if (orderBtnEl) {
      gsap.fromTo(orderBtnEl, 
        { scale: 0.94 }, 
        { scale: 1, duration: 0.25, ease: 'power2.out', force3D: true }
      );
    }
    store.isCartOpen = true;
  }

  $effect(() => {
    // Bloquear scroll de la página de fondo mientras la pantalla esté activa
    const prevOverflow = document.body.style.overflow;
    document.body.style.overflow = 'hidden';

    // Cerrar modal con la tecla Escape
    const handleKeyDown = (e) => {
      if (e.key === 'Escape') {
        if (showIngredients) {
          closeIngredientsModal();
        } else if (showOrder) {
          closeOrderModal();
        }
      }
    };
    window.addEventListener('keydown', handleKeyDown);

    // Activar entrada fluida en el siguiente fotograma
    const timer = setTimeout(() => {
      isEntering = false;
    }, 20);

    return () => {
      clearTimeout(timer);
      window.removeEventListener('keydown', handleKeyDown);
      document.body.style.overflow = prevOverflow;
    };
  });

  function handleBack() {
    if (isClosing) return;
    isClosing = true;
    setTimeout(() => {
      onBack?.();
    }, 500);
  }

  // 6 Hamburguesas de Autor de Mías
  const burgers = [
    {
      id: 1,
      category_id: 1,
      slug: 'original',
      name: 'Original',
      shortDesc: 'Pan pretzel artesanal, 160g de res al carbón, tocineta crujiente y mozzarella fundido.',
      description: 'Pan pretzel, 160gr de carne de res premium, tocineta crujiente, queso mozzarella, cebolla caramelizada, tomate, salsa de la casa.',
      price: 17000,
      bread_type: 'pretzel',
      image_url: '/imagenes/hamburguesas/Orignal.webp',
      is_bestseller: true,
      is_chef_choice: false,
      badge: 'La Clásica de Autor',
      highlight: '160g Res al Carbón',
      highlightIndex: 2,
      ingredients: [
        'PAN PRETZEL ARTESANAL',
        '160G RES PREMIUM',
        'TOCINETA CRUJIENTE',
        'QUESO MOZZARELLA FUNDIDO',
        'CEBOLLA CARAMELIZADA',
        'TOMATE FRESCO',
        'SALSA DE LA CASA'
      ]
    },
    {
      id: 2,
      category_id: 1,
      slug: 'crispy',
      name: 'Crispy',
      shortDesc: 'Pan pretzel artesanal, pechuga apanada crujiente, coleslaw de repollo y miel mostaza.',
      description: 'Pan pretzel, carne de pechuga apanada, queso mozzarella, ensalada coleslaw de repollo, pepinillos agridulces encurtidos, salsa miel mostaza artesanal.',
      price: 17000,
      bread_type: 'pretzel',
      image_url: '/imagenes/hamburguesas/Crispy.webp',
      is_bestseller: false,
      is_chef_choice: false,
      badge: 'Pechuga Crujiente',
      highlight: 'Coleslaw & Miel Mostaza',
      highlightIndex: 1,
      ingredients: [
        'PAN PRETZEL ARTESANAL',
        'PECHUGA APANADA CRUJIENTE',
        'QUESO MOZZARELLA',
        'COLESLAW DE REPOLLO',
        'PEPINILLOS AGRIDULCES',
        'SALSA MIEL MOSTAZA'
      ]
    },
    {
      id: 3,
      category_id: 1,
      slug: 'dorada',
      name: 'Dorada',
      shortDesc: 'Pan pretzel, 160g de res al carbón, queso Paipa fundido, piña glaseada al ron oscuro y alioli.',
      description: 'Pan pretzel, 160gr de carne de res premium, queso Paipa fundido, piña glaseada al ron oscuro, tocineta crujiente, cebolla caramelizada, alioli de ajo rostizado artesanal.',
      price: 23000,
      bread_type: 'pretzel',
      image_url: '/imagenes/hamburguesas/Golden.webp',
      is_bestseller: true,
      is_chef_choice: true,
      badge: 'Elección del Chef',
      highlight: 'Queso Paipa & Ron Oscuro',
      highlightIndex: 2,
      ingredients: [
        'PAN PRETZEL ARTESANAL',
        '160G RES AL CARBÓN',
        'QUESO PAIPA FUNDIDO',
        'PIÑA GLASEADA AL RON',
        'TOCINETA CRUJIENTE',
        'CEBOLLA CARAMELIZADA',
        'ALIOLI DE AJO ROSTIZADO'
      ]
    },
    {
      id: 4,
      category_id: 1,
      slug: 'criolla',
      name: 'Criolla',
      shortDesc: 'Pan pretzel, 160g de res al carbón, suero costeño, plátano maduro en cubitos y cebolla morada.',
      description: 'Pan pretzel, 160gr de carne de res premium, base de suero costeño, cubitos de plátano maduro, cebolla morada encurtida, lechuga en chiffonade.',
      price: 23000,
      bread_type: 'pretzel',
      image_url: '/imagenes/hamburguesas/Criolla.webp',
      is_bestseller: false,
      is_chef_choice: false,
      badge: 'Sabor Tradición',
      highlight: 'Suero Costeño & Plátano',
      highlightIndex: 3,
      ingredients: [
        'PAN PRETZEL ARTESANAL',
        '160G RES AL CARBÓN',
        'SUERO COSTEÑO ARTESANAL',
        'PLÁTANO MADURO EN CUBOS',
        'CEBOLLA MORADA ENCURTIDA',
        'LECHUGA EN CHIFFONADE'
      ]
    },
    {
      id: 5,
      category_id: 1,
      slug: 'brava',
      name: 'Brava',
      shortDesc: 'Pan pretzel, 160g de res al carbón, queso colby jack, aguacate, cebolla al tequila y chipotle.',
      description: 'Pan pretzel, 160gr de carne de res premium, queso colby jack, crema de aguacate, mermelada de cebolla al tequila, slaw de repollo y cilantro, mayonesa de chipotle artesanal.',
      price: 26000,
      bread_type: 'pretzel',
      image_url: '/imagenes/hamburguesas/Brava.webp',
      is_bestseller: false,
      is_chef_choice: true,
      badge: 'Gourmet Picante',
      highlight: 'Cebolla al Tequila & Chipotle',
      highlightIndex: 4,
      ingredients: [
        'PAN PRETZEL ARTESANAL',
        '160G RES AL CARBÓN',
        'QUESO COLBY JACK',
        'CREMA DE AGUACATE',
        'CEBOLLA AL TEQUILA',
        'SLAW DE CILANTRO',
        'MAYONESA DE CHIPOTLE'
      ]
    },
    {
      id: 6,
      category_id: 1,
      slug: 'suprema',
      name: 'Suprema',
      shortDesc: 'Pan pretzel, doble carne premium de 320g, cheddar americano, tocineta y cebolla caramelizada.',
      description: 'Pan pretzel, doble carne premium (res, pechuga apanada o mixta), queso cheddar americano, cebolla caramelizada, tocineta crujiente, salsas de la casa.',
      price: 28000,
      bread_type: 'pretzel',
      image_url: '/imagenes/hamburguesas/Suprema.webp',
      is_bestseller: true,
      is_chef_choice: false,
      badge: 'Doble Poder 320g',
      highlight: 'Doble Res & Cheddar',
      highlightIndex: 1,
      ingredients: [
        'PAN PRETZEL ARTESANAL',
        'DOBLE RES 320G AL CARBÓN',
        'DOBLE CHEDDAR AMERICANO',
        'TOCINETA CRUJIENTE',
        'CEBOLLA CARAMELIZADA',
        'SALSAS DE LA CASA'
      ]
    }
  ];

  // Estado del carrusel de hamburguesas
  let currentIndex = $state(2); // Inicia en 'Dorada' (la favorita del chef)
  let currentBurger = $derived(burgers[currentIndex]);

  let isFavorite = $state(false);
  let isAddingToCart = $state(false);

  function formatCOP(val) {
    return new Intl.NumberFormat('es-CO', {
      style: 'currency',
      currency: 'COP',
      maximumFractionDigits: 0
    }).format(val);
  }

  function nextBurger() {
    currentIndex = (currentIndex + 1) % burgers.length;
    isFavorite = false;
  }

  function prevBurger() {
    currentIndex = (currentIndex - 1 + burgers.length) % burgers.length;
    isFavorite = false;
  }

  function selectBurger(idx) {
    if (idx === currentIndex) return;
    currentIndex = idx;
    isFavorite = false;
  }

  // Detección de Swipe táctil en ruleta / media luna para móvil
  let touchStartX = 0;
  let touchStartY = 0;
  let dragX = $state(0);
  let isDragging = $state(false);

  // Navegación por teclado en escritorio (Flechas izquierda/derecha y Escape para volver/cerrar)
  function handleKeyDown(e) {
    if (e.key === 'Escape') {
      if (showIngredients) {
        closeIngredientsModal(e);
      } else if (showOrder) {
        closeOrderModal(e);
      } else {
        handleBack();
      }
      return;
    }
    if (showIngredients || showOrder) return;
    if (e.key === 'ArrowLeft') prevBurger();
    if (e.key === 'ArrowRight') nextBurger();
  }

  function handleTouchStart(e) {
    if (e.touches && e.touches[0]) {
      touchStartX = e.touches[0].clientX;
      touchStartY = e.touches[0].clientY;
      isDragging = true;
      dragX = 0;
    }
  }

  function handleTouchMove(e) {
    if (!isDragging || !e.touches || !e.touches[0]) return;
    const dx = e.touches[0].clientX - touchStartX;
    const dy = e.touches[0].clientY - touchStartY;
    if (Math.abs(dx) > Math.abs(dy)) {
      dragX = dx;
      if (e.cancelable) e.preventDefault();
    }
  }

  function handleTouchEnd() {
    if (!isDragging) return;
    isDragging = false;
    const threshold = 40;
    if (dragX < -threshold) {
      nextBurger();
    } else if (dragX > threshold) {
      prevBurger();
    }
    dragX = 0;
  }

  // Cálculo de posición relativa circular [-3, 2] respecto a currentIndex
  function getRelativeOffset(i, current, length) {
    let diff = i - current;
    while (diff > length / 2) diff -= length;
    while (diff <= -length / 2) diff += length;
    return diff;
  }

  function getBurgerItemStyle(i, current, length, currentDragX, dragging) {
    const offset = getRelativeOffset(i, current, length);
    
    // Si está fuera del campo visual [-2, 2], ocultar
    if (Math.abs(offset) > 2) {
      return 'display: none;';
    }

    // Espaciado adaptativo: 260px en desktop, escalado suave en móviles angostos
    const spacing = Math.min(260, Math.max(185, Math.round(windowWidth * 0.54)));
    const rawX = offset * spacing + currentDragX;
    const absX = Math.abs(rawX);

    // Trayectoria suave en arco de media luna: parábola sutil
    const y = Math.min(75, Math.pow(rawX * 0.075, 2));
    const rotate = rawX * 0.065;
    const scale = Math.max(0.68, 1 - absX / 650);

    // Opacidad continua (1 en centro, 0.3 en laterales en reposo, 0 más allá)
    let opacity = 0;
    if (absX < 60) {
      opacity = 1;
    } else if (absX < spacing * 1.5) {
      const p = (absX - 60) / (spacing * 1.5 - 60);
      opacity = Math.max(0.28, 1 - p * 0.72);
    }

    const zIndex = 20 - Math.round(absX / 15);
    const transition = dragging 
      ? 'none' 
      : 'transform 0.42s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.38s ease-out';

    const cursor = offset === 0 ? 'default' : 'pointer';

    return `transform: translate3d(${rawX}px, ${y}px, 0) rotate(${rotate}deg) scale(${scale}); opacity: ${opacity}; z-index: ${zIndex}; transition: ${transition}; cursor: ${cursor}; will-change: transform, opacity;`;
  }

  // Añadir al pedido
  function handleAddToCart() {
    isAddingToCart = true;
    store.addToCart(currentBurger, false, []);

    setTimeout(() => {
      isAddingToCart = false;
    }, 800);
  }
</script>

<svelte:window onkeydown={handleKeyDown} bind:innerWidth={windowWidth} bind:innerHeight={windowHeight} />

<section 
  class="fixed inset-0 z-40 h-[100dvh] w-full bg-[#FAF6EE] text-[#3A3A3C] flex flex-col justify-between select-none overflow-x-hidden overflow-y-hidden will-change-[transform,opacity] transition-all duration-500 {isEntering ? 'opacity-0 translate-y-8 scale-[0.98]' : isClosing ? 'opacity-0 translate-y-8 scale-[0.98] pointer-events-none' : 'opacity-100 translate-y-0 scale-100'} {isStandalone ? 'is-pwa-standalone' : ''}"
  style="transition-timing-function: cubic-bezier(0.16, 1, 0.3, 1); padding-top: max(0.5rem, calc(env(safe-area-inset-top, 0px) + 0.35rem)); padding-bottom: max({isStandalone ? '2rem' : '1.2rem'}, calc(env(safe-area-inset-bottom, 0px) + {isStandalone ? '1.5rem' : '0.85rem'}));"
>

  <!-- CONTENEDOR CENTRAL LIMITADO EN ESCRITORIO -->
  <div class="w-full max-w-md md:max-w-4xl lg:max-w-5xl mx-auto flex-1 flex flex-col justify-between min-h-0 overflow-visible">
    
    <!-- 1. BARRA SUPERIOR: BOTÓN ATRÁS (<), BADGE CATEGORÍA Y BOTÓN FAVORITO -->
    <header class="flex items-center justify-between py-1.5 sm:py-2 z-30 shrink-0 px-5 sm:px-6">
      <!-- Botón Atrás Circular Blanco Plano (Mismo tamaño y diseño que Home) -->
      <button 
        type="button"
        onclick={handleBack}
        class="w-14 h-14 sm:w-16 sm:h-16 rounded-full bg-white border border-black/5 flex items-center justify-center text-[#3A3A3C] hover:bg-black/5 active:scale-90 transition-all cursor-pointer group shrink-0"
        aria-label="Volver a la carta principal"
      >
        <ChevronLeft class="w-6 h-6 sm:w-7 sm:h-7 stroke-[2.2] -translate-x-0.5 group-hover:-translate-x-1 transition-transform" />
      </button>

      <!-- Indicador de Posición -->
      <div class="flex items-center gap-1.5 font-mono text-xs font-bold text-[#3A3A3C]">
        <span>0{currentIndex + 1}</span>
        <span class="text-[10px] text-ash/60">/</span>
        <span class="text-[11px] text-ash/80">0{burgers.length}</span>
      </div>

      <!-- Botón Favorito / Corazón Circular (Mismo tamaño y diseño que Home) -->
      <button 
        type="button"
        onclick={() => isFavorite = !isFavorite}
        class="w-14 h-14 sm:w-16 sm:h-16 rounded-full bg-white border border-black/5 flex items-center justify-center transition-all cursor-pointer active:scale-95 shrink-0 {isFavorite ? 'text-[#FF2400] border-[#FF2400]/40' : 'text-ash hover:text-[#3A3A3C]'}"
        aria-label="Marcar como favorita"
      >
        <Heart class="w-6 h-6 sm:w-6.5 sm:h-6.5 {isFavorite ? 'fill-[#FF2400]' : ''}" />
      </button>
    </header>

    <!-- 2. TÍTULO DE LA HAMBURGUESA -->
    <div class="text-center pt-1 sm:pt-2 pb-1 sm:pb-2 z-20 shrink-0 px-5 sm:px-6">
      <h1 class="font-flourish text-[2.45rem] sm:text-4xl md:text-5xl lg:text-6xl font-medium tracking-tight text-[#FF2400] transition-all duration-300 leading-none">
        {currentBurger.name}
      </h1>
    </div>

    <!-- 3. ÁREA HERO: SHOWCASE DE LA HAMBURGUESA EN RULETA / MEDIA LUNA (REFERENCIAS 1 & 2) -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div 
      class="relative flex-1 min-h-0 flex items-center justify-center my-0 sm:my-1 md:my-2 touch-pan-y overflow-visible w-full"
      ontouchstart={handleTouchStart}
      ontouchmove={handleTouchMove}
      ontouchend={handleTouchEnd}
    >
      <!-- Marca de Agua Arquitectónica en Escritorio (Referencia 2) -->
      <div class="hidden md:flex absolute inset-0 items-center justify-center pointer-events-none select-none overflow-hidden">
        <span class="font-display font-black text-black/[0.025] tracking-[0.25em] text-5xl lg:text-7xl xl:text-8xl uppercase whitespace-nowrap">
          MÍAS • AL CARBÓN
        </span>
      </div>

      <!-- Controles Laterales en Escritorio (Referencia 2) -->
      <button
        type="button"
        onclick={prevBurger}
        class="hidden md:flex absolute left-2 lg:left-6 z-30 w-11 h-11 lg:w-12 lg:h-12 rounded-full bg-white/85 hover:bg-white border border-black/10 text-[#3A3A3C] items-center justify-center shadow-xs hover:shadow-md hover:scale-105 active:scale-95 transition-all cursor-pointer group"
        aria-label="Ver hamburguesa anterior"
      >
        <ChevronLeft class="w-5 h-5 lg:w-6 lg:h-6 stroke-[2] group-hover:-translate-x-0.5 transition-transform" />
      </button>

      <button
        type="button"
        onclick={nextBurger}
        class="hidden md:flex absolute right-2 lg:right-6 z-30 w-11 h-11 lg:w-12 lg:h-12 rounded-full bg-white/85 hover:bg-white border border-black/10 text-[#3A3A3C] items-center justify-center shadow-xs hover:shadow-md hover:scale-105 active:scale-95 transition-all cursor-pointer group"
        aria-label="Ver hamburguesa siguiente"
      >
        <ChevronRight class="w-5 h-5 lg:w-6 lg:h-6 stroke-[2] group-hover:translate-x-0.5 transition-transform" />
      </button>

      <!-- Escenario Ruleta / Media Luna Continuo y Permanente (60fps en Mobile y Desktop sin parpadeo) -->
      <div class="burger-roulette-container relative z-20 flex items-center justify-center w-full max-w-[260px] sm:max-w-[310px] md:max-w-[340px] lg:max-w-[360px] max-h-[32vh] aspect-square -translate-y-16 sm:-translate-y-18 md:-translate-y-12 lg:-translate-y-14">
        <!-- Sombra elíptica independiente en el suelo -->
        <div 
          class="absolute -bottom-2 sm:-bottom-3 w-40 sm:w-52 h-5 rounded-[100%] bg-black/10 blur-[5px] pointer-events-none transition-transform duration-300"
          style="transform: scale({isDragging ? Math.max(0.68, 1 - Math.abs(dragX) / 400) : 1});"
        ></div>

        <!-- Pista de 6 hamburguesas montadas de forma persistente en GPU (cero parpadeos, cero re-montajes) -->
        {#each burgers as b, i (b.id)}
          {@const itemStyle = getBurgerItemStyle(i, currentIndex, burgers.length, dragX, isDragging)}
          {@const offset = getRelativeOffset(i, currentIndex, burgers.length)}
          <!-- svelte-ignore a11y_click_events_have_key_events -->
          <!-- svelte-ignore a11y_no_static_element_interactions -->
          <div 
            class="absolute inset-0 flex items-center justify-center select-none"
            style={itemStyle}
            onclick={() => {
              if (offset === -1) prevBurger();
              if (offset === 1) nextBurger();
            }}
          >
            <img 
              src={b.image_url} 
              alt={b.name} 
              class="w-full h-full max-h-[32vh] object-contain select-none pointer-events-none {b.slug === 'original' ? 'scale-[1.18]' : ''}"
              loading="eager"
              decoding="async"
            />
          </div>
        {/each}

        <!-- Indicadores de Posición / Dots al Nivel Inferior de la Hamburguesa -->
        <div class="absolute -bottom-7 sm:-bottom-8 left-1/2 -translate-x-1/2 flex items-center gap-1.5 z-30 pointer-events-auto">
          {#each burgers as b, idx (b.id)}
            <button
              type="button"
              onclick={(e) => { e.stopPropagation(); selectBurger(idx); }}
              ontouchstart={(e) => e.stopPropagation()}
              class="h-2 rounded-full transition-all duration-300 cursor-pointer {idx === currentIndex ? 'w-7 bg-[#FF2400]' : 'w-2 bg-black/15 hover:bg-black/30'}"
              aria-label="Ir a {b.name}"
            ></button>
          {/each}
        </div>

        <!-- Botones Gordos Grandes en Píldora con Curva Apple (Mismo grosor y altura que la carta) -->
        <div class="pill-buttons-container absolute top-[calc(100%+4.4rem)] sm:top-[calc(100%+5rem)] left-1/2 -translate-x-1/2 w-[calc(100vw-2.5rem)] max-w-md md:max-w-xl flex items-center justify-center gap-3 sm:gap-4 z-30 pointer-events-auto">
          <!-- 1. Botón Ingredientes (Blanco, gordo alto, estilo carta) -->
          <button
            bind:this={ingredientsBtnEl}
            type="button"
            onclick={openIngredientsModal}
            class="pill-action-btn flex-1 min-h-[76px] sm:min-h-[82px] md:min-h-[88px] rounded-full bg-white text-[#3A3A3C] font-display font-extrabold text-lg sm:text-xl md:text-2xl tracking-tight border border-black/30 shadow-xs hover:shadow-sm hover:bg-[#FAF6EE] active:scale-[0.97] transition-all duration-150 flex items-center justify-center cursor-pointer select-none px-4"
          >
            Ingredientes
          </button>

          <!-- 2. Botón Añadir (Rojo vivo, gordo alto, estilo carta) -->
          <button
            type="button"
            onclick={(e) => { e.stopPropagation(); handleAddToCart(); }}
            ontouchstart={(e) => e.stopPropagation()}
            disabled={isAddingToCart}
            class="pill-action-btn flex-1 min-h-[76px] sm:min-h-[82px] md:min-h-[88px] rounded-full bg-[#FF2400] text-white font-display font-extrabold text-lg sm:text-xl md:text-2xl tracking-tight border border-[#FF2400] shadow-sm hover:shadow-md hover:bg-[#E02000] active:scale-[0.97] transition-all duration-150 flex items-center justify-center gap-2 cursor-pointer select-none px-4"
          >
            {#if isAddingToCart}
              <Check class="w-6 h-6 stroke-[2.5] animate-in zoom-in-50 duration-150" />
              <span>Añadido</span>
            {:else}
              <span>Añadir</span>
            {/if}
          </button>
        </div>
      </div>
    </div>

    <!-- 4. BARRA INFERIOR DE ACCIÓN: PRECIO Y MI ORDEN -->
    <div class="burger-bottom-bar flex items-center justify-between pt-2.5 pb-1 border-t border-black/[0.06] mt-auto shrink-0 z-20 px-5 sm:px-6">
      <!-- Precio a la izquierda -->
      <div class="flex flex-col">
        <span class="font-mono text-[10px] sm:text-xs uppercase text-ash tracking-wider">
          Precio
        </span>
        <span class="font-display font-extrabold text-2xl sm:text-3xl lg:text-4xl text-[#3A3A3C] tracking-tight">
          {formatCOP(currentBurger.price)}
        </span>
      </div>

      <!-- Botón Mi Orden (Mismo tamaño, grosor y diseño que los otros dos botones) -->
      <button
        bind:this={orderBtnEl}
        type="button"
        onclick={openOrderModal}
        class="pill-action-btn min-h-[76px] sm:min-h-[82px] md:min-h-[88px] rounded-full bg-white text-[#3A3A3C] font-display font-extrabold text-lg sm:text-xl md:text-2xl tracking-tight border border-black/30 shadow-xs hover:shadow-sm hover:bg-[#FAF6EE] active:scale-[0.97] transition-all duration-150 flex items-center justify-center gap-3 cursor-pointer select-none px-6 sm:px-8 shrink-0"
        aria-label="Ver mi orden ({store.cartCount} productos añadidos)"
      >
        <span>Mi orden</span>
        {#if store.cartCount > 0}
          <span class="w-7 h-7 sm:w-8 sm:h-8 rounded-full bg-[#FF2400] text-white text-xs sm:text-sm font-mono font-bold flex items-center justify-center shadow-xs animate-in zoom-in duration-200 shrink-0">
            {store.cartCount}
          </span>
        {/if}
      </button>
    </div>

  </div>

  <!-- TARJETA EDITORIAL DE INGREDIENTES (IDÉNTICA A LA TARJETA DE LA PANTALLA DE INICIO CON ANIMACIÓN GSAP) -->

  <!-- TARJETA EDITORIAL DE INGREDIENTES (IDÉNTICA A LA TARJETA DE LA PANTALLA DE INICIO CON ANIMACIÓN GSAP) -->
  {#if showIngredients}
    <!-- svelte-ignore a11y_click_events_have_key_events -->
    <!-- svelte-ignore a11y_no_static_element_interactions -->
    <div 
      bind:this={modalBackdropEl}
      data-lenis-prevent
      class="welcome-modal-backdrop fixed inset-0 h-[100dvh] z-[100] flex items-center justify-center bg-[#050507]/94 lg:backdrop-blur-sm cursor-pointer will-change-[opacity]"
      style="touch-action: pan-y;"
      onclick={closeIngredientsModal}
    >
      <div 
        bind:this={cardEl}
        data-lenis-prevent
        class="welcome-card-modal w-full h-full lg:max-w-[94vw] xl:max-w-[1440px] 2xl:max-w-[1560px] lg:h-[88vh] xl:h-[90vh] bg-[#FAF6EE] shadow-2xl lg:shadow-[0_25px_60px_-15px_rgba(0,0,0,0.6)] relative border border-black/30 rounded-[38px] sm:rounded-[48px] overflow-hidden cursor-default flex flex-col p-6 sm:p-10 md:p-14 lg:p-16 justify-between text-[#3A3A3C]"
        style="container-type: inline-size; touch-action: pan-y;"
        onclick={(e) => e.stopPropagation()}
      >
        <!-- Botón Cerrar Flotante Apple Standard (top-5 right-5 idéntico a IntroScreen) -->
        <button 
          type="button"
          onclick={closeIngredientsModal}
          class="absolute top-5 right-5 lg:top-7 lg:right-7 w-10 h-10 rounded-full bg-black/[0.06] hover:bg-black/[0.12] active:scale-95 flex items-center justify-center text-char/80 hover:text-char transition-all cursor-pointer shadow-sm z-30 shrink-0"
          aria-label="Cerrar ingredientes"
        >
          <X class="w-5 h-5 stroke-[2.2]" />
        </button>

        <!-- Nombre de la Hamburguesa (Editorial de alta legibilidad Playfair Display / Cormorant) -->
        <header class="select-none pr-14 shrink-0">
          <h2 
            class="burger-modal-title uppercase tracking-tight leading-none text-3xl sm:text-5xl md:text-6xl lg:text-7xl text-[#1A1A1D] font-extrabold"
            style="font-family: 'Playfair Display', 'Cormorant Garamond', Georgia, serif;"
          >
            {currentBurger.name}
          </h2>
        </header>

        <!-- Lista de Ingredientes (Tipografía Editorial de alta legibilidad, índices 01, 02... y resalte rojo como en la referencia) -->
        <div class="ingredients-list-container my-auto py-6 sm:py-8 lg:py-10 flex flex-col space-y-2.5 sm:space-y-3.5 md:space-y-4 lg:space-y-4.5 select-none overflow-y-auto" style="scrollbar-width: none;">
          {#each currentBurger.ingredients as item, idx}
            {@const isHighlighted = idx === currentBurger.highlightIndex}
            <div class="flex items-start group">
              <span 
                class="ingredient-item-text uppercase tracking-normal leading-[1.14] select-none text-xl sm:text-2xl md:text-3xl lg:text-4xl xl:text-5xl transition-colors duration-200 {isHighlighted ? 'text-[#FF2400] font-bold' : 'text-[#1A1A1D] font-semibold'}"
                style="font-family: 'Playfair Display', 'Cormorant Garamond', Georgia, serif;"
              >
                {item}
              </span>
              <span 
                class="ingredient-item-num font-mono text-[10px] sm:text-xs md:text-sm font-semibold ml-1.5 sm:ml-2.5 pt-0.5 sm:pt-1 select-none leading-none shrink-0 {isHighlighted ? 'text-[#FF2400]' : 'text-ash/60'}"
              >
                {String(idx + 1).padStart(2, '0')}
              </span>
            </div>
          {/each}
        </div>
      </div>
    </div>
  {/if}

</section>

<style>
  .welcome-modal-backdrop {
    padding-left: 12px;
    padding-right: 12px;
    padding-bottom: 12px;
    padding-top: max(14px, calc(env(safe-area-inset-top, 0px) + 6px));
  }

  @media (min-width: 640px) {
    .welcome-modal-backdrop {
      padding: 1.25rem;
    }
  }

  @media (min-width: 768px) {
    .welcome-modal-backdrop {
      padding: 1.5rem;
    }
  }

  @media (min-width: 1024px) {
    .welcome-modal-backdrop {
      padding: 1.75rem;
    }
  }

  @media (min-width: 1280px) {
    .welcome-modal-backdrop {
      padding: 2.25rem;
    }
  }

  .welcome-card-modal {
    border: 1px solid rgba(0, 0, 0, 0.30);
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

  /* ========================================================
     ADAPTABILIDAD RESPONSIVA (VIEWPORT HEIGHT & WIDTH)
     Preserva el 100% del aspecto, posición y tamaño en pantallas
     estándar, y escala fluidamente en alturas y anchos compactos.
     ======================================================== */

  /* En pantallas de escritorio / laptop apaisadas con altura compacta (ej. 1270x626) */
  @media (orientation: landscape) and (max-height: 720px) {
    .burger-roulette-container {
      transform: translateY(-2.85rem) !important;
      max-height: 25vh !important;
    }
    .pill-buttons-container {
      top: calc(100% + 2.75rem) !important;
    }
    .pill-action-btn {
      min-height: 66px !important;
      font-size: 1.15rem !important;
    }
    .welcome-card-modal {
      padding: 1.75rem 2.25rem !important;
    }
    .burger-modal-title {
      font-size: clamp(2rem, 5.5vh, 3.25rem) !important;
    }
    .ingredients-list-container {
      padding-top: 0.75rem !important;
      padding-bottom: 0.75rem !important;
      gap: 0.5rem !important;
    }
    .ingredient-item-text {
      font-size: clamp(1.15rem, 3.6vh, 2.35rem) !important;
      line-height: 1.15 !important;
      letter-spacing: 0.012em !important;
    }
  }

  /* Modo apaisado ultra-compacto (< 540px de alto) */
  @media (orientation: landscape) and (max-height: 540px) {
    .burger-roulette-container {
      transform: translateY(-2.1rem) !important;
      max-height: 22vh !important;
    }
    .pill-buttons-container {
      top: calc(100% + 2.1rem) !important;
    }
    .pill-action-btn {
      min-height: 54px !important;
      font-size: 1rem !important;
    }
  }

  /* En móviles verticales (Brave, Safari, iOS, Android): tamaño pleno del héroe y separación natural de los botones */
  @media (orientation: portrait) {
    .burger-roulette-container {
      max-height: 35vh !important;
    }
    .pill-buttons-container {
      top: calc(100% + 4.5rem) !important;
    }
    .pill-action-btn {
      min-height: 76px !important;
      font-size: 1.22rem !important;
    }
  }

  /* Tarjeta modal en móvil vertical */
  @media (max-width: 640px) and (orientation: portrait) {
    .welcome-card-modal {
      padding: 1.75rem 1.5rem !important;
    }
    .burger-modal-title {
      font-size: clamp(2.1rem, 7.5vw, 3rem) !important;
    }
    .ingredients-list-container {
      gap: 0.85rem !important;
      padding-top: 1rem !important;
      padding-bottom: 1rem !important;
    }
    .ingredient-item-text {
      font-size: clamp(1.12rem, 4.4vw, 1.6rem) !important;
      line-height: 1.16 !important;
      letter-spacing: 0.012em !important;
    }
    .ingredient-item-num {
      font-size: 0.75rem !important;
    }
  }

  /* ========================================================
     PWA STANDALONE (SAFARI / IOS PWA EN PANTALLA DE INICIO)
     Eleva la barra inferior y el precio para que no queden pegados
     a la barra negra de inicio (home indicator) del iPhone.
     ======================================================== */
  @media all and (display-mode: standalone) {
    .burger-bottom-bar {
      padding-top: 1.25rem !important;
      padding-bottom: max(1.25rem, calc(env(safe-area-inset-bottom, 0px) + 0.5rem)) !important;
      margin-bottom: 0.75rem !important;
    }
  }

  .is-pwa-standalone .burger-bottom-bar {
    padding-top: 1.25rem !important;
    padding-bottom: max(1.25rem, calc(env(safe-area-inset-bottom, 0px) + 0.5rem)) !important;
    margin-bottom: 0.75rem !important;
  }
</style>
