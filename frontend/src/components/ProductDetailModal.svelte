<script>
  import { store } from '../lib/store.svelte.js';
  import { X, Check, Plus, Flame, Sparkles } from 'lucide-svelte';

  let product = $derived(store.selectedProduct);
  let includeCombo = $state(false);
  let selectedExtras = $state([]);

  const availableExtras = [
    { id: 'papas-nativas', name: 'Papa Nativa', price: 9000, img: '/imagenes/de_la_casa/salchi_papa.png' },
    { id: 'tocineta', name: 'Tocineta Extra', price: 5000, img: '/imagenes/marca/isotipo.webp' },
    { id: 'bretana', name: 'Bretaña en Vidrio', price: 5000, img: '/imagenes/bebidas/bretana-vidrio.png' },
    { id: 'cocacola', name: 'Coca-Cola 400ml', price: 4000, img: '/imagenes/bebidas/cocacola.webp' }
  ];

  let totalPrice = $derived(() => {
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
    store.isCartOpen = true; // Despliega el carrito inferior
  }

  function formatPrice(val) {
    return new Intl.NumberFormat('es-CO', { style: 'currency', currency: 'COP', maximumFractionDigits: 0 }).format(val);
  }
</script>

{#if product}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div 
    class="fixed inset-0 z-50 flex items-end sm:items-center justify-center bg-black/80 backdrop-blur-md p-0 sm:p-4 overflow-y-auto"
    onclick={() => store.selectedProduct = null}
  >
    <div 
      class="bg-char-card w-full max-w-lg sm:rounded-3xl border hairline-dark overflow-hidden flex flex-col max-h-[92vh] sm:max-h-[88vh] shadow-2xl relative"
      onclick={(e) => e.stopPropagation()}
    >
      <!-- Botón de Cierre -->
      <button 
        onclick={() => store.selectedProduct = null}
        class="absolute top-4 right-4 z-10 w-9 h-9 rounded-full bg-char/70 text-bone flex items-center justify-center border hairline-dark hover:bg-ember transition-colors"
      >
        <X class="w-5 h-5" />
      </button>

      <!-- Cabecera Fotográfica en Gran Formato -->
      <div class="relative bg-gradient-to-b from-white/5 to-char-card pt-8 pb-4 flex items-center justify-center border-b hairline-dark">
        <img 
          src={product.image_url || '/imagenes/marca/isotipo.webp'} 
          alt={product.name}
          class="max-h-56 w-auto object-contain drop-shadow-2xl hover:scale-105 transition-transform duration-300"
        />
      </div>

      <!-- Contenido Scrollable -->
      <div class="p-6 overflow-y-auto space-y-6">
        <!-- Título y Pan -->
        <div>
          <div class="flex items-center gap-2 mb-2">
            {#if product.bread_type === 'pretzel'}
              <span class="text-xs font-mono uppercase tracking-widest text-gold bg-gold/10 px-2.5 py-0.5 rounded border border-gold/30">
                Pan Pretzel 160g
              </span>
            {:else if product.bread_type === 'brioche'}
              <span class="text-xs font-mono uppercase tracking-widest text-gold bg-gold/10 px-2.5 py-0.5 rounded border border-gold/30">
                Pan Brioche Artesanal
              </span>
            {/if}
            <span class="text-xs font-mono text-ash">• Al Carbón «Rayo»</span>
          </div>

          <h2 class="font-serif text-3xl sm:text-4xl font-bold text-bone tracking-tight mb-2">
            {product.name}
          </h2>
          <p class="text-sm text-ash leading-relaxed font-sans">
            {product.description}
          </p>
        </div>

        <!-- Banner de Combo Promocional ($11.000) -->
        <div 
          onclick={() => includeCombo = !includeCombo}
          class="p-4 rounded-2xl border transition-all cursor-pointer flex items-center justify-between {includeCombo ? 'bg-ember/15 border-ember shadow-ember' : 'bg-white/5 border-white/10 hover:border-gold/30'}"
        >
          <div class="flex items-center gap-3">
            <div class="w-6 h-6 rounded-md flex items-center justify-center border {includeCombo ? 'bg-ember border-ember text-white' : 'border-white/20 text-transparent'}">
              <Check class="w-4 h-4" />
            </div>
            <div>
              <div class="flex items-center gap-2">
                <span class="text-xs font-bold uppercase tracking-wider text-bone">Convertir en Combo</span>
                <span class="text-[10px] bg-ember text-white font-mono px-1.5 py-0.2 rounded font-bold">RECOMENDADO</span>
              </div>
              <p class="text-xs text-ash">Papa a la francesa crujiente + Gaseosa 400ml</p>
            </div>
          </div>
          <span class="font-mono text-sm font-bold text-gold">+ $11.000</span>
        </div>

        <!-- Retícula de Tarjetas Cuadradas de Acompañamiento (Inspirado en Surf Coffee - Ref 2) -->
        <div>
          <div class="flex items-center justify-between mb-3">
            <span class="text-xs font-mono uppercase tracking-widest text-ash">Adicionales sugeridos</span>
            <span class="text-[11px] font-mono text-gold/80">Selecciona para sumar</span>
          </div>

          <div class="grid grid-cols-2 gap-3">
            {#each availableExtras as extra}
              {@const isSelected = selectedExtras.some((e) => e.id === extra.id)}
              <div
                onclick={() => toggleExtra(extra)}
                class="p-3 rounded-xl border transition-all cursor-pointer flex flex-col justify-between {isSelected ? 'bg-gold/15 border-gold shadow-gold' : 'bg-white/5 border-white/10 hover:border-white/20'}"
              >
                <div class="flex items-center justify-between mb-2">
                  <span class="text-xs font-medium text-bone line-clamp-1">{extra.name}</span>
                  <div class="w-4 h-4 rounded-full flex items-center justify-center border {isSelected ? 'bg-gold border-gold text-char' : 'border-white/20 text-transparent'}">
                    <Check class="w-3 h-3" />
                  </div>
                </div>
                <span class="font-mono text-xs text-gold font-semibold">+{formatPrice(extra.price)}</span>
              </div>
            {/each}
          </div>
        </div>
      </div>

      <!-- Barra Inferior Fija de Acción (Thumb Zone) -->
      <div class="p-4 bg-char border-t hairline-dark pb-safe">
        <button
          onclick={handleAddAndOrder}
          class="w-full py-3.5 px-6 rounded-full bg-ember hover:bg-ember-light text-white font-mono font-bold tracking-wider text-sm flex items-center justify-between transition-all shadow-ember active:scale-[0.98]"
        >
          <span>Pedir y pagar</span>
          <span>{formatPrice(totalPrice())}</span>
        </button>
      </div>
    </div>
  </div>
{/if}
