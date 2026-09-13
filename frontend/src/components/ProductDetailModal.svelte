<script>
  import { store } from '../lib/store.svelte.js';
  import { X, Check, Plus, ShoppingBag } from 'lucide-svelte';

  let product = $derived(store.selectedProduct);
  let includeCombo = $state(false);
  let selectedExtras = $state([]);

  const availableExtras = [
    { id: 'papas-nativas', name: 'Papa Francesa Extra', price: 8000 },
    { id: 'tocineta', name: 'Tocineta Ahumada Extra', price: 5000 },
    { id: 'bretana', name: 'Bretaña en Vidrio', price: 5000 },
    { id: 'cocacola', name: 'Coca-Cola 400ml', price: 4000 }
  ];

  let totalPrice = $derived.by(() => {
    if (!product) return 0;
    let total = product.price;
    if (includeCombo) total += 11000;
    for (const ex of selectedExtras) {
      total += ex.price;
    }
    return total;
  });

  function toggleExtra(extra) {
    const idx = selectedExtras.findIndex((e) => e.id === extra.id);
    if (idx > -1) {
      selectedExtras = selectedExtras.filter((e) => e.id !== extra.id);
    } else {
      selectedExtras = [...selectedExtras, extra];
    }
  }

  function handleAddAndOrder() {
    store.addToCart(product, includeCombo, selectedExtras);
    store.selectedProduct = null;
    includeCombo = false;
    selectedExtras = [];
    store.isCartOpen = true; // Abre inmediatamente la comanda
  }

  function formatCOP(val) {
    return new Intl.NumberFormat('es-CO', { 
      style: 'currency', 
      currency: 'COP', 
      maximumFractionDigits: 0 
    }).format(val || 0);
  }
</script>

