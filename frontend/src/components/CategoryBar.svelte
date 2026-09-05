<script>
  import { store } from '../lib/store.svelte.js';

  let { selectedCategory, onSelect } = $props();

  const categories = [
    { id: 1, slug: 'hamburguesas', label: 'HAMBURGUESAS', index: '01' },
    { id: 2, slug: 'perros', label: 'PERROS', index: '02' },
    { id: 3, slug: 'de_la_casa', label: 'DE LA CASA', index: '03' },
    { id: 6, slug: 'temporada', label: 'TEMPORADA', index: '04', badge: 'Amor & Amistad' },
    { id: 4, slug: 'bebidas', label: 'BEBIDAS', index: '05' }
  ];
</script>

<div class="border-b hairline-dark py-4 px-4 bg-char/80 sticky top-[57px] z-30 backdrop-blur-md">
  <div class="max-w-6xl mx-auto flex items-center gap-3 overflow-x-auto no-scrollbar py-1">
    {#each categories as cat}
      <button
        onclick={() => onSelect(cat.slug)}
        class="flex-shrink-0 flex items-center gap-2 px-4 py-2 rounded-full border transition-all duration-200 text-xs font-mono tracking-wider {selectedCategory === cat.slug ? 'border-ember bg-ember/15 text-bone font-semibold shadow-ember' : 'border-white/10 text-ash hover:text-bone hover:border-white/20'}"
      >
        <span class="{selectedCategory === cat.slug ? 'text-ember' : 'text-gold/70'} font-mono">[{cat.index}]</span>
        <span>{cat.label}</span>
        {#if cat.badge}
          <span class="bg-ember text-white text-[9px] px-1.5 py-0.5 rounded font-sans font-bold uppercase tracking-wider ml-1">
            {cat.badge}
          </span>
        {/if}
      </button>
    {/each}
  </div>
</div>
