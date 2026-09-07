<script>
  import { store } from '../lib/store.svelte.js';
  import { User, Award, BookOpen, Utensils, Shield, ChevronLeft } from 'lucide-svelte';

  // Validar horario en tiempo real: Sábado a Jueves de 3:00 PM a 9:30 PM (Viernes cerrado)
  let isOpen = $derived.by(() => {
    const now = new Date();
    if (now.getDay() === 5) return false; // Viernes cerrado
    const hour = now.getHours();
    const minute = now.getMinutes();
    const totalMinutes = hour * 60 + minute;

    const openMinutes = 15 * 60; // 3:00 PM
    const closeMinutes = 21 * 60 + 30; // 9:30 PM

    return totalMinutes >= openMinutes && totalMinutes <= closeMinutes;
  });
</script>

<header class="sticky top-0 z-40 bg-[#FAF6EE]/95 backdrop-blur-md border-b border-black/10 px-4 py-3 transition-all">
  <div class="max-w-6xl mx-auto flex items-center justify-between">
    <!-- Logotipo Oficial en Negro sobre Crema -->
    <button 
      type="button"
      onclick={() => store.currentView = 'menu'} 
      class="flex items-center gap-3 text-left group focus:outline-none cursor-pointer"
    >
      <img 
        src="/imagenes/marca/logo_en_negro.webp" 
        alt="Mías Hamburguesas de Lujo" 
        class="h-8 sm:h-9 object-contain group-hover:opacity-80 transition-opacity"
      />
    </button>

    <!-- Indicador de Horario en Vivo -->
    <div class="hidden md:flex items-center gap-2 px-3.5 py-1.5 rounded-full bg-[#EDE8DC] border border-black/5 text-xs font-mono">
      <span class="w-2 h-2 rounded-full {isOpen ? 'bg-emerald-500 animate-pulse' : 'bg-[#FF2400]'}"></span>
      {#if isOpen}
        <span class="text-[#0C0C0E] font-medium">Abierto hoy hasta 9:30 PM</span>
      {:else}
        <span class="text-[#8E8E93]">Cerrado ahora · Abrimos 3:00 PM (Vie. cerrado)</span>
      {/if}
    </div>

    <!-- Navegación Rápida -->
    <div class="flex items-center gap-2">
      <!-- Menú -->
      <button 
        type="button"
        onclick={() => store.currentView = 'menu'}
        class="px-3 py-2 rounded-full text-xs font-mono font-bold uppercase tracking-wider transition-all cursor-pointer {store.currentView === 'menu' ? 'text-[#FF2400] bg-white border border-black/10' : 'text-[#3A3A3C] hover:text-[#0C0C0E]'}"
        title="Carta Gastronómica"
      >
        <span class="hidden sm:inline">Menú</span>
        <Utensils class="w-4 h-4 sm:hidden inline" />
      </button>

      <!-- Mis Sellos -->
      <button 
        type="button"
        onclick={() => store.currentView = 'stamps'}
        class="relative px-3 py-2 rounded-full text-xs font-mono font-bold uppercase tracking-wider transition-all cursor-pointer {store.currentView === 'stamps' ? 'text-white bg-[#FF2400]' : 'text-[#3A3A3C] hover:text-[#0C0C0E]'}"
        title="Mis Sellos"
      >
        <div class="flex items-center gap-1.5">
          <Award class="w-4 h-4" />
          <span class="hidden sm:inline">Sellos</span>
          {#if store.stampBalance > 0}
            <span class="{store.currentView === 'stamps' ? 'bg-white text-[#FF2400]' : 'bg-[#FF2400] text-white'} text-[10px] font-black px-1.5 py-0.2 rounded-full">
              {store.stampBalance}/10
            </span>
          {/if}
        </div>
      </button>

      <!-- Acceso de Personal o Perfil -->
      {#if store.user}
        <div class="flex items-center gap-2 ml-1 pl-2 border-l border-black/10">
          <button 
            type="button"
            onclick={() => store.isAuthModalOpen = true}
            class="text-xs text-[#0C0C0E] hover:text-[#FF2400] font-mono font-bold transition-colors hidden sm:inline cursor-pointer"
          >
            {store.user.full_name}
          </button>

          {#if store.user.role === 'admin' || store.user.role === 'cashier' || store.user.role === 'waiter'}
            <button 
              type="button"
              onclick={() => store.currentView = 'staff'}
              class="px-3.5 py-1.5 rounded-full bg-[#0C0C0E] text-white text-xs font-mono font-bold uppercase tracking-wider hover:bg-black transition-colors cursor-pointer"
            >
              Panel Staff
            </button>
          {/if}

          <button 
            type="button"
            onclick={() => store.logout()}
            class="text-xs text-[#8E8E93] hover:text-[#FF2400] font-mono ml-1 cursor-pointer"
          >
            Salir
          </button>
        </div>
      {:else}
        <button 
          type="button"
          onclick={() => store.isAuthModalOpen = true}
          class="flex items-center gap-1.5 px-4 py-2 rounded-full bg-white border border-black/10 text-[#0C0C0E] hover:border-[#FF2400] text-xs font-mono font-bold tracking-wide transition-all cursor-pointer ml-1"
        >
          <User class="w-3.5 h-3.5" />
          <span>Ingresar</span>
        </button>
      {/if}
    </div>
  </div>
</header>
