# Reglas Normativas de Diseño Visual y Anti-Patrones

Basado en `.standard/14. ANTI-PATRONES DE DISEÑO VISUAL.md` (v2.2), `RFC-2026-003`, `docs/05_BRAND_SYSTEM.md` y `.agents/rules/apple_cards_standard.md`.

## 1. Detección y Prohibición de Firmas Visuales de IA (Familia N - CRÍTICA)
Está prohibido adoptar por inercia las siguientes estéticas generalizadas en generadores de UI (v0, Lovable, Bolt, Shadcn):
- **Estética "Linear":** Contenedor oscuro flotando sobre lienzo claro con un solo acento neón, sin justificación de marca.
- **Estética "DevTools / Terminal":** Uso decorativo de tipografía mono en textos no técnicos o badges verdes tipo salida de consola fuera de contextos DevOps.
- **Bento Grid por defecto:** Cuadrículas simétricas de cajas con esquinas ultra-redondeadas para agrupar cualquier dato sin relación conceptual.
- **Estética "v0 / Shadcn":** Paleta monocromática Zinc/Slate repetitiva con fuentes Inter/Geist por automatismo de entrenamiento.
- **Badges tipo CI/CD:** Píldoras con puntos pulsantes verdes o checks de pipeline para estados ordinarios del negocio.
- **Brutalist Display Extrudido:** Tipografías pesadas desmedidas (`Monument`, `Clash Display`) sin coherencia editorial con la marca.

## 2. Principios de Composición y Jerarquía
- **Estructura nacida del contenido (Familia A):** Prohibido el layout universal predecible de landing SaaS (Hero → Logos → 3 Cards → Métricas → Testimonios → CTA). El contenido real dicta la disposición.
- **Justificación funcional de elementos (Familia B):** Cero adornos superfluos; prohibidos gradientes sintéticos morado-azul, aurora backgrounds, noise, blobs y acumulación de efectos (borde + sombra + blur + glow simultáneos).
- **Veracidad de datos (Familia C):** Prohibido inventar métricas decorativas (500+, 99%), testimonios ficticios o logotipos sin base de clientes real.
- **Uso justificado de tarjetas (Familia D):** No encapsular párrafos ni listas continuas en tarjetas individuales sin separación conceptual.
- **Tipografía con jerarquía semántica (Familia F):** Prohibido alternar cursiva en títulos para fingir sofisticación, enfatizar una palabra fija por párrafo o fragmentar oraciones artificialmente.
- **Diseño para interacción real (Familia G):** Sin animaciones, sliders o marquesinas decorativas continuas que entorpezcan la usabilidad.
- **Feedback y estados completos (Familia H):** Respuestas diferenciadas de hover/active según el elemento (no aplicar `scale-105` indiscriminadamente). Estados vacíos, de carga y de error con contexto claro y acciones directas.

## 3. Identidad de Marca Obligatoria del Proyecto (`docs/05_BRAND_SYSTEM.md`)
- **Estilo:** Diseño editorial gastronómico suizo (retícula de 4 columnas en móvil, líneas hairlines divisorias de 1px con `rgba(0,0,0,0.08)` o `rgba(255,255,255,0.08)`).
- **Paleta oficial:**
  - `cream`: `#F8F5EE` (o `#FAF6EE` en tarjetas marfil).
  - `bone`: `#EDE8DC`.
  - `char`: `#0C0C0E` (carbón profundo).
  - `char-card`: `#161619` (carbón elevado).
  - `scarlet` / `ember`: `#C1121F` (rojo candado y horno Rayo).
  - `gold`: `#C5A059` (latón pulido de lujo).
  - `ash`: `#8E8E93` (metadatos y textos técnicos).
- **Tipografías:**
  - Títulos: `Cormorant Garamond` / `Playfair Display`.
  - UI y lectura: `Plus Jakarta Sans` / `Outfit`.
  - Comandas y tickets: `JetBrains Mono`.
- **Curvatura Concave/Squircle (`.agents/rules/apple_cards_standard.md`):**
  - Móvil (`< 640px`): `rounded-[42px]` (`rounded-card`) con marco perimetral de 12px.
  - Tablet/Escritorio: `sm:rounded-[48px]` (`sm:rounded-card-lg`).
  - Fondo de tarjeta marfil: `#FAF6EE` con borde `border-black/5` y sombra profunda.
