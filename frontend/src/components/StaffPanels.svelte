<script>
  import { store } from '../lib/store.svelte.js';
  import { playChime } from '../lib/audio.js';
  import { 
    CheckCircle, XCircle, Printer, Volume2, Shield, Utensils, 
    Bike, Award, DollarSign, TrendingUp, Sparkles, AlertTriangle 
  } from 'lucide-svelte';

  let currentStaffTab = $state('cashier'); // 'waiter', 'cashier', 'admin'
  let printedOrder = $state(null);

  // Muestras operativas en memoria para control de personal
  let inStoreOrders = $state([
    { id: '1', code: 'ORD-A101', customer: 'Andrés Gómez', total: 38000, status: 'PENDING_PAYMENT', items: '2x Dorada, 1x Combo', time: '18:42' },
    { id: '2', code: 'ORD-A102', customer: 'Valentina Ruiz', total: 23000, status: 'PAID', items: '1x Criolla', time: '18:45' }
  ]);

  let deliveryOrders = $state([
    { id: '3', code: 'DOM-B201', customer: 'Felipe Santos', phone: '310 987 6543', address: 'Cra 27 # 45-12, Apto 501', total: 51000, status: 'IN_PREPARATION', time: '18:50', payment: 'Efectivo (Paga $100.000)' }
  ]);

  let stampRequests = $state([
    { id: '4', receiptNumber: '00124', customer: 'Carlos Mendoza', orderCode: 'ORD-A098', burgersCount: 2, total: 46000, time: '18:40' },
    { id: '5', receiptNumber: '00125', customer: 'Lucía Morales', orderCode: 'ORD-A099', burgersCount: 1, total: 28000, time: '18:48' }
  ]);

  let redemptions = $state([
    { id: '6', customer: 'Marcos Peña', stampsCount: 10, targetItem: 'Hamburguesa Original Gratis', time: '18:52' }
  ]);

  function handleMarkPaid(order) {
    order.status = 'PAID';
    playChime();
  }

  function handlePrintReceipt(order) {
    printedOrder = order;
    setTimeout(() => {
      window.print();
    }, 150);
  }

  function handleApproveStamp(req) {
    stampRequests = stampRequests.filter(r => r.id !== req.id);
    playChime();
    alert(`Factura #${req.receiptNumber} aprobada. Se acreditaron ${req.burgersCount} sellos al cliente.`);
  }

  function handleRejectStamp(req) {
    stampRequests = stampRequests.filter(r => r.id !== req.id);
    alert(`Factura #${req.receiptNumber} rechazada.`);
  }

  function handleApproveRedemption(red) {
    redemptions = redemptions.filter(r => r.id !== red.id);
    playChime();
    alert(`Redención de ${red.customer} aprobada. Entregar Hamburguesa Original sin costo.`);
  }
</script>

