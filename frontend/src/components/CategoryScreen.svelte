<script>
  import { store } from '../lib/store.svelte.js';
  import { ChevronLeft, Plus, Check } from 'lucide-svelte';

  let {
    categoryId,
    title,
    kicker = 'Mías • Línea de Autor',
    heroImage = null,
    heroSubtitle = '',
    featuredSlug = null,
    featuredBadge = '★ RECOMENDADO DEL CHEF',
    defaultBadge = 'RECETA TRADICIONAL',
    breadInfo = null,
    layout = 'list',
    onBack
  } = $props();

  let addedItemId = $state(null);

  function formatCOP(val) {
    return new Intl.NumberFormat('es-CO', {
      style: 'currency',
      currency: 'COP',
      maximumFractionDigits: 0
    }).format(val);
  }

  // Filtrar productos correspondientes a la categoría
  let categoryProducts = $derived(
    store.products.filter(p => p.category_id === categoryId)
  );

  function handleAddToCart(product) {
    store.addToCart(product, false, []);
    addedItemId = product.id;
    setTimeout(() => {
      if (addedItemId === product.id) {
        addedItemId = null;
      }
    }, 1200);
  }

  function openOrder() {
    store.isCartOpen = true;
  }
</script>

<section 
  class="relative min-h-[100dvh] w-full flex flex-col justify-between bg-char text-bone select-none"
  style="padding-top: max(1rem, env(safe-area-inset-top, 0px)); padding-bottom: max(1.25rem, env(safe-area-inset-bottom, 0px)); padding-left: max(1.25rem, env(safe-area-inset-left, 0px)); padding-right: max(1.25rem, env(safe-area-inset-right, 0px));"
