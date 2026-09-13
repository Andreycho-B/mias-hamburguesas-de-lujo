<script>
  import { store } from '../lib/store.svelte.js';
  import { requestStampAPI, requestRedemptionAPI } from '../lib/api.js';
  import { Award, Lock, Check, Clock, Sparkles, AlertCircle, ChevronLeft } from 'lucide-svelte';
  import confetti from 'canvas-confetti';

  let receiptModalOpen = $state(false);
  let selectedOrderForStamp = $state(null);
  let receiptNumberInput = $state('');
  let pendingStampRequests = $state([]);
  let isAwaitingRedemptionApproval = $state(false);
  let stampError = $state('');

  function openReceiptModal(order) {
    selectedOrderForStamp = order;
    receiptNumberInput = '';
    stampError = '';
    receiptModalOpen = true;
  }

  async function submitStampRequest() {
    if (!receiptNumberInput.trim()) {
      stampError = 'Por favor digita el número de factura física';
      return;
    }

    const receiptNum = receiptNumberInput.trim();
    const burgersCount = selectedOrderForStamp?.items?.filter(it => it.product?.bread_type === 'pretzel')?.reduce((s, it) => s + it.quantity, 0) || 1;

    const isOnlineOrder = selectedOrderForStamp?.id && !selectedOrderForStamp.id.startsWith('ord-');
    let apiSuccess = false;

    try {
      if (isOnlineOrder) {
        await requestStampAPI(selectedOrderForStamp.id, receiptNum);
        apiSuccess = true;
      }
    } catch (err) {
      console.warn('[Stamps] Error conectando con API, registrando en Anillo 0', err);
    }

    const newReq = {
      id: 'req-' + Date.now(),
      orderCode: selectedOrderForStamp?.code || 'ORD-' + Math.floor(1000 + Math.random() * 9000),
      receiptNumber: receiptNum,
      burgersCount,
      status: 'PENDING',
      createdAt: new Date().toLocaleTimeString()
    };

    pendingStampRequests = [newReq, ...pendingStampRequests];
    receiptModalOpen = false;

    // Solo si opera 100% offline o en modo demo local sin backend, simular tras 5s
    if (!apiSuccess && !isOnlineOrder) {
      setTimeout(() => {
        newReq.status = 'APPROVED';
        const newTotal = Math.min(10, store.stampBalance + newReq.burgersCount);
        store.setStampBalance(newTotal);

        if (newTotal === 10) {
          confetti({
            particleCount: 120,
            spread: 70,
            origin: { y: 0.6 }
          });
        }
      }, 5000);
    }
  }

  async function handleRedeem() {
    if (store.stampBalance < 10) return;
    isAwaitingRedemptionApproval = true;

    try {
      await requestRedemptionAPI();
    } catch (err) {
      console.warn('[Redemption] Error registrando redención en API', err);
    }
  }
</script>

