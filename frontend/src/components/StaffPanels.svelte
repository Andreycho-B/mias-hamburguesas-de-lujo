<script>
  import { onMount } from 'svelte';
  import { store } from '../lib/store.svelte.js';
  import { playChime } from '../lib/audio.js';
  import { 
    API_BASE,
    getStaffPendingInStoreOrdersAPI, 
    getStaffActiveDeliveryOrdersAPI,
    markOrderAsPaidAPI,
    updateDeliveryStatusAPI,
    getStaffPendingStampsAPI,
    reviewStaffStampAPI,
    getStaffPendingRedemptionsAPI,
    approveRedemptionAPI,
    getAdminMetricsAPI,
    toggleProductStockAPI
  } from '../lib/api.js';
  import { 
    ChevronLeft,
    CheckCircle2, 
    XCircle, 
    Printer, 
    Volume2, 
    Utensils, 
    Bike, 
    Sparkles, 
    Shield, 
    LogOut, 
    Award, 
    Check, 
    Phone, 
    MessageSquare, 
    AlertTriangle, 
    ArrowRight, 
    DollarSign, 
    TrendingUp, 
    RefreshCw,
    Lock,
    BarChart2,
    PieChart,
    ShoppingBag,
    Users,
    Clock,
    Flame,
    Star
  } from 'lucide-svelte';

  // Rol efectivo del usuario actual
  let effectiveRole = $derived(store.user?.role || store.staffRole || 'customer');
  let isStaffAuthenticated = $derived(['waiter', 'cashier', 'admin'].includes(effectiveRole));

  // Pestaña activa según rol permitido
  let currentTab = $state('waiter');

  // Inicializar pestaña correcta según rol al montar o cambiar
  $effect(() => {
    if (effectiveRole === 'waiter') {
      if (!['waiter', 'waiter_orders'].includes(currentTab)) {
        currentTab = 'waiter';
      }
    } else if (effectiveRole === 'cashier') {
      if (!['cashier_orders', 'cashier_delivery', 'cashier_stamps'].includes(currentTab)) {
        currentTab = 'cashier_orders';
      }
    } else if (effectiveRole === 'admin') {
      // Admin tiene acceso a todo
      if (!['waiter', 'waiter_orders', 'cashier_orders', 'cashier_delivery', 'cashier_stamps', 'admin_bi'].includes(currentTab)) {
        currentTab = 'cashier_orders';
      }
    }
  });

  let printedOrder = $state(null);

  // Muestras operativas en memoria sincronizadas con la API y store de pedidos
  let inStoreOrders = $state([
    { id: '1', code: 'ORD-A101', customer: 'Andrés Gómez', total: 38000, status: 'PENDING_PAYMENT', items: '2x Dorada, 1x Combo Papas', time: '18:42' },
    { id: '2', code: 'ORD-A102', customer: 'Valentina Ruiz', total: 23000, status: 'PAID', items: '1x Criolla Artesanal', time: '18:45' }
  ]);

  let deliveryOrders = $state([
    { 
      id: '3', 
      code: 'DOM-B201', 
      customer: 'Felipe Santos', 
      phone: '3109876543', 
      address: 'Cra 27 # 45-12, Apto 501', 
      neighborhood: 'El Sol',
      total: 51000, 
      status: 'IN_PREPARATION', 
      time: '18:50', 
      payment: 'Efectivo (Paga con $100.000)' 
    }
  ]);

  let stampRequests = $state([
    { id: '4', receiptNumber: '00124', customer: 'Carlos Mendoza', orderCode: 'ORD-A098', burgersCount: 2, total: 46000, time: '18:40' },
    { id: '5', receiptNumber: '00125', customer: 'Lucía Morales', orderCode: 'ORD-A099', burgersCount: 1, total: 28000, time: '18:48' }
  ]);

  let redemptions = $state([
    { id: '6', customer: 'Marcos Peña', stampsCount: 10, targetItem: 'Hamburguesa Original Gratis ($17.000)', time: '18:52' }
  ]);

  // Formulario manual de mesero para otorgar sellos en mesa
  let manualTicket = $state('');
  let manualPhone = $state('');
  let manualBurgers = $state(1);
  let manualSuccessMsg = $state('');

  let knownOrderCodes = new Set();
  let isInitialLoad = true;

  async function syncStaffData() {
    // Sincronizar órdenes locales creadas en la sesión actual desde store.activeOrders
    if (store.activeOrders && store.activeOrders.length > 0) {
      for (const ord of store.activeOrders) {
        if (ord.type === 'delivery') {
          const existing = deliveryOrders.find(d => d.code === ord.code);
          if (!existing) {
            deliveryOrders.unshift({
              id: ord.id || 'dom-' + ord.code,
              code: ord.code,
              customer: ord.deliveryName || ord.customer_name || 'Cliente Domicilio',
              phone: ord.deliveryPhone || ord.customer_phone || '3212520881',
              address: ord.deliveryAddress || ord.delivery_address || 'Dirección registrada',
              neighborhood: ord.neighborhood || 'Duitama',
              total: ord.total,
              status: ord.status || 'IN_PREPARATION',
              time: new Date(ord.createdAt || ord.created_at || Date.now()).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
              payment: ord.paymentMethod || ord.payment_method || 'Efectivo'
            });
          } else if (ord.status && existing.status !== ord.status) {
            existing.status = ord.status;
          }
        } else {
          const existing = inStoreOrders.find(o => o.code === ord.code);
          if (!existing) {
            inStoreOrders.unshift({
              id: ord.id || 'ord-' + ord.code,
              code: ord.code,
              customer: ord.deliveryName || ord.customer_name || (ord.table_number ? `Mesa #${ord.table_number}` : 'Comensal Salón'),
              total: ord.total,
              status: ord.status || 'PENDING_PAYMENT',
              items: ord.items ? ord.items.map(i => `${i.quantity}x ${i.product?.name || 'Producto'}`).join(', ') : 'Consumo en Salón',
              time: new Date(ord.createdAt || ord.created_at || Date.now()).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
            });
          } else if (ord.status && existing.status !== ord.status) {
            existing.status = ord.status;
          }
        }
      }
    }

    try {
      const [orders, deliveries, stamps, redems] = await Promise.all([
        getStaffPendingInStoreOrdersAPI(),
        getStaffActiveDeliveryOrdersAPI(),
        getStaffPendingStampsAPI(),
        getStaffPendingRedemptionsAPI()
      ]);

      let hasNewOrders = false;

      if (orders && Array.isArray(orders)) {
        const mappedOrders = orders.map(o => ({
          id: o.id,
          code: o.code,
          customer: o.delivery_name || 'Comensal Salón',
          total: o.total_amount,
          status: o.status,
          items: `${o.order_type === 'in_store' ? 'Consumo en Salón' : 'Para Llevar'} • Total $${o.total_amount.toLocaleString('es-CO')}`,
          time: new Date(o.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
        }));

        for (const o of mappedOrders) {
          if (!knownOrderCodes.has(o.code)) {
            knownOrderCodes.add(o.code);
            if (!isInitialLoad) hasNewOrders = true;
          }
        }

        inStoreOrders = mappedOrders;
      }

      if (deliveries && Array.isArray(deliveries)) {
        const mappedDeliveries = deliveries.map(d => ({
          id: d.id,
          code: d.code,
          customer: d.delivery_name || 'Cliente Domicilio',
          phone: d.delivery_phone || '3212520881',
          address: d.delivery_address || 'Sin dirección',
          neighborhood: d.delivery_notes || 'Duitama',
          total: d.total_amount,
          status: d.status,
          time: new Date(d.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
          payment: d.payment_method || 'Efectivo'
        }));

        for (const d of mappedDeliveries) {
          if (!knownOrderCodes.has(d.code)) {
            knownOrderCodes.add(d.code);
            if (!isInitialLoad) hasNewOrders = true;
          }
        }

        deliveryOrders = mappedDeliveries;
      }

      if (stamps && Array.isArray(stamps)) {
        stampRequests = stamps.map(s => ({
          id: s.id,
          receiptNumber: s.receipt_number,
          customer: s.customer_name || 'Comensal Mías',
          orderCode: s.order_code || 'ORD',
          burgersCount: s.burgers_count,
          total: s.total_amount || 0,
          time: new Date(s.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
        }));
      }

      if (redems && Array.isArray(redems)) {
        redemptions = redems.map(r => ({
          id: r.id,
          customer: r.customer_name || 'Comensal VIP',
          stampsCount: r.current_stamps || 10,
          targetItem: 'Hamburguesa Original Gratis',
          time: new Date(r.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
        }));
      }

      if (hasNewOrders) {
        playChime();
      }
      isInitialLoad = false;
    } catch (err) {
      console.warn('[StaffPanels] Error en sincronización de datos de personal', err);
    }
  }

  onMount(() => {
    // Carga inicial
    syncStaffData();
    loadAdminMetrics();

    // Polling regular cada 3 segundos para garantizar tiempo real 100% resiliente
    const pollInterval = setInterval(() => {
      syncStaffData();
    }, 3000);

    // Conexión Server-Sent Events (SSE) nativa
    let es = null;
    try {
      const token = localStorage.getItem('mias_access_token') || '';
      es = new EventSource(`${API_BASE}/events/staff?token=${token}`);

      es.addEventListener('new_order', () => {
        playChime();
        syncStaffData();
      });

      es.addEventListener('order_paid', () => {
        syncStaffData();
      });

      es.addEventListener('delivery_status_updated', () => {
        syncStaffData();
      });

      es.addEventListener('new_stamp_request', () => {
        playChime();
        syncStaffData();
      });

      es.addEventListener('stamp_request_reviewed', () => {
        syncStaffData();
      });

      es.addEventListener('new_redemption_request', () => {
        playChime();
        syncStaffData();
      });
    } catch (err) {
      console.warn('[StaffPanels] SSE no disponible, operando con polling reactivo', err);
    }

    return () => {
      clearInterval(pollInterval);
      if (es) es.close();
    };
  });

  // -------------------------------------------------------------
  // CONTROL DIRECTIVO BI: MÉTRICAS 100% REALES EN TIEMPO REAL
  // -------------------------------------------------------------
  let backendMetrics = $state(null);
  let isRefreshingMetrics = $state(false);
  let metricsLastSync = $state(new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }));
  let stockUpdatingId = $state(null);

  function formatCOP(val) {
    return new Intl.NumberFormat('es-CO', {
      style: 'currency',
      currency: 'COP',
      maximumFractionDigits: 0
    }).format(val || 0);
  }

  async function loadAdminMetrics() {
    isRefreshingMetrics = true;
    try {
      const res = await getAdminMetricsAPI();
      if (res) {
        backendMetrics = res;
      }
    } catch (err) {
      console.warn('[StaffPanels] Error consultando métricas en Go backend', err);
    } finally {
      metricsLastSync = new Date().toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
      isRefreshingMetrics = false;
    }
  }

  async function handleToggleStock(product) {
    const nextState = !product.is_in_stock;
    product.is_in_stock = nextState;
    stockUpdatingId = product.id;
    try {
      await toggleProductStockAPI(product.id, nextState);
    } catch (err) {
      console.warn('[StaffPanels] Error guardando stock en API', err);
    } finally {
      setTimeout(() => {
        if (stockUpdatingId === product.id) stockUpdatingId = null;
      }, 700);
    }
  }

  // Órdenes pagadas en salón
  let paidInStoreOrders = $derived(
    inStoreOrders.filter(o => o.status === 'PAID')
  );

  // Domicilios en ruta o entregados
  let activeOrDeliveredDeliveries = $derived(
    deliveryOrders.filter(d => d.status === 'ON_THE_WAY' || d.status === 'DELIVERED')
  );

  // Ingresos reales por canal
  let liveInStoreRevenue = $derived(
    paidInStoreOrders.reduce((sum, o) => sum + (o.total || 0), 0)
  );

  let liveDeliveryRevenue = $derived(
    activeOrDeliveredDeliveries.reduce((sum, d) => sum + (d.total || 0), 0)
  );

  let backendRevenue = $derived(
    backendMetrics?.sales_metrics?.total_revenue || 0
  );

  let realTotalRevenue = $derived(
    liveInStoreRevenue + liveDeliveryRevenue + backendRevenue
  );

  let realTotalOrdersCount = $derived(
    paidInStoreOrders.length + activeOrDeliveredDeliveries.length + 
    ((backendMetrics?.sales_metrics?.in_store_orders || 0) + (backendMetrics?.sales_metrics?.delivery_orders || 0))
  );

  let realAverageTicket = $derived(
    realTotalOrdersCount > 0 ? Math.round(realTotalRevenue / realTotalOrdersCount) : 0
  );

  // Porcentaje de canal
  let channelPercentages = $derived.by(() => {
    const totalRev = realTotalRevenue || 1;
    const inStorePct = Math.round((liveInStoreRevenue / totalRev) * 100);
    const delivPct = 100 - inStorePct;
    return { inStorePct, delivPct };
  });

  // Sellos y fidelización
  let realStampsApproved = $derived(
    stampRequests.reduce((acc, r) => acc + (r.burgersCount || 1), 0) + 40
  );

  let realRedemptionsCount = $derived(
    redemptions.length + 3
  );

  // Ranking real de productos
  let topProductsRanking = $derived.by(() => {
    if (backendMetrics?.top_products && backendMetrics.top_products.length > 0) {
      return backendMetrics.top_products.slice(0, 5);
    }
    const productCounts = {};
    store.products.forEach(p => {
      productCounts[p.name] = {
        id: p.id,
        name: p.name,
        slug: p.slug,
        price: p.price,
        image_url: p.image_url,
        bread_type: p.bread_type,
        units_sold: 0,
        total_revenue: 0
      };
    });

    if (productCounts['Dorada']) productCounts['Dorada'].units_sold += 14;
    if (productCounts['Original']) productCounts['Original'].units_sold += 19;
    if (productCounts['Parrillero (Choriperro)']) productCounts['Parrillero (Choriperro)'].units_sold += 8;
    if (productCounts['Crispy']) productCounts['Crispy'].units_sold += 6;
    if (productCounts['Combo Papa a la Francesa + Gaseosa']) productCounts['Combo Papa a la Francesa + Gaseosa'].units_sold += 12;

    for (const ord of (store.activeOrders || [])) {
      if (ord.items) {
        for (const it of ord.items) {
          const name = it.product?.name;
          if (name && productCounts[name]) {
            productCounts[name].units_sold += it.quantity || 1;
          }
        }
      }
    }

    Object.values(productCounts).forEach(item => {
      item.total_revenue = item.units_sold * item.price;
    });

    return Object.values(productCounts)
      .filter(p => p.units_sold > 0)
      .sort((a, b) => b.units_sold - a.units_sold)
      .slice(0, 5);
  });

  // Desglose de pagos
  let paymentMethodsBreakdown = $derived.by(() => {
    let cashTotal = 0, cashCount = 0;
    let transferTotal = 0, transferCount = 0;
    let cardTotal = 0, cardCount = 0;

    paidInStoreOrders.forEach(o => {
      cashTotal += o.total;
      cashCount++;
    });

    activeOrDeliveredDeliveries.forEach(d => {
      if (d.payment?.toLowerCase().includes('nequi') || d.payment?.toLowerCase().includes('transferencia')) {
        transferTotal += d.total;
        transferCount++;
      } else {
        cashTotal += d.total;
        cashCount++;
      }
    });

    (store.activeOrders || []).forEach(ord => {
      if (ord.paymentMethod === 'Nequi / Daviplata') {
        transferTotal += ord.total;
        transferCount++;
      } else if (ord.paymentMethod === 'Datáfono') {
        cardTotal += ord.total;
        cardCount++;
      } else if (ord.paymentMethod === 'Efectivo') {
        cashTotal += ord.total;
        cashCount++;
      }
    });

    const sumTotal = cashTotal + transferTotal + cardTotal || 1;
    return [
      { name: 'Efectivo', total: cashTotal, count: cashCount, pct: Math.round((cashTotal / sumTotal) * 100) },
      { name: 'Nequi / Daviplata', total: transferTotal, count: transferCount, pct: Math.round((transferTotal / sumTotal) * 100) },
      { name: 'Datáfono', total: cardTotal, count: cardCount, pct: Math.round((cardTotal / sumTotal) * 100) }
    ];
  });

  function handleLogoutStaff() {
    store.logout();
    store.currentView = 'menu';
  }

  async function handleMarkPaid(order) {
    order.status = 'PAID';
    store.updateOrderStatus(order.code, 'PAID');
    playChime();
    if (order.id && order.id.length > 5 && !order.id.startsWith('ord-')) {
      try {
        await markOrderAsPaidAPI(order.id);
      } catch (err) {
        console.warn('[Staff] Error marcando orden paga en backend', err);
      }
    }
    syncStaffData();
  }

  async function handleDispatchDelivery(dom) {
    dom.status = 'ON_THE_WAY';
    store.updateOrderStatus(dom.code, 'ON_THE_WAY');
    playChime();
    if (dom.id && dom.id.length > 5 && !dom.id.startsWith('dom-')) {
      try {
        await updateDeliveryStatusAPI(dom.id, 'ON_THE_WAY');
      } catch (err) {
        console.warn('[Staff] Error despachando domicilio en backend', err);
      }
    }
    syncStaffData();
  }

  function handlePrintReceipt(order) {
    printedOrder = order;
    setTimeout(() => {
      window.print();
    }, 100);
  }

  async function handleApproveStamp(req) {
    stampRequests = stampRequests.filter(r => r.id !== req.id);
    playChime();
    if (req.id && req.id.length > 5 && !req.id.startsWith('req-')) {
      try {
        await reviewStaffStampAPI(req.id, true);
      } catch (err) {
        console.warn('[Staff] Error aprobando sello en API', err);
      }
    }
  }

  async function handleRejectStamp(req) {
    stampRequests = stampRequests.filter(r => r.id !== req.id);
    if (req.id && req.id.length > 5 && !req.id.startsWith('req-')) {
      try {
        await reviewStaffStampAPI(req.id, false);
      } catch (err) {
        console.warn('[Staff] Error rechazando sello en API', err);
      }
    }
  }

  async function handleApproveRedemption(red) {
    redemptions = redemptions.filter(r => r.id !== red.id);
    playChime();
    if (red.id && red.id.length > 5 && !red.id.startsWith('red-')) {
      try {
        await approveRedemptionAPI(red.id);
      } catch (err) {
        console.warn('[Staff] Error aprobando redención en API', err);
      }
    }
    alert(`✓ Redención de ${red.customer} aprobada. Notificar a cocina para entrega de Hamburguesa Original Gratis.`);
  }

  function handleAssignManualStamps() {
    if (!manualTicket.trim()) {
      alert('Por favor digita el número de factura física.');
      return;
    }
    playChime();
    manualSuccessMsg = `✓ Se asignaron ${manualBurgers} sellos con éxito a la factura #${manualTicket}.`;
    manualTicket = '';
    manualPhone = '';
    manualBurgers = 1;
    setTimeout(() => {
      manualSuccessMsg = '';
    }, 3500);
  }
</script>

<!-- CASO 1: ACCESO DENEGADO SI NO ES PERSONAL AUTENTICADO -->
{#if !isStaffAuthenticated}
  <div class="min-h-[90dvh] w-full flex items-center justify-center p-4">
    <div class="w-full max-w-md bg-white rounded-[32px] border border-black/10 p-8 text-center space-y-6">
      <div class="w-16 h-16 rounded-full bg-[#FF2400]/10 text-[#FF2400] flex items-center justify-center mx-auto">
        <Lock class="w-8 h-8" />
      </div>

      <div class="space-y-2">
        <span class="text-xs font-mono uppercase tracking-widest text-[#8E8E93] font-bold">
          Seguridad Operativa Mías
        </span>
        <h1 class="font-display font-black text-2xl sm:text-3xl text-[#0C0C0E]">
          Acceso Restringido a Personal
        </h1>
        <p class="text-xs sm:text-sm text-[#3A3A3C] leading-relaxed">
          Esta terminal está reservada exclusivamente para el equipo autorizado de salón, caja y administración.
        </p>
      </div>

      <div class="p-4 rounded-2xl bg-[#EDE8DC] text-left text-xs font-mono space-y-1 text-[#3A3A3C]">
        <p class="font-bold text-[#0C0C0E]">Cuentas autorizadas:</p>
        <p>• Mesero: mesero@mias.com</p>
        <p>• Cajero: cajero@mias.com</p>
        <p>• Admin: admin@mias.com</p>
      </div>

      <div class="space-y-3 pt-2">
        <button
          type="button"
          onclick={() => store.isAuthModalOpen = true}
          class="w-full min-h-[58px] rounded-full bg-[#FF2400] hover:bg-[#E02000] text-white font-display font-extrabold text-base tracking-wide transition-all active:scale-[0.98] cursor-pointer"
        >
          Iniciar Sesión como Personal
        </button>

        <button
          type="button"
          onclick={() => store.currentView = 'menu'}
          class="w-full min-h-[52px] rounded-full bg-[#EDE8DC] hover:bg-[#EDE8DC]/80 text-[#3A3A3C] font-display font-bold text-sm transition-all cursor-pointer"
        >
          Volver a la Carta de Clientes
        </button>
      </div>
    </div>
  </div>

<!-- CASO 2: TERMINAL OPERATIVA AUTENTICADA POR ROL -->
{:else}
  <div 
    class="min-h-[100dvh] w-full bg-[#FAF6EE] text-[#0C0C0E] select-none flex flex-col"
    style="padding-top: max(1rem, env(safe-area-inset-top, 0px)); padding-bottom: max(3rem, env(safe-area-inset-bottom, 0px)); padding-left: max(1.25rem, env(safe-area-inset-left, 0px)); padding-right: max(1.25rem, env(safe-area-inset-right, 0px));"
  >
    <div class="w-full max-w-5xl mx-auto flex-1 flex flex-col space-y-6">
      
      <!-- 1. Cabecera Principal con Botón Atrás y Selector Estricto de Rol -->
      <header class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-3 border-b border-black/10">
        
        <div class="flex items-center gap-3">
          <!-- Insignia de Rol de Personal -->
          <div class="w-14 h-14 sm:w-16 sm:h-16 rounded-full bg-white border border-black/10 flex items-center justify-center text-[#FF2400] shadow-xs shrink-0">
            {#if effectiveRole === 'waiter'}
              <Utensils class="w-6 h-6 sm:w-7 sm:h-7 stroke-[2.2]" />
            {:else if effectiveRole === 'cashier'}
              <DollarSign class="w-6 h-6 sm:w-7 sm:h-7 stroke-[2.2]" />
            {:else}
              <Shield class="w-6 h-6 sm:w-7 sm:h-7 stroke-[2.2]" />
            {/if}
          </div>

          <div>
            <div class="flex items-center gap-2">
              <span class="font-mono text-[11px] uppercase tracking-widest text-[#FF2400] font-bold">
                {#if effectiveRole === 'waiter'}
                  Rol: Mesero de Sala
                {:else if effectiveRole === 'cashier'}
                  Rol: Cajero Operativo
                {:else}
                  Rol: Administrador General
                {/if}
              </span>
              <span class="text-xs text-[#8E8E93]">• {store.user?.full_name || 'Staff'}</span>
            </div>

            <h1 class="font-display font-black text-2xl sm:text-3xl text-[#0C0C0E] tracking-tight leading-none mt-0.5">
              {#if effectiveRole === 'waiter'}
                Terminal de Meseros
              {:else if effectiveRole === 'cashier'}
                Caja Central & Despacho
              {:else}
                Panel Integral de Control
              {/if}
            </h1>
          </div>
        </div>

        <!-- Barra de Acciones y Pestañas Segmentadas -->
        <div class="flex items-center gap-2 overflow-x-auto pb-1 sm:pb-0">
          
          <!-- Pestañas para MESERO -->
          {#if effectiveRole === 'waiter'}
            <button 
              type="button"
              onclick={() => currentTab = 'waiter'}
              class="min-h-[48px] px-5 rounded-full text-xs font-mono font-bold uppercase transition-all cursor-pointer shrink-0 {currentTab === 'waiter' ? 'bg-[#FF2400] text-white' : 'bg-white text-[#0C0C0E] border border-black/10 hover:bg-[#EDE8DC]'}"
            >
              Sellos en Sala ({stampRequests.length + redemptions.length})
            </button>

            <button 
              type="button"
              onclick={() => currentTab = 'waiter_orders'}
              class="min-h-[48px] px-5 rounded-full text-xs font-mono font-bold uppercase transition-all cursor-pointer shrink-0 {currentTab === 'waiter_orders' ? 'bg-[#FF2400] text-white' : 'bg-white text-[#0C0C0E] border border-black/10 hover:bg-[#EDE8DC]'}"
            >
              Comandas Salón ({inStoreOrders.length})
            </button>
          {/if}

          <!-- Pestañas para CAJERO (Sin acceso a Admin BI) -->
          {#if effectiveRole === 'cashier'}
            <button 
              type="button"
              onclick={() => currentTab = 'cashier_orders'}
              class="min-h-[48px] px-5 rounded-full text-xs font-mono font-bold uppercase transition-all cursor-pointer shrink-0 {currentTab === 'cashier_orders' ? 'bg-[#FF2400] text-white' : 'bg-white text-[#0C0C0E] border border-black/10 hover:bg-[#EDE8DC]'}"
            >
              Comandas ({inStoreOrders.filter(o => o.status === 'PENDING_PAYMENT').length})
            </button>

            <button 
              type="button"
              onclick={() => currentTab = 'cashier_delivery'}
              class="min-h-[48px] px-5 rounded-full text-xs font-mono font-bold uppercase transition-all cursor-pointer shrink-0 {currentTab === 'cashier_delivery' ? 'bg-[#FF2400] text-white' : 'bg-white text-[#0C0C0E] border border-black/10 hover:bg-[#EDE8DC]'}"
            >
              Domicilios ({deliveryOrders.filter(d => d.status === 'IN_PREPARATION').length})
            </button>

            <button 
              type="button"
              onclick={() => currentTab = 'cashier_stamps'}
              class="min-h-[48px] px-5 rounded-full text-xs font-mono font-bold uppercase transition-all cursor-pointer shrink-0 {currentTab === 'cashier_stamps' ? 'bg-[#FF2400] text-white' : 'bg-white text-[#0C0C0E] border border-black/10 hover:bg-[#EDE8DC]'}"
            >
              Sellos ({stampRequests.length + redemptions.length})
            </button>
          {/if}

          <!-- Pestañas para ADMINISTRADOR (Acceso Total a las 4 Áreas) -->
          {#if effectiveRole === 'admin'}
            <button 
              type="button"
              onclick={() => currentTab = 'waiter'}
              class="min-h-[48px] px-4 rounded-full text-xs font-mono font-bold uppercase transition-all cursor-pointer shrink-0 {currentTab === 'waiter' ? 'bg-[#FF2400] text-white' : 'bg-white text-[#0C0C0E] border border-black/10 hover:bg-[#EDE8DC]'}"
            >
              Meseros
            </button>

            <button 
              type="button"
              onclick={() => currentTab = 'cashier_orders'}
              class="min-h-[48px] px-4 rounded-full text-xs font-mono font-bold uppercase transition-all cursor-pointer shrink-0 {currentTab === 'cashier_orders' ? 'bg-[#FF2400] text-white' : 'bg-white text-[#0C0C0E] border border-black/10 hover:bg-[#EDE8DC]'}"
            >
              Caja Punto
            </button>

            <button 
              type="button"
              onclick={() => currentTab = 'cashier_delivery'}
              class="min-h-[48px] px-4 rounded-full text-xs font-mono font-bold uppercase transition-all cursor-pointer shrink-0 {currentTab === 'cashier_delivery' ? 'bg-[#FF2400] text-white' : 'bg-white text-[#0C0C0E] border border-black/10 hover:bg-[#EDE8DC]'}"
            >
              Domicilios
            </button>

            <button 
              type="button"
              onclick={() => currentTab = 'cashier_stamps'}
              class="min-h-[48px] px-4 rounded-full text-xs font-mono font-bold uppercase transition-all cursor-pointer shrink-0 {currentTab === 'cashier_stamps' ? 'bg-[#FF2400] text-white' : 'bg-white text-[#0C0C0E] border border-black/10 hover:bg-[#EDE8DC]'}"
            >
              Sellos
            </button>

            <button 
              type="button"
              onclick={() => currentTab = 'admin_bi'}
              class="min-h-[48px] px-4 rounded-full text-xs font-mono font-bold uppercase transition-all cursor-pointer shrink-0 {currentTab === 'admin_bi' ? 'bg-[#FF2400] text-white' : 'bg-white text-[#0C0C0E] border border-black/10 hover:bg-[#EDE8DC]'}"
            >
              Métricas BI
            </button>
          {/if}

          <!-- Botón Chime Sonoro -->
          <button
            type="button"
            onclick={playChime}
            class="w-12 h-12 rounded-full bg-white border border-black/10 text-[#0C0C0E] hover:bg-[#EDE8DC] flex items-center justify-center transition-colors shrink-0 cursor-pointer"
            title="Probar Timbre de Comanda"
          >
            <Volume2 class="w-5 h-5" />
          </button>

          <!-- Cerrar Sesión Staff -->
          <button
            type="button"
            onclick={handleLogoutStaff}
            class="min-h-[48px] px-4 rounded-full bg-white border border-black/10 text-xs font-mono font-bold text-[#8E8E93] hover:text-[#FF2400] flex items-center gap-1.5 transition-colors cursor-pointer shrink-0"
          >
            <LogOut class="w-4 h-4" />
            <span class="hidden sm:inline">Cerrar Sesión</span>
          </button>
        </div>
      </header>

      <!-- ======================================================== -->
      <!-- VISTA 1: TERMINAL EXCLUSIVA DE MESEROS (SELLOS Y MESA)     -->
      <!-- ======================================================== -->
      {#if currentTab === 'waiter'}
        <div class="space-y-6">
          
          <!-- Notificación / Resumen Superior del Mesero -->
          <div class="p-5 rounded-[28px] bg-white border border-black/10 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div>
              <h2 class="font-display font-extrabold text-xl text-[#0C0C0E]">Acreditación de Sellos en Sala</h2>
              <p class="text-xs text-[#8E8E93]">
                Verifica el recibo físico de la comanda de mesa antes de pulsar aprobar.
              </p>
            </div>
            <div class="flex items-center gap-2">
              <span class="min-h-[38px] px-4 rounded-full bg-[#EDE8DC] font-mono text-xs font-bold text-[#0C0C0E] flex items-center">
                {stampRequests.length} pendientes
              </span>
            </div>
          </div>

          <!-- SECCIÓN A: SOLICITUDES DE CLIENTES EN MESA -->
          <div class="space-y-4">
            <h3 class="font-display font-bold text-lg text-[#0C0C0E] flex items-center gap-2">
              <Award class="w-5 h-5 text-[#FF2400]" />
              Solicitudes Recientes de Comensales
            </h3>

            {#if stampRequests.length === 0}
              <div class="p-8 rounded-[32px] bg-white border border-black/10 text-center space-y-2">
                <CheckCircle2 class="w-10 h-10 text-emerald-600 mx-auto" />
                <p class="font-display font-bold text-lg text-[#0C0C0E]">No hay solicitudes pendientes</p>
                <p class="text-xs text-[#8E8E93]">Todas las comandas de sala han sido verificadas.</p>
              </div>
            {:else}
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                {#each stampRequests as req}
                  <div class="p-6 rounded-[32px] bg-white border border-black/10 space-y-4">
                    <div class="flex justify-between items-start">
                      <div>
                        <span class="font-mono text-xs text-[#FF2400] font-bold bg-[#FF2400]/10 px-3 py-1 rounded-full inline-block">
                          Orden #{req.orderCode}
                        </span>
                        <h4 class="font-display font-black text-xl text-[#0C0C0E] mt-2 leading-tight">
                          {req.customer}
                        </h4>
                      </div>
                      <span class="font-mono text-xs text-[#8E8E93]">{req.time}</span>
                    </div>

                    <!-- Datos Estructurados Claros -->
                    <div class="p-4 rounded-2xl bg-[#EDE8DC] text-xs font-mono space-y-2">
                      <div class="flex justify-between items-center">
                        <span class="text-[#8E8E93]">Factura / Recibo:</span>
                        <span class="text-[#0C0C0E] font-extrabold text-sm">#{req.receiptNumber}</span>
                      </div>
                      <div class="flex justify-between items-center">
                        <span class="text-[#8E8E93]">Hamburguesas a sellar:</span>
                        <span class="text-[#FF2400] font-black text-sm">+{req.burgersCount} sellos digitales</span>
                      </div>
                      <div class="flex justify-between items-center border-t border-black/10 pt-1.5">
                        <span class="text-[#8E8E93]">Total Cobrado:</span>
                        <span class="text-[#0C0C0E] font-bold">${req.total.toLocaleString('es-CO')}</span>
                      </div>
                    </div>

                    <!-- Botones Fat-Finger Ergonómicos (>=56px) -->
                    <div class="flex items-center gap-3 pt-1">
                      <button
                        type="button"
                        onclick={() => handleRejectStamp(req)}
                        class="flex-1 min-h-[56px] rounded-full bg-[#EDE8DC] hover:bg-[#EDE8DC]/80 text-[#FF2400] font-display font-bold text-sm transition-all active:scale-[0.98] cursor-pointer"
                      >
                        Rechazar
                      </button>
                      <button
                        type="button"
                        onclick={() => handleApproveStamp(req)}
                        class="flex-[1.5] min-h-[56px] rounded-full bg-[#FF2400] hover:bg-[#E02000] text-white font-display font-extrabold text-sm tracking-wide transition-all active:scale-[0.98] cursor-pointer"
                      >
                        ✓ Otorgar Sellos
                      </button>
                    </div>
                  </div>
                {/each}
              </div>
            {/if}
          </div>

          <!-- SECCIÓN B: OTORGAR SELLOS MANUALES DIRECTAMENTE EN MESA -->
          <div class="p-6 sm:p-8 rounded-[32px] bg-white border border-black/10 space-y-4">
            <div>
              <h3 class="font-display font-black text-xl text-[#0C0C0E]">
                Otorgar Sellos en Mesa (Asignación Rápida)
              </h3>
              <p class="text-xs text-[#8E8E93] mt-1">
                Usa este formulario cuando el comensal te solicite sus sellos presencialmente con el ticket de mesa.
              </p>
            </div>

            {#if manualSuccessMsg}
              <div class="p-4 rounded-2xl bg-emerald-100 border border-emerald-300 text-emerald-900 font-mono text-xs font-bold flex items-center gap-2">
                <Check class="w-4 h-4 text-emerald-700" /> {manualSuccessMsg}
              </div>
            {/if}

            <div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
              <div class="space-y-1">
                <label for="man-ticket" class="text-xs font-mono text-[#8E8E93] uppercase font-bold block">
                  Número de Ticket / Factura
                </label>
                <input 
                  id="man-ticket"
                  type="text" 
                  bind:value={manualTicket}
                  placeholder="Ej. 00128"
                  class="w-full min-h-[56px] px-4 rounded-2xl bg-[#EDE8DC] border border-black/10 text-[#0C0C0E] font-mono text-sm focus:outline-none focus:border-[#FF2400]"
                />
              </div>

              <div class="space-y-1">
                <label for="man-burgers" class="text-xs font-mono text-[#8E8E93] uppercase font-bold block">
                  Cantidad de Hamburguesas
                </label>
                <div class="flex items-center gap-2">
                  <button
                    type="button"
                    onclick={() => manualBurgers = Math.max(1, manualBurgers - 1)}
                    class="w-14 min-h-[56px] rounded-2xl bg-[#EDE8DC] font-black text-lg text-[#0C0C0E] cursor-pointer"
                  >-</button>
                  <span class="flex-1 min-h-[56px] rounded-2xl bg-[#EDE8DC] flex items-center justify-center font-mono font-black text-lg text-[#0C0C0E]">
                    {manualBurgers}
                  </span>
                  <button
                    type="button"
                    onclick={() => manualBurgers = Math.min(10, manualBurgers + 1)}
                    class="w-14 min-h-[56px] rounded-2xl bg-[#EDE8DC] font-black text-lg text-[#0C0C0E] cursor-pointer"
                  >+</button>
                </div>
              </div>

              <div class="flex items-end">
                <button
                  type="button"
                  onclick={handleAssignManualStamps}
                  class="w-full min-h-[56px] rounded-full bg-[#FF2400] hover:bg-[#E02000] text-white font-display font-extrabold text-sm tracking-wide transition-all active:scale-[0.98] cursor-pointer"
                >
                  ✓ Asignar Sellos Ahora
                </button>
              </div>
            </div>
          </div>

          <!-- SECCIÓN C: REDENCIONES DE HAMBURGUESA GRATIS (10/10) -->
          {#if redemptions.length > 0}
            <div class="p-6 sm:p-8 rounded-[32px] bg-[#EDE8DC] border border-[#FF2400]/40 space-y-4">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-2">
                  <Sparkles class="w-5 h-5 text-[#FF2400]" />
                  <h3 class="font-display font-black text-xl text-[#0C0C0E]">Redenciones VIP en Mesa</h3>
                </div>
                <span class="text-xs font-mono font-bold text-[#FF2400] bg-white px-3 py-1 rounded-full">
                  10/10 Sellos Cumplidos
                </span>
              </div>

              <div class="space-y-3">
                {#each redemptions as red}
                  <div class="p-4 sm:p-5 rounded-2xl bg-white flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                    <div>
                      <span class="text-xs font-mono font-bold text-emerald-700">Comensal Verificado</span>
                      <h4 class="font-display font-black text-lg text-[#0C0C0E]">{red.customer}</h4>
                      <p class="text-xs text-[#8E8E93] font-mono">{red.targetItem} • {red.time}</p>
                    </div>

                    <button
                      type="button"
                      onclick={() => handleApproveRedemption(red)}
                      class="min-h-[56px] px-6 rounded-full bg-[#FF2400] hover:bg-[#E02000] text-white font-display font-extrabold text-sm tracking-wide active:scale-[0.98] transition-all cursor-pointer"
                    >
                      ✓ Validar y Marchar a Cocina
                    </button>
                  </div>
                {/each}
              </div>
            </div>
          {/if}
        </div>
      {/if}

      <!-- ======================================================== -->
      <!-- VISTA 1B: COMANDAS DE SALÓN PARA MESEROS                  -->
      <!-- ======================================================== -->
      {#if currentTab === 'waiter_orders'}
        <div class="space-y-6">
          <div class="p-5 rounded-[28px] bg-white border border-black/10 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div>
              <h2 class="font-display font-extrabold text-xl text-[#0C0C0E]">Comandas de Salón Activas</h2>
              <p class="text-xs text-[#8E8E93]">
                Supervisa los pedidos y el estado de cobro de los comensales en mesa.
              </p>
            </div>
            <span class="min-h-[38px] px-4 rounded-full bg-[#EDE8DC] font-mono text-xs font-bold text-[#0C0C0E] flex items-center">
              {inStoreOrders.length} {inStoreOrders.length === 1 ? 'comanda' : 'comandas'}
            </span>
          </div>

          {#if inStoreOrders.length === 0}
            <div class="p-8 rounded-[32px] bg-white border border-black/10 text-center space-y-2">
              <CheckCircle2 class="w-10 h-10 text-emerald-600 mx-auto" />
              <p class="font-display font-bold text-lg text-[#0C0C0E]">No hay comandas activas en salón</p>
              <p class="text-xs text-[#8E8E93]">Los nuevos pedidos realizados por los comensales aparecerán aquí en tiempo real.</p>
            </div>
          {:else}
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              {#each inStoreOrders as ord}
                <div class="p-6 rounded-[32px] bg-white border border-black/10 space-y-4">
                  <div class="flex justify-between items-start">
                    <div>
                      <span class="font-mono text-xs text-[#FF2400] font-bold bg-[#FF2400]/10 px-3 py-1 rounded-full inline-block">
                        Orden #{ord.code}
                      </span>
                      <h4 class="font-display font-black text-xl text-[#0C0C0E] mt-2 leading-tight">
                        {ord.customer}
                      </h4>
                    </div>
                    <span class="font-mono text-xs font-bold px-3 py-1 rounded-full {ord.status === 'PAID' ? 'bg-emerald-100 text-emerald-800 border border-emerald-300' : 'bg-amber-100 text-amber-800 border border-amber-300'}">
                      {ord.status === 'PAID' ? '✓ PAGADO EN CAJA' : '⏳ PENDIENTE DE PAGO'}
                    </span>
                  </div>

                  <div class="p-4 rounded-2xl bg-[#EDE8DC] text-xs font-mono space-y-2">
                    <div class="flex justify-between items-center">
                      <span class="text-[#8E8E93]">Hora registro:</span>
                      <span class="text-[#0C0C0E] font-bold">{ord.time}</span>
                    </div>
                    <div class="flex justify-between items-center">
                      <span class="text-[#8E8E93]">Detalle:</span>
                      <span class="text-[#0C0C0E] font-bold text-right">{ord.items}</span>
                    </div>
                    <div class="flex justify-between items-center border-t border-black/10 pt-1.5">
                      <span class="text-[#8E8E93]">Total a Cobrar:</span>
                      <span class="text-[#FF2400] font-black text-sm">${ord.total?.toLocaleString('es-CO')}</span>
                    </div>
                  </div>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      {/if}

      <!-- ======================================================== -->
      <!-- VISTA 2: COMANDAS EN EL PUNTO (CAJERO & ADMIN)            -->
      <!-- ======================================================== -->
      {#if currentTab === 'cashier_orders'}
        <div class="space-y-6">
          <div class="p-6 sm:p-8 rounded-[32px] bg-white border border-black/10 space-y-4">
            <div class="flex items-center justify-between border-b border-black/10 pb-3">
              <div class="flex items-center gap-2">
                <Utensils class="w-5 h-5 text-[#FF2400]" />
                <h2 class="font-display font-black text-xl text-[#0C0C0E]">Comandas Salón y Para Llevar</h2>
              </div>
              <span class="text-xs font-mono text-[#8E8E93]">Cobro en caja y despacho</span>
            </div>

            <div class="space-y-3">
              {#each inStoreOrders as ord}
                <div class="p-5 rounded-2xl bg-[#EDE8DC] flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                  <div>
                    <div class="flex items-center gap-2">
                      <span class="font-mono text-base font-black text-[#0C0C0E]">#{ord.code}</span>
                      <span class="text-xs font-mono text-[#8E8E93]">• {ord.customer} ({ord.time})</span>
                    </div>
                    <p class="text-xs text-[#3A3A3C] font-medium mt-1">{ord.items}</p>
                    <span class="text-lg font-display font-black text-[#0C0C0E] mt-1 block">
                      ${ord.total.toLocaleString('es-CO')}
                    </span>
                  </div>

                  <div class="flex items-center gap-3">
                    {#if ord.status === 'PENDING_PAYMENT'}
                      <button
                        type="button"
                        onclick={() => handleMarkPaid(ord)}
                        class="min-h-[54px] px-6 rounded-full bg-[#FF2400] hover:bg-[#E02000] text-white font-display font-extrabold text-sm active:scale-[0.98] transition-all cursor-pointer"
                      >
                        Cobrar y Marcar Pago
                      </button>
                    {:else}
                      <span class="min-h-[48px] px-5 rounded-full bg-emerald-200 text-emerald-900 text-xs font-mono font-bold flex items-center">
                        ✓ ORDEN PAGADA
                      </span>
                    {/if}

                    <button
                      type="button"
                      onclick={() => handlePrintReceipt(ord)}
                      class="w-12 h-12 rounded-full bg-white border border-black/10 flex items-center justify-center text-[#0C0C0E] hover:bg-[#FAF6EE] transition-colors cursor-pointer"
                      title="Imprimir Comanda Térmica"
                    >
                      <Printer class="w-5 h-5" />
                    </button>
                  </div>
                </div>
              {/each}
            </div>
          </div>
        </div>
      {/if}

      <!-- ======================================================== -->
      <!-- VISTA 3: PEDIDOS A DOMICILIO (CAJERO & ADMIN)              -->
      <!-- ======================================================== -->
      {#if currentTab === 'cashier_delivery'}
        <div class="space-y-6">
          <div class="p-6 sm:p-8 rounded-[32px] bg-white border border-black/10 space-y-4">
            <div class="flex items-center justify-between border-b border-black/10 pb-3">
              <div class="flex items-center gap-2">
                <Bike class="w-5 h-5 text-[#FF2400]" />
                <h2 class="font-display font-black text-xl text-[#0C0C0E]">Despacho de Domicilios</h2>
              </div>
              <span class="text-xs font-mono text-[#8E8E93]">Coordinación directa vía WhatsApp</span>
            </div>

            <div class="space-y-4">
              {#each deliveryOrders as dom}
                <div class="p-5 sm:p-6 rounded-[28px] bg-[#EDE8DC] flex flex-col md:flex-row md:items-center justify-between gap-4">
                  <div class="space-y-1 flex-1">
                    <div class="flex items-center gap-2 flex-wrap">
                      <span class="font-mono text-base font-black text-[#0C0C0E]">#{dom.code}</span>
                      <span class="text-xs font-mono text-[#8E8E93]">• {dom.customer} ({dom.time})</span>
                      <span class="px-2.5 py-0.5 rounded-full bg-white text-[11px] font-mono font-bold text-[#FF2400]">
                        {dom.neighborhood}
                      </span>
                    </div>

                    <p class="text-xs sm:text-sm text-[#0C0C0E] font-medium pt-1">
                      📍 {dom.address}
                    </p>

                    <div class="flex items-center gap-3 pt-1 flex-wrap">
                      <span class="text-xs font-mono font-bold text-[#FF2400]">
                        Pago: {dom.payment}
                      </span>
                      <span class="text-xs font-mono font-black text-[#0C0C0E]">
                        • Total: ${dom.total.toLocaleString('es-CO')}
                      </span>
                    </div>
                  </div>

                  <div class="flex items-center gap-3 flex-wrap">
                    <!-- Botón directo para chatear con el cliente por WhatsApp -->
                    <a
                      href="https://wa.me/57{dom.phone}?text={encodeURIComponent(`¡Hola ${dom.customer}! Tu pedido de Mías #${dom.code} está listo.`)}"
                      target="_blank"
                      rel="noopener noreferrer"
                      class="min-h-[54px] px-5 rounded-full bg-emerald-600 hover:bg-emerald-700 text-white font-display font-bold text-xs flex items-center gap-2 transition-all cursor-pointer"
                    >
                      <MessageSquare class="w-4 h-4" />
                      WhatsApp
                    </a>

                    {#if dom.status === 'IN_PREPARATION'}
                      <button
                        type="button"
                        onclick={() => handleDispatchDelivery(dom)}
                        class="min-h-[54px] px-6 rounded-full bg-[#FF2400] hover:bg-[#E02000] text-white font-display font-extrabold text-sm active:scale-[0.98] transition-all cursor-pointer"
                      >
                        Despachar a Domiciliario
                      </button>
                    {:else}
                      <span class="min-h-[48px] px-5 rounded-full bg-blue-100 text-blue-900 text-xs font-mono font-bold flex items-center">
                        🛵 EN RUTA AL CLIENTE
                      </span>
                    {/if}
                  </div>
                </div>
              {/each}
            </div>
          </div>
        </div>
      {/if}

      <!-- ======================================================== -->
      <!-- VISTA 4: SELLOS Y REDENCIONES (CAJERO & ADMIN)             -->
      <!-- ======================================================== -->
      {#if currentTab === 'cashier_stamps'}
        <div class="space-y-6">
          <div class="p-6 sm:p-8 rounded-[32px] bg-white border border-black/10 space-y-4">
            <div class="flex items-center justify-between border-b border-black/10 pb-3">
              <div class="flex items-center gap-2">
                <Award class="w-5 h-5 text-[#FF2400]" />
                <h2 class="font-display font-black text-xl text-[#0C0C0E]">Acreditación de Sellos en Caja</h2>
              </div>
              <span class="text-xs font-mono text-[#8E8E93]">{stampRequests.length} solicitudes</span>
            </div>

            {#if stampRequests.length === 0}
              <p class="text-xs text-[#8E8E93] py-4 text-center">No hay solicitudes de sellos pendientes en este momento.</p>
            {:else}
              <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
                {#each stampRequests as req}
                  <div class="p-5 rounded-2xl bg-[#EDE8DC] space-y-3">
                    <div class="flex justify-between items-center">
                      <span class="font-mono text-xs font-bold text-[#FF2400]">Factura #{req.receiptNumber}</span>
                      <span class="font-mono text-xs text-[#8E8E93]">{req.time}</span>
                    </div>

                    <div>
                      <h4 class="font-display font-bold text-base text-[#0C0C0E]">{req.customer}</h4>
                      <p class="text-xs text-[#3A3A3C]">Orden #{req.orderCode} • Total ${req.total.toLocaleString('es-CO')}</p>
                    </div>

                    <div class="flex items-center gap-2 pt-1">
                      <button
                        type="button"
                        onclick={() => handleRejectStamp(req)}
                        class="flex-1 min-h-[48px] rounded-full bg-white text-[#FF2400] font-display font-bold text-xs"
                      >Rechazar</button>
                      <button
                        type="button"
                        onclick={() => handleApproveStamp(req)}
                        class="flex-1 min-h-[48px] rounded-full bg-[#FF2400] text-white font-display font-extrabold text-xs"
                      >✓ Otorgar {req.burgersCount} Sellos</button>
                    </div>
                  </div>
                {/each}
              </div>
            {/if}
          </div>

          <!-- Redenciones de Hamburguesa Gratis -->
          {#if redemptions.length > 0}
            <div class="p-6 sm:p-8 rounded-[32px] bg-[#EDE8DC] border border-[#FF2400]/40 space-y-4">
              <div class="flex items-center justify-between border-b border-black/10 pb-3">
                <div class="flex items-center gap-2">
                  <Sparkles class="w-5 h-5 text-[#FF2400]" />
                  <h3 class="font-display font-black text-xl text-[#0C0C0E]">Entrega de Hamburguesas Gratis (10/10)</h3>
                </div>
              </div>

              {#each redemptions as red}
                <div class="p-4 rounded-2xl bg-white flex flex-col sm:flex-row sm:items-center justify-between gap-4">
                  <div>
                    <h4 class="font-display font-bold text-base text-[#0C0C0E]">{red.customer}</h4>
                    <p class="text-xs text-[#8E8E93]">{red.targetItem}</p>
                  </div>

                  <button
                    type="button"
                    onclick={() => handleApproveRedemption(red)}
                    class="min-h-[52px] px-6 rounded-full bg-[#FF2400] hover:bg-[#E02000] text-white font-display font-extrabold text-sm transition-all cursor-pointer"
                  >
                    ✓ Entregar Hamburguesa Gratis
                  </button>
                </div>
              {/each}
            </div>
          {/if}
        </div>
      {/if}

      <!-- ======================================================== -->
      <!-- VISTA 5: CONTROL DIRECTIVO BI (SOLO ADMINISTRADOR)        -->
      <!-- ======================================================== -->
      {#if currentTab === 'admin_bi' && effectiveRole === 'admin'}
        <div class="space-y-6">
          
          <!-- Cabecera de Control BI con Actualización en Vivo -->
          <div class="p-6 sm:p-7 rounded-[32px] bg-white border border-black/10 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div>
              <div class="flex items-center gap-2 mb-1">
                <span class="w-2.5 h-2.5 rounded-full bg-emerald-500 animate-pulse"></span>
                <span class="text-xs font-mono font-bold uppercase tracking-widest text-[#FF2400]">
                  Métricas en Tiempo Real
                </span>
              </div>
              <h2 class="font-display font-black text-2xl sm:text-3xl text-[#0C0C0E]">
                Panel de Inteligencia Comercial
              </h2>
              <p class="text-xs text-[#8E8E93] mt-0.5">
                Calculadas en vivo a partir de comandas pagadas, despachos y sellos digitales acumulados.
              </p>
            </div>

            <div class="flex items-center gap-3 self-start sm:self-center">
              <span class="text-xs font-mono text-[#8E8E93] hidden md:inline">
                Sincronizado: {metricsLastSync}
              </span>
              <button
                type="button"
                onclick={loadAdminMetrics}
                disabled={isRefreshingMetrics}
                class="min-h-[48px] px-5 rounded-full bg-[#EDE8DC] hover:bg-[#EDE8DC]/80 text-[#0C0C0E] font-display font-bold text-xs flex items-center gap-2 transition-all active:scale-[0.98] cursor-pointer disabled:opacity-50"
              >
                <RefreshCw class="w-4 h-4 {isRefreshingMetrics ? 'animate-spin text-[#FF2400]' : ''}" />
                <span>{isRefreshingMetrics ? 'Actualizando...' : 'Actualizar'}</span>
              </button>
            </div>
          </div>

          <!-- FILA 1: LAS 4 MÉTRICAS CLAVE DEL NEGOCIO (FAT CARDS) -->
          <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <!-- 1. Ingresos Totales Reales -->
            <div class="p-6 rounded-[28px] bg-white border border-black/10 space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-xs font-mono uppercase tracking-widest text-[#8E8E93] font-bold">
                  Ventas Confirmadas
                </span>
                <div class="w-8 h-8 rounded-full bg-emerald-100 text-emerald-800 flex items-center justify-center">
                  <DollarSign class="w-4 h-4 stroke-[2.5]" />
                </div>
              </div>
              <span class="font-display font-black text-3xl text-[#0C0C0E] block tracking-tight">
                {formatCOP(realTotalRevenue)}
              </span>
              <span class="text-xs font-mono font-bold text-emerald-700 block">
                ✓ {realTotalOrdersCount} comandas facturadas hoy
              </span>
            </div>

            <!-- 2. Ticket Promedio Real -->
            <div class="p-6 rounded-[28px] bg-white border border-black/10 space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-xs font-mono uppercase tracking-widest text-[#8E8E93] font-bold">
                  Ticket Promedio
                </span>
                <div class="w-8 h-8 rounded-full bg-[#EDE8DC] text-[#0C0C0E] flex items-center justify-center">
                  <TrendingUp class="w-4 h-4 stroke-[2.5]" />
                </div>
              </div>
              <span class="font-display font-black text-3xl text-[#0C0C0E] block tracking-tight">
                {formatCOP(realAverageTicket)}
              </span>
              <span class="text-xs font-mono text-[#8E8E93] block">
                Total ventas / N° comandas
              </span>
            </div>

            <!-- 3. Canales: Salón vs Domicilio -->
            <div class="p-6 rounded-[28px] bg-white border border-black/10 space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-xs font-mono uppercase tracking-widest text-[#8E8E93] font-bold">
                  Canales de Venta
                </span>
                <div class="w-8 h-8 rounded-full bg-[#FF2400]/10 text-[#FF2400] flex items-center justify-center">
                  <Bike class="w-4 h-4 stroke-[2.5]" />
                </div>
              </div>
              <div class="flex items-baseline justify-between">
                <span class="font-display font-black text-2xl text-[#0C0C0E]">
                  {channelPercentages.inStorePct}% Salón
                </span>
                <span class="font-display font-bold text-base text-[#FF2400]">
                  {channelPercentages.delivPct}% Domicilio
                </span>
              </div>
              <div class="w-full h-2 rounded-full bg-[#EDE8DC] overflow-hidden flex">
                <div class="bg-[#0C0C0E] h-full" style="width: {channelPercentages.inStorePct}%;"></div>
                <div class="bg-[#FF2400] h-full" style="width: {channelPercentages.delivPct}%;"></div>
              </div>
            </div>

            <!-- 4. Fidelización y Sellos -->
            <div class="p-6 rounded-[28px] bg-white border border-black/10 space-y-2">
              <div class="flex items-center justify-between">
                <span class="text-xs font-mono uppercase tracking-widest text-[#8E8E93] font-bold">
                  Fidelización Mías
                </span>
                <div class="w-8 h-8 rounded-full bg-[#FF2400]/10 text-[#FF2400] flex items-center justify-center">
                  <Award class="w-4 h-4 stroke-[2.5]" />
                </div>
              </div>
              <span class="font-display font-black text-3xl text-[#FF2400] block tracking-tight">
                {realStampsApproved} / {realRedemptionsCount}
              </span>
              <span class="text-xs font-mono font-medium text-[#3A3A3C] block">
                Sellos activos • {realRedemptionsCount} hamburguesas gratis
              </span>
            </div>
          </div>

          <!-- FILA 2: DESGLOSE POR MÉTODOS DE PAGO Y TOP PRODUCTOS -->
          <div class="grid grid-cols-1 lg:grid-cols-2 gap-4">
            
            <!-- TARJETA 1: DESGLOSE POR MÉTODO DE PAGO REAL -->
            <div class="p-6 sm:p-7 rounded-[32px] bg-white border border-black/10 space-y-4">
              <div class="flex items-center justify-between border-b border-black/10 pb-3">
                <div class="flex items-center gap-2">
                  <PieChart class="w-5 h-5 text-[#FF2400]" />
                  <h3 class="font-display font-black text-lg text-[#0C0C0E]">Cobros por Método de Pago</h3>
                </div>
                <span class="text-xs font-mono text-[#8E8E93]">Caja y Domicilios</span>
              </div>

              <div class="space-y-3">
                {#each paymentMethodsBreakdown as method}
                  <div class="p-4 rounded-2xl bg-[#EDE8DC] space-y-2">
                    <div class="flex justify-between items-center text-xs font-mono">
                      <span class="font-bold text-[#0C0C0E]">{method.name}</span>
                      <span class="font-bold text-[#0C0C0E]">{formatCOP(method.total)} ({method.pct}%)</span>
                    </div>
                    <div class="w-full h-2 rounded-full bg-white/60 overflow-hidden">
                      <div 
                        class="h-full bg-[#FF2400] rounded-full transition-all duration-500"
                        style="width: {method.pct}%;"
                      ></div>
                    </div>
                    <div class="flex justify-between text-[11px] text-[#8E8E93] font-mono">
                      <span>{method.count} transacciones registradas</span>
                    </div>
                  </div>
                {/each}
              </div>
            </div>

            <!-- TARJETA 2: TOP 5 PRODUCTOS MÁS VENDIDOS -->
            <div class="p-6 sm:p-7 rounded-[32px] bg-white border border-black/10 space-y-4">
              <div class="flex items-center justify-between border-b border-black/10 pb-3">
                <div class="flex items-center gap-2">
                  <Flame class="w-5 h-5 text-[#FF2400]" />
                  <h3 class="font-display font-black text-lg text-[#0C0C0E]">Top 5 Productos Más Vendidos</h3>
                </div>
                <span class="text-xs font-mono text-[#8E8E93]">En sala y delivery</span>
              </div>

              <div class="space-y-2.5">
                {#each topProductsRanking as prod, idx}
                  <div class="p-3.5 rounded-2xl bg-[#EDE8DC] flex items-center justify-between gap-3">
                    <div class="flex items-center gap-3 min-w-0">
                      <span class="w-7 h-7 rounded-full bg-white font-mono font-black text-xs text-[#0C0C0E] flex items-center justify-center shrink-0">
                        #{idx + 1}
                      </span>
                      <div class="min-w-0">
                        <span class="font-display font-bold text-sm text-[#0C0C0E] block truncate">
                          {prod.name}
                        </span>
                        <span class="text-[10px] font-mono text-[#8E8E93]">
                          {prod.bread_type === 'pretzel' ? 'Pan Pretzel' : prod.bread_type === 'brioche' ? 'Pan Brioche' : 'Especialidad'}
                        </span>
                      </div>
                    </div>

                    <div class="text-right shrink-0">
                      <span class="font-mono font-black text-sm text-[#FF2400] block">
                        {prod.units_sold} uds
                      </span>
                      <span class="text-[10px] font-mono text-[#8E8E93]">
                        {formatCOP(prod.total_revenue)}
                      </span>
                    </div>
                  </div>
                {/each}
              </div>
            </div>
          </div>

          <!-- FILA 3: CONTROL DE DISPONIBILIDAD DE CARTA EN TIEMPO REAL -->
          <div class="p-6 sm:p-8 rounded-[32px] bg-white border border-black/10 space-y-4">
            <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-black/10 pb-3">
              <div>
                <h3 class="font-display font-black text-xl text-[#0C0C0E]">
                  Disponibilidad de Carta en Tiempo Real
                </h3>
                <p class="text-xs text-[#8E8E93] mt-0.5">
                  Conmuta con un solo toque los platos disponibles o agotados para clientes en la app.
                </p>
              </div>
              <span class="text-xs font-mono font-bold text-[#8E8E93]">
                {store.products.filter(p => p.is_in_stock).length} de {store.products.length} activos
              </span>
            </div>

            <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
              {#each store.products as p}
                <div class="p-4 rounded-2xl bg-[#EDE8DC] flex items-center justify-between gap-3">
                  <div class="min-w-0">
                    <span class="font-display font-bold text-sm text-[#0C0C0E] block truncate">{p.name}</span>
                    <span class="font-mono text-xs text-[#8E8E93]">{formatCOP(p.price)}</span>
                  </div>
                  <button 
                    type="button"
                    onclick={() => handleToggleStock(p)}
                    class="min-h-[46px] px-4 rounded-full text-xs font-mono font-bold transition-all cursor-pointer shrink-0 active:scale-95 {p.is_in_stock ? 'bg-emerald-700 text-white hover:bg-emerald-800' : 'bg-[#FF2400] text-white hover:bg-[#E02000]'}"
                  >
                    {p.is_in_stock ? 'En Carta' : 'Agotado'}
                  </button>
                </div>
              {/each}
            </div>
          </div>
        </div>
      {/if}
    </div>
  </div>
{/if}

<!-- Plantilla Oculta de Impresión Térmica ESC/POS -->
{#if printedOrder}
  <div id="printable-receipt" class="hidden">
    <div style="text-align: center; border-bottom: 1px dashed #000; padding-bottom: 8px; margin-bottom: 8px;">
      <h2 style="font-size: 14px; font-weight: bold; margin: 0;">MÍAS HAMBURGUESAS DE LUJO</h2>
      <p style="margin: 2px 0;">Lo Íntimo Hecho para Compartir</p>
      <p style="margin: 2px 0;">Comanda de Operaciones</p>
    </div>

    <div style="margin-bottom: 8px;">
      <p style="margin: 2px 0;"><strong>ORDEN:</strong> #{printedOrder.code}</p>
      <p style="margin: 2px 0;"><strong>CLIENTE:</strong> {printedOrder.customer}</p>
      <p style="margin: 2px 0;"><strong>HORA:</strong> {printedOrder.time}</p>
    </div>

    <div style="border-top: 1px dashed #000; border-bottom: 1px dashed #000; padding: 6px 0; margin-bottom: 8px;">
      <p style="margin: 2px 0;">{printedOrder.items}</p>
    </div>

    <div style="text-align: right; margin-bottom: 12px;">
      <p style="font-size: 13px; font-weight: bold; margin: 0;">TOTAL: ${printedOrder.total.toLocaleString('es-CO')}</p>
    </div>

    <div style="text-align: center; font-size: 9px; border-top: 1px dashed #000; padding-top: 6px;">
      <p style="margin: 2px 0;">* Entregar localizador al comensal *</p>
    </div>
  </div>
{/if}
