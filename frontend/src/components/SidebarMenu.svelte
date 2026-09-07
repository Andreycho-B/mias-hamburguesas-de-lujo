<script>
  import { fly, fade } from 'svelte/transition';
  import { cubicOut } from 'svelte/easing';
  import { store } from '../lib/store.svelte.js';
  import { 
    X, 
    ArrowRight, 
    Award, 
    BookOpen, 
    User, 
    Utensils, 
    Shield, 
    Bike, 
    Clock, 
    Sparkles,
    ChevronRight,
    LogOut,
    Phone
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

  function navigateTo(viewSlug) {
    store.closeSidebar();
    store.currentView = viewSlug;
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
  <!-- BACKDROP OSCURO CON DESENFOQUE CINEMATOGRÁFICO -->
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div 
    transition:fade={{ duration: 220 }}
    onclick={() => store.closeSidebar()}
    class="fixed inset-0 z-50 bg-black/50 backdrop-blur-xs select-none"
  >
    <!-- CAJÓN LATERAL (DRAWER) CON CURVATURA APPLE EN EL BORDE DERECHO -->
    <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
    <aside
      transition:fly={{ x: -380, duration: 320, easing: cubicOut }}
      onclick={(e) => e.stopPropagation()}
      class="w-[86vw] max-w-[360px] sm:max-w-[400px] h-[100dvh] bg-[#FAF6EE] rounded-r-[38px] sm:rounded-r-[44px] border-r border-black/[0.08] shadow-[0_0_60px_rgba(0,0,0,0.35)] flex flex-col justify-between overflow-hidden relative"
      style="overscroll-behavior: contain;"
      onwheel={(e) => e.stopPropagation()}
      ontouchmove={(e) => e.stopPropagation()}
    >
      
      <!-- CONTENIDO DESPLAZABLE -->
      <div 
        class="flex-1 overflow-y-auto px-5 sm:px-6 pt-5 pb-6 space-y-6"
        style="scrollbar-width: none; overscroll-behavior: contain; touch-action: pan-y;"
      >
        <!-- CABECERA: LOGO + BOTÓN CERRAR -->
        <div class="flex items-center justify-between pb-2 border-b border-black/[0.06]">
          <button 
            type="button"
            onclick={() => navigateTo('menu')}
            class="flex items-center gap-2.5 text-left cursor-pointer active:scale-95 transition-transform"
          >
            <img 
              src="/imagenes/marca/logo_en_negro.webp" 
              alt="Mías Hamburguesas de Lujo" 
              class="h-8 sm:h-9 w-auto object-contain"
            />
          </button>

          <button
            type="button"
            onclick={() => store.closeSidebar()}
            class="w-10 h-10 rounded-full bg-white border border-black/10 hover:border-black/20 text-[#0C0C0E] flex items-center justify-center transition-all active:scale-95 shadow-2xs cursor-pointer"
            aria-label="Cerrar menú lateral"
          >
            <X class="w-5 h-5 stroke-[2]" />
          </button>
        </div>

        <!-- TARJETA DE PERFIL / FIDELIZACIÓN -->
        <div class="bg-white rounded-[26px] p-4 border border-black/[0.06] shadow-2xs">
          {#if store.user}
            <div class="flex items-center justify-between gap-3">
              <div class="flex items-center gap-3 min-w-0">
                <div class="w-11 h-11 rounded-full bg-[#18181b] text-white flex items-center justify-center font-display font-black text-sm shrink-0 border border-black/10">
                  {store.user.full_name?.charAt(0).toUpperCase() || 'M'}
                </div>
                <div class="min-w-0">
                  <span class="text-[10px] font-mono uppercase text-ash tracking-wider block">Comensal Mías</span>
                  <h3 class="font-display font-extrabold text-sm sm:text-base text-[#0C0C0E] truncate">
                    {store.user.full_name}
                  </h3>
                </div>
              </div>

              <!-- Sellos y Desconexión -->
              <div class="flex items-center gap-2 shrink-0">
                <button
                  type="button"
                  onclick={() => navigateTo('stamps')}
                  class="flex items-center gap-1.5 px-2.5 py-1 rounded-full bg-amber-500/10 border border-amber-500/20 text-[#0C0C0E] hover:border-amber-500/40 transition-all cursor-pointer"
                  title="Ver pasaporte de sellos"
                >
                  <Award class="w-3.5 h-3.5 text-amber-600" />
                  <span class="font-mono text-xs font-bold text-amber-700">{store.stampBalance}</span>
                </button>

                <button
                  type="button"
                  onclick={handleLogout}
                  class="w-8 h-8 rounded-full hover:bg-black/5 text-ash hover:text-[#FF2400] flex items-center justify-center transition-colors cursor-pointer"
                  title="Cerrar sesión"
                >
                  <LogOut class="w-4 h-4 stroke-[1.8]" />
                </button>
              </div>
            </div>
          {:else}
            <!-- Invitación a registrarse para fidelización -->
            <div class="space-y-3 text-left">
              <div class="flex items-start gap-3">
                <div class="w-10 h-10 rounded-full bg-[#FAF6EE] text-[#FF2400] flex items-center justify-center shrink-0 mt-0.5 border border-black/[0.04]">
                  <Sparkles class="w-5 h-5 stroke-[2]" />
                </div>
                <div>
                  <h3 class="font-display font-extrabold text-sm text-[#0C0C0E] leading-snug">
                    Pasaporte de Fidelización
                  </h3>
                  <p class="text-xs text-ash mt-0.5 leading-relaxed">
                    Acumula sellos en cada visita o domicilio y redime hamburguesas gratis.
                  </p>
                </div>
              </div>

              <button
                type="button"
                onclick={openAuth}
                class="w-full py-2.5 rounded-full bg-[#FF2400] hover:bg-[#E02000] active:scale-95 text-white font-display font-extrabold text-xs sm:text-sm tracking-tight flex items-center justify-center gap-1.5 transition-all shadow-2xs cursor-pointer"
              >
                <span>Iniciar Sesión / Registro</span>
                <ArrowRight class="w-3.5 h-3.5" />
              </button>
            </div>
          {/if}
        </div>

        <!-- SECCIÓN 1: CARTA GASTRONÓMICA -->
        <div class="space-y-2.5 text-left">
          <div class="flex items-center justify-between px-1">
            <span class="text-[10px] font-mono font-bold tracking-widest text-ash uppercase">
              Carta Gastronómica
            </span>
            <button
              type="button"
              onclick={() => navigateTo('menu')}
              class="text-[11px] font-mono font-bold text-[#FF2400] hover:underline cursor-pointer"
            >
              Ver Carta →
            </button>
          </div>

          <div class="space-y-1.5">
            {#each menuCategories as cat}
              <button
                type="button"
                onclick={() => navigateTo(cat.slug)}
                class="w-full bg-white hover:bg-[#FFFDF5] border border-black/[0.06] hover:border-[#FF2400]/40 rounded-[20px] p-3 text-left transition-all active:scale-[0.985] cursor-pointer flex items-center justify-between group shadow-2xs"
              >
                <div class="flex items-center gap-3 min-w-0">
                  <span class="text-xl sm:text-2xl shrink-0 group-hover:scale-110 transition-transform">
                    {cat.icon}
                  </span>
                  <div class="min-w-0">
                    <div class="flex items-center gap-1.5">
                      <h4 class="font-display font-bold text-xs sm:text-sm text-[#0C0C0E] group-hover:text-[#FF2400] transition-colors truncate">
                        {cat.title}
                      </h4>
                      {#if cat.badge}
                        <span class="text-[9px] font-mono font-bold uppercase text-[#FF2400] bg-[#FF2400]/10 px-1.5 py-0.2 rounded-full shrink-0">
                          {cat.badge}
                        </span>
                      {/if}
                    </div>
                    <p class="text-[11px] text-ash truncate mt-0.5">
                      {cat.desc}
                    </p>
                  </div>
                </div>

                <ChevronRight class="w-4 h-4 text-ash/70 group-hover:text-[#FF2400] group-hover:translate-x-0.5 transition-all shrink-0 ml-2" />
              </button>
            {/each}
          </div>
        </div>

        <!-- SECCIÓN 2: EXPERIENCIAS & HISTORIA -->
        <div class="space-y-2 text-left pt-1">
          <span class="text-[10px] font-mono font-bold tracking-widest text-ash uppercase px-1 block">
            Experiencias & Tradición
          </span>

          <div class="grid grid-cols-1 gap-1.5">
            <!-- Sellos -->
            <button
              type="button"
              onclick={() => navigateTo('stamps')}
              class="w-full bg-white hover:bg-[#FFFDF5] border border-black/[0.06] rounded-[20px] p-3 text-left transition-all active:scale-[0.985] cursor-pointer flex items-center justify-between group shadow-2xs"
            >
              <div class="flex items-center gap-3">
                <div class="w-8 h-8 rounded-xl bg-amber-500/10 text-amber-700 flex items-center justify-center shrink-0">
                  <Award class="w-4 h-4 stroke-[2.2]" />
                </div>
                <div>
                  <h4 class="font-display font-bold text-xs sm:text-sm text-[#0C0C0E] group-hover:text-amber-700 transition-colors">
                    Pasaporte de Sellos
                  </h4>
                  <p class="text-[11px] text-ash">
                    10 sellos = 1 hamburguesa de autor gratis
                  </p>
                </div>
              </div>
              <ChevronRight class="w-4 h-4 text-ash/70 group-hover:text-[#0C0C0E] transition-colors" />
            </button>

            <!-- Historia -->
            <button
              type="button"
              onclick={() => navigateTo('history')}
              class="w-full bg-white hover:bg-[#FFFDF5] border border-black/[0.06] rounded-[20px] p-3 text-left transition-all active:scale-[0.985] cursor-pointer flex items-center justify-between group shadow-2xs"
            >
              <div class="flex items-center gap-3">
                <div class="w-8 h-8 rounded-xl bg-[#0C0C0E]/5 text-[#0C0C0E] flex items-center justify-center shrink-0">
                  <BookOpen class="w-4 h-4 stroke-[2]" />
                </div>
                <div>
                  <h4 class="font-display font-bold text-xs sm:text-sm text-[#0C0C0E] group-hover:text-[#FF2400] transition-colors">
                    Historia & Horno Rayo
                  </h4>
                  <p class="text-[11px] text-ash">
                    El secreto de nuestra brasa a carbón
                  </p>
                </div>
              </div>
              <ChevronRight class="w-4 h-4 text-ash/70 group-hover:text-[#0C0C0E] transition-colors" />
            </button>

            <!-- Terminal de Personal (Solo si tiene rol o acceso) -->
            <button
              type="button"
              onclick={() => navigateTo('staff')}
              class="w-full bg-white hover:bg-[#FFFDF5] border border-black/[0.06] rounded-[20px] p-3 text-left transition-all active:scale-[0.985] cursor-pointer flex items-center justify-between group shadow-2xs"
            >
              <div class="flex items-center gap-3">
                <div class="w-8 h-8 rounded-xl bg-purple-500/10 text-purple-700 flex items-center justify-center shrink-0">
                  <Shield class="w-4 h-4 stroke-[2]" />
                </div>
                <div>
                  <h4 class="font-display font-bold text-xs sm:text-sm text-[#0C0C0E] group-hover:text-purple-700 transition-colors">
                    Terminal de Personal
                  </h4>
                  <p class="text-[11px] text-ash">
                    Comandas y caja para meseros y cajeros
                  </p>
                </div>
              </div>
              <ChevronRight class="w-4 h-4 text-ash/70 group-hover:text-[#0C0C0E] transition-colors" />
            </button>
          </div>
        </div>

        <!-- SECCIÓN 3: ATENCIÓN & HORARIO -->
        <div class="bg-white rounded-[24px] p-3.5 border border-black/[0.06] space-y-2.5 text-left shadow-2xs">
          <div class="flex items-center justify-between">
            <span class="text-[10px] font-mono font-bold tracking-wider text-ash uppercase flex items-center gap-1.5">
              <Clock class="w-3.5 h-3.5" />
              Horario de Atención
            </span>
            <span class="flex items-center gap-1 text-[10px] font-mono font-bold {isOpenNow ? 'text-emerald-600' : 'text-[#FF2400]'}">
              <span class="w-1.5 h-1.5 rounded-full {isOpenNow ? 'bg-emerald-500 animate-pulse' : 'bg-[#FF2400]'}"></span>
              {isOpenNow ? 'Abierto Ahora' : 'Cerrado Ahora'}
            </span>
          </div>
          <p class="text-xs font-mono text-[#0C0C0E]">
            Sáb a Jue: <strong>3:00 PM – 9:30 PM</strong> (Vie. cerrado)
          </p>

          <!-- WhatsApp de Atención -->
          <a
            href="https://api.whatsapp.com/send?phone=573212520881&text=Hola%20M%C3%ADas%2C%20quisiera%20hacer%20un%20pedido%20o%20consultar%20la%20carta."
            target="_blank"
            rel="noopener noreferrer"
            class="w-full py-2 px-3 rounded-xl bg-emerald-600/10 hover:bg-emerald-600/20 text-emerald-800 font-mono text-xs font-bold flex items-center justify-between transition-colors mt-1"
          >
            <div class="flex items-center gap-2">
              <Phone class="w-3.5 h-3.5 text-emerald-700" />
              <span>WhatsApp: 321 252 0881</span>
            </div>
            <ArrowRight class="w-3.5 h-3.5 text-emerald-700" />
          </a>
        </div>
      </div>

      <!-- FOOTER DEL MENÚ LATERAL -->
      <footer class="p-4 sm:p-5 bg-white border-t border-black/[0.06] shrink-0 text-center select-none">
        <div class="flex items-center justify-center gap-2 mb-1">
          <img src="/imagenes/marca/isotipo.webp" alt="Mías" class="w-5 h-5 object-contain" />
          <span class="font-display font-bold text-xs text-[#0C0C0E]">Mías Hamburguesas de Lujo</span>
        </div>
        <p class="text-[10px] font-mono text-ash">
          Duitama, Boyacá • Lo Íntimo Hecho para Compartir
        </p>
      </footer>
    </aside>
  </div>
{/if}
