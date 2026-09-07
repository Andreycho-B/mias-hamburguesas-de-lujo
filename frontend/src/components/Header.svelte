<script>
  import { store } from '../lib/store.svelte.js';
  import { User, Award, BookOpen, Utensils, Shield } from 'lucide-svelte';

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

<header class="sticky top-0 z-40 bg-char/95 backdrop-blur-md border-b hairline-dark px-4 py-3 transition-all">
  <div class="max-w-6xl mx-auto flex items-center justify-between">
    <!-- Logotipo Oficial -->
    <button 
      onclick={() => store.currentView = 'menu'} 
      class="flex items-center gap-3 text-left group focus:outline-none"
    >
      <img 
        src="/imagenes/marca/logo_en_blanco.webp" 
        alt="Mías Hamburguesas de Lujo" 
        class="h-8 sm:h-9 object-contain group-hover:opacity-90 transition-opacity"
      />
    </button>

    <!-- Indicador de Horario en Vivo (Sáb a Jue 3:00 PM - 9:30 PM, Vie cerrado) -->
    <div class="hidden md:flex items-center gap-2 px-3 py-1 rounded-full bg-white/5 border hairline-dark text-xs font-mono">
      <span class="w-2 h-2 rounded-full {isOpen ? 'bg-emerald-500 animate-pulse' : 'bg-ember'}"></span>
      {#if isOpen}
        <span class="text-bone/80">Abierto hoy hasta 9:30 PM</span>
      {:else}
        <span class="text-ash">Cerrado ahora · Abrimos 3:00 PM (Vie. cerrado)</span>
      {/if}
    </div>

    <!-- Navegación Rápida -->
    <div class="flex items-center gap-2">
      <!-- Menú -->
      <button 
        onclick={() => store.currentView = 'menu'}
        class="p-2 rounded-lg text-xs font-medium uppercase tracking-wider transition-colors {store.currentView === 'menu' ? 'text-gold bg-white/5' : 'text-ash hover:text-bone'}"
        title="Carta Gastronómica"
      >
        <span class="hidden sm:inline">Menú</span>
        <Utensils class="w-4 h-4 sm:hidden inline" />
      </button>

      <!-- Mis Sellos -->
      <button 
        onclick={() => store.currentView = 'stamps'}
        class="relative p-2 rounded-lg text-xs font-medium uppercase tracking-wider transition-colors {store.currentView === 'stamps' ? 'text-ember bg-white/5' : 'text-ash hover:text-bone'}"
        title="Mis Sellos"
      >
        <div class="flex items-center gap-1.5">
          <Award class="w-4 h-4 text-ember" />
          <span class="hidden sm:inline">Sellos</span>
          {#if store.stampBalance > 0}
            <span class="bg-ember text-white text-[10px] font-bold px-1.5 py-0.2 rounded-full">
              {store.stampBalance}/10
            </span>
          {/if}
        </div>
      </button>

      <!-- Historia -->
      <button 
        onclick={() => store.currentView = 'history'}
        class="p-2 rounded-lg text-xs font-medium uppercase tracking-wider transition-colors {store.currentView === 'history' ? 'text-gold bg-white/5' : 'text-ash hover:text-bone'}"
        title="Nuestra Historia"
      >
        <span class="hidden sm:inline">Historia</span>
        <BookOpen class="w-4 h-4 sm:hidden inline" />
      </button>

      <!-- Acceso de Personal o Perfil -->
      {#if store.user}
        <div class="flex items-center gap-2 ml-2 pl-2 border-l hairline-dark">
          <span class="text-xs text-ash font-mono hidden lg:inline">{store.user.full_name}</span>
          {#if store.user.role === 'admin' || store.user.role === 'cashier' || store.user.role === 'waiter'}
            <button 
              onclick={() => store.currentView = 'staff'}
              class="px-2.5 py-1 rounded bg-amber-500/20 text-gold border border-gold/30 text-xs font-bold uppercase tracking-wider"
            >
              Panel {store.user.role}
            </button>
          {/if}
          <button 
            onclick={() => store.logout()}
            class="text-xs text-ash hover:text-ember underline ml-1"
          >
            Salir
          </button>
        </div>
      {:else}
        <button 
          onclick={() => store.isAuthModalOpen = true}
          class="flex items-center gap-1.5 px-3 py-1.5 rounded-md border border-gold/40 text-gold hover:bg-gold/10 text-xs font-medium tracking-wide transition-all ml-1"
        >
          <User class="w-3.5 h-3.5" />
          <span>Ingresar</span>
        </button>
      {/if}
    </div>
  </div>
</header>