>
  <div class="w-full max-w-md md:max-w-4xl lg:max-w-5xl mx-auto flex-1 flex flex-col">
    
    <!-- 1. CABECERA CON BOTÓN ATRÁS Y TÍTULO EDITORIAL -->
    <header class="flex items-center justify-between py-2 sm:py-3 z-20 shrink-0">
      <button
        type="button"
        onclick={onBack}
        class="w-14 h-14 sm:w-16 sm:h-16 rounded-full bg-char-card border border-white/10 flex items-center justify-center text-bone hover:bg-char-card-hover active:scale-90 transition-all cursor-pointer group shrink-0 shadow-2xs"
        aria-label="Volver a la carta"
      >
        <ChevronLeft class="w-6 h-6 sm:w-7 sm:h-7 stroke-[2.2] -translate-x-0.5 group-hover:-translate-x-1 transition-transform" />
      </button>

      <div class="text-center flex-1 px-3">
        <span class="font-mono text-[10px] sm:text-xs uppercase tracking-widest text-ash font-bold block">
          {kicker}
        </span>
        <h1 class="font-flourish text-2xl sm:text-3xl md:text-4xl lg:text-5xl font-normal tracking-tight text-bone">
          {title}
        </h1>
      </div>

      <!-- Espaciador simétrico para centrado exacto del título -->
      <div class="w-14 sm:w-16 shrink-0"></div>
    </header>

    <!-- 2. HERO PRINCIPAL: IMAGEN O SUBTÍTULO EDITORIAL -->
    {#if heroImage}
      <div class="mt-3 sm:mt-5 mb-5 sm:mb-7 flex flex-col items-center justify-center text-center">
        <div class="relative w-full max-w-[320px] sm:max-w-[380px] md:max-w-[440px] aspect-[16/10] flex items-center justify-center">
          <div class="absolute bottom-1 w-52 sm:w-64 h-4 rounded-[100%] bg-black/40 blur-[5px] pointer-events-none"></div>
          <img 
            src={heroImage} 
            alt={title} 
            class="w-full h-full object-contain pointer-events-none select-none drop-shadow-[0_10px_20px_rgba(0,0,0,0.5)]"
            loading="eager"
          />
        </div>
        {#if heroSubtitle}
          <p class="font-sans text-xs sm:text-sm text-ash font-medium max-w-md mt-1">
            {heroSubtitle}
          </p>
        {/if}
      </div>
    {:else if heroSubtitle}
      <div class="text-center my-3 sm:my-4">
        <p class="font-sans text-xs sm:text-sm text-ash font-medium max-w-md mx-auto">
          {heroSubtitle}
        </p>
      </div>
    {/if}

    <!-- 3. LISTADO EDITORIAL DE PRODUCTOS -->
    {#if layout === 'grid'}
      <div class="grid grid-cols-1 md:grid-cols-2 gap-3.5 sm:gap-4 pb-24 sm:pb-28 mt-2">
        {#each categoryProducts as item (item.id)}
          {@const isAdded = addedItemId === item.id}
          {@const isFeatured = featuredSlug ? item.slug === featuredSlug : (item.is_chef_choice || item.is_bestseller)}
          <div class="bg-char-card rounded-[28px] sm:rounded-[34px] border {isFeatured ? 'border-[#FF2400]/40 shadow-sm' : 'border-white/10 shadow-2xs'} p-5 flex flex-col justify-between transition-all select-none hover:border-white/20">
            <div class="min-w-0 pr-1">
              <div class="flex items-center gap-1.5 mb-1">
                <span class="text-[10px] sm:text-xs font-mono font-bold uppercase tracking-wider {isFeatured ? 'text-[#FF2400]' : 'text-ash'}">
                  {isFeatured ? featuredBadge : defaultBadge}
                </span>
                {#if breadInfo}
                  <span class="text-[10px] sm:text-xs font-mono text-ash/70">• {breadInfo}</span>
                {/if}
              </div>

              <h2 class="font-display font-black text-lg sm:text-xl text-bone tracking-tight">
                {item.name}
              </h2>

              <p class="font-sans text-xs sm:text-sm text-ash leading-relaxed mt-1 line-clamp-2">
                {item.description}
              </p>
            </div>

            <div class="mt-4 pt-3 border-t border-white/10 flex items-center justify-between gap-3">
              <span class="font-display font-black text-lg sm:text-xl text-[#FF2400] tracking-tight">
                {formatCOP(item.price)}
              </span>

              <button
                type="button"
                onclick={() => handleAddToCart(item)}
                class="min-h-[48px] px-5 rounded-full {isAdded ? 'bg-emerald-600' : 'bg-[#FF2400] hover:bg-[#E02000]'} active:scale-[0.97] text-white font-display font-extrabold text-sm tracking-tight transition-all flex items-center justify-center gap-1.5 cursor-pointer shadow-xs"
                aria-label="Añadir {item.name}"
              >
                {#if isAdded}
                  <Check class="w-4 h-4 stroke-[2.5]" />
                  <span>Listo</span>
                {:else}
                  <Plus class="w-4 h-4 stroke-[2.5]" />
                  <span>Añadir</span>
                {/if}
              </button>
            </div>
          </div>
        {/each}
      </div>
    {:else}
      <div class="flex-1 space-y-3.5 sm:space-y-4 pb-24 sm:pb-28">
        {#each categoryProducts as item (item.id)}
          {@const isAdded = addedItemId === item.id}
          {@const isFeatured = featuredSlug ? item.slug === featuredSlug : (item.is_chef_choice || item.is_bestseller)}
          <div class="bg-char-card rounded-[30px] sm:rounded-[36px] border {isFeatured ? 'border-[#FF2400]/40 shadow-sm' : 'border-white/10 shadow-2xs'} p-5 sm:p-6 transition-all hover:border-white/20 flex flex-col sm:flex-row sm:items-center justify-between gap-4 select-none relative">
            <!-- Contenido Descriptivo -->
            <div class="flex-1 min-w-0 pr-0 sm:pr-4">
              <div class="flex items-center gap-2 mb-1">
                <span class="text-[10px] sm:text-xs font-mono font-bold uppercase tracking-wider {isFeatured ? 'text-[#FF2400]' : 'text-ash'}">
                  {isFeatured ? featuredBadge : defaultBadge}
                </span>
                {#if breadInfo}
                  <span class="text-[10px] sm:text-xs font-mono text-ash/70">• {breadInfo}</span>
                {/if}
              </div>

              <h2 class="font-display font-black text-xl sm:text-2xl text-bone tracking-tight">
                {item.name}
              </h2>

              <p class="font-sans text-xs sm:text-sm text-ash leading-relaxed mt-1 line-clamp-2">
                {item.description}
              </p>

              <div class="font-display font-black text-lg sm:text-xl text-[#FF2400] mt-2 tracking-tight">
                {formatCOP(item.price)}
              </div>
            </div>

            <!-- Botón Ergonómico "+ Añadir a la Orden" -->
            <div class="shrink-0 pt-1 sm:pt-0">
              <button
                type="button"
                onclick={() => handleAddToCart(item)}
                class="w-full sm:w-auto min-h-[58px] sm:min-h-[64px] px-6 sm:px-8 rounded-full {isAdded ? 'bg-emerald-600' : 'bg-[#FF2400] hover:bg-[#E02000]'} active:scale-[0.97] text-white font-display font-extrabold text-base sm:text-lg tracking-tight transition-all flex items-center justify-center gap-2 cursor-pointer shadow-xs hover:shadow"
                aria-label="Añadir {item.name} a mi orden"
              >
                {#if isAdded}
                  <Check class="w-5 h-5 stroke-[2.5]" />
                  <span>Añadido</span>
                {:else}
                  <Plus class="w-5 h-5 stroke-[2.5]" />
                  <span>Añadir</span>
                {/if}
              </button>
            </div>
          </div>
        {/each}
      </div>
    {/if}

  </div>

  <!-- 4. BARRA INFERIOR FIJA: BOTÓN MI ORDEN CON PRIORIDAD TÁCTIL -->
  <div class="fixed bottom-0 left-0 right-0 z-40 bg-gradient-to-t from-char via-char/95 to-transparent pt-6 pb-5 px-5 sm:px-8 pointer-events-none">
    <div class="max-w-md md:max-w-4xl lg:max-w-5xl mx-auto flex items-center justify-between pointer-events-auto gap-4">
      <div class="hidden sm:flex flex-col">
        <span class="font-mono text-xs uppercase tracking-wider text-ash">Mías • Orden Activa</span>
        <span class="font-display font-bold text-sm text-bone">{store.cartCount} productos en tu carrito</span>
      </div>

      <button
        type="button"
        onclick={openOrder}
        class="w-full sm:w-auto ml-auto min-h-[72px] sm:min-h-[76px] px-8 sm:px-10 rounded-full bg-char-card text-bone font-display font-extrabold text-lg sm:text-xl tracking-tight border border-white/10 shadow-lg hover:bg-char-card-hover active:scale-[0.97] transition-all flex items-center justify-center gap-3.5 cursor-pointer select-none"
        aria-label="Ver mi orden ({store.cartCount} productos)"
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
</section>
