<script>
  import { tick } from 'svelte';
  import { fly, scale } from 'svelte/transition';
  import { backOut, cubicOut } from 'svelte/easing';
  import gsap from 'gsap';
  import { store } from '../lib/store.svelte.js';
  import { createOrderAPI } from '../lib/api.js';
  import { 
    X, 
    Plus, 
    Minus, 
    Trash2, 
    ShoppingBag, 
    Sparkles, 
    ArrowRight, 
    Check, 
    ChevronLeft, 
    UtensilsCrossed, 
    Bike, 
    Send, 
    Copy, 
    CheckCircle2,
    Pencil,
    MessageSquare,
    DollarSign,
    CreditCard,
    QrCode,
    Bell,
    Building2
  } from 'lucide-svelte';

  let orderBtnEl = $state(null);
  let orderCardEl = $state(null);
  let orderBackdropEl = $state(null);
  let isModalAnimating = false;
  let recentlyAddedUpsellId = $state(null);

  // Máquina de estados interna de la comanda
  let cartStep = $state('cart'); // 'cart' | 'mode' | 'delivery_form' | 'ticket'
  let selectedMode = $state('in_store'); // 'in_store' | 'takeout' | 'delivery'
  let waiterStep = $state('name'); // 'name' | 'phone' | 'address' | 'notes' | 'payment'
  let chatScrollEl = $state(null);
  let waiterFeedback = $state('');
  let tableNumber = $state('');
  let customerName = $state(store.user?.full_name || '');
  let customerPhone = $state(store.user?.phone || '');
  let deliveryAddress = $state('');
  let deliveryNeighborhood = $state('');
  let deliveryNotes = $state('');
  let paymentMethod = $state('Efectivo');
  let cashAmount = $state('');
  let deliveryFee = 6000;
  let createdOrder = $state(null);
  let isCopied = $state(false);

  let deliveryTotal = $derived(store.cartSubtotal + deliveryFee);
  let cashChange = $derived.by(() => {
    const val = parseInt(cashAmount, 10);
    if (!isNaN(val) && val >= deliveryTotal) {
      return val - deliveryTotal;
    }
    return 0;
  });

  function formatCOP(val) {
    return new Intl.NumberFormat('es-CO', {
      style: 'currency',
      currency: 'COP',
      maximumFractionDigits: 0
    }).format(val || 0);
  }

  // Catálogo de productos sugeridos para la sección "¿Algo más?"
  const upsellItems = [
    // 1. Bebidas
    {
      id: 14,
      category_id: 4,
      categoryLabel: 'Bebida',
      categorySlug: 'bebidas',
      slug: 'gaseosa',
      name: 'Coca-Cola 400ml',
      price: 4000,
      image_url: '/imagenes/bebidas/cocacola.webp'
    },
    {
      id: 15,
      category_id: 4,
      categoryLabel: 'Bebida',
      categorySlug: 'bebidas',
      slug: 'bretana',
      name: 'Bretaña en Vidrio',
      price: 5000,
      image_url: '/imagenes/bebidas/bretana-vidrio.png'
    },
    {
      id: 13,
      category_id: 4,
      categoryLabel: 'Bebida',
      categorySlug: 'bebidas',
      slug: 'agua-con-gas',
      name: 'Agua con Gas',
      price: 3000,
      image_url: '/imagenes/bebidas/agua_con_gas.webp'
    },
    // 2. Adicionales
    {
      id: 17,
      category_id: 5,
      categoryLabel: 'Adicional',
      categorySlug: 'adicionales',
      slug: 'papa-francesa',
      name: 'Papa Francesa',
      price: 8000,
      image_url: ''
    },
    {
      id: 18,
      category_id: 5,
      categoryLabel: 'Adicional',
      categorySlug: 'adicionales',
      slug: 'papa-nativa',
      name: 'Papa Nativa al Carbón',
      price: 9000,
      image_url: ''
    },
    {
      id: 19,
      category_id: 5,
      categoryLabel: 'Adicional',
      categorySlug: 'adicionales',
      slug: 'tocineta-extra',
      name: 'Tocineta Crujiente Extra',
      price: 5000,
      image_url: ''
    },
    // 3. De la Casa
    {
      id: 10,
      category_id: 3,
      categoryLabel: 'De la Casa',
      categorySlug: 'de_la_casa',
      slug: 'salchipapa',
      name: 'Salchipapa Tradicional',
      price: 22000,
      image_url: '/imagenes/de_la_casa/salchi_papa.png'
    }
  ];

  function handleAddUpsell(item) {
    store.addToCart(item, false, []);
    recentlyAddedUpsellId = item.id;
    setTimeout(() => {
      if (recentlyAddedUpsellId === item.id) {
        recentlyAddedUpsellId = null;
      }
    }, 1200);
  }

  function goToCategory(slug) {
    closeOrderModal();
    store.currentView = slug;
  }

  function handleContinueToMode() {
    if (store.cart.length === 0) return;
    cartStep = 'mode';
  }

  async function scrollToBottom() {
    await tick();
    if (chatScrollEl) {
      chatScrollEl.scrollTo({
        top: chatScrollEl.scrollHeight + 300,
        behavior: 'smooth'
      });
    }
  }

  function handleBackStep() {
    waiterFeedback = '';
    if (cartStep === 'delivery_form') {
      if (waiterStep === 'payment') {
        waiterStep = 'notes';
      } else if (waiterStep === 'notes') {
        waiterStep = 'address';
      } else if (waiterStep === 'address') {
        waiterStep = 'phone';
      } else if (waiterStep === 'phone') {
        waiterStep = 'name';
      } else {
        cartStep = 'mode';
      }
      scrollToBottom();
    } else if (cartStep === 'mode') {
      cartStep = 'cart';
    } else if (cartStep === 'ticket') {
      resetAndClose();
    }
  }

  function handleSelectMode(mode) {
    selectedMode = mode;
    waiterFeedback = '';
    if (mode === 'delivery') {
      waiterStep = 'name';
      cartStep = 'delivery_form';
      scrollToBottom();
    } else {
      createInStoreOrTakeoutOrder(mode);
    }
  }

  function submitName() {
    waiterFeedback = '';
    if (!customerName.trim()) {
      waiterFeedback = '¿Me indicas tu nombre, por favor? Así sabremos a nombre de quién va el pedido.';
      return;
    }
    if (!customerPhone.trim()) {
      waiterStep = 'phone';
    } else if (!deliveryAddress.trim() || !deliveryNeighborhood.trim()) {
      waiterStep = 'address';
    } else {
      waiterStep = 'notes';
    }
    scrollToBottom();
  }

  function submitPhone() {
    waiterFeedback = '';
    if (!customerPhone.trim()) {
      waiterFeedback = 'Indícanos tu celular o WhatsApp para mantenerte informado del despacho.';
      return;
    }
    if (!deliveryAddress.trim() || !deliveryNeighborhood.trim()) {
      waiterStep = 'address';
    } else {
      waiterStep = 'notes';
    }
    scrollToBottom();
  }

  function submitAddress() {
    waiterFeedback = '';
    if (!deliveryAddress.trim()) {
      waiterFeedback = 'Por favor escribe la dirección exacta (calle, carrera, número, torre/apto).';
      return;
    }
    if (!deliveryNeighborhood.trim()) {
      waiterFeedback = 'Indícanos también el barrio o sector para coordinar la ruta del repartidor.';
      return;
    }
    waiterStep = 'notes';
    scrollToBottom();
  }

  function toggleQuickNote(note) {
    waiterFeedback = '';
    if (note.includes('Sin notas')) {
      deliveryNotes = 'Sin notas adicionales';
      return;
    }
    if (deliveryNotes === 'Sin notas adicionales') {
      deliveryNotes = note;
      return;
    }
    if (deliveryNotes.includes(note)) {
      deliveryNotes = deliveryNotes
        .replace(note, '')
        .replace(/,\s*,/g, ',')
        .replace(/^,\s*|,\s*$/g, '')
        .trim();
    } else {
      deliveryNotes = deliveryNotes ? `${deliveryNotes}, ${note}` : note;
    }
  }

  function submitNotes() {
    waiterFeedback = '';
    waiterStep = 'payment';
    scrollToBottom();
  }

  function setCashPreset(amount) {
    cashAmount = amount.toString();
  }

  async function createInStoreOrTakeoutOrder(mode) {
    let code = 'ORD-' + Math.floor(1000 + Math.random() * 9000);
    let orderId = 'ord-' + Date.now();

    const apiItems = store.cart.map(it => ({
      product_id: it.product.id,
      quantity: it.quantity,
      customizations: {
        combo: it.combo,
        extras: it.extras || []
      }
    }));

    try {
      const apiRes = await createOrderAPI({
        order_type: 'in_store',
        items: apiItems,
        delivery_name: customerName.trim() || 'Cliente Mías',
        payment_method: 'Pago en Caja Física'
      });
      if (apiRes && apiRes.order) {
        code = apiRes.order.code;
        orderId = apiRes.order.id;
      }
    } catch (err) {
      console.warn('[Order] Operando localmente con Anillo 0', err);
    }

    const order = {
      id: orderId,
      code,
      type: mode,
      modeLabel: mode === 'in_store' ? 'Consumo en Salón' : 'Para Llevar en Barra',
      status: 'PENDING_PAYMENT',
      customer_name: customerName.trim() || 'Cliente Mías',
      table_number: tableNumber.trim() || '',
      items: [...store.cart],
      subtotal: store.cartSubtotal,
      delivery_fee: 0,
      total: store.cartSubtotal,
      payment_method: 'Pago en Caja Física',
      created_at: new Date().toISOString()
    };

    store.addOrder(order);
    createdOrder = order;
    store.clearCart();
    cartStep = 'ticket';
  }

  async function handleConfirmDelivery() {
    waiterFeedback = '';
    if (!customerName.trim()) {
      waiterStep = 'name';
      waiterFeedback = 'Falta registrar el nombre del cliente.';
      return;
    }
    if (!customerPhone.trim()) {
      waiterStep = 'phone';
      waiterFeedback = 'Falta el número de WhatsApp de contacto.';
      return;
    }
    if (!deliveryAddress.trim() || !deliveryNeighborhood.trim()) {
      waiterStep = 'address';
      waiterFeedback = 'Falta la dirección exacta o el barrio de entrega.';
      return;
    }

    let code = 'DOM-' + Math.floor(1000 + Math.random() * 9000);
    let orderId = 'dom-' + Date.now();

    const apiItems = store.cart.map(it => ({
      product_id: it.product.id,
      quantity: it.quantity,
      customizations: {
        combo: it.combo,
        extras: it.extras || []
      }
    }));

    try {
      const apiRes = await createOrderAPI({
        order_type: 'delivery',
        items: apiItems,
        delivery_name: customerName.trim(),
        delivery_phone: customerPhone.trim(),
        delivery_address: `${deliveryAddress.trim()}, Barrio ${deliveryNeighborhood.trim()}`,
        delivery_notes: deliveryNotes.trim(),
        payment_method: paymentMethod,
        cash_change_amount: parseInt(cashAmount, 10) || 0
      });
      if (apiRes && apiRes.order) {
        code = apiRes.order.code;
        orderId = apiRes.order.id;
      }
    } catch (err) {
      console.warn('[Delivery] Operando localmente con Anillo 0', err);
    }

    const order = {
      id: orderId,
      code,
      type: 'delivery',
      modeLabel: 'A Domicilio',
      status: 'IN_PREPARATION',
      customer_name: customerName.trim(),
      customer_phone: customerPhone.trim(),
      delivery_address: `${deliveryAddress.trim()}, Barrio ${deliveryNeighborhood.trim()}`,
      delivery_notes: deliveryNotes.trim(),
      items: [...store.cart],
      subtotal: store.cartSubtotal,
      delivery_fee: deliveryFee,
      total: deliveryTotal,
      payment_method: paymentMethod,
      cash_change_amount: parseInt(cashAmount, 10) || 0,
      created_at: new Date().toISOString()
    };

    store.addOrder(order);
    createdOrder = order;
    store.clearCart();
    cartStep = 'ticket';
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
    if (createdOrder.table_number) {
      msg += `*Mesa:* ${createdOrder.table_number}\n`;
    }
    if (createdOrder.delivery_notes) {
      msg += `*Notas:* ${createdOrder.delivery_notes}\n`;
    }

    msg += `\n*Detalle de Productos:*\n`;
    createdOrder.items.forEach((it) => {
      msg += `• ${it.quantity}x ${it.product.name} (${formatCOP(it.totalPrice)})\n`;
      if (it.combo) msg += `   - Con Combo (Papas + Gaseosa)\n`;
    });

    msg += `\n*Subtotal:* ${formatCOP(createdOrder.subtotal)}\n`;
    if (createdOrder.delivery_fee > 0) {
      msg += `*Costo Domicilio:* ${formatCOP(createdOrder.delivery_fee)}\n`;
    }
    msg += `*TOTAL:* ${formatCOP(createdOrder.total)}\n` +
      `*Pago:* ${createdOrder.payment_method}\n`;

    const encoded = encodeURIComponent(msg);
    const waURL = `https://api.whatsapp.com/send?phone=573212520881&text=${encoded}`;
    window.open(waURL, '_blank');
  }

  function copyOrderCode() {
    if (!createdOrder) return;
    try {
      navigator.clipboard?.writeText(createdOrder.code);
      isCopied = true;
      setTimeout(() => isCopied = false, 2000);
    } catch {}
  }

  function resetAndClose() {
    closeOrderModal();
    setTimeout(() => {
      cartStep = 'cart';
      createdOrder = null;
    }, 280);
  }

  async function openOrderModal(e) {
    if (e) e.stopPropagation();
    if (isModalAnimating) return;
    isModalAnimating = true;

    // Feedback táctil del botón de origen
    if (orderBtnEl) {
      gsap.fromTo(orderBtnEl, 
        { scale: 0.94 }, 
        { scale: 1, duration: 0.25, ease: 'power2.out', force3D: true }
      );
    }

    store.isCartOpen = true;
    await tick();

    if (!orderCardEl || !orderBackdropEl) {
      isModalAnimating = false;
      return;
    }

    let originX = '50%';
    let originY = '50%';

    if (orderBtnEl) {
      const btnRect = orderBtnEl.getBoundingClientRect();
      const cardRect = orderCardEl.getBoundingClientRect();
      const btnCenterX = btnRect.left + btnRect.width / 2;
      const btnCenterY = btnRect.top + btnRect.height / 2;
      originX = `${Math.round(btnCenterX - cardRect.left)}px`;
      originY = `${Math.round(btnCenterY - cardRect.top)}px`;
    }

    gsap.killTweensOf([orderBackdropEl, orderCardEl]);

    gsap.set(orderCardEl, {
      transformOrigin: `${originX} ${originY}`,
      scale: 0.18,
      opacity: 0,
      force3D: true
    });

    gsap.set(orderBackdropEl, {
      opacity: 0,
      force3D: true
    });

    const tl = gsap.timeline({
      onComplete: () => {
        isModalAnimating = false;
        if (orderCardEl) {
          gsap.set(orderCardEl, { clearProps: 'transform' });
        }
      }
    });

    // 1. Fondo oscuro cinematográfico
    tl.to(orderBackdropEl, {
      opacity: 1,
      duration: 0.28,
      ease: 'power2.out',
      force3D: true
    }, 0);

    // 2. Expansión fluida de la tarjeta estilo Apple Spring
    tl.to(orderCardEl, {
      scale: 1,
      opacity: 1,
      duration: 0.30,
      ease: 'power2.out',
      force3D: true
    }, 0);
  }

  function closeOrderModal(e) {
    if (e) e.stopPropagation();
    if (isModalAnimating || !store.isCartOpen) return;
    if (!orderCardEl || !orderBackdropEl) {
      store.isCartOpen = false;
      return;
    }

    isModalAnimating = true;

    let originX = '50%';
    let originY = '50%';

    if (orderBtnEl) {
      const btnRect = orderBtnEl.getBoundingClientRect();
      const cardRect = orderCardEl.getBoundingClientRect();
      const btnCenterX = btnRect.left + btnRect.width / 2;
      const btnCenterY = btnRect.top + btnRect.height / 2;
      originX = `${Math.round(btnCenterX - cardRect.left)}px`;
      originY = `${Math.round(btnCenterY - cardRect.top)}px`;
    }

    gsap.killTweensOf([orderBackdropEl, orderCardEl]);

    gsap.set(orderCardEl, {
      transformOrigin: `${originX} ${originY}`,
      force3D: true
    });

    const tl = gsap.timeline({
      onComplete: () => {
        store.isCartOpen = false;
        isModalAnimating = false;
      }
    });

    // Colapso en espejo de la tarjeta hacia el botón
    tl.to(orderCardEl, {
      scale: 0.18,
      opacity: 0,
      duration: 0.22,
      ease: 'power2.in',
      force3D: true
    }, 0);

    // Desvanecimiento del fondo oscuro
    tl.to(orderBackdropEl, {
      opacity: 0,
      duration: 0.20,
      ease: 'power2.in',
      force3D: true
    }, 0.02);
  }

  $effect(() => {
    // Si isCartOpen se activó externamente
    if (store.isCartOpen && !isModalAnimating && !orderCardEl) {
      openOrderModal();
    }
  });

  $effect(() => {
    const handleKeyDown = (e) => {
      if (e.key === 'Escape' && store.isCartOpen) {
        closeOrderModal();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  });
</script>

<!-- BOTÓN FLOTANTE "MI ORDEN" (Activo en la pantalla de inicio Carta cuando hay productos) -->
{#if store.cartCount > 0 && !store.isCartOpen && store.currentView === 'menu'}
  <div 
    in:fly={{ y: 28, duration: 340, easing: backOut }}
    out:scale={{ start: 0.85, duration: 220, opacity: 0, easing: cubicOut }}
    class="fixed bottom-6 right-6 z-40 pb-safe will-change-transform"
  >
    <button
      bind:this={orderBtnEl}
      type="button"
      onclick={openOrderModal}
      class="pill-action-btn min-h-[76px] sm:min-h-[82px] md:min-h-[88px] rounded-full bg-white text-[#3A3A3C] font-display font-extrabold text-lg sm:text-xl md:text-2xl tracking-tight border border-black/30 shadow-xl hover:shadow-2xl hover:bg-[#FAF6EE] active:scale-[0.97] transition-all duration-150 flex items-center justify-center gap-3.5 cursor-pointer select-none px-7 sm:px-9 shrink-0"
      aria-label="Ver mi orden ({store.cartCount} productos añadidos)"
    >
      <span>Mi orden</span>
      <span class="w-7 h-7 sm:w-8 sm:h-8 rounded-full bg-[#FF2400] text-white text-xs sm:text-sm font-mono font-bold flex items-center justify-center shadow-xs animate-in zoom-in duration-200 shrink-0">
        {store.cartCount}
      </span>
    </button>
  </div>
{/if}

<!-- TARJETA MODAL "MI ORDEN" (Global con flujo de comanda integrado, aislamiento de scroll y ticket de autor) -->
{#if store.isCartOpen}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div 
    bind:this={orderBackdropEl}
    data-lenis-prevent
    class="welcome-modal-backdrop fixed inset-0 h-[100dvh] z-[100] flex items-center justify-center bg-[#050507]/94 lg:backdrop-blur-sm cursor-pointer will-change-[opacity]"
    style="touch-action: pan-y;"
    onclick={closeOrderModal}
  >
    <div 
      bind:this={orderCardEl}
      data-lenis-prevent
      class="welcome-card-modal w-full h-full lg:max-w-[94vw] xl:max-w-[1440px] 2xl:max-w-[1560px] lg:h-[88vh] xl:h-[90vh] bg-[#FAF6EE] shadow-2xl lg:shadow-[0_25px_60px_-15px_rgba(0,0,0,0.6)] relative border border-black/30 rounded-[38px] sm:rounded-[48px] overflow-hidden cursor-default flex flex-col p-5 sm:p-7 md:p-9 lg:p-10 justify-between text-[#3A3A3C]"
      style="container-type: inline-size; touch-action: pan-y;"
      onclick={(e) => e.stopPropagation()}
    >
      <!-- BOTÓN VOLVER FLOTANTE (Visible en los pasos internos para poder editar la orden) -->
      {#if cartStep !== 'cart'}
        <button 
          type="button"
          onclick={handleBackStep}
          class="absolute top-5 left-5 lg:top-7 lg:left-7 w-10 h-10 rounded-full bg-white border border-black/10 hover:bg-black/5 active:scale-95 flex items-center justify-center text-char/80 hover:text-char transition-all cursor-pointer shadow-xs z-30 shrink-0 animate-in fade-in zoom-in duration-150"
          aria-label="Volver al paso anterior"
          title="Volver"
        >
          <ChevronLeft class="w-5 h-5 stroke-[2.5] -translate-x-0.5" />
        </button>
      {/if}

      <!-- BOTÓN CERRAR FLOTANTE APPLE STANDARD (top-5 right-5) -->
      <button 
        type="button"
        onclick={closeOrderModal}
        class="absolute top-5 right-5 lg:top-7 lg:right-7 w-10 h-10 rounded-full bg-black/[0.06] hover:bg-black/[0.12] active:scale-95 flex items-center justify-center text-char/80 hover:text-char transition-all cursor-pointer shadow-sm z-30 shrink-0"
        aria-label="Cerrar orden"
      >
        <X class="w-5 h-5 stroke-[2.2]" />
      </button>

      <!-- CONTENEDOR CENTRALIZADO -->
      <div class="w-full max-w-md md:max-w-lg lg:max-w-xl mx-auto flex-1 flex flex-col justify-between min-h-0 h-full">

        <!-- =======================================================
             FASE 1: LISTADO DE PRODUCTOS Y SECCIÓN "¿ALGO MÁS?"
             ======================================================= -->
        {#if cartStep === 'cart'}
          <!-- Cabecera -->
          <header class="select-none pr-14 shrink-0 flex items-center justify-between pt-1">
            <div>
              <h2 class="font-display font-black text-2xl sm:text-3xl md:text-4xl tracking-tight text-[#0C0C0E] leading-none">
                Mi Orden
              </h2>
              <span class="text-[10px] sm:text-xs font-mono font-semibold uppercase tracking-widest text-ash mt-1 block">
                Mías • Lo Íntimo Hecho para Compartir
              </span>
            </div>
          </header>

          {#if store.cart.length === 0}
            <!-- Estado Vacío -->
            <div class="flex-1 flex flex-col items-center justify-center text-center p-4 select-none my-auto">
              <div class="w-20 h-20 sm:w-24 sm:h-24 rounded-full bg-white border border-black/[0.06] flex items-center justify-center shadow-xs mb-4">
                <ShoppingBag class="w-9 h-9 text-ash/60 stroke-[1.6]" />
              </div>
              <p class="font-display font-bold text-xl sm:text-2xl text-[#0C0C0E]">
                Tu orden está vacía
              </p>
              <p class="text-xs sm:text-sm text-ash mt-1 max-w-xs">
                Explora nuestra carta gastronómica y añade tus platos favoritos para comenzar.
              </p>
              <button
                type="button"
                onclick={() => {
                  closeOrderModal();
                  store.currentView = 'hamburguesas';
                }}
                class="mt-5 px-6 py-3 rounded-full bg-[#FF2400] hover:bg-[#E02000] text-white font-display font-extrabold text-sm sm:text-base shadow-sm hover:shadow active:scale-95 transition-all cursor-pointer flex items-center gap-2"
              >
                <span>Ver Hamburguesas</span>
              </button>
            </div>
          {:else}
            <!-- Lista Desplazable de Productos con aislamiento de scroll -->
            <div 
              class="flex-1 min-h-0 overflow-y-auto my-3 sm:my-4 pr-0.5 space-y-3 sm:space-y-3.5 select-none"
              style="scrollbar-width: none; overscroll-behavior: contain; touch-action: pan-y;"
              onwheel={(e) => e.stopPropagation()}
              ontouchmove={(e) => e.stopPropagation()}
            >
              {#each store.cart as item (item.id)}
                <div class="bg-white rounded-[28px] sm:rounded-[34px] border border-black/[0.08] shadow-xs hover:shadow-sm transition-all flex items-center justify-between p-3 sm:p-4 gap-2.5 sm:gap-3.5 select-none relative">
                  <!-- Miniatura -->
                  <div class="w-16 h-16 sm:w-20 sm:h-20 shrink-0 flex items-center justify-center">
                    <img 
                      src={item.product?.image_url || '/imagenes/hamburguesas/Orignal.webp'} 
                      alt={item.product?.name || 'Producto'} 
                      class="w-full h-full object-contain pointer-events-none select-none drop-shadow-sm"
                      loading="lazy"
                    />
                  </div>

                  <!-- Detalles -->
                  <div class="flex-1 min-w-0 flex flex-col justify-center">
                    <span class="text-[10px] sm:text-xs font-mono font-bold text-ash uppercase tracking-wider">
                      {item.product?.category_id === 1 ? 'HAMBURGUESA' : item.product?.category_id === 2 ? 'PERRO CALIENTE' : 'DE AUTOR'}
                    </span>
                    <h3 class="font-display font-extrabold text-base sm:text-lg text-[#0C0C0E] tracking-tight truncate mt-0.5">
                      {item.product?.name || 'Producto'}
                    </h3>
                    {#if item.combo}
                      <span class="text-[10px] sm:text-xs font-mono text-ash/80 mt-0.5">
                        + Combo (Papas y Bebida)
                      </span>
                    {/if}
                    <div class="font-display font-black text-base sm:text-lg text-[#FF2400] mt-1 tracking-tight">
                      {formatCOP(item.totalPrice)}
                    </div>
                  </div>

                  <!-- Controles -->
                  <div class="flex items-center gap-2 sm:gap-2.5 shrink-0 ml-auto">
                    <div class="flex items-center gap-1 sm:gap-1.5 bg-[#FAF6EE] p-1 sm:p-1.5 rounded-full border border-black/[0.08] shadow-2xs">
                      <button
                        type="button"
                        onclick={() => store.updateQuantity(item.id, -1)}
                        class="w-8 h-8 sm:w-9 sm:h-9 rounded-full bg-white hover:bg-black/5 active:scale-90 border border-black/10 flex items-center justify-center text-[#0C0C0E] transition-all cursor-pointer shadow-2xs"
                        aria-label="Disminuir cantidad"
                      >
                        <Minus class="w-3.5 h-3.5 sm:w-4 sm:h-4 stroke-[2.5]" />
                      </button>

                      <span class="font-display font-extrabold text-sm sm:text-base text-[#0C0C0E] min-w-[20px] sm:min-w-[24px] text-center select-none px-1">
                        {item.quantity}
                      </span>

                      <button
                        type="button"
                        onclick={() => store.updateQuantity(item.id, 1)}
                        class="w-8 h-8 sm:w-9 sm:h-9 rounded-full bg-[#FF2400] hover:bg-[#E02000] active:scale-90 text-white flex items-center justify-center transition-all cursor-pointer shadow-xs"
                        aria-label="Aumentar cantidad"
                      >
                        <Plus class="w-3.5 h-3.5 sm:w-4 sm:h-4 stroke-[2.5]" />
                      </button>
                    </div>

                    <button
                      type="button"
                      onclick={() => store.updateQuantity(item.id, -item.quantity)}
                      class="w-10 sm:w-11 h-14 sm:h-16 rounded-[18px] sm:rounded-[22px] bg-[#FF2400]/10 hover:bg-[#FF2400]/20 active:scale-95 text-[#FF2400] flex items-center justify-center transition-all cursor-pointer shadow-2xs shrink-0 border border-[#FF2400]/15"
                      aria-label="Eliminar producto"
                      title="Eliminar producto"
                    >
                      <Trash2 class="w-4 h-4 sm:w-5 sm:h-5 stroke-[2] text-[#FF2400]" />
                    </button>
                  </div>
                </div>
              {/each}

              <!-- SECCIÓN "¿ALGO MÁS?" EN CUADRADOS REDONDEADOS Y NAVEGACIÓN DIRECTA -->
              <div class="mt-4 pt-4 border-t border-black/[0.08] select-none">
                <div class="flex items-center justify-between mb-2.5 px-0.5">
                  <div>
                    <h3 class="font-display font-black text-lg sm:text-xl text-[#0C0C0E] tracking-tight">
                      ¿Algo más?
                    </h3>
                    <span class="text-[10px] sm:text-xs font-mono text-ash font-medium">
                      Bebidas frías, adicionales y de la casa
                    </span>
                  </div>

                  <!-- Enlaces directos a las vistas completas -->
                  <div class="flex items-center gap-1 sm:gap-1.5">
                    <button
                      type="button"
                      onclick={() => goToCategory('bebidas')}
                      class="px-2.5 py-1 rounded-full bg-white border border-black/15 hover:border-[#FF2400] text-[10px] sm:text-[11px] font-mono font-bold text-[#3A3A3C] hover:text-[#FF2400] transition-all cursor-pointer shadow-2xs hover:shadow-xs active:scale-95"
                      title="Ver toda la carta de Bebidas"
                    >
                      Bebidas →
                    </button>
                    <button
                      type="button"
                      onclick={() => goToCategory('adicionales')}
                      class="px-2.5 py-1 rounded-full bg-white border border-black/15 hover:border-[#FF2400] text-[10px] sm:text-[11px] font-mono font-bold text-[#3A3A3C] hover:text-[#FF2400] transition-all cursor-pointer shadow-2xs hover:shadow-xs active:scale-95"
                      title="Ver toda la carta de Adicionales"
                    >
                      Adicionales →
                    </button>
                    <button
                      type="button"
                      onclick={() => goToCategory('de_la_casa')}
                      class="hidden sm:inline-flex px-2.5 py-1 rounded-full bg-white border border-black/15 hover:border-[#FF2400] text-[10px] sm:text-[11px] font-mono font-bold text-[#3A3A3C] hover:text-[#FF2400] transition-all cursor-pointer shadow-2xs hover:shadow-xs active:scale-95"
                      title="Ver toda la carta de De la Casa"
                    >
                      De la Casa →
                    </button>
                  </div>
                </div>

                <!-- Fila de Cuadrados Redondeados Desplazable -->
                <div 
                  class="flex gap-2.5 sm:gap-3 overflow-x-auto py-1 px-0.5 -mx-0.5" 
                  style="scrollbar-width: none; overscroll-behavior: contain;"
                >
                  {#each upsellItems as upsell (upsell.id)}
                    {@const isJustAdded = recentlyAddedUpsellId === upsell.id}
                    <div class="w-32 sm:w-36 shrink-0 bg-white rounded-[24px] sm:rounded-[28px] border border-black/[0.08] p-2.5 sm:p-3 flex flex-col justify-between shadow-2xs hover:shadow-xs transition-all select-none group">
                      
                      <div class="h-16 sm:h-20 w-full flex items-center justify-center overflow-hidden mb-1">
                        {#if upsell.image_url}
                          <img 
                            src={upsell.image_url} 
                            alt={upsell.name} 
                            class="h-full w-auto max-w-full object-contain pointer-events-none drop-shadow-2xs group-hover:scale-105 transition-transform"
                            loading="lazy"
                          />
                        {:else}
                          <div class="w-11 h-11 rounded-full bg-[#FAF6EE] flex items-center justify-center text-[#FF2400] border border-black/[0.04]">
                            <Sparkles class="w-5 h-5 stroke-[2]" />
                          </div>
                        {/if}
                      </div>

                      <div class="min-w-0">
                        <span class="text-[9px] font-mono font-bold text-ash uppercase tracking-wider block truncate">
                          {upsell.categoryLabel}
                        </span>
                        <h4 class="font-display font-extrabold text-xs sm:text-sm text-[#0C0C0E] tracking-tight truncate mt-0.5" title={upsell.name}>
                          {upsell.name}
                        </h4>
                        <span class="font-display font-black text-xs sm:text-sm text-[#FF2400] tracking-tight block mt-0.5">
                          {formatCOP(upsell.price)}
                        </span>
                      </div>

                      <button
                        type="button"
                        onclick={() => handleAddUpsell(upsell)}
                        class="mt-2.5 w-full py-1.5 sm:py-2 rounded-full {isJustAdded ? 'bg-emerald-600 text-white' : 'bg-[#FF2400] hover:bg-[#E02000] text-white'} active:scale-95 font-display font-extrabold text-[11px] sm:text-xs flex items-center justify-center gap-1 transition-all cursor-pointer shadow-2xs"
                        aria-label="Añadir {upsell.name} a la orden"
                      >
                        {#if isJustAdded}
                          <Check class="w-3 h-3 stroke-[3]" />
                          <span>Añadido</span>
                        {:else}
                          <Plus class="w-3 h-3 stroke-[3]" />
                          <span>Añadir</span>
                        {/if}
                      </button>
                    </div>
                  {/each}
                </div>
              </div>
            </div>

            <!-- Footer con Totales y Botón "ENVIAR Y PAGAR" -->
            <footer class="pt-3 sm:pt-4 border-t border-black/[0.08] shrink-0">
              <div class="flex items-center justify-between">
                <span class="font-mono text-xs sm:text-sm font-bold tracking-widest text-ash uppercase">
                  SUBTOTAL
                </span>
                <span class="font-display font-extrabold text-base sm:text-xl text-[#0C0C0E]">
                  {formatCOP(store.cartSubtotal)}
                </span>
              </div>

              <div class="border-t border-black/[0.08] my-2.5 sm:my-3"></div>

              <div class="flex items-center justify-between">
                <span class="font-display font-extrabold text-2xl sm:text-3xl text-[#0C0C0E]">
                  Total
                </span>
                <span class="font-display font-extrabold text-2xl sm:text-3xl text-[#FF2400]">
                  {formatCOP(store.cartSubtotal)}
                </span>
              </div>

              <button
                type="button"
                onclick={handleContinueToMode}
                class="w-full mt-4 sm:mt-5 min-h-[76px] sm:min-h-[82px] md:min-h-[88px] rounded-full bg-[#FF2400] hover:bg-[#E02000] active:scale-[0.97] text-white font-display font-extrabold text-lg sm:text-xl md:text-2xl tracking-tight uppercase shadow-md hover:shadow-lg transition-all flex items-center justify-center cursor-pointer select-none px-6"
              >
                ENVIAR Y PAGAR
              </button>
            </footer>
          {/if}

        <!-- =======================================================
             FASE 2: SELECCIÓN DE MODALIDAD DE CONSUMO (LIMPIEZA DE CONTENIDO EN LA MISMA CARD)
             ======================================================= -->
        {:else if cartStep === 'mode'}
          <header class="select-none pl-12 pr-14 shrink-0 flex items-center justify-between pt-1">
            <div>
              <h2 class="font-display font-black text-2xl sm:text-3xl md:text-4xl tracking-tight text-[#0C0C0E] leading-none">
                ¿Dónde disfrutas?
              </h2>
              <span class="text-[10px] sm:text-xs font-mono font-semibold uppercase tracking-widest text-ash mt-1 block">
                Selecciona la modalidad de tu comanda
              </span>
            </div>
          </header>

          <div 
            class="flex-1 min-h-0 overflow-y-auto my-auto py-4 space-y-3.5 select-none"
            style="scrollbar-width: none; overscroll-behavior: contain;"
            onwheel={(e) => e.stopPropagation()}
            ontouchmove={(e) => e.stopPropagation()}
          >
            <!-- Opción 1: Consumo en Salón / Mesa -->
            <button
              type="button"
              onclick={() => handleSelectMode('in_store')}
              class="w-full text-left bg-white rounded-[28px] sm:rounded-[34px] border border-black/[0.08] hover:border-[#FF2400]/40 p-5 sm:p-6 shadow-xs hover:shadow-md transition-all active:scale-[0.985] cursor-pointer flex items-start gap-4 group"
            >
              <div class="w-12 h-12 rounded-2xl bg-[#FF2400]/10 text-[#FF2400] flex items-center justify-center shrink-0 mt-0.5 group-hover:scale-105 transition-transform">
                <UtensilsCrossed class="w-6 h-6 stroke-[2.2]" />
              </div>
              <div class="flex-1 min-w-0">
                <div class="flex items-center justify-between">
                  <h3 class="font-display font-extrabold text-lg sm:text-xl text-[#0C0C0E] group-hover:text-[#FF2400] transition-colors">
                    Consumo en el Punto / Salón
                  </h3>
                  <ArrowRight class="w-4 h-4 text-ash group-hover:text-[#FF2400] group-hover:translate-x-1 transition-all shrink-0" />
                </div>
                <p class="text-xs text-ash mt-1 leading-relaxed">
                  Pagas en caja física, recibes tu identificador de madera y te llevamos los platos a la mesa.
                </p>
              </div>
            </button>

            <!-- Opción 2: Para Llevar -->
            <button
              type="button"
              onclick={() => handleSelectMode('takeout')}
              class="w-full text-left bg-white rounded-[28px] sm:rounded-[34px] border border-black/[0.08] hover:border-[#FF2400]/40 p-5 sm:p-6 shadow-xs hover:shadow-md transition-all active:scale-[0.985] cursor-pointer flex items-start gap-4 group"
            >
              <div class="w-12 h-12 rounded-2xl bg-[#0C0C0E]/5 text-[#0C0C0E] flex items-center justify-center shrink-0 mt-0.5 group-hover:scale-105 transition-transform">
                <ShoppingBag class="w-6 h-6 stroke-[2.2]" />
              </div>
              <div class="flex-1 min-w-0">
                <div class="flex items-center justify-between">
                  <h3 class="font-display font-extrabold text-lg sm:text-xl text-[#0C0C0E] group-hover:text-[#FF2400] transition-colors">
                    Para Llevar (Take Away)
                  </h3>
                  <ArrowRight class="w-4 h-4 text-ash group-hover:text-[#FF2400] group-hover:translate-x-1 transition-all shrink-0" />
                </div>
                <p class="text-xs text-ash mt-1 leading-relaxed">
                  Empacado térmico hermético con sello de seguridad. Recoges directo en la barra.
                </p>
              </div>
            </button>

            <!-- Opción 3: A Domicilio -->
            <button
              type="button"
              onclick={() => handleSelectMode('delivery')}
              class="w-full text-left bg-white rounded-[28px] sm:rounded-[34px] border border-black/[0.08] hover:border-[#FF2400]/40 p-5 sm:p-6 shadow-xs hover:shadow-md transition-all active:scale-[0.985] cursor-pointer flex items-start gap-4 group"
            >
              <div class="w-12 h-12 rounded-2xl bg-[#FF2400] text-white flex items-center justify-center shrink-0 mt-0.5 group-hover:scale-105 transition-transform shadow-xs">
                <Bike class="w-6 h-6 stroke-[2.2]" />
              </div>
              <div class="flex-1 min-w-0">
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-2">
                    <h3 class="font-display font-extrabold text-lg sm:text-xl text-[#0C0C0E] group-hover:text-[#FF2400] transition-colors">
                      A Domicilio
                    </h3>
                    <span class="text-[10px] font-mono font-bold text-[#FF2400] bg-[#FF2400]/10 px-2 py-0.5 rounded-full">
                      +$6.000
                    </span>
                  </div>
                  <ArrowRight class="w-4 h-4 text-ash group-hover:text-[#FF2400] group-hover:translate-x-1 transition-all shrink-0" />
                </div>
                <p class="text-xs text-ash mt-1 leading-relaxed">
                  Coordinado directamente a tu puerta y despachado por WhatsApp oficial al restaurante.
                </p>
              </div>
            </button>
          </div>

          <footer class="pt-3 border-t border-black/[0.08] text-center text-xs font-mono text-ash shrink-0">
            <span>Puedes volver a tu orden en cualquier momento tocando la flecha superior</span>
          </footer>

        <!-- =======================================================
             FASE 2.5: EXPERIENCIA CONVERSACIONAL CON EL MESERO DE MÍAS
             ======================================================= -->
        {:else if cartStep === 'delivery_form'}
          <header class="select-none pl-12 pr-12 shrink-0 flex items-center justify-between pt-1 border-b border-black/[0.06] pb-3">
            <div class="flex items-center gap-3">
              <!-- Avatar del Mesero con indicador online -->
              <div class="relative shrink-0">
                <div class="w-10 h-10 sm:w-11 sm:h-11 rounded-full bg-[#18181b] text-white flex items-center justify-center shadow-xs border-2 border-white">
                  <UtensilsCrossed class="w-4 h-4 sm:w-5 sm:h-5 stroke-[2.2] text-[#FAF6EE]" />
                </div>
                <span class="absolute bottom-0 right-0 w-3 h-3 rounded-full bg-emerald-500 border-2 border-white ring-1 ring-emerald-400/50"></span>
              </div>
              <div class="min-w-0">
                <h2 class="font-display font-black text-base sm:text-lg tracking-tight text-[#0C0C0E] leading-tight truncate">
                  El Mesero de Mías
                </h2>
                <span class="text-[10px] sm:text-xs font-mono text-ash flex items-center gap-1.5 font-medium">
                  <span class="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse"></span>
                  Tomando tu orden a domicilio
                </span>
              </div>
            </div>
          </header>

          <!-- CONTENEDOR CONVERSACIONAL CON AUTO-SCROLL Y AISLAMIENTO DE SCROLL -->
          <div 
            bind:this={chatScrollEl}
            class="flex-1 min-h-0 overflow-y-auto my-2 py-3 space-y-3.5 pr-1 select-none text-left"
            style="scrollbar-width: none; overscroll-behavior: contain; touch-action: pan-y;"
            onwheel={(e) => e.stopPropagation()}
            ontouchmove={(e) => e.stopPropagation()}
          >
            <!-- 1. Saludo Inicial del Mesero -->
            <div class="flex items-start gap-2.5 max-w-[92%] sm:max-w-[85%]">
              <div class="w-7 h-7 sm:w-8 sm:h-8 rounded-full bg-[#18181b] text-white flex items-center justify-center shrink-0 mt-0.5 shadow-2xs">
                <UtensilsCrossed class="w-3.5 h-3.5 text-[#FAF6EE]" />
              </div>
              <div class="bg-white rounded-[22px] rounded-tl-sm p-3.5 sm:p-4 border border-black/[0.08] shadow-2xs text-[#0C0C0E]">
                <p class="text-xs sm:text-sm leading-relaxed">
                  ¡Qué delicia de comanda has armado! Me encargo personalmente de tu despacho para que todo llegue perfecto y caliente.
                </p>
                <p class="text-xs sm:text-sm font-bold mt-1.5 text-[#0C0C0E]">
                  ¿A nombre de quién tomamos el pedido hoy?
                </p>
              </div>
            </div>

            <!-- Respuesta 1 del Cliente (Nombre) -->
            {#if customerName.trim() && waiterStep !== 'name'}
              <div class="flex justify-end">
                <button 
                  type="button"
                  onclick={() => { waiterStep = 'name'; scrollToBottom(); }}
                  class="group max-w-[85%] bg-[#FF2400] text-white rounded-[22px] rounded-tr-sm px-4 py-2.5 shadow-2xs flex items-center gap-2.5 hover:bg-[#E02000] active:scale-98 transition-all cursor-pointer text-left"
                  title="Tocar para editar nombre"
                >
                  <div class="min-w-0">
                    <span class="text-[10px] font-mono text-white/70 block uppercase tracking-wider">Cliente</span>
                    <span class="text-xs sm:text-sm font-bold truncate block">{customerName}</span>
                  </div>
                  <Pencil class="w-3.5 h-3.5 text-white/70 group-hover:text-white shrink-0 ml-1" />
                </button>
              </div>
            {/if}

            <!-- 2. Pregunta del Mesero: Teléfono WhatsApp -->
            {#if (customerName.trim() && waiterStep !== 'name') || waiterStep === 'phone'}
              <div class="flex items-start gap-2.5 max-w-[92%] sm:max-w-[85%]">
                <div class="w-7 h-7 sm:w-8 sm:h-8 rounded-full bg-[#18181b] text-white flex items-center justify-center shrink-0 mt-0.5 shadow-2xs">
                  <UtensilsCrossed class="w-3.5 h-3.5 text-[#FAF6EE]" />
                </div>
                <div class="bg-white rounded-[22px] rounded-tl-sm p-3.5 sm:p-4 border border-black/[0.08] shadow-2xs text-[#0C0C0E]">
                  <p class="text-xs sm:text-sm leading-relaxed">
                    Mucho gusto, <strong class="text-[#FF2400] font-bold">{customerName}</strong>. 
                    ¿A qué número de WhatsApp te notificamos cuando la carne esté al carbón y el repartidor vaya en camino?
                  </p>
                </div>
              </div>
            {/if}

            <!-- Respuesta 2 del Cliente (Teléfono) -->
            {#if customerPhone.trim() && waiterStep !== 'name' && waiterStep !== 'phone'}
              <div class="flex justify-end">
                <button 
                  type="button"
                  onclick={() => { waiterStep = 'phone'; scrollToBottom(); }}
                  class="group max-w-[85%] bg-[#FF2400] text-white rounded-[22px] rounded-tr-sm px-4 py-2.5 shadow-2xs flex items-center gap-2.5 hover:bg-[#E02000] active:scale-98 transition-all cursor-pointer text-left"
                  title="Tocar para editar celular"
                >
                  <div class="min-w-0">
                    <span class="text-[10px] font-mono text-white/70 block uppercase tracking-wider">WhatsApp</span>
                    <span class="text-xs sm:text-sm font-mono font-bold truncate block">{customerPhone}</span>
                  </div>
                  <Pencil class="w-3.5 h-3.5 text-white/70 group-hover:text-white shrink-0 ml-1" />
                </button>
              </div>
            {/if}

            <!-- 3. Pregunta del Mesero: Dirección y Barrio -->
            {#if (customerPhone.trim() && waiterStep !== 'name' && waiterStep !== 'phone') || waiterStep === 'address'}
              <div class="flex items-start gap-2.5 max-w-[92%] sm:max-w-[85%]">
                <div class="w-7 h-7 sm:w-8 sm:h-8 rounded-full bg-[#18181b] text-white flex items-center justify-center shrink-0 mt-0.5 shadow-2xs">
                  <UtensilsCrossed class="w-3.5 h-3.5 text-[#FAF6EE]" />
                </div>
                <div class="bg-white rounded-[22px] rounded-tl-sm p-3.5 sm:p-4 border border-black/[0.08] shadow-2xs text-[#0C0C0E]">
                  <p class="text-xs sm:text-sm leading-relaxed">
                    Excelente. Ahora cuéntame, ¿a qué dirección exacta y barrio te llevamos este banquete?
                  </p>
                </div>
              </div>
            {/if}

            <!-- Respuesta 3 del Cliente (Dirección) -->
            {#if deliveryAddress.trim() && deliveryNeighborhood.trim() && waiterStep !== 'name' && waiterStep !== 'phone' && waiterStep !== 'address'}
              <div class="flex justify-end">
                <button 
                  type="button"
                  onclick={() => { waiterStep = 'address'; scrollToBottom(); }}
                  class="group max-w-[85%] bg-[#FF2400] text-white rounded-[22px] rounded-tr-sm px-4 py-2.5 shadow-2xs flex items-center gap-2.5 hover:bg-[#E02000] active:scale-98 transition-all cursor-pointer text-left"
                  title="Tocar para editar dirección"
                >
                  <div class="min-w-0">
                    <span class="text-[10px] font-mono text-white/70 block uppercase tracking-wider">Lugar de Entrega</span>
                    <span class="text-xs sm:text-sm font-bold truncate block">{deliveryAddress}</span>
                    <span class="text-[11px] text-white/80 block">Barrio {deliveryNeighborhood}</span>
                  </div>
                  <Pencil class="w-3.5 h-3.5 text-white/70 group-hover:text-white shrink-0 ml-1" />
                </button>
              </div>
            {/if}

            <!-- 4. Pregunta del Mesero: Instrucciones & Notas -->
            {#if (deliveryAddress.trim() && deliveryNeighborhood.trim() && waiterStep !== 'name' && waiterStep !== 'phone' && waiterStep !== 'address') || waiterStep === 'notes'}
              <div class="flex items-start gap-2.5 max-w-[92%] sm:max-w-[85%]">
                <div class="w-7 h-7 sm:w-8 sm:h-8 rounded-full bg-[#18181b] text-white flex items-center justify-center shrink-0 mt-0.5 shadow-2xs">
                  <UtensilsCrossed class="w-3.5 h-3.5 text-[#FAF6EE]" />
                </div>
                <div class="bg-white rounded-[22px] rounded-tl-sm p-3.5 sm:p-4 border border-black/[0.08] shadow-2xs text-[#0C0C0E]">
                  <p class="text-xs sm:text-sm leading-relaxed">
                    Anotada la ruta. ¿Deseas dejarnos alguna indicación para la cocina o para cuando el repartidor llegue a tu puerta?
                  </p>
                </div>
              </div>
            {/if}

            <!-- Respuesta 4 del Cliente (Notas) -->
            {#if waiterStep === 'payment'}
              <div class="flex justify-end">
                <button 
                  type="button"
                  onclick={() => { waiterStep = 'notes'; scrollToBottom(); }}
                  class="group max-w-[85%] bg-[#FF2400] text-white rounded-[22px] rounded-tr-sm px-4 py-2.5 shadow-2xs flex items-center gap-2.5 hover:bg-[#E02000] active:scale-98 transition-all cursor-pointer text-left"
                  title="Tocar para editar notas"
                >
                  <div class="min-w-0">
                    <span class="text-[10px] font-mono text-white/70 block uppercase tracking-wider">Indicaciones</span>
                    <span class="text-xs sm:text-sm font-medium block">{deliveryNotes.trim() || 'Sin notas adicionales'}</span>
                  </div>
                  <Pencil class="w-3.5 h-3.5 text-white/70 group-hover:text-white shrink-0 ml-1" />
                </button>
              </div>
            {/if}

            <!-- 5. Mensaje del Mesero: Resumen y Pago -->
            {#if waiterStep === 'payment'}
              <div class="flex items-start gap-2.5 max-w-[92%] sm:max-w-[85%]">
                <div class="w-7 h-7 sm:w-8 sm:h-8 rounded-full bg-[#18181b] text-white flex items-center justify-center shrink-0 mt-0.5 shadow-2xs">
                  <UtensilsCrossed class="w-3.5 h-3.5 text-[#FAF6EE]" />
                </div>
                <div class="bg-white rounded-[22px] rounded-tl-sm p-3.5 sm:p-4 border border-black/[0.08] shadow-2xs text-[#0C0C0E]">
                  <p class="text-xs sm:text-sm leading-relaxed">
                    ¡Casi listos para encender fogones! Tu subtotal es de <strong>{formatCOP(store.cartSubtotal)}</strong> + domicilio oficial <strong>{formatCOP(deliveryFee)}</strong> para un total de:
                  </p>
                  <p class="font-display font-black text-xl sm:text-2xl text-[#FF2400] mt-1">
                    {formatCOP(deliveryTotal)}
                  </p>
                  <p class="text-xs sm:text-sm text-ash mt-1">
                    ¿Cómo prefieres cancelar tu comanda?
                  </p>
                </div>
              </div>
            {/if}
          </div>

          <!-- CONTROLES INTERACTIVOS SEGÚN EL PASO ACTUAL -->
          <footer class="pt-2 border-t border-black/[0.08] shrink-0 select-none">
            {#if waiterStep === 'name'}
              <div class="bg-white rounded-[26px] p-3 sm:p-3.5 border border-black/[0.08] shadow-xs space-y-2">
                {#if store.user?.full_name && !customerName}
                  <div class="flex items-center gap-1.5 pb-1">
                    <span class="text-[10px] font-mono text-ash">Sugerencia:</span>
                    <button 
                      type="button" 
                      onclick={() => { customerName = store.user.full_name; submitName(); }}
                      class="px-2.5 py-1 rounded-full bg-[#FAF6EE] border border-black/10 hover:border-[#FF2400] text-xs font-medium text-[#0C0C0E] transition-all cursor-pointer"
                    >
                      Soy {store.user.full_name}
                    </button>
                  </div>
                {/if}
                <div class="flex items-center gap-2">
                  <div class="relative flex-1">
                    <input 
                      type="text" 
                      bind:value={customerName}
                      placeholder="Escribe tu nombre completo..."
                      class="w-full bg-[#FAF6EE] border border-black/10 rounded-2xl px-4 py-3 text-xs sm:text-sm font-sans text-[#0C0C0E] placeholder:text-ash/70 focus:outline-none focus:border-[#FF2400] transition-colors"
                      onkeydown={(e) => { if (e.key === 'Enter') submitName(); }}
                    />
                  </div>
                  <button 
                    type="button" 
                    onclick={submitName}
                    class="h-11 px-5 rounded-2xl bg-[#FF2400] hover:bg-[#E02000] active:scale-95 text-white font-display font-bold text-xs sm:text-sm flex items-center justify-center gap-1.5 transition-all cursor-pointer shadow-2xs shrink-0"
                  >
                    <span>Continuar</span>
                    <ArrowRight class="w-3.5 h-3.5" />
                  </button>
                </div>
                {#if waiterFeedback}
                  <span class="text-[11px] text-[#FF2400] font-medium block px-1">{waiterFeedback}</span>
                {/if}
              </div>

            {:else if waiterStep === 'phone'}
              <div class="bg-white rounded-[26px] p-3 sm:p-3.5 border border-black/[0.08] shadow-xs space-y-2">
                {#if store.user?.phone && !customerPhone}
                  <div class="flex items-center gap-1.5 pb-1">
                    <span class="text-[10px] font-mono text-ash">Sugerencia:</span>
                    <button 
                      type="button" 
                      onclick={() => { customerPhone = store.user.phone; submitPhone(); }}
                      class="px-2.5 py-1 rounded-full bg-[#FAF6EE] border border-black/10 hover:border-[#FF2400] text-xs font-mono font-medium text-[#0C0C0E] transition-all cursor-pointer"
                    >
                      {store.user.phone}
                    </button>
                  </div>
                {/if}
                <div class="flex items-center gap-2">
                  <div class="relative flex-1">
                    <input 
                      type="tel" 
                      bind:value={customerPhone}
                      placeholder="Ej: 300 123 4567"
                      class="w-full bg-[#FAF6EE] border border-black/10 rounded-2xl px-4 py-3 text-xs sm:text-sm font-mono text-[#0C0C0E] placeholder:text-ash/70 focus:outline-none focus:border-[#FF2400] transition-colors"
                      onkeydown={(e) => { if (e.key === 'Enter') submitPhone(); }}
                    />
                  </div>
                  <button 
                    type="button" 
                    onclick={submitPhone}
                    class="h-11 px-5 rounded-2xl bg-[#FF2400] hover:bg-[#E02000] active:scale-95 text-white font-display font-bold text-xs sm:text-sm flex items-center justify-center gap-1.5 transition-all cursor-pointer shadow-2xs shrink-0"
                  >
                    <span>Listo</span>
                    <ArrowRight class="w-3.5 h-3.5" />
                  </button>
                </div>
                {#if waiterFeedback}
                  <span class="text-[11px] text-[#FF2400] font-medium block px-1">{waiterFeedback}</span>
                {/if}
              </div>

            {:else if waiterStep === 'address'}
              <div class="bg-white rounded-[26px] p-3 sm:p-3.5 border border-black/[0.08] shadow-xs space-y-2">
                <div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
                  <input 
                    type="text" 
                    bind:value={deliveryAddress}
                    placeholder="Dirección exacta (Calle, Cra, Apto)..."
                    class="w-full bg-[#FAF6EE] border border-black/10 rounded-2xl px-4 py-2.5 text-xs sm:text-sm font-sans text-[#0C0C0E] placeholder:text-ash/70 focus:outline-none focus:border-[#FF2400] transition-colors"
                    onkeydown={(e) => { if (e.key === 'Enter') submitAddress(); }}
                  />
                  <input 
                    type="text" 
                    bind:value={deliveryNeighborhood}
                    placeholder="Barrio o sector..."
                    class="w-full bg-[#FAF6EE] border border-black/10 rounded-2xl px-4 py-2.5 text-xs sm:text-sm font-sans text-[#0C0C0E] placeholder:text-ash/70 focus:outline-none focus:border-[#FF2400] transition-colors"
                    onkeydown={(e) => { if (e.key === 'Enter') submitAddress(); }}
                  />
                </div>
                <button 
                  type="button" 
                  onclick={submitAddress}
                  class="w-full h-11 rounded-2xl bg-[#FF2400] hover:bg-[#E02000] active:scale-95 text-white font-display font-bold text-xs sm:text-sm flex items-center justify-center gap-1.5 transition-all cursor-pointer shadow-2xs"
                >
                  <span>Confirmar Dirección</span>
                  <ArrowRight class="w-3.5 h-3.5" />
                </button>
                {#if waiterFeedback}
                  <span class="text-[11px] text-[#FF2400] font-medium block px-1">{waiterFeedback}</span>
                {/if}
              </div>

            {:else if waiterStep === 'notes'}
              <div class="bg-white rounded-[26px] p-3 sm:p-3.5 border border-black/[0.08] shadow-xs space-y-2.5">
                <span class="text-[10px] font-mono font-bold uppercase text-ash block">Opciones rápidas con un toque:</span>
                <div class="flex flex-wrap gap-1.5">
                  {#each [
                    '🔔 Tocar el timbre',
                    '🏢 Dejar en portería',
                    '🛵 Llamar al llegar',
                    '🧅 Sin cebolla',
                    '🥫 Salsas adicionales',
                    '✨ Sin notas adicionales'
                  ] as chip}
                    {@const isSelected = deliveryNotes.includes(chip.replace(/^[^\s]+\s/, '')) || (chip.includes('Sin notas') && deliveryNotes === 'Sin notas adicionales')}
                    <button 
                      type="button"
                      onclick={() => toggleQuickNote(chip.replace(/^[^\s]+\s/, ''))}
                      class="px-2.5 py-1 rounded-full text-xs font-medium border transition-all cursor-pointer active:scale-95 {isSelected ? 'bg-[#FF2400] text-white border-[#FF2400]' : 'bg-[#FAF6EE] text-[#3A3A3C] border-black/10 hover:border-black/30'}"
                    >
                      {chip}
                    </button>
                  {/each}
                </div>
                <div class="flex items-center gap-2 pt-1">
                  <input 
                    type="text" 
                    bind:value={deliveryNotes}
                    placeholder="O escribe alguna otra indicación..."
                    class="flex-1 bg-[#FAF6EE] border border-black/10 rounded-2xl px-4 py-2.5 text-xs sm:text-sm font-sans text-[#0C0C0E] placeholder:text-ash/70 focus:outline-none focus:border-[#FF2400] transition-colors"
                    onkeydown={(e) => { if (e.key === 'Enter') submitNotes(); }}
                  />
                  <button 
                    type="button" 
                    onclick={submitNotes}
                    class="h-10 px-4 rounded-2xl bg-[#FF2400] hover:bg-[#E02000] active:scale-95 text-white font-display font-bold text-xs sm:text-sm flex items-center justify-center gap-1 transition-all cursor-pointer shadow-2xs shrink-0"
                  >
                    <span>Continuar</span>
                    <ArrowRight class="w-3 h-3" />
                  </button>
                </div>
              </div>

            {:else if waiterStep === 'payment'}
              <div class="space-y-2.5 pt-1">
                <!-- Métodos de pago táctiles -->
                <div class="grid grid-cols-3 gap-2">
                  {#each [
                    { id: 'Nequi / Bre-B', label: 'Nequi / Bancolombia', sub: 'Transferencia' },
                    { id: 'Efectivo', label: 'Efectivo', sub: 'Contra entrega' },
                    { id: 'Datáfono', label: 'Datáfono', sub: 'Tarjeta' }
                  ] as m}
                    <button
                      type="button"
                      onclick={() => { paymentMethod = m.id; }}
                      class="p-2.5 rounded-2xl border transition-all text-center flex flex-col items-center justify-center cursor-pointer active:scale-95 {paymentMethod === m.id ? 'bg-[#FF2400] text-white border-[#FF2400] shadow-xs' : 'bg-white text-[#0C0C0E] border-black/10 hover:border-black/25'}"
                    >
                      <span class="text-xs sm:text-sm font-display font-extrabold leading-tight block">{m.label}</span>
                      <span class="text-[10px] font-mono {paymentMethod === m.id ? 'text-white/80' : 'text-ash'} block mt-0.5">{m.sub}</span>
                    </button>
                  {/each}
                </div>

                <!-- Detalle si es Efectivo -->
                {#if paymentMethod === 'Efectivo'}
                  <div class="bg-white rounded-2xl p-3 border border-black/10 space-y-2">
                    <span class="text-xs font-mono font-bold text-[#0C0C0E] block">
                      ¿Con cuánto vas a pagar? (Llevamos tu cambio exacto):
                    </span>
                    <div class="flex flex-wrap gap-1.5">
                      <button
                        type="button"
                        onclick={() => setCashPreset(deliveryTotal)}
                        class="px-2.5 py-1 rounded-full text-xs font-mono font-bold border cursor-pointer {parseInt(cashAmount, 10) === deliveryTotal ? 'bg-[#FF2400] text-white border-[#FF2400]' : 'bg-[#FAF6EE] text-[#0C0C0E] border-black/10'}"
                      >
                        Exacto ({formatCOP(deliveryTotal)})
                      </button>
                      {#if deliveryTotal < 50000}
                        <button
                          type="button"
                          onclick={() => setCashPreset(50000)}
                          class="px-2.5 py-1 rounded-full text-xs font-mono font-bold border cursor-pointer {parseInt(cashAmount, 10) === 50000 ? 'bg-[#FF2400] text-white border-[#FF2400]' : 'bg-[#FAF6EE] text-[#0C0C0E] border-black/10'}"
                        >
                          $50.000
                        </button>
                      {/if}
                      {#if deliveryTotal < 100000}
                        <button
                          type="button"
                          onclick={() => setCashPreset(100000)}
                          class="px-2.5 py-1 rounded-full text-xs font-mono font-bold border cursor-pointer {parseInt(cashAmount, 10) === 100000 ? 'bg-[#FF2400] text-white border-[#FF2400]' : 'bg-[#FAF6EE] text-[#0C0C0E] border-black/10'}"
                        >
                          $100.000
                        </button>
                      {/if}
                    </div>
                    <input 
                      type="number" 
                      bind:value={cashAmount}
                      placeholder="O escribe el valor del billete..."
                      class="w-full bg-[#FAF6EE] border border-black/10 rounded-xl px-3 py-2 text-xs sm:text-sm font-mono text-[#0C0C0E] focus:outline-none focus:border-[#FF2400]"
                    />
                    {#if cashChange > 0}
                      <span class="text-xs font-mono text-emerald-700 font-bold block">
                        ✓ El repartidor te llevará {formatCOP(cashChange)} de cambio.
                      </span>
                    {/if}
                  </div>
                {/if}

                <!-- Botón gigante de confirmación de comanda -->
                <button
                  type="button"
                  onclick={handleConfirmDelivery}
                  class="w-full min-h-[60px] sm:min-h-[68px] rounded-full bg-[#FF2400] hover:bg-[#E02000] active:scale-[0.97] text-white font-display font-extrabold text-base sm:text-lg tracking-tight uppercase shadow-md hover:shadow-lg transition-all flex items-center justify-center gap-2 cursor-pointer select-none"
                >
                  <Sparkles class="w-5 h-5 stroke-[2.2]" />
                  <span>CONFIRMAR Y GENERAR TICKET</span>
                </button>
              </div>
            {/if}
          </footer>

        <!-- =======================================================
             FASE 3: EL TICKET DE COMANDA PERFORADO (DISEÑO INSPIRADO EN LA CAPTURA)
             ======================================================= -->
        {:else if cartStep === 'ticket' && createdOrder}
          <div 
            class="flex-1 min-h-0 overflow-y-auto py-2 pr-0.5 flex flex-col justify-between select-none"
            style="scrollbar-width: none; overscroll-behavior: contain;"
            onwheel={(e) => e.stopPropagation()}
            ontouchmove={(e) => e.stopPropagation()}
          >
            <!-- TARJETA DEL TICKET PERFORADO -->
            <div class="ticket-card w-full max-w-sm mx-auto bg-[#FFFDF5] border border-black/15 shadow-xl relative overflow-hidden rounded-[26px] sm:rounded-[30px] my-auto">
              
              <!-- Dientes Perforados Superiores (Scalloped Teeth) -->
              <div class="flex justify-between items-center -mt-2 px-1 overflow-hidden pointer-events-none" aria-hidden="true">
                {#each Array(15) as _}
                  <div class="w-3.5 h-3.5 rounded-full bg-[#FAF6EE] shrink-0 border-b border-black/10"></div>
                {/each}
              </div>

              <!-- Cabecera del Ticket -->
              <div class="pt-5 pb-3 px-6 text-center">
                <div class="flex items-center justify-center gap-2 mb-2">
                  <img src="/imagenes/marca/isotipo.webp" alt="Mías" class="w-7 h-7 object-contain" />
                  <span class="font-display font-black text-sm uppercase tracking-wider text-[#0C0C0E]">
                    Mías Hamburguesas
                  </span>
                </div>

                <h3 class="font-display font-black text-2xl sm:text-3xl text-[#0C0C0E] tracking-tight leading-none mt-1.5">
                  ¡Comanda Lista!
                </h3>
                <p class="font-sans text-xs text-ash mt-1 font-medium">
                  Lo íntimo hecho para compartir.
                </p>
              </div>

              <!-- Líneas Guía de Puntos (Igual a la referencia del usuario) -->
              <div class="px-6 py-2.5 space-y-2 font-mono text-xs text-[#3A3A3C]">
                <div class="flex items-baseline justify-between">
                  <span class="text-ash shrink-0">Modalidad</span>
                  <span class="flex-1 mx-2 border-b border-dotted border-black/30 relative -top-1"></span>
                  <span class="font-bold text-[#0C0C0E] shrink-0">{createdOrder.modeLabel}</span>
                </div>

                {#if createdOrder.table_number}
                  <div class="flex items-baseline justify-between">
                    <span class="text-ash shrink-0">Mesa</span>
                    <span class="flex-1 mx-2 border-b border-dotted border-black/30 relative -top-1"></span>
                    <span class="font-bold text-[#FF2400] shrink-0">Mesa {createdOrder.table_number}</span>
                  </div>
                {/if}

                <div class="flex items-baseline justify-between">
                  <span class="text-ash shrink-0">Cliente</span>
                  <span class="flex-1 mx-2 border-b border-dotted border-black/30 relative -top-1"></span>
                  <span class="font-bold text-[#0C0C0E] shrink-0 truncate max-w-[130px]">{createdOrder.customer_name}</span>
                </div>

                <div class="flex items-baseline justify-between">
                  <span class="text-ash shrink-0">Productos</span>
                  <span class="flex-1 mx-2 border-b border-dotted border-black/30 relative -top-1"></span>
                  <span class="font-bold text-[#0C0C0E] shrink-0">{createdOrder.items.reduce((s, i) => s + i.quantity, 0)} ítems</span>
                </div>

                <div class="flex items-baseline justify-between">
                  <span class="text-ash shrink-0">Subtotal</span>
                  <span class="flex-1 mx-2 border-b border-dotted border-black/30 relative -top-1"></span>
                  <span class="font-bold text-[#0C0C0E] shrink-0">{formatCOP(createdOrder.subtotal)}</span>
                </div>

                {#if createdOrder.delivery_fee > 0}
                  <div class="flex items-baseline justify-between">
                    <span class="text-ash shrink-0">Domicilio</span>
                    <span class="flex-1 mx-2 border-b border-dotted border-black/30 relative -top-1"></span>
                    <span class="font-bold text-[#0C0C0E] shrink-0">{formatCOP(createdOrder.delivery_fee)}</span>
                  </div>
                {/if}

                <div class="flex items-baseline justify-between pt-1 text-sm font-bold">
                  <span class="text-[#0C0C0E] shrink-0">Total comanda</span>
                  <span class="flex-1 mx-2 border-b border-dotted border-black/40 relative -top-1"></span>
                  <span class="font-display font-black text-base text-[#FF2400] shrink-0">{formatCOP(createdOrder.total)}</span>
                </div>
              </div>

              <!-- Muescas de Rasgado Lateral y Línea Troquelada Central -->
              <div class="relative my-3 py-1 flex items-center justify-center pointer-events-none" aria-hidden="true">
                <div class="absolute -left-3.5 w-7 h-7 rounded-full bg-[#FAF6EE] border-r border-black/15 shadow-inner"></div>
                <div class="w-full border-t-2 border-dashed border-black/25 mx-4"></div>
                <div class="absolute -right-3.5 w-7 h-7 rounded-full bg-[#FAF6EE] border-l border-black/15 shadow-inner"></div>
              </div>

              <!-- Sección Inferior con Gran Identificador Numérico (Idéntico a la referencia) -->
              <div class="pt-2 pb-6 px-6 text-center">
                <span class="text-[10px] font-mono uppercase tracking-widest text-ash block mb-1">
                  Número de Identificación
                </span>
                <button
                  type="button"
                  onclick={copyOrderCode}
                  class="group inline-flex items-center justify-center gap-2 cursor-pointer active:scale-95 transition-transform"
                  title="Toca para copiar código"
                >
                  <span class="font-display font-black text-4xl sm:text-5xl tracking-tight text-[#FF2400] drop-shadow-2xs">
                    #{createdOrder.code}
                  </span>
                  <Copy class="w-4 h-4 text-ash group-hover:text-[#FF2400] transition-colors" />
                </button>
                {#if isCopied}
                  <span class="text-[10px] font-mono text-emerald-600 font-bold block mt-1 animate-in fade-in">
                    ✓ Código copiado al portapapeles
                  </span>
                {/if}
              </div>

              <!-- Dientes Perforados Inferiores (Scalloped Teeth) -->
              <div class="flex justify-between items-center -mb-2 px-1 overflow-hidden pointer-events-none" aria-hidden="true">
                {#each Array(15) as _}
                  <div class="w-3.5 h-3.5 rounded-full bg-[#FAF6EE] shrink-0 border-t border-black/10"></div>
                {/each}
              </div>

            </div>

            <!-- Acciones Posteriores al Ticket -->
            <div class="w-full max-w-sm mx-auto mt-4 space-y-2.5 shrink-0">
              {#if createdOrder.type === 'delivery'}
                <button
                  type="button"
                  onclick={dispatchWhatsApp}
                  class="w-full min-h-[64px] rounded-full bg-emerald-600 hover:bg-emerald-500 active:scale-[0.97] text-white font-display font-extrabold text-base tracking-tight shadow-md transition-all flex items-center justify-center gap-2 cursor-pointer"
                >
                  <Send class="w-5 h-5 stroke-[2.2]" />
                  <span>Enviar a WhatsApp (321 252 0881)</span>
                </button>
              {/if}

              <button
                type="button"
                onclick={resetAndClose}
                class="w-full py-3.5 rounded-full bg-white border border-black/15 hover:bg-black/5 active:scale-95 text-[#0C0C0E] font-display font-extrabold text-sm tracking-tight transition-all cursor-pointer shadow-xs"
              >
                Finalizar y Volver a la Carta
              </button>
            </div>

          </div>
        {/if}

      </div>

    </div>
  </div>
{/if}

<style>
  .welcome-modal-backdrop {
    padding-left: 12px;
    padding-right: 12px;
    padding-bottom: 12px;
    padding-top: max(14px, calc(env(safe-area-inset-top, 0px) + 6px));
  }

  @media (min-width: 640px) {
    .welcome-modal-backdrop {
      padding: 1.25rem;
    }
  }

  @media (min-width: 768px) {
    .welcome-modal-backdrop {
      padding: 1.5rem;
    }
  }

  @media (min-width: 1024px) {
    .welcome-modal-backdrop {
      padding: 1.75rem;
    }
  }

  @media (min-width: 1280px) {
    .welcome-modal-backdrop {
      padding: 2.25rem;
    }
  }

  .welcome-card-modal {
    border: 1px solid rgba(0, 0, 0, 0.30);
    border-radius: 38px;
    -webkit-border-radius: 38px;
    overflow: hidden;
    isolation: isolate;
    will-change: transform, opacity;
    transform: translateZ(0);
    -webkit-backface-visibility: hidden;
    backface-visibility: hidden;
    contain: layout paint;
  }

  @media (min-width: 640px) {
    .welcome-card-modal {
      border-radius: 48px;
      -webkit-border-radius: 48px;
    }
  }

  @media (orientation: portrait) {
    .pill-action-btn {
      min-height: 76px !important;
      font-size: 1.22rem !important;
    }
  }

  @media (orientation: landscape) and (max-height: 720px) {
    .pill-action-btn {
      min-height: 66px !important;
      font-size: 1.15rem !important;
    }
  }
</style>
