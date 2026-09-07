<script>
  import { store } from '../lib/store.svelte.js';
  import { Plus, Flame, Star } from 'lucide-svelte';

  let { product } = $props();

  function formatPrice(val) {
    return new Intl.NumberFormat('es-CO', { style: 'currency', currency: 'COP', maximumFractionDigits: 0 }).format(val);
  }
</script>

<!-- svelte-ignore a11y_click_events_have_key_events -->
<!-- svelte-ignore a11y_no_static_element_interactions -->
<div 
  onclick={() => store.selectedProduct = product}
  class="group relative bg-char-card rounded-2xl border hairline-dark p-5 flex flex-col justify-between hover:border-gold/40 transition-all duration-300 cursor-pointer hover:shadow-lux hover:-translate-y-1"
>
  <!-- Badges Superiores -->
  <div class="flex items-center justify-between gap-2 mb-3">
    <div class="flex items-center gap-1.5">
      {#if product.bread_type === 'pretzel'}
        <span class="text-[10px] font-mono uppercase tracking-widest text-gold bg-gold/10 px-2 py-0.5 rounded border border-gold/20">
          Pan Pretzel 160g
        </span>
      {:else if product.bread_type === 'brioche'}
        <span class="text-[10px] font-mono uppercase tracking-widest text-gold bg-gold/10 px-2 py-0.5 rounded border border-gold/20">
          Brioche Artesanal
        </span>
      {/if}
    </div>

    <div class="flex items-center gap-1">
      {#if product.is_bestseller}
        <span class="flex items-center gap-1 text-[10px] font-bold uppercase tracking-wider text-ember bg-ember/15 px-2 py-0.5 rounded border border-ember/30">
          <Flame class="w-3 h-3" /> Top
        </span>
      {/if}
      {#if product.is_chef_choice}
        <span class="flex items-center gap-1 text-[10px] font-bold uppercase tracking-wider text-gold bg-gold/15 px-2 py-0.5 rounded border border-gold/30">
          <Star class="w-3 h-3" /> Chef
        </span>
      {/if}
    </div>
  </div>

  <!-- Fotografía con Sombra de Contacto -->
  <div class="relative py-4 flex items-center justify-center overflow-hidden">
    <div class="absolute inset-0 bg-gradient-to-t from-gold/5 to-transparent rounded-full filter blur-2xl opacity-0 group-hover:opacity-100 transition-opacity"></div>
    <img 
      src={product.image_url || '/imagenes/marca/isotipo.webp'} 
      alt={product.name}
      class="max-h-44 w-auto object-contain drop-shadow-2xl group-hover:scale-105 transition-transform duration-300"
      loading="lazy"
      decoding="async"
    />
  </div>

  <!-- Información del Producto -->
  <div class="mt-2">
    <div class="flex items-baseline justify-between gap-2 mb-1.5">
      <h3 class="font-serif text-2xl font-bold text-bone tracking-wide group-hover:text-gold transition-colors">
        {product.name}
      </h3>
      <span class="font-mono text-base font-bold text-gold">
        {formatPrice(product.price)}
      </span>
    </div>

    <p class="text-xs text-ash line-clamp-2 leading-relaxed font-sans mb-4">
      {product.description}
    </p>

    <!-- Botón de Selección Rápida -->
    <div class="pt-3 border-t hairline-dark flex items-center justify-between">
      <span class="text-[11px] font-mono text-ash/80 group-hover:text-bone transition-colors">
        Ver receta y acompañar →
      </span>
      <button 
        onclick={(e) => {
          e.stopPropagation();
          store.addToCart(product);
        }}
        class="w-8 h-8 rounded-full bg-white/5 hover:bg-ember text-bone flex items-center justify-center transition-colors border border-white/10"
        title="Agregar directamente"
      >
        <Plus class="w-4 h-4" />
      </button>
    </div>
  </div>
</div>
