<script>
  import { store } from '../lib/store.svelte.js';
  import { User, Award, BookOpen, Utensils, Shield, ChevronLeft, Receipt } from 'lucide-svelte';

  const activeStatuses = ['PENDING', 'CONFIRMED', 'PREPARING', 'IN_DELIVERY', 'READY'];
  let activeOrdersCount = $derived(
    store.activeOrders?.filter((o) => activeStatuses.includes(o.status)).length || 0
  );

  // Validar horario en tiempo real: Lunes a Viernes 3:00 PM a 9:00 PM, Sábados y Domingos 3:00 PM a 10:00 PM
  let isWeekend = $derived.by(() => {
    const day = new Date().getDay();
    return day === 0 || day === 6; // 0 = Domingo, 6 = Sábado
  });

  let isOpen = $derived.by(() => {
    const now = new Date();
    const day = now.getDay();
    const weekend = day === 0 || day === 6;
    const hour = now.getHours();
    const minute = now.getMinutes();
    const totalMinutes = hour * 60 + minute;

    const openMinutes = 15 * 60; // 3:00 PM
    const closeMinutes = weekend ? 22 * 60 : 21 * 60; // 10:00 PM en fin de semana, 9:00 PM entre semana

    return totalMinutes >= openMinutes && totalMinutes <= closeMinutes;
  });
</script>

<header class="sticky top-0 z-40 bg-char/95 backdrop-blur-md border-b border-white/10 px-4 py-3 transition-all">
  <div class="max-w-6xl mx-auto flex items-center justify-between">
    <!-- Logotipo Oficial en Blanco sobre Carbón -->
    <button 
      type="button"
      onclick={() => store.currentView = 'menu'} 
      class="flex items-center gap-3 cursor-pointer group"
    >
      <img 
        src="/imagenes/marca/logo_en_blanco.webp" 
        alt="Mías Hamburguesas de Lujo" 
        class="h-9 w-auto object-contain transition-transform group-hover:scale-105"
      />
      <div class="hidden md:flex flex-col border-l border-white/10 pl-3">
        <span class="font-serif text-sm font-bold text-bone tracking-wide">Mías Hamburguesas de Lujo</span>
        <span class="font-mono text-[10px] text-ash uppercase tracking-wider">Alta Gastronomía</span>
      </div>
    </button>

    <!-- Estado del Local en Tiempo Real -->
    <div class="hidden sm:flex items-center gap-2 px-3 py-1.5 rounded-full bg-char-card border border-white/10 text-xs font-mono">
      <span class="w-2 h-2 rounded-full {isOpen ? 'bg-emerald-500 animate-pulse' : 'bg-red-500'}"></span>
      {#if isOpen}
        <span class="text-bone/90">Abierto hoy hasta {isWeekend ? '10:00 PM' : '9:00 PM'}</span>
      {:else}
        <span class="text-[#8E8E93]">Cerrado · Servicio de 3:00 PM a {isWeekend ? '10:00 PM' : '9:00 PM'}</span>
      {/if}
    </div>

    <!-- Navegación Rápida -->
    <div class="flex items-center gap-2">
      <!-- Menú -->
      <button 
        type="button"
        onclick={() => store.currentView = 'menu'}
        class="px-3 py-2 rounded-full text-xs font-mono font-bold uppercase tracking-wider transition-all cursor-pointer {store.currentView === 'menu' ? 'text-white bg-[#FF2400]' : 'text-bone/80 hover:text-bone'}"
        title="Carta Gastronómica"
      >
        <span class="hidden sm:inline">Menú</span>
        <Utensils class="w-4 h-4 sm:hidden inline" />
      </button>

      <!-- Mis Pedidos -->
      <button 
        type="button"
        onclick={() => store.currentView = 'orders'}
        class="relative px-3 py-2 rounded-full text-xs font-mono font-bold uppercase tracking-wider transition-all cursor-pointer {store.currentView === 'orders' ? 'text-white bg-[#FF2400]' : 'text-bone/80 hover:text-bone'}"
        title="Mis Pedidos y Comandas"
      >
        <div class="flex items-center gap-1.5">
          <Receipt class="w-4 h-4" />
          <span class="hidden sm:inline">Pedidos</span>
          {#if activeOrdersCount > 0}
            <span class="{store.currentView === 'orders' ? 'bg-white text-[#FF2400]' : 'bg-[#FF2400] text-white'} text-[10px] font-black px-1.5 py-0.2 rounded-full animate-pulse">
              {activeOrdersCount}
            </span>
          {/if}
        </div>
      </button>

      <!-- Mis Sellos -->
      <button 
        type="button"
        onclick={() => store.currentView = 'stamps'}
        class="relative px-3 py-2 rounded-full text-xs font-mono font-bold uppercase tracking-wider transition-all cursor-pointer {store.currentView === 'stamps' ? 'text-white bg-[#FF2400]' : 'text-bone/80 hover:text-bone'}"
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
        <div class="flex items-center gap-2 ml-1 pl-2 border-l border-white/10">
          <button 
            type="button"
            onclick={() => store.isAuthModalOpen = true}
            class="text-xs text-bone hover:text-[#FF2400] font-mono font-bold transition-colors hidden sm:inline cursor-pointer"
          >
            {store.user.full_name}
          </button>

          {#if store.user.role === 'admin' || store.user.role === 'cashier' || store.user.role === 'waiter'}
            <button 
              type="button"
              onclick={() => store.currentView = 'staff'}
              class="px-3.5 py-1.5 rounded-full bg-char-card border border-white/10 text-white text-xs font-mono font-bold uppercase tracking-wider hover:bg-white/10 transition-colors cursor-pointer"
            >
              Panel Staff
            </button>
          {/if}

          <button 
            type="button"
            onclick={() => store.logout()}
            class="text-xs text-ash hover:text-[#FF2400] font-mono ml-1 cursor-pointer"
          >
            Salir
          </button>
        </div>
      {:else}
        <button 
          type="button"
          onclick={() => store.isAuthModalOpen = true}
          class="flex items-center gap-1.5 px-4 py-2 rounded-full bg-char-card border border-white/10 text-bone hover:border-[#FF2400] text-xs font-mono font-bold tracking-wide transition-all cursor-pointer ml-1"
        >
          <User class="w-3.5 h-3.5" />
          <span>Ingresar</span>
        </button>
      {/if}
    </div>
  </div>
</header>
