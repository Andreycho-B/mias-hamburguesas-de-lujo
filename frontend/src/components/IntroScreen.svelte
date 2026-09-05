<script>
  import { onMount } from 'svelte';
  import gsap from 'gsap';

  let { onComplete } = $props();
  let introContainer = $state(null);
  let isVisible = $state(true);

  onMount(() => {
    // 1. Validar si ya se mostró en los últimos 60 minutos (3600 segundos)
    const seenAt = localStorage.getItem('mias_intro_seen_at');
    if (seenAt) {
      const diffSeconds = (Date.now() - parseInt(seenAt, 10)) / 1000;
      if (diffSeconds < 3600) {
        isVisible = false;
        onComplete();
        return;
      }
    }

    // 2. Animación de entrada suave con GSAP
    const tl = gsap.timeline({
      onComplete: () => {
        finishIntro();
      }
    });

    tl.fromTo('.intro-logo', 
      { opacity: 0, scale: 0.9, y: 15 }, 
      { opacity: 1, scale: 1, y: 0, duration: 1.0, ease: 'power3.out' }
    )
    .fromTo('.intro-title', 
      { opacity: 0, y: 20 }, 
      { opacity: 1, y: 0, duration: 0.9, ease: 'power3.out' }, 
      '-=0.6'
    )
    .fromTo('.intro-subtitle', 
      { opacity: 0, y: 10 }, 
      { opacity: 1, y: 0, duration: 0.8, ease: 'power2.out' }, 
      '-=0.5'
    )
    .fromTo('.intro-dismiss', 
      { opacity: 0 }, 
      { opacity: 0.6, duration: 0.6 }, 
      '-=0.3'
    )
    // Esperar hasta completar 3.5 segundos en total antes del desvanecimiento
    .to({}, { duration: 0.6 })
    .to(introContainer, {
      opacity: 0,
      scale: 1.02,
      duration: 0.6,
      ease: 'power2.inOut'
    });
  });

  function finishIntro() {
    if (!isVisible) return;
    localStorage.setItem('mias_intro_seen_at', Date.now().toString());
    gsap.to(introContainer, {
      opacity: 0,
      duration: 0.35,
      ease: 'power2.inOut',
      onComplete: () => {
        isVisible = false;
        onComplete();
      }
    });
  }
</script>

{#if isVisible}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <!-- svelte-ignore a11y_no_static_element_interactions -->
  <div 
    bind:this={introContainer}
    onclick={finishIntro}
    class="fixed inset-0 z-50 flex flex-col items-center justify-center bg-cream text-char cursor-pointer select-none p-6 text-center transition-all overflow-hidden"
  >
    <!-- Isotipo del Candado de Corazón -->
    <div class="intro-logo mb-6">
      <img 
        src="/imagenes/marca/isotipo.webp" 
        alt="Mías Candado de Corazón" 
        class="w-20 h-20 mx-auto drop-shadow-md"
      />
    </div>

    <!-- Titular Monumental en Serif de Lujo -->
    <h1 class="intro-title font-serif text-5xl sm:text-7xl font-bold tracking-tight text-char uppercase mb-2">
      BIENVENIDO
    </h1>

    <!-- Subtítulo en caja baja y Plus Jakarta Sans -->
    <p class="intro-subtitle font-sans text-base sm:text-lg text-char/80 tracking-wide font-normal mb-8">
      a mías hamburguesas de lujo
    </p>

    <!-- Línea Divisoria Hairline Suiza -->
    <div class="w-16 h-[1px] bg-char/20 mx-auto mb-12"></div>

    <!-- Texto Discreto para Continuar -->
    <p class="intro-dismiss font-sans text-xs uppercase tracking-widest text-char/50 animate-pulse">
      da clic en cualquier parte para continuar
    </p>
  </div>
{/if}
