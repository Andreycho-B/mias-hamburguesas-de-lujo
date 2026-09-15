<script>
  import { store } from '../lib/store.svelte.js';
  import { createOrderAPI } from '../lib/api.js';
  import {
    X,
    Plus,
    Minus,
    ShoppingBag,
    Sparkles,
    ArrowRight,
    Check,
    ChevronLeft,
    UtensilsCrossed,
    Bike,
    Send,
    Copy,
    DollarSign,
    CreditCard,
    Award,
    MapPin,
    Phone,
    User,
    LogIn,
    Flame,
    Utensils,
    Clock,
    Receipt
  } from 'lucide-svelte';

  // ═══════════════════════════════════════════════════════════════════════
  // ESTADOS DEL FLUJO MULTI-PÁGINA
  // Pasos: 'upsell' → 'review' → 'mode' → 'delivery_form' → 'ticket'
  // En salón: upsell → review → mode (auto-finaliza) → ticket
  // ═══════════════════════════════════════════════════════════════════════
  let cartStep = $state('upsell');
  let selectedMode = $state('delivery');

  // Estado de upsell (burbujas de acompañamientos)
  let selectedUpsells = $state([]);
  let highlightedUpsell = $state(null);

  // Datos del cliente para la orden
  let customerName = $state(store.user?.full_name || '');
  let customerPhone = $state(store.user?.phone || '');
  let deliveryAddress = $state(store.user?.address || '');
  let deliveryNeighborhood = $state('Centro');
  let deliveryNotes = $state('');
  let tableNumber = $state('1');

  // Método de pago
  let paymentMethod = $state('Efectivo');
  let cashAmount = $state('');
  let formError = $state('');

  const deliveryFee = 0;
  let createdOrder = $state(null);
  let isCopied = $state(false);

  // Producto hero: primer ítem del carrito (imagen central del upsell)
  let heroProduct = $derived(store.cart.length > 0 ? store.cart[0] : null);

  // Desglose reactivo de unidades individuales añadidas al pedido para posicionamiento en escena
  let cartUnits = $derived.by(() => {
    const units = [];
    for (const item of store.cart) {
      const qty = item.quantity || 1;
      for (let q = 0; q < qty; q++) {
        units.push({
          id: `${item.id}-${q}`,
          name: item.product?.name || 'Producto',
          image_url: item.product?.image_url || '',
          price: item.product?.price || 0,
          slug: item.product?.slug || ''
        });
      }
    }
    return units;
  });

  let liveTicketOrder = $derived(
    createdOrder ? (store.activeOrders.find(o => o.code === createdOrder.code) || createdOrder) : null
  );

  // Sincronizar datos si el usuario inicia sesión
  $effect(() => {
    if (store.user) {
      if (!customerName) customerName = store.user.full_name || '';
      if (!customerPhone) customerPhone = store.user.phone || '';
      if (!deliveryAddress && store.user.address) deliveryAddress = store.user.address;
    }
  });

  // Redirigir a upsell si el carrito queda vacío durante review
  $effect(() => {
    if (store.cart.length === 0 && cartStep === 'review') {
      cartStep = 'upsell';
    }
  });

  // Reset a upsell cuando hay nuevos ítems tras una orden completada
  $effect(() => {
    if (store.isCartOpen && store.cart.length > 0 && createdOrder) {
      createdOrder = null;
      cartStep = 'upsell';
      selectedUpsells = [];
      highlightedUpsell = null;
      formError = '';
    }
  });

  let currentDeliveryFee = $derived(selectedMode === 'delivery' ? deliveryFee : 0);
  let currentTotal = $derived(store.cartSubtotal + currentDeliveryFee);

  let cashChange = $derived.by(() => {
    const val = parseInt(cashAmount, 10);
    if (!isNaN(val) && val >= currentTotal) return val - currentTotal;
    return 0;
  });

  // ═══════════════════════════════════════════════════════════════════════
  // CATÁLOGO AMPLIADO DE ACOMPAÑAMIENTOS (6 BURBUJAS)
  // ═══════════════════════════════════════════════════════════════════════
  const upsellItems = [
    { id: 14, category_id: 4, name: 'Coca-Cola 400ml', shortName: 'Coca-Cola', price: 4000, image_url: '/imagenes/bebidas/cocacola.webp', type: 'beverage' },
    { id: 15, category_id: 4, name: 'Bretaña en Vidrio', shortName: 'Bretaña', price: 5000, image_url: '/imagenes/bebidas/bretana-vidrio.png', type: 'beverage' },
    { id: 17, category_id: 5, name: 'Papas Francesas', shortName: 'Papas', price: 8000, image_url: '', type: 'fries' },
    { id: 18, category_id: 5, name: 'Papa Nativa', shortName: 'P. Nativa', price: 9000, image_url: '', type: 'fries' },
    { id: 19, category_id: 5, name: 'Tocineta Crujiente', shortName: 'Tocineta', price: 5000, image_url: '', type: 'meat' },
    { id: 21, category_id: 5, name: 'Queso Fundido', shortName: 'Queso', price: 5000, image_url: '', type: 'cheese' }
  ];

  // ═══════════════════════════════════════════════════════════════════════
  // FUNCIONES DE FORMATO
  // ═══════════════════════════════════════════════════════════════════════
  function formatCOP(val) {
    return new Intl.NumberFormat('es-CO', {
      style: 'currency',
      currency: 'COP',
      maximumFractionDigits: 0
    }).format(val || 0);
  }

  function formatTicketTime(dateStr) {
    try {
      const d = dateStr ? new Date(dateStr) : new Date();
      return d.toLocaleTimeString('en-US', { hour: '2-digit', minute: '2-digit', hour12: true });
    } catch {
      return '03:00 PM';
    }
  }

  function formatTicketDate(dateStr) {
    try {
      const d = dateStr ? new Date(dateStr) : new Date();
      const meses = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
      const dias = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
      return `${meses[d.getMonth()]} ${d.getDate()}, ${dias[d.getDay()]}`;
    } catch {
      return 'Sep 14, Sun';
    }
  }

  // ═══════════════════════════════════════════════════════════════════════
  // NAVEGACIÓN ENTRE PASOS
  // ═══════════════════════════════════════════════════════════════════════
  function handleToggleUpsell(item) {
    const idx = selectedUpsells.findIndex(u => u.id === item.id);
    if (idx > -1) {
      selectedUpsells = selectedUpsells.filter(u => u.id !== item.id);
      highlightedUpsell = selectedUpsells.length > 0 ? selectedUpsells[selectedUpsells.length - 1] : null;
    } else {
      selectedUpsells = [...selectedUpsells, item];
      highlightedUpsell = item;
    }
  }

  function handleContinueFromUpsell() {
    for (const item of selectedUpsells) {
      store.addToCart(item, false, []);
    }
    selectedUpsells = [];
    highlightedUpsell = null;
    cartStep = 'review';
  }

  function handleContinueFromReview() {
    if (store.cart.length === 0) return;
    if (!store.user) {
      formError = 'Inicia sesión o regístrate para enviar tu comanda.';
      store.isAuthModalOpen = true;
      return;
    }
    formError = '';
    cartStep = 'mode';
  }

  function handleBackStep() {
    formError = '';
    if (cartStep === 'delivery_form') {
      cartStep = 'mode';
    } else if (cartStep === 'mode') {
      cartStep = 'review';
    } else if (cartStep === 'review') {
      cartStep = 'upsell';
    } else if (cartStep === 'ticket') {
      cartStep = 'upsell';
    }
  }

  async function handleSelectMode(mode) {
    if (!store.user) {
      formError = 'Inicia sesión o regístrate para ordenar.';
      store.isAuthModalOpen = true;
      return;
    }
    selectedMode = mode;
    formError = '';
    if (mode === 'in_store') {
      // Consumo en salón: sin fricción, directo a cocina
      await handleFinalizeOrder();
    } else {
      cartStep = 'delivery_form';
    }
  }

  function selectQuickCash(amount) {
    if (amount === 'exact') {
      cashAmount = String(currentTotal);
    } else {
      cashAmount = String(amount);
    }
  }

  function toggleQuickNote(note) {
    if (deliveryNotes.includes(note)) {
      deliveryNotes = deliveryNotes.replace(note, '').replace(/,\s*,/g, ',').trim();
    } else {
      deliveryNotes = deliveryNotes ? `${deliveryNotes}, ${note}` : note;
    }
  }

  // ═══════════════════════════════════════════════════════════════════════
  // FINALIZACIÓN DE ORDEN (PRESERVADO ÍNTEGRO)
  // ═══════════════════════════════════════════════════════════════════════
  async function handleFinalizeOrder() {
    formError = '';

    if (!store.user) {
      formError = 'Debes iniciar sesión o registrarte para realizar tu comanda.';
      store.isAuthModalOpen = true;
      return;
    }

    const effectiveCustomerName = (customerName.trim() || store.user?.full_name || 'Comensal Mías');
    const effectiveCustomerPhone = (customerPhone.trim() || store.user?.phone || '');

    if (selectedMode !== 'in_store') {
      if (!effectiveCustomerName) {
        formError = 'Por favor ingresa tu nombre para registrar la comanda.';
        return;
      }
      if (!effectiveCustomerPhone) {
        formError = 'Por favor ingresa tu número de WhatsApp para notificarte.';
        return;
      }
    }

    if (selectedMode === 'delivery') {
      if (!deliveryAddress.trim() || !deliveryNeighborhood.trim()) {
        formError = 'Por favor completa la dirección y el barrio de entrega en Duitama.';
        return;
      }
    }

    let code = (selectedMode === 'delivery' ? 'DOM-' : 'ORD-') + Math.floor(1000 + Math.random() * 9000);
    let orderId = (selectedMode === 'delivery' ? 'dom-' : 'ord-') + Date.now();

    const itemsSnapshot = [...store.cart];
    const subtotalSnapshot = store.cartSubtotal;
    const feeSnapshot = currentDeliveryFee;
    const totalSnapshot = currentTotal;

    const order = {
      id: orderId,
      code,
      type: selectedMode,
      modeLabel: selectedMode === 'delivery' ? 'A Domicilio' : selectedMode === 'takeout' ? 'Para Llevar' : 'Consumo en Salón',
      status: selectedMode === 'delivery' ? 'IN_PREPARATION' : 'PENDING_PAYMENT',
      customer_name: effectiveCustomerName,
      customer_phone: effectiveCustomerPhone,
      delivery_address: selectedMode === 'delivery' ? `${deliveryAddress.trim()}, Barrio ${deliveryNeighborhood.trim()}` : '',
      table_number: selectedMode === 'in_store' ? (tableNumber.trim() || '1') : '',
      delivery_notes: selectedMode === 'takeout'
        ? (`[PARA LLEVAR] ${deliveryNotes.trim()}`).trim()
        : selectedMode === 'in_store'
        ? (deliveryNotes.trim() || 'Consumo en Salón')
        : deliveryNotes.trim(),
      items: itemsSnapshot,
      subtotal: subtotalSnapshot,
      delivery_fee: feeSnapshot,
      total: totalSnapshot,
      payment_method: selectedMode === 'in_store' ? 'Pago en Caja' : paymentMethod,
      cash_change_amount: parseInt(cashAmount, 10) || 0,
      created_at: new Date().toISOString()
    };

    // Redirección inmediata a ticket sin bloqueo de red
    store.addOrder(order);
    createdOrder = order;
    store.clearCart();
    cartStep = 'ticket';

    // Persistencia asíncrona en Backend Anillo 0
    const apiItems = itemsSnapshot.map(it => ({
      product_id: it.product.id,
      quantity: it.quantity,
      customizations: {
        combo: it.combo,
        extras: it.extras || []
      }
    }));

    (async () => {
      try {
        const apiRes = await createOrderAPI({
          order_type: order.type === 'delivery' ? 'delivery' : 'in_store',
          items: apiItems,
          delivery_name: order.customer_name,
          delivery_phone: order.customer_phone,
          delivery_address: order.delivery_address,
          delivery_notes: order.delivery_notes,
          payment_method: order.payment_method,
          cash_change_amount: order.cash_change_amount
        });
        if (apiRes && apiRes.order) {
          order.code = apiRes.order.code;
          order.id = apiRes.order.id;
          createdOrder = { ...order };
        }
      } catch (err) {
        console.warn('[Order API] Guardado local Anillo 0', err);
      }
    })();
  }

  function dispatchWhatsApp() {
    if (!createdOrder) return;
    let msg = `🍔 *NUEVA COMANDA — MÍAS HAMBURGUESAS* 🍔\n` +
      `*Orden:* #${createdOrder.code}\n` +
      `*Modalidad:* ${createdOrder.modeLabel}\n` +
      `*Cliente:* ${createdOrder.customer_name}` + (createdOrder.customer_phone ? ` (${createdOrder.customer_phone})` : '') + `\n`;

    if (createdOrder.delivery_address) {
      msg += `*Dirección:* ${createdOrder.delivery_address}\n`;
    }
    if (createdOrder.delivery_notes) {
      msg += `*Notas:* ${createdOrder.delivery_notes}\n`;
    }

    msg += `\n*Detalle de Platos:*\n`;
    createdOrder.items.forEach((it) => {
      msg += `• ${it.quantity}x ${it.product.name} (${formatCOP(it.totalPrice)})\n`;
      if (it.combo) msg += `   - Con Combo (Papas y Bebida)\n`;
    });

    msg += `\n*Subtotal:* ${formatCOP(createdOrder.subtotal)}\n`;
    if (createdOrder.delivery_fee > 0) {
      msg += `*Costo Domicilio:* ${formatCOP(createdOrder.delivery_fee)}\n`;
    }
    msg += `*TOTAL A PAGAR:* ${formatCOP(createdOrder.total)}\n` +
      `*Método de Pago:* ${createdOrder.payment_method}\n`;

    if (createdOrder.payment_method === 'Efectivo' && createdOrder.cash_change_amount > 0) {
      msg += `*Paga con:* ${formatCOP(createdOrder.cash_change_amount)}\n` +
        `*Cambio requerido:* ${formatCOP(createdOrder.cash_change_amount - createdOrder.total)}\n`;
    }

    const encoded = encodeURIComponent(msg);
    const waURL = `https://api.whatsapp.com/send?phone=573212520881&text=${encoded}`;
    window.open(waURL, '_blank');
  }

  function copyNequi() {
    navigator.clipboard?.writeText('3212520881');
    isCopied = true;
    setTimeout(() => isCopied = false, 2000);
  }

  function goToStampsFromTicket() {
    store.isCartOpen = false;
    store.openStampModal(liveTicketOrder || createdOrder);
  }

  function closeOrderModal() {
    store.isCartOpen = false;
  }

  function addMoreProducts() {
    store.isCartOpen = false;
    store.currentView = 'menu';
  }
