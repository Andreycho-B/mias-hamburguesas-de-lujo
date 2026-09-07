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
    MessageSquare,
    DollarSign,
    CreditCard,
    Award,
    MapPin,
    Phone,
    User,
    FileText,
    Flame,
    Utensils
  } from 'lucide-svelte';

  let orderBtnEl = $state(null);
  let orderCardEl = $state(null);
  let orderBackdropEl = $state(null);
  let isModalAnimating = false;
  let recentlyAddedUpsellId = $state(null);

  // Pasos de la comanda: 'cart' | 'mode' | 'delivery_form' | 'ticket'
  let cartStep = $state('cart');
  let selectedMode = $state('delivery'); // 'delivery' | 'takeout' | 'in_store'
  
  // Datos del cliente para la orden
  let customerName = $state(store.user?.full_name || '');
  let customerPhone = $state(store.user?.phone || '');
  let deliveryAddress = $state(store.user?.address || '');
  let deliveryNeighborhood = $state('Centro');
  let deliveryNotes = $state('');
  let tableNumber = $state('1');

  // Método de pago
  let paymentMethod = $state('Efectivo'); // 'Efectivo' | 'Nequi / Daviplata' | 'Datáfono contraentrega'
  let cashAmount = $state('');
  let formError = $state('');

  const deliveryFee = 6000;
  let createdOrder = $state(null);
  let isCopied = $state(false);

  let liveTicketOrder = $derived(
    createdOrder ? (store.activeOrders.find(o => o.code === createdOrder.code) || createdOrder) : null
  );

  // Sincronizar datos si el usuario inicia sesión o cambia su perfil
  $effect(() => {
    if (store.user) {
      if (!customerName) customerName = store.user.full_name || '';
      if (!customerPhone) customerPhone = store.user.phone || '';
      if (!deliveryAddress && store.user.address) deliveryAddress = store.user.address;
    }
  });

  let currentDeliveryFee = $derived(selectedMode === 'delivery' ? deliveryFee : 0);
  let currentTotal = $derived(store.cartSubtotal + currentDeliveryFee);

  let cashChange = $derived.by(() => {
    const val = parseInt(cashAmount, 10);
    if (!isNaN(val) && val >= currentTotal) {
      return val - currentTotal;
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

  // Catálogo de productos sugeridos ("¿Deseas acompañar tu hamburguesa?")
  const upsellItems = [
    {
      id: 14,
      category_id: 4,
      name: 'Coca-Cola 400ml',
      price: 4000,
      image_url: '/imagenes/bebidas/cocacola.webp'
    },
    {
      id: 15,
      category_id: 4,
      name: 'Bretaña en Vidrio',
      price: 5000,
      image_url: '/imagenes/bebidas/bretana-vidrio.png'
    },
    {
      id: 17,
      category_id: 5,
      name: 'Papas Francesas',
      price: 8000,
      image_url: ''
    }
  ];

  function handleAddUpsell(product) {
    store.addToCart(product, false, []);
    recentlyAddedUpsellId = product.id;
    setTimeout(() => {
      if (recentlyAddedUpsellId === product.id) {
        recentlyAddedUpsellId = null;
      }
    }, 1200);
  }

  function handleContinueToMode() {
    if (store.cart.length === 0) return;
    cartStep = 'mode';
    formError = '';
  }

  function handleSelectMode(mode) {
    selectedMode = mode;
    cartStep = 'delivery_form';
    formError = '';
  }

  function handleBackStep() {
    formError = '';
    if (cartStep === 'ticket') {
      cartStep = 'cart';
    } else if (cartStep === 'delivery_form') {
      cartStep = 'mode';
    } else if (cartStep === 'mode') {
      cartStep = 'cart';
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

  async function handleFinalizeOrder() {
    formError = '';

    if (!customerName.trim()) {
      formError = 'Por favor ingresa tu nombre para registrar la comanda.';
      return;
    }
    if (!customerPhone.trim()) {
      formError = 'Por favor ingresa tu número de WhatsApp para notificarte.';
      return;
    }

    if (selectedMode === 'delivery') {
      if (!deliveryAddress.trim() || !deliveryNeighborhood.trim()) {
        formError = 'Por favor completa la dirección y el barrio de entrega en Duitama.';
        return;
      }
    } else if (selectedMode === 'in_store') {
      if (!tableNumber.trim()) {
        formError = 'Por favor ingresa el número de mesa en el restaurante.';
        return;
      }
    }

    let code = (selectedMode === 'delivery' ? 'DOM-' : 'ORD-') + Math.floor(1000 + Math.random() * 9000);
    let orderId = (selectedMode === 'delivery' ? 'dom-' : 'ord-') + Date.now();

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
        order_type: selectedMode,
        items: apiItems,
        delivery_name: customerName.trim(),
        delivery_phone: customerPhone.trim(),
        delivery_address: selectedMode === 'delivery' ? `${deliveryAddress.trim()}, Barrio ${deliveryNeighborhood.trim()}` : '',
        delivery_notes: deliveryNotes.trim(),
        payment_method: paymentMethod,
        cash_change_amount: parseInt(cashAmount, 10) || 0
      });
      if (apiRes && apiRes.order) {
        code = apiRes.order.code;
        orderId = apiRes.order.id;
      }
    } catch (err) {
      console.warn('[Order API] Guardado local Anillo 0', err);
    }

    const order = {
      id: orderId,
      code,
      type: selectedMode,
      modeLabel: selectedMode === 'delivery' ? 'A Domicilio' : selectedMode === 'takeout' ? 'Para Llevar' : 'Consumo en Salón',
      status: selectedMode === 'delivery' ? 'IN_PREPARATION' : 'PENDING_PAYMENT',
      customer_name: customerName.trim(),
      customer_phone: customerPhone.trim(),
      delivery_address: selectedMode === 'delivery' ? `${deliveryAddress.trim()}, Barrio ${deliveryNeighborhood.trim()}` : '',
      table_number: selectedMode === 'in_store' ? tableNumber.trim() : '',
      delivery_notes: deliveryNotes.trim(),
      items: [...store.cart],
      subtotal: store.cartSubtotal,
      delivery_fee: currentDeliveryFee,
      total: currentTotal,
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
    store.currentView = 'stamps';
  }

  function closeOrderModal() {
    store.isCartOpen = false;
  }
</script>

<!-- BOTÓN FLOTANTE "MI ORDEN" (VISIBLE CUANDO HAY PRODUCTOS Y EL CARRITO ESTÁ CERRADO) -->
{#if store.cartCount > 0 && !store.isCartOpen && store.currentView === 'menu'}
  <div 
    in:fly={{ y: 28, duration: 340, easing: backOut }}
    out:scale={{ start: 0.85, duration: 220, opacity: 0, easing: cubicOut }}
    class="fixed bottom-6 right-6 z-40 pb-safe will-change-transform"
  >
    <button
      bind:this={orderBtnEl}
      type="button"
      onclick={() => store.isCartOpen = true}
      class="min-h-[70px] sm:min-h-[76px] rounded-full bg-[#0C0C0E] text-white font-display font-black text-lg sm:text-xl tracking-tight border border-black/10 shadow-xl hover:bg-black active:scale-[0.97] transition-all flex items-center justify-center gap-3.5 cursor-pointer select-none px-7 sm:px-8 shrink-0"
      aria-label="Ver comanda ({store.cartCount} items)"
    >
      <ShoppingBag class="w-5 h-5 text-[#FF2400]" />
      <span>Mi Comanda</span>
      <span class="w-7 h-7 rounded-full bg-[#FF2400] text-white text-xs font-mono font-bold flex items-center justify-center shadow-xs shrink-0">
        {store.cartCount}
      </span>
    </button>
  </div>
{/if}

<!-- VISTA DE COMANDA Y CHECKOUT A PANTALLA COMPLETA FAT-FINGER -->
{#if store.isCartOpen}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div 
    data-lenis-prevent
    onwheel={(e) => e.stopPropagation()}
    ontouchmove={(e) => e.stopPropagation()}
    class="fixed inset-0 z-50 w-full h-full min-h-[100dvh] bg-[#FAF6EE] text-[#0C0C0E] overflow-y-auto overscroll-contain"
    style="touch-action: pan-y; -webkit-overflow-scrolling: touch; padding-top: max(1rem, env(safe-area-inset-top, 0px)); padding-bottom: max(5rem, calc(env(safe-area-inset-bottom, 0px) + 3rem)); padding-left: max(1rem, env(safe-area-inset-left, 0px)); padding-right: max(1rem, env(safe-area-inset-right, 0px));"
  >
    <div data-lenis-prevent class="w-full max-w-lg mx-auto py-2 sm:py-4 flex flex-col space-y-6">
      
      <!-- 1. Cabecera Unificada con Botón Volver y Botón Cerrar -->
      <header class="flex items-center justify-between w-full pb-1 shrink-0">
        {#if cartStep !== 'cart'}
          <button 
            type="button"
            onclick={handleBackStep}
            class="w-14 h-14 sm:w-16 sm:h-16 rounded-full bg-white border border-black/5 flex items-center justify-center text-[#3A3A3C] hover:bg-black/5 active:scale-90 transition-all cursor-pointer group shrink-0"
            aria-label="Volver al paso anterior"
          >
            <ChevronLeft class="w-6 h-6 sm:w-7 sm:h-7 stroke-[2.2] -translate-x-0.5 group-hover:-translate-x-1 transition-transform" />
          </button>
        {:else}
          <div class="w-14 h-14 sm:w-16 sm:h-16 flex items-center justify-center shrink-0">
            <img 
              src="/imagenes/marca/isotipo.webp" 
              alt="Mías" 
              class="w-7 h-7 object-contain opacity-85"
            />
          </div>
        {/if}

        <div class="text-center flex-1 px-2">
          <span class="font-mono text-[10px] sm:text-xs uppercase tracking-widest text-[#8E8E93] font-bold block">
            {#if cartStep === 'cart'}
              Mías • Comanda Gastronómica
            {:else if cartStep === 'mode'}
              Paso 1 de 2 • Modalidad
            {:else if cartStep === 'delivery_form'}
              Paso 2 de 2 • Entrega y Pago
            {:else}
              Orden Procesada
            {/if}
          </span>
          <h1 class="font-display font-black text-xl sm:text-2xl text-[#0C0C0E] tracking-tight leading-tight">
            {#if cartStep === 'cart'}
              Tu Comanda
            {:else if cartStep === 'mode'}
              ¿Dónde Disfrutas?
            {:else if cartStep === 'delivery_form'}
              {selectedMode === 'delivery' ? 'Datos de Domicilio' : selectedMode === 'takeout' ? 'Datos de Retiro' : 'Mesa en Salón'}
            {:else}
              ¡Comanda en Marcha!
            {/if}
          </h1>
        </div>

        <button 
          type="button"
          onclick={closeOrderModal}
          class="w-14 h-14 sm:w-16 sm:h-16 rounded-full bg-white border border-black/5 flex items-center justify-center text-[#3A3A3C] hover:text-[#FF2400] active:scale-90 transition-all cursor-pointer shrink-0"
          aria-label="Cerrar comanda"
        >
          <X class="w-6 h-6 stroke-[2.2]" />
        </button>
      </header>

      <!-- ========================================================================= -->
      <!-- FASE 1: LISTADO DE PRODUCTOS Y SELECCIÓN "¿ALGO MÁS?"                     -->
      <!-- ========================================================================= -->
      {#if cartStep === 'cart'}
        {#if store.cart.length === 0}
          <div class="bg-white border border-black/10 rounded-[36px] p-8 text-center space-y-4 my-auto shadow-none">
            <div class="w-20 h-20 rounded-full bg-[#EDE8DC] mx-auto flex items-center justify-center text-[#8E8E93]">
              <ShoppingBag class="w-8 h-8 stroke-[1.8]" />
            </div>
            <h2 class="font-display font-extrabold text-2xl text-[#0C0C0E]">Tu comanda está vacía</h2>
            <p class="text-xs sm:text-sm text-[#8E8E93] max-w-xs mx-auto">
              Explora nuestra carta gastronómica y añade tus hamburguesas favoritas al carbón.
            </p>
            <button
              type="button"
              onclick={() => { closeOrderModal(); store.currentView = 'hamburguesas'; }}
              class="w-full min-h-[58px] rounded-full bg-[#FF2400] text-white font-display font-extrabold text-base hover:bg-[#E02000] active:scale-[0.97] transition-all cursor-pointer"
            >
              Explorar Hamburguesas
            </button>
          </div>
        {:else}
          <!-- Lista de Tarjetas Gordas de Productos -->
          <div class="space-y-3.5">
            {#each store.cart as item (item.id)}
              <div class="bg-white border border-black/10 rounded-[32px] p-4 sm:p-5 flex items-center gap-3.5 sm:gap-4 shadow-none">
                <!-- Imagen o Icono del Producto -->
                <div class="w-20 h-20 sm:w-22 sm:h-22 rounded-2xl bg-[#EDE8DC] flex items-center justify-center shrink-0 overflow-hidden border border-black/5">
                  {#if item.product?.image_url}
                    <img 
                      src={item.product.image_url} 
                      alt={item.product?.name} 
                      class="w-full h-full object-contain"
                    />
                  {:else}
                    <!-- Icono representativo para adicionales y acompañamientos -->
                    <div class="w-full h-full flex flex-col items-center justify-center text-[#FF2400] p-1 text-center select-none">
                      {#if item.product?.slug?.includes('papa')}
                        <Utensils class="w-7 h-7 sm:w-8 sm:h-8 stroke-[2.2]" />
                        <span class="text-[9px] font-mono font-bold text-[#0C0C0E] mt-0.5 uppercase tracking-tighter">Papas</span>
                      {:else if item.product?.slug?.includes('tocineta') || item.product?.slug?.includes('carne')}
                        <Flame class="w-7 h-7 sm:w-8 sm:h-8 stroke-[2.2]" />
                        <span class="text-[9px] font-mono font-bold text-[#0C0C0E] mt-0.5 uppercase tracking-tighter">Extra</span>
                      {:else if item.product?.slug?.includes('queso')}
                        <Sparkles class="w-7 h-7 sm:w-8 sm:h-8 stroke-[2.2]" />
                        <span class="text-[9px] font-mono font-bold text-[#0C0C0E] mt-0.5 uppercase tracking-tighter">Queso</span>
                      {:else if item.product?.slug?.includes('combo')}
                        <ShoppingBag class="w-7 h-7 sm:w-8 sm:h-8 stroke-[2.2]" />
                        <span class="text-[9px] font-mono font-bold text-[#0C0C0E] mt-0.5 uppercase tracking-tighter">Combo</span>
                      {:else if item.product?.slug?.includes('salsa') || item.product?.slug?.includes('pepinillos')}
                        <Utensils class="w-7 h-7 sm:w-8 sm:h-8 stroke-[2.2]" />
                        <span class="text-[9px] font-mono font-bold text-[#0C0C0E] mt-0.5 uppercase tracking-tighter">Extra</span>
                      {:else}
                        <Plus class="w-7 h-7 sm:w-8 sm:h-8 stroke-[2.5]" />
                        <span class="text-[9px] font-mono font-bold text-[#0C0C0E] mt-0.5 uppercase tracking-tighter">Adicional</span>
                      {/if}
                    </div>
                  {/if}
                </div>

                <!-- Datos del Producto -->
                <div class="flex-1 min-w-0">
                  <div class="flex items-center gap-2 flex-wrap">
                    {#if item.product?.bread_type}
                      <span class="px-2 py-0.5 rounded-full bg-[#EDE8DC] text-[10px] font-mono font-bold uppercase text-[#0C0C0E]">
                        {item.product.bread_type === 'pretzel' ? 'Pan Pretzel' : 'Pan Brioche'}
                      </span>
                    {:else}
                      <span class="px-2 py-0.5 rounded-full bg-[#EDE8DC] text-[10px] font-mono font-bold uppercase text-[#8E8E93]">
                        {item.product?.category_id === 5 ? 'Adicional' : item.product?.category_id === 4 ? 'Bebida' : 'Acompañamiento'}
                      </span>
                    {/if}
                    {#if item.combo}
                      <span class="px-2 py-0.5 rounded-full bg-[#FF2400]/10 text-[10px] font-mono font-bold text-[#FF2400]">
                        + Combo
                      </span>
                    {/if}
                  </div>
                  <h3 class="font-display font-black text-lg sm:text-xl text-[#0C0C0E] truncate mt-1">
                    {item.product?.name}
                  </h3>
                  {#if item.extras && item.extras.length > 0}
                    <p class="text-[11px] font-mono text-[#8E8E93] truncate">
                      + {item.extras.map(e => e.name).join(', ')}
                    </p>
                  {/if}
                  <span class="font-display font-extrabold text-base sm:text-lg text-[#FF2400] block mt-0.5">
                    {formatCOP(item.totalPrice)}
                  </span>
                </div>

                <!-- Controles Táctiles Gordos (Touch Targets >= 48px) -->
                <div class="flex flex-col items-center gap-1.5 shrink-0">
                  <div class="flex items-center gap-1 bg-[#EDE8DC] p-1 rounded-full border border-black/5">
                    <button
                      type="button"
                      onclick={() => store.updateQuantity(item.id, item.quantity - 1)}
                      class="w-10 h-10 rounded-full bg-white text-[#0C0C0E] flex items-center justify-center hover:bg-black/5 active:scale-90 transition-all cursor-pointer"
                      aria-label="Disminuir cantidad"
                    >
                      <Minus class="w-4 h-4 stroke-[2.5]" />
                    </button>
                    <span class="w-7 text-center font-mono font-black text-sm text-[#0C0C0E]">
                      {item.quantity}
                    </span>
                    <button
                      type="button"
                      onclick={() => store.updateQuantity(item.id, item.quantity + 1)}
                      class="w-10 h-10 rounded-full bg-white text-[#0C0C0E] flex items-center justify-center hover:bg-black/5 active:scale-90 transition-all cursor-pointer"
                      aria-label="Aumentar cantidad"
                    >
                      <Plus class="w-4 h-4 stroke-[2.5]" />
                    </button>
                  </div>
                  <button
                    type="button"
                    onclick={() => store.removeFromCart(item.id)}
                    class="text-[11px] font-mono text-[#8E8E93] hover:text-[#FF2400] transition-colors py-0.5 cursor-pointer"
                  >
                    Eliminar
                  </button>
                </div>
              </div>
            {/each}
          </div>

          <!-- Sección ¿Deseas acompañar tu hamburguesa? (Adicionales Rápidos) -->
          <div class="bg-white border border-black/10 rounded-[36px] p-5 sm:p-6 space-y-3.5 shadow-none">
            <div class="flex items-center justify-between">
              <span class="text-xs font-mono font-bold uppercase tracking-wider text-[#8E8E93]">
                ¿Deseas acompañar tu orden?
              </span>
              <span class="text-[11px] font-mono text-[#FF2400] font-bold">1 toque para añadir</span>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-3 gap-2.5">
              {#each upsellItems as item}
                <button
                  type="button"
                  onclick={() => handleAddUpsell(item)}
                  class="p-3.5 rounded-2xl bg-[#FAF6EE] border border-black/5 hover:border-[#FF2400]/40 flex items-center justify-between sm:flex-col sm:items-start text-left transition-all active:scale-[0.98] cursor-pointer"
                >
                  <div class="min-w-0 pr-2">
                    <span class="font-display font-extrabold text-sm text-[#0C0C0E] block truncate">{item.name}</span>
                    <span class="font-mono text-xs text-[#FF2400] font-bold block mt-0.5">+{formatCOP(item.price)}</span>
                  </div>
                  <div class="w-8 h-8 rounded-full bg-white border border-black/10 flex items-center justify-center text-[#0C0C0E] shrink-0 sm:mt-2 sm:self-end">
                    {#if recentlyAddedUpsellId === item.id}
                      <Check class="w-4 h-4 text-emerald-600 stroke-[3]" />
                    {:else}
                      <Plus class="w-4 h-4 stroke-[2.5]" />
                    {/if}
                  </div>
                </button>
              {/each}
            </div>
          </div>

          <!-- Total de Comanda y Botón Fat-Finger CTA -->
          <div class="bg-white border border-black/10 rounded-[36px] p-6 space-y-4 shadow-none">
            <div class="flex items-center justify-between">
              <div>
                <span class="text-xs font-mono text-[#8E8E93] uppercase tracking-wider block">Subtotal de platos</span>
                <span class="font-display font-extrabold text-2xl sm:text-3xl text-[#0C0C0E] leading-tight">
                  {formatCOP(store.cartSubtotal)}
                </span>
              </div>
              <span class="text-xs font-mono font-bold text-[#8E8E93] bg-[#EDE8DC] px-3 py-1.5 rounded-full">
                {store.cartCount} {store.cartCount === 1 ? 'producto' : 'productos'}
              </span>
            </div>

            <button
              type="button"
              onclick={handleContinueToMode}
              class="w-full min-h-[64px] rounded-full bg-[#FF2400] hover:bg-[#E02000] active:scale-[0.97] text-white font-display font-extrabold text-lg sm:text-xl tracking-tight transition-all flex items-center justify-center gap-2 cursor-pointer shadow-none"
            >
              <span>Continuar con el Pedido</span>
              <ArrowRight class="w-5 h-5 stroke-[2.5]" />
            </button>
          </div>
        {/if}

      <!-- ========================================================================= -->
      <!-- FASE 2: SELECCIÓN DE MODALIDAD (TARJETAS GORDAS TÁCTILES)                  -->
      <!-- ========================================================================= -->
      {:else if cartStep === 'mode'}
        <div class="space-y-4">
          <!-- Opción 1: A Domicilio -->
          <button
            type="button"
            onclick={() => handleSelectMode('delivery')}
            class="w-full text-left bg-white rounded-[32px] border-2 p-5 sm:p-6 transition-all active:scale-[0.985] cursor-pointer flex items-center gap-4.5 {selectedMode === 'delivery' ? 'border-[#FF2400] bg-white' : 'border-black/10 hover:border-black/20'}"
          >
            <div class="w-14 h-14 rounded-2xl bg-[#FF2400] text-white flex items-center justify-center shrink-0">
              <Bike class="w-7 h-7 stroke-[2.2]" />
            </div>
            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between">
                <h2 class="font-display font-black text-xl text-[#0C0C0E]">A Domicilio</h2>
                <span class="px-2.5 py-1 rounded-full bg-[#FF2400]/10 text-xs font-mono font-bold text-[#FF2400]">
                  + $6.000
                </span>
              </div>
              <p class="text-xs text-[#8E8E93] mt-1 leading-relaxed">
                Directo a tu puerta en Duitama, empacado hermético con sello de calor.
              </p>
            </div>
          </button>

          <!-- Opción 2: Para Llevar -->
          <button
            type="button"
            onclick={() => handleSelectMode('takeout')}
            class="w-full text-left bg-white rounded-[32px] border-2 p-5 sm:p-6 transition-all active:scale-[0.985] cursor-pointer flex items-center gap-4.5 {selectedMode === 'takeout' ? 'border-[#FF2400] bg-white' : 'border-black/10 hover:border-black/20'}"
          >
            <div class="w-14 h-14 rounded-2xl bg-[#0C0C0E] text-white flex items-center justify-center shrink-0">
              <ShoppingBag class="w-7 h-7 stroke-[2.2]" />
            </div>
            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between">
                <h2 class="font-display font-black text-xl text-[#0C0C0E]">Para Llevar (Take Away)</h2>
                <span class="px-2.5 py-1 rounded-full bg-[#EDE8DC] text-xs font-mono font-bold text-[#0C0C0E]">
                  Sin recargo
                </span>
              </div>
              <p class="text-xs text-[#8E8E93] mt-1 leading-relaxed">
                Empacado para retirar en barra sin hacer filas.
              </p>
            </div>
          </button>

          <!-- Opción 3: Consumo en Salón -->
          <button
            type="button"
            onclick={() => handleSelectMode('in_store')}
            class="w-full text-left bg-white rounded-[32px] border-2 p-5 sm:p-6 transition-all active:scale-[0.985] cursor-pointer flex items-center gap-4.5 {selectedMode === 'in_store' ? 'border-[#FF2400] bg-white' : 'border-black/10 hover:border-black/20'}"
          >
            <div class="w-14 h-14 rounded-2xl bg-[#EDE8DC] text-[#0C0C0E] flex items-center justify-center shrink-0">
              <UtensilsCrossed class="w-7 h-7 stroke-[2.2]" />
            </div>
            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between">
                <h2 class="font-display font-black text-xl text-[#0C0C0E]">Consumo en el Salón</h2>
                <span class="px-2.5 py-1 rounded-full bg-[#EDE8DC] text-xs font-mono font-bold text-[#0C0C0E]">
                  En mesa
                </span>
              </div>
              <p class="text-xs text-[#8E8E93] mt-1 leading-relaxed">
                Disfruta la experiencia en nuestro punto físico, te lo llevamos a la mesa.
              </p>
            </div>
          </button>
        </div>

      <!-- ========================================================================= -->
      <!-- FASE 3: DATOS DE ENTREGA Y PAGO (FAT-FINGER DESIGN DE ALTA CONVERSIÓN)     -->
      <!-- ========================================================================= -->
      {:else if cartStep === 'delivery_form'}
        <div class="space-y-4">
          
          <!-- SECCIÓN 1: ¿A QUIÉN ENTREGAMOS? -->
          <div class="bg-white border border-black/10 rounded-[34px] p-5 sm:p-6 space-y-3.5 shadow-none">
            <div class="flex items-center gap-2">
              <User class="w-5 h-5 text-[#FF2400]" />
              <h2 class="font-display font-extrabold text-lg text-[#0C0C0E]">
                ¿A quién entregamos?
              </h2>
            </div>

            <div class="space-y-2.5">
              <div class="relative">
                <input 
                  type="text" 
                  bind:value={customerName}
                  placeholder="Nombre completo"
                  class="w-full min-h-[58px] rounded-2xl px-5 bg-[#EDE8DC] border border-black/10 text-[#0C0C0E] placeholder:text-[#8E8E93] text-base font-sans focus:outline-none focus:border-black/30 transition-colors"
                />
              </div>

              <div class="relative">
                <input 
                  type="tel" 
                  bind:value={customerPhone}
                  placeholder="Número de WhatsApp (ej. 321 252 0881)"
                  class="w-full min-h-[58px] rounded-2xl px-5 bg-[#EDE8DC] border border-black/10 text-[#0C0C0E] placeholder:text-[#8E8E93] text-base font-sans focus:outline-none focus:border-black/30 transition-colors"
                />
              </div>
            </div>
          </div>

          <!-- SECCIÓN 2: LUGAR DE ENTREGA O MESA -->
          {#if selectedMode === 'delivery'}
            <div class="bg-white border border-black/10 rounded-[34px] p-5 sm:p-6 space-y-3.5 shadow-none">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2">
                  <MapPin class="w-5 h-5 text-[#FF2400]" />
                  <h2 class="font-display font-extrabold text-lg text-[#0C0C0E]">
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
                  class="w-full min-h-[58px] rounded-2xl px-5 bg-[#EDE8DC] border border-black/10 text-[#0C0C0E] placeholder:text-[#8E8E93] text-base font-sans focus:outline-none focus:border-black/30 transition-colors"
                />

                <input 
                  type="text" 
                  bind:value={deliveryNeighborhood}
                  placeholder="Barrio o sector (ej. Centro, El Recreo, Sugamuxi)"
                  class="w-full min-h-[58px] rounded-2xl px-5 bg-[#EDE8DC] border border-black/10 text-[#0C0C0E] placeholder:text-[#8E8E93] text-base font-sans focus:outline-none focus:border-black/30 transition-colors"
                />

                <input 
                  type="text" 
                  bind:value={deliveryNotes}
                  placeholder="Indicaciones para el repartidor (ej. Casa blanca, timbre 201)"
                  class="w-full min-h-[54px] rounded-2xl px-5 bg-[#EDE8DC] border border-black/10 text-[#0C0C0E] placeholder:text-[#8E8E93] text-sm font-sans focus:outline-none focus:border-black/30 transition-colors"
                />

                <!-- Chips Rápidos para Notas de Entrega -->
                <div class="flex flex-wrap gap-1.5 pt-1">
                  {#each ['🔔 Tocar timbre', '🏢 Dejar en portería', '🛵 Llamar al llegar'] as chip}
                    <button
                      type="button"
                      onclick={() => toggleQuickNote(chip)}
                      class="px-3 py-1.5 rounded-full text-xs font-medium border transition-all cursor-pointer {deliveryNotes.includes(chip) ? 'bg-[#FF2400] text-white border-[#FF2400]' : 'bg-[#EDE8DC] text-[#0C0C0E] border-black/10'}"
                    >
                      {chip}
                    </button>
                  {/each}
                </div>
              </div>
            </div>
          {:else if selectedMode === 'in_store'}
            <div class="bg-white border border-black/10 rounded-[34px] p-5 sm:p-6 space-y-3 shadow-none">
              <div class="flex items-center gap-2">
                <UtensilsCrossed class="w-5 h-5 text-[#FF2400]" />
                <h2 class="font-display font-extrabold text-lg text-[#0C0C0E]">Número de Mesa</h2>
              </div>
              <input 
                type="text" 
                bind:value={tableNumber}
                placeholder="Número de mesa (ej. 4)"
                class="w-full min-h-[58px] rounded-2xl px-5 bg-[#EDE8DC] border border-black/10 text-[#0C0C0E] text-base font-sans focus:outline-none focus:border-black/30 transition-colors"
              />
            </div>
          {/if}

          <!-- SECCIÓN 3: MÉTODO DE PAGO (TARJETAS GORDAS) -->
          <div class="bg-white border border-black/10 rounded-[34px] p-5 sm:p-6 space-y-3.5 shadow-none">
            <div class="flex items-center justify-between">
              <h2 class="font-display font-extrabold text-lg text-[#0C0C0E]">Método de Pago</h2>
              <span class="text-xs font-mono font-bold text-[#8E8E93]">{paymentMethod}</span>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-3 gap-2.5">
              <!-- Efectivo -->
              <button
                type="button"
                onclick={() => paymentMethod = 'Efectivo'}
                class="p-4 rounded-2xl border-2 text-left transition-all cursor-pointer flex flex-col justify-between min-h-[84px] {paymentMethod === 'Efectivo' ? 'border-[#FF2400] bg-[#FAF6EE]' : 'border-black/10 bg-white'}"
              >
                <DollarSign class="w-5 h-5 text-[#FF2400]" />
                <div>
                  <span class="font-display font-extrabold text-sm text-[#0C0C0E] block">Efectivo</span>
                  <span class="text-[11px] text-[#8E8E93]">Contraentrega</span>
                </div>
              </button>

              <!-- Nequi / Daviplata -->
              <button
                type="button"
                onclick={() => paymentMethod = 'Nequi / Daviplata'}
                class="p-4 rounded-2xl border-2 text-left transition-all cursor-pointer flex flex-col justify-between min-h-[84px] {paymentMethod === 'Nequi / Daviplata' ? 'border-[#FF2400] bg-[#FAF6EE]' : 'border-black/10 bg-white'}"
              >
                <Phone class="w-5 h-5 text-[#FF2400]" />
                <div>
                  <span class="font-display font-extrabold text-sm text-[#0C0C0E] block">Nequi / Davi</span>
                  <span class="text-[11px] text-[#8E8E93]">Transferencia</span>
                </div>
              </button>

              <!-- Datáfono -->
              <button
                type="button"
                onclick={() => paymentMethod = 'Datáfono contraentrega'}
                class="p-4 rounded-2xl border-2 text-left transition-all cursor-pointer flex flex-col justify-between min-h-[84px] {paymentMethod === 'Datáfono contraentrega' ? 'border-[#FF2400] bg-[#FAF6EE]' : 'border-black/10 bg-white'}"
              >
                <CreditCard class="w-5 h-5 text-[#FF2400]" />
                <div>
                  <span class="font-display font-extrabold text-sm text-[#0C0C0E] block">Datáfono</span>
                  <span class="text-[11px] text-[#8E8E93]">Tarjeta física</span>
                </div>
              </button>
            </div>

            <!-- Detalles según método de pago seleccionado -->
            {#if paymentMethod === 'Efectivo'}
              <div class="p-4 rounded-2xl bg-[#EDE8DC] space-y-2.5">
                <span class="text-xs font-mono font-bold text-[#0C0C0E] block">
                  ¿Con cuánto cancelas? (Para llevar tu cambio exacto)
                </span>
                <div class="grid grid-cols-3 gap-2">
                  <button
                    type="button"
                    onclick={() => selectQuickCash(50000)}
                    class="py-2.5 rounded-xl bg-white border border-black/10 text-xs font-bold font-mono text-[#0C0C0E] hover:border-[#FF2400] cursor-pointer"
                  >
                    $50.000
                  </button>
                  <button
                    type="button"
                    onclick={() => selectQuickCash(100000)}
                    class="py-2.5 rounded-xl bg-white border border-black/10 text-xs font-bold font-mono text-[#0C0C0E] hover:border-[#FF2400] cursor-pointer"
                  >
                    $100.000
                  </button>
                  <button
                    type="button"
                    onclick={() => selectQuickCash('exact')}
                    class="py-2.5 rounded-xl bg-white border border-black/10 text-xs font-bold font-mono text-[#0C0C0E] hover:border-[#FF2400] cursor-pointer"
                  >
                    Exacto
                  </button>
                </div>

                <div class="flex items-center gap-2 pt-1">
                  <input 
                    type="number"
                    bind:value={cashAmount}
                    placeholder="Otro valor en efectivo..."
                    class="flex-1 min-h-[48px] rounded-xl px-4 bg-white border border-black/10 text-xs sm:text-sm font-mono text-[#0C0C0E] focus:outline-none focus:border-[#FF2400]"
                  />
                  {#if cashChange > 0}
                    <span class="text-xs font-mono font-bold text-emerald-700 bg-white px-3 py-3 rounded-xl border border-black/10 shrink-0">
                      Vueltas: {formatCOP(cashChange)}
                    </span>
                  {/if}
                </div>
              </div>
            {:else if paymentMethod === 'Nequi / Daviplata'}
              <div class="p-4 rounded-2xl bg-[#EDE8DC] flex items-center justify-between">
                <div>
                  <span class="text-[11px] font-mono uppercase text-[#8E8E93] block">Cuenta Nequi & Daviplata Mías</span>
                  <span class="font-display font-black text-lg text-[#0C0C0E]">321 252 0881</span>
                </div>
                <button
                  type="button"
                  onclick={copyNequi}
                  class="min-h-[42px] px-4 rounded-xl bg-white border border-black/10 text-xs font-mono font-bold text-[#0C0C0E] hover:border-[#FF2400] transition-colors cursor-pointer flex items-center gap-1.5"
                >
                  <Copy class="w-3.5 h-3.5" />
                  <span>{isCopied ? '¡Copiado!' : 'Copiar'}</span>
                </button>
              </div>
            {/if}
          </div>

          <!-- SECCIÓN 4: RESUMEN Y BOTÓN FAT-FINGER FINALIZAR -->
          <div class="bg-white border border-black/10 rounded-[34px] p-6 space-y-4 shadow-none">
            <div class="space-y-2 text-sm">
              <div class="flex justify-between text-[#8E8E93]">
                <span>Subtotal de platos:</span>
                <span class="font-mono font-medium text-[#0C0C0E]">{formatCOP(store.cartSubtotal)}</span>
              </div>
              <div class="flex justify-between text-[#8E8E93]">
                <span>Costo de domicilio:</span>
                <span class="font-mono font-medium text-[#0C0C0E]">
                  {currentDeliveryFee > 0 ? formatCOP(currentDeliveryFee) : 'Gratis'}
                </span>
              </div>
              <div class="border-t border-black/10 pt-2 flex items-center justify-between">
                <span class="font-display font-black text-xl text-[#0C0C0E]">Total a Pagar</span>
                <span class="font-display font-black text-2xl sm:text-3xl text-[#FF2400]">
                  {formatCOP(currentTotal)}
                </span>
              </div>
            </div>

            {#if formError}
              <div class="p-3.5 rounded-2xl bg-red-50 border border-red-200 text-xs font-bold text-[#FF2400] text-center">
                {formError}
              </div>
            {/if}

            <button
              type="button"
              onclick={handleFinalizeOrder}
              class="w-full min-h-[64px] rounded-full bg-[#FF2400] hover:bg-[#E02000] active:scale-[0.97] text-white font-display font-extrabold text-lg sm:text-xl tracking-tight transition-all flex items-center justify-center gap-2 cursor-pointer shadow-none"
            >
              <span>Confirmar y Finalizar Pedido</span>
              <Check class="w-5 h-5 stroke-[2.5]" />
            </button>
          </div>
        </div>

      <!-- ========================================================================= -->
      <!-- FASE 4: COMANDA CONFIRMADA (TICKET Y ACCIONES INMEDIATAS)                  -->
      <!-- ========================================================================= -->
      {:else if cartStep === 'ticket' && (liveTicketOrder || createdOrder)}
        {@const ord = liveTicketOrder || createdOrder}
        <div class="bg-white border border-black/10 rounded-[38px] p-6 sm:p-8 space-y-6 shadow-none text-center">
          
          <!-- Encabezado de Éxito con Estado en Vivo -->
          <div class="space-y-2">
            <div class="w-16 h-16 rounded-full {ord.status === 'PAID' ? 'bg-emerald-600' : 'bg-[#FF2400]'} text-white mx-auto flex items-center justify-center shadow-xs transition-colors duration-500">
              <Check class="w-8 h-8 stroke-[3]" />
            </div>
            <span class="font-mono text-xs uppercase font-bold tracking-widest text-[#FF2400] block">
              Comanda #{ord.code}
            </span>
            <h2 class="font-display font-black text-2xl sm:text-3xl text-[#0C0C0E]">
              {#if ord.status === 'PAID'}
                ¡Comanda Cobrada y Aprobada!
              {:else if ord.status === 'IN_PREPARATION'}
                ¡Tu comanda está al fuego!
              {:else if ord.status === 'ON_THE_WAY'}
                ¡Tu comanda va en camino!
              {:else if ord.status === 'DELIVERED'}
                ¡Pedido Entregado!
              {:else}
                ¡Tu comanda está registrada!
              {/if}
            </h2>
            <p class="text-xs sm:text-sm text-[#8E8E93] max-w-sm mx-auto">
              {#if ord.status === 'PAID'}
                La caja ha confirmado tu pago satisfactoriamente. Tu orden se está procesando.
              {:else}
                Para despachar de inmediato al fuego y coordinar la entrega, envía tu comanda por WhatsApp oficial.
              {/if}
            </p>
          </div>

          <!-- Resumen del Ticket Térmico con Estado en Vivo -->
          <div class="p-5 rounded-3xl bg-[#EDE8DC] text-left space-y-3 font-mono text-xs">
            <div class="flex justify-between items-center border-b border-black/10 pb-2">
              <span class="text-[#8E8E93]">Estado en Vivo:</span>
              <span class="font-mono font-bold text-[11px] uppercase px-2.5 py-1 rounded-full {
                ord.status === 'PAID' ? 'bg-emerald-200 text-emerald-900 border border-emerald-300' :
                ord.status === 'IN_PREPARATION' ? 'bg-amber-200 text-amber-900 border border-amber-300' :
                ord.status === 'ON_THE_WAY' ? 'bg-blue-200 text-blue-900 border border-blue-300' :
                ord.status === 'DELIVERED' ? 'bg-purple-200 text-purple-900 border border-purple-300' :
                'bg-orange-100 text-[#FF2400] border border-[#FF2400]/30 animate-pulse'
              }">
                {#if ord.status === 'PAID'}
                  ✓ Pago Aprobado en Caja
                {:else if ord.status === 'IN_PREPARATION'}
                  🔥 En Preparación al Fuego
                {:else if ord.status === 'ON_THE_WAY'}
                  🛵 En Camino al Domicilio
                {:else if ord.status === 'DELIVERED'}
                  ✨ Pedido Entregado
                {:else}
                  ⏳ Pendiente de Cobro en Caja
                {/if}
              </span>
            </div>
            <div class="flex justify-between border-b border-black/10 pb-2">
              <span class="text-[#8E8E93]">Modalidad:</span>
              <span class="font-bold text-[#0C0C0E]">{ord.modeLabel || ord.type}</span>
            </div>
            <div class="flex justify-between border-b border-black/10 pb-2">
              <span class="text-[#8E8E93]">Cliente:</span>
              <span class="font-bold text-[#0C0C0E]">{ord.customer_name || ord.deliveryName}</span>
            </div>
            {#if ord.delivery_address || ord.deliveryAddress}
              <div class="flex justify-between border-b border-black/10 pb-2">
                <span class="text-[#8E8E93]">Dirección:</span>
                <span class="font-bold text-[#0C0C0E] text-right">{ord.delivery_address || ord.deliveryAddress}</span>
              </div>
            {/if}
            {#if ord.table_number || ord.tableNumber}
              <div class="flex justify-between border-b border-black/10 pb-2">
                <span class="text-[#8E8E93]">Mesa:</span>
                <span class="font-bold text-[#0C0C0E]">#{ord.table_number || ord.tableNumber}</span>
              </div>
            {/if}
            <div class="flex justify-between pt-1">
              <span class="font-display font-extrabold text-sm text-[#0C0C0E]">Total Comanda:</span>
              <span class="font-display font-black text-base text-[#FF2400]">{formatCOP(ord.total)}</span>
            </div>
          </div>

          <!-- Botones Fat-Finger de Alta Conversión -->
          <div class="space-y-3 pt-2">
            <!-- Botón 1: Enviar por WhatsApp -->
            <button
              type="button"
              onclick={dispatchWhatsApp}
              class="w-full min-h-[64px] rounded-full bg-[#FF2400] hover:bg-[#E02000] active:scale-[0.97] text-white font-display font-extrabold text-base sm:text-lg tracking-tight transition-all flex items-center justify-center gap-2.5 cursor-pointer shadow-none"
            >
              <Send class="w-5 h-5" />
              <span>Enviar Pedido por WhatsApp</span>
            </button>

            <!-- Botón 2: Solicitar Sellos para la Cartilla -->
            <button
              type="button"
              onclick={goToStampsFromTicket}
              class="w-full min-h-[58px] rounded-full bg-white hover:bg-[#FAF6EE] text-[#0C0C0E] border border-black/15 font-display font-extrabold text-sm sm:text-base tracking-tight transition-all flex items-center justify-center gap-2 cursor-pointer"
            >
              <Award class="w-5 h-5 text-[#FF2400]" />
              <span>Solicitar Sellos en mi Cartilla</span>
            </button>

            <!-- Botón 3: Volver a la Carta -->
            <button
              type="button"
              onclick={() => { closeOrderModal(); store.currentView = 'menu'; }}
              class="w-full py-2.5 text-xs font-mono text-[#8E8E93] hover:text-[#0C0C0E] transition-colors cursor-pointer"
            >
              Volver a la carta principal
            </button>
          </div>
        </div>
      {/if}

    </div>
  </div>
{/if}