<div class="max-w-3xl mx-auto px-4 py-6 sm:py-8 space-y-6">
  <!-- Barra Superior con Botón Atrás Oficial del Menú -->
  <div class="flex items-center justify-between">
    <button
      type="button"
      onclick={() => store.currentView = 'menu'}
      class="w-14 h-14 sm:w-16 sm:h-16 rounded-full bg-char-card border border-white/10 flex items-center justify-center text-bone hover:bg-white/10 active:scale-90 transition-all cursor-pointer group shrink-0"
      aria-label="Volver al Menú"
    >
      <ChevronLeft class="w-6 h-6 sm:w-7 sm:h-7 stroke-[2.2] -translate-x-0.5 group-hover:-translate-x-1 transition-transform" />
    </button>

    <span class="text-xs font-mono font-bold tracking-widest text-ash uppercase">
      Mías Fidelización
    </span>

    <div class="w-14 sm:w-16"></div>
  </div>

  <!-- Cabecera Editorial con Foto Real del Candado -->
  <div class="rounded-[32px] bg-char-card border border-white/10 shadow-xl p-6 sm:p-8 flex flex-col md:flex-row items-center gap-6 sm:gap-8">
    <div class="relative w-32 h-44 sm:w-40 sm:h-52 shrink-0 rounded-2xl overflow-hidden border border-white/10">
      <img 
        src="/imagenes/historia/candado_branding_historia.webp" 
        alt="Candado de Corazón Mías" 
        class="w-full h-full object-cover"
      />
      <div class="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent"></div>
      <span class="absolute bottom-2 left-2 text-[10px] font-mono text-white/90 uppercase tracking-widest font-bold">
        Amor para siempre
      </span>
    </div>

    <div class="space-y-3 text-center md:text-left flex-1">
      <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-[#FF2400]/10 border border-[#FF2400]/20 text-[#FF2400] text-xs font-mono font-bold uppercase tracking-wider">
        <Award class="w-4 h-4" /> Fidelización Mías
      </div>
      <h1 class="font-display font-black text-2xl sm:text-4xl text-bone tracking-tight leading-tight">
        Tu Cartilla de Sellos
      </h1>
      <p class="text-xs sm:text-sm text-ash leading-relaxed font-sans">
        Por cada hamburguesa al carbón que pidas y pagues, acumulas <strong class="text-bone">1 sello digital</strong>. 
        Al completar <strong class="text-bone">10 sellos</strong>, ganas una <strong class="text-bone">Hamburguesa Original gratis</strong> ($17.000 COP).
      </p>
      <p class="text-[11px] font-mono text-ash/80">
        * Válido 6 meses. Redención presencial en el local con autorización del equipo.
      </p>
    </div>
  </div>

  <!-- Retícula de 10 Sellos Digitales -->
  <div class="p-6 sm:p-8 rounded-[32px] bg-char-card border border-white/10 shadow-xl space-y-6">
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-white/10 pb-4">
      <div>
        <span class="text-xs font-mono uppercase tracking-widest text-ash font-bold">Progreso Actual</span>
        <h2 class="font-display font-extrabold text-2xl sm:text-3xl text-bone">
          {store.stampBalance} de 10 Sellos
        </h2>
      </div>
      <div>
        {#if store.stampBalance >= 10}
          <span class="inline-flex items-center gap-1.5 px-4 py-2 rounded-full bg-emerald-950/60 text-emerald-400 border border-emerald-700/30 text-xs font-mono font-bold">
            <Sparkles class="w-4 h-4" /> ¡META ALCANZADA!
          </span>
        {:else}
          <span class="text-xs font-mono text-[#FF2400] font-bold bg-[#FF2400]/10 border border-[#FF2400]/20 px-3.5 py-1.5 rounded-full inline-block">
            Faltan {10 - store.stampBalance} para tu gratis
          </span>
        {/if}
      </div>
    </div>

    <!-- Barra de Progreso Visual -->
    <div class="w-full h-3 rounded-full bg-char border border-white/5 overflow-hidden">
      <div 
        class="h-full bg-[#FF2400] transition-all duration-500 rounded-full"
        style="width: {(store.stampBalance / 10) * 100}%;"
      ></div>
    </div>

    <!-- Grid 2x5 de Casillas Fat-Finger -->
    <div class="grid grid-cols-5 gap-2.5 sm:gap-4">
      {#each Array(10) as _, index}
        {@const isStamped = index < store.stampBalance}
        {@const isLast = index === 9}
        <div 
          class="aspect-square rounded-2xl border transition-all duration-300 flex flex-col items-center justify-center p-2 relative select-none {isStamped ? 'bg-[#FF2400]/10 border-[#FF2400] text-[#FF2400] shadow-[0_0_12px_rgba(255,36,0,0.15)]' : 'bg-char border-white/10 text-ash'} {isLast && !isStamped ? 'border-dashed border-[#FF2400]/60' : ''}"
        >
          {#if isStamped}
            <img 
              src="/imagenes/marca/isotipo.webp" 
              alt="Sello" 
              class="w-7 h-7 sm:w-9 sm:h-9 object-contain"
            />
            <span class="text-[10px] font-mono font-black text-[#FF2400] mt-1">✓ Sello</span>
          {:else}
            <Lock class="w-4 h-4 sm:w-5 sm:h-5 opacity-40 mb-0.5" />
            <span class="text-[10px] font-mono font-bold opacity-60">{String(index + 1).padStart(2, '0')}</span>
            {#if isLast}
              <span class="text-[8px] font-mono font-black text-[#FF2400] uppercase mt-0.5">GRATIS</span>
            {/if}
          {/if}
        </div>
      {/each}
    </div>

    <!-- Banner de Redención Si tiene 10/10 -->
    {#if store.stampBalance >= 10}
      <div class="p-6 rounded-[28px] bg-char border border-[#FF2400]/30 text-center space-y-3">
        <Sparkles class="w-8 h-8 text-[#FF2400] mx-auto animate-bounce" />
        <h3 class="font-display font-extrabold text-2xl sm:text-3xl text-bone">¡Hamburguesa Original Gratis Lista!</h3>
        <p class="text-xs text-ash max-w-md mx-auto">
          Presenta tu app al mesero o en la caja física de Mías para que validen y entreguen tu hamburguesa.
        </p>
        
        {#if !isAwaitingRedemptionApproval}
          <button
            type="button"
            onclick={handleRedeem}
            class="min-h-[58px] px-8 rounded-full bg-[#FF2400] hover:bg-[#E02000] text-white font-display font-extrabold text-base tracking-wide transition-all active:scale-[0.98] cursor-pointer shadow-none"
          >
            Redimir Hamburguesa Original Gratis
          </button>
        {:else}
          <div class="min-h-[54px] px-6 rounded-full bg-char-card border border-white/10 text-bone font-mono text-xs flex items-center justify-center gap-2 max-w-xs mx-auto animate-pulse">
            <Clock class="w-4 h-4 text-[#FF2400]" /> A la espera de aprobación en caja...
          </div>
        {/if}
      </div>
    {/if}
  </div>

  <!-- Órdenes del Cliente y Solicitud de Sellos -->
  <div class="p-6 sm:p-8 rounded-[32px] bg-char-card border border-white/10 shadow-xl space-y-4">
    <div class="flex items-center justify-between border-b border-white/10 pb-3">
      <div>
        <h2 class="font-display font-extrabold text-xl sm:text-2xl text-bone">Tus Órdenes Recientes</h2>
        <span class="text-xs font-mono text-ash">Solicita tus sellos tras realizar tu pedido</span>
      </div>
    </div>

    {#if store.activeOrders.length === 0}
      <div class="py-8 text-center space-y-3">
        <p class="text-sm text-ash">Aún no tienes órdenes registradas.</p>
        <button
          type="button"
          onclick={() => store.currentView = 'menu'}
          class="min-h-[50px] px-6 rounded-full bg-white/10 hover:bg-white/15 text-bone border border-white/10 font-display font-bold text-xs uppercase tracking-wider transition-all cursor-pointer"
        >
          Explorar la Carta
        </button>
      </div>
    {:else}
      <div class="space-y-3">
        {#each store.activeOrders as ord}
          <div class="p-4 sm:p-5 rounded-2xl bg-char border border-white/5 flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div>
              <div class="flex items-center gap-2">
                <span class="font-mono text-sm font-bold text-bone">#{ord.code}</span>
                <span class="text-[10px] font-mono uppercase px-2.5 py-1 rounded-full font-bold {
                  ord.status === 'PAID' ? 'bg-emerald-950/60 text-emerald-400 border border-emerald-700/30' :
                  ord.status === 'IN_PREPARATION' ? 'bg-amber-950/60 text-amber-400 border border-amber-700/30' :
                  ord.status === 'ON_THE_WAY' ? 'bg-blue-950/60 text-blue-400 border border-blue-700/30' :
                  ord.status === 'DELIVERED' ? 'bg-purple-950/60 text-purple-400 border border-purple-700/30' :
                  'bg-orange-950/60 text-orange-400 border border-orange-700/30'
                }">
                  {#if ord.status === 'PAID'}
                    ✓ ORDEN PAGA
                  {:else if ord.status === 'IN_PREPARATION'}
                    🔥 EN PREPARACIÓN
                  {:else if ord.status === 'ON_THE_WAY'}
                    🛵 EN CAMINO
                  {:else if ord.status === 'DELIVERED'}
                    ✨ ENTREGADO
                  {:else}
                    ⏳ PENDIENTE DE PAGO
                  {/if}
                </span>
                <span class="text-xs text-ash font-mono">• {ord.type === 'in_store' ? 'En Salón' : ord.type === 'takeout' ? 'Para Llevar' : 'A Domicilio'}</span>
              </div>
              <span class="text-sm text-bone font-display font-bold mt-1 block">${ord.total.toLocaleString('es-CO')}</span>
            </div>

            <div>
              <button
                type="button"
                onclick={() => openReceiptModal(ord)}
                class="min-h-[48px] px-5 rounded-full bg-[#FF2400] hover:bg-[#E02000] text-white font-display font-extrabold text-xs flex items-center gap-2 transition-all active:scale-[0.98] cursor-pointer"
              >
                <Award class="w-4 h-4" /> Solicitar Sellos
              </button>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<!-- Modal Fat-Finger para Digitar el Número de Factura Física -->
{#if receiptModalOpen}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div 
    data-lenis-prevent
    onwheel={(e) => e.stopPropagation()}
    ontouchmove={(e) => e.stopPropagation()}
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4"
  >
    <div class="bg-char w-full max-w-md rounded-[32px] border border-white/10 p-6 sm:p-8 space-y-5 relative text-bone">
      <div>
        <span class="text-xs font-mono text-ash uppercase tracking-wider font-bold block mb-1">
          Acreditación de Compra
        </span>
        <h3 class="font-display font-black text-2xl text-bone leading-tight">
          Registrar Factura Física
        </h3>
      </div>

      <p class="text-xs text-ash leading-relaxed">
        Ingresa el número impreso en el ticket o recibo que te entregaron en caja o con tu domicilio (ej. <strong class="text-bone">00124</strong> o <strong class="text-bone">4580</strong>).
      </p>

      <div class="space-y-1.5">
        <label for="sb-receipt" class="text-xs font-mono text-ash uppercase tracking-wider block font-bold">
          Número de Factura / Ticket
        </label>
        <input 
          id="sb-receipt"
          type="text" 
          bind:value={receiptNumberInput}
          placeholder="Ej. 00124"
          class="w-full min-h-[56px] px-4 rounded-2xl bg-char-card border border-white/10 text-bone font-mono text-base focus:outline-none focus:border-[#FF2400] transition-colors"
        />
        {#if stampError}
          <p class="text-xs text-[#FF2400] font-medium flex items-center gap-1 pt-1">
            <AlertCircle class="w-3.5 h-3.5" /> {stampError}
          </p>
        {/if}
      </div>

      <div class="flex items-center gap-3 pt-2">
        <button 
          type="button"
          onclick={() => receiptModalOpen = false}
          class="flex-1 min-h-[56px] rounded-full bg-char-card hover:bg-char-card-hover text-bone font-display font-bold text-sm transition-all cursor-pointer border border-white/10"
        >
          Cancelar
        </button>
        <button 
          type="button"
          onclick={submitStampRequest}
          class="flex-1 min-h-[56px] rounded-full bg-[#FF2400] hover:bg-[#E02000] text-white font-display font-extrabold text-sm tracking-wide transition-all active:scale-[0.98] cursor-pointer"
        >
          Enviar Solicitud
        </button>
      </div>
    </div>
  </div>
{/if}
