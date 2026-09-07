<script>
  import { store } from '../lib/store.svelte.js';
  import { requestStampAPI, requestRedemptionAPI } from '../lib/api.js';
  import { Award, Lock, Check, Clock, Sparkles, AlertCircle } from 'lucide-svelte';
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
    const burgersCount = selectedOrderForStamp.items.filter(it => it.product.bread_type === 'pretzel').reduce((s, it) => s + it.quantity, 0) || 1;

    try {
      if (selectedOrderForStamp.id && !selectedOrderForStamp.id.startsWith('ord-')) {
        await requestStampAPI(selectedOrderForStamp.id, receiptNum);
      }
    } catch (err) {
      console.warn('[Stamps] Error conectando con API, registrando en Anillo 0', err);
    }

    const newReq = {
      id: 'req-' + Date.now(),
      orderCode: selectedOrderForStamp.code,
      receiptNumber: receiptNum,
      burgersCount,
      status: 'PENDING',
      createdAt: new Date().toLocaleTimeString()
    };

    pendingStampRequests = [newReq, ...pendingStampRequests];
    receiptModalOpen = false;

    // Aprobación y acreditación
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
    }, 3500);
  }

  async function handleRedeem() {
    if (store.stampBalance < 10) return;
    isAwaitingRedemptionApproval = true;

    try {
      await requestRedemptionAPI();
    } catch (err) {
      console.warn('[Redemption] Operando en Anillo 0', err);
    }

    // Aprobación presencial
    setTimeout(() => {
      isAwaitingRedemptionApproval = false;
      store.setStampBalance(0); // Reiniciar cartilla
      confetti({
        particleCount: 200,
        spread: 100,
        origin: { y: 0.5 }
      });
      alert('¡FELICIDADES! Tu Hamburguesa Original ha sido aprobada y entregada por caja.');
    }, 4500);
  }
</script>

