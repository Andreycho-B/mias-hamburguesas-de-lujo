<script>
  import { store } from '../lib/store.svelte.js';
  import { staffLoginAPI } from '../lib/api.js';
  import { X, Shield, User, Lock, Mail, Loader2, AlertCircle } from 'lucide-svelte';

  let mode = $state('customer'); // 'customer', 'staff'
  let email = $state('cajero@mias.com');
  let password = $state('CajeroMias2026!');
  let staffRoleSelect = $state('cashier'); // 'waiter', 'cashier', 'admin'
  let isLoading = $state(false);
  let errorMessage = $state('');

  function selectPreset(role) {
    staffRoleSelect = role;
    if (role === 'admin') {
      email = 'admin@mias.com';
      password = 'AdminMias2026!';
    } else if (role === 'cashier') {
      email = 'cajero@mias.com';
      password = 'CajeroMias2026!';
    } else if (role === 'waiter') {
      email = 'mesero@mias.com';
      password = 'MeseroMias2026!';
    }
  }

  function handleGoogleLogin() {
    store.setUser({
      id: 'usr-' + Math.random().toString(36).substring(2, 9),
      full_name: 'Cliente Mías',
      email: 'cliente@gmail.com',
      role: 'customer'
    });
    store.isAuthModalOpen = false;
  }

  async function handleStaffLogin() {
    if (!email.trim() || !password.trim()) {
      errorMessage = 'Por favor ingresa usuario y contraseña';
      return;
    }

    isLoading = true;
    errorMessage = '';

    try {
      const data = await staffLoginAPI(email.trim(), password.trim());
      if (data && data.tokens) {
        localStorage.setItem('mias_access_token', data.tokens.access_token);
        localStorage.setItem('mias_refresh_token', data.tokens.refresh_token);
        store.setUser(data.user);
        store.staffRole = data.user.role;
      } else {
        // Fallback local
        store.setUser({
          id: 'staff-' + staffRoleSelect,
          full_name: staffRoleSelect === 'admin' ? 'Administrador General' : staffRoleSelect === 'cashier' ? 'Cajero Principal' : 'Mesero Sala',
          email: email.trim(),
          role: staffRoleSelect
        });
        store.staffRole = staffRoleSelect;
      }
      store.currentView = 'staff';
      store.isAuthModalOpen = false;
    } catch (err) {
      errorMessage = err.message || 'Error al autenticar con el servidor';
    } finally {
      isLoading = false;
    }
  }
</script>

