<script>
  import { onMount } from 'svelte';
  import Lenis from 'lenis';
  import { store } from './lib/store.svelte.js';
  import IntroScreen from './components/IntroScreen.svelte';
  import Header from './components/Header.svelte';
  import HomeHero from './components/HomeHero.svelte';
  import BurgersScreen from './components/BurgersScreen.svelte';
  import CategoryScreen from './components/CategoryScreen.svelte';
  import ProductDetailModal from './components/ProductDetailModal.svelte';
  import FloatingCart from './components/FloatingCart.svelte';
  import StampBooklet from './components/StampBooklet.svelte';
  import HistorySection from './components/HistorySection.svelte';
  import StaffPanels from './components/StaffPanels.svelte';
  import AuthModal from './components/AuthModal.svelte';
  import SidebarMenu from './components/SidebarMenu.svelte';
  import { Sparkles, Award } from 'lucide-svelte';

  const ONE_HOUR_MS = 60 * 60 * 1000;
  const INTRO_STORAGE_KEY = 'mias_intro_seen_at';

  function isIntroSuppressed() {
    if (typeof window === 'undefined') return false;
    try {
      const saved = localStorage.getItem(INTRO_STORAGE_KEY);
      if (!saved) return false;
      const timestamp = parseInt(saved, 10);
      if (isNaN(timestamp)) return false;
      return (Date.now() - timestamp) < ONE_HOUR_MS;
    } catch {
      return false;
    }
  }

  let homeRootEl = $state(null);

  let selectedCategory = $state('hamburguesas');
  let introCompleted = $state(isIntroSuppressed());

  function handleIntroComplete() {
    introCompleted = true;
    try {
      localStorage.setItem(INTRO_STORAGE_KEY, Date.now().toString());
    } catch {
      // Ignorar restricciones en entornos con almacenamiento bloqueado
    }
    if (typeof document !== 'undefined') {
      document.documentElement.classList.remove('theme-intro');
      document.documentElement.classList.add('theme-home');
      document.documentElement.style.backgroundColor = '#FAF6EE';
      document.documentElement.style.colorScheme = 'light';
      const metaTheme = document.getElementById('theme-color-meta') || document.querySelector('meta[name="theme-color"]');
      if (metaTheme) metaTheme.setAttribute('content', '#FAF6EE');
      const metaApple = document.getElementById('apple-status-bar-meta') || document.querySelector('meta[name="apple-mobile-web-app-status-bar-style"]');
      if (metaApple) metaApple.setAttribute('content', 'default');
    }
  }

  // Inicializar Lenis Smooth Scroll y bloquear zoom/pellizco permitiendo solo deslizamiento vertical
  onMount(() => {
    const lenis = new Lenis({
      duration: 1.2,
      easing: (t) => Math.min(1, 1.001 - Math.pow(2, -10 * t)),
      smoothWheel: true
    });

    function raf(time) {
      lenis.raf(time);
      requestAnimationFrame(raf);
    }
    requestAnimationFrame(raf);

    // Bloquear gestos multitáctiles de pellizco (zoom)
    const preventGesture = (e) => e.preventDefault();
    window.addEventListener('gesturestart', preventGesture, { passive: false });
    window.addEventListener('gesturechange', preventGesture, { passive: false });

    // Bloquear zoom con touchpad / Ctrl + rueda en desktop (el scroll vertical libre)
    const preventCtrlWheel = (e) => {
      if (e.ctrlKey) e.preventDefault();
    };
    window.addEventListener('wheel', preventCtrlWheel, { passive: false });

    // Bloquear atajos de teclado de zoom (Ctrl/Cmd con +, -, =, 0)
    const preventZoomKeys = (e) => {
      if ((e.ctrlKey || e.metaKey) && ['+', '-', '=', '0'].includes(e.key)) {
        e.preventDefault();
      }
    };
    window.addEventListener('keydown', preventZoomKeys);

    if (typeof window !== 'undefined') {
      window.__lenis = lenis;
    }

    return () => {
      lenis.destroy();
      if (typeof window !== 'undefined') {
        delete window.__lenis;
      }
      window.removeEventListener('gesturestart', preventGesture);
      window.removeEventListener('gesturechange', preventGesture);
      window.removeEventListener('wheel', preventCtrlWheel);
      window.removeEventListener('keydown', preventZoomKeys);
    };
  });

  // Pausar Lenis y congelar scroll del lienzo cuando la comanda o el menú lateral estén abiertos
  $effect(() => {
    if (typeof window !== 'undefined') {
      if (store.isCartOpen || store.isSidebarOpen) {
        window.__lenis?.stop();
        document.body.style.overflow = 'hidden';
      } else {
        window.__lenis?.start();
        document.body.style.overflow = '';
      }
    }
  });

  const isMenuCategoryView = $derived(['menu', 'hamburguesas', 'perros', 'de_la_casa', 'adicionales', 'bebidas'].includes(store.currentView));