</script>


<!-- ═══════════════════════════════════════════════════════════════════════ -->
<!-- FLUJO MULTI-PÁGINA DE COMANDA — ZERO SCROLL                          -->
<!-- ═══════════════════════════════════════════════════════════════════════ -->
{#if store.isCartOpen}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    data-lenis-prevent
    onwheel={(e) => e.stopPropagation()}
    ontouchmove={(e) => e.stopPropagation()}
    class="fixed inset-0 z-50 {cartStep === 'upsell' ? 'bg-[#000000]' : 'bg-char'} text-bone flex flex-col"
    style="height: 100dvh; touch-action: pan-y; -webkit-overflow-scrolling: touch; {cartStep === 'upsell' ? 'padding-top: max(0.5rem, calc(env(safe-area-inset-top, 0px) + 0.35rem)); padding-bottom: max(1rem, env(safe-area-inset-bottom, 0px)); padding-left: max(0px, env(safe-area-inset-left, 0px)); padding-right: max(0px, env(safe-area-inset-right, 0px));' : 'padding-top: max(1rem, env(safe-area-inset-top, 0px)); padding-bottom: max(1rem, env(safe-area-inset-bottom, 0px)); padding-left: max(1rem, env(safe-area-inset-left, 0px)); padding-right: max(1rem, env(safe-area-inset-right, 0px));'}"
  >
    <div data-lenis-prevent class="w-full {cartStep === 'upsell' ? 'max-w-md md:max-w-4xl lg:max-w-5xl' : cartStep === 'mode' ? 'max-w-2xl' : 'max-w-lg'} mx-auto flex-1 flex flex-col min-h-0">

      <!-- ════════════════════════════════════════════════════════════════ -->
      <!-- CABECERA COMPARTIDA                                            -->
      <!-- ════════════════════════════════════════════════════════════════ -->
      <header class="flex items-center justify-between w-full {cartStep === 'upsell' ? 'py-1.5 sm:py-2 px-5 sm:px-6' : 'py-2 sm:py-3'} shrink-0">
        {#if cartStep === 'upsell'}
          <!-- Vista vaciada: solo botón de cierre en la misma posición que el corazón de Hamburguesas -->
          <div class="w-14 h-14 sm:w-16 sm:h-16 shrink-0"></div>
          <div class="flex-1"></div>
          <button
            type="button"
            onclick={closeOrderModal}
            class="btn-gold-outline w-14 h-14 sm:w-16 sm:h-16 rounded-full flex items-center justify-center text-bone hover:text-[#FF2400] active:scale-95 transition-all cursor-pointer shrink-0"
            aria-label="Cerrar comanda"
          >
            <X class="w-6 h-6 sm:w-6.5 sm:h-6.5 stroke-[2.2]" />
          </button>
        {:else}
          <!-- Izquierda: Logo o Atrás -->
          {#if cartStep === 'ticket'}
            <div class="w-14 h-14 sm:w-16 sm:h-16 flex items-center justify-center shrink-0">
              <img src="/imagenes/marca/isotipo.webp" alt="Mías" class="w-7 h-7 object-contain opacity-85" />
            </div>
          {:else}
            <button
              type="button"
              onclick={handleBackStep}
              class="w-14 h-14 sm:w-16 sm:h-16 rounded-full bg-char-card border border-white/10 flex items-center justify-center text-bone hover:bg-char-card-hover active:scale-90 transition-all cursor-pointer group shrink-0"
              aria-label="Volver al paso anterior"
            >
              <ChevronLeft class="w-6 h-6 sm:w-7 sm:h-7 stroke-[2.2] -translate-x-0.5 group-hover:-translate-x-1 transition-transform" />
            </button>
          {/if}

          <!-- Centro: Kicker + Título -->
          <div class="text-center flex-1 px-2">
            <span class="font-mono text-[10px] sm:text-xs uppercase tracking-widest text-[#8E8E93] font-bold block">
              {#if cartStep === 'review'}
                Mías • Tu Pedido
              {:else if cartStep === 'mode'}
                Mías • Modalidad
              {:else if cartStep === 'delivery_form'}
                Mías • Entrega
              {:else}
                Orden Procesada
              {/if}
            </span>
            <h1 class="font-display font-black text-xl sm:text-2xl text-bone tracking-tight leading-tight">
              {#if cartStep === 'review'}
                Confirma tu orden
              {:else if cartStep === 'mode'}
                ¿Dónde disfrutas?
              {:else if cartStep === 'delivery_form'}
                {selectedMode === 'delivery' ? 'Datos de Domicilio' : 'Datos de Retiro'}
              {:else}
                ¡Comanda en Marcha!
              {/if}
            </h1>
          </div>

          <!-- Derecha: Cerrar -->
          <button
            type="button"
            onclick={closeOrderModal}
            class="btn-gold-outline w-14 h-14 sm:w-16 sm:h-16 rounded-full flex items-center justify-center text-bone hover:text-[#FF2400] active:scale-95 transition-all cursor-pointer shrink-0"
            aria-label="Cerrar comanda"
          >
            <X class="w-6 h-6 sm:w-6.5 sm:h-6.5 stroke-[2.2]" />
          </button>
        {/if}
      </header>


      <!-- ════════════════════════════════════════════════════════════════ -->
      <!-- PASO 1 — UPSELL: ESCENARIO DE PRODUCTOS APILADOS EN PROFUNDIDAD   -->
      <!-- ════════════════════════════════════════════════════════════════ -->
      {#if cartStep === 'upsell'}
        <div class="flex-1 flex flex-col items-center justify-center w-full px-4 relative overflow-hidden">
          {#if cartUnits.length === 0}
            <!-- Estado sin productos en comanda -->
            <div class="text-center py-12 space-y-2 select-none">
              <p class="font-serif text-lg text-bone/60">Tu orden está vacía</p>
              <p class="font-sans text-xs text-ash">Agrega productos para personalizarlos</p>
            </div>
          {:else}
            <!-- Escenario interactivo de productos apilados hacia la derecha y en profundidad -->
            <div 
              class="relative flex items-center justify-center w-full select-none"
              style="min-height: 280px; max-height: 380px;"
            >
              <!-- Sombra de contacto rasante en el piso -->
              <div 
                class="absolute bottom-3 sm:bottom-5 w-52 sm:w-64 h-6 rounded-[100%] bg-black/90 blur-xl pointer-events-none"
                style="transform: scale({cartUnits.length === 1 ? 1.15 : 1 + Math.min(2.2, (cartUnits.length - 1) * 0.35)});"
              ></div>

              <!-- Composición: producto principal centrado y protagónico, secundarios a los costados -->
              {#each cartUnits as unit, i (unit.id)}
                {@const isHero = i === 0}
                {@const side = (i % 2 === 1) ? 1 : -1}
                {@const rank = Math.ceil(i / 2)}
                {@const stepX = 68}
                {@const stepY = 10}
                {@const posX = isHero ? 0 : side * (rank * stepX)}
                {@const posY = isHero ? 0 : -(rank * stepY)}
                {@const scale = isHero ? 1.16 : Math.max(0.68, 0.84 - ((rank - 1) * 0.12))}
                {@const zIndex = isHero ? 40 : 35 - (rank * 8) - (side < 0 ? 1 : 0)}
                {@const brightness = isHero ? 1.0 : Math.max(0.74, 0.90 - ((rank - 1) * 0.08))}

                <div 
                  class="absolute flex items-center justify-center transition-all duration-500 ease-out pointer-events-none"
                  style="
                    transform: translate3d({posX}px, {posY}px, 0) scale({scale});
                    z-index: {zIndex};
                    filter: brightness({brightness}) drop-shadow(0 16px 30px rgba(0,0,0,0.85));
                  "
                >
                  <div class="w-48 h-48 sm:w-56 sm:h-56 md:w-64 md:h-64 flex items-center justify-center">
                    {#if unit.image_url}
                      <img 
                        src={unit.image_url} 
                        alt={unit.name} 
                        class="w-full h-full object-contain select-none pointer-events-none"
                        loading="eager"
                      />
                    {:else}
                      <div class="w-36 h-36 rounded-full bg-char-card border border-white/10 flex items-center justify-center text-bone">
                        <ShoppingBag class="w-12 h-12 text-ash" />
                      </div>
                    {/if}
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>


      <!-- ════════════════════════════════════════════════════════════════ -->
      <!-- PASO 2 — REVIEW: RESUMEN INTERACTIVO DEL PEDIDO                -->
      <!-- ════════════════════════════════════════════════════════════════ -->
      {:else if cartStep === 'review'}
        <div class="flex-1 flex flex-col min-h-0 py-2">
          <!-- Subtítulo -->
          <p class="text-center text-sm text-ash mb-4">
            ¿Quieres agregar o quitar algo?
          </p>

          <!-- Lista de ítems (scroll aislado si excede el viewport) -->
          <div class="flex-1 min-h-0 overflow-y-auto space-y-2 overscroll-contain" style="touch-action: pan-y;">
            {#each store.cart as item (item.id)}
              <div class="flex items-center gap-3 p-3 sm:p-3.5 bg-char-card rounded-2xl border border-white/10">
                <!-- Miniatura -->
                <div class="w-11 h-11 sm:w-12 sm:h-12 rounded-xl bg-char flex items-center justify-center shrink-0 overflow-hidden border border-white/10">
                  {#if item.product?.image_url}
                    <img src={item.product.image_url} alt={item.product?.name} class="w-full h-full object-contain" />
                  {:else if item.product?.slug?.includes('papa')}
                    <Utensils class="w-5 h-5 text-[#FF2400] stroke-[2]" />
                  {:else if item.product?.slug?.includes('tocineta') || item.product?.slug?.includes('carne')}
                    <Flame class="w-5 h-5 text-[#FF2400] stroke-[2]" />
                  {:else if item.product?.slug?.includes('queso')}
                    <Sparkles class="w-5 h-5 text-[#FF2400] stroke-[2]" />
                  {:else}
                    <ShoppingBag class="w-5 h-5 text-ash stroke-[1.8]" />
                  {/if}
                </div>

                <!-- Nombre + extras -->
                <div class="flex-1 min-w-0">
                  <h3 class="font-display font-bold text-sm sm:text-base text-bone truncate">{item.product?.name}</h3>
                  {#if item.combo}
                    <span class="text-[10px] font-mono text-[#FF2400]">+ Combo</span>
                  {/if}
                  {#if item.extras && item.extras.length > 0}
                    <span class="text-[10px] font-mono text-ash block truncate">+ {item.extras.map(e => e.name).join(', ')}</span>
                  {/if}
                </div>

                <!-- Precio -->
                <span class="font-mono text-sm text-[#FF2400] font-bold shrink-0">{formatCOP(item.totalPrice)}</span>

                <!-- Controles ± compactos -->
                <div class="flex items-center gap-0.5 shrink-0 bg-char rounded-full p-0.5 border border-white/10">
                  <button
                    type="button"
                    onclick={() => store.updateQuantity(item.id, -1)}
                    class="w-8 h-8 rounded-full bg-char-card text-bone flex items-center justify-center hover:bg-char-card-hover active:scale-90 transition-all cursor-pointer"
                    aria-label="Disminuir cantidad"
                  >
                    <Minus class="w-3.5 h-3.5 stroke-[2.5]" />
                  </button>
                  <span class="w-5 text-center font-mono text-xs font-bold text-bone">{item.quantity}</span>
                  <button
                    type="button"
                    onclick={() => store.updateQuantity(item.id, 1)}
                    class="w-8 h-8 rounded-full bg-char-card text-bone flex items-center justify-center hover:bg-char-card-hover active:scale-90 transition-all cursor-pointer"
                    aria-label="Aumentar cantidad"
                  >
                    <Plus class="w-3.5 h-3.5 stroke-[2.5]" />
                  </button>
                </div>
              </div>
            {/each}
          </div>

          <!-- Subtotal -->
          <div class="shrink-0 pt-4">
            <div class="flex items-center justify-between px-1 mb-1">
              <span class="text-xs font-mono text-ash uppercase tracking-wider">Subtotal</span>
              <span class="font-mono text-xs font-bold text-[#8E8E93] bg-[#EDE8DC] px-3 py-1 rounded-full">
                {store.cartCount} {store.cartCount === 1 ? 'producto' : 'productos'}
              </span>
            </div>
            <span class="font-display font-extrabold text-2xl sm:text-3xl text-bone block px-1">
              {formatCOP(store.cartSubtotal)}
            </span>
          </div>
        </div>

        <!-- CTA Review -->
        <div class="shrink-0 space-y-2 pt-2 pb-2">
          {#if !store.user}
            <div class="p-3 rounded-2xl bg-char border border-amber-500/20 flex items-center gap-3">
              <User class="w-4 h-4 text-amber-400 shrink-0" />
              <span class="text-xs text-ash flex-1">Inicia sesión para continuar con tu pedido</span>
            </div>
          {/if}

          {#if formError}
            <div class="p-3 rounded-2xl bg-red-950/60 border border-red-800/40 text-xs font-bold text-red-400 text-center">
              {formError}
            </div>
          {/if}

          <button
            type="button"
            onclick={handleContinueFromReview}
            class="w-full min-h-[62px] rounded-full bg-[#FF2400] hover:bg-[#E02000] active:scale-[0.97] text-white font-display font-extrabold text-base sm:text-lg tracking-tight transition-all flex items-center justify-center gap-2.5 cursor-pointer"
          >
            {#if !store.user}
              <LogIn class="w-5 h-5 stroke-[2.5]" />
              <span>Identifícate para Continuar</span>
            {:else}
              <span>Confirmar Pedido</span>
              <ArrowRight class="w-5 h-5 stroke-[2.5]" />
            {/if}
          </button>

          <button
            type="button"
            onclick={addMoreProducts}
            class="w-full py-2 text-xs font-mono text-ash hover:text-bone transition-colors cursor-pointer text-center"
          >
            Agregar más productos
          </button>
        </div>


      <!-- ════════════════════════════════════════════════════════════════ -->
      <!-- PASO 3 — MODALIDAD: ¿DÓNDE DISFRUTAS?                         -->
      <!-- En salón: auto-finaliza sin solicitar datos                    -->
      <!-- ════════════════════════════════════════════════════════════════ -->
      {:else if cartStep === 'mode'}
        <div class="flex-1 flex flex-col justify-center gap-3.5 py-2 sm:py-4">
          {#if formError}
            <div class="p-3.5 rounded-2xl bg-red-950/60 border border-red-800/40 text-xs font-bold text-red-400 text-center mb-1">
              {formError}
            </div>
          {/if}

          <!-- Opción 1: Consumo en Salón (auto-finaliza) -->
          <button
            type="button"
            onclick={() => handleSelectMode('in_store')}
            class="w-full text-left rounded-[28px] border-2 p-5 sm:p-6 transition-all active:scale-[0.985] cursor-pointer flex items-center gap-4 group border-white/10 hover:border-white/25 bg-char-card"
          >
            <div class="w-14 h-14 sm:w-16 sm:h-16 rounded-2xl bg-[#FF2400] text-white flex items-center justify-center shrink-0 group-hover:scale-105 transition-transform">
              <UtensilsCrossed class="w-7 h-7 sm:w-8 sm:h-8 stroke-[2.2]" />
            </div>
            <div class="flex-1 min-w-0 overflow-hidden">
              <h2 class="font-display font-black text-lg sm:text-2xl text-bone truncate">Consumo en el Salón</h2>
              <span class="inline-block mt-1 px-2.5 py-0.5 rounded-full bg-emerald-950/60 border border-emerald-500/40 text-[10px] sm:text-xs font-mono font-bold text-emerald-300">
                ⚡ 1 Toque • A la Mesa
              </span>
              <p class="text-xs sm:text-sm text-ash mt-1.5 leading-relaxed line-clamp-2">
                Envío directo a cocina. Disfruta en nuestro punto y la cuenta queda lista para pagar en caja.
              </p>
            </div>
          </button>

          <!-- Opción 2: A Domicilio -->
          <button
            type="button"
            onclick={() => handleSelectMode('delivery')}
            class="w-full text-left rounded-[28px] border-2 p-5 sm:p-6 transition-all active:scale-[0.985] cursor-pointer flex items-center gap-4 group border-white/10 hover:border-white/25 bg-char-card"
          >
            <div class="w-14 h-14 sm:w-16 sm:h-16 rounded-2xl bg-char-card border border-white/10 text-white flex items-center justify-center shrink-0 group-hover:scale-105 transition-transform">
              <Bike class="w-7 h-7 sm:w-8 sm:h-8 stroke-[2.2] text-[#FF2400]" />
            </div>
            <div class="flex-1 min-w-0 overflow-hidden">
              <h2 class="font-display font-black text-lg sm:text-2xl text-bone truncate">A Domicilio</h2>
              <span class="inline-block mt-1 px-2.5 py-0.5 rounded-full bg-white/5 border border-white/10 text-[10px] sm:text-xs font-mono font-bold text-ash">
                Duitama
              </span>
              <p class="text-xs sm:text-sm text-ash mt-1.5 leading-relaxed line-clamp-2">
                Empacado hermético con sello de calor directo a tu dirección.
              </p>
            </div>
          </button>

          <!-- Opción 3: Para Llevar -->
          <button
            type="button"
            onclick={() => handleSelectMode('takeout')}
            class="w-full text-left rounded-[28px] border-2 p-5 sm:p-6 transition-all active:scale-[0.985] cursor-pointer flex items-center gap-4 group border-white/10 hover:border-white/25 bg-char-card"
          >
            <div class="w-14 h-14 sm:w-16 sm:h-16 rounded-2xl bg-char-card border border-white/10 text-white flex items-center justify-center shrink-0 group-hover:scale-105 transition-transform">
              <ShoppingBag class="w-7 h-7 sm:w-8 sm:h-8 stroke-[2.2] text-[#FF2400]" />
            </div>
            <div class="flex-1 min-w-0 overflow-hidden">
              <h2 class="font-display font-black text-lg sm:text-2xl text-bone truncate">Para Llevar</h2>
              <span class="inline-block mt-1 px-2.5 py-0.5 rounded-full bg-white/5 border border-white/10 text-[10px] sm:text-xs font-mono font-bold text-ash">
                Punto Físico
              </span>
              <p class="text-xs sm:text-sm text-ash mt-1.5 leading-relaxed line-clamp-2">
                Pasa a retirar en barra listo para llevar sin hacer filas.
              </p>
            </div>
          </button>
        </div>


      <!-- ════════════════════════════════════════════════════════════════ -->
      <!-- PASO 4 — DATOS Y PAGO (SOLO DELIVERY / TAKEOUT)               -->
      <!-- ════════════════════════════════════════════════════════════════ -->
      {:else if cartStep === 'delivery_form'}
        <div class="flex-1 min-h-0 overflow-y-auto overscroll-contain space-y-4 py-2" style="touch-action: pan-y;">

          <!-- SECCIÓN 1: ¿A QUIÉN ENTREGAMOS? -->
          <div class="bg-char-card border border-white/10 rounded-[34px] p-5 sm:p-6 space-y-3.5">
            <div class="flex items-center gap-2">
              <User class="w-5 h-5 text-[#FF2400]" />
              <h2 class="font-display font-extrabold text-lg text-bone">
                ¿A quién entregamos?
              </h2>
            </div>
            <div class="space-y-2.5">
              <input
                type="text"
                bind:value={customerName}
                placeholder="Nombre completo"
                class="w-full min-h-[58px] rounded-2xl px-5 bg-char border border-white/10 text-bone placeholder:text-ash/60 text-base font-sans focus:outline-none focus:border-[#FF2400] transition-colors"
              />
              <input
                type="tel"
                bind:value={customerPhone}
                placeholder="Número de WhatsApp (ej. 321 252 0881)"
                class="w-full min-h-[58px] rounded-2xl px-5 bg-char border border-white/10 text-bone placeholder:text-ash/60 text-base font-sans focus:outline-none focus:border-[#FF2400] transition-colors"
              />
            </div>
          </div>

          <!-- SECCIÓN 2: DIRECCIÓN (SOLO DELIVERY) O NOTAS (TAKEOUT) -->
          {#if selectedMode === 'delivery'}
            <div class="bg-char-card border border-white/10 rounded-[34px] p-5 sm:p-6 space-y-3.5">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2">
                  <MapPin class="w-5 h-5 text-[#FF2400]" />
                  <h2 class="font-display font-extrabold text-lg text-bone">
                    Dirección en Duitama
                  </h2>
                </div>
                <span class="text-xs font-mono font-bold text-[#FF2400]">Domicilio Oficial</span>
              </div>
              <div class="space-y-2.5">
                <input
                  type="text"
                  bind:value={deliveryAddress}
                  placeholder="Dirección exacta (ej. Carrera 16 # 14-32)"
                  class="w-full min-h-[58px] rounded-2xl px-5 bg-char border border-white/10 text-bone placeholder:text-ash/60 text-base font-sans focus:outline-none focus:border-[#FF2400] transition-colors"
                />
                <input
                  type="text"
                  bind:value={deliveryNeighborhood}
                  placeholder="Barrio o sector (ej. Centro, El Recreo)"
                  class="w-full min-h-[58px] rounded-2xl px-5 bg-char border border-white/10 text-bone placeholder:text-ash/60 text-base font-sans focus:outline-none focus:border-[#FF2400] transition-colors"
                />
                <input
                  type="text"
                  bind:value={deliveryNotes}
                  placeholder="Indicaciones para el repartidor"
                  class="w-full min-h-[54px] rounded-2xl px-5 bg-char border border-white/10 text-bone placeholder:text-ash/60 text-sm font-sans focus:outline-none focus:border-[#FF2400] transition-colors"
                />
                <!-- Chips Rápidos -->
                <div class="flex flex-wrap gap-1.5 pt-1">
                  {#each ['🔔 Tocar timbre', '🏢 Dejar en portería', '🛵 Llamar al llegar'] as chip}
                    <button
                      type="button"
                      onclick={() => toggleQuickNote(chip)}
                      class="px-3 py-1.5 rounded-full text-xs font-medium border transition-all cursor-pointer {deliveryNotes.includes(chip) ? 'bg-[#FF2400] text-white border-[#FF2400]' : 'bg-char text-ash border-white/10 hover:text-bone'}"
                    >
                      {chip}
                    </button>
                  {/each}
                </div>
              </div>
            </div>
          {:else if selectedMode === 'takeout'}
            <div class="bg-char-card border border-white/10 rounded-[34px] p-5 sm:p-6 space-y-3">
              <div class="flex items-center gap-2">
                <ShoppingBag class="w-5 h-5 text-[#FF2400]" />
                <h2 class="font-display font-extrabold text-lg text-bone">Notas de Retiro</h2>
              </div>
              <input
                type="text"
                bind:value={deliveryNotes}
                placeholder="Indicaciones adicionales (opcional)"
                class="w-full min-h-[54px] rounded-2xl px-5 bg-char border border-white/10 text-bone placeholder:text-ash/60 text-sm font-sans focus:outline-none focus:border-[#FF2400] transition-colors"
              />
            </div>
          {/if}

          <!-- SECCIÓN 3: MÉTODO DE PAGO -->
          <div class="bg-char-card border border-white/10 rounded-[34px] p-5 sm:p-6 space-y-3.5">
            <div class="flex items-center justify-between">
              <h2 class="font-display font-extrabold text-lg text-bone">Método de Pago</h2>
              <span class="text-xs font-mono font-bold text-ash">{paymentMethod}</span>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-3 gap-2.5">
              <button
                type="button"
                onclick={() => paymentMethod = 'Efectivo'}
                class="p-4 rounded-2xl border-2 text-left transition-all cursor-pointer flex flex-col justify-between min-h-[84px] {paymentMethod === 'Efectivo' ? 'border-[#FF2400] bg-char' : 'border-white/10 bg-char'}"
              >
                <DollarSign class="w-5 h-5 text-[#FF2400]" />
                <div>
                  <span class="font-display font-extrabold text-sm text-bone block">Efectivo</span>
                  <span class="text-[11px] text-ash">Contraentrega</span>
                </div>
              </button>
              <button
                type="button"
                onclick={() => paymentMethod = 'Nequi / Daviplata'}
                class="p-4 rounded-2xl border-2 text-left transition-all cursor-pointer flex flex-col justify-between min-h-[84px] {paymentMethod === 'Nequi / Daviplata' ? 'border-[#FF2400] bg-char' : 'border-white/10 bg-char'}"
              >
                <Phone class="w-5 h-5 text-[#FF2400]" />
                <div>
                  <span class="font-display font-extrabold text-sm text-bone block">Nequi / Davi</span>
                  <span class="text-[11px] text-ash">Transferencia</span>
                </div>
              </button>
              <button
                type="button"
                onclick={() => paymentMethod = 'Datáfono contraentrega'}
                class="p-4 rounded-2xl border-2 text-left transition-all cursor-pointer flex flex-col justify-between min-h-[84px] {paymentMethod === 'Datáfono contraentrega' ? 'border-[#FF2400] bg-char' : 'border-white/10 bg-char'}"
              >
                <CreditCard class="w-5 h-5 text-[#FF2400]" />
                <div>
                  <span class="font-display font-extrabold text-sm text-bone block">Datáfono</span>
                  <span class="text-[11px] text-ash">Tarjeta física</span>
                </div>
              </button>
            </div>

            {#if paymentMethod === 'Efectivo'}
              <div class="p-4 rounded-2xl bg-char border border-white/5 space-y-2.5">
                <span class="text-xs font-mono font-bold text-bone block">
                  ¿Con cuánto cancelas? (Para llevar tu cambio exacto)
                </span>
                <div class="grid grid-cols-3 gap-2">
                  <button type="button" onclick={() => selectQuickCash(50000)} class="py-2.5 rounded-xl bg-char-card border border-white/10 text-xs font-bold font-mono text-bone hover:border-[#FF2400] cursor-pointer transition-colors">$50.000</button>
                  <button type="button" onclick={() => selectQuickCash(100000)} class="py-2.5 rounded-xl bg-char-card border border-white/10 text-xs font-bold font-mono text-bone hover:border-[#FF2400] cursor-pointer transition-colors">$100.000</button>
                  <button type="button" onclick={() => selectQuickCash('exact')} class="py-2.5 rounded-xl bg-char-card border border-white/10 text-xs font-bold font-mono text-bone hover:border-[#FF2400] cursor-pointer transition-colors">Exacto</button>
                </div>
                <div class="flex items-center gap-2 pt-1">
                  <input
                    type="number"
                    bind:value={cashAmount}
                    placeholder="Otro valor en efectivo..."
                    class="flex-1 min-h-[48px] rounded-xl px-4 bg-char-card border border-white/10 text-xs sm:text-sm font-mono text-bone placeholder:text-ash/60 focus:outline-none focus:border-[#FF2400]"
                  />
                  {#if cashChange > 0}
                    <span class="text-xs font-mono font-bold text-emerald-400 bg-emerald-950/60 px-3 py-3 rounded-xl border border-emerald-700/30 shrink-0">
                      Vueltas: {formatCOP(cashChange)}
                    </span>
                  {/if}
                </div>
              </div>
            {:else if paymentMethod === 'Nequi / Daviplata'}
              <div class="p-4 rounded-2xl bg-char border border-white/5 flex items-center justify-between">
                <div>
                  <span class="text-[11px] font-mono uppercase text-ash block">Cuenta Nequi & Daviplata Mías</span>
                  <span class="font-display font-black text-lg text-bone">321 252 0881</span>
                </div>
                <button
                  type="button"
                  onclick={copyNequi}
                  class="min-h-[42px] px-4 rounded-xl bg-char-card border border-white/10 text-xs font-mono font-bold text-bone hover:border-[#FF2400] transition-colors cursor-pointer flex items-center gap-1.5"
                >
                  <Copy class="w-3.5 h-3.5" />
                  <span>{isCopied ? '¡Copiado!' : 'Copiar'}</span>
                </button>
              </div>
            {/if}
          </div>

          <!-- SECCIÓN 4: RESUMEN Y BOTÓN FINALIZAR -->
          <div class="bg-char-card border border-white/10 rounded-[34px] p-6 space-y-4">
            <div class="space-y-2 text-sm">
              <div class="flex justify-between text-ash">
                <span>Subtotal de platos:</span>
                <span class="font-mono font-medium text-bone">{formatCOP(store.cartSubtotal)}</span>
              </div>
              <div class="flex justify-between text-ash">
                <span>Costo de domicilio:</span>
                <span class="font-mono font-medium text-bone">
                  {currentDeliveryFee > 0 ? formatCOP(currentDeliveryFee) : 'Gratis'}
                </span>
              </div>
              <div class="border-t border-white/10 pt-2 flex items-center justify-between">
                <span class="font-display font-black text-xl text-bone">Total a Pagar</span>
                <span class="font-display font-black text-2xl sm:text-3xl text-[#FF2400]">
                  {formatCOP(currentTotal)}
                </span>
              </div>
            </div>

            {#if formError}
              <div class="p-3.5 rounded-2xl bg-red-950/60 border border-red-800/40 text-xs font-bold text-red-400 text-center">
                {formError}
              </div>
            {/if}

            <button
              type="button"
              onclick={handleFinalizeOrder}
              class="w-full min-h-[64px] rounded-full bg-[#FF2400] hover:bg-[#E02000] active:scale-[0.97] text-white font-display font-extrabold text-lg sm:text-xl tracking-tight transition-all flex items-center justify-center gap-2 cursor-pointer"
            >
              <span>Confirmar y Finalizar Pedido</span>
              <Check class="w-5 h-5 stroke-[2.5]" />
            </button>
          </div>
        </div>


      <!-- ════════════════════════════════════════════════════════════════ -->
      <!-- PASO 5 — TICKET: FACTURA DIGITAL BOARDING PASS                 -->
      <!-- ════════════════════════════════════════════════════════════════ -->
      {:else if cartStep === 'ticket' && (liveTicketOrder || createdOrder)}
        {@const ord = liveTicketOrder || createdOrder}
        {@const itemCount = (ord.items || []).reduce((s, it) => s + (it.quantity || 1), 0)}
        <div class="flex-1 flex flex-col items-center py-2 min-h-0 overflow-y-auto overscroll-contain" style="touch-action: pan-y;">
          <div class="flex-1 flex items-center justify-center w-full">
            <!-- Boarding Pass Card -->
            <div class="w-full max-w-sm mx-auto rounded-[28px] bg-white text-[#1C1C1E] overflow-visible relative" style="box-shadow: 0 8px 40px rgba(0,0,0,0.35);">
              <!-- Top Section -->
              <div class="px-6 pt-6 pb-4">
                <div class="flex items-center justify-between mb-5">
                  <span class="font-mono text-[11px] font-bold text-[#8E8E93] tracking-wide uppercase">
                    {ord.code || 'ORD-0000'}
                  </span>
                  <span class="font-mono text-[11px] font-bold text-[#8E8E93] tracking-wide uppercase">
                    MÍAS BURGERS
                  </span>
                </div>

                <!-- Route: Origin → Destination -->
                <div class="flex items-start justify-between mb-1">
                  <div class="text-left">
                    <span class="font-display font-black text-[42px] leading-none tracking-tight text-[#1C1C1E] block">
                      {#if ord.type === 'in_store'}SLN{:else if ord.type === 'delivery'}DMC{:else}TKA{/if}
                    </span>
                    <span class="font-mono text-[11px] text-[#8E8E93] mt-0.5 block">
                      {#if ord.type === 'in_store'}En el Salón{:else if ord.type === 'delivery'}Domicilio, Duitama{:else}Para Llevar{/if}
                    </span>
                  </div>
                  <div class="flex-1 flex items-center justify-center pt-3 px-4">
                    <div class="w-full flex items-center gap-1.5">
                      <div class="h-px flex-1 bg-[#D1D1D6]"></div>
                      <div class="w-2 h-2 rounded-full border-2 border-[#D1D1D6]"></div>
                      <div class="h-px flex-1 bg-[#D1D1D6]"></div>
                    </div>
                  </div>
                  <div class="text-right">
                    <span class="font-display font-black text-[42px] leading-none tracking-tight text-[#1C1C1E] block">MÍA</span>
                    <span class="font-mono text-[11px] text-[#8E8E93] mt-0.5 block">Cocina Mías</span>
                  </div>
                </div>

                <!-- Metadata Grid -->
                <div class="grid grid-cols-3 gap-x-3 mt-5 pt-4 border-t border-[#E5E5EA]">
                  <div>
                    <span class="font-mono text-[10px] text-[#8E8E93] uppercase tracking-wider block">Hora registro</span>
                    <span class="font-display font-extrabold text-sm text-[#1C1C1E] block mt-0.5">{formatTicketTime(ord.created_at)}</span>
                  </div>
                  <div>
                    <span class="font-mono text-[10px] text-[#8E8E93] uppercase tracking-wider block">Modalidad</span>
                    <span class="font-display font-extrabold text-sm text-[#1C1C1E] block mt-0.5">{ord.modeLabel || 'Salón'}</span>
                  </div>
                  <div class="text-right">
                    <span class="font-mono text-[10px] text-[#8E8E93] uppercase tracking-wider block">Total</span>
                    <span class="font-display font-extrabold text-sm text-[#1C1C1E] block mt-0.5">{formatCOP(ord.total)}</span>
                  </div>
                </div>

                <!-- Date + Status -->
                <div class="flex items-center justify-between mt-4 pt-3 border-t border-[#E5E5EA]">
                  <div>
                    <span class="font-mono text-[10px] text-[#8E8E93] uppercase tracking-wider block">Fecha</span>
                    <span class="font-display font-extrabold text-sm text-[#1C1C1E] block mt-0.5">{formatTicketDate(ord.created_at)}</span>
                  </div>
                  <div class="text-right">
                    <span class="font-mono text-[10px] text-[#8E8E93] uppercase tracking-wider block">Estado</span>
                    <span class="font-display font-extrabold text-sm block mt-0.5 {
                      ord.status === 'PAID' ? 'text-emerald-600' :
                      ord.status === 'IN_PREPARATION' ? 'text-amber-600' :
                      ord.status === 'ON_THE_WAY' ? 'text-blue-600' :
                      ord.status === 'DELIVERED' ? 'text-purple-600' :
                      'text-orange-600'
                    }">
                      {#if ord.status === 'PAID'}Cobrado{:else if ord.status === 'IN_PREPARATION'}Al Fuego{:else if ord.status === 'ON_THE_WAY'}En Camino{:else if ord.status === 'DELIVERED'}Entregado{:else}Pendiente{/if}
                    </span>
                  </div>
                </div>

                <!-- Comensal -->
                <div class="mt-4 pt-3 border-t border-[#E5E5EA]">
                  <span class="font-mono text-[10px] text-[#8E8E93] uppercase tracking-wider block">Comensal</span>
                  <span class="font-display font-black text-xl text-[#1C1C1E] block mt-0.5 truncate">
                    {ord.customer_name || 'Comensal'}
                  </span>
                </div>
              </div>

              <!-- Perforated Tear Line -->
              <div class="relative h-8 flex items-center">
                <div class="absolute -left-3.5 top-1/2 -translate-y-1/2 w-7 h-7 rounded-full bg-char z-10"></div>
                <div class="absolute -right-3.5 top-1/2 -translate-y-1/2 w-7 h-7 rounded-full bg-char z-10"></div>
                <div class="w-full border-t-2 border-dashed border-[#D1D1D6] mx-6"></div>
              </div>

              <!-- Bottom Section: Barcode -->
              <div class="px-6 pb-6 pt-2">
                <div class="flex items-end justify-center gap-px h-14 mb-2" aria-hidden="true">
                  {#each Array(42) as _, i}
                    {@const w = (i % 3 === 0) ? 3 : (i % 5 === 0) ? 1.5 : 2}
                    {@const h = (i % 7 === 0) ? 75 : (i % 4 === 0) ? 90 : 100}
                    <div
                      class="bg-[#1C1C1E] rounded-[0.5px]"
                      style="width: {w}px; height: {h}%; min-height: 10px;"
                    ></div>
                  {/each}
                </div>
                <p class="text-center font-mono text-[11px] text-[#8E8E93] tracking-[0.3em] mb-5">
                  {(ord.code || 'ORD-0000').replace('-', ' ')}
                </p>
                <div class="grid grid-cols-3 gap-3">
                  <div class="bg-[#F2F2F7] rounded-xl p-3 text-center">
                    <span class="font-mono text-[9px] text-[#8E8E93] uppercase tracking-wider block">
                      {#if ord.type === 'delivery'}Barrio{:else}Salón{/if}
                    </span>
                    <span class="font-display font-black text-2xl text-[#1C1C1E] block mt-1">
                      {#if ord.type === 'delivery'}{(ord.delivery_address || '').split('Barrio')[1]?.trim()?.charAt(0) || 'D'}{:else}1{/if}
                    </span>
                  </div>
                  <div class="bg-[#F2F2F7] rounded-xl p-3 text-center">
                    <span class="font-mono text-[9px] text-[#8E8E93] uppercase tracking-wider block">
                      {#if ord.type === 'in_store'}Modalidad{:else if ord.type === 'delivery'}Zona{:else}Turno{/if}
                    </span>
                    <span class="font-display font-black text-2xl text-[#1C1C1E] block mt-1">
                      {#if ord.type === 'in_store'}Local{:else if ord.type === 'delivery'}{(ord.delivery_address || 'D').charAt(0)}{:else}—{/if}
                    </span>
                  </div>
                  <div class="bg-[#F2F2F7] rounded-xl p-3 text-center">
                    <span class="font-mono text-[9px] text-[#8E8E93] uppercase tracking-wider block">Platos</span>
                    <span class="font-display font-black text-2xl text-[#1C1C1E] block mt-1">{itemCount}</span>
                  </div>
                </div>
              </div>
            </div>
          </div>

          <!-- Action Buttons -->
          <div class="w-full max-w-sm mx-auto mt-6 space-y-3 shrink-0">
            {#if ord.type === 'in_store'}
              <div class="w-full min-h-[64px] rounded-full bg-amber-500/20 border-2 border-amber-500/40 text-amber-300 font-display font-extrabold text-base sm:text-lg tracking-tight flex items-center justify-center gap-2.5 select-none">
                <Clock class="w-5 h-5 animate-pulse text-amber-400" />
                <span>A la espera del cobro</span>
              </div>
            {:else}
              <button
                type="button"
                onclick={dispatchWhatsApp}
                class="w-full min-h-[64px] rounded-full bg-[#FF2400] hover:bg-[#E02000] active:scale-[0.97] text-white font-display font-extrabold text-base sm:text-lg tracking-tight transition-all flex items-center justify-center gap-2.5 cursor-pointer"
              >
                <Send class="w-5 h-5" />
                <span>Enviar Pedido por WhatsApp</span>
              </button>
            {/if}

            {#if store.orderHasBurgers(liveTicketOrder || createdOrder)}
              <button
                type="button"
                onclick={goToStampsFromTicket}
                class="w-full min-h-[58px] rounded-full bg-char hover:bg-char-card text-bone border border-white/10 font-display font-extrabold text-sm sm:text-base tracking-tight transition-all flex items-center justify-center gap-2 cursor-pointer"
              >
                <Award class="w-5 h-5 text-[#FF2400]" />
                <span>Solicitar Sellos en mi Cartilla</span>
              </button>
            {/if}

            <button
              type="button"
              onclick={() => { closeOrderModal(); store.currentView = 'orders'; }}
              class="w-full min-h-[54px] rounded-full bg-char-card hover:bg-white/10 text-bone border border-white/15 font-display font-extrabold text-sm tracking-tight transition-all flex items-center justify-center gap-2 cursor-pointer"
            >
              <Receipt class="w-4 h-4 text-amber-400" />
              <span>Ver el Estado de Mis Pedidos</span>
            </button>

            <button
              type="button"
              onclick={() => { closeOrderModal(); store.currentView = 'menu'; }}
              class="w-full py-2.5 text-xs font-mono text-ash hover:text-bone transition-colors cursor-pointer"
            >
              Volver a la carta principal
            </button>
          </div>
        </div>
      {/if}

    </div>
  </div>
{/if}
