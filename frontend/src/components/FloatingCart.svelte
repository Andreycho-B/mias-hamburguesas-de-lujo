<script>
import { store } from '../lib/store.svelte.js';
import { createOrderAPI } from '../lib/api.js';
import { onDestroy } from 'svelte';
  import {
    X,
    Plus,
    Minus,
    ShoppingBag,
    Sparkles,
    ArrowRight,
    Check,
    ChevronLeft,
    ChevronRight,
    ChevronUp,
    Sandwich,
    Drumstick,
    CookingPot,
    CupSoda,
    Star,
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

  // Estado de sugeridos ("¿Algo más?")
  let showCategorySheet = $state(false);

  // Datos del cliente para la orden
  let customerName = $state(store.user?.full_name || '');
  let customerPhone = $state(store.user?.phone || '');
  let deliveryAddress = $state(store.user?.address || '');
  let deliveryNeighborhood = $state('');
  let deliveryNotes = $state('');
  let tableNumber = $state('1');

  // Método de pago
  let paymentMethod = $state('Efectivo');
  let cashAmount = $state('');
  let formError = $state('');

  const deliveryFee = 6000;
  let createdOrder = $state(null);
  let isCopied = $state(false);

  // ═══════════════════════════════════════════════════════════════════════
  // CURADURÍA DE SUGERIDOS ("¿ALGO MÁS?") — lista única, datos vivos
  // Los nombres, precios, stock e imágenes canónicas se resuelven desde
  // store.products (PostgreSQL); aquí solo vive la curaduría: IDs y nombre
  // corto de chip. Así el precio visible y el cobrado no pueden divergir
  // (PAP-002: reutilizar antes de duplicar).
  // ═══════════════════════════════════════════════════════════════════════
  const UPSELL_FEATURED = [
    { id: 14, shortName: 'Coca-Cola', type: 'beverage' },
    { id: 17, shortName: 'Papa Francesa', type: 'fries' },
    { id: 15, shortName: 'Bretaña', type: 'beverage' },
    { id: 13, shortName: 'Botella de Agua', type: 'beverage' },
    { id: 10, shortName: 'Salchipapa', type: 'house' }
  ];

  // Nombre corto por ID para los chips de acompañamientos en escena
  // (cubre también ítems fuera de la ruleta, ej. Papa Nativa o extras).
  const UPSELL_SHORT_NAMES = new Map([
    [14, 'Coca-Cola'],
    [17, 'Papa Francesa'],
    [15, 'Bretaña'],
    [13, 'Botella de Agua'],
    [10, 'Salchipapa'],
    [18, 'Papa Nativa'],
    [16, 'Combo Clásico'],
    [11, 'Choripapa'],
    [19, 'Tocineta Extra'],
    [21, 'Queso Extra'],
    [23, 'Salsa Extra'],
    [20, 'Carne Extra'],
    [22, 'Pepinillos']
  ]);

  // Imágenes de respaldo para productos sin foto en DB (combos y adicionales
  // de la categoría 5). Clave por slug: estable ante cambios de ID o precio.
  const UPSELL_IMAGE_FALLBACK = {
    'combo-papa-gaseosa': '/imagenes/de_la_casa/papas_francesas.webp',
    'papa-francesa': '/imagenes/de_la_casa/papas_francesas.webp',
    'papa-nativa': '/imagenes/de_la_casa/papas_francesas.webp',
    'tocineta-extra': '/imagenes/de_la_casa/papas_francesas.webp',
    'queso-extra': '/imagenes/de_la_casa/salchi_papa.png',
    'salsa-de-la-casa': '/imagenes/de_la_casa/salchi_papa.png',
    'carne-extra': '/imagenes/hamburguesas/Orignal.webp',
    'pepinillos-extra': '/imagenes/hamburguesas/Crispy.webp'
  };

  function normalizeUpsellImage(url) {
    return (url || '').replace('papas_francesas.jpg', 'papas_francesas.webp');
  }

  // Resuelve una referencia curada contra el catálogo vivo. Retorna null si
  // el producto no existe o está agotado (respeta el toggle de stock admin).
  function resolveUpsellItem(ref) {
    const p = store.products.find((prod) => prod.id === ref.id);
    if (!p || p.is_in_stock === false) return null;
    return {
      id: p.id,
      category_id: p.category_id,
      name: p.name,
      shortName: ref.shortName,
      price: p.price,
      image_url: normalizeUpsellImage(p.image_url || UPSELL_IMAGE_FALLBACK[p.slug]),
      type: ref.type
    };
  }

  let upsellList = $derived(
    UPSELL_FEATURED.map(resolveUpsellItem).filter((it) => it !== null)
  );

  // ═══════════════════════════════════════════════════════════════════════
  // DESGLOSE REACTIVO DEL ESCENARIO HORIZONTAL (ORDEN CRONOLÓGICO)
  // Regla de distribución:
  // 1º producto: Centro y de primeras claramente
  // 2º producto: Derecha detrás (80% visible)
  // 3º producto: Izquierda detrás (80% visible por la izquierda)
  // ... sucesivamente hacia los márgenes, respetando tamaño y capa.
  // ═══════════════════════════════════════════════════════════════════════
  // Orden de exhibición = orden de agregado (estable: agregar nunca mueve
  // lo existente, solo anexa al margen). Las bebidas, delgadas, abren más
  // lateralmente vía `spread` en el template para no quedar tras el héroe.
  let allSelectedSceneUnits = $derived.by(() => {
    return store.cart.map((item) => {
      const p = item.product || {};
      const fallbackImg = normalizeUpsellImage(p.image_url || UPSELL_IMAGE_FALLBACK[p.slug]);
      return {
        id: item.id,
        productId: p.id,
        name: p.name || 'Producto',
        image_url: fallbackImg,
        price: p.price || 0,
        quantity: item.quantity || 1,
        category_id: p.category_id
      };
    });
  });

  // Chips interactivos de acompañamientos debajo del escenario
  let upsellSceneChips = $derived.by(() => {
    return store.cart.filter(
      (it) => it.product?.category_id === 4 || it.product?.category_id === 5 || it.product?.category_id === 3
    ).map((it) => {
      const short = UPSELL_SHORT_NAMES.get(it.product?.id ?? it.productId);
      const chipLabel = short || (it.product?.name?.replace(/400ml|en Vidrio|Manantial|Porción|Crujiente|Tradicional/g, '').trim() || it.product?.name);
      return {
        id: it.id,
        productId: it.product?.id,
        name: it.product?.name || 'Acompañamiento',
        shortName: chipLabel,
        quantity: it.quantity || 1,
        price: it.totalPrice
      };
    });
  });

  // Cuenta en tiempo real: sincronizada 100% con store.cart
  let currentUpsellTotal = $derived(store.cartSubtotal);

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
  // RULETA DE PRODUCTOS SUGERIDOS EN MEDIA LUNA (Inercia & Gestos)
  // ═══════════════════════════════════════════════════════════════════════
  let wheelPosition = $state(0); // Posición virtual continua en unidades de slot
  let activeWheelIndex = $derived.by(() => {
    const total = upsellList.length || 1;
    return ((Math.round(wheelPosition) % total) + total) % total;
  });
  let activeWheelItem = $derived(upsellList[activeWheelIndex] || upsellList[0]);
  let activeWheelQty = $derived(getUpsellQuantity(activeWheelItem?.id));
  let isWheelDragging = $state(false);
  let swipeOffsetY = $state(0); // Desplazamiento vertical al arrastrar hacia arriba
  let swipeAxis = $state(null); // 'horizontal' | 'vertical' | null
  let isFlyingUp = $state(false); // Transición de vuelo hacia la escena superior

  let wheelTouchStartX = 0;
  let wheelTouchStartY = 0;
  let wheelStartPos = 0;
  let wheelLastX = 0;
  let wheelLastTime = 0;
  let wheelVelocity = 0;
  let wheelAnimFrame = 0;
  let wheelDragDistance = 0; // px máximos recorridos en el gesto actual
  let flyUpTimeout = 0;

  // Limpieza de ciclo de vida: ningún rAF ni timeout sobrevive al desmonte
  // del modal (el bloque {#if store.isCartOpen} destruye el componente).
  onDestroy(() => {
    stopWheelAnimation();
    if (flyUpTimeout) {
      clearTimeout(flyUpTimeout);
      flyUpTimeout = 0;
    }
  });

  function prefersReducedMotion() {
    return (
      typeof window !== 'undefined' &&
      typeof window.matchMedia === 'function' &&
      window.matchMedia('(prefers-reduced-motion: reduce)').matches
    );
  }

  // Slots virtuales visibles continuos: cada slot k conserva invariante su ítem y su imagen (cero parpadeos)
  let visibleWheelSlots = $derived.by(() => {
    const total = upsellList.length;
    if (total === 0) return [];
    const base = Math.floor(wheelPosition);
    // Renderizamos base - 2 hasta base + 3 (6 slots continuos) para garantizar entrada/salida fluida con opacidad 0
    const slots = [];
    for (let k = base - 2; k <= base + 3; k++) {
      const cyclicIdx = ((k % total) + total) % total;
      const item = upsellList[cyclicIdx];
      if (item) {
        slots.push({ k, item, cyclicIdx });
      }
    }
    return slots;
  });

  function stopWheelAnimation() {
    if (wheelAnimFrame) {
      cancelAnimationFrame(wheelAnimFrame);
      wheelAnimFrame = 0;
    }
  }

  function handleWheelPointerDown(e) {
    stopWheelAnimation();
    const clientX = e.touches ? e.touches[0].clientX : e.clientX;
    const clientY = e.touches ? e.touches[0].clientY : e.clientY;
    wheelTouchStartX = clientX;
    wheelTouchStartY = clientY;
    wheelLastX = clientX;
    wheelLastTime = performance.now();
    wheelStartPos = wheelPosition;
    wheelVelocity = 0;
    wheelDragDistance = 0;
    if (flyUpTimeout) {
      clearTimeout(flyUpTimeout);
      flyUpTimeout = 0;
    }
    swipeOffsetY = 0;
    swipeAxis = null;
    isWheelDragging = true;
  }

  function handleWheelPointerMove(e) {
    if (!isWheelDragging) return;
    const clientX = e.touches ? e.touches[0].clientX : e.clientX;
    const clientY = e.touches ? e.touches[0].clientY : e.clientY;
    const dx = clientX - wheelTouchStartX;
    const dy = clientY - wheelTouchStartY;
    wheelDragDistance = Math.max(wheelDragDistance, Math.abs(dx), Math.abs(dy));

    if (!swipeAxis) {
      if (dy < -12 && Math.abs(dy) > Math.abs(dx) * 0.7) {
        swipeAxis = 'vertical';
      } else if (Math.abs(dx) > 10) {
        swipeAxis = 'horizontal';
      }
    }

    if (swipeAxis === 'vertical') {
      if (dy < 0) {
        swipeOffsetY = dy < -120 ? -120 + (dy + 120) * 0.3 : dy;
      } else {
        swipeOffsetY = dy * 0.15;
      }
    } else if (swipeAxis === 'horizontal' || !swipeAxis) {
      if (Math.abs(dx) > Math.abs(dy) * 0.45) {
        const now = performance.now();
        const dt = Math.max(1, now - wheelLastTime);
        const stepDelta = (wheelLastX - clientX) / 128;
        
        const currentV = stepDelta / dt;
        wheelVelocity = wheelVelocity * 0.35 + currentV * 0.65;
        
        wheelLastX = clientX;
        wheelLastTime = now;
        
        wheelPosition = wheelStartPos - (dx / 128);
      }
    }
  }

  function snapWheelToNearest(targetPos) {
    stopWheelAnimation();
    const target = targetPos !== undefined ? targetPos : Math.round(wheelPosition);
    // Movimiento reducido: salto instantáneo, sin resortes ni inercias.
    if (prefersReducedMotion()) {
      wheelPosition = target;
      return;
    }
    function springStep() {
      const diff = target - wheelPosition;
      if (Math.abs(diff) < 0.003) {
        wheelPosition = target;
        wheelAnimFrame = 0;
        return;
      }
      wheelPosition += diff * 0.18;
      wheelAnimFrame = requestAnimationFrame(springStep);
    }
    wheelAnimFrame = requestAnimationFrame(springStep);
  }

  function handleWheelPointerUp() {
    if (!isWheelDragging) return;
    isWheelDragging = false;

    if (swipeAxis === 'vertical') {
      if (swipeOffsetY < -30) {
        const centerInt = Math.round(wheelPosition);
        const total = upsellList.length || 1;
        const item = upsellList[((centerInt % total) + total) % total];
        if (item) {
          isFlyingUp = true;
          swipeOffsetY = -170;

          flyUpTimeout = setTimeout(() => {
            flyUpTimeout = 0;
            addUpsellItem(item);
            isFlyingUp = false;
            swipeOffsetY = 0;
            swipeAxis = null;
          }, 220);
        } else {
          swipeOffsetY = 0;
          swipeAxis = null;
        }
      } else {
        swipeOffsetY = 0;
        swipeAxis = null;
      }
      return;
    }

    swipeAxis = null;
    swipeOffsetY = 0;

    if (prefersReducedMotion()) {
      snapWheelToNearest();
      return;
    }

    if (Math.abs(wheelVelocity) > 0.002) {
      let v = wheelVelocity * 16;
      v = Math.max(-0.40, Math.min(0.40, v));
      const friction = 0.93;

      function inertiaStep() {
        wheelPosition += v;
        v *= friction;

        if (Math.abs(v) < 0.012) {
          snapWheelToNearest();
          return;
        }
        wheelAnimFrame = requestAnimationFrame(inertiaStep);
      }
      wheelAnimFrame = requestAnimationFrame(inertiaStep);
    } else {
      snapWheelToNearest();
    }
  }

  function getWheelSlotPositionStyle(offset) {
    const spacingX = 128;
    const x = offset * spacingX;
    const isCenterSlot = Math.abs(offset) < 0.35;
    const baseOffsetY = isCenterSlot ? swipeOffsetY : 0;
    const y = Math.pow(offset, 2) * 18 + baseOffsetY;
    
    let scale = Math.max(0.72, 1 - Math.abs(offset) * 0.08);
    let opacity = Math.abs(offset) >= 2.4 ? 0 : Math.max(0, 1 - Math.pow(Math.abs(offset) / 2.3, 1.8));
    
    if (isCenterSlot && isFlyingUp) {
      scale = 0.55;
      opacity = 0.15;
    }
    
    const zIndex = Math.round(50 - Math.abs(offset) * 10);

    const transitionRule = isFlyingUp
      ? 'transform 0.22s cubic-bezier(0.2, 0.9, 0.3, 1.2), opacity 0.2s ease'
      : (isWheelDragging || wheelAnimFrame ? 'none' : 'transform 0.32s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.25s ease');

    return `
      transform: translate3d(${x.toFixed(1)}px, ${y.toFixed(1)}px, 0) scale(${scale.toFixed(3)});
      opacity: ${opacity.toFixed(2)};
      z-index: ${zIndex};
      transition: ${transitionRule};
      pointer-events: ${opacity < 0.1 ? 'none' : 'auto'};
    `;
  }

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
      return d.toLocaleTimeString('es-CO', { hour: '2-digit', minute: '2-digit', hour12: true });
    } catch {
      return '03:00 p. m.';
    }
  }

  function formatTicketDate(dateStr) {
    try {
      const d = dateStr ? new Date(dateStr) : new Date();
      return d.toLocaleDateString('es-CO', { weekday: 'short', day: 'numeric', month: 'short' });
    } catch {
      return 'dom, 14 sept';
    }
  }

  // ═══════════════════════════════════════════════════════════════════════
  // NAVEGACIÓN ENTRE PASOS Y MANEJO REACTIVO DIRECTO DEL CARRITO
  // ═══════════════════════════════════════════════════════════════════════
  // La ruleta gestiona UNIDADES SUELTAS: una Tocineta suelta no es la misma
  // línea que la Tocineta como extra dentro de una hamburguesa
  // (ProductDetailModal). Badge y stepper cuentan solo líneas sueltas.
  function getUpsellQuantity(itemId) {
    if (!itemId) return 0;
    const found = store.cart.find(
      (it) => it.product?.id === itemId && !it.combo && (!it.extras || it.extras.length === 0)
    );
    return found ? (found.quantity || 0) : 0;
  }

  function addUpsellItem(item) {
    if (!item) return;
    const base = store.products.find((p) => p.id === item.id) || item;
    const fullProduct = { ...base };
    if (!fullProduct.image_url && item.image_url) {
      fullProduct.image_url = item.image_url;
    }
    store.addToCart(fullProduct, false, []);
  }

  // Dos formas de llamada, dos identidades:
  // - Chips de escena: `item.id` YA es el ID de la línea del carrito
  //   (viene de `upsellSceneChips`) → se opera directo sobre esa línea.
  // - Ruleta/stepper: ítem de catálogo (`id` = ID de producto) → se opera
  //   sobre la línea suelta de ese producto, sin tocar combos ni extras
  //   de hamburguesas (ver `getUpsellQuantity`).
  function removeUpsellItem(item) {
    if (!item) return;
    if (item.productId !== undefined) {
      store.updateQuantity(item.id, -1);
      return;
    }
    const found = store.cart.find(
      (it) => it.product?.id === item.id && !it.combo && (!it.extras || it.extras.length === 0)
    );
    if (found) {
      store.updateQuantity(found.id, -1);
    }
  }

  function handleContinueFromUpsell() {
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

    if (selectedMode !== 'in_store' && paymentMethod === 'Efectivo' && cashAmount.trim() !== '') {
      const cashVal = parseInt(cashAmount, 10);
      if (isNaN(cashVal) || cashVal < currentTotal) {
        formError = 'El monto en efectivo es menor al total a pagar.';
        return;
      }
    }

    let code = (selectedMode === 'delivery' ? 'DOM-' : selectedMode === 'takeout' ? 'TKA-' : 'ORD-') + Math.floor(1000 + Math.random() * 9000);
    let orderId = (selectedMode === 'delivery' ? 'dom-' : selectedMode === 'takeout' ? 'tka-' : 'ord-') + Date.now();

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
          order_type: order.type,
          items: apiItems,
          delivery_name: order.customer_name,
          delivery_phone: order.customer_phone,
          delivery_address: order.delivery_address,
          delivery_notes: order.delivery_notes,
          payment_method: order.payment_method,
          delivery_fee: feeSnapshot,
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
    let msg = `*NUEVA COMANDA — MÍAS HAMBURGUESAS*\n` +
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

  let activeSelectedSlug = $state(null);

  // Animación física con inercia Liquid Glass (Resorte elástico con rebote, masa y opacidad sólida, sin desvanecido borroso)
  function iosFluidMenu(node, { duration = 320 } = {}) {
    return {
      duration,
      css: (t) => {
        // En t=0 colapso total limpio
        if (t <= 0.001) {
          return 'transform: scale(0); opacity: 0; pointer-events: none;';
        }

        // Inercia de cristal líquido (momentum y rebote elástico):
        // Se expande con aceleración inicial, alcanza un overshoot (~1.045) en t ≈ 0.65 y asienta con tensión elástica en 1.00
        const ease = 1 - Math.pow(1 - t, 2.8) + Math.sin(t * Math.PI) * Math.pow(1 - t, 0.9) * 0.24;
        const scale = Math.max(0, ease);

        // Deformación líquida orgánica (squash & stretch elástico durante la inercia)
        const stretch = Math.sin(t * Math.PI) * 0.032 * (1 - t);
        const scaleX = scale * (1 - stretch);
        const scaleY = scale * (1 + stretch);
        const translateY = (1 - t) * -6;

        return `
          transform-origin: top right;
          transform: scale3d(${scaleX.toFixed(4)}, ${scaleY.toFixed(4)}, 1) translate3d(0, ${translateY.toFixed(2)}px, 0);
          opacity: 1;
        `;
      }
    };
  }

  function iosBackdrop(node, { duration = 180 } = {}) {
    return {
      duration,
      css: (t) => `opacity: ${t};`
    };
  }

  function closeOrderModal() {
    store.isCartOpen = false;
    showCategorySheet = false;
    activeSelectedSlug = null;
  }

  // Cierre con Escape (patrón del equipo, cf. BurgersScreen): primero el
  // sheet contextual, luego el modal. Solo cuando la comanda está abierta.
  $effect(() => {
    if (!store.isCartOpen) return;
    const handleKeyDown = (e) => {
      if (e.key !== 'Escape') return;
      if (showCategorySheet) showCategorySheet = false;
      else closeOrderModal();
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  });

  function addMoreProducts() {
    store.isCartOpen = false;
    showCategorySheet = false;
    activeSelectedSlug = null;
    store.currentView = 'menu';
  }

  // Icono elegante por categoría del sheet "+" (Lucide, sin emojis).
  // Las categorías se derivan de store.categories (DB) como en HomeHero.
  const SHEET_CATEGORY_ICONS = {
    hamburguesas: Sandwich,
    perros: Drumstick,
    de_la_casa: CookingPot,
    bebidas: CupSoda,
    adicionales: Sparkles,
    temporada: Star
  };

  const SHEET_FALLBACK_CATEGORIES = [
    { slug: 'hamburguesas', name: 'Hamburguesas' },
    { slug: 'perros', name: 'Perros Calientes' },
    { slug: 'de_la_casa', name: 'De la Casa' },
    { slug: 'adicionales', name: 'Adicionales' },
    { slug: 'bebidas', name: 'Bebidas' }
  ];

  let sheetCategories = $derived(
    store.categories && store.categories.length > 0
      ? store.categories.map((c) => ({ slug: c.slug, name: c.name }))
      : SHEET_FALLBACK_CATEGORIES
  );

  function handleSelectCategoryFromSheet(slug) {
    activeSelectedSlug = slug;
    // Feedback táctil y salida fluida antes de cambiar de vista
    setTimeout(() => {
      showCategorySheet = false;
      setTimeout(() => {
        store.isCartOpen = false;
        store.currentView = slug;
        activeSelectedSlug = null;
      }, 220);
    }, 70);
  }
</script>


<!-- ═══════════════════════════════════════════════════════════════════════ -->
<!-- FLUJO MULTI-PÁGINA DE COMANDA — ZERO SCROLL                          -->
<!-- ═══════════════════════════════════════════════════════════════════════ -->
{#if store.isCartOpen}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div
    data-lenis-prevent
    role="dialog"
    tabindex="-1"
    aria-modal="true"
    aria-label="Comanda actual"
    onwheel={(e) => e.stopPropagation()}
    ontouchmove={(e) => e.stopPropagation()}
    class="fixed inset-0 z-50 bg-char text-bone flex flex-col overflow-visible"
    style="height: 100dvh; touch-action: pan-y; -webkit-overflow-scrolling: touch; {cartStep === 'upsell' ? 'padding-top: max(0.5rem, calc(env(safe-area-inset-top, 0px) + 0.35rem)); padding-bottom: max(0.85rem, calc(env(safe-area-inset-bottom, 0px) + 0.5rem)); padding-left: max(0px, env(safe-area-inset-left, 0px)); padding-right: max(0px, env(safe-area-inset-right, 0px));' : 'padding-top: max(1rem, env(safe-area-inset-top, 0px)); padding-bottom: max(1rem, env(safe-area-inset-bottom, 0px)); padding-left: max(1rem, env(safe-area-inset-left, 0px)); padding-right: max(1rem, env(safe-area-inset-right, 0px));'}"
  >
    <div data-lenis-prevent class="w-full {cartStep === 'upsell' ? 'max-w-md md:max-w-4xl lg:max-w-5xl' : cartStep === 'mode' ? 'max-w-2xl' : 'max-w-lg'} mx-auto flex-1 flex flex-col min-h-0 overflow-visible">

      <!-- ════════════════════════════════════════════════════════════════ -->
      <!-- CABECERA COMPARTIDA                                            -->
      <!-- ════════════════════════════════════════════════════════════════ -->
      <header class="flex items-center justify-between w-full {cartStep === 'upsell' ? 'py-1.5 sm:py-2 px-5 sm:px-6' : 'py-2 sm:py-3'} shrink-0">
        {#if cartStep === 'upsell'}
          <!-- Botón Izquierdo: Volver atrás (flecha ChevronLeft idéntica a las demás vistas) -->
          <button
            type="button"
            onclick={closeOrderModal}
            class="btn-gold-outline w-14 h-14 sm:w-16 sm:h-16 rounded-full flex items-center justify-center text-bone active:scale-90 transition-all cursor-pointer group shrink-0"
            aria-label="Volver a la carta"
          >
            <ChevronLeft class="w-6 h-6 sm:w-7 sm:h-7 stroke-[2.2] -translate-x-0.5 group-hover:-translate-x-1 transition-transform" />
          </button>

          <div class="flex-1 text-center px-2">
            <span class="font-mono text-[10px] sm:text-xs uppercase tracking-widest text-[#8E8E93] font-bold block">
              Mías • Tu Cuenta
            </span>
            <h2 class="font-display font-black text-xl sm:text-2xl text-bone tracking-tight leading-tight">
              ¿Algo más?
            </h2>
          </div>

          <!-- Botón Derecho: Agregar / Más (+) con Menú Desplegable Contextual estilo iOS -->
          <div class="relative shrink-0">
            <button
              type="button"
              onclick={() => showCategorySheet = !showCategorySheet}
              class="btn-gold-outline w-14 h-14 sm:w-16 sm:h-16 rounded-full flex items-center justify-center text-bone hover:text-[#FF2400] active:scale-90 transition-all cursor-pointer group shrink-0 {showCategorySheet ? 'ring-2 ring-gold/40 bg-white/5' : ''}"
              aria-label="Agregar más productos"
            >
              <Plus class="w-6 h-6 sm:w-7 sm:h-7 stroke-[2.2] {showCategorySheet ? 'rotate-45 text-gold scale-95' : 'group-hover:rotate-90'} transition-transform duration-300 ease-[cubic-bezier(0.34,1.56,0.64,1)]" />
            </button>

            <!-- MENÚ DESPLEGABLE CONTEXTUAL FLOTANTE ESTILO iOS (DARK VIBRANCY CON INERCIA LIQUID GLASS) -->
            {#if showCategorySheet}
              <!-- Backdrop invisible para cerrar al tocar fuera (teclado: Escape) -->
              <!-- svelte-ignore a11y_click_events_have_key_events -->
              <!-- svelte-ignore a11y_no_static_element_interactions -->
              <div 
                transition:iosBackdrop={{ duration: 180 }}
                class="fixed inset-0 z-40 bg-black/50 backdrop-blur-[2px]"
                onclick={() => showCategorySheet = false}
              ></div>

              <!-- Tarjeta flotante iOS Context Menu con borde dorado biselado de lujo e inercia liquid glass -->
              <div 
                transition:iosFluidMenu={{ duration: 320 }}
                class="gold-border-card absolute right-0 top-[calc(100%+0.85rem)] z-50 w-72 sm:w-80 backdrop-blur-2xl rounded-card sm:rounded-card-lg p-5 sm:p-6 select-none will-change-transform origin-top-right"
                style="--gold-card-bg: rgba(26, 26, 29, 0.96); box-shadow: 0 30px 70px -15px rgba(0,0,0,0.92), 0 8px 24px -6px rgba(197, 160, 89, 0.22), inset 0 1px 1px rgba(255, 245, 207, 0.20);"
              >

                <!-- Sección 2: Añadir platos de la carta principal -->
                <div>
                  <span class="font-mono text-[10px] uppercase tracking-widest text-[#8E8E93] font-bold block px-2 mb-2">
                    Añadir Platos a la Carta
                  </span>
                  <div class="space-y-1">
                    {#each sheetCategories as cat (cat.slug)}
                      {@const CatIcon = SHEET_CATEGORY_ICONS[cat.slug] || UtensilsCrossed}
                      <button
                        type="button"
                        onclick={() => handleSelectCategoryFromSheet(cat.slug)}
                        class="w-full px-3.5 py-2.5 rounded-[20px] hover:bg-white/[0.08] active:bg-white/[0.14] flex items-center justify-between text-left transition-all cursor-pointer group"
                      >
                        <div class="flex items-center gap-3 min-w-0">
                          <CatIcon class="w-5 h-5 shrink-0 text-white/70 group-hover:text-gold transition-colors stroke-[1.8]" />
                          <span class="font-display font-bold text-sm text-white block truncate group-hover:text-gold transition-colors">
                            {cat.name}
                          </span>
                        </div>
                        <ChevronRight class="w-4 h-4 text-white/35 group-hover:text-gold group-hover:translate-x-0.5 transition-all shrink-0 stroke-[2]" />
                      </button>
                    {/each}
                  </div>
                </div>
              </div>
            {/if}
          </div>
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
        <div class="flex-1 flex flex-col items-center justify-between w-full px-4 relative overflow-visible min-h-0">
          {#if store.cart.length === 0}
            <!-- Estado sin productos en comanda -->
            <div class="flex-1 flex flex-col items-center justify-center text-center py-12 px-6 space-y-4 select-none">
              <div class="w-16 h-16 rounded-full bg-char-card border border-white/10 flex items-center justify-center text-ash">
                <ShoppingBag class="w-8 h-8 stroke-[1.8]" />
              </div>
              <div class="space-y-1 max-w-xs">
                <h3 class="font-display font-black text-xl text-bone">Tu comanda está vacía</h3>
                <p class="font-sans text-xs text-ash">Agrega tus productos favoritos de la carta para continuar.</p>
              </div>
              <button
                type="button"
                onclick={closeOrderModal}
                class="px-6 py-3 rounded-full bg-[#FF2400] hover:bg-[#E02000] text-white font-display font-bold text-sm transition-all active:scale-95 cursor-pointer shadow-md"
              >
                Explorar la Carta
              </button>
            </div>
          {:else}
            <!-- Escenario superior: producto principal en escena / plato -->
            <div class="flex-1 flex flex-col items-center justify-center w-full min-h-0 relative py-1">
              <div 
                class="relative z-20 flex items-center justify-center w-full h-[27vh] sm:h-[31vh] max-h-[280px] sm:max-h-[340px] select-none overflow-visible"
                style="perspective: 1100px; perspective-origin: 50% 50%; transform-style: preserve-3d;"
              >
                <!-- Todos los productos seleccionados en orden de agregado estilo Cover Flow (ref: cascada secuencial):
                     Héroe: el 1º agregado, centro al frente, plano. Acompañantes: alternan
                     derecha/izquierda, inclinados hacia el centro con retroceso Z, escala
                     y brillo decrecientes por rango. Estable: agregar anexa al margen
                     sin mover lo existente. Las bebidas, delgadas, abren 1.6× más
                     para no quedar tras el héroe. -->
                {#each allSelectedSceneUnits as unit, i (unit.id)}
                  {@const isHero = i === 0}
                  {@const side = (i % 2 === 1) ? 1 : -1}
                  {@const rank = isHero ? 0 : Math.ceil(i / 2)}
                  {@const zIndex = isHero ? 50 : (side === 1 ? 50 - (rank * 2) : 50 - (rank * 2) - 1)}
                  {@const spread = isHero ? 1 : (unit.category_id === 4 ? 1.6 : 1)}
                  {@const xOffset = isHero ? '0px' : (
                    rank === 1 
                      ? `calc(${side} * ${spread} * clamp(56px, 7vw, 78px))`
                      : `calc(${side} * ${spread} * (clamp(56px, 7vw, 78px) + ${rank - 1} * clamp(44px, 5.5vw, 62px)))`
                  )}
                  {@const rotY = isHero ? 0 : (side === 1 ? -Math.min(45, 20 + rank * 8) : Math.min(45, 20 + rank * 8))}
                  {@const depthZ = isHero ? '0px' : `-${rank * 70}px`}
                  {@const scale = isHero ? 1.04 : (unit.category_id === 4 ? Math.max(0.8, 0.97 - rank * 0.06) : Math.max(0.6, 0.97 - rank * 0.12))}
                  {@const brightness = isHero ? 1 : Math.max(0.6, 1 - rank * 0.14)}

                  <div 
                    class="absolute inset-0 flex items-center justify-center transition-all duration-500 cubic-bezier(0.16, 1, 0.3, 1) pointer-events-none"
                    style="
                      transform: translate3d({xOffset}, 0, {depthZ}) rotateY({rotY}deg) scale({scale});
                      z-index: {zIndex};
                      filter: drop-shadow(0 20px 38px rgba(0,0,0,0.92)) brightness({brightness});
                      transform-origin: center center;
                    "
                  >
                    <div class="relative flex items-center justify-center h-full w-full">
                      <!-- Sombra rasante individual en el piso -->
                      <div class="absolute bottom-1 sm:bottom-2 w-32 sm:w-44 h-5 rounded-[100%] bg-black/60 blur-[6px] pointer-events-none"></div>

                      {#if unit.image_url}
                        <div class="relative flex items-center justify-center">
                          <img 
                            src={unit.image_url} 
                            alt={unit.name} 
                            class="max-h-[25vh] sm:max-h-[29vh] max-w-[240px] sm:max-w-[300px] object-contain select-none pointer-events-none {unit.image_url?.endsWith('.jpg') ? 'mix-blend-lighten' : ''}"
                            loading="eager"
                          />
                          {#if unit.quantity > 1}
                            <div class="absolute -top-1 -right-1 sm:-top-2 sm:-right-2 min-w-[28px] h-7 px-1.5 rounded-full bg-[#FF2400] text-white flex items-center justify-center font-mono font-black text-xs shadow-[0_4px_14px_rgba(255,36,0,0.5)] border border-black/40 z-30 animate-in zoom-in-75 duration-200 pointer-events-none">
                              x{unit.quantity}
                            </div>
                          {/if}
                        </div>
                      {:else}
                        <div class="w-36 h-36 rounded-full bg-char-card border border-white/10 flex items-center justify-center text-bone">
                          <ShoppingBag class="w-12 h-12 text-ash" />
                        </div>
                      {/if}
                    </div>
                  </div>
                {/each}
              </div>

              <!-- Fila organizada de acompañamientos elegidos con retiro en 1 toque -->
              {#if upsellSceneChips.length > 0}
                <div class="flex items-center justify-center gap-1.5 flex-wrap px-4 pt-1 z-30 select-none max-w-lg">
                  {#each upsellSceneChips as u (u.id)}
                    <div class="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full bg-char-card/95 border border-white/15 text-bone font-mono text-xs shadow-md animate-in fade-in duration-200">
                      <span class="font-sans font-medium text-bone truncate max-w-[110px]">{u.shortName || u.name}</span>
                      <span class="font-bold text-[#FF2400] font-mono">x{u.quantity}</span>
                      <button
                        type="button"
                        onclick={() => removeUpsellItem(u)}
                        class="w-4 h-4 rounded-full flex items-center justify-center bg-white/10 hover:bg-white/25 text-ash hover:text-white transition-colors cursor-pointer"
                        aria-label="Quitar una unidad"
                      >
                        <X class="w-2.5 h-2.5 stroke-[2.5]" />
                      </button>
                    </div>
                  {/each}
                </div>
              {/if}
            </div>

            <!-- PARTE INFERIOR: RULETA DE CÍRCULOS EN MEDIA LUNA (Touch & Gesture Physics) -->
            <!-- svelte-ignore a11y_no_static_element_interactions -->
            <!-- svelte-ignore a11y_no_noninteractive_element_interactions -->
            <div
              role="group"
              aria-label="Ruleta de productos sugeridos"
              class="w-full flex flex-col items-center justify-center shrink-0 pt-0 pb-1 select-none relative z-30"
              ontouchstart={handleWheelPointerDown}
              ontouchmove={handleWheelPointerMove}
              ontouchend={handleWheelPointerUp}
              onmousedown={handleWheelPointerDown}
              onmousemove={handleWheelPointerMove}
              onmouseup={handleWheelPointerUp}
              onmouseleave={handleWheelPointerUp}
              style="touch-action: none;"
            >
              <!-- Arco de círculos en media luna / ruleta infinita (Agrandados: 110px) -->
              <div class="relative w-full h-[130px] sm:h-[140px] flex items-center justify-center overflow-visible">
                {#each visibleWheelSlots as slot (slot.k)}
                  {@const offset = slot.k - wheelPosition}
                  {@const item = slot.item}
                  {@const isCenter = Math.abs(offset) < 0.35}
                  {@const itemQty = getUpsellQuantity(item.id)}
                  <button
                    type="button"
                    class="absolute bg-transparent border-0 p-0 cursor-pointer flex items-center justify-center will-change-transform"
                    style={getWheelSlotPositionStyle(offset)}
                    aria-label={isCenter ? `Añadir ${item.name} por ${formatCOP(item.price)}` : `Centrar ${item.name} en la ruleta`}
                    onclick={() => {
                      // Un arrastre no es un toque: evita altas accidentales
                      // al soltar sobre el centro tras girar la ruleta.
                      if (wheelDragDistance > 12) return;
                      if (isCenter) {
                        addUpsellItem(item);
                      } else {
                        snapWheelToNearest(slot.k);
                      }
                    }}
                  >
                    <!-- Flechita interactiva en movimiento indicando deslizar hacia arriba o tocar para agregar -->
                    {#if isCenter}
                      <div class="absolute -top-7 sm:-top-8.5 left-1/2 -translate-x-1/2 flex items-center gap-1 pointer-events-none select-none z-50 transition-opacity duration-150 {swipeOffsetY < -10 ? 'opacity-0' : 'opacity-100'}">
                        <ChevronUp class="w-5 h-5 sm:w-6 sm:h-6 text-white animate-swipe-up-glow drop-shadow-[0_2px_8px_rgba(255,255,255,0.85)]" />
                        {#if itemQty > 0}
                          <span class="font-mono text-[10px] font-bold text-white/90 animate-swipe-up-glow">+1</span>
                        {/if}
                      </div>
                    {/if}

                    <!-- Círculo de acompañamiento estilo Apple (110px: +30% de escala) -->
                    <div 
                      class="relative w-[106px] h-[106px] sm:w-[116px] sm:h-[116px] shrink-0 rounded-full flex items-center justify-center active:scale-95 transition-all duration-300 {isCenter ? 'bg-[#222226] ring-2 ring-white/80 shadow-[0_14px_32px_rgba(0,0,0,0.85),0_0_24px_rgba(255,255,255,0.14)]' : 'bg-[#161619]/90 border border-white/15 hover:border-white/25 shadow-md'}"
                    >
                      {#if item.image_url}
                        <img 
                          src={item.image_url} 
                          alt={item.name} 
                          class="w-[70px] h-[70px] sm:w-[78px] sm:h-[78px] max-w-[70px] max-h-[70px] sm:max-w-[78px] sm:max-h-[78px] object-contain select-none pointer-events-none drop-shadow-sm"
                          loading="eager"
                        />
                      {:else}
                        <Sparkles class="w-7 h-7 text-gold" />
                      {/if}

                      <!-- Badge con número indicativo de unidades añadidas -->
                      {#if itemQty > 0}
                        <div class="absolute -top-1.5 -right-1.5 min-w-[28px] h-7 px-1.5 rounded-full bg-[#FF2400] text-white flex items-center justify-center shadow-lg border border-black/40 font-mono font-black text-xs animate-in zoom-in duration-150">
                          {itemQty}
                        </div>
                      {/if}
                    </div>
                  </button>
                {/each}
              </div>

              <!-- Nombre del producto sugerido centrado y selector de cantidad -->
              <div class="text-center mt-1.5 mb-1 flex flex-col items-center justify-center min-h-[46px]">
                <span class="font-display font-black text-base sm:text-lg text-bone block truncate tracking-tight max-w-xs">
                  {activeWheelItem ? activeWheelItem.name : 'Acompañamiento'}
                </span>

                {#if activeWheelItem}
                  {#if activeWheelQty > 0}
                    <!-- Cápsula interactiva de control de cantidad -->
                    <div class="mt-1 inline-flex items-center gap-2 px-2.5 py-1 rounded-full bg-[#1C1C1F] border border-white/20 text-xs font-mono shadow-md">
                      <button
                        type="button"
                        onclick={() => removeUpsellItem(activeWheelItem)}
                        class="w-6 h-6 rounded-full bg-white/10 hover:bg-white/20 text-bone flex items-center justify-center active:scale-90 transition-all cursor-pointer"
                        aria-label="Quitar una unidad"
                      >
                        <Minus class="w-3.5 h-3.5 stroke-[2.5]" />
                      </button>
                      <span class="font-bold text-white min-w-[20px] text-center font-mono text-sm">{activeWheelQty}</span>
                      <button
                        type="button"
                        onclick={() => addUpsellItem(activeWheelItem)}
                        class="w-6 h-6 rounded-full bg-[#FF2400] hover:bg-[#E02000] text-white flex items-center justify-center active:scale-90 transition-all cursor-pointer shadow-sm"
                        aria-label="Agregar otra unidad"
                      >
                        <Plus class="w-3.5 h-3.5 stroke-[2.5]" />
                      </button>
                      <span class="text-gold font-bold ml-1 text-xs">+{formatCOP(activeWheelItem.price * activeWheelQty)}</span>
                    </div>
                  {:else}
                    <button
                      type="button"
                      onclick={() => addUpsellItem(activeWheelItem)}
                      class="mt-1 inline-flex items-center gap-1.5 px-3.5 py-1 rounded-full text-xs font-mono font-bold transition-all cursor-pointer bg-white/10 text-gold hover:bg-white/15 border border-white/15 active:scale-95"
                    >
                      <Plus class="w-3.5 h-3.5 stroke-[2.5]" />
                      <span>Añadir • +{formatCOP(activeWheelItem.price)}</span>
                    </button>
                  {/if}
                {/if}
              </div>
            </div>

            <!-- Botón grande "Continuar" con la cuenta / total en tiempo real -->
            <div class="w-full max-w-sm sm:max-w-md mx-auto pt-0 pb-1.5 sm:pb-2.5 px-4 shrink-0 z-30 overflow-visible">
              <button
                type="button"
                onclick={handleContinueFromUpsell}
                class="btn-gold-red w-full min-h-[74px] sm:min-h-[80px] md:min-h-[86px] rounded-full text-white px-6 sm:px-8 flex items-center justify-between active:scale-[0.97] transition-all duration-150 cursor-pointer select-none shadow-[0_12px_32px_rgba(255,36,0,0.45)]"
              >
                <div class="flex flex-col text-left">
                  <span class="font-sans text-[11px] sm:text-xs uppercase tracking-wider text-white/80 font-bold">
                    {store.cartCount > 0 ? `Cuenta (${store.cartCount} ${store.cartCount === 1 ? 'item' : 'items'})` : 'Total Cuenta'}
                  </span>
                  <span class="font-mono font-black text-xl sm:text-2xl text-white tracking-tight leading-none">
                    {formatCOP(currentUpsellTotal)}
                  </span>
                </div>
                <div class="flex items-center gap-2">
                  <span class="font-display font-black text-lg sm:text-xl tracking-tight text-white">Continuar</span>
                  <ChevronRight class="w-6 h-6 sm:w-7 sm:h-7 stroke-[2.8]" />
                </div>
              </button>
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
                1 Toque • A la Mesa
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
                  {#each ['Tocar timbre', 'Dejar en portería', 'Llamar al llegar'] as chip}
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