</script>

{#if !introCompleted}
  <IntroScreen onComplete={handleIntroComplete} />
{/if}

<div 
  bind:this={homeRootEl}
  class="min-h-screen flex flex-col bg-[#FAF6EE] text-char selection:bg-ember selection:text-white"
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

    {:else if store.currentView === 'hamburguesas'}
      <!-- Pantalla Completa de Hamburguesas de Autor -->
      <BurgersScreen onBack={() => store.currentView = 'menu'} />

    {:else if store.currentView === 'perros'}
      <!-- Pantalla Completa de Perros Calientes -->
      <CategoryScreen 
        categoryId={2}
        title="Perros Calientes"
        kicker="Mías • Línea de Autor"
        heroImage="/imagenes/perros/perro.png"
        heroSubtitle="Pan brioche artesanal dorado a la mantequilla con embutidos premium y salsas de la casa."
        featuredSlug="parrillero-choriperro"
        featuredBadge="★ RECOMENDADO DEL CHEF"
        defaultBadge="RECETA TRADICIONAL"
        breadInfo="Pan Brioche"
        onBack={() => store.currentView = 'menu'} 
      />

    {:else if store.currentView === 'de_la_casa'}
      <!-- Pantalla Completa De la Casa -->
      <CategoryScreen 
        categoryId={3}
        title="De la Casa"
        kicker="Mías • Tradición Gastronómica"
        heroImage="/imagenes/de_la_casa/salchi_papa.png"
        heroSubtitle="Recetas tradicionales de la casa preparadas con papas seleccionadas y carnes premium."
        featuredSlug="salchipapa-especial"
        featuredBadge="★ FAVORITA DE LA CASA"
        defaultBadge="ESPECIALIDAD"
        onBack={() => store.currentView = 'menu'} 
      />

    {:else if store.currentView === 'adicionales'}
      <!-- Pantalla Completa de Adicionales y Acompañamientos -->
      <CategoryScreen 
        categoryId={5}
        title="Adicionales"
        kicker="Mías • Acompañamientos y Extras"
        heroSubtitle="Personaliza cada bocado con porciones adicionales, salsas de autor y papas rústicas al carbón."
        featuredSlug="combo-papa-gaseosa"
        featuredBadge="★ COMBO RECOMENDADO"
        defaultBadge="PORCIÓN EXTRA"
        layout="grid"
        onBack={() => store.currentView = 'menu'} 
      />

    {:else if store.currentView === 'bebidas'}
      <!-- Pantalla Completa de Bebidas frías y en vidrio -->
      <CategoryScreen 
        categoryId={4}
        title="Bebidas"
        kicker="Mías • Línea en Vidrio y Refrescos"
        heroImage="/imagenes/bebidas/bretana-vidrio.png"
        heroSubtitle="Gaseosas frías y bebidas refrescantes servidas en vidrio tradicional."
        featuredSlug="bretana-vidrio"
        featuredBadge="★ CLÁSICA EN VIDRIO"
        defaultBadge="REFRESCO"
        onBack={() => store.currentView = 'menu'} 
      />

    {:else if store.currentView === 'stamps'}
      <!-- Vista de Cartilla de Sellos Digital -->
      <StampBooklet />

    {:else if store.currentView === 'history'}
      <!-- Vista de Historia y Galería Real -->
      <HistorySection />

    {:else if store.currentView === 'staff'}
      <!-- Terminal Operativa de Personal (Meseros, Cajero, Admin) -->
      <StaffPanels />
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
        Todos los días: 3:00 PM – 9:30 PM • Domicilios: 321 252 0881
      </p>
      <p class="text-[10px] text-ash/50 font-mono">
        Sistema de Alta Fidelización y Comandas • Estándar Senior de Ingeniería 2026
      </p>
    </footer>
  {/if}

  <!-- Modales y Componentes Flotantes -->
  <ProductDetailModal />
  <FloatingCart />
  <AuthModal />
  <SidebarMenu />
</div>
