<script>
  import { onMount } from 'svelte';
  import Lenis from 'lenis';
  import { store } from './lib/store.svelte.js';
  import IntroScreen from './components/IntroScreen.svelte';
  import Header from './components/Header.svelte';
  import CategoryBar from './components/CategoryBar.svelte';
  import ProductCard from './components/ProductCard.svelte';
  import ProductDetailModal from './components/ProductDetailModal.svelte';
  import FloatingCart from './components/FloatingCart.svelte';
  import OrderModeModal from './components/OrderModeModal.svelte';
  import DeliveryWizard from './components/DeliveryWizard.svelte';
  import StampBooklet from './components/StampBooklet.svelte';
  import HistorySection from './components/HistorySection.svelte';
  import StaffPanels from './components/StaffPanels.svelte';
  import AuthModal from './components/AuthModal.svelte';
  import { Sparkles, Award } from 'lucide-svelte';

  let selectedCategory = $state('hamburguesas');
  let introCompleted = $state(false);

  // Inicializar Lenis Smooth Scroll
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

    return () => {
      lenis.destroy();
    };
  });

  // Filtrar productos por categoría
  let filteredProducts = $derived(() => {
    const cat = store.categories.find(c => c.slug === selectedCategory);
    if (!cat) return store.products;
    return store.products.filter(p => p.category_id === cat.id);
  });
</script>

<!-- 1. Pantalla de Bienvenida (Intro Screen: 3.5s o clic, fondo crema, 60m cooldown) -->
<IntroScreen onComplete={() => introCompleted = true} />

<div class="min-h-screen flex flex-col bg-char text-bone selection:bg-ember selection:text-white">
  <!-- Encabezado Principal -->
  <Header />

  <main class="flex-1 pb-24">
    {#if store.currentView === 'menu'}
      <!-- Hero Banner Editorial Gastronómico -->
      <section class="relative px-4 py-8 sm:py-12 border-b hairline-dark bg-gradient-to-b from-char-card to-char text-center overflow-hidden">
        <div class="max-w-4xl mx-auto space-y-4">
          <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-white/5 border hairline-dark text-gold text-xs font-mono">
            <Sparkles class="w-3.5 h-3.5 text-gold" />
            <span>Pan Pretzel Artesanal • Al Carbón «Rayo»</span>
          </div>

          <h1 class="font-serif text-4xl sm:text-6xl font-bold tracking-tight text-bone">
            Hamburguesas de Lujo
          </h1>

          <p class="text-xs sm:text-sm text-ash max-w-lg mx-auto font-sans leading-relaxed">
            Lo íntimo hecho para compartir. Cada hamburguesa es una experiencia de alta cocina preparada con devoción al fuego del carbón.
          </p>

          <!-- Banner Promocional de Fidelización Rápido -->
          <button 
            type="button"
            onclick={() => store.currentView = 'stamps'}
            class="inline-flex items-center gap-3 p-2.5 px-4 rounded-full bg-ember/10 border border-ember/30 text-ember hover:bg-ember/20 transition-all cursor-pointer text-xs font-mono font-bold"
          >
            <Award class="w-4 h-4" />
            <span>10 Sellos = 1 Original Gratis</span>
            <span class="text-bone/70 font-normal">→ Ver mi progreso</span>
          </button>
        </div>
      </section>

      <!-- Selector de Categorías Suizo -->
      <CategoryBar 
        {selectedCategory} 
        onSelect={(slug) => selectedCategory = slug} 
      />

      <!-- Retícula de Productos -->
      <section class="max-w-6xl mx-auto px-4 py-8">
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
          {#each filteredProducts() as product (product.id)}
            <ProductCard {product} />
          {/each}
        </div>
      </section>

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

  <!-- Footer Editorial -->
  <footer class="border-t hairline-dark bg-char-card py-10 px-4 text-center text-xs text-ash space-y-4">
    <div class="flex items-center justify-center gap-2">
      <img src="/imagenes/marca/isotipo.webp" alt="Mías" class="w-6 h-6 object-contain opacity-70" />
      <span class="font-serif text-lg font-bold text-bone tracking-wide">Mías Hamburguesas de Lujo</span>
    </div>
    <p class="font-mono text-[11px] text-ash/80">
      Sábado a Jueves: 3:00 PM – 9:30 PM • Viernes: Cerrado • Domicilios: 321 252 0881
    </p>
    <p class="text-[10px] text-ash/50 font-mono">
      Sistema de Alta Fidelización y Comandas • Estándar Senior de Ingeniería 2026
    </p>
  </footer>

  <!-- Modales y Componentes Flotantes -->
  <ProductDetailModal />
  <FloatingCart />
  <OrderModeModal />
  <DeliveryWizard />
  <AuthModal />
</div>