{#if product}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div 
    data-lenis-prevent
    onwheel={(e) => e.stopPropagation()}
    ontouchmove={(e) => e.stopPropagation()}
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-xs p-4 sm:p-6 overflow-y-auto overscroll-contain"
    style="touch-action: pan-y;"
    onclick={() => store.selectedProduct = null}
  >
    <div 
      data-lenis-prevent
      class="bg-char text-bone w-full max-w-lg rounded-[38px] sm:rounded-[44px] border border-white/10 overflow-hidden flex flex-col max-h-[90vh] shadow-2xl relative"
      onclick={(e) => e.stopPropagation()}
    >
      <!-- Botón de Cierre Circular Dark (Fat-Finger) -->
      <button 
        type="button"
        onclick={() => store.selectedProduct = null}
        class="absolute top-4 right-4 z-10 w-12 h-12 sm:w-14 sm:h-14 rounded-full bg-char-card text-bone flex items-center justify-center border border-white/10 hover:bg-char-card-hover active:scale-90 transition-all cursor-pointer shrink-0"
        aria-label="Cerrar detalle"
      >
        <X class="w-5 h-5 stroke-[2.2]" />
      </button>

      <!-- Cabecera Fotográfica en Gran Formato -->
      <div class="relative bg-char-card pt-8 pb-6 flex items-center justify-center border-b border-white/10 shrink-0">
        <img 
          src={product.image_url || '/imagenes/marca/isotipo.webp'} 
          alt={product.name}
          class="max-h-52 sm:max-h-60 w-auto object-contain select-none pointer-events-none drop-shadow-[0_12px_24px_rgba(0,0,0,0.6)]"
        />
      </div>

      <!-- Contenido con Scroll Aislado -->
      <div 
        class="p-5 sm:p-6 overflow-y-auto space-y-5 flex-1 min-h-0"
        style="touch-action: pan-y;"
      >
        <!-- Título y Tipo de Pan -->
        <div>
          <div class="flex items-center gap-2 mb-2">
            {#if product.bread_type === 'pretzel'}
              <span class="text-[11px] font-mono uppercase tracking-wider font-bold text-bone bg-char-card border border-white/10 px-3 py-1 rounded-full">
                Pan Pretzel 160g
              </span>
            {:else if product.bread_type === 'brioche'}
              <span class="text-[11px] font-mono uppercase tracking-wider font-bold text-bone bg-char-card border border-white/10 px-3 py-1 rounded-full">
                Pan Brioche Artesanal
              </span>
            {/if}
            <span class="text-xs font-mono text-ash">• Al Carbón Artesanal</span>
          </div>

          <h2 class="font-display font-black text-2xl sm:text-3xl text-bone tracking-tight leading-tight mb-1.5">
            {product.name}
          </h2>
          <p class="text-xs sm:text-sm text-ash leading-relaxed font-sans">
            {product.description}
          </p>
        </div>

        <!-- Tarjeta de Combo Promocional ($11.000) -->
        <button
          type="button"
          onclick={() => includeCombo = !includeCombo}
          class="w-full text-left p-4 sm:p-5 rounded-2xl border-2 transition-all cursor-pointer flex items-center justify-between {includeCombo ? 'bg-char-card border-[#FF2400] ring-1 ring-[#FF2400]' : 'bg-char-card border-white/10 hover:border-white/20'}"
        >
          <div class="flex items-center gap-3.5">
            <div class="w-6 h-6 rounded-md flex items-center justify-center border transition-colors {includeCombo ? 'bg-[#FF2400] border-[#FF2400] text-white' : 'border-white/20 text-transparent'}">
              <Check class="w-4 h-4 stroke-[3]" />
            </div>
            <div>
              <div class="flex items-center gap-2">
                <span class="text-xs font-display font-extrabold uppercase tracking-wide text-bone">Convertir en Combo</span>
                <span class="text-[10px] bg-[#FF2400] text-white font-mono px-2 py-0.5 rounded-full font-bold">RECOMENDADO</span>
              </div>
              <p class="text-xs text-ash mt-0.5">Papa a la francesa crujiente + Bebida en vidrio</p>
            </div>
          </div>
          <span class="font-display font-black text-sm text-[#FF2400] shrink-0">+ $11.000</span>
        </button>

        <!-- Adicionales Sugeridos (Tarjetas Planas Gordas) -->
        <div>
          <div class="flex items-center justify-between mb-2.5">
            <span class="text-xs font-mono font-bold uppercase tracking-wider text-ash">Adicionales</span>
            <span class="text-[11px] font-mono text-ash">Toca para sumar</span>
          </div>

          <div class="grid grid-cols-2 gap-2.5">
            {#each availableExtras as extra}
              {@const isSelected = selectedExtras.some((e) => e.id === extra.id)}
              <button
                type="button"
                onclick={() => toggleExtra(extra)}
                class="p-3.5 rounded-2xl border text-left transition-all cursor-pointer flex flex-col justify-between min-h-[72px] {isSelected ? 'bg-char-card border-[#FF2400] ring-1 ring-[#FF2400]' : 'bg-char-card border-white/10 hover:border-white/20'}"
              >
                <div class="flex items-center justify-between gap-1">
                  <span class="text-xs font-display font-bold text-bone line-clamp-1">{extra.name}</span>
                  <div class="w-4 h-4 rounded-full flex items-center justify-center border shrink-0 {isSelected ? 'bg-[#FF2400] border-[#FF2400] text-white' : 'border-white/20 text-transparent'}">
                    <Check class="w-3 h-3 stroke-[3]" />
                  </div>
                </div>
                <span class="font-mono text-xs text-[#FF2400] font-bold mt-1">+{formatCOP(extra.price)}</span>
              </button>
            {/each}
          </div>
        </div>
      </div>

      <!-- Barra Inferior Fija de Acción (Fat-Finger Button >= 60px) -->
      <div class="p-4 sm:p-5 bg-char border-t border-white/10 shrink-0">
        <button
          type="button"
          onclick={handleAddAndOrder}
          class="w-full min-h-[62px] rounded-full bg-[#FF2400] hover:bg-[#E02000] active:scale-[0.97] text-white font-display font-black text-base sm:text-lg tracking-tight flex items-center justify-between px-7 transition-all cursor-pointer shadow-none"
        >
          <div class="flex items-center gap-2">
            <ShoppingBag class="w-5 h-5" />
            <span>Añadir a la Comanda</span>
          </div>
          <span class="font-mono text-base font-bold">{formatCOP(totalPrice)}</span>
        </button>
      </div>
    </div>
  </div>
{/if}
