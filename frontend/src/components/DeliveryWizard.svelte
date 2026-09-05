<script>
  import { store } from '../lib/store.svelte.js';
  import { X, ArrowRight, ArrowLeft, Send, CheckCircle2 } from 'lucide-svelte';

  let currentStep = $state(1); // 1: Datos personales, 2: Dirección, 3: Pago, 4: Resumen
  let name = $state(store.user?.full_name || '');
  let phone = $state(store.user?.phone || '');
  let address = $state('');
  let neighborhood = $state('');
  let notes = $state('');
  let paymentMethod = $state('Efectivo'); // 'Efectivo', 'Nequi / Bre-B'
  let cashAmount = $state('');
  let deliveryFee = 6000;

  let totalAmount = $derived(store.cartSubtotal + deliveryFee);
  let calculatedChange = $derived(() => {
    const val = parseInt(cashAmount, 10);
    if (!isNaN(val) && val >= totalAmount) {
      return val - totalAmount;
    }
    return 0;
  });

  function handleNext() {
    if (currentStep === 1 && (!name.trim() || !phone.trim())) {
      alert('Por favor ingresa tu nombre y número de celular');
      return;
    }
    if (currentStep === 2 && (!address.trim() || !neighborhood.trim())) {
      alert('Por favor ingresa la dirección y el barrio');
      return;
    }
    currentStep += 1;
  }

  function handleBack() {
    if (currentStep > 1) {
      currentStep -= 1;
    }
  }

  function confirmAndDispatchWhatsApp() {
    const code = 'DOM-' + Math.floor(1000 + Math.random() * 9000);
    const order = {
      id: 'dom-' + Date.now(),
      code,
      type: 'delivery',
      status: 'IN_PREPARATION',
      subtotal: store.cartSubtotal,
      delivery_fee: deliveryFee,
      total: totalAmount,
      delivery_name: name,
      delivery_phone: phone,
      delivery_address: `${address}, Barrio ${neighborhood}`,
      delivery_notes: notes,
      payment_method: paymentMethod,
      cash_change_amount: parseInt(cashAmount, 10) || 0,
      items: [...store.cart],
      created_at: new Date().toISOString()
    };

    store.addOrder(order);

    // Formatear Mensaje Oficial para WhatsApp al 321 252 0881
    let msg = `🍔 *NUEVO PEDIDO A DOMICILIO — MÍAS HAMBURGUESAS* 🍔\n` +
      `*Orden:* #${code}\n` +
      `*Cliente:* ${name} (${phone})\n` +
      `*Dirección:* ${address}, Barrio ${neighborhood}\n`;

    if (notes.trim()) {
      msg += `*Indicaciones:* ${notes}\n`;
    }

    msg += `\n*Detalle del Pedido:*\n`;
    store.cart.forEach((it) => {
      msg += `• ${it.quantity}x ${it.product.name} ($${it.totalPrice.toLocaleString('es-CO')})\n`;
      if (it.combo) msg += `  - Con Combo Papa + Gaseosa\n`;
      if (it.extras && it.extras.length > 0) {
        it.extras.forEach((ex) => {
          msg += `  - Extra: ${ex.name}\n`;
        });
      }
    });

    msg += `\n*Subtotal:* $${store.cartSubtotal.toLocaleString('es-CO')}\n` +
      `*Domicilio:* $${deliveryFee.toLocaleString('es-CO')}\n` +
      `*Total a Pagar:* $${totalAmount.toLocaleString('es-CO')}\n` +
      `*Método de Pago:* ${paymentMethod}\n`;

    if (paymentMethod === 'Efectivo' && parseInt(cashAmount, 10) > 0) {
      msg += `*Paga con:* $${parseInt(cashAmount, 10).toLocaleString('es-CO')} (Cambio: $${calculatedChange().toLocaleString('es-CO')})\n`;
    }

    store.clearCart();
    store.isDeliveryWizardOpen = false;

    // Abrir WhatsApp Web / App
    const encoded = encodeURIComponent(msg);
    const waURL = `https://api.whatsapp.com/send?phone=573212520881&text=${encoded}`;
    window.open(waURL, '_blank');
  }
</script>

