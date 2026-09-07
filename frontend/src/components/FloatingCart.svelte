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
    DollarSign, 
    CreditCard, 
    Award, 
    MapPin, 
    Phone, 
    User, 
    Clock
  } from 'lucide-svelte';

  let recentlyAddedUpsellId = $state(null);

  // Pasos de la comanda interactiva:
  // 'cart' | 'mode' | 'delivery_address' | 'delivery_neighborhood' | 'delivery_payment' | 'in_store_confirm' | 'takeout_confirm' | 'ticket'
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

  const duitamaNeighborhoods = ['Centro', 'El Sol', 'Cándido Quintero', 'La Floresta', 'San Antonio', 'Higueras', 'El Recreo'];

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
    formError = '';
    if (mode === 'delivery') {
      cartStep = 'delivery_address';
    } else if (mode === 'in_store') {
      cartStep = 'in_store_confirm';
    } else if (mode === 'takeout') {
      cartStep = 'takeout_confirm';
    }
  }

  function handleBackStep() {
    formError = '';
    if (cartStep === 'ticket') {
      cartStep = 'cart';
    } else if (cartStep === 'delivery_payment') {
      cartStep = 'delivery_neighborhood';
    } else if (cartStep === 'delivery_neighborhood') {
      cartStep = 'delivery_address';
    } else if (cartStep === 'delivery_address' || cartStep === 'in_store_confirm' || cartStep === 'takeout_confirm') {
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

  function copyNequi() {
    if (typeof navigator !== 'undefined' && navigator.clipboard) {
      navigator.clipboard.writeText('3212520881');
      isCopied = true;
      setTimeout(() => isCopied = false, 2000);
    }
  }

  // Finalizar Comanda
  async function handleFinalizeOrder() {
    formError = '';

    // REQUISITO EXIGIDO: Si el usuario no tiene cuenta, salta la opción de crearla antes de enviar el pedido
    if (!store.user) {
      formError = 'Crea tu cuenta o inicia sesión para asociar tu pedido y registrar tus sellos de fidelización.';
      store.isAuthModalOpen = true;
      return;
    }

    if (selectedMode === 'delivery') {
      if (!deliveryAddress.trim()) {
        formError = 'Por favor ingresa la dirección de entrega en Duitama.';
        cartStep = 'delivery_address';
        return;
      }
      if (!deliveryNeighborhood.trim()) {
        formError = 'Por favor ingresa el barrio de entrega en Duitama.';
        cartStep = 'delivery_neighborhood';
        return;
      }
    }

    const clientName = store.user?.full_name || customerName.trim() || 'Cliente Mías';
    const clientPhone = store.user?.phone || customerPhone.trim() || '3212520881';

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
        delivery_name: clientName,
        delivery_phone: clientPhone,
        delivery_address: selectedMode === 'delivery' ? `${deliveryAddress.trim()}, Barrio ${deliveryNeighborhood.trim()}` : '',
        delivery_notes: selectedMode === 'delivery' ? deliveryNotes.trim() : (selectedMode === 'in_store' ? `Mesa #${tableNumber}` : 'Para llevar'),
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
      customer_name: clientName,
      customer_phone: clientPhone,
      delivery_address: selectedMode === 'delivery' ? `${deliveryAddress.trim()}, Barrio ${deliveryNeighborhood.trim()}` : '',
      table_number: selectedMode === 'in_store' ? tableNumber.trim() : '',
      delivery_notes: selectedMode === 'delivery' ? deliveryNotes.trim() : '',
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
    let msg = `🍔 *NUEVA COMANDA — MÍAS HAMBURGUESAS (DUITAMA)* 🍔\n` +
      `*Orden:* #${createdOrder.code}\n` +
      `*Modalidad:* ${createdOrder.modeLabel}\n` +
      `*Cliente:* ${createdOrder.customer_name}` + (createdOrder.customer_phone ? ` (${createdOrder.customer_phone})` : '') + `\n`;

    if (createdOrder.delivery_address) {
      msg += `*Dirección:* ${createdOrder.delivery_address}\n`;
    }
    if (createdOrder.table_number) {
      msg += `*Mesa:* #${createdOrder.table_number}\n`;
    }
    if (createdOrder.delivery_notes) {
      msg += `*Notas:* ${createdOrder.delivery_notes}\n`;
    }
    msg += `*Pago:* ${createdOrder.payment_method}\n`;
    if (createdOrder.cash_change_amount > 0) {
      msg += `*Paga con:* ${formatCOP(createdOrder.cash_change_amount)} (Vueltas: ${formatCOP(createdOrder.cash_change_amount - createdOrder.total)})\n`;
    }
    msg += `\n*DETALLE DE PLATOS:*\n`;
    createdOrder.items.forEach(it => {
      msg += `• ${it.quantity}x ${it.product.name}`;
      if (it.combo) msg += ` + [Combo Papa + Gaseosa]`;
      if (it.extras && it.extras.length > 0) {
        msg += ` (${it.extras.map(e => e.name).join(', ')})`;
      }
      msg += ` — ${formatCOP(it.itemTotal)}\n`;
    });
    msg += `\n*TOTAL A PAGAR:* ${formatCOP(createdOrder.total)}\n`;
    msg += `\n📍 Sede Duitama, Boyacá • Por favor confirmar despacho`;

    const url = `https://api.whatsapp.com/send?phone=573212520881&text=${encodeURIComponent(msg)}`;
    window.open(url, '_blank');
  }

  function goToStampsFromTicket() {
    store.closeCart();
    store.currentView = 'stamps';
  }

  function closeOrderModal() {
    store.closeCart();
    cartStep = 'cart';
    createdOrder = null;
    formError = '';
  }
</script>

<!-- BOTÓN FLOTANTE INFERIOR (DISPARADOR) -->
{#if store.cartCount > 0 && !store.isCartOpen}
  <div 
    transition:scale={{ duration: 250, start: 0.85, easing: backOut }}
    class="fixed bottom-6 inset-x-0 z-40 flex justify-center px-4 pointer-events-none"
  >
    <button
      type="button"
      onclick={() => store.openCart()}
      class="pointer-events-auto min-h-[58px] px-6 rounded-full bg-[#0C0C0E] hover:bg-black active:scale-[0.97] text-white border border-white/10 shadow-[0_12px_40px_rgba(0,0,0,0.4)] flex items-center gap-3.5 transition-all cursor-pointer font-display font-black text-sm sm:text-base tracking-tight"
    >
      <div class="w-8 h-8 rounded-full bg-[#FF2400] text-white flex items-center justify-center shrink-0">
        <ShoppingBag class="w-4 h-4 stroke-[2.5]" />
      </div>
      <span>Ver Comanda</span>
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
    class="fixed inset-0 z-50 w-full h-full min-h-[100dvh] bg-[#FAF6EE] text-[#0C0C0E] overflow-y-auto overscroll-contain flex flex-col"
    style="touch-action: pan-y; -webkit-overflow-scrolling: touch; padding-top: max(1rem, env(safe-area-inset-top, 0px)); padding-bottom: max(3rem, calc(env(safe-area-inset-bottom, 0px) + 2rem)); padding-left: max(1rem, env(safe-area-inset-left, 0px)); padding-right: max(1rem, env(safe-area-inset-right, 0px));"
  >
    <div data-lenis-prevent class="w-full max-w-lg mx-auto py-2 sm:py-4 flex-1 flex flex-col space-y-5 sm:space-y-6">
      
      <!-- 1. Cabecera Unificada con Botón Volver y Botón Cerrar -->
      <header class="flex items-center justify-between w-full pb-1 shrink-0">
        {#if cartStep !== 'cart' && cartStep !== 'ticket'}
          <button 
            type="button"
            onclick={handleBackStep}
            class="w-12 h-12 sm:w-14 sm:h-14 rounded-full bg-white border border-black/10 flex items-center justify-center text-[#3A3A3C] hover:bg-black/5 active:scale-90 transition-all cursor-pointer group shrink-0 shadow-2xs"
            aria-label="Volver al paso anterior"
          >
            <ChevronLeft class="w-6 h-6 stroke-[2.5] -translate-x-0.5 group-hover:-translate-x-1 transition-transform" />
          </button>
        {:else}
          <div class="w-12 h-12 sm:w-14 sm:h-14 flex items-center justify-center shrink-0">
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
              Paso 1 • Modalidad de Pedido
            {:else if cartStep === 'delivery_address'}
              Paso 1 de 3 • Dirección
            {:else if cartStep === 'delivery_neighborhood'}
              Paso 2 de 3 • Barrio Duitama
            {:else if cartStep === 'delivery_payment'}
              Paso 3 de 3 • Pago y Envío
            {:else if cartStep === 'in_store_confirm'}
              Consumo en Salón • Mesa
            {:else if cartStep === 'takeout_confirm'}
              Para Llevar • Retiro
            {:else}
              Orden Procesada
            {/if}
          </span>
          <h1 class="font-display font-black text-xl sm:text-2xl text-[#0C0C0E] tracking-tight leading-tight">
            {#if cartStep === 'cart'}
              Tu Comanda
            {:else if cartStep === 'mode'}
              ¿Dónde Disfrutas?
            {:else if cartStep === 'delivery_address'}
              ¿A Dónde lo Llevamos?
            {:else if cartStep === 'delivery_neighborhood'}
              Barrio en Duitama
            {:else if cartStep === 'delivery_payment'}
              Forma de Pago
            {:else if cartStep === 'in_store_confirm'}
              Mesa en el Salón
            {:else if cartStep === 'takeout_confirm'}
              Retiro en Barra
            {:else}
              ¡Comanda Recibida!
            {/if}
          </h1>
        </div>

        <!-- Botón Circular Blanco Plano de Cierre -->
        <button 
          type="button"
          onclick={closeOrderModal}
          class="w-12 h-12 sm:w-14 sm:h-14 rounded-full bg-white border border-black/10 flex items-center justify-center text-[#0C0C0E] hover:bg-black/5 active:scale-90 transition-all cursor-pointer shrink-0 shadow-2xs"
          aria-label="Cerrar comanda"
        >
          <X class="w-5 h-5 sm:w-6 sm:h-6 stroke-[2.2]" />
        </button>
      </header>

      <!-- Mensaje de Error / Alerta si aplica -->
      {#if formError}
        <div class="p-4 rounded-2xl bg-red-50 border border-red-200 text-xs sm:text-sm font-bold text-[#FF2400] text-center shadow-xs">
          {formError}
        </div>
      {/if}

      <!-- ========================================================================= -->
      <!-- FASE 1: LISTADO DE PRODUCTOS Y UPSELLING                                  -->
      <!-- ========================================================================= -->
      {#if cartStep === 'cart'}
        {#if store.cart.length === 0}
          <div class="bg-white border border-black/10 rounded-[38px] p-8 sm:p-12 text-center space-y-4 shadow-none my-auto">
            <div class="w-20 h-20 rounded-full bg-[#EDE8DC] text-[#8E8E93] mx-auto flex items-center justify-center">
              <ShoppingBag class="w-10 h-10 stroke-[1.5]" />
            </div>
            <h2 class="font-display font-black text-2xl text-[#0C0C0E]">Tu bolsa está vacía</h2>
            <p class="text-sm text-[#8E8E93] max-w-xs mx-auto">
              Descubre nuestras hamburguesas al carbón con pan pretzel artesanal.
            </p>
            <button
              type="button"
              onclick={() => { closeOrderModal(); store.currentView = 'menu'; }}
              class="min-h-[52px] px-8 rounded-full bg-[#FF2400] text-white font-display font-extrabold text-sm hover:bg-[#E02000] active:scale-95 transition-all cursor-pointer"
            >
              Explorar la Carta
            </button>
          </div>
        {:else}
          <div class="space-y-4">
            <!-- Lista de Ítems en la Comanda -->
            <div class="space-y-3">
              {#each store.cart as item, index (item.product.id + '-' + index)}
                <div class="bg-white border border-black/10 rounded-[28px] p-4.5 sm:p-5 flex items-center justify-between gap-4 shadow-none">
                  <div class="flex items-center gap-3.5 min-w-0">
                    {#if item.product.image_url}
                      <img 
                        src={item.product.image_url} 
                        alt={item.product.name} 
                        class="w-16 h-16 sm:w-20 sm:h-20 object-contain rounded-2xl bg-[#FAF6EE] p-1 border border-black/5 shrink-0" 
                      />
                    {:else}
                      <div class="w-16 h-16 sm:w-20 sm:h-20 rounded-2xl bg-[#EDE8DC] flex items-center justify-center shrink-0">
                        <UtensilsCrossed class="w-6 h-6 text-[#8E8E93]" />
                      </div>
                    {/if}

                    <div class="min-w-0">
                      <h3 class="font-display font-extrabold text-base sm:text-lg text-[#0C0C0E] truncate">
                        {item.product.name}
                      </h3>
                      <div class="text-xs text-[#8E8E93] font-medium space-y-0.5">
                        {#if item.combo}
                          <p class="text-[#FF2400] font-bold">+ Combo Papa + Gaseosa</p>
                        {/if}
                        {#if item.extras && item.extras.length > 0}
                          <p class="truncate">{item.extras.map(e => e.name).join(', ')}</p>
                        {/if}
                      </div>
                      <span class="font-mono text-sm sm:text-base font-bold text-[#FF2400] mt-1 block">
                        {formatCOP(item.itemTotal)}
                      </span>
                    </div>
                  </div>

                  <!-- Controles de Cantidad -->
                  <div class="flex items-center gap-2 shrink-0">
                    <button
                      type="button"
                      onclick={() => store.updateQuantity(index, item.quantity - 1)}
                      class="w-9 h-9 rounded-full bg-[#EDE8DC] hover:bg-black/10 active:scale-90 text-[#0C0C0E] flex items-center justify-center transition-all cursor-pointer"
                      aria-label="Disminuir"
                    >
                      <Minus class="w-4 h-4 stroke-[2.5]" />
                    </button>

                    <span class="w-7 text-center font-mono font-bold text-sm text-[#0C0C0E]">
                      {item.quantity}
                    </span>

                    <button
                      type="button"
                      onclick={() => store.updateQuantity(index, item.quantity + 1)}
                      class="w-9 h-9 rounded-full bg-[#FF2400] hover:bg-[#E02000] active:scale-90 text-white flex items-center justify-center transition-all cursor-pointer shadow-xs"
                      aria-label="Aumentar"
                    >
                      <Plus class="w-4 h-4 stroke-[2.5]" />
                    </button>
                  </div>
                </div>
              {/each}
            </div>

            <!-- Sugerencias de Acompañamiento (Upsell) -->
            <div class="bg-white border border-black/10 rounded-[32px] p-5 space-y-3 shadow-none">
              <div class="flex items-center gap-2">
                <Sparkles class="w-4 h-4 text-[#FF2400]" />
                <h4 class="font-display font-extrabold text-sm text-[#0C0C0E]">
                  ¿Deseas acompañar tu pedido?
                </h4>
              </div>

              <div class="grid grid-cols-3 gap-2.5">
                {#each upsellItems as up}
                  <button
                    type="button"
                    onclick={() => handleAddUpsell(up)}
                    class="p-3 rounded-2xl border border-black/10 bg-[#FAF6EE] hover:bg-white text-center transition-all active:scale-95 cursor-pointer flex flex-col items-center justify-between min-h-[105px] group relative"
                  >
                    {#if up.image_url}
                      <img src={up.image_url} alt={up.name} class="w-9 h-9 object-contain group-hover:scale-110 transition-transform" />
                    {:else}
                      <div class="w-9 h-9 rounded-full bg-[#EDE8DC] flex items-center justify-center text-xs">🍟</div>
                    {/if}
                    <div>
                      <span class="text-[11px] font-bold text-[#0C0C0E] block leading-tight truncate max-w-full">
                        {up.name}
                      </span>
                      <span class="text-[10px] font-mono font-bold text-[#FF2400]">
                        +{formatCOP(up.price)}
                      </span>
                    </div>
                  </button>
                {/each}
              </div>
            </div>

            <!-- Resumen y Botón Continuar a Modalidad -->
            <div class="bg-white border border-black/10 rounded-[34px] p-6 space-y-4 shadow-none">
              <div class="flex justify-between items-center text-lg">
                <span class="font-display font-extrabold text-[#0C0C0E]">Subtotal:</span>
                <span class="font-display font-black text-2xl text-[#FF2400]">
                  {formatCOP(store.cartSubtotal)}
                </span>
              </div>

              <button
                type="button"
                onclick={handleContinueToMode}
                class="w-full min-h-[64px] rounded-full bg-[#FF2400] hover:bg-[#E02000] active:scale-[0.98] text-white font-display font-extrabold text-lg sm:text-xl tracking-tight transition-all flex items-center justify-center gap-2.5 cursor-pointer shadow-sm"
              >
                <span>Continuar con el Pedido</span>
                <ArrowRight class="w-5 h-5 stroke-[2.5]" />
              </button>
            </div>
          </div>
        {/if}

      <!-- ========================================================================= -->
      <!-- FASE 2: SELECCIÓN DE MODALIDAD ("¿DÓNDE DISFRUTAS?")                      -->
      <!-- 3 OPCIONES EN PANTALLA COMPLETA CON ESPACIO AMPLIO Y CERO COLISIÓN       -->
      <!-- ========================================================================= -->
      {:else if cartStep === 'mode'}
        <div class="flex-1 flex flex-col justify-between gap-4 py-2 min-h-[60vh] sm:min-h-[65vh]">
          
          <!-- Opción 1: A Domicilio -->
          <button
            type="button"
            onclick={() => handleSelectMode('delivery')}
            class="w-full flex-1 min-h-[125px] sm:min-h-[145px] text-left bg-white rounded-[32px] sm:rounded-[36px] border-2 p-5 sm:p-6 transition-all active:scale-[0.985] cursor-pointer flex items-center gap-5 sm:gap-6 shadow-2xs hover:shadow-md {selectedMode === 'delivery' ? 'border-[#FF2400] bg-white' : 'border-black/10 hover:border-black/20'}"
          >
            <div class="w-16 h-16 sm:w-20 sm:h-20 rounded-2xl bg-[#FF2400] text-white flex items-center justify-center shrink-0 shadow-sm">
              <Bike class="w-8 h-8 sm:w-9 sm:h-9 stroke-[2.2]" />
            </div>
            <div class="flex-1 min-w-0 pr-1">
              <div class="flex items-center justify-between gap-2">
                <h2 class="font-display font-black text-xl sm:text-2xl text-[#0C0C0E]">A Domicilio</h2>
                <span class="px-3 py-1 rounded-full bg-[#FF2400]/10 text-xs sm:text-sm font-mono font-bold text-[#FF2400] shrink-0">
                  + $6.000
                </span>
              </div>
              <p class="text-xs sm:text-sm text-[#8E8E93] mt-1.5 leading-relaxed">
                Directo a tu puerta en Duitama, empacado hermético con sello de calor.
              </p>
            </div>
          </button>

          <!-- Opción 2: Para Llevar -->
          <button
            type="button"
            onclick={() => handleSelectMode('takeout')}
            class="w-full flex-1 min-h-[125px] sm:min-h-[145px] text-left bg-white rounded-[32px] sm:rounded-[36px] border-2 p-5 sm:p-6 transition-all active:scale-[0.985] cursor-pointer flex items-center gap-5 sm:gap-6 shadow-2xs hover:shadow-md {selectedMode === 'takeout' ? 'border-[#FF2400] bg-white' : 'border-black/10 hover:border-black/20'}"
          >
            <div class="w-16 h-16 sm:w-20 sm:h-20 rounded-2xl bg-[#0C0C0E] text-white flex items-center justify-center shrink-0 shadow-sm">
              <ShoppingBag class="w-8 h-8 sm:w-9 sm:h-9 stroke-[2.2]" />
            </div>
            <div class="flex-1 min-w-0 pr-1">
              <div class="flex items-center justify-between gap-2">
                <h2 class="font-display font-black text-xl sm:text-2xl text-[#0C0C0E]">Para Llevar</h2>
                <span class="px-3 py-1 rounded-full bg-[#EDE8DC] text-xs sm:text-sm font-mono font-bold text-[#0C0C0E] shrink-0">
                  Sin recargo
                </span>
              </div>
              <p class="text-xs sm:text-sm text-[#8E8E93] mt-1.5 leading-relaxed">
                Empacado para retirar en barra en nuestro punto físico sin filas.
              </p>
            </div>
          </button>

          <!-- Opción 3: Consumo en Salón -->
          <button
            type="button"
            onclick={() => handleSelectMode('in_store')}
            class="w-full flex-1 min-h-[125px] sm:min-h-[145px] text-left bg-white rounded-[32px] sm:rounded-[36px] border-2 p-5 sm:p-6 transition-all active:scale-[0.985] cursor-pointer flex items-center gap-5 sm:gap-6 shadow-2xs hover:shadow-md {selectedMode === 'in_store' ? 'border-[#FF2400] bg-white' : 'border-black/10 hover:border-black/20'}"
          >
            <div class="w-16 h-16 sm:w-20 sm:h-20 rounded-2xl bg-[#EDE8DC] text-[#0C0C0E] flex items-center justify-center shrink-0 shadow-sm">
              <UtensilsCrossed class="w-8 h-8 sm:w-9 sm:h-9 stroke-[2.2]" />
            </div>
            <div class="flex-1 min-w-0 pr-1">
              <div class="flex items-center justify-between gap-2">
                <h2 class="font-display font-black text-xl sm:text-2xl text-[#0C0C0E]">En el Salón</h2>
                <span class="px-3 py-1 rounded-full bg-emerald-100 text-xs sm:text-sm font-mono font-bold text-emerald-800 shrink-0">
                  En mesa • Sin datos
                </span>
              </div>
              <p class="text-xs sm:text-sm text-[#8E8E93] mt-1.5 leading-relaxed">
                Disfruta la experiencia en nuestro punto físico, servido directo a tu mesa.
              </p>
            </div>
          </button>

        </div>

      <!-- ========================================================================= -->
      <!-- FASE 3A: DOMICILIO - PANTALLA 1: DIRECCIÓN (FAT-FINGER INTERACTIVO)       -->
      <!-- ========================================================================= -->
      {:else if cartStep === 'delivery_address'}
        <div class="flex-1 flex flex-col justify-between space-y-6 py-2">
          <div class="bg-white border border-black/10 rounded-[36px] p-6 sm:p-7 space-y-5 shadow-none">
            <div class="w-14 h-14 rounded-2xl bg-[#FF2400]/10 text-[#FF2400] flex items-center justify-center">
              <MapPin class="w-7 h-7 stroke-[2.2]" />
            </div>
            <div>
              <h2 class="font-display font-black text-2xl sm:text-3xl text-[#0C0C0E] leading-tight">
                ¿A dónde te lo llevamos?
              </h2>
              <p class="text-xs sm:text-sm text-[#8E8E93] mt-1">
                Ingresa tu dirección exacta en Duitama para asignarte el repartidor más cercano.
              </p>
            </div>

            <div class="pt-2">
              <label for="delivery-addr" class="text-xs font-mono font-bold text-[#8E8E93] uppercase block mb-2">
                Dirección exacta
              </label>
              <input 
                id="delivery-addr"
                type="text" 
                bind:value={deliveryAddress}
                placeholder="ej. Carrera 16 # 14-32"
                class="w-full min-h-[72px] rounded-3xl px-6 bg-[#FAF6EE] border-2 border-black/10 focus:border-[#FF2400] text-[#0C0C0E] text-lg sm:text-xl font-display font-bold focus:outline-none shadow-xs transition-colors"
              />
            </div>
          </div>

          <button
            type="button"
            onclick={() => {
              if (!deliveryAddress.trim()) {
                formError = 'Por favor ingresa tu dirección de entrega en Duitama.';
                return;
              }
              formError = '';
              cartStep = 'delivery_neighborhood';
            }}
            class="w-full min-h-[64px] rounded-full bg-[#FF2400] hover:bg-[#E02000] active:scale-[0.98] text-white font-display font-extrabold text-lg sm:text-xl tracking-tight transition-all flex items-center justify-center gap-2.5 cursor-pointer shadow-sm"
          >
            <span>Continuar al Barrio</span>
            <ArrowRight class="w-5 h-5 stroke-[2.5]" />
          </button>
        </div>

      <!-- ========================================================================= -->
      <!-- FASE 3B: DOMICILIO - PANTALLA 2: BARRIO Y NOTAS                           -->
      <!-- ========================================================================= -->
      {:else if cartStep === 'delivery_neighborhood'}
        <div class="flex-1 flex flex-col justify-between space-y-6 py-2">
          <div class="bg-white border border-black/10 rounded-[36px] p-6 sm:p-7 space-y-5 shadow-none">
            <div class="w-14 h-14 rounded-2xl bg-amber-500/10 text-amber-600 flex items-center justify-center">
              <MapPin class="w-7 h-7 stroke-[2.2]" />
            </div>
            <div>
              <h2 class="font-display font-black text-2xl sm:text-3xl text-[#0C0C0E] leading-tight">
                ¿En qué barrio estás?
              </h2>
              <p class="text-xs sm:text-sm text-[#8E8E93] mt-1">
                Toca tu sector o escribe el nombre para optimizar la ruta de entrega.
              </p>
            </div>

            <div class="space-y-4 pt-1">
              <input 
                type="text" 
                bind:value={deliveryNeighborhood}
                placeholder="ej. El Sol, Cándido, Centro"
                class="w-full min-h-[68px] rounded-3xl px-6 bg-[#FAF6EE] border-2 border-black/10 focus:border-[#FF2400] text-[#0C0C0E] text-lg font-display font-bold focus:outline-none shadow-xs transition-colors"
              />

              <!-- Chips Táctiles de Barrios en Duitama -->
              <div class="flex flex-wrap gap-2">
                {#each duitamaNeighborhoods as neigh}
                  <button
                    type="button"
                    onclick={() => deliveryNeighborhood = neigh}
                    class="min-h-[42px] px-4 rounded-full text-xs font-mono font-bold border transition-all cursor-pointer active:scale-95 {deliveryNeighborhood === neigh ? 'bg-[#FF2400] text-white border-[#FF2400]' : 'bg-[#EDE8DC] text-[#0C0C0E] border-black/10 hover:border-black/25'}"
                  >
                    {neigh}
                  </button>
                {/each}
              </div>

              <div class="pt-2">
                <label for="delivery-notes" class="text-xs font-mono font-bold text-[#8E8E93] uppercase block mb-1.5">
                  Indicaciones para el repartidor (opcional)
                </label>
                <input 
                  id="delivery-notes"
                  type="text" 
                  bind:value={deliveryNotes}
                  placeholder="ej. Apto 302, timbre blanco, portón negro"
                  class="w-full min-h-[58px] rounded-2xl px-5 bg-[#FAF6EE] border border-black/10 text-[#0C0C0E] text-sm font-sans focus:outline-none focus:border-[#FF2400]"
                />

                <!-- Chips Rápidos de Notas -->
                <div class="flex flex-wrap gap-1.5 pt-2">
                  {#each ['🔔 Tocar timbre', '🏢 Dejar en portería', '🛵 Llamar al llegar'] as chip}
                    <button
                      type="button"
                      onclick={() => toggleQuickNote(chip)}
                      class="px-3 py-1.5 rounded-full text-xs font-medium border transition-all cursor-pointer {deliveryNotes.includes(chip) ? 'bg-[#0C0C0E] text-white border-[#0C0C0E]' : 'bg-[#EDE8DC] text-[#0C0C0E] border-black/10'}"
                    >
                      {chip}
                    </button>
                  {/each}
                </div>
              </div>
            </div>
          </div>

          <button
            type="button"
            onclick={() => {
              if (!deliveryNeighborhood.trim()) {
                formError = 'Por favor ingresa tu barrio en Duitama.';
                return;
              }
              formError = '';
              cartStep = 'delivery_payment';
            }}
            class="w-full min-h-[64px] rounded-full bg-[#FF2400] hover:bg-[#E02000] active:scale-[0.98] text-white font-display font-extrabold text-lg sm:text-xl tracking-tight transition-all flex items-center justify-center gap-2.5 cursor-pointer shadow-sm"
          >
            <span>Continuar al Pago</span>
            <ArrowRight class="w-5 h-5 stroke-[2.5]" />
          </button>
        </div>

      <!-- ========================================================================= -->
      <!-- FASE 3C: DOMICILIO - PANTALLA 3: MÉTODO DE PAGO Y CONFIRMACIÓN            -->
      <!-- ========================================================================= -->
      {:else if cartStep === 'delivery_payment'}
        <div class="flex-1 flex flex-col justify-between space-y-5 py-2">
          
          <!-- Banner si el usuario aún no tiene cuenta -->
          {#if !store.user}
            <div class="bg-amber-500/10 border-2 border-amber-500/30 rounded-[28px] p-4.5 flex items-center justify-between gap-3 shadow-2xs">
              <div class="flex items-center gap-3 min-w-0">
                <div class="w-10 h-10 rounded-full bg-[#FF2400] text-white flex items-center justify-center shrink-0">
                  <Sparkles class="w-5 h-5 stroke-[2]" />
                </div>
                <div class="min-w-0">
                  <h4 class="font-display font-bold text-sm text-[#0C0C0E]">¿Aún no tienes cuenta?</h4>
                  <p class="text-xs text-[#8E8E93]">Crea tu cuenta antes de enviar para acumular sellos.</p>
                </div>
              </div>
              <button
                type="button"
                onclick={() => store.isAuthModalOpen = true}
                class="min-h-[42px] px-3.5 py-1.5 rounded-full bg-[#FF2400] text-white text-xs font-display font-bold shrink-0 hover:bg-[#E02000] active:scale-95 cursor-pointer"
              >
                Crear Cuenta
              </button>
            </div>
          {/if}

          <!-- Selección de Método de Pago -->
          <div class="bg-white border border-black/10 rounded-[34px] p-5 sm:p-6 space-y-4 shadow-none">
            <h2 class="font-display font-black text-xl text-[#0C0C0E]">
              ¿Cómo deseas pagar?
            </h2>

            <div class="grid grid-cols-1 sm:grid-cols-3 gap-2.5">
              <!-- Efectivo -->
              <button
                type="button"
                onclick={() => paymentMethod = 'Efectivo'}
                class="p-4 rounded-2xl border-2 text-left transition-all cursor-pointer flex flex-col justify-between min-h-[88px] {paymentMethod === 'Efectivo' ? 'border-[#FF2400] bg-[#FAF6EE]' : 'border-black/10 bg-white'}"
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
                class="p-4 rounded-2xl border-2 text-left transition-all cursor-pointer flex flex-col justify-between min-h-[88px] {paymentMethod === 'Nequi / Daviplata' ? 'border-[#FF2400] bg-[#FAF6EE]' : 'border-black/10 bg-white'}"
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
                class="p-4 rounded-2xl border-2 text-left transition-all cursor-pointer flex flex-col justify-between min-h-[88px] {paymentMethod === 'Datáfono contraentrega' ? 'border-[#FF2400] bg-[#FAF6EE]' : 'border-black/10 bg-white'}"
              >
                <CreditCard class="w-5 h-5 text-[#FF2400]" />
                <div>
                  <span class="font-display font-extrabold text-sm text-[#0C0C0E] block">Datáfono</span>
                  <span class="text-[11px] text-[#8E8E93]">Tarjeta en puerta</span>
                </div>
              </button>
            </div>

            <!-- Detalles según método -->
            {#if paymentMethod === 'Efectivo'}
              <div class="p-4 rounded-2xl bg-[#EDE8DC] space-y-2.5">
                <span class="text-xs font-mono font-bold text-[#0C0C0E] block">
                  ¿Con cuánto cancelas? (Para llevar tu cambio exacto)
                </span>
                <div class="grid grid-cols-3 gap-2">
                  <button
                    type="button"
                    onclick={() => selectQuickCash(50000)}
                    class="min-h-[44px] rounded-xl bg-white border border-black/10 text-xs font-bold font-mono text-[#0C0C0E] hover:border-[#FF2400] cursor-pointer"
                  >
                    $50.000
                  </button>
                  <button
                    type="button"
                    onclick={() => selectQuickCash(100000)}
                    class="min-h-[44px] rounded-xl bg-white border border-black/10 text-xs font-bold font-mono text-[#0C0C0E] hover:border-[#FF2400] cursor-pointer"
                  >
                    $100.000
                  </button>
                  <button
                    type="button"
                    onclick={() => selectQuickCash('exact')}
                    class="min-h-[44px] rounded-xl bg-white border border-black/10 text-xs font-bold font-mono text-[#0C0C0E] hover:border-[#FF2400] cursor-pointer"
                  >
                    Exacto
                  </button>
                </div>
                {#if cashChange > 0}
                  <span class="text-xs font-mono font-bold text-emerald-800 bg-white p-2.5 rounded-xl border border-black/10 block text-center">
                    Vueltas a devolver: {formatCOP(cashChange)}
                  </span>
                {/if}
              </div>
            {:else if paymentMethod === 'Nequi / Daviplata'}
              <div class="p-4 rounded-2xl bg-[#EDE8DC] flex items-center justify-between">
                <div>
                  <span class="text-[10px] font-mono uppercase text-[#8E8E93] block font-bold">Nequi Oficial Duitama</span>
                  <span class="font-display font-black text-lg text-[#0C0C0E]">321 252 0881</span>
                </div>
                <button
                  type="button"
                  onclick={copyNequi}
                  class="min-h-[44px] px-4 rounded-xl bg-white border border-black/10 text-xs font-mono font-bold text-[#0C0C0E] hover:border-[#FF2400] cursor-pointer flex items-center gap-1.5"
                >
                  <Copy class="w-3.5 h-3.5" />
                  <span>{isCopied ? '¡Copiado!' : 'Copiar'}</span>
                </button>
              </div>
            {/if}
          </div>

          <!-- Resumen de Costos y Botón Finalizar -->
          <div class="bg-white border border-black/10 rounded-[34px] p-6 space-y-4 shadow-none">
            <div class="space-y-1.5 text-sm">
              <div class="flex justify-between text-[#8E8E93]">
                <span>Subtotal de comanda:</span>
                <span class="font-mono font-medium text-[#0C0C0E]">{formatCOP(store.cartSubtotal)}</span>
              </div>
              <div class="flex justify-between text-[#8E8E93]">
                <span>Domicilio Duitama:</span>
                <span class="font-mono font-medium text-[#0C0C0E]">{formatCOP(currentDeliveryFee)}</span>
              </div>
              <div class="border-t border-black/10 pt-2 flex items-center justify-between">
                <span class="font-display font-black text-xl text-[#0C0C0E]">Total a Pagar</span>
                <span class="font-display font-black text-2xl sm:text-3xl text-[#FF2400]">
                  {formatCOP(currentTotal)}
                </span>
              </div>
            </div>

            <button
              type="button"
              onclick={handleFinalizeOrder}
              class="w-full min-h-[68px] rounded-full bg-[#FF2400] hover:bg-[#E02000] active:scale-[0.98] text-white font-display font-extrabold text-lg sm:text-xl tracking-tight transition-all flex items-center justify-center gap-2.5 cursor-pointer shadow-sm"
            >
              <span>Confirmar Pedido a Domicilio</span>
              <Check class="w-6 h-6 stroke-[2.5]" />
            </button>
          </div>
        </div>

      <!-- ========================================================================= -->
      <!-- FASE 3D: CONSUMO EN SALÓN (SIN PEDIR DATOS PERSONALES)                     -->
      <!-- ========================================================================= -->
      {:else if cartStep === 'in_store_confirm'}
        <div class="flex-1 flex flex-col justify-between space-y-5 py-2">
          
          <div class="bg-white border border-black/10 rounded-[36px] p-6 sm:p-7 space-y-5 shadow-none">
            <div class="w-14 h-14 rounded-2xl bg-emerald-100 text-emerald-800 flex items-center justify-center">
              <UtensilsCrossed class="w-7 h-7 stroke-[2.2]" />
            </div>
            <div>
              <h2 class="font-display font-black text-2xl sm:text-3xl text-[#0C0C0E] leading-tight">
                Consumo en el Salón
              </h2>
              <p class="text-xs sm:text-sm text-[#8E8E93] mt-1">
                Servicio directo a tu mesa en nuestro punto de Duitama. No te pediremos direcciones ni formularios.
              </p>
            </div>

            <!-- Selector Rápido de Mesa (Opcional) -->
            <div class="space-y-2 pt-1">
              <span class="text-xs font-mono font-bold text-[#8E8E93] uppercase block">
                Selecciona tu Mesa
              </span>
              <div class="grid grid-cols-5 gap-2">
                {#each [1, 2, 3, 4, 5, 6, 7, 8, 9, 10] as n}
                  <button
                    type="button"
                    onclick={() => tableNumber = String(n)}
                    class="min-h-[48px] rounded-2xl font-mono font-bold text-sm border transition-all cursor-pointer active:scale-95 {tableNumber === String(n) ? 'bg-[#0C0C0E] text-white border-[#0C0C0E]' : 'bg-[#FAF6EE] text-[#0C0C0E] border-black/10 hover:border-black/30'}"
                  >
                    #{n}
                  </button>
                {/each}
              </div>
            </div>

            <!-- Método de Pago en Mesa -->
            <div class="space-y-2 pt-2">
              <span class="text-xs font-mono font-bold text-[#8E8E93] uppercase block">
                Forma de Pago en Mesa
              </span>
              <div class="grid grid-cols-3 gap-2">
                {#each ['Efectivo', 'Nequi / Daviplata', 'Datáfono'] as method}
                  <button
                    type="button"
                    onclick={() => paymentMethod = method}
                    class="min-h-[48px] p-2 rounded-xl text-xs font-mono font-bold border transition-all cursor-pointer text-center {paymentMethod === method ? 'bg-[#FF2400] text-white border-[#FF2400]' : 'bg-[#FAF6EE] text-[#0C0C0E] border-black/10'}"
                  >
                    {method}
                  </button>
                {/each}
              </div>
            </div>
          </div>

          <!-- Banner si el usuario no tiene cuenta -->
          {#if !store.user}
            <div class="bg-amber-500/10 border border-amber-500/30 rounded-[28px] p-4 flex items-center justify-between gap-3 shadow-2xs">
              <div class="flex items-center gap-2.5">
                <Sparkles class="w-5 h-5 text-[#FF2400] shrink-0" />
                <span class="text-xs font-bold text-[#0C0C0E]">Crea tu cuenta para acumular tus sellos</span>
              </div>
              <button
                type="button"
                onclick={() => store.isAuthModalOpen = true}
                class="min-h-[38px] px-3.5 rounded-full bg-[#FF2400] text-white text-xs font-bold cursor-pointer shrink-0"
              >
                Crear Cuenta
              </button>
            </div>
          {/if}

          <!-- Botón Confirmar Comanda en Mesa -->
          <div class="bg-white border border-black/10 rounded-[34px] p-6 space-y-4 shadow-none">
            <div class="flex justify-between items-center">
              <span class="font-display font-extrabold text-base text-[#0C0C0E]">Total Comanda:</span>
              <span class="font-display font-black text-2xl sm:text-3xl text-[#FF2400]">
                {formatCOP(store.cartSubtotal)}
              </span>
            </div>

            <button
              type="button"
              onclick={handleFinalizeOrder}
              class="w-full min-h-[64px] rounded-full bg-[#0C0C0E] hover:bg-black active:scale-[0.98] text-white font-display font-extrabold text-lg sm:text-xl tracking-tight transition-all flex items-center justify-center gap-2.5 cursor-pointer shadow-sm"
            >
              <span>Enviar a Cocina (Mesa #{tableNumber})</span>
              <Check class="w-5 h-5 stroke-[2.5]" />
            </button>
          </div>
        </div>

      <!-- ========================================================================= -->
      <!-- FASE 3E: PARA LLEVAR (TAKE AWAY - RETIRO EN BARRA)                        -->
      <!-- ========================================================================= -->
      {:else if cartStep === 'takeout_confirm'}
        <div class="flex-1 flex flex-col justify-between space-y-5 py-2">
          
          <div class="bg-white border border-black/10 rounded-[36px] p-6 sm:p-7 space-y-5 shadow-none">
            <div class="w-14 h-14 rounded-2xl bg-[#0C0C0E] text-white flex items-center justify-center">
              <ShoppingBag class="w-7 h-7 stroke-[2.2]" />
            </div>
            <div>
              <h2 class="font-display font-black text-2xl sm:text-3xl text-[#0C0C0E] leading-tight">
                Para Llevar (Take Away)
              </h2>
              <p class="text-xs sm:text-sm text-[#8E8E93] mt-1">
                Empacamos herméticamente tu comanda para que retires en barra sin hacer filas.
              </p>
            </div>

            <div class="space-y-2 pt-2">
              <span class="text-xs font-mono font-bold text-[#8E8E93] uppercase block">
                Forma de Pago al Retirar
              </span>
              <div class="grid grid-cols-3 gap-2">
                {#each ['Efectivo', 'Nequi / Daviplata', 'Datáfono'] as method}
                  <button
                    type="button"
                    onclick={() => paymentMethod = method}
                    class="min-h-[48px] p-2 rounded-xl text-xs font-mono font-bold border transition-all cursor-pointer text-center {paymentMethod === method ? 'bg-[#FF2400] text-white border-[#FF2400]' : 'bg-[#FAF6EE] text-[#0C0C0E] border-black/10'}"
                  >
                    {method}
                  </button>
                {/each}
              </div>
            </div>
          </div>

          <!-- Banner si el usuario no tiene cuenta -->
          {#if !store.user}
            <div class="bg-amber-500/10 border border-amber-500/30 rounded-[28px] p-4 flex items-center justify-between gap-3 shadow-2xs">
              <div class="flex items-center gap-2.5">
                <Sparkles class="w-5 h-5 text-[#FF2400] shrink-0" />
                <span class="text-xs font-bold text-[#0C0C0E]">Crea tu cuenta para acumular tus sellos</span>
              </div>
              <button
                type="button"
                onclick={() => store.isAuthModalOpen = true}
                class="min-h-[38px] px-3.5 rounded-full bg-[#FF2400] text-white text-xs font-bold cursor-pointer shrink-0"
              >
                Crear Cuenta
              </button>
            </div>
          {/if}

          <!-- Botón Confirmar Retiro -->
          <div class="bg-white border border-black/10 rounded-[34px] p-6 space-y-4 shadow-none">
            <div class="flex justify-between items-center">
              <span class="font-display font-extrabold text-base text-[#0C0C0E]">Total Comanda:</span>
              <span class="font-display font-black text-2xl sm:text-3xl text-[#FF2400]">
                {formatCOP(store.cartSubtotal)}
              </span>
            </div>

            <button
              type="button"
              onclick={handleFinalizeOrder}
              class="w-full min-h-[64px] rounded-full bg-[#0C0C0E] hover:bg-black active:scale-[0.98] text-white font-display font-extrabold text-lg sm:text-xl tracking-tight transition-all flex items-center justify-center gap-2.5 cursor-pointer shadow-sm"
            >
              <span>Confirmar Pedido para Llevar</span>
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
