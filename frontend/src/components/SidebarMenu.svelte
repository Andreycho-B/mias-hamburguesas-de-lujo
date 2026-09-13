<script>
  import { fly, fade } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import { store } from '../lib/store.svelte.js';
  import { 
    X, 
    ArrowRight, 
    Award, 
    Shield, 
    Clock, 
    Sparkles,
    ChevronRight,
    LogOut,
    Phone,
    ShoppingBag,
    MapPin,
    User
  } from 'lucide-svelte';

  // Validar horario en tiempo real: Sábado a Jueves de 3:00 PM a 9:30 PM (Viernes cerrado)
  let isOpenNow = $derived.by(() => {
    const now = new Date();
    if (now.getDay() === 5) return false; // Viernes cerrado
    const hour = now.getHours();
    const minute = now.getMinutes();
    const totalMinutes = hour * 60 + minute;
    const openMinutes = 15 * 60; // 3:00 PM
    const closeMinutes = 21 * 60 + 30; // 9:30 PM
    return totalMinutes >= openMinutes && totalMinutes <= closeMinutes;
  });

  const menuCategories = [
    {
      slug: 'hamburguesas',
      title: 'Hamburguesas de Autor',
      desc: 'Pan pretzel artesanal y 160g de res al carbón',
      icon: '🍔',
      badge: 'Favoritas'
    },
    {
      slug: 'perros',
      title: 'Perros Calientes',
      desc: 'Pan brioche artesanal y chorizo santarrosano',
      icon: '🌭',
      badge: null
    },
    {
      slug: 'de_la_casa',
      title: 'De la Casa',
      desc: 'Salchipapa tradicional, choripapa y desgranados',
      icon: '🌽',
      badge: null
    },
    {
      slug: 'adicionales',
      title: 'Adicionales & Papas',
      desc: 'Papas nativas al carbón, francesas y tocineta',
      icon: '🍟',
      badge: null
    },
    {
      slug: 'bebidas',
      title: 'Línea de Bebidas',
      desc: 'Refrescos en vidrio y agua mineral fría',
      icon: '🥤',
      badge: null
    }
  ];

  let isStaffUser = $derived(Boolean(store.user && ['waiter', 'cashier', 'admin'].includes(store.user.role)));
  let cartItemsCount = $derived(store.cart?.reduce((acc, i) => acc + (i.quantity || 1), 0) || 0);
  let activeOrdersCount = $derived(store.activeOrders?.filter(o => o.status !== 'DELIVERED' && o.status !== 'CANCELLED').length || 0);

  function navigateTo(viewSlug) {
    store.closeSidebar();
    store.currentView = viewSlug;
  }

  function openCart() {
    store.closeSidebar();
    store.isCartOpen = true;
  }

  function openAuth() {
    store.closeSidebar();
    store.isAuthModalOpen = true;
  }

  function handleLogout() {
    store.logout();
  }

  function handleKeydown(e) {
    if (e.key === 'Escape' && store.isSidebarOpen) {
      store.closeSidebar();
    }
  }
</script>

<svelte:window onkeydown={handleKeydown} />