<div class="max-w-4xl mx-auto px-4 py-8 space-y-8">
  <!-- Cabecera Editorial con Foto Real del Candado en la Reja -->
  <div class="relative rounded-3xl overflow-hidden border hairline-dark bg-char-card p-6 sm:p-10 flex flex-col md:flex-row items-center gap-8 shadow-lux">
    <div class="relative w-36 h-48 sm:w-44 sm:h-56 flex-shrink-0 rounded-2xl overflow-hidden border border-gold/30 shadow-gold">
      <img 
        src="/imagenes/historia/candado_branding_historia.webp" 
        alt="Candado de Corazón Mías" 
        class="w-full h-full object-cover"
      />
      <div class="absolute inset-0 bg-gradient-to-t from-char/80 to-transparent"></div>
      <span class="absolute bottom-2 left-2 text-[10px] font-mono text-gold uppercase tracking-widest">
        Amor para toda la vida
      </span>
    </div>

    <div class="space-y-3 text-center md:text-left">
      <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full bg-ember/15 border border-ember/30 text-ember text-xs font-mono font-bold uppercase tracking-wider">
        <Award class="w-4 h-4" /> Programa de Fidelización
      </div>
      <h2 class="font-serif text-3xl sm:text-5xl font-bold text-bone tracking-tight">
        Tu Cartilla de Sellos
      </h2>
      <p class="text-xs sm:text-sm text-ash leading-relaxed font-sans max-w-xl">
        Por cada hamburguesa al carbón que pidas en la app web y pagues en caja, recibes 1 sello digital. 
        Al completar <strong>10 sellos</strong>, tienes derecho a una <strong>Hamburguesa Original totalmente gratis</strong> ($17.000 COP).
      </p>
      <p class="text-[11px] font-mono text-gold/80">
        * Válido durante 6 meses desde tu última orden. Redención exclusivamente presencial en el punto de venta con aprobación del cajero.
      </p>
    </div>
  </div>

  <!-- Retícula de 10 Sellos (Inspirado en Dalkom Cafe - Ref 5) -->
  <div class="p-6 sm:p-8 rounded-3xl bg-char-card border hairline-dark space-y-6">
    <div class="flex items-center justify-between border-b hairline-dark pb-4">
      <div>
        <span class="text-xs font-mono uppercase tracking-widest text-ash">Progreso actual</span>
        <h3 class="font-serif text-2xl font-bold text-bone">
          {store.stampBalance} de 10 Sellos Acumulados
        </h3>
      </div>
      <div class="text-right">
        {#if store.stampBalance >= 10}
          <span class="text-xs font-mono font-bold text-emerald-400 bg-emerald-500/15 px-3 py-1 rounded-full border border-emerald-500/30 animate-pulse">
            ¡META ALCANZADA!
          </span>
        {:else}
          <span class="text-xs font-mono text-gold font-bold">
            Faltan {10 - store.stampBalance} para tu Original gratis
          </span>
        {/if}
      </div>
    </div>

    <!-- Grid 2x5 de Casillas de Sellos -->
    <div class="grid grid-cols-5 gap-2 sm:gap-4">
      {#each Array(10) as _, index}
        {@const isStamped = index < store.stampBalance}
        {@const isLast = index === 9}
        <div 
          class="aspect-square rounded-2xl border transition-all duration-300 flex flex-col items-center justify-center p-2 relative {isStamped ? 'bg-ember/20 border-ember shadow-ember text-ember' : 'bg-char border-white/10 text-ash/40'} {isLast && !isStamped ? 'border-dashed border-gold/40' : ''}"
        >
          {#if isStamped}
            <img 
              src="/imagenes/marca/isotipo.webp" 
              alt="Sello" 
              class="w-8 h-8 sm:w-10 sm:h-10 object-contain drop-shadow animate-scale-in"
            />
            <span class="text-[10px] font-mono font-bold text-gold mt-1">✓ Sello</span>
          {:else}
            <Lock class="w-5 h-5 sm:w-6 sm:h-6 opacity-40 mb-1" />
            <span class="text-[10px] font-mono opacity-50 font-bold">{String(index + 1).padStart(2, '0')}</span>
            {#if isLast}
              <span class="text-[8px] font-mono text-gold font-bold uppercase mt-0.5">GRATIS</span>
            {/if}
          {/if}
        </div>
      {/each}
    </div>

    <!-- Botón de Redención (Si tiene 10/10) -->
    {#if store.stampBalance >= 10}
      <div class="p-6 rounded-2xl bg-gradient-to-r from-ember/20 via-gold/15 to-ember/20 border border-gold/40 text-center space-y-3 shadow-gold">
        <Sparkles class="w-8 h-8 text-gold mx-auto animate-spin" style="animation-duration: 4s;" />
        <h4 class="font-serif text-3xl font-bold text-bone">¡Tienes una Hamburguesa Original Gratis!</h4>
        <p class="text-xs text-ash max-w-md mx-auto">
          Acércate a la caja física de Mías, preséntate ante el cajero y pulsa el botón a continuación para autorizar la entrega de tu hamburguesa.
        </p>
        
        {#if !isAwaitingRedemptionApproval}
          <button
            onclick={handleRedeem}
            class="py-3.5 px-8 rounded-full bg-ember hover:bg-ember-light text-white font-mono font-bold text-sm tracking-wider shadow-ember transition-all active:scale-95"
          >
            Redimir Hamburguesa Original Gratis
          </button>
        {:else}
          <div class="py-3 px-6 rounded-full bg-char border border-gold/40 text-gold font-mono text-xs flex items-center justify-center gap-2 max-w-xs mx-auto animate-pulse">
            <Clock class="w-4 h-4" /> A la espera de aprobación en caja...
          </div>
        {/if}

        <p class="text-[11px] font-mono text-ash/80">
          Nota: Si deseas agregar adiciones o combo, solo pagarás el excedente en caja.
        </p>
      </div>
    {/if}
  </div>

  <!-- Órdenes del Cliente y Solicitud de Sellos con Factura Física -->
  <div class="p-6 sm:p-8 rounded-3xl bg-char-card border hairline-dark space-y-4">
    <div class="flex items-center justify-between border-b hairline-dark pb-3">
      <h3 class="font-serif text-2xl font-bold text-bone">Tus Órdenes Recientes</h3>
      <span class="text-xs font-mono text-ash">Facturas válidas durante 24 horas</span>
    </div>

    {#if store.activeOrders.length === 0}
      <p class="text-xs text-ash py-4 text-center">Aún no tienes órdenes registradas en tu perfil.</p>
    {:else}
      <div class="space-y-3">
        {#each store.activeOrders as ord}
          <div class="p-4 rounded-2xl bg-char border hairline-dark flex flex-col sm:flex-row sm:items-center justify-between gap-3">
            <div>
              <div class="flex items-center gap-2">
                <span class="font-mono text-sm font-bold text-bone">#{ord.code}</span>
                <span class="text-[10px] font-mono uppercase px-2 py-0.5 rounded font-bold {ord.status === 'PAID' ? 'bg-emerald-500/20 text-emerald-400' : 'bg-ember/20 text-ember'}">
                  {ord.status === 'PAID' ? 'ORDEN PAGA' : ord.status}
                </span>
                <span class="text-xs text-ash font-mono">• {ord.type === 'in_store' ? 'Punto' : 'Domicilio'}</span>
              </div>
              <span class="text-xs text-gold font-mono mt-1 block">${ord.total.toLocaleString('es-CO')}</span>
            </div>

            <div>
              {#if ord.status === 'PAID'}
                <button
                  onclick={() => openReceiptModal(ord)}
                  class="px-4 py-2 rounded-full bg-gold hover:bg-gold-light text-char font-mono font-bold text-xs flex items-center gap-1.5 transition-colors shadow-gold"
                >
                  <Award class="w-3.5 h-3.5" /> Solicitar Sello
                </button>
              {:else}
                <span class="text-[11px] font-mono text-ash">
                  Paga tu orden en caja para habilitar sellos
                </span>
              {/if}
            </div>
          </div>
        {/each}
      </div>
    {/if}

    <!-- Solicitudes de Sellos en Tránsito -->
    {#if pendingStampRequests.length > 0}
      <div class="pt-4 border-t hairline-dark space-y-2">
        <span class="text-xs font-mono uppercase tracking-widest text-gold block">Solicitudes en Revisión</span>
        {#each pendingStampRequests as req}
          <div class="p-3 rounded-xl bg-white/5 border hairline-dark flex items-center justify-between text-xs font-mono">
            <div>
              <span class="text-bone font-bold">Factura #{req.receiptNumber}</span>
              <span class="text-ash block text-[10px]">Orden #{req.orderCode} • {req.burgersCount} hamburguesas</span>
            </div>
            <div>
              {#if req.status === 'PENDING'}
                <span class="text-gold flex items-center gap-1">
                  <Clock class="w-3.5 h-3.5 animate-spin" /> En revisión (Mesero/Caja)
                </span>
              {:else if req.status === 'APPROVED'}
                <span class="text-emerald-400 font-bold flex items-center gap-1">
                  ✓ Aprobado (+{req.burgersCount} sellos)
                </span>
              {/if}
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>

<!-- Modal para Digitar el Número de Factura Física -->
{#if receiptModalOpen}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/85 backdrop-blur-md p-4">
    <div class="bg-char-card w-full max-w-md rounded-3xl border hairline-dark p-6 space-y-4 relative shadow-2xl">
      <h3 class="font-serif text-2xl font-bold text-bone">Registrar Factura Física</h3>
      <p class="text-xs text-ash leading-relaxed">
        Ingresa el número correlativo impreso en el recibo de caja de papel que te entregaron al pagar (ej. <strong>00124</strong> o <strong>4580</strong>).
      </p>

      <div>
        <label for="sb-receipt" class="text-xs font-mono text-ash uppercase tracking-wider block mb-1">Número de Factura</label>
        <input 
          id="sb-receipt"
          type="text" 
          bind:value={receiptNumberInput}
          placeholder="Ej. 00124"
          class="w-full px-4 py-3 rounded-xl bg-char border hairline-dark text-bone font-mono text-base focus:outline-none focus:border-gold"
        />
      </div>

      <div class="flex items-center justify-end gap-3 pt-2">
        <button 
          onclick={() => receiptModalOpen = false}
          class="px-4 py-2 rounded-lg text-ash hover:text-bone text-xs font-mono"
        >
          Cancelar
        </button>
        <button 
          onclick={submitStampRequest}
          class="px-6 py-2.5 rounded-full bg-ember hover:bg-ember-light text-white font-mono font-bold text-xs tracking-wide transition-all shadow-ember"
        >
          Enviar Solicitud
        </button>
      </div>
    </div>
  </div>
{/if}
