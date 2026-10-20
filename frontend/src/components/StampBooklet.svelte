<script>
  import { onMount } from 'svelte';
  import { store } from '../lib/store.svelte.js';
  import { requestStampAPI, requestRedemptionAPI, oauthLoginAPI, getOrderByCodeAPI } from '../lib/api.js';
  import { Award, Lock, Check, Clock, Sparkles, AlertCircle, ChevronLeft, Send, RefreshCw, Receipt } from 'lucide-svelte';
  import confetti from 'canvas-confetti';

  let receiptNumberInput = $state('');
  let customerEmailInput = $state('');
  let stampError = $state('');
  let stampSuccess = $state('');
  let isSubmitting = $state(false);
  let isAwaitingRedemptionApproval = $state(false);
  let redemptionError = $state('');

  // Orden activa para el modal de sellos
  let orderForModal = $derived(store.selectedOrderForStamp);

  // Autocompletar correo cuando se abre el modal de factura si no hay sesión
  $effect(() => {
    if (store.isReceiptModalOpen && !customerEmailInput) {
      if (store.user?.email) {
        customerEmailInput = store.user.email;
      } else if (orderForModal?.customer_phone) {
        customerEmailInput = `${orderForModal.customer_phone.replace(/\D/g, '')}@gmail.com`;
      }
    }
  });

  $effect(() => {
    if (!store.isReceiptModalOpen) return;
    const handleKeyDown = (e) => {
      if (e.key === 'Escape') {
        store.closeStampModal();
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  });

  async function submitStampRequest() {
    stampError = '';
    stampSuccess = '';

    const receiptNum = receiptNumberInput.trim();
    if (!receiptNum) {
      stampError = 'Por favor digita el número de factura o recibo físico (ej. 00124)';
      return;
    }

    const ord = orderForModal;
    if (!ord) {
      stampError = 'No se encontró la comanda vinculada.';
      return;
    }

    isSubmitting = true;

    // 1. Re-verificación inmediata contra el servidor
    if (ord.code) {
      try {
        const liveRes = await getOrderByCodeAPI(ord.code);
        if (liveRes && liveRes.order) {
          ord.status = liveRes.order.status;
          if (liveRes.order.id) ord.id = liveRes.order.id;
          store.updateOrderStatus(ord.code, liveRes.order.status, liveRes.order.id);
        }
      } catch {}
    }

    if (ord.status !== 'PAID' && ord.status !== 'DELIVERED') {
      isSubmitting = false;
      stampError = `La orden #${ord.code} aún no ha sido pagada en caja.`;
      return;
    }

    // 2. Si el cliente no tiene sesión iniciada, registrar/iniciar automáticamente
    let token = typeof window !== 'undefined' ? localStorage.getItem('mias_access_token') : null;
    if (!store.user || !token) {
      const emailToUse = customerEmailInput.trim() || 
        (ord.customer_phone ? `${ord.customer_phone.replace(/\D/g, '')}@gmail.com` : '') ||
        (ord.delivery_phone ? `${ord.delivery_phone.replace(/\D/g, '')}@gmail.com` : '');

      if (!emailToUse) {
        isSubmitting = false;
        stampError = 'Por favor ingresa tu correo electrónico para guardar tus sellos.';
        return;
      }

      try {
        const authData = await oauthLoginAPI({
          email: emailToUse,
          full_name: ord.customer_name || ord.delivery_name || emailToUse.split('@')[0],
          phone: ord.customer_phone || ord.delivery_phone || '',
          habeas_data_accepted: true
        });
        if (authData?.tokens) {
          localStorage.setItem('mias_access_token', authData.tokens.access_token);
          localStorage.setItem('mias_refresh_token', authData.tokens.refresh_token);
          store.setUser(authData.user);
          token = authData.tokens.access_token;
        }
      } catch (err) {
        isSubmitting = false;
        stampError = 'Error al registrar tu usuario: ' + (err.message || 'Verifica tu correo');
        return;
      }
    }

    // 3. Enviar solicitud de sellos
    try {
      if (!ord.id || ord.id.startsWith('ord-') || ord.id.startsWith('dom-')) {
        throw new Error('La orden aún no tiene un identificador de servidor válido. Intenta en unos segundos.');
      }

      await requestStampAPI(ord.id, receiptNum);

      // Actualizar inmediatamente el estado en el cliente
      store.updateOrderStampStatus(ord.code, 'PENDING');
      store.updateOrderStampStatus(ord.id, 'PENDING');
      if (orderForModal) {
        orderForModal.stamp_status = 'PENDING';
      }

      confetti({
        particleCount: 60,
        spread: 70,
        origin: { y: 0.6 }
      });

      stampSuccess = `¡Solicitud enviada para la factura #${receiptNum}! El cajero la autorizará en caja.`;
      receiptNumberInput = '';
      setTimeout(() => {
        store.closeStampModal();
      }, 900);
      await store.refreshStampBalance();
      await store.refreshUserOrders();
    } catch (err) {
      stampError = err.message || 'Error al conectar con el servidor para registrar sellos';
    } finally {
      isSubmitting = false;
    }
  }

  async function handleRedeem() {
    if (store.stampBalance < 10) return;
    redemptionError = '';
    isAwaitingRedemptionApproval = true;

    try {
      await requestRedemptionAPI();
    } catch (err) {
      isAwaitingRedemptionApproval = false;
      redemptionError = err.message || 'Error al solicitar redención en caja';
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

        {#if redemptionError}
          <p class="text-xs text-[#FF2400] font-medium flex items-center justify-center gap-1.5 pt-1">
            <AlertCircle class="w-3.5 h-3.5 shrink-0" /> {redemptionError}
          </p>
        {/if}
      </div>
    {/if}
  </div>

  <!-- ENLACE AL HISTORIAL DE PEDIDOS -->
  <div class="pt-4 text-center">
    <button
      type="button"
      onclick={() => store.currentView = 'orders'}
      class="w-full min-h-[60px] sm:min-h-[64px] rounded-full bg-char-card hover:bg-white/10 border-2 border-white/15 text-bone font-display font-black text-sm sm:text-base tracking-wide transition-all flex items-center justify-center gap-3 cursor-pointer shadow-xl active:scale-[0.97]"
    >
      <Receipt class="w-5 h-5 text-amber-400 stroke-[2.2]" />
      <span>Ver Mis Pedidos & Solicitar Sellos</span>
    </button>
  </div>
</div>

<!-- Modal Fat-Finger para Digitar el Número de Factura Física de la Orden -->
{#if store.isReceiptModalOpen}
  <div 
    data-lenis-prevent
    role="dialog"
    tabindex="-1"
    aria-modal="true"
    aria-label="Registrar Factura Física"
    onwheel={(e) => e.stopPropagation()}
    ontouchmove={(e) => e.stopPropagation()}
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/75 backdrop-blur-sm p-4"
  >
    <div class="bg-char w-full max-w-md rounded-[32px] border border-white/10 p-6 sm:p-8 space-y-5 relative text-bone shadow-2xl">
      <div>
        <span class="text-xs font-mono text-ash uppercase tracking-wider font-bold block mb-1">
          Acreditación de Compra
        </span>
        <h3 class="font-display font-black text-2xl text-bone leading-tight">
          Registrar Factura Física
        </h3>
      </div>

      {#if orderForModal}
        <div class="p-3.5 rounded-xl bg-char-card border border-white/10 flex items-center justify-between text-xs font-mono">
          <span>Comanda: <strong class="text-bone">#{orderForModal.code}</strong> • ${orderForModal.total?.toLocaleString('es-CO')}</span>
          <span class="px-2.5 py-0.5 rounded-full font-bold {
            orderForModal.status === 'PAID' ? 'bg-emerald-950/70 text-emerald-400' :
            orderForModal.status === 'DELIVERED' ? 'bg-purple-950/70 text-purple-400' :
            'bg-amber-950/70 text-amber-400'
          }">
            {orderForModal.status === 'PAID' ? 'PAGADA' : orderForModal.status === 'DELIVERED' ? 'ENTREGADA' : 'PENDIENTE'}
          </span>
        </div>
      {/if}

      <p class="text-xs text-ash leading-relaxed">
        Ingresa el número impreso en el ticket o recibo entregado en caja física o con tu comanda (ej. <strong class="text-bone">00124</strong> o <strong class="text-bone">4580</strong>).
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
      </div>

      {#if !store.user}
        <div class="space-y-1.5">
          <label for="sb-email" class="text-xs font-mono text-ash uppercase tracking-wider block font-bold">
            Tu Correo Electrónico (para guardar tus sellos)
          </label>
          <input 
            id="sb-email"
            type="email" 
            bind:value={customerEmailInput}
            placeholder="ejemplo@correo.com"
            class="w-full min-h-[54px] px-4 rounded-2xl bg-char-card border border-white/10 text-bone font-sans text-sm focus:outline-none focus:border-[#FF2400] transition-colors"
          />
        </div>
      {/if}

      {#if stampError}
        <p class="text-xs text-[#FF2400] font-medium flex items-center gap-1 pt-1">
          <AlertCircle class="w-3.5 h-3.5 shrink-0" /> {stampError}
        </p>
      {/if}

      {#if stampSuccess}
        <p class="text-xs text-emerald-400 font-medium flex items-center gap-1 pt-1">
          <Check class="w-3.5 h-3.5 shrink-0" /> {stampSuccess}
        </p>
      {/if}

      <div class="flex items-center gap-3 pt-2">
        <button 
          type="button"
          onclick={() => store.closeStampModal()}
          class="flex-1 min-h-[56px] rounded-full bg-char-card hover:bg-char-card-hover text-bone font-display font-bold text-sm transition-all cursor-pointer border border-white/10"
        >
          Cancelar
        </button>
        <button 
          type="button"
          onclick={submitStampRequest}
          disabled={isSubmitting}
          class="flex-1 min-h-[56px] rounded-full bg-[#FF2400] hover:bg-[#E02000] text-white font-display font-extrabold text-sm tracking-wide transition-all active:scale-[0.98] cursor-pointer disabled:opacity-50"
        >
          {isSubmitting ? 'Enviando...' : 'Enviar Solicitud'}
        </button>
      </div>
    </div>
  </div>
{/if}