{#if store.isSidebarOpen}
  <!-- MODAL EN PANTALLA COMPLETA ESTILO APPLE CARD -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div 
    data-lenis-prevent
    transition:fade={{ duration: 200 }}
    onclick={() => store.closeSidebar()}
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/65 backdrop-blur-xs p-3 sm:p-5 overflow-y-auto overscroll-contain select-none"
    style="touch-action: pan-y;"
  >
    <!-- TARJETA APPLE CARD CON CURVATURA 42px / 48px -->
    <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
    <div
      transition:fly={{ y: 24, duration: 280, easing: cubicOut }}
      onclick={(e) => e.stopPropagation()}
      class="bg-char text-bone w-full max-w-lg rounded-[38px] sm:rounded-[44px] border border-white/10 overflow-hidden flex flex-col max-h-[92vh] shadow-[0_25px_60px_-15px_rgba(0,0,0,0.8)] relative"
    >
      <!-- 1. CABECERA EDITORIAL DE LA CARD -->
      <header class="px-5 sm:px-6 pt-5 pb-3 border-b border-white/10 flex items-center justify-between shrink-0 bg-char">
        <button 
          type="button"
          onclick={() => navigateTo('menu')}
          class="flex items-center gap-3 text-left cursor-pointer active:scale-95 transition-transform min-h-[44px]"
        >
          <img 
            src="/imagenes/marca/logo_en_blanco.webp" 
            alt="Mías Hamburguesas de Lujo" 
            class="h-8 sm:h-9 w-auto object-contain"
          />
          <div class="border-l border-white/10 pl-2.5">
            <span class="text-[9px] font-mono font-bold tracking-widest text-ash uppercase block">
              Menú & Carta
            </span>
            <span class="text-xs font-serif font-bold text-bone italic">
              Sede Duitama
            </span>
          </div>
        </button>

        <!-- Botón de Cierre Circular Fat-Finger (44x44px) -->
        <button
          type="button"
          onclick={() => store.closeSidebar()}
          class="w-11 h-11 min-w-[44px] min-h-[44px] rounded-full bg-char-card border border-white/10 hover:bg-char-card-hover text-bone flex items-center justify-center transition-all active:scale-90 shadow-2xs cursor-pointer"
          aria-label="Cerrar menú"
        >
          <X class="w-5 h-5 stroke-[2.2]" />
        </button>
      </header>
      
      <!-- 2. CONTENIDO DESPLAZABLE DENTRO DE LA CARD -->
      <div 
        class="flex-1 overflow-y-auto px-5 sm:px-6 py-5 space-y-6"
        style="scrollbar-width: none; overscroll-behavior: contain; touch-action: pan-y;"
      >
        <!-- TARJETA DE PERFIL / FIDELIZACIÓN -->
        <div class="bg-char-card rounded-[28px] p-4.5 border border-white/10 shadow-2xs">
          {#if store.user}
            <div class="flex items-center justify-between gap-3">
              <button 
                type="button"
                onclick={openAuth}
                class="flex items-center gap-3 min-w-0 text-left cursor-pointer group flex-1 min-h-[48px]"
              >
                {#if store.user.avatar_url}
                  <img src={store.user.avatar_url} alt={store.user.full_name} class="w-12 h-12 rounded-full object-cover border border-white/10 shrink-0" />
                {:else}
                  <div class="w-12 h-12 rounded-full bg-char text-bone flex items-center justify-center font-display font-black text-base shrink-0 border border-white/10">
                    {store.user.full_name?.charAt(0).toUpperCase() || 'M'}
                  </div>
                {/if}
                <div class="min-w-0">
                  <div class="flex items-center gap-1.5">
                    <span class="text-[9px] font-mono uppercase text-ash tracking-wider font-bold block">Mi Cuenta</span>
                    {#if store.user.role && store.user.role !== 'customer'}
                      <span class="px-1.5 py-0.5 rounded-full bg-purple-950/60 text-purple-400 border border-purple-800/40 text-[9px] font-mono font-bold uppercase">
                        {store.user.role}
                      </span>
                    {/if}
                  </div>
                  <h3 class="font-display font-black text-sm sm:text-base text-bone truncate group-hover:text-[#FF2400] transition-colors">
                    {store.user.full_name}
                  </h3>
                  <span class="text-[11px] font-mono text-ash flex items-center gap-1">
                    <MapPin class="w-3 h-3 text-[#FF2400] shrink-0" /> Duitama
                  </span>
                </div>
              </button>

              <!-- Sellos y Salir -->
              <div class="flex items-center gap-2 shrink-0">
                <button
                  type="button"
                  onclick={() => navigateTo('stamps')}
                  class="min-h-[44px] flex items-center gap-1.5 px-3 py-1.5 rounded-full bg-amber-500/10 border border-amber-500/25 text-bone hover:border-amber-500/50 active:scale-95 transition-all cursor-pointer shadow-2xs"
                  title="Ver pasaporte de sellos"
                >
                  <Award class="w-4 h-4 text-amber-500 shrink-0" />
                  <span class="font-mono text-xs font-bold text-amber-400">{store.stampBalance}</span>
                </button>

                <button
                  type="button"
                  onclick={handleLogout}
                  class="w-11 h-11 min-w-[44px] min-h-[44px] rounded-full hover:bg-white/10 active:scale-95 text-ash hover:text-[#FF2400] flex items-center justify-center transition-colors cursor-pointer"
                  title="Cerrar sesión"
                  aria-label="Cerrar sesión"
                >
                  <LogOut class="w-4 h-4 stroke-[2]" />
                </button>
              </div>
            </div>
          {:else}
            <!-- Invitación a Registrarse en el Pasaporte de Sellos -->
            <div class="space-y-3.5 text-left">
              <div class="flex items-start gap-3">
                <div class="w-11 h-11 min-w-[44px] min-h-[44px] rounded-full bg-char text-[#FF2400] flex items-center justify-center shrink-0 mt-0.5 border border-white/10 shadow-2xs">
                  <Sparkles class="w-5 h-5 stroke-[2]" />
                </div>
                <div>
                  <h3 class="font-serif font-bold text-base text-bone leading-tight">
                    Pasaporte de Fidelización
                  </h3>
                  <p class="text-xs text-ash mt-0.5 leading-relaxed">
                    Acumula 10 sellos en Duitama y redime tu hamburguesa de autor favorita gratis.
                  </p>
                </div>
              </div>

              <button
                type="button"
                onclick={openAuth}
                class="w-full min-h-[48px] py-3 rounded-full bg-[#FF2400] hover:bg-[#E02000] active:scale-[0.98] text-white font-display font-extrabold text-xs sm:text-sm tracking-tight flex items-center justify-center gap-2 transition-all shadow-sm cursor-pointer"
              >
                <span>Iniciar Sesión / Pasaporte</span>
                <ArrowRight class="w-4 h-4" />
              </button>
            </div>
          {/if}
        </div>

        <!-- SECCIÓN 1: CARTA GASTRONÓMICA COMPLETA -->
        <div class="space-y-2.5 text-left">
          <div class="flex items-center justify-between px-1">
            <span class="text-[10px] font-mono font-bold tracking-widest text-ash uppercase">
              Categorías de Carta
            </span>
            <button
              type="button"
              onclick={() => navigateTo('menu')}
              class="min-h-[36px] flex items-center text-xs font-mono font-bold text-[#FF2400] hover:underline cursor-pointer active:scale-95"
            >
              Ver Carta Completa →
            </button>
          </div>

          <div class="space-y-2">
            {#each menuCategories as cat}
              <button
                type="button"
                onclick={() => navigateTo(cat.slug)}
                class="w-full min-h-[58px] bg-char-card hover:bg-char-card-hover border rounded-[24px] p-3.5 text-left transition-all active:scale-[0.985] cursor-pointer flex items-center justify-between group shadow-2xs {store.currentView === cat.slug ? 'border-[#FF2400] bg-[#FF2400]/10 text-[#FF2400]' : 'border-white/10 hover:border-white/20'}"
              >
                <div class="flex items-center gap-3.5 min-w-0">
                  <span class="text-2xl shrink-0 group-hover:scale-110 transition-transform">
                    {cat.icon}
                  </span>
                  <div class="min-w-0">
                    <div class="flex items-center gap-2">
                      <h4 class="font-display font-bold text-sm text-bone group-hover:text-[#FF2400] transition-colors truncate">
                        {cat.title}
                      </h4>
                      {#if cat.badge}
                        <span class="text-[9px] font-mono font-bold uppercase text-[#FF2400] bg-[#FF2400]/10 px-2 py-0.5 rounded-full shrink-0">
                          {cat.badge}
                        </span>
                      {/if}
                    </div>
                    <p class="text-xs text-ash truncate mt-0.5">
                      {cat.desc}
                    </p>
                  </div>
                </div>

                <ChevronRight class="w-5 h-5 text-ash group-hover:text-[#FF2400] group-hover:translate-x-0.5 transition-all shrink-0 ml-2" />
              </button>
            {/each}
          </div>
        </div>

        <!-- SECCIÓN 2: ACCIONES DIRECTAS DEL CLIENTE -->
        <div class="space-y-2.5 text-left pt-1">
          <span class="text-[10px] font-mono font-bold tracking-widest text-ash uppercase px-1 block">
            Acciones & Fidelización
          </span>

          <div class="grid grid-cols-1 gap-2">
            <!-- 1. Pasaporte de Sellos -->
            <button
              type="button"
              onclick={() => navigateTo('stamps')}
              class="w-full min-h-[56px] bg-char-card hover:bg-char-card-hover border border-white/10 hover:border-amber-500/40 rounded-[24px] p-3.5 text-left transition-all active:scale-[0.985] cursor-pointer flex items-center justify-between group shadow-2xs {store.currentView === 'stamps' ? 'border-amber-500 bg-amber-500/10' : ''}"
            >
              <div class="flex items-center gap-3 min-w-0">
                <div class="w-10 h-10 rounded-2xl bg-amber-500/10 text-amber-500 flex items-center justify-center shrink-0">
                  <Award class="w-5 h-5 stroke-[2.2]" />
                </div>
                <div class="min-w-0">
                  <h4 class="font-display font-bold text-sm text-bone group-hover:text-amber-500 transition-colors">
                    Pasaporte de Sellos
                  </h4>
                  <p class="text-xs text-ash truncate">
                    10 sellos = 1 hamburguesa gratis
                  </p>
                </div>
              </div>
              <div class="flex items-center gap-1.5 shrink-0">
                <span class="text-xs font-mono font-bold text-amber-400 bg-amber-500/20 px-2 py-0.5 rounded-full">
                  {store.stampBalance}
                </span>
                <ChevronRight class="w-4 h-4 text-ash group-hover:text-bone transition-colors" />
              </div>
            </button>

            <!-- 2. Mis Pedidos & Carrito Activo -->
            <button
              type="button"
              onclick={openCart}
              class="w-full min-h-[56px] bg-char-card hover:bg-char-card-hover border border-white/10 hover:border-[#FF2400]/40 rounded-[24px] p-3.5 text-left transition-all active:scale-[0.985] cursor-pointer flex items-center justify-between group shadow-2xs"
            >
              <div class="flex items-center gap-3 min-w-0">
                <div class="w-10 h-10 rounded-2xl bg-[#FF2400]/10 text-[#FF2400] flex items-center justify-center shrink-0">
                  <ShoppingBag class="w-5 h-5 stroke-[2]" />
                </div>
                <div class="min-w-0">
                  <h4 class="font-display font-bold text-sm text-bone group-hover:text-[#FF2400] transition-colors">
                    Mi Pedido & Domicilio
                  </h4>
                  <p class="text-xs text-ash truncate">
                    {cartItemsCount > 0 ? `${cartItemsCount} productos en tu bolsa` : 'Revisa tus órdenes y comanda activa'}
                  </p>
                </div>
              </div>
              <div class="flex items-center gap-1.5 shrink-0">
                {#if cartItemsCount > 0 || activeOrdersCount > 0}
                  <span class="text-xs font-mono font-bold text-white bg-[#FF2400] px-2.5 py-0.5 rounded-full shadow-2xs">
                    {activeOrdersCount > 0 ? `${activeOrdersCount} en curso` : `${cartItemsCount}`}
                  </span>
                {/if}
                <ChevronRight class="w-4 h-4 text-ash group-hover:text-bone transition-colors" />
              </div>
            </button>

            <!-- 3. Terminal de Personal (SOLO SI ES PERSONAL STAFF) -->
            {#if isStaffUser}
              <button
                type="button"
                onclick={() => navigateTo('staff')}
                class="w-full min-h-[56px] bg-char-card hover:bg-char-card-hover border border-white/10 hover:border-purple-500/40 rounded-[24px] p-3.5 text-left transition-all active:scale-[0.985] cursor-pointer flex items-center justify-between group shadow-2xs"
              >
                <div class="flex items-center gap-3 min-w-0">
                  <div class="w-10 h-10 rounded-2xl bg-purple-500/10 text-purple-400 flex items-center justify-center shrink-0">
                    <Shield class="w-5 h-5 stroke-[2]" />
                  </div>
                  <div class="min-w-0">
                    <h4 class="font-display font-bold text-sm text-bone group-hover:text-purple-400 transition-colors">
                      Terminal Operativa
                    </h4>
                    <p class="text-xs text-ash truncate">
                      Comandas y caja para personal
                    </p>
                  </div>
                </div>
                <div class="flex items-center gap-1.5 shrink-0">
                  <span class="text-[10px] font-mono font-bold uppercase text-purple-300 bg-purple-900/40 px-2 py-0.5 rounded-full border border-purple-500/20">
                    {store.user.role}
                  </span>
                  <ChevronRight class="w-4 h-4 text-ash group-hover:text-bone transition-colors" />
                </div>
              </button>
            {/if}
          </div>
        </div>

        <!-- SECCIÓN 3: ATENCIÓN & HORARIO DUITAMA -->
        <div class="bg-char-card rounded-[28px] p-4.5 border border-white/10 space-y-3 text-left shadow-2xs">
          <div class="flex items-center justify-between">
            <span class="text-[10px] font-mono font-bold tracking-wider text-ash uppercase flex items-center gap-1.5">
              <Clock class="w-4 h-4 text-bone" />
              Horario Duitama
            </span>
            <span class="flex items-center gap-1.5 text-xs font-mono font-bold {isOpenNow ? 'text-emerald-400' : 'text-[#FF2400]'}">
              <span class="w-2 h-2 rounded-full {isOpenNow ? 'bg-emerald-500 animate-pulse' : 'bg-[#FF2400]'}"></span>
              {isOpenNow ? 'Abierto Ahora' : 'Cerrado Ahora'}
            </span>
          </div>
          <p class="text-xs font-mono text-bone">
            Sáb a Jue: <strong>3:00 PM – 9:30 PM</strong> (Vie. cerrado)
          </p>

          <!-- WhatsApp de Atención Oficial Duitama -->
          <a
            href="https://api.whatsapp.com/send?phone=573212520881&text=Hola%20M%C3%ADas%20Duitama%2C%20quisiera%20hacer%20un%20pedido%20o%20consultar%20la%20carta."
            target="_blank"
            rel="noopener noreferrer"
            class="w-full min-h-[48px] py-2.5 px-4 rounded-2xl bg-emerald-600/15 hover:bg-emerald-600/25 active:scale-[0.98] text-emerald-300 font-mono text-xs font-bold flex items-center justify-between transition-colors cursor-pointer border border-emerald-500/20"
          >
            <div class="flex items-center gap-2.5">
              <Phone class="w-4 h-4 text-emerald-400 shrink-0" />
              <span>WhatsApp: 321 252 0881</span>
            </div>
            <ArrowRight class="w-4 h-4 text-emerald-400 shrink-0" />
          </a>
        </div>
      </div>

      <!-- FOOTER DEL MENÚ LATERAL -->
      <footer class="p-4 sm:p-5 bg-char border-t border-white/10 shrink-0 text-center select-none">
        <div class="flex items-center justify-center gap-2 mb-1">
          <img src="/imagenes/marca/isotipo.webp" alt="Mías" class="w-5 h-5 object-contain" />
          <span class="font-serif font-bold text-sm text-bone">Mías Hamburguesas de Lujo</span>
        </div>
        <p class="text-[10px] font-mono text-ash">
          Duitama, Boyacá • Lo Íntimo Hecho para Compartir
        </p>
      </footer>
    </div>
  </div>
{/if}
