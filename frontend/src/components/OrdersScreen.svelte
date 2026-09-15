<script>
  import { onMount } from 'svelte';
  import { store } from '../lib/store.svelte.js';
  import { getOrderByCodeAPI, getMyOrdersAPI } from '../lib/api.js';
  import {
    ChevronLeft,
    Clock,
    Flame,
    Bike,
    Check,
    RotateCcw,
    Search,
    RefreshCw,
    ShoppingBag,
    Utensils,
    AlertCircle,
    CheckCircle2,
    CreditCard,
    Award
  } from 'lucide-svelte';

  let { onBack = () => store.currentView = 'menu' } = $props();

  let isRefreshing = $state(false);
  let showSearchInput = $state(false);
  let searchQuery = $state('');
  let isSearching = $state(false);
  let searchError = $state('');

  // Historial unificado cronológico (más recientes primero)
  let allOrders = $derived(
    [...store.activeOrders].sort((a, b) => new Date(b.created_at || 0) - new Date(a.created_at || 0))
  );

  onMount(() => {
    store.refreshUserOrders();
  });

  async function handleRefresh() {
    isRefreshing = true;
    try {
      await store.refreshUserOrders();
    } finally {
      setTimeout(() => {
        isRefreshing = false;
      }, 400);
    }
  }

  function handleReorder(order) {
    if (!order.items || order.items.length === 0) {
      store.currentView = 'menu';
      return;
    }

    for (const it of order.items) {
      const prod = store.products.find(
        (p) => p.id === it.product_id || p.name?.toLowerCase() === (it.product_name || '').toLowerCase()
      );
      if (prod) {
        store.addToCart(prod, Boolean(it.is_combo), it.extras || []);
      }
    }
    store.isCartOpen = true;
  }

  async function handleSearch() {
    const q = searchQuery.trim();
    if (!q) return;

    isSearching = true;
    searchError = '';

    try {
      const cleanCode = q.replace('#', '').toUpperCase();
      const res = await getOrderByCodeAPI(cleanCode);

      if (res && res.order) {
        const o = res.order;
        const mapped = {
          id: o.id,
          code: o.code,
          type: o.order_type,
          modeLabel: o.order_type === 'delivery' ? 'A Domicilio' : o.order_type === 'takeout' ? 'Para Llevar' : 'En el Local',
          status: o.status,
          customer_name: o.delivery_name || '',
          customer_phone: o.delivery_phone || '',
          delivery_address: o.delivery_address || '',
          table_number: '',
          total: o.total_amount,
          items: res.items || [],
          created_at: o.created_at
        };

        const exists = store.activeOrders.some((item) => item.code === mapped.code);
        if (!exists) {
          store.activeOrders = [mapped, ...store.activeOrders];
          if (typeof window !== 'undefined') {
            localStorage.setItem('mias_orders', JSON.stringify(store.activeOrders));
          }
        }

        searchQuery = '';
        showSearchInput = false;
        return;
      }

      const phoneRes = await getMyOrdersAPI([], q);
      if (phoneRes && phoneRes.length > 0) {
        await store.refreshUserOrders();
        searchQuery = '';
        showSearchInput = false;
        return;
      }

      searchError = 'No encontramos ningún pedido con ese código o teléfono.';
    } catch {
      searchError = 'Error de conexión. Intenta nuevamente.';
    } finally {
      isSearching = false;
    }
  }

  function formatCOP(val) {
    return new Intl.NumberFormat('es-CO', {
      style: 'currency',
      currency: 'COP',
      maximumFractionDigits: 0
    }).format(val || 0);
  }

  function formatOrderTime(dateStr) {
    if (!dateStr) return 'Reciente';
    try {
      const d = new Date(dateStr);
      if (isNaN(d.getTime())) return dateStr;
      const now = new Date();
      const isToday = d.toDateString() === now.toDateString();
      const time = d.toLocaleTimeString('es-CO', { hour: 'numeric', minute: '2-digit', hour12: true });
      return isToday ? `Hoy · ${time}` : `${d.toLocaleDateString('es-CO', { day: 'numeric', month: 'short' })} · ${time}`;
    } catch {
      return dateStr;
    }
  }

  function getPaymentInfo(ord) {
    if (ord.status === 'PAID') {
      return {
        isPaid: true,
        label: 'Pagado en caja',
        subtext: 'Cobro verificado',
        badgeClass: 'border-white/10 bg-white/[0.03] text-bone',
        icon: Check
      };
    }

    if (ord.status === 'DELIVERED') {
      return {
        isPaid: true,
        label: 'Entregado y pagado',
        subtext: 'Orden finalizada',
        badgeClass: 'border-white/10 bg-white/[0.03] text-bone',
        icon: Check
      };
    }

    if (ord.status === 'CANCELLED') {
      return {
        isPaid: false,
        isCancelled: true,
        label: 'Pedido cancelado',
        subtext: 'Sin cobro registrado',
        badgeClass: 'border-ember/30 bg-ember/10 text-ember-light',
        icon: AlertCircle
      };
    }

    if (ord.type === 'in_store') {
      return {
        isPaid: false,
        needsPayment: true,
        label: 'Falta pagar en caja',
        subtext: `Pasa a caja física con tu código #${ord.code || ''}`,
        badgeClass: 'border-gold/40 bg-gold/[0.06] text-gold',
        icon: CreditCard
      };
    }

    return {
      isPaid: false,
      needsPayment: true,
      label: 'Pago contra entrega',
      subtext: `Pagas al recibir: ${formatCOP(ord.total)}`,
      badgeClass: 'border-gold/40 bg-gold/[0.06] text-gold',
      icon: CreditCard
    };
  }

  function getPreparationLabel(status) {
    switch (status) {
      case 'PENDING':
        return { text: 'En espera de cocina', color: 'text-gold', icon: Clock };
      case 'CONFIRMED':
      case 'PREPARING':
        return { text: 'En el horno al carbón', color: 'text-ember-light', icon: Flame };
      case 'IN_DELIVERY':
        return { text: 'En camino con el repartidor', color: 'text-gold', icon: Bike };
      case 'READY':
        return { text: 'Listo para entrega', color: 'text-bone', icon: CheckCircle2 };
      case 'DELIVERED':
      case 'PAID':
        return { text: 'Completado', color: 'text-bone', icon: Check };
      default:
        return { text: 'En proceso', color: 'text-ash', icon: Clock };
    }
  }
