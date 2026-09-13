<script>
  import { store } from '../lib/store.svelte.js';
  import { staffLoginAPI, oauthLoginAPI } from '../lib/api.js';
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
    ShoppingBag,
    LogOut,
    Check,
    ArrowRight,
    Utensils,
    Sparkles,
    MessageCircle,
    MapPin
  } from 'lucide-svelte';

  // Modos de vista: 'login' o 'register' (cuando no hay sesión activa)
  let viewMode = $state('login');
  
  // Campos del formulario de autenticación
  let email = $state('');
  let password = $state('');
  let fullName = $state('');
  let rememberMe = $state(true);
  let showPassword = $state(false);

  // Estados de interacción
  let isLoading = $state(false);
  let errorMessage = $state('');
  let profileSavedMessage = $state('');

  // Datos editables de perfil
  let profilePhone = $state('');
  let profileAddress = $state('');
  let favoriteBread = $state('pretzel');
  let notifyWhatsApp = $state(true);

  // Sincronizar datos del usuario cuando se abre el perfil
  $effect(() => {
    if (store.user) {
      profilePhone = store.user.phone || '';
      profileAddress = store.user.address || 'Carrera 16 # 14-32, Duitama';
      if (typeof window !== 'undefined') {
        const savedBread = localStorage.getItem('mias_pref_bread');
        if (savedBread) favoriteBread = savedBread;
        const savedNotify = localStorage.getItem('mias_pref_notify');
        if (savedNotify !== null) notifyWhatsApp = savedNotify === 'true';
      }
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

  // Flujo Google OAuth (Píldora Gorda)
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
            console.error('[Google OAuth] Error de autenticación:', tokenResponse.error);
            errorMessage = 'No fue posible autenticar con Google. Por favor intenta de nuevo o ingresa con tu correo.';
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
                avatar_url: googleProfile.picture,
                role: 'customer'
              });
            }
            store.isAuthModalOpen = false;
          } catch (err) {
            console.error('[Google OAuth] Error procesando perfil:', err);
            errorMessage = 'No fue posible iniciar sesión con Google. Por favor intenta nuevamente o ingresa con tu correo.';
          } finally {
            isLoading = false;
          }
        },
        error_callback: (err) => {
          isLoading = false;
          if (err && err.type === 'popup_closed') {
            return;
          }
          console.error('[Google OAuth] Error de origen/conexión:', err, '(Verifica que el dominio esté autorizado en Google Cloud Console)');
          errorMessage = 'No fue posible conectar con Google en este momento. Por favor intenta nuevamente o inicia sesión con tu correo.';
        }
      });

      client.requestAccessToken();
    } catch (err) {
      isLoading = false;
      console.error('[Google OAuth] Error iniciando conexión:', err);
      errorMessage = 'No fue posible conectar con Google en este momento. Por favor intenta con tu correo.';
    }
  }

  // Envío Unificado: Maneja acceso transparente para personal y clientes
  async function handleLoginSubmit() {
    const cleanEmail = email.trim();
    const cleanPassword = password.trim();

    if (!cleanEmail || !cleanPassword) {
      errorMessage = 'Ingresa tu correo y contraseña';
      return;
    }

    if (viewMode === 'register' && !fullName.trim()) {
      errorMessage = 'Ingresa tu nombre completo';
      return;
    }

    isLoading = true;
    errorMessage = '';

    if (viewMode === 'register') {
      try {
        const authData = await oauthLoginAPI({
          email: cleanEmail,
          full_name: fullName.trim(),
          avatar_url: '',
          phone: '',
          habeas_data_accepted: true
        }).catch(() => null);

        if (authData?.tokens) {
          localStorage.setItem('mias_access_token', authData.tokens.access_token);
          localStorage.setItem('mias_refresh_token', authData.tokens.refresh_token);
          store.setUser(authData.user);
        } else {
          store.setUser({
            id: 'usr-' + Math.random().toString(36).substring(2, 9),
            full_name: fullName.trim(),
            email: cleanEmail,
            role: 'customer'
          });
        }
        store.isAuthModalOpen = false;
      } catch (err) {
        console.error('[Auth] Error al registrar cuenta:', err);
        errorMessage = 'No fue posible crear la cuenta en este momento. Por favor verifica tus datos e intenta de nuevo.';
      } finally {
        isLoading = false;
      }
      return;
    }

    // Modo Login
    try {
      const staffData = await staffLoginAPI(cleanEmail, cleanPassword).catch(() => null);

      if (staffData && staffData.user && (staffData.user.role === 'waiter' || staffData.user.role === 'cashier' || staffData.user.role === 'admin')) {
        if (staffData.tokens) {
          localStorage.setItem('mias_access_token', staffData.tokens.access_token);
          localStorage.setItem('mias_refresh_token', staffData.tokens.refresh_token);
        }
        store.setUser(staffData.user);
        store.staffRole = staffData.user.role;
        store.currentView = 'staff';
        store.isAuthModalOpen = false;
        return;
      }

      // Fallback staff offline demo
      if (cleanEmail === 'admin@mias.com' && cleanPassword === 'AdminMias2026!') {
        store.setUser({ id: 'staff-admin', full_name: 'Administrador General', email: cleanEmail, role: 'admin' });
        store.staffRole = 'admin';
        store.currentView = 'staff';
        store.isAuthModalOpen = false;
        return;
      }
      if (cleanEmail === 'cajero@mias.com' && cleanPassword === 'CajeroMias2026!') {
        store.setUser({ id: 'staff-cashier', full_name: 'Cajero Principal', email: cleanEmail, role: 'cashier' });
        store.staffRole = 'cashier';
        store.currentView = 'staff';
        store.isAuthModalOpen = false;
        return;
      }
      if (cleanEmail === 'mesero@mias.com' && cleanPassword === 'MeseroMias2026!') {
        store.setUser({ id: 'staff-waiter', full_name: 'Mesero Sala', email: cleanEmail, role: 'waiter' });
        store.staffRole = 'waiter';
        store.currentView = 'staff';
        store.isAuthModalOpen = false;
        return;
      }

      // Autenticar cliente
      const clientData = await oauthLoginAPI({
        email: cleanEmail,
        full_name: cleanEmail.split('@')[0],
        avatar_url: '',
        phone: '',
        habeas_data_accepted: true
      }).catch(() => null);

      if (clientData?.tokens) {
        localStorage.setItem('mias_access_token', clientData.tokens.access_token);
        localStorage.setItem('mias_refresh_token', clientData.tokens.refresh_token);
        store.setUser(clientData.user);
      } else {
        store.setUser({
          id: 'usr-' + Math.random().toString(36).substring(2, 9),
          full_name: cleanEmail.split('@')[0],
          email: cleanEmail,
          role: 'customer'
        });
      }
      store.isAuthModalOpen = false;
    } catch (err) {
      errorMessage = 'Credenciales no válidas';
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
    if (typeof window !== 'undefined') {
      localStorage.setItem('mias_pref_bread', favoriteBread);
      localStorage.setItem('mias_pref_notify', String(notifyWhatsApp));
    }
    profileSavedMessage = 'Datos guardados correctamente';
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
</script>

{#if store.isAuthModalOpen}
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <!-- VISTA A PANTALLA COMPLETA RESPONSIVA CON data-lenis-prevent (GARANTIZA SCROLL FLUIDO EN MÓVIL Y DESKTOP) -->
  <div 
    data-lenis-prevent
    onwheel={(e) => e.stopPropagation()}
    ontouchmove={(e) => e.stopPropagation()}
    class="fixed inset-0 z-50 w-full h-full min-h-[100dvh] bg-char text-bone overflow-y-auto overscroll-contain"
    style="touch-action: pan-y; -webkit-overflow-scrolling: touch; padding-top: max(1rem, env(safe-area-inset-top, 0px)); padding-bottom: max(5rem, calc(env(safe-area-inset-bottom, 0px) + 3rem)); padding-left: max(1rem, env(safe-area-inset-left, 0px)); padding-right: max(1rem, env(safe-area-inset-right, 0px));"
  >
    <!-- Contenedor con flujo vertical natural -->
    <div data-lenis-prevent class="w-full max-w-md mx-auto py-2 sm:py-4 flex flex-col space-y-6">
      
      <!-- 1. Cabecera Superior: Botón Atrás Circular Dark Plano del Menú -->
      <header class="flex items-center justify-between w-full pb-1 shrink-0">
        <button 
          type="button"
          onclick={handleBack}
          class="w-14 h-14 sm:w-16 sm:h-16 rounded-full bg-char-card border border-white/10 flex items-center justify-center text-bone hover:bg-char-card-hover active:scale-90 transition-all cursor-pointer group shrink-0"
          aria-label="Volver"
        >
          <ChevronLeft class="w-6 h-6 sm:w-7 sm:h-7 stroke-[2.2] -translate-x-0.5 group-hover:-translate-x-1 transition-transform" />
        </button>

        <span class="font-display font-extrabold text-base sm:text-lg text-bone tracking-tight">
          {store.user ? 'Mi Perfil' : 'Acceso Mías'}
        </span>

        <!-- Botón Salir o Isotipo -->
        {#if store.user}
          <button
            type="button"
            onclick={handleLogout}
            class="w-14 h-14 sm:w-16 sm:h-16 rounded-full bg-char-card border border-white/10 flex items-center justify-center text-ash hover:text-[#FF2400] active:scale-90 transition-all cursor-pointer shrink-0"
            title="Cerrar sesión"
            aria-label="Cerrar sesión"
          >
            <LogOut class="w-5 h-5 stroke-[2]" />
          </button>
        {:else}
          <div class="w-14 h-14 sm:w-16 sm:h-16 flex items-center justify-center shrink-0">
            <img 
              src="/imagenes/marca/isotipo.webp" 
              alt="Mías" 
              class="w-7 h-7 object-contain opacity-85"
            />
          </div>
        {/if}
      </header>

      {#if store.user}
        <!-- ========================================================================= -->
        <!-- CASO A: PERFIL DE USUARIO                                                 -->
        <!-- ========================================================================= -->
        
        <!-- SECCIÓN 1: TARJETA DE PERFIL HERO (EXACTA A LA REFERENCIA CON PALETA MÍAS) -->
        <div class="bg-char-card border border-white/10 rounded-[38px] overflow-hidden shadow-none">
          
          <!-- Banner Superior Gastronómico con Botón Tipo "Follow +" -->
          <div class="h-40 sm:h-48 w-full relative bg-[#0C0C0E] overflow-hidden">
            <img 
              src="/imagenes/local/restaurante_mias_desde_dentro.webp" 
              alt="Ambiente Mías" 
              class="w-full h-full object-cover opacity-80"
            />
            <div class="absolute inset-0 bg-gradient-to-t from-black/60 via-black/20 to-black/10"></div>

            <!-- Botón Flotante Superior Derecho (Fiel a la referencia "Follow +") -->
            <div class="absolute top-4 right-4">
              <span class="px-4 py-1.5 rounded-full bg-char-card/95 backdrop-blur-md border border-white/10 text-xs font-mono font-bold text-bone flex items-center gap-1.5 shadow-xs">
                <span>Club Mías</span>
                <span class="text-[#FF2400] font-black">★</span>
              </span>
            </div>
          </div>

          <!-- Avatar Superpuesto y Barra de Experiencia / Nivel "exp. ||||||||||||" -->
          <div class="px-6 pt-0 pb-6 -mt-14 sm:-mt-16 relative z-10">
            <div class="flex items-end justify-between gap-3">
              <!-- Avatar Circular con Anillo de Acento Rojo Vivo (#FF2400) -->
              <div class="w-24 h-24 sm:w-26 sm:h-26 rounded-full border-4 border-char overflow-hidden bg-[#0C0C0E] text-white flex items-center justify-center font-display font-black text-2xl sm:text-3xl shrink-0 ring-2 ring-[#FF2400]/50 shadow-sm">
                {#if store.user.avatar_url}
                  <img 
                    src={store.user.avatar_url} 
                    alt={store.user.full_name} 
                    class="w-full h-full object-cover"
                  />
                {:else}
                  <span>{store.user.full_name?.charAt(0).toUpperCase() || 'M'}</span>
                {/if}
              </div>

              <!-- Indicador "exp. ||||||||||||" fiel a la imagen de referencia -->
              <div class="pb-1 text-right">
                <div class="flex items-center gap-1.5 justify-end">
                  <span class="text-xs font-mono text-ash lowercase">exp.</span>
                  <div class="flex items-center gap-0.5 sm:gap-1">
                    {#each Array(16) as _, i}
                      <div 
                        class="w-1 sm:w-1.5 h-3.5 rounded-full transition-all {i < Math.min(16, (store.stampBalance * 2) || 4) ? 'bg-[#FF2400]' : 'bg-white/10'}"
                      ></div>
                    {/each}
                  </div>
                </div>
                <span class="text-[10px] font-mono uppercase text-ash tracking-widest block mt-1.5">
                  Nivel • {store.stampBalance >= 10 ? 'VIP Gold' : (store.stampBalance >= 5 ? 'Sibarita' : 'Comensal')}
                </span>
              </div>
            </div>

            <!-- Nombre y Biografía del Comensal -->
            <div class="mt-4">
              <h1 class="font-display font-black text-2xl sm:text-3xl text-bone tracking-tight leading-tight">
                {store.user.full_name}
              </h1>
              <p class="text-xs sm:text-sm text-ash font-medium mt-1 leading-snug">
                Comensal de Honor • Experiencia Gastronómica al Carbón
              </p>
            </div>

            <!-- Fila de 3 Métricas en Caja Píldora -->
            <div class="mt-5 p-3.5 sm:p-4 rounded-2xl bg-char border border-white/10 grid grid-cols-3 text-center divide-x divide-white/10">
              <div>
                <span class="font-display font-extrabold text-lg sm:text-xl text-[#FF2400] block leading-none">
                  {store.stampBalance}
                </span>
                <span class="text-[10px] font-mono uppercase text-ash tracking-wider block mt-1.5">
                  Sellos
                </span>
              </div>
              <div>
                <span class="font-display font-extrabold text-lg sm:text-xl text-bone block leading-none">
                  {store.activeOrders.length || 0}
                </span>
                <span class="text-[10px] font-mono uppercase text-ash tracking-wider block mt-1.5">
                  Comandas
                </span>
              </div>
              <div>
                <span class="font-display font-extrabold text-lg sm:text-xl text-bone block leading-none">
                  {store.stampBalance >= 10 ? '1 Gratis' : `${10 - (store.stampBalance % 10)} faltan`}
                </span>
                <span class="text-[10px] font-mono uppercase text-ash tracking-wider block mt-1.5">
                  Meta Gratis
                </span>
              </div>
            </div>

            <!-- Accesos Rápidos de Redes y Contacto -->
            <div class="mt-4 pt-4 border-t border-white/10 grid grid-cols-3 divide-x divide-white/10 text-center text-ash">
              <a 
                href="https://instagram.com" 
                target="_blank" 
                rel="noreferrer"
                class="flex items-center justify-center gap-1.5 text-xs font-mono hover:text-[#FF2400] transition-colors py-1 cursor-pointer"
                title="Instagram @mias"
              >
                <svg class="w-4 h-4 fill-none stroke-current stroke-2 shrink-0" viewBox="0 0 24 24">
                  <rect width="20" height="20" x="2" y="2" rx="5" ry="5"/>
                  <path d="M16 11.37A4 4 0 1 1 12.63 8 4 4 0 0 1 16 11.37z"/>
                  <line x1="17.5" x2="17.51" y1="6.5" y2="6.5"/>
                </svg>
                <span>@mias</span>
              </a>
              <a 
                href="https://wa.me/573212520881" 
                target="_blank" 
                rel="noreferrer"
                class="flex items-center justify-center gap-1.5 text-xs font-mono hover:text-[#FF2400] transition-colors py-1 cursor-pointer"
                title="WhatsApp de Atención"
              >
                <svg class="w-4 h-4 fill-none stroke-current stroke-2 shrink-0" viewBox="0 0 24 24">
                  <path d="M22 16.92v3a2 2 0 0 1-2.18 2 19.79 19.79 0 0 1-8.63-3.07 19.5 19.5 0 0 1-6-6 19.79 19.79 0 0 1-3.07-8.67A2 2 0 0 1 4.11 2h3a2 2 0 0 1 2 1.72 12.84 12.84 0 0 0 .7 2.81 2 2 0 0 1-.45 2.11L8.09 9.91a16 16 0 0 0 6 6l1.27-1.27a2 2 0 0 1 2.11-.45 12.84 12.84 0 0 0 2.81.7A2 2 0 0 1 22 16.92z"/>
                </svg>
                <span>WhatsApp</span>
              </a>
              <span class="flex items-center justify-center gap-1.5 text-xs font-mono text-ash py-1">
                <MapPin class="w-3.5 h-3.5 text-[#FF2400] shrink-0" />
                <span>Duitama</span>
              </span>
            </div>
          </div>
        </div>

        <!-- Acceso Directo a Terminal para Personal (Staff) -->
        {#if store.user.role === 'admin' || store.user.role === 'cashier' || store.user.role === 'waiter'}
          <div class="p-5 rounded-[28px] bg-char-card border border-white/10 text-center space-y-2">
            <span class="text-xs font-mono font-bold uppercase text-[#FF2400] block tracking-wider">
              Turno Operativo Activo ({store.user.role})
            </span>
            <button
              type="button"
              onclick={() => { store.isAuthModalOpen = false; store.currentView = 'staff'; }}
              class="w-full min-h-[56px] rounded-full bg-[#FF2400] hover:bg-[#E02000] text-white font-display font-extrabold text-base flex items-center justify-center gap-2 active:scale-[0.97] transition-all cursor-pointer"
            >
              <span>Acceder a la Terminal</span>
              <ArrowRight class="w-5 h-5" />
            </button>
          </div>
        {/if}

        <!-- SECCIÓN 2: FIDELIZACIÓN (PROGRAMA DE SELLOS) -->
        <div class="bg-char-card border border-white/10 rounded-[36px] p-6 sm:p-7 space-y-5 shadow-none">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <Award class="w-5 h-5 text-[#FF2400]" />
              <h2 class="font-display font-extrabold text-xl sm:text-2xl text-bone tracking-tight">
                Programa de Sellos
              </h2>
            </div>
            <span class="min-h-[32px] px-3.5 rounded-full bg-char border border-white/10 text-xs font-mono font-bold text-bone flex items-center">
              {store.stampBalance} de 10 Sellos
            </span>
          </div>

          <!-- Leyenda de Estados de Sellos -->
          <div class="flex items-center justify-center gap-6 py-1 border-y border-white/10 text-xs font-medium">
            <div class="flex items-center gap-2">
              <div class="w-4 h-4 rounded-full bg-[#FF2400] flex items-center justify-center text-white">
                <Check class="w-2.5 h-2.5 stroke-[3]" />
              </div>
              <span class="text-bone">Acumulados ({store.stampBalance})</span>
            </div>
            <div class="flex items-center gap-2">
              <div class="w-4 h-4 rounded-full bg-char border border-white/20"></div>
              <span class="text-ash">Pendientes ({Math.max(0, 10 - store.stampBalance)})</span>
            </div>
          </div>

          <!-- Grilla Táctil de 10 Sellos (2 Filas de 5 Círculos Grandes) -->
          <div class="grid grid-cols-5 gap-2.5 py-1">
            {#each Array(10) as _, i}
              <div class="flex flex-col items-center gap-1">
                <div 
                  class="w-12 h-12 sm:w-14 sm:h-14 rounded-full flex items-center justify-center font-display font-extrabold text-sm transition-all {i < store.stampBalance ? 'bg-[#FF2400] text-white shadow-xs' : 'bg-char text-ash/80 border border-white/10'}"
                >
                  {#if i < store.stampBalance}
                    <Check class="w-5 h-5 stroke-[2.5]" />
                  {:else}
                    <span>0{i + 1}</span>
                  {/if}
                </div>
                <span class="text-[10px] font-mono text-ash">S-{i + 1}</span>
              </div>
            {/each}
          </div>

          <div class="p-4 rounded-2xl bg-char border border-white/10">
            <p class="text-xs text-bone leading-relaxed">
              {#if store.stampBalance >= 10}
                <strong class="text-emerald-400 font-bold block mb-1">¡Felicidades! Has completado tu cartilla.</strong>
                Pide tu Hamburguesa Original gratis con tu mesero o en tu próximo pedido en línea.
              {:else}
                Acumula sellos por cada hamburguesa registrada en tus comandas. Te faltan <strong class="text-[#FF2400] font-bold">{10 - store.stampBalance} sellos</strong> para tu hamburguesa gratis.
              {/if}
            </p>
          </div>

          <!-- Botón Fat-Finger para ir a la Cartilla Completa -->
          <button
            type="button"
            onclick={goToStamps}
            class="w-full min-h-[60px] sm:min-h-[64px] rounded-full bg-[#FF2400] hover:bg-[#E02000] active:scale-[0.97] text-white font-display font-extrabold text-base sm:text-lg tracking-tight transition-all duration-150 flex items-center justify-center gap-2 cursor-pointer pt-0.5"
          >
            <span>Ver Pasaporte y Redimir</span>
            <ArrowRight class="w-5 h-5" />
          </button>
        </div>

        <!-- SECCIÓN 3: COMANDAS Y PEDIDOS RECIENTES -->
        <div class="bg-char-card border border-white/10 rounded-[36px] p-6 sm:p-7 space-y-4 shadow-none">
          <div class="flex items-center justify-between">
            <div class="flex items-center gap-2">
              <ShoppingBag class="w-5 h-5 text-[#FF2400]" />
              <h2 class="font-display font-extrabold text-xl text-bone">Mis Comandas</h2>
            </div>
            <span class="text-xs font-mono font-bold text-ash">
              {store.activeOrders.length} {store.activeOrders.length === 1 ? 'orden' : 'órdenes'}
            </span>
          </div>

          {#if store.activeOrders.length > 0}
            <div class="space-y-2.5 pt-1">
              {#each store.activeOrders.slice(0, 3) as order}
                <div class="p-4 rounded-2xl bg-char border border-white/10 flex items-center justify-between">
                  <div>
                    <span class="font-mono text-xs font-bold text-[#FF2400]">#{order.code}</span>
                    <p class="text-sm text-bone font-display font-extrabold mt-0.5">${order.total?.toLocaleString('es-CO')}</p>
                  </div>
                  <span class="text-xs font-mono font-bold px-3 py-1 rounded-full bg-char-card border border-white/10 text-bone">
                    {order.status || 'En cocina'}
                  </span>
                </div>
              {/each}
            </div>
          {:else}
            <div class="p-5 rounded-2xl bg-char/60 border border-white/10 text-center space-y-3">
              <p class="text-xs text-ash">
                No tienes comandas registradas en este dispositivo actualmente.
              </p>
              <button
                type="button"
                onclick={() => { store.isAuthModalOpen = false; store.currentView = 'menu'; }}
                class="min-h-[46px] px-6 rounded-full bg-[#FF2400] text-white text-xs font-display font-bold hover:bg-[#E02000] active:scale-95 transition-all cursor-pointer"
              >
                Explorar la Carta
              </button>
            </div>
          {/if}
        </div>

        <!-- SECCIÓN 4: DATOS DE CONTACTO Y ENTREGA A DOMICILIO -->
        <div class="bg-char-card border border-white/10 rounded-[36px] p-6 sm:p-7 space-y-4 shadow-none">
          <div class="flex items-center gap-2">
            <Phone class="w-5 h-5 text-[#FF2400]" />
            <h2 class="font-display font-extrabold text-xl text-bone">
              Contacto y Domicilio
            </h2>
          </div>
          <p class="text-xs text-ash leading-relaxed">
            Recibe avisos de estado por WhatsApp y registra tu dirección habitual para entregas a domicilio en Duitama.
          </p>

          <div class="space-y-3 pt-1">
            <!-- Teléfono WhatsApp -->
            <div class="relative">
              <div class="absolute inset-y-0 left-0 pl-5 flex items-center pointer-events-none text-ash">
                <Phone class="w-5 h-5 stroke-[2]" />
              </div>
              <input 
                type="tel"
                bind:value={profilePhone}
                placeholder="Teléfono WhatsApp (ej. 321 252 0881)"
                class="w-full min-h-[58px] rounded-full py-4 pl-14 pr-5 bg-char border border-white/10 text-bone placeholder:text-ash/60 text-base font-sans focus:outline-none focus:border-[#FF2400]/60 transition-colors"
              />
            </div>

            <!-- Dirección habitual -->
            <div class="relative">
              <div class="absolute inset-y-0 left-0 pl-5 flex items-center pointer-events-none text-ash">
                <MapPin class="w-5 h-5 stroke-[2]" />
              </div>
              <input 
                type="text"
                bind:value={profileAddress}
                placeholder="Dirección o barrio (ej. Carrera 11 # 15-40)"
                class="w-full min-h-[58px] rounded-full py-4 pl-14 pr-5 bg-char border border-white/10 text-bone placeholder:text-ash/60 text-base font-sans focus:outline-none focus:border-[#FF2400]/60 transition-colors"
              />
            </div>
          </div>

          {#if profileSavedMessage}
            <span class="text-xs font-bold text-emerald-400 block text-center py-1">{profileSavedMessage}</span>
          {/if}

          <button
            type="button"
            onclick={handleSaveProfile}
            class="w-full min-h-[56px] rounded-full bg-[#FF2400] hover:bg-[#E02000] text-white font-display font-extrabold text-sm active:scale-[0.97] transition-all cursor-pointer"
          >
            Guardar Datos
          </button>
        </div>

        <!-- SECCIÓN 5: PREFERENCIAS GASTRONÓMICAS -->
        <div class="bg-char-card border border-white/10 rounded-[36px] p-6 sm:p-7 space-y-4 shadow-none">
          <div class="flex items-center gap-2">
            <Utensils class="w-5 h-5 text-[#FF2400]" />
            <h2 class="font-display font-extrabold text-xl text-bone">
              Preferencias del Comensal
            </h2>
          </div>
          <p class="text-xs text-ash">
            Personaliza tu experiencia culinaria para agilizar tus futuras comandas.
          </p>

          <div class="space-y-3 pt-1">
            <!-- Selección de Pan Predilecto -->
            <div>
              <span class="text-xs font-mono font-bold text-bone block mb-2">
                Pan Artesanal Favorito
              </span>
              <div class="grid grid-cols-2 gap-2">
                <button
                  type="button"
                  onclick={() => favoriteBread = 'pretzel'}
                  class="min-h-[52px] p-3 rounded-2xl border text-xs font-display font-bold transition-all text-center cursor-pointer {favoriteBread === 'pretzel' ? 'bg-[#FF2400] text-white border-[#FF2400]' : 'bg-char text-bone border-white/10 hover:border-white/20'}"
                >
                  Pan Pretzel (Insignia)
                </button>
                <button
                  type="button"
                  onclick={() => favoriteBread = 'brioche'}
                  class="min-h-[52px] p-3 rounded-2xl border text-xs font-display font-bold transition-all text-center cursor-pointer {favoriteBread === 'brioche' ? 'bg-[#FF2400] text-white border-[#FF2400]' : 'bg-char text-bone border-white/10 hover:border-white/20'}"
                >
                  Pan Brioche Dorado
                </button>
              </div>
            </div>

            <!-- Notificaciones WhatsApp -->
            <div class="pt-2 flex items-center justify-between p-3.5 rounded-2xl bg-char border border-white/10">
              <div>
                <span class="text-xs font-bold text-bone block">Avisos de Comanda por WhatsApp</span>
                <span class="text-[11px] text-ash">Te escribimos cuando tu smash esté lista</span>
              </div>
              <button
                type="button"
                onclick={() => notifyWhatsApp = !notifyWhatsApp}
                class="w-12 h-7 rounded-full transition-colors cursor-pointer relative p-0.5 {notifyWhatsApp ? 'bg-[#FF2400]' : 'bg-white/10'}"
                aria-label="Alternar notificaciones WhatsApp"
              >
                <div class="w-6 h-6 rounded-full bg-white transition-transform {notifyWhatsApp ? 'translate-x-5' : 'translate-x-0'}"></div>
              </button>
            </div>
          </div>
        </div>

        <!-- SECCIÓN 6: CERRAR SESIÓN (FAT-FINGER) -->
        <div class="pt-2 pb-6">
          <button
            type="button"
            onclick={handleLogout}
            class="w-full min-h-[60px] sm:min-h-[64px] rounded-full bg-char-card hover:bg-char-card-hover text-ash hover:text-[#FF2400] active:scale-[0.97] border border-white/10 font-display font-extrabold text-base tracking-tight transition-all duration-150 flex items-center justify-center gap-2 cursor-pointer"
          >
            <LogOut class="w-5 h-5" />
            <span>Cerrar Sesión</span>
          </button>
        </div>

      {:else}
        <!-- ========================================================================= -->
        <!-- CASO B: USUARIO NO AUTENTICADO -> LOGIN / REGISTRO                       -->
        <!-- ========================================================================= -->
        <main class="space-y-6">
          <!-- Título y Subtítulo Fieles a la Referencia -->
          <div>
            {#if viewMode === 'register'}
              <h1 class="font-display font-extrabold text-3xl sm:text-4xl text-bone tracking-tight leading-none">
                Crear Cuenta
              </h1>
              <p class="text-sm text-ash mt-2 font-medium">
                Regístrate para continuar
              </p>
            {:else}
              <h1 class="font-display font-extrabold text-3xl sm:text-4xl text-bone tracking-tight leading-none">
                ¡Bienvenido de nuevo!
              </h1>
              <p class="text-sm text-ash mt-2 font-medium">
                Inicia sesión para continuar
              </p>
            {/if}
          </div>

          <!-- Alerta de Error (Diseño Plano) -->
          {#if errorMessage}
            <div class="p-4 rounded-2xl bg-[#FF2400]/10 border border-[#FF2400]/30 text-[#FF2400] text-sm flex items-center gap-3 mb-6 font-medium">
              <AlertCircle class="w-5 h-5 shrink-0 stroke-[2.5]" />
              <span>{errorMessage}</span>
            </div>
          {/if}

          <form onsubmit={(e) => { e.preventDefault(); handleLoginSubmit(); }} class="space-y-4">
            
            {#if viewMode === 'register'}
              <!-- Nombre Completo en Registro -->
              <div class="relative">
                <div class="absolute inset-y-0 left-0 pl-5 flex items-center pointer-events-none text-ash">
                  <User class="w-5 h-5 stroke-[2]" />
                </div>
                <input 
                  type="text"
                  bind:value={fullName}
                  placeholder="Tu Nombre"
                  class="w-full min-h-[58px] sm:min-h-[60px] rounded-full py-3.5 sm:py-4 pl-14 pr-5 bg-char-card border border-white/10 text-bone placeholder:text-ash/60 text-base font-sans focus:outline-none focus:border-[#FF2400]/60 transition-colors"
                  required
                />
              </div>
            {/if}

            <!-- Campo Correo (Cápsula Redonda Gorda con Icono) -->
            <div class="relative">
              <div class="absolute inset-y-0 left-0 pl-5 flex items-center pointer-events-none text-ash">
                <Mail class="w-5 h-5 stroke-[2]" />
              </div>
              <input 
                type="email"
                bind:value={email}
                placeholder="Tu Correo"
                class="w-full min-h-[58px] sm:min-h-[60px] rounded-full py-3.5 sm:py-4 pl-14 pr-5 bg-char-card border border-white/10 text-bone placeholder:text-ash/60 text-base font-sans focus:outline-none focus:border-[#FF2400]/60 transition-colors"
                required
              />
            </div>

            <!-- Campo Contraseña (Cápsula Redonda Gorda con Icono y Toggle) -->
            <div class="relative">
              <div class="absolute inset-y-0 left-0 pl-5 flex items-center pointer-events-none text-ash">
                <Lock class="w-5 h-5 stroke-[2]" />
              </div>
              <input 
                type={showPassword ? 'text' : 'password'}
                bind:value={password}
                placeholder="Tu Contraseña"
                class="w-full min-h-[58px] sm:min-h-[60px] rounded-full py-3.5 sm:py-4 pl-14 pr-14 bg-char-card border border-white/10 text-bone placeholder:text-ash/60 text-base font-sans focus:outline-none focus:border-[#FF2400]/60 transition-colors"
                required
              />
              <button 
                type="button" 
                onclick={() => showPassword = !showPassword}
                class="absolute inset-y-0 right-0 pr-5 flex items-center text-ash hover:text-bone cursor-pointer"
                aria-label="Ver u ocultar contraseña"
              >
                {#if showPassword}
                  <EyeOff class="w-5 h-5" />
                {:else}
                  <Eye class="w-5 h-5" />
                {/if}
              </button>
            </div>

            <!-- Fila Opciones: Recordarme / Olvidaste Contraseña -->
            <div class="flex items-center justify-between px-2 pt-1 pb-1">
              <label class="flex items-center gap-3 cursor-pointer select-none py-1">
                <input 
                  type="checkbox" 
                  bind:checked={rememberMe}
                  class="w-5 h-5 rounded-md border-white/20 text-[#FF2400] focus:ring-0 focus:ring-offset-0 bg-char-card cursor-pointer"
                />
                <span class="text-sm font-medium text-bone">
                  Recordarme
                </span>
              </label>

              {#if viewMode === 'login'}
                <button 
                  type="button" 
                  onclick={() => alert('Para restablecer tu contraseña puedes ingresar directamente con Google o solicitar asistencia en mostrador.')}
                  class="text-sm font-medium text-ash hover:text-[#FF2400] transition-colors py-1 cursor-pointer"
                >
                  ¿Olvidaste tu contraseña?
                </button>
              {/if}
            </div>

            <!-- BOTÓN PRINCIPAL 1: Iniciar Sesión (Rojo Vivo #FF2400) -->
            <div class="pt-2">
              <button
                type="submit"
                disabled={isLoading}
                class="w-full min-h-[60px] sm:min-h-[64px] rounded-full bg-[#FF2400] hover:bg-[#E02000] active:scale-[0.97] text-white font-display font-extrabold text-lg sm:text-xl tracking-tight transition-all duration-150 flex items-center justify-center gap-2 cursor-pointer disabled:opacity-60"
              >
                {#if isLoading}
                  <Loader2 class="w-5 h-5 animate-spin" />
                  <span>Procesando...</span>
                {:else}
                  <span>{viewMode === 'register' ? 'Registrarse' : 'Iniciar Sesión'}</span>
                {/if}
              </button>
            </div>

            <!-- BOTÓN PRINCIPAL 2: Píldora Gorda para Google justo abajo -->
            <button
              type="button"
              onclick={handleGoogleClick}
              disabled={isLoading}
              class="w-full min-h-[60px] sm:min-h-[64px] rounded-full bg-char-card hover:bg-char-card-hover active:scale-[0.97] text-bone font-display font-extrabold text-base sm:text-lg tracking-tight border border-white/10 transition-all duration-150 flex items-center justify-center gap-3 cursor-pointer disabled:opacity-60"
            >
              <svg class="w-5 h-5 shrink-0" viewBox="0 0 24 24">
                <path fill="#4285F4" d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z"/>
                <path fill="#34A853" d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z"/>
                <path fill="#FBBC05" d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.06H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.94l2.85-2.22.81-.63z"/>
                <path fill="#EA4335" d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.06l3.66 2.84c.87-2.6 3.3-4.52 6.16-4.52z"/>
              </svg>
              <span>{viewMode === 'register' ? 'Registrarse con Google' : 'Continuar con Google'}</span>
            </button>
          </form>

          <!-- Pie de Página: Alternar entre Iniciar Sesión y Registrarse -->
          <footer class="pt-4 pb-1 text-center shrink-0">
            {#if viewMode === 'login'}
              <p class="text-sm text-ash font-medium">
                ¿No tienes una cuenta aún?
                <button 
                  type="button" 
                  onclick={() => setMode('register')}
                  class="font-bold text-bone hover:text-[#FF2400] ml-1 transition-colors underline underline-offset-4 cursor-pointer"
                >
                  Regístrate
                </button>
              </p>
            {:else if viewMode === 'register'}
              <p class="text-sm text-ash font-medium">
                ¿Ya tienes una cuenta?
                <button 
                  type="button" 
                  onclick={() => setMode('login')}
                  class="font-bold text-bone hover:text-[#FF2400] ml-1 transition-colors underline underline-offset-4 cursor-pointer"
                >
                  Inicia Sesión
                </button>
              </p>
            {/if}
          </footer>
        </main>
      {/if}

    </div>
  </div>
{/if}
