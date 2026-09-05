<script>
  import { store } from '../lib/store.svelte.js';
  import { Heart, Flame, Sparkles } from 'lucide-svelte';
</script>

<div class="max-w-4xl mx-auto px-4 py-8 space-y-12">
  <!-- Cabecera Editorial -->
  <div class="text-center space-y-4 max-w-2xl mx-auto">
    <span class="text-xs font-mono uppercase tracking-widest text-gold flex items-center justify-center gap-2">
      <Sparkles class="w-3.5 h-3.5" /> Amor para toda la vida
    </span>
    <h2 class="font-serif text-4xl sm:text-6xl font-bold text-bone tracking-tight">
      Lo Íntimo Hecho para Compartir
    </h2>
    <div class="w-16 h-[1px] bg-gold/40 mx-auto"></div>
  </div>

  <!-- Ensayo y Narrativa Original de la Marca -->
  <div class="p-8 sm:p-12 rounded-3xl bg-char-card border hairline-dark space-y-6 shadow-lux">
    <p class="font-serif text-xl sm:text-2xl text-bone/90 leading-relaxed italic text-center">
      "Mías nació en el mismo lugar donde comenzó mi historia; una casa llena de amor donde crecí alrededor del carbón, compartiendo la mesa, las recetas y los momentos más valiosos con mi familia."
    </p>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-8 pt-6 border-t hairline-dark items-center text-xs sm:text-sm text-ash leading-relaxed font-sans">
      <p>
        Fue allí donde entendí que la comida también puede ser una forma de expresar cariño. Hoy regreso a ese mismo lugar para convertir esos recuerdos en hamburguesas al carbón, preparadas con ingredientes de calidad, dedicación y mucho corazón.
      </p>
      <p>
        Cada hamburguesa que sale de nuestra cocina lleva un pedazo de esa historia: el esfuerzo de una familia, de los sueños y la certeza de que todo lo que se hace con amor deja huella.
      </p>
    </div>
  </div>

  <!-- Galería Fotográfica Real del Restaurante -->
  <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
    <!-- 1. El Horno Rayo al Carbón -->
    <div class="rounded-2xl overflow-hidden border hairline-dark bg-char-card group hover:border-gold/50 transition-all">
      <div class="h-64 overflow-hidden relative">
        <img 
          src="/imagenes/historia/horno_a_carbon.webp" 
          alt="Horno al Carbón Rayo" 
          class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
        />
        <div class="absolute inset-0 bg-gradient-to-t from-char to-transparent"></div>
        <span class="absolute bottom-3 left-3 text-xs font-mono text-gold font-bold">
          Horno a Carbón «Rayo»
        </span>
      </div>
      <div class="p-4">
        <p class="text-xs text-ash">
          Acabado en rojo escarlata y bronce, el corazón de nuestro sabor ahumado inconfundible.
        </p>
      </div>
    </div>

    <!-- 2. El Ritual del Candado de Amor -->
    <div class="rounded-2xl overflow-hidden border hairline-dark bg-char-card group hover:border-gold/50 transition-all">
      <div class="h-64 overflow-hidden relative">
        <img 
          src="/imagenes/historia/candado_branding_historia.webp" 
          alt="Candado de Corazón en la Reja" 
          class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
        />
        <div class="absolute inset-0 bg-gradient-to-t from-char to-transparent"></div>
        <span class="absolute bottom-3 left-3 text-xs font-mono text-gold font-bold">
          El Candado de Amor
        </span>
      </div>
      <div class="p-4">
        <p class="text-xs text-ash">
          El símbolo que une a nuestros visitantes: deja tu candado sellado en la reja del local.
        </p>
      </div>
    </div>

    <!-- 3. El Salón de Lámparas de Cristal -->
    <div class="rounded-2xl overflow-hidden border hairline-dark bg-char-card group hover:border-gold/50 transition-all sm:col-span-2 lg:col-span-1">
      <div class="h-64 overflow-hidden relative">
        <img 
          src="/imagenes/local/restaurante_mias_desde_dentro.webp" 
          alt="Salón Restaurante Mías" 
          class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
        />
        <div class="absolute inset-0 bg-gradient-to-t from-char to-transparent"></div>
        <span class="absolute bottom-3 left-3 text-xs font-mono text-gold font-bold">
          El Salón Íntimo
        </span>
      </div>
      <div class="p-4">
        <p class="text-xs text-ash">
          Sillas de terciopelo escarlata y candelabros de cristal en un ambiente cálido y boutique.
        </p>
      </div>
    </div>
  </div>

  <!-- Botón de Retorno al Menú -->
  <div class="text-center pt-4">
    <button
      onclick={() => store.currentView = 'menu'}
      class="py-3 px-8 rounded-full bg-gold hover:bg-gold-light text-char font-mono font-bold text-xs uppercase tracking-wider transition-colors shadow-gold"
    >
      Volver a la Carta Gastronómica →
    </button>
  </div>
</div>