{#if store.isAuthModalOpen}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div 
    class="fixed inset-0 z-50 flex items-center justify-center bg-black/85 backdrop-blur-md p-4"
    onclick={() => store.isAuthModalOpen = false}
  >
    <div 
      class="bg-char-card w-full max-w-md rounded-3xl border hairline-dark overflow-hidden shadow-2xl p-6 sm:p-8 relative"
      onclick={(e) => e.stopPropagation()}
    >
      <button 
        onclick={() => store.isAuthModalOpen = false}
        class="absolute top-5 right-5 p-2 rounded-full hover:bg-white/5 text-ash hover:text-bone"
      >
        <X class="w-5 h-5" />
      </button>

      <div class="text-center mb-6">
        <img 
          src="/imagenes/marca/isotipo.webp" 
          alt="Mías" 
          class="w-12 h-12 mx-auto mb-2"
        />
        <h3 class="font-serif text-3xl font-bold text-bone">Iniciar Sesión</h3>
        <p class="text-xs text-ash mt-1">Accede a tu perfil o terminal operativa</p>
      </div>

      <!-- Selector Modo Cliente vs Personal -->
      <div class="grid grid-cols-2 gap-2 p-1 rounded-xl bg-char border hairline-dark mb-6 text-xs font-mono">
        <button
          onclick={() => mode = 'customer'}
          class="py-2 rounded-lg font-bold transition-all {mode === 'customer' ? 'bg-gold text-char shadow' : 'text-ash hover:text-bone'}"
        >
          Cliente
        </button>
        <button
          onclick={() => mode = 'staff'}
          class="py-2 rounded-lg font-bold transition-all {mode === 'staff' ? 'bg-gold text-char shadow' : 'text-ash hover:text-bone'}"
        >
          Personal / Caja
        </button>
      </div>

      {#if mode === 'customer'}
        <div class="space-y-4 text-center">
          <p class="text-xs text-ash leading-relaxed">
            Inicia sesión con Google para acumular sellos, consultar tus comandas y redimir tu hamburguesa gratis.
          </p>

          <button
            onclick={handleGoogleLogin}
            class="w-full py-3.5 px-6 rounded-full bg-white hover:bg-cream text-char font-mono font-bold text-xs flex items-center justify-center gap-3 transition-colors shadow-lg"
          >
            <svg class="w-4 h-4" viewBox="0 0 24 24">
              <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
              <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
              <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.06H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.94l2.85-2.22.81-.63z"/>
              <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06l3.66 2.84c.87-2.6 3.3-4.52 6.16-4.52z"/>
            </svg>
            Continuar con Google
          </button>
        </div>
      {:else}
        <!-- Preajustes Rápidos de Demostración y Auditoría -->
        <div class="mb-4">
          <span class="text-[10px] font-mono uppercase tracking-widest text-ash block mb-1.5">Acceso Rápido Auditores / Staff:</span>
          <div class="grid grid-cols-3 gap-1.5">
            <button
              type="button"
              onclick={() => selectPreset('cashier')}
              class="py-1 px-2 rounded-lg text-[10px] font-mono border transition-colors {staffRoleSelect === 'cashier' ? 'bg-gold/15 border-gold text-gold font-bold' : 'bg-char border-white/5 text-ash hover:text-bone'}"
            >
              Caja
            </button>
            <button
              type="button"
              onclick={() => selectPreset('waiter')}
              class="py-1 px-2 rounded-lg text-[10px] font-mono border transition-colors {staffRoleSelect === 'waiter' ? 'bg-gold/15 border-gold text-gold font-bold' : 'bg-char border-white/5 text-ash hover:text-bone'}"
            >
              Mesero
            </button>
            <button
              type="button"
              onclick={() => selectPreset('admin')}
              class="py-1 px-2 rounded-lg text-[10px] font-mono border transition-colors {staffRoleSelect === 'admin' ? 'bg-gold/15 border-gold text-gold font-bold' : 'bg-char border-white/5 text-ash hover:text-bone'}"
            >
              Admin
            </button>
          </div>
        </div>

        {#if errorMessage}
          <div class="p-3 rounded-xl bg-ember/15 border border-ember/30 text-ember-light text-xs flex items-center gap-2 mb-4 font-mono">
            <AlertCircle class="w-4 h-4 shrink-0" />
            <span>{errorMessage}</span>
          </div>
        {/if}

        <form onsubmit={(e) => { e.preventDefault(); handleStaffLogin(); }} class="space-y-4">
          <div>
            <label for="am-role" class="text-xs font-mono uppercase tracking-wider text-ash block mb-1">Puesto de Operación</label>
            <select
              id="am-role"
              bind:value={staffRoleSelect}
              onchange={(e) => selectPreset(e.target.value)}
              class="w-full px-4 py-2.5 rounded-xl bg-char border hairline-dark text-bone font-mono text-xs focus:outline-none focus:border-gold"
            >
              <option value="waiter">Mesero (Sala / Sellos)</option>
              <option value="cashier">Cajero (Caja Central)</option>
              <option value="admin">Administrador (Control Total BI)</option>
            </select>
          </div>

          <div>
            <label for="am-email" class="text-xs font-mono uppercase tracking-wider text-ash block mb-1">Correo Electrónico</label>
            <input 
              id="am-email"
              type="email"
              bind:value={email}
              placeholder="personal@mias.com"
              class="w-full px-4 py-2.5 rounded-xl bg-char border hairline-dark text-bone text-xs focus:outline-none focus:border-gold"
              required
            />
          </div>

          <div>
            <label for="am-pass" class="text-xs font-mono uppercase tracking-wider text-ash block mb-1">Contraseña</label>
            <input 
              id="am-pass"
              type="password"
              bind:value={password}
              placeholder="••••••••"
              class="w-full px-4 py-2.5 rounded-xl bg-char border hairline-dark text-bone text-xs focus:outline-none focus:border-gold"
              required
            />
          </div>

          <button
            type="submit"
            disabled={isLoading}
            class="w-full py-3 px-6 rounded-full bg-gold hover:bg-gold-light text-char font-mono font-bold text-xs uppercase tracking-wider transition-colors shadow-gold mt-2 flex items-center justify-center gap-2 disabled:opacity-50"
          >
            {#if isLoading}
              <Loader2 class="w-4 h-4 animate-spin" />
              <span>Verificando Credenciales...</span>
            {:else}
              <span>Acceder a la Terminal</span>
            {/if}
          </button>
        </form>
      {/if}
    </div>
  </div>
{/if}