{#if store.isDeliveryWizardOpen}
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/90 backdrop-blur-md p-4">
    <div class="bg-char-card w-full max-w-lg rounded-3xl border hairline-dark overflow-hidden shadow-2xl p-6 sm:p-8 relative">
      <!-- Cierre -->
      <button 
        onclick={() => store.isDeliveryWizardOpen = false}
        class="absolute top-5 right-5 p-2 rounded-full hover:bg-white/5 text-ash hover:text-bone"
      >
        <X class="w-5 h-5" />
      </button>

      <!-- Barra de Progreso Superior -->
      <div class="mb-6">
        <div class="flex items-center justify-between text-xs font-mono text-ash mb-2">
          <span>Paso {currentStep} de 4</span>
          <span class="text-gold font-bold">{currentStep === 1 ? 'Contacto' : currentStep === 2 ? 'Dirección' : currentStep === 3 ? 'Pago' : 'Confirmación'}</span>
        </div>
        <div class="w-full h-1 bg-white/10 rounded-full overflow-hidden">
          <div 
            class="h-full bg-ember transition-all duration-300"
            style="width: {(currentStep / 4) * 100}%"
          ></div>
        </div>
      </div>

      <!-- PASO 1: Datos de Contacto -->
      {#if currentStep === 1}
        <div class="space-y-4">
          <div>
            <h3 class="font-serif text-3xl font-bold text-bone mb-1">¿A nombre de quién?</h3>
            <p class="text-xs text-ash">Ingresa tus datos para coordinar la entrega y llamarte al llegar.</p>
          </div>

          <div class="space-y-3">
            <div>
              <label for="dw-name" class="text-xs font-mono uppercase tracking-wider text-ash block mb-1">Nombre Completo</label>
              <input 
                id="dw-name"
                type="text" 
                bind:value={name}
                placeholder="Ej. Carlos Mendoza" 
                class="w-full px-4 py-3 rounded-xl bg-char border hairline-dark text-bone focus:outline-none focus:border-gold text-sm"
              />
            </div>
            <div>
              <label for="dw-phone" class="text-xs font-mono uppercase tracking-wider text-ash block mb-1">Celular de Contacto</label>
              <input 
                id="dw-phone"
                type="tel" 
                bind:value={phone}
                placeholder="Ej. 300 123 4567" 
                class="w-full px-4 py-3 rounded-xl bg-char border hairline-dark text-bone focus:outline-none focus:border-gold text-sm"
              />
            </div>
          </div>
        </div>

      <!-- PASO 2: Dirección y Barrio -->
      {:else if currentStep === 2}
        <div class="space-y-4">
          <div>
            <h3 class="font-serif text-3xl font-bold text-bone mb-1">¿A dónde lo llevamos?</h3>
            <p class="text-xs text-ash">Indícanos la dirección y detalles para el domiciliario.</p>
          </div>

          <div class="space-y-3">
            <div>
              <label for="dw-address" class="text-xs font-mono uppercase tracking-wider text-ash block mb-1">Dirección Exacta</label>
              <input 
                id="dw-address"
                type="text" 
                bind:value={address}
                placeholder="Ej. Calle 45 # 12-34" 
                class="w-full px-4 py-3 rounded-xl bg-char border hairline-dark text-bone focus:outline-none focus:border-gold text-sm"
              />
            </div>
            <div>
              <label for="dw-neighborhood" class="text-xs font-mono uppercase tracking-wider text-ash block mb-1">Barrio</label>
              <input 
                id="dw-neighborhood"
                type="text" 
                bind:value={neighborhood}
                placeholder="Ej. El Poblado / Cañaveral" 
                class="w-full px-4 py-3 rounded-xl bg-char border hairline-dark text-bone focus:outline-none focus:border-gold text-sm"
              />
            </div>
            <div>
              <label for="dw-notes" class="text-xs font-mono uppercase tracking-wider text-ash block mb-1">Indicaciones (Torre, Apto, timbre)</label>
              <input 
                id="dw-notes"
                type="text" 
                bind:value={notes}
                placeholder="Ej. Torre 2 Apto 402, dejar en portería" 
                class="w-full px-4 py-3 rounded-xl bg-char border hairline-dark text-bone focus:outline-none focus:border-gold text-sm"
              />
            </div>
          </div>
        </div>

      <!-- PASO 3: Método de Pago -->
      {:else if currentStep === 3}
        <div class="space-y-4">
          <div>
            <h3 class="font-serif text-3xl font-bold text-bone mb-1">¿Cómo deseas pagar?</h3>
            <p class="text-xs text-ash">El pago se coordina directamente con el domiciliario al llegar.</p>
          </div>

          <div class="space-y-3">
            <!-- Selector Efectivo vs Transferencia -->
            <div class="grid grid-cols-2 gap-3">
              <button
                onclick={() => paymentMethod = 'Efectivo'}
                class="p-4 rounded-xl border text-left transition-all {paymentMethod === 'Efectivo' ? 'bg-gold/15 border-gold text-bone' : 'bg-char border-white/10 text-ash hover:text-bone'}"
              >
                <span class="font-serif text-lg font-bold block">💵 Efectivo</span>
                <span class="text-[11px] opacity-70">Pago contra entrega</span>
              </button>

              <button
                onclick={() => paymentMethod = 'Transferencia Nequi / Bre-B'}
                class="p-4 rounded-xl border text-left transition-all {paymentMethod.includes('Nequi') ? 'bg-gold/15 border-gold text-bone' : 'bg-char border-white/10 text-ash hover:text-bone'}"
              >
                <span class="font-serif text-lg font-bold block">📱 Nequi / Bre-B</span>
                <span class="text-[11px] opacity-70">Transferencia digital</span>
              </button>
            </div>

            {#if paymentMethod === 'Efectivo'}
              <div class="p-4 rounded-2xl bg-char border hairline-dark space-y-2 mt-3">
                <label for="dw-cash" class="text-xs font-mono text-ash block">¿Con qué billete vas a pagar? (Para llevarte cambio exacto)</label>
                <input 
                  id="dw-cash"
                  type="number" 
                  bind:value={cashAmount}
                  placeholder="Ej. 50000" 
                  class="w-full px-4 py-2.5 rounded-xl bg-char-card border hairline-dark text-gold font-mono font-bold text-base focus:outline-none focus:border-gold"
                />
                {#if calculatedChange() > 0}
                  <p class="text-xs font-mono text-emerald-400">
                    ✓ Tu cambio será de: <strong>${calculatedChange().toLocaleString('es-CO')}</strong>
                  </p>
                {/if}
              </div>
            {/if}
          </div>
        </div>

      <!-- PASO 4: Resumen y Despacho -->
      {:else if currentStep === 4}
        <div class="space-y-4">
          <div>
            <span class="text-xs font-mono uppercase tracking-widest text-emerald-400 block mb-1 flex items-center gap-1.5">
              <CheckCircle2 class="w-4 h-4" /> Todo listo para enviar
            </span>
            <h3 class="font-serif text-3xl font-bold text-bone">Resumen de Entrega</h3>
          </div>

          <div class="p-4 rounded-2xl bg-char border hairline-dark text-xs font-mono space-y-2">
            <div class="flex justify-between text-ash">
              <span>Cliente:</span>
              <span class="text-bone">{name} ({phone})</span>
            </div>
            <div class="flex justify-between text-ash">
              <span>Entrega en:</span>
              <span class="text-bone">{address}, {neighborhood}</span>
            </div>
            <div class="flex justify-between text-ash">
              <span>Método de Pago:</span>
              <span class="text-gold">{paymentMethod}</span>
            </div>
            <div class="border-t hairline-dark pt-2 mt-2 flex justify-between text-ash">
              <span>Subtotal comanda:</span>
              <span class="text-bone">${store.cartSubtotal.toLocaleString('es-CO')}</span>
            </div>
            <div class="flex justify-between text-ash">
              <span>Costo Domicilio:</span>
              <span class="text-bone">${deliveryFee.toLocaleString('es-CO')}</span>
            </div>
            <div class="border-t hairline-dark pt-2 flex justify-between text-base font-bold">
              <span class="text-bone">Total a Cancelar:</span>
              <span class="text-gold">${totalAmount.toLocaleString('es-CO')}</span>
            </div>
          </div>
        </div>
      {/if}

      <!-- Botones de Navegación Inferior -->
      <div class="flex items-center justify-between gap-3 mt-8 pt-4 border-t hairline-dark">
        {#if currentStep > 1}
          <button
            onclick={handleBack}
            class="px-4 py-3 rounded-full border hairline-dark hover:bg-white/5 text-ash hover:text-bone text-xs font-mono flex items-center gap-1.5"
          >
            <ArrowLeft class="w-4 h-4" /> Atrás
          </button>
        {:else}
          <div></div>
        {/if}

        {#if currentStep < 4}
          <button
            onclick={handleNext}
            class="px-6 py-3 rounded-full bg-gold hover:bg-gold-light text-char font-mono font-bold text-xs flex items-center gap-2 transition-all shadow-gold"
          >
            Siguiente <ArrowRight class="w-4 h-4" />
          </button>
        {:else}
          <button
            onclick={confirmAndDispatchWhatsApp}
            class="px-6 py-3.5 rounded-full bg-emerald-600 hover:bg-emerald-500 text-white font-mono font-bold text-xs flex items-center gap-2 transition-all shadow-lg active:scale-95"
          >
            <Send class="w-4 h-4" /> Confirmar y Despachar a WhatsApp (321 252 0881)
          </button>
        {/if}
      </div>
    </div>
  </div>
{/if}