<div class="max-w-6xl mx-auto px-4 py-8 space-y-6">
  <!-- Selector de Rol del Personal y Prueba de Chime Sonoro -->
  <div class="p-4 rounded-2xl bg-char-card border hairline-dark flex flex-wrap items-center justify-between gap-4">
    <div class="flex items-center gap-2">
      <Shield class="w-5 h-5 text-gold" />
      <span class="font-serif text-xl font-bold text-bone">Terminal de Operación Staff</span>
    </div>

    <div class="flex items-center gap-2">
      <!-- Selector de Vistas -->
      <button 
        onclick={() => currentStaffTab = 'waiter'}
        class="px-3.5 py-1.5 rounded-lg text-xs font-mono font-bold uppercase transition-colors {currentStaffTab === 'waiter' ? 'bg-gold text-char' : 'bg-char text-ash border hairline-dark'}"
      >
        Meseros
      </button>
      <button 
        onclick={() => currentStaffTab = 'cashier'}
        class="px-3.5 py-1.5 rounded-lg text-xs font-mono font-bold uppercase transition-colors {currentStaffTab === 'cashier' ? 'bg-gold text-char' : 'bg-char text-ash border hairline-dark'}"
      >
        Caja Central
      </button>
      <button 
        onclick={() => currentStaffTab = 'admin'}
        class="px-3.5 py-1.5 rounded-lg text-xs font-mono font-bold uppercase transition-colors {currentStaffTab === 'admin' ? 'bg-gold text-char' : 'bg-char text-ash border hairline-dark'}"
      >
        Admin BI
      </button>

      <!-- Botón de Prueba de Sonido -->
      <button
        onclick={playChime}
        class="p-2 rounded-lg bg-white/5 border hairline-dark text-gold hover:bg-gold/10 transition-colors ml-2"
        title="Probar sonido de notificación (Web Audio API)"
      >
        <Volume2 class="w-4 h-4" />
      </button>
    </div>
  </div>

  <!-- VISTA 1: PANEL DE MESERO (MÓVIL / BANDEJA DE SELLOS) -->
  {#if currentStaffTab === 'waiter'}
    <div class="space-y-4">
      <div class="flex items-center justify-between">
        <div>
          <h3 class="font-serif text-3xl font-bold text-bone">Bandeja de Sellos (Meseros)</h3>
          <p class="text-xs text-ash">Revisa las facturas registradas en mesa y aprueba los sellos correspondientes.</p>
        </div>
        <span class="font-mono text-xs text-gold bg-gold/10 px-3 py-1 rounded-full border border-gold/30">
          {stampRequests.length} Solicitudes
        </span>
      </div>

      <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
        {#each stampRequests as req}
          <div class="p-5 rounded-2xl bg-char-card border hairline-dark space-y-3 shadow-lux">
            <div class="flex justify-between items-start">
              <div>
                <span class="font-mono text-xs text-ember font-bold">Orden #{req.orderCode}</span>
                <h4 class="font-serif text-xl font-bold text-bone">{req.customer}</h4>
              </div>
              <span class="font-mono text-xs text-ash">{req.time}</span>
            </div>

            <div class="p-3 rounded-xl bg-char border hairline-dark text-xs font-mono space-y-1">
              <div class="flex justify-between">
                <span class="text-ash">Factura Física:</span>
                <span class="text-gold font-bold">#{req.receiptNumber}</span>
              </div>
              <div class="flex justify-between">
                <span class="text-ash">Hamburguesas a sellar:</span>
                <span class="text-bone font-bold">{req.burgersCount} sellos</span>
              </div>
              <div class="flex justify-between">
                <span class="text-ash">Monto Factura:</span>
                <span class="text-bone">${req.total.toLocaleString('es-CO')}</span>
              </div>
            </div>

            <div class="flex items-center gap-2 pt-2">
              <button
                onclick={() => handleRejectStamp(req)}
                class="flex-1 py-2.5 rounded-xl bg-white/5 hover:bg-ember/20 text-ember border border-ember/30 font-mono text-xs font-bold transition-colors"
              >
                Rechazar
              </button>
              <button
                onclick={() => handleApproveStamp(req)}
                class="flex-1 py-2.5 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-mono text-xs font-bold transition-colors shadow-lg"
              >
                ✓ Aprobar Sello
              </button>
            </div>
          </div>
        {/each}
      </div>
    </div>

  <!-- VISTA 2: PANEL DE CAJA CENTRAL (TABLET / DESKTOP) -->
  {:else if currentStaffTab === 'cashier'}
    <div class="space-y-6">
      <!-- 1. Comandas en el Punto: Por Cobrar -> Orden Paga -->
      <div class="p-6 rounded-3xl bg-char-card border hairline-dark space-y-4">
        <div class="flex items-center justify-between border-b hairline-dark pb-3">
          <div class="flex items-center gap-2">
            <Utensils class="w-5 h-5 text-gold" />
            <h3 class="font-serif text-2xl font-bold text-bone">Comandas en el Punto</h3>
          </div>
          <span class="text-xs font-mono text-ash">Cobro y entrega de localizador</span>
        </div>

        <div class="space-y-3">
          {#each inStoreOrders as ord}
            <div class="p-4 rounded-2xl bg-char border hairline-dark flex flex-col md:flex-row md:items-center justify-between gap-4">
              <div>
                <div class="flex items-center gap-2">
                  <span class="font-mono text-base font-bold text-bone">#{ord.code}</span>
                  <span class="text-xs text-ash">• {ord.customer} ({ord.time})</span>
                </div>
                <p class="text-xs text-ash mt-1">{ord.items}</p>
                <span class="text-sm font-mono font-bold text-gold mt-1 block">${ord.total.toLocaleString('es-CO')}</span>
              </div>

              <div class="flex items-center gap-3">
                {#if ord.status === 'PENDING_PAYMENT'}
                  <button
                    onclick={() => handleMarkPaid(ord)}
                    class="py-2.5 px-5 rounded-full bg-ember hover:bg-ember-light text-white font-mono font-bold text-xs shadow-ember transition-colors"
                  >
                    Por Cobrar (Recibir Pago)
                  </button>
                {:else}
                  <span class="text-xs font-mono font-bold text-emerald-400 bg-emerald-500/20 px-3 py-1 rounded-full border border-emerald-500/30">
                    ✓ ORDEN PAGA
                  </span>
                {/if}

                <!-- Botón Imprimir Comanda Térmica -->
                <button
                  onclick={() => handlePrintReceipt(ord)}
                  class="p-2.5 rounded-full bg-white/5 hover:bg-gold/20 text-gold border hairline-dark transition-colors"
                  title="Imprimir comanda térmica (80mm)"
                >
                  <Printer class="w-4 h-4" />
                </button>
              </div>
            </div>
          {/each}
        </div>
      </div>

      <!-- 2. Pedidos a Domicilio -->
      <div class="p-6 rounded-3xl bg-char-card border hairline-dark space-y-4">
        <div class="flex items-center justify-between border-b hairline-dark pb-3">
          <div class="flex items-center gap-2">
            <Bike class="w-5 h-5 text-ember" />
            <h3 class="font-serif text-2xl font-bold text-bone">Pedidos a Domicilio</h3>
          </div>
          <span class="text-xs font-mono text-gold">Coordinados vía WhatsApp</span>
        </div>

        <div class="space-y-3">
          {#each deliveryOrders as dom}
            <div class="p-4 rounded-2xl bg-char border hairline-dark flex flex-col md:flex-row md:items-center justify-between gap-4">
              <div>
                <div class="flex items-center gap-2">
                  <span class="font-mono text-base font-bold text-bone">#{dom.code}</span>
                  <span class="text-xs text-ash">• {dom.customer} ({dom.phone})</span>
                </div>
                <p class="text-xs text-ash mt-0.5">{dom.address}</p>
                <span class="text-xs font-mono text-gold block mt-1">{dom.payment}</span>
              </div>

              <div class="flex items-center gap-3">
                {#if dom.status === 'IN_PREPARATION'}
                  <button
                    onclick={() => dom.status = 'ON_THE_WAY'}
                    class="py-2.5 px-5 rounded-full bg-gold hover:bg-gold-light text-char font-mono font-bold text-xs shadow-gold transition-colors"
                  >
                    Despachar (En Camino)
                  </button>
                {:else}
                  <span class="text-xs font-mono font-bold text-blue-400 bg-blue-500/20 px-3 py-1 rounded-full border border-blue-500/30">
                    🛵 EN CAMINO
                  </span>
                {/if}
              </div>
            </div>
          {/each}
        </div>
      </div>

      <!-- 3. Redenciones de Hamburguesa Gratis -->
      {#if redemptions.length > 0}
        <div class="p-6 rounded-3xl bg-char-card border border-gold/40 space-y-4 shadow-gold">
          <div class="flex items-center justify-between border-b hairline-dark pb-3">
            <div class="flex items-center gap-2">
              <Sparkles class="w-5 h-5 text-gold animate-spin" style="animation-duration: 4s;" />
              <h3 class="font-serif text-2xl font-bold text-bone">Redenciones de Fidelización</h3>
            </div>
            <span class="text-xs font-mono text-gold bg-gold/15 px-2.5 py-0.5 rounded">Presencial</span>
          </div>

          {#each redemptions as red}
            <div class="p-4 rounded-2xl bg-char border hairline-dark flex items-center justify-between">
              <div>
                <span class="font-mono text-xs text-emerald-400 font-bold">10/10 Sellos Verificados</span>
                <h4 class="font-serif text-xl font-bold text-bone">{red.customer}</h4>
                <p class="text-xs text-ash">{red.targetItem} ($17.000 COP sin costo)</p>
              </div>

              <button
                onclick={() => handleApproveRedemption(red)}
                class="py-2.5 px-6 rounded-full bg-emerald-600 hover:bg-emerald-500 text-white font-mono font-bold text-xs shadow-lg"
              >
                ✓ Entregar y Quemar Sellos
              </button>
            </div>
          {/each}
        </div>
      {/if}
    </div>

  <!-- VISTA 3: DASHBOARD ADMINISTRATIVO BI (DUEÑO) -->
  {:else if currentStaffTab === 'admin'}
    <div class="space-y-6">
      <!-- Métricas Superiores -->
      <div class="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <div class="p-5 rounded-2xl bg-char-card border hairline-dark">
          <span class="text-xs font-mono uppercase tracking-widest text-ash block mb-1">Ventas Hoy</span>
          <span class="font-mono text-3xl font-bold text-gold">$1.480.000</span>
          <span class="text-[11px] text-emerald-400 block mt-1">↑ +18% vs semana anterior</span>
        </div>

        <div class="p-5 rounded-2xl bg-char-card border hairline-dark">
          <span class="text-xs font-mono uppercase tracking-widest text-ash block mb-1">Sellos Emitidos / Redimidos</span>
          <span class="font-mono text-3xl font-bold text-ember">42 / 4</span>
          <span class="text-[11px] text-ash block mt-1">9.5% de tasa de redención</span>
        </div>

        <div class="p-5 rounded-2xl bg-char-card border hairline-dark">
          <span class="text-xs font-mono uppercase tracking-widest text-ash block mb-1">Ticket Promedio</span>
          <span class="font-mono text-3xl font-bold text-bone">$39.500</span>
          <span class="text-[11px] text-ash block mt-1">Consumo en punto domina 65%</span>
        </div>
      </div>

      <!-- CMS: Conmutador Inmediato de Stock -->
      <div class="p-6 rounded-3xl bg-char-card border hairline-dark space-y-4">
        <h3 class="font-serif text-2xl font-bold text-bone">Control de Disponibilidad en Tiempo Real</h3>
        <p class="text-xs text-ash">Desactiva productos agotados con 1 solo clic si se acaban ingredientes en el servicio.</p>

        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
          {#each store.products.slice(0, 6) as p}
            <div class="p-3 rounded-xl bg-char border hairline-dark flex items-center justify-between">
              <div>
                <span class="font-serif text-base font-bold text-bone">{p.name}</span>
                <span class="font-mono text-xs text-gold block">${p.price.toLocaleString('es-CO')}</span>
              </div>
              <button 
                onclick={() => p.is_in_stock = !p.is_in_stock}
                class="px-3 py-1 rounded-full text-xs font-mono font-bold transition-colors {p.is_in_stock ? 'bg-emerald-500/20 text-emerald-400 border border-emerald-500/30' : 'bg-red-500/20 text-red-400 border border-red-500/30'}"
              >
                {p.is_in_stock ? 'En Stock' : 'Agotado'}
              </button>
            </div>
          {/each}
        </div>
      </div>
    </div>
  {/if}
</div>

<!-- Plantilla Oculta de Impresión Térmica ESC/POS (Visible solo al imprimir) -->
{#if printedOrder}
  <div id="printable-receipt" class="hidden">
    <div style="text-align: center; border-bottom: 1px dashed #000; padding-bottom: 8px; margin-bottom: 8px;">
      <h2 style="font-size: 14px; font-weight: bold; margin: 0;">MÍAS HAMBURGUESAS DE LUJO</h2>
      <p style="margin: 2px 0;">Lo Íntimo Hecho para Compartir</p>
      <p style="margin: 2px 0;">Comanda de Cocina / Sala</p>
    </div>

    <div style="margin-bottom: 8px;">
      <p style="margin: 2px 0;"><strong>ORDEN:</strong> #{printedOrder.code}</p>
      <p style="margin: 2px 0;"><strong>CLIENTE:</strong> {printedOrder.customer}</p>
      <p style="margin: 2px 0;"><strong>FECHA/HORA:</strong> {printedOrder.time}</p>
    </div>

    <div style="border-top: 1px dashed #000; border-bottom: 1px dashed #000; padding: 6px 0; margin-bottom: 8px;">
      <p style="margin: 2px 0;">{printedOrder.items}</p>
    </div>

    <div style="text-align: right; margin-bottom: 12px;">
      <p style="font-size: 13px; font-weight: bold; margin: 0;">TOTAL: ${printedOrder.total.toLocaleString('es-CO')}</p>
    </div>

    <div style="text-align: center; font-size: 9px; border-top: 1px dashed #000; padding-top: 6px;">
      <p style="margin: 2px 0;">* Entregar localizador físico al cliente *</p>
    </div>
  </div>
{/if}
