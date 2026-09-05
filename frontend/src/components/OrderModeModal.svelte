<script>
  import { store } from '../lib/store.svelte.js';
  import { UtensilsCrossed, Bike, X } from 'lucide-svelte';

  let createdInStoreOrder = $state(null);

  function handleInStore() {
    // Generar código de comanda en el punto ORD-XXXX
    const code = 'ORD-' + Math.floor(1000 + Math.random() * 9000);
    const order = {
      id: 'ord-' + Date.now(),
      code,
      type: 'in_store',
      status: 'PENDING_PAYMENT',
      total: store.cartSubtotal,
      items: [...store.cart],
      created_at: new Date().toISOString()
    };

    store.addOrder(order);
    createdInStoreOrder = order;
    store.clearCart();
  }

  function handleDelivery() {
    store.isOrderModeModalOpen = false;
    store.isDeliveryWizardOpen = true; // Abre wizard conversacional
  }

  function closeAll() {
    store.isOrderModeModalOpen = false;
    createdInStoreOrder = null;
  }
</script>

{#if store.isOrderModeModalOpen}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div 
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/85 backdrop-blur-md p-4"
    onclick={closeAll}
  >
    <div 
      class="bg-char-card w-full max-w-md rounded-3xl border hairline-dark overflow-hidden shadow-2xl p-6 relative"
      onclick={(e) => e.stopPropagation()}
    >
      <button 
        onclick={closeAll}
        class="absolute top-4 right-4 p-2 rounded-full hover:bg-white/5 text-ash hover:text-bone"
      >
        <X class="w-5 h-5" />
      </button>

      {#if !createdInStoreOrder}
        <div class="text-center mb-6 pt-2">
          <span class="text-xs font-mono uppercase tracking-widest text-gold mb-1 block">Modalidad de Comanda</span>
          <h3 class="font-serif text-3xl font-bold text-bone">¿Dónde vas a disfrutar?</h3>
          <p class="text-xs text-ash mt-1">Selecciona cómo deseas recibir tu orden</p>
        </div>

        <div class="space-y-4">
          <!-- Opción 1: Consumo en el Punto -->
          <div 
            onclick={handleInStore}
            class="p-5 rounded-2xl bg-white/5 border hairline-dark hover:border-gold/50 transition-all cursor-pointer flex items-center gap-4 group hover:shadow-lux"
          >
            <div class="w-12 h-12 rounded-xl bg-gold/15 text-gold flex items-center justify-center border border-gold/30 group-hover:scale-110 transition-transform">
              <UtensilsCrossed class="w-6 h-6" />
            </div>
            <div>
              <h4 class="font-serif text-xl font-bold text-bone group-hover:text-gold transition-colors">Consumo en el Punto</h4>
              <p class="text-xs text-ash mt-0.5">Pagas en caja, recibes tu identificador físico y te llevamos la orden a la mesa.</p>
            </div>
          </div>

          <!-- Opción 2: Domicilio -->
          <div 
            onclick={handleDelivery}
            class="p-5 rounded-2xl bg-white/5 border hairline-dark hover:border-ember/50 transition-all cursor-pointer flex items-center gap-4 group hover:shadow-ember"
          >
            <div class="w-12 h-12 rounded-xl bg-ember/15 text-ember flex items-center justify-center border border-ember/30 group-hover:scale-110 transition-transform">
              <Bike class="w-6 h-6" />
            </div>
            <div>
              <div class="flex items-center gap-2">
                <h4 class="font-serif text-xl font-bold text-bone group-hover:text-ember transition-colors">A Domicilio</h4>
                <span class="text-[10px] font-mono text-gold bg-gold/10 px-2 py-0.5 rounded">+$6.000</span>
              </div>
              <p class="text-xs text-ash mt-0.5">Coordinado a tu puerta y despachado por WhatsApp al restaurante.</p>
            </div>
          </div>
        </div>
      {:else}
        <!-- Comprobante Suizo para Punto (Inspirado en Ticket TF2026 - Ref 1) -->
        <div class="text-center py-2">
          <div class="inline-block p-2 rounded-full bg-ember/10 border border-ember/30 text-ember mb-3">
            <UtensilsCrossed class="w-6 h-6 mx-auto" />
          </div>

          <span class="text-xs font-mono uppercase tracking-widest text-ember block mb-1">¡Comanda Generada!</span>
          <h3 class="font-serif text-3xl font-bold text-bone mb-4">#{createdInStoreOrder.code}</h3>

          <div class="p-4 rounded-2xl bg-char border hairline-dark text-left space-y-2 text-xs font-mono mb-6">
            <div class="flex justify-between text-ash">
              <span>Estado:</span>
              <span class="text-ember font-bold">POR COBRAR EN CAJA</span>
            </div>
            <div class="flex justify-between text-ash">
              <span>Total a pagar:</span>
              <span class="text-gold font-bold text-sm">${createdInStoreOrder.total.toLocaleString('es-CO')}</span>
            </div>
            <div class="border-t hairline-dark pt-2 mt-2 text-ash text-[11px] leading-relaxed">
              <p class="text-bone font-medium mb-1">Pasos siguientes:</p>
              <p>1. Acércate a la caja física con este código <strong>#{createdInStoreOrder.code}</strong>.</p>
              <p>2. Realiza tu pago y recibe tu <strong>factura de papel</strong> y tu <strong>localizador de mesa</strong>.</p>
              <p>3. Sube a las mesas; una vez tu factura esté paga podrás solicitar tus sellos en tu perfil.</p>
            </div>
          </div>

          <button
            onclick={closeAll}
            class="w-full py-3.5 px-6 rounded-full bg-gold hover:bg-gold-light text-char font-mono font-bold text-xs uppercase tracking-wider transition-colors"
          >
            Entendido, ir a esperar mi orden
          </button>
        </div>
      {/if}
    </div>
  </div>
{/if}
