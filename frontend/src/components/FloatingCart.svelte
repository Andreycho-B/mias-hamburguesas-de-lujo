<script>
  import { store } from '../lib/store.svelte.js';
  import { ShoppingBag, X, Plus, Minus, Trash2, ShieldCheck } from 'lucide-svelte';

  let showHabeasDataCard = $state(false);

  function formatPrice(val) {
    return new Intl.NumberFormat('es-CO', { style: 'currency', currency: 'COP', maximumFractionDigits: 0 }).format(val);
  }

  function handleContinue() {
    if (!store.user) {
      showHabeasDataCard = true;
    } else {
      store.isCartOpen = false;
      store.isOrderModeModalOpen = true; // Abre selector punto vs domicilio
    }
  }

  function acceptHabeasAndLogin() {
    // Simular login con Google rápido de 1 clic (OAuth)
    const mockUser = {
      id: 'usr-' + Math.random().toString(36).substring(2, 9),
      full_name: 'Cliente Mías',
      email: 'cliente@gmail.com',
      role: 'customer'
    };
    store.setUser(mockUser);
    showHabeasDataCard = false;
    store.isCartOpen = false;
    store.isOrderModeModalOpen = true;
  }
</script>

<!-- Botón Circular Flotante (56px) -->
{#if store.cartCount > 0 && !store.isCartOpen}
  <div class="fixed bottom-6 right-6 z-40 pb-safe">
    <button
      onclick={() => store.isCartOpen = true}
      class="relative w-14 h-14 rounded-full bg-char-card hover:bg-char border-2 border-gold text-gold flex items-center justify-center shadow-gold transition-all duration-300 hover:scale-105 active:scale-95"
      aria-label="Ver carrito"
    >
      <ShoppingBag class="w-6 h-6 text-gold" />
      <span class="absolute -top-1.5 -right-1.5 bg-ember text-white text-xs font-mono font-bold w-6 h-6 rounded-full flex items-center justify-center border-2 border-char shadow-md animate-bounce">
        {store.cartCount}
      </span>
    </button>
  </div>
{/if}

<!-- Drawer Inferior del Carrito -->
{#if store.isCartOpen}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div 
    class="fixed inset-0 z-50 flex items-end justify-center bg-black/80 backdrop-blur-sm"
    onclick={() => store.isCartOpen = false}
  >
    <div 
      class="bg-char-card w-full max-w-lg rounded-t-3xl border-t hairline-dark overflow-hidden flex flex-col max-h-[85vh] shadow-2xl pb-safe"
      onclick={(e) => e.stopPropagation()}
    >
      <!-- Cabecera del Carrito -->
      <div class="p-5 border-b hairline-dark flex items-center justify-between">
        <div class="flex items-center gap-2">
          <ShoppingBag class="w-5 h-5 text-gold" />
          <h3 class="font-serif text-2xl font-bold text-bone">Tu Comanda</h3>
          <span class="text-xs font-mono text-ash">({store.cartCount} items)</span>
        </div>
        <button 
          onclick={() => store.isCartOpen = false}
          class="p-2 rounded-full hover:bg-white/5 text-ash hover:text-bone transition-colors"
        >
          <X class="w-5 h-5" />
        </button>
      </div>

      <!-- Lista de Items -->
      <div class="p-5 overflow-y-auto space-y-4 flex-1">
        {#if store.cart.length === 0}
          <div class="py-12 text-center text-ash">
            <ShoppingBag class="w-12 h-12 mx-auto mb-3 opacity-30" />
            <p class="font-serif text-lg text-bone/60">Tu comanda está vacía</p>
            <p class="text-xs mt-1">Explora la carta y agrega tus favoritas al carbón</p>
          </div>
        {:else}
          {#each store.cart as item (item.id)}
            <div class="flex items-center justify-between gap-3 p-3 rounded-xl bg-white/5 border hairline-dark">
              <div class="flex-1">
                <h4 class="font-serif text-lg font-bold text-bone">{item.product.name}</h4>
                {#if item.combo}
                  <span class="text-[11px] font-mono text-gold block">+ Combo Papas & Gaseosa</span>
                {/if}
                {#if item.extras && item.extras.length > 0}
                  {#each item.extras as ex}
                    <span class="text-[10px] font-mono text-ash block">+ {ex.name}</span>
                  {/each}
                {/if}
                <span class="font-mono text-xs font-bold text-gold mt-1 block">
                  {formatPrice(item.totalPrice)}
                </span>
              </div>

              <!-- Controles de Cantidad (Touch Target 44px) -->
              <div class="flex items-center gap-2 bg-char px-2 py-1 rounded-full border hairline-dark">
                <button 
                  onclick={() => store.updateQuantity(item.id, -1)}
                  class="w-8 h-8 rounded-full flex items-center justify-center hover:text-ember text-ash transition-colors"
                  aria-label="Restar uno"
                >
                  {#if item.quantity === 1}
                    <Trash2 class="w-3.5 h-3.5 text-ember" />
                  {:else}
                    <Minus class="w-3.5 h-3.5" />
                  {/if}
                </button>
                <span class="font-mono text-xs font-bold text-bone w-4 text-center">{item.quantity}</span>
                <button 
                  onclick={() => store.updateQuantity(item.id, 1)}
                  class="w-8 h-8 rounded-full flex items-center justify-center hover:text-gold text-ash transition-colors"
                  aria-label="Sumar uno"
                >
                  <Plus class="w-3.5 h-3.5" />
                </button>
              </div>
            </div>
          {/each}
        {/if}
      </div>

      <!-- Tarjeta Legal Habeas Data (Banner Inferior) -->
      {#if showHabeasDataCard}
        <div class="p-4 bg-char border-t border-ember/30 mx-4 mb-3 rounded-2xl animate-fade-in shadow-ember">
          <div class="flex items-start gap-3">
            <ShieldCheck class="w-6 h-6 text-gold flex-shrink-0 mt-0.5" />
            <div class="text-xs">
              <p class="font-bold text-bone mb-1">Protección de Datos (Ley 1581)</p>
              <p class="text-ash leading-relaxed mb-3">
                Para coordinar tu orden y acumular sellos de fidelización, autorizas el tratamiento de tus datos personales conforme a nuestra política de privacidad.
              </p>
              <div class="flex items-center gap-2">
                <button 
                  onclick={acceptHabeasAndLogin}
                  class="px-4 py-2 rounded-lg bg-gold text-char font-mono font-bold text-xs hover:bg-gold-light transition-colors"
                >
                  Aceptar y continuar con Google
                </button>
                <button 
                  onclick={() => showHabeasDataCard = false}
                  class="px-3 py-2 rounded-lg text-ash hover:text-bone text-xs"
                >
                  Rechazar
                </button>
              </div>
            </div>
          </div>
        </div>
      {/if}

      <!-- Resumen y Botón de Avance -->
      {#if store.cart.length > 0}
        <div class="p-5 bg-char border-t hairline-dark space-y-4">
          <div class="flex items-center justify-between font-mono text-sm">
            <span class="text-ash uppercase tracking-wider">Subtotal comanda</span>
            <span class="text-bone font-bold text-base">{formatPrice(store.cartSubtotal)}</span>
          </div>

          <button
            onclick={handleContinue}
            class="w-full py-4 px-6 rounded-full bg-ember hover:bg-ember-light text-white font-mono font-bold tracking-wider text-sm flex items-center justify-between transition-all shadow-ember active:scale-[0.98]"
          >
            <span>Continuar con la orden</span>
            <span>{formatPrice(store.cartSubtotal)} →</span>
          </button>
        </div>
      {/if}
    </div>
  </div>
{/if}
