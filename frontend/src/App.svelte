<script>
  import { onMount } from 'svelte';
  import Lenis from 'lenis';
  import { store } from './lib/store.svelte.js';
  import Header from './components/Header.svelte';
  import HomeHero from './components/HomeHero.svelte';
  import BurgersScreen from './components/BurgersScreen.svelte';
  import CategoryScreen from './components/CategoryScreen.svelte';
  import ProductDetailModal from './components/ProductDetailModal.svelte';
  import FloatingCart from './components/FloatingCart.svelte';
  import StampBooklet from './components/StampBooklet.svelte';
  import StaffPanels from './components/StaffPanels.svelte';
  import AuthModal from './components/AuthModal.svelte';
  import IntroScreen from './components/IntroScreen.svelte';
  import OrdersScreen from './components/OrdersScreen.svelte';
  import { Sparkles, Award } from 'lucide-svelte';

  let homeRootEl = $state(null);
  let selectedCategory = $state('hamburguesas');
  let introCompleted = $state(false);

  // Comprobar persistencia de intro (TRD §5): Si se visualizó hace menos de 1 hora o es personal staff, omitir
  if (typeof window !== 'undefined') {
    try {
      const seenAt = localStorage.getItem('mias_intro_seen_at');
      const savedUser = localStorage.getItem('mias_user');
      const isStaffUser = savedUser && ['waiter', 'cashier', 'admin'].includes(JSON.parse(savedUser)?.role);
      if (isStaffUser || (seenAt && (Date.now() - parseInt(seenAt, 10)) < 3600 * 1000)) {
        introCompleted = true;
      }
    } catch {}
  }

  function handleIntroComplete() {
    introCompleted = true;
    if (typeof window !== 'undefined') {
      try {
        localStorage.setItem('mias_intro_seen_at', Date.now().toString());
      } catch {}
    }
  }

  // Inicializar Lenis Smooth Scroll solo en Desktop (preservando scroll nativo por GPU en iPhone)
  onMount(() => {
    const isTouchDevice = typeof window !== 'undefined' && 
      (('ontouchstart' in window) || (navigator.maxTouchPoints > 0) || window.matchMedia('(pointer: coarse)').matches);

    let lenis = null;
    let rafId = null;

    // En pantallas táctiles móviles (iPhone), el scroll nativo de Safari corre 100% por hardware a 120 FPS
    // sin loops de JavaScript en reposo, lo que mantiene el procesador frío y ahorra batería.
    if (!isTouchDevice) {
      lenis = new Lenis({
        duration: 1.2,
        easing: (t) => Math.min(1, 1.001 - Math.pow(2, -10 * t)),
        smoothWheel: true,
        prevent: (node) => node.closest?.('[data-lenis-prevent]') !== null
      });

      function raf(time) {
        lenis.raf(time);
        rafId = requestAnimationFrame(raf);
      }
      rafId = requestAnimationFrame(raf);
      window.__lenis = lenis;
    }

    return () => {
      if (lenis) {
        lenis.destroy();
      }
      if (rafId) {
        cancelAnimationFrame(rafId);
      }
      if (typeof window !== 'undefined') {
        delete window.__lenis;
      }
    };
  });

  // Pausar Lenis y congelar scroll del lienzo cuando la comanda, menú lateral o login estén abiertos
  $effect(() => {
    if (typeof window !== 'undefined') {
      if (store.isCartOpen || store.isSidebarOpen || store.isAuthModalOpen) {
        window.__lenis?.stop();
        document.body.style.overflow = 'hidden';
      } else {
        window.__lenis?.start();
        document.body.style.overflow = '';
      }
    }
  });

  const categoryConfig = {
    hamburguesas: {
      type: 'burgers'
    },
    perros: {
      kicker: 'Mías • Línea de Autor',
      heroImage: '/imagenes/perros/perro.png',
      heroSubtitle: 'Pan brioche artesanal dorado a la mantequilla con embutidos premium y salsas de la casa.',
      featuredSlug: 'parrillero-choriperro',
      featuredBadge: '★ RECOMENDADO DEL CHEF',
      defaultBadge: 'RECETA TRADICIONAL',
      breadInfo: 'Pan Brioche'
    },
    de_la_casa: {
      kicker: 'Mías • Tradición Gastronómica',
      heroImage: '/imagenes/de_la_casa/salchi_papa.png',
      heroSubtitle: 'Recetas tradicionales de la casa preparadas con papas seleccionadas y carnes premium.',
      featuredSlug: 'salchipapa-especial',
      featuredBadge: '★ FAVORITA DE LA CASA',
      defaultBadge: 'ESPECIALIDAD'
    },
    adicionales: {
      kicker: 'Mías • Acompañamientos y Extras',
      heroSubtitle: 'Personaliza cada bocado con porciones adicionales, salsas de autor y papas rústicas al carbón.',
      featuredSlug: 'combo-papa-gaseosa',
      featuredBadge: '★ COMBO RECOMENDADO',
      defaultBadge: 'PORCIÓN EXTRA',
      layout: 'grid'
    },
    bebidas: {
      kicker: 'Mías • Línea en Vidrio y Refrescos',
      heroImage: '/imagenes/bebidas/bretana-vidrio.png',
      heroSubtitle: 'Gaseosas frías y bebidas refrescantes servidas en vidrio tradicional.',
      featuredSlug: 'bretana-vidrio',
      featuredBadge: '★ CLÁSICA EN VIDRIO',
      defaultBadge: 'REFRESCO'
    }
  };

  const activeCategory = $derived(store.categories.find(c => c.slug === store.currentView));
  const isMenuCategoryView = $derived(['menu', ...store.categories.map(c => c.slug)].includes(store.currentView));
  const isStaff = $derived(Boolean(store.user && ['waiter', 'cashier', 'admin'].includes(store.user.role)) || store.currentView === 'staff');