</script>

<div class="min-h-screen bg-char text-bone pb-32 pt-2 px-3 sm:px-6 max-w-xl mx-auto select-none">

  <!-- ═══════════════════════════════════════════════════════════════════════ -->
  <!-- BARRA SUPERIOR EDITORIAL                                               -->
  <!-- ═══════════════════════════════════════════════════════════════════════ -->
  <header class="flex items-center justify-between py-4 mb-4">
    <button
      type="button"
      onclick={onBack}
      class="w-12 h-12 rounded-full bg-char-card border hairline-dark flex items-center justify-center text-bone hover:border-white/20 active:scale-95 transition-all cursor-pointer shadow-lux"
      aria-label="Volver a la carta"
    >
      <ChevronLeft class="w-5 h-5 stroke-[2.2]" />
    </button>

    <div class="text-center px-2">
      <h1 class="font-serif text-2xl sm:text-3xl text-bone font-normal tracking-tight leading-tight">
        Historial de Pedidos
      </h1>
      <span class="font-sans text-xs text-ash block mt-0.5">
        Comandas activas y recientes
      </span>
    </div>

    <button
      type="button"
      onclick={handleRefresh}
      class="w-12 h-12 rounded-full bg-char-card border hairline-dark flex items-center justify-center text-bone hover:border-white/20 active:scale-95 transition-all cursor-pointer shadow-lux"
      aria-label="Actualizar comanda"
    >
      <RefreshCw class="w-4 h-4 {isRefreshing ? 'animate-spin text-gold' : ''}" />
    </button>
  </header>

  <!-- ═══════════════════════════════════════════════════════════════════════ -->
  <!-- LISTA CONTINUA DE COMANDAS (ESTÁNDAR APPLE CARDS 42px / 48px)          -->
  <!-- ═══════════════════════════════════════════════════════════════════════ -->
  {#if allOrders.length === 0}
    <!-- ESTADO VACÍO EDITORIAL -->
    <div class="py-16 px-6 text-center rounded-card sm:rounded-card-lg bg-char-card border hairline-dark flex flex-col items-center justify-center space-y-4 shadow-lux">
      <div class="w-16 h-16 rounded-full bg-gold/10 border border-gold/30 flex items-center justify-center text-gold">
        <ShoppingBag class="w-8 h-8 stroke-[1.8]" />
      </div>

      <div class="space-y-1 max-w-xs">
        <h2 class="font-serif text-2xl text-bone font-normal">
          Sin pedidos registrados
        </h2>
        <p class="font-sans text-xs sm:text-sm text-ash leading-relaxed">
          Tus comandas en el restaurante o a domicilio se registrarán automáticamente aquí.
        </p>
      </div>

      <button
        type="button"
        onclick={onBack}
        class="w-full min-h-[56px] rounded-full bg-ember hover:bg-ember-light active:scale-[0.98] text-white font-sans font-bold text-base tracking-wide transition-all flex items-center justify-center gap-2.5 cursor-pointer shadow-lux mt-2"
      >
        <Utensils class="w-4 h-4 stroke-[2]" />
        <span>Ver Carta y Ordenar</span>
      </button>
    </div>
  {:else}
    <!-- LISTADO DE TARJETAS APPLE CARD -->
    <div class="space-y-5">
      {#each allOrders as ord, idx (ord.code || ord.id || idx)}
        {@const pay = getPaymentInfo(ord)}
        {@const prep = getPreparationLabel(ord.status)}
        {@const PrepIcon = prep.icon}

        <article class="rounded-card sm:rounded-card-lg bg-char-card border hairline-dark p-5 sm:p-7 space-y-4 shadow-lux transition-all">
          
          <!-- 1. IDENTIFICACIÓN Y TOTAL DE COMANDA -->
          <div class="flex items-start justify-between gap-3">
            <div>
              <div class="flex items-center gap-2">
                <span class="font-mono font-bold text-2xl sm:text-3xl text-bone tracking-tight">
                  #{ord.code || 'ORD'}
                </span>
              </div>
              <div class="flex items-center gap-2 mt-1">
                <span class="font-sans text-xs font-medium text-ash">
                  {#if ord.type === 'in_store'}
                    Consumo en el Local
                  {:else if ord.type === 'delivery'}
                    A Domicilio
                  {:else}
                    Para Llevar
                  {/if}
                </span>
                <span class="text-ash/40">•</span>
                <span class="font-sans text-xs text-ash">
                  {formatOrderTime(ord.created_at)}
                </span>
              </div>
            </div>

            <div class="text-right">
              <span class="font-sans text-[11px] uppercase tracking-wider text-ash block">Total</span>
              <span class="font-mono font-bold text-2xl sm:text-3xl text-gold block leading-tight">
                {formatCOP(ord.total)}
              </span>
            </div>
          </div>

          <!-- 2. ESTADO DE PAGO EDITORIAL (SIN CI/CD NI ANIMATE-PING) -->
          <div class="w-full rounded-2xl border p-3.5 flex items-center justify-between gap-3 {pay.badgeClass}">
            <div class="flex items-center gap-2.5 min-w-0">
              <pay.icon class="w-4 h-4 shrink-0" />
              <div class="min-w-0">
                <span class="font-sans font-bold text-xs sm:text-sm block leading-tight truncate">
                  {pay.label}
                </span>
                <span class="font-sans text-[11px] opacity-80 block leading-tight truncate mt-0.5">
                  {pay.subtext}
                </span>
              </div>
            </div>
          </div>

          <!-- 3. ESTADO OPERATIVO DE COCINA -->
          <div class="flex items-center gap-3 px-1 py-1">
            <PrepIcon class="w-4 h-4 {prep.color} shrink-0" />
            <span class="font-sans text-xs text-ash">
              Cocina: <strong class="text-bone font-medium">{prep.text}</strong>
            </span>
          </div>

          <!-- 4. DESGLOSE DE PLATOS -->
          {#if ord.items && ord.items.length > 0}
            <div class="border-t hairline-dark pt-3 space-y-1.5">
              <span class="font-sans text-[11px] text-ash uppercase font-semibold tracking-wider block px-0.5">
                Platos ({ord.items.reduce((acc, i) => acc + (i.quantity || 1), 0)})
              </span>
              <div class="space-y-1.5">
                {#each ord.items as it}
                  <div class="flex items-center justify-between text-xs py-0.5">
                    <span class="font-sans text-bone truncate pr-2">
                      <span class="font-mono font-bold text-gold">{it.quantity || 1}x</span> {it.product_name}
                      {#if it.is_combo}
                        <span class="font-sans text-[10px] text-gold bg-gold/10 px-1.5 py-0.2 rounded-sm ml-1">Combo</span>
                      {/if}
                    </span>
                    <span class="font-mono text-bone/80 shrink-0 font-medium">
                      {formatCOP((it.unit_price || 0) * (it.quantity || 1))}
                    </span>
                  </div>
                {/each}
              </div>
            </div>
          {/if}

          <!-- 5. ACCIONES DE COMANDA -->
          <div class="pt-2 space-y-2.5">
            
            <!-- Acreditación de Sellos (solo órdenes con hamburguesas) -->
            {#if store.orderHasBurgers(ord)}
              {#if ord.stamp_status === 'APPROVED'}
                <div class="w-full min-h-[50px] rounded-full bg-char border border-gold/30 text-gold font-sans font-medium text-xs sm:text-sm flex items-center justify-center gap-2">
                  <Check class="w-4 h-4 text-gold stroke-[2.5]" />
                  <span>Sello acreditado en tu cartilla (+1)</span>
                </div>
              {:else if ord.stamp_status === 'PENDING'}
                <div class="w-full min-h-[50px] rounded-full bg-char border hairline-dark text-gold/80 font-sans font-medium text-xs sm:text-sm flex items-center justify-center gap-2">
                  <Clock class="w-4 h-4 text-gold" />
                  <span>Sello en validación de caja...</span>
                </div>
              {:else if pay.needsPayment}
                <div class="w-full min-h-[50px] rounded-full bg-char/70 border hairline-dark text-ash font-sans text-xs px-4 py-2 flex items-center justify-between gap-2">
                  <div class="flex items-center gap-2 min-w-0">
                    <AlertCircle class="w-4 h-4 text-gold shrink-0" />
                    <span class="truncate">Paga en caja para registrar tu sello</span>
                  </div>
                  <span class="font-sans text-[11px] font-medium text-gold shrink-0">Pendiente</span>
                </div>
              {:else}
                <button
                  type="button"
                  onclick={() => store.openStampModal(ord)}
                  class="w-full min-h-[54px] sm:min-h-[56px] rounded-full bg-gold/10 hover:bg-gold/20 border border-gold/40 text-gold active:scale-[0.98] font-sans font-bold text-sm sm:text-base tracking-wide transition-all flex items-center justify-center gap-2 cursor-pointer shadow-lux"
                >
                  <Award class="w-4 h-4 text-gold stroke-[2]" />
                  <span>Solicitar Sello en mi Cartilla</span>
                </button>
              {/if}
            {/if}

            <!-- Pedir de Nuevo -->
            <button
              type="button"
              onclick={() => handleReorder(ord)}
              class="w-full min-h-[56px] sm:min-h-[58px] rounded-full bg-ember hover:bg-ember-light active:scale-[0.98] text-white font-sans font-bold text-base tracking-wide transition-all flex items-center justify-center gap-2.5 cursor-pointer shadow-lux"
            >
              <RotateCcw class="w-4 h-4 stroke-[2.2]" />
              <span>Pedir de nuevo</span>
            </button>
          </div>

        </article>
      {/each}
    </div>
  {/if}

  <!-- ═══════════════════════════════════════════════════════════════════════ -->
  <!-- BUSCADOR DISCRETO                                                       -->
  <!-- ═══════════════════════════════════════════════════════════════════════ -->
  <div class="mt-8 text-center pt-4 border-t hairline-dark">
    {#if !showSearchInput}
      <button
        type="button"
        onclick={() => showSearchInput = true}
        class="min-h-[46px] px-5 py-2 rounded-full bg-char-card border hairline-dark text-xs font-sans text-ash hover:text-bone hover:border-white/20 transition-all inline-flex items-center gap-2 cursor-pointer active:scale-95"
      >
        <Search class="w-3.5 h-3.5" />
        <span>¿No ves una comanda previa? Buscar por código</span>
      </button>
    {:else}
      <div class="p-5 sm:p-6 rounded-card sm:rounded-card-lg bg-char-card border hairline-dark text-left space-y-3 shadow-lux">
        <span class="text-xs font-sans font-medium text-bone block">
          Ingresa el código (#ORD-...) o tu teléfono:
        </span>
        <div class="flex flex-col sm:flex-row gap-2.5">
          <input
            type="text"
            bind:value={searchQuery}
            onkeydown={(e) => e.key === 'Enter' && handleSearch()}
            placeholder="Ej: ORD-A102 o 3212520881"
            class="flex-1 min-h-[50px] px-4 bg-char border hairline-dark rounded-full text-sm font-mono text-bone focus:outline-none focus:border-gold transition-colors"
          />
          <button
            type="button"
            onclick={handleSearch}
            disabled={isSearching || !searchQuery.trim()}
            class="min-h-[50px] px-6 rounded-full bg-ember hover:bg-ember-light active:scale-95 text-white font-sans font-bold text-sm tracking-wide shrink-0 cursor-pointer disabled:opacity-40 transition-all"
          >
            {isSearching ? 'Buscando...' : 'Buscar'}
          </button>
        </div>
        {#if searchError}
          <p class="text-xs font-sans text-ember-light">{searchError}</p>
        {/if}
      </div>
    {/if}
  </div>

</div>
