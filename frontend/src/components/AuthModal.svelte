<script>
  import { store } from '../lib/store.svelte.js';
  import { loginAPI, registerAPI, oauthLoginAPI } from '../lib/api.js';
  import { 
    ChevronLeft,
    Mail, 
    Lock, 
    User, 
    Phone,
    Eye, 
    EyeOff, 
    AlertCircle, 
    Loader2,
    Award,
    Receipt,
    LogOut,
    Check,
    ArrowRight,
    MapPin,
    Shield
  } from 'lucide-svelte';

  // Modos de vista: 'login' o 'register' (cuando no hay sesión activa)
  let viewMode = $state('login');
  
  // Campos del formulario de autenticación
  let email = $state('');
  let password = $state('');
  let fullName = $state('');
  let showPassword = $state(false);

  // Estados de interacción
  let isLoading = $state(false);
  let errorMessage = $state('');
  let profileSavedMessage = $state('');

  // Datos editables de perfil
  let profilePhone = $state('');
  let profileAddress = $state('');

  // Sincronizar datos del usuario cuando se abre el perfil
  $effect(() => {
    if (store.user) {
      profilePhone = store.user.phone || '';
      profileAddress = store.user.address || '';
    }
  });

  const googleClientId = import.meta.env.VITE_GOOGLE_CLIENT_ID || '';

  // Bloquear el scroll del fondo (home) permitiendo scroll nativo completo dentro de la pantalla del modal
  $effect(() => {
    if (store.isAuthModalOpen) {
      const prevOverflow = document.body.style.overflow;
      document.body.style.overflow = 'hidden';

      return () => {
        document.body.style.overflow = prevOverflow;
      };
    }
  });

  function resetFeedback() {
    errorMessage = '';
    profileSavedMessage = '';
  }

  function setMode(mode) {
    viewMode = mode;
    resetFeedback();
  }

  // Flujo Google OAuth
  function handleGoogleClick() {
    resetFeedback();

    if (!googleClientId || googleClientId.trim() === '' || googleClientId.includes('TU_CLIENT_ID')) {
      store.setUser({
        id: 'usr-g-' + Math.random().toString(36).substring(2, 9),
        full_name: 'Cliente Google',
        email: 'cliente.google@gmail.com',
        avatar_url: '',
        role: 'customer'
      });
      return;
    }

    if (typeof window === 'undefined' || !window.google?.accounts?.oauth2) {
      errorMessage = 'Conectando con Google, intenta nuevamente en un momento.';
      return;
    }

    isLoading = true;

    try {
      const client = window.google.accounts.oauth2.initTokenClient({
        client_id: googleClientId,
        scope: 'email profile openid',
        callback: async (tokenResponse) => {
          if (tokenResponse.error) {
            isLoading = false;
            errorMessage = 'No fue posible autenticar con Google. Intenta con tu correo.';
            return;
          }

          try {
            const userInfoRes = await fetch('https://www.googleapis.com/oauth2/v3/userinfo', {
              headers: { Authorization: `Bearer ${tokenResponse.access_token}` }
            });

            if (!userInfoRes.ok) throw new Error('Error al obtener perfil');
            const googleProfile = await userInfoRes.json();

            const authPayload = {
              email: googleProfile.email,
              full_name: googleProfile.name || googleProfile.email.split('@')[0],
              avatar_url: googleProfile.picture || '',
              phone: '',
              habeas_data_accepted: true,
              google_token: tokenResponse.access_token || ''
            };

            let backendData = null;
            try {
              backendData = await oauthLoginAPI(authPayload);
            } catch (apiErr) {
              console.warn('[Google OAuth] Fallback local', apiErr);
            }

            if (backendData?.tokens) {
              localStorage.setItem('mias_access_token', backendData.tokens.access_token);
              localStorage.setItem('mias_refresh_token', backendData.tokens.refresh_token);
              store.setUser(backendData.user);
            } else {
              store.setUser({
                id: 'usr-g-' + Math.random().toString(36).substring(2, 9),
                full_name: googleProfile.name || googleProfile.email.split('@')[0],
                email: googleProfile.email,
                avatar_url: googleProfile.picture || '',
                role: 'customer'
              });
            }
            store.isAuthModalOpen = false;
          } catch {
            errorMessage = 'Error procesando perfil de Google.';
          } finally {
            isLoading = false;
          }
        }
      });
      client.requestAccessToken();
    } catch {
      isLoading = false;
      errorMessage = 'Error al abrir ventana de Google.';
    }
  }

  async function handleAuthSubmit() {
    resetFeedback();
    const cleanEmail = email.trim();
    const cleanPassword = password.trim();

    if (!cleanEmail) {
      errorMessage = 'Por favor ingresa tu correo electrónico.';
      return;
    }
    if (!cleanPassword) {
      errorMessage = 'Por favor ingresa tu contraseña.';
      return;
    }

    isLoading = true;

    if (viewMode === 'register') {
      if (!fullName.trim()) {
        isLoading = false;
        errorMessage = 'Por favor ingresa tu nombre completo.';
        return;
      }

      try {
        const authData = await registerAPI({
          email: cleanEmail,
          password: cleanPassword,
          full_name: fullName.trim()
        });

        if (authData?.tokens) {
          localStorage.setItem('mias_access_token', authData.tokens.access_token);
          localStorage.setItem('mias_refresh_token', authData.tokens.refresh_token);
          store.setUser(authData.user);
          store.isAuthModalOpen = false;
        } else {
          errorMessage = 'No fue posible crear la cuenta.';
        }
      } catch (err) {
        errorMessage = err.message || 'No fue posible registrar la cuenta. Verifica tus datos.';
      } finally {
        isLoading = false;
      }
      return;
    }

    try {
      const authData = await loginAPI(cleanEmail, cleanPassword);

      if (authData?.tokens) {
        localStorage.setItem('mias_access_token', authData.tokens.access_token);
        localStorage.setItem('mias_refresh_token', authData.tokens.refresh_token);
      }

      store.setUser(authData.user);
      if (['waiter', 'cashier', 'admin'].includes(authData.user?.role)) {
        store.staffRole = authData.user.role;
        store.currentView = 'staff';
      }
      store.isAuthModalOpen = false;
    } catch (err) {
      errorMessage = err.message || 'Correo o contraseña incorrectos.';
    } finally {
      isLoading = false;
    }
  }

  function handleSaveProfile() {
    if (!store.user) return;
    const updatedUser = { 
      ...store.user, 
      phone: profilePhone.trim(),
      address: profileAddress.trim()
    };
    store.setUser(updatedUser);
    profileSavedMessage = 'Datos de entrega guardados correctamente.';
    setTimeout(() => {
      profileSavedMessage = '';
    }, 2500);
  }

  function handleLogout() {
    store.logout();
    setMode('login');
  }

  function handleBack() {
    if (!store.user && viewMode === 'register') {
      viewMode = 'login';
    } else {
      store.isAuthModalOpen = false;
    }
  }

  function goToStamps() {
    store.isAuthModalOpen = false;
    store.currentView = 'stamps';
  }

  function goToOrders() {
    store.isAuthModalOpen = false;
    store.currentView = 'orders';
  }

  let isStaff = $derived(store.user && ['admin', 'cashier', 'waiter'].includes(store.user.role));
  let staffRoleLabel = $derived({
    admin: 'Administrador',
    cashier: 'Cajero',
    waiter: 'Mesero'
  }[store.user?.role] || '');

  $effect(() => {
    if (!store.isAuthModalOpen) return;
    const handleKeyDown = (e) => {
      if (e.key === 'Escape') {
        store.isAuthModalOpen = false;
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  });
</script>

{#if store.isAuthModalOpen}
  <div 
    data-lenis-prevent
    role="dialog"
    tabindex="-1"
    aria-modal="true"
    aria-label={store.user ? 'Mi Perfil' : (viewMode === 'register' ? 'Crear Cuenta' : 'Iniciar Sesión')}
    onwheel={(e) => e.stopPropagation()}
    ontouchmove={(e) => e.stopPropagation()}
    class="fixed inset-0 z-[70] w-full h-full min-h-[100dvh] bg-char text-bone overflow-y-auto overscroll-contain select-none"
    style="touch-action: pan-y; -webkit-overflow-scrolling: touch; padding-top: max(1rem, env(safe-area-inset-top, 0px)); padding-bottom: max(5rem, calc(env(safe-area-inset-bottom, 0px) + 3rem)); padding-left: max(1rem, env(safe-area-inset-left, 0px)); padding-right: max(1rem, env(safe-area-inset-right, 0px));"
  >
    <div data-lenis-prevent class="w-full max-w-md mx-auto py-2 sm:py-4 flex flex-col space-y-5">
      
      <!-- 1. Cabecera Editorial -->
      <header class="flex items-center justify-between w-full pb-1 shrink-0">
        <button 
          type="button"
          onclick={handleBack}
          class="w-12 h-12 rounded-full bg-char-card border hairline-dark flex items-center justify-center text-bone hover:border-white/20 active:scale-95 transition-all cursor-pointer shadow-lux"
          aria-label="Volver"
        >
          <ChevronLeft class="w-5 h-5 stroke-[2.2]" />
        </button>

        <h1 class="font-serif text-2xl text-bone font-normal tracking-tight leading-tight">
          {store.user ? 'Mi Perfil' : (viewMode === 'register' ? 'Crear Cuenta' : 'Iniciar Sesión')}
        </h1>

        {#if store.user}
          <button
            type="button"
            onclick={handleLogout}
            class="w-12 h-12 rounded-full bg-char-card border hairline-dark flex items-center justify-center text-ash hover:text-ember active:scale-95 transition-all cursor-pointer shadow-lux"
            title="Cerrar sesión"
            aria-label="Cerrar sesión"
          >
            <LogOut class="w-4 h-4 stroke-[2]" />
          </button>
        {:else}
          <div class="w-12 h-12 flex items-center justify-center">
            <img 
              src="/imagenes/marca/isotipo.webp" 
              alt="Mías" 
              class="w-6 h-6 object-contain opacity-70"
            />
          </div>
        {/if}
      </header>

      {#if store.user}
        <!-- ========================================================================= -->
        <!-- CASO A: PERFIL DEL COMENSAL (APPLE CARD GASTRONÓMICA)                     -->
        <!-- ========================================================================= -->
        
        <!-- 1. TARJETA DE IDENTIDAD -->
        <article class="rounded-card sm:rounded-card-lg bg-char-card border hairline-dark p-6 space-y-4 shadow-lux text-center">
          <div class="mx-auto w-20 h-20 rounded-full border-2 border-gold/40 p-1 bg-char flex items-center justify-center shadow-lux">
            {#if store.user.avatar_url}
              <img 
                src={store.user.avatar_url} 
                alt={store.user.full_name} 
                class="w-full h-full rounded-full object-cover"
              />
            {:else}
              <span class="font-serif font-bold text-2xl text-gold">
                {store.user.full_name?.charAt(0).toUpperCase() || 'M'}
              </span>
            {/if}
          </div>

          <div>
            <h2 class="font-serif text-2xl sm:text-3xl text-bone font-normal tracking-tight">
              {store.user.full_name}
            </h2>
            <p class="font-sans text-xs text-ash mt-0.5">
              {store.user.email}
            </p>
            {#if isStaff}
              <span class="inline-flex items-center gap-1.5 mt-2 px-3 py-1 rounded-full bg-gold/10 border border-gold/30 text-gold font-sans text-xs font-semibold">
                <Shield class="w-3.5 h-3.5" />
                <span>Equipo Mías • {staffRoleLabel}</span>
              </span>
            {/if}
          </div>
        </article>

        <!-- 2. TARJETA PASAPORTE DE FIDELIZACIÓN -->
        <article class="rounded-card sm:rounded-card-lg bg-char-card border hairline-dark p-5 sm:p-6 space-y-4 shadow-lux">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2.5">
              <Award class="w-5 h-5 text-gold" />
              <h3 class="font-serif text-xl text-bone font-normal">
                Pasaporte de Sellos
              </h3>
            </div>
            <span class="font-mono font-bold text-sm text-gold">
              {store.stampBalance} / 10
            </span>
          </div>

          <!-- Barra de Progreso de Sellos -->
          <div class="w-full h-2 rounded-full bg-char border hairline-dark overflow-hidden p-0.5">
            <div 
              class="h-full rounded-full bg-gold transition-all duration-300"
              style="width: {Math.min(100, (store.stampBalance / 10) * 100)}%;"
            ></div>
          </div>

          <p class="font-sans text-xs text-ash leading-relaxed">
            {#if store.stampBalance >= 10}
              ¡Tu cartilla está completa! Redime tu Hamburguesa Original gratis con tu mesero o en caja física.
            {:else}
              Acumula sellos en tus comandas con hamburguesas. Te faltan <strong class="text-gold font-semibold">{Math.max(0, 10 - store.stampBalance)} sellos</strong> para tu hamburguesa gratis.
            {/if}
          </p>

          <button
            type="button"
            onclick={goToStamps}
            class="w-full min-h-[52px] rounded-full bg-gold/10 hover:bg-gold/20 border border-gold/40 text-gold active:scale-[0.98] font-sans font-bold text-sm tracking-wide transition-all flex items-center justify-center gap-2 cursor-pointer shadow-lux"
          >
            <span>Ver mi Cartilla y Redimir</span>
            <ArrowRight class="w-4 h-4" />
          </button>
        </article>

        <!-- 3. TARJETA HISTORIAL DE COMANDAS -->
        <article class="rounded-card sm:rounded-card-lg bg-char-card border hairline-dark p-5 sm:p-6 space-y-3 shadow-lux">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2.5">
              <Receipt class="w-5 h-5 text-bone" />
              <h3 class="font-serif text-xl text-bone font-normal">
                Mis Pedidos
              </h3>
            </div>
            <span class="font-sans text-xs text-ash">
              {store.activeOrders.length} {store.activeOrders.length === 1 ? 'comanda' : 'comandas'}
            </span>
          </div>

          <p class="font-sans text-xs text-ash leading-relaxed">
            Revisa el estado de cocción en vivo, tus comprobantes de pago y solicita sellos para tus órdenes.
          </p>

          <button
            type="button"
            onclick={goToOrders}
            class="w-full min-h-[52px] rounded-full bg-char border hairline-dark hover:border-white/20 text-bone active:scale-[0.98] font-sans font-bold text-sm tracking-wide transition-all flex items-center justify-center gap-2 cursor-pointer shadow-lux"
          >
            <span>Ver mis Comandas</span>
            <ArrowRight class="w-4 h-4" />
          </button>
        </article>

        <!-- 4. DATOS DE ENTREGA A DOMICILIO (DUITAMA) -->
        <article class="rounded-card sm:rounded-card-lg bg-char-card border hairline-dark p-5 sm:p-6 space-y-4 shadow-lux">
          <div class="flex items-center gap-2.5">
            <MapPin class="w-5 h-5 text-gold" />
            <h3 class="font-serif text-xl text-bone font-normal">
              Datos de Entrega (Duitama)
            </h3>
          </div>

          <div class="space-y-3">
            <div>
              <label for="profile-phone" class="font-sans text-xs font-medium text-ash block mb-1">
                Teléfono de contacto
              </label>
              <div class="relative">
                <Phone class="w-4 h-4 text-ash absolute left-4 top-1/2 -translate-y-1/2 pointer-events-none" />
                <input 
                  id="profile-phone"
                  type="tel"
                  bind:value={profilePhone}
                  placeholder="Ej: 321 252 0881"
                  class="w-full min-h-[50px] rounded-full pl-11 pr-4 bg-char border hairline-dark text-bone font-sans text-sm focus:border-gold outline-none transition-colors"
                />
              </div>
            </div>

            <div>
              <label for="profile-address" class="font-sans text-xs font-medium text-ash block mb-1">
                Dirección o barrio habitual
              </label>
              <div class="relative">
                <MapPin class="w-4 h-4 text-ash absolute left-4 top-1/2 -translate-y-1/2 pointer-events-none" />
                <input 
                  id="profile-address"
                  type="text"
                  bind:value={profileAddress}
                  placeholder="Ej: Carrera 16 # 14-32"
                  class="w-full min-h-[50px] rounded-full pl-11 pr-4 bg-char border hairline-dark text-bone font-sans text-sm focus:border-gold outline-none transition-colors"
                />
              </div>
            </div>
          </div>

          {#if profileSavedMessage}
            <p class="font-sans text-xs font-medium text-gold text-center py-1">
              {profileSavedMessage}
            </p>
          {/if}

          <button
            type="button"
            onclick={handleSaveProfile}
            class="w-full min-h-[50px] rounded-full bg-char hover:bg-white/5 border border-gold/30 text-gold active:scale-[0.98] font-sans font-bold text-xs tracking-wide transition-all cursor-pointer"
          >
            Guardar Datos
          </button>
        </article>

        <!-- 5. TERMINAL DE PERSONAL (SOLO STAFF) -->
        {#if isStaff}
          <button
            type="button"
            onclick={() => { store.isAuthModalOpen = false; store.currentView = 'staff'; }}
            class="w-full min-h-[54px] rounded-full bg-gold/15 hover:bg-gold/25 border border-gold/40 text-gold font-sans font-bold text-sm tracking-wide transition-all flex items-center justify-center gap-2 cursor-pointer shadow-lux"
          >
            <Shield class="w-4 h-4" />
            <span>Acceder a la Terminal ({staffRoleLabel})</span>
          </button>
        {/if}

        <!-- 6. CERRAR SESIÓN -->
        <div class="pt-2">
          <button
            type="button"
            onclick={handleLogout}
            class="w-full min-h-[52px] rounded-full bg-char-card border hairline-dark text-ash hover:text-ember active:scale-[0.98] font-sans font-medium text-sm transition-all flex items-center justify-center gap-2 cursor-pointer"
          >
            <LogOut class="w-4 h-4" />
            <span>Cerrar Sesión</span>
          </button>
        </div>

      {:else}
        <!-- ========================================================================= -->
        <!-- CASO B: LOGIN / REGISTRO DEL CLIENTE                                     -->
        <!-- ========================================================================= -->
        <article class="rounded-card sm:rounded-card-lg bg-char-card border hairline-dark p-6 sm:p-7 space-y-5 shadow-lux">
          
          <div class="text-center space-y-1">
            <h2 class="font-serif text-2xl sm:text-3xl text-bone font-normal tracking-tight">
              {viewMode === 'register' ? 'Crea tu Pasaporte' : 'Bienvenido a Mías'}
            </h2>
            <p class="font-sans text-xs text-ash">
              {viewMode === 'register' ? 'Regístrate para acumular sellos en cada pedido' : 'Inicia sesión para consultar tus sellos y comandas'}
            </p>
          </div>

          <!-- Botón de Google OAuth -->
          <button
            type="button"
            onclick={handleGoogleClick}
            disabled={isLoading}
            class="w-full min-h-[52px] rounded-full bg-white text-char hover:bg-bone active:scale-[0.98] font-sans font-bold text-sm flex items-center justify-center gap-3 transition-all cursor-pointer shadow-lux disabled:opacity-50"
          >
            <svg class="w-4 h-4" viewBox="0 0 24 24">
              <path fill="#4285F4" d="M23.745 12.27c0-.7-.06-1.4-.19-2.07H12v4.51h6.6c-.29 1.52-1.14 2.82-2.4 3.68v3.05h3.88c2.27-2.09 3.665-5.17 3.665-9.17z"/>
              <path fill="#34A853" d="M12 24c3.24 0 5.95-1.08 7.93-2.91l-3.88-3.05c-1.08.72-2.45 1.16-4.05 1.16-3.12 0-5.77-2.1-6.72-4.93H1.25v3.15C3.26 21.36 7.33 24 12 24z"/>
              <path fill="#FBBC05" d="M5.28 14.27c-.25-.72-.38-1.49-.38-2.27s.13-1.55.38-2.27V6.58H1.25C.45 8.18 0 9.99 0 12s.45 3.82 1.25 5.42l4.03-3.15z"/>
              <path fill="#EA4335" d="M12 4.75c1.77 0 3.35.61 4.6 1.8l3.42-3.42C17.95 1.19 15.24 0 12 0 7.33 0 3.26 2.64 1.25 6.58l4.03 3.15c.95-2.83 3.6-4.98 6.72-4.98z"/>
            </svg>
            <span>Continuar con Google</span>
          </button>

          <div class="flex items-center gap-3">
            <div class="flex-1 border-t hairline-dark"></div>
            <span class="font-sans text-[11px] text-ash uppercase">o con correo</span>
            <div class="flex-1 border-t hairline-dark"></div>
          </div>

          {#if errorMessage}
            <div class="p-3.5 rounded-2xl bg-ember/10 border border-ember/30 text-ember-light font-sans text-xs flex items-center gap-2">
              <AlertCircle class="w-4 h-4 shrink-0" />
              <span>{errorMessage}</span>
            </div>
          {/if}

          <!-- Formulario de Credenciales con Labels Visibles (Regla M) -->
          <form onsubmit={(e) => { e.preventDefault(); handleAuthSubmit(); }} class="space-y-3">
            
            {#if viewMode === 'register'}
              <div>
                <label for="auth-fullname" class="font-sans text-xs font-medium text-ash block mb-1">
                  Nombre Completo
                </label>
                <div class="relative">
                  <User class="w-4 h-4 text-ash absolute left-4 top-1/2 -translate-y-1/2 pointer-events-none" />
                  <input
                    id="auth-fullname"
                    type="text"
                    bind:value={fullName}
                    placeholder="Tu nombre y apellido"
                    class="w-full min-h-[50px] rounded-full pl-11 pr-4 bg-char border hairline-dark text-bone font-sans text-sm focus:border-gold outline-none transition-colors"
                  />
                </div>
              </div>
            {/if}

            <div>
              <label for="auth-email" class="font-sans text-xs font-medium text-ash block mb-1">
                Correo Electrónico
              </label>
              <div class="relative">
                <Mail class="w-4 h-4 text-ash absolute left-4 top-1/2 -translate-y-1/2 pointer-events-none" />
                <input
                  id="auth-email"
                  type="email"
                  bind:value={email}
                  placeholder="ejemplo@correo.com"
                  class="w-full min-h-[50px] rounded-full pl-11 pr-4 bg-char border hairline-dark text-bone font-sans text-sm focus:border-gold outline-none transition-colors"
                />
              </div>
            </div>

            <div>
              <label for="auth-password" class="font-sans text-xs font-medium text-ash block mb-1">
                Contraseña
              </label>
              <div class="relative">
                <Lock class="w-4 h-4 text-ash absolute left-4 top-1/2 -translate-y-1/2 pointer-events-none" />
                <input
                  id="auth-password"
                  type={showPassword ? 'text' : 'password'}
                  bind:value={password}
                  placeholder="••••••••"
                  class="w-full min-h-[50px] rounded-full pl-11 pr-11 bg-char border hairline-dark text-bone font-sans text-sm focus:border-gold outline-none transition-colors"
                />
                <button
                  type="button"
                  onclick={() => showPassword = !showPassword}
                  class="w-9 h-9 rounded-full absolute right-2 top-1/2 -translate-y-1/2 flex items-center justify-center text-ash hover:text-bone transition-colors"
                  aria-label={showPassword ? 'Ocultar contraseña' : 'Ver contraseña'}
                >
                  {#if showPassword}
                    <EyeOff class="w-4 h-4" />
                  {:else}
                    <Eye class="w-4 h-4" />
                  {/if}
                </button>
              </div>
            </div>

            <button
              type="submit"
              disabled={isLoading}
              class="w-full min-h-[54px] rounded-full bg-ember hover:bg-ember-light active:scale-[0.98] text-white font-sans font-bold text-base tracking-wide transition-all flex items-center justify-center gap-2 cursor-pointer shadow-lux disabled:opacity-50 mt-4"
            >
              {#if isLoading}
                <Loader2 class="w-4 h-4 animate-spin" />
                <span>Procesando...</span>
              {:else}
                <span>{viewMode === 'register' ? 'Crear mi Cuenta' : 'Entrar a Mías'}</span>
                <ArrowRight class="w-4 h-4" />
              {/if}
            </button>
          </form>

          <!-- Alternar Modo Login / Registro -->
          <div class="text-center pt-2">
            {#if viewMode === 'register'}
              <p class="font-sans text-xs text-ash">
                ¿Ya tienes una cuenta registrada? 
                <button 
                  type="button" 
                  onclick={() => setMode('login')} 
                  class="text-gold font-semibold hover:underline ml-1 cursor-pointer"
                >
                  Inicia sesión aquí
                </button>
              </p>
            {:else}
              <p class="font-sans text-xs text-ash">
                ¿Aún no tienes cuenta? 
                <button 
                  type="button" 
                  onclick={() => setMode('register')} 
                  class="text-gold font-semibold hover:underline ml-1 cursor-pointer"
                >
                  Regístrate gratis
                </button>
              </p>
            {/if}
          </div>

        </article>
      {/if}

    </div>
  </div>
{/if}