</script>

{#if !introCompleted}
  <IntroScreen onComplete={handleIntroComplete} />
{/if}

{#if isStaff}
  <!-- TERMINAL EXCLUSIVA DEL PERSONAL (MESEROS / CAJA / ADMIN) -->
  <StaffPanels />
{:else}
  <!-- EXPERIENCIA EXCLUSIVA DEL CLIENTE -->
  <div 
    bind:this={homeRootEl}
    class="min-h-screen flex flex-col bg-char text-bone selection:bg-ember selection:text-white"
  >
    {#if !isMenuCategoryView}
      <!-- Encabezado exclusivo para paneles del sistema / sellos / historia -->
      <Header />
    {/if}

    <main class="flex-1">
      {#if store.currentView === 'menu'}
        <!-- Pantalla de Inicio Carta: Saludo, logo central, botón navegación, sellos y píldoras de categorías -->
        <HomeHero 
          selectedCategory={selectedCategory}
          onSelectCategory={(slug) => {
            selectedCategory = slug;
            store.currentView = slug;
          }}
        />

      {:else if activeCategory}
        {#if activeCategory.slug === 'hamburguesas' || activeCategory.id === 1}
          <!-- Pantalla Completa de Hamburguesas de Autor -->
          <BurgersScreen onBack={() => store.currentView = 'menu'} />
        {:else}
          {@const config = categoryConfig[activeCategory.slug] || {}}
          <!-- Pantalla Dinámica de Categoría Gobernada por DB -->
          <CategoryScreen 
            categoryId={activeCategory.id}
            title={activeCategory.name}
            kicker={activeCategory.metadata?.kicker || config.kicker || `Mías • ${activeCategory.name}`}
            heroImage={activeCategory.metadata?.heroImage || config.heroImage || ''}
            heroSubtitle={activeCategory.metadata?.heroSubtitle || config.heroSubtitle || `Platos seleccionados de ${activeCategory.name}.`}
            featuredSlug={activeCategory.metadata?.featuredSlug || config.featuredSlug || ''}
            featuredBadge={activeCategory.metadata?.featuredBadge || config.featuredBadge || '★ RECOMENDADO'}
            defaultBadge={activeCategory.metadata?.defaultBadge || config.defaultBadge || 'ESPECIALIDAD'}
            breadInfo={activeCategory.metadata?.breadInfo || config.breadInfo || ''}
            layout={activeCategory.metadata?.layout || config.layout || 'list'}
            onBack={() => store.currentView = 'menu'} 
          />
        {/if}

      {:else if store.currentView === 'orders'}
        <!-- Vista de Mis Pedidos y Seguimiento en Vivo -->
        <OrdersScreen onBack={() => store.currentView = 'menu'} />

      {:else if store.currentView === 'stamps'}
        <!-- Vista de Cartilla de Sellos Digital -->
        <StampBooklet />
      {/if}
    </main>

    {#if !isMenuCategoryView}
      <!-- Footer Editorial para vistas secundarias operativas -->
      <footer class="border-t hairline-dark bg-char-card py-10 px-4 text-center text-xs text-ash space-y-4">
        <div class="flex items-center justify-center gap-2">
          <img src="/imagenes/marca/isotipo.webp" alt="Mías" class="w-6 h-6 object-contain opacity-70" />
          <span class="font-serif text-lg font-bold text-bone tracking-wide">Mías Hamburguesas de Lujo</span>
        </div>
        <p class="font-mono text-[11px] text-ash/80">
          Sábado a Jueves: 3:00 PM – 9:30 PM (Viernes cerrado) • Domicilios: 321 252 0881 • Duitama, Boyacá
        </p>
        <p class="text-[10px] text-ash/50 font-mono">
          Duitama, Boyacá, Colombia • Todos los derechos reservados.
        </p>
      </footer>
    {/if}

    <!-- Modales y Componentes Flotantes del Cliente -->
    <ProductDetailModal />
    <FloatingCart />
  </div>
{/if}

<!-- Modal de Autenticación Unificado -->
<AuthModal />
