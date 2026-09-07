# Reglas Normativas de Ingeniería Frontend

Basado en `.standard/15. FRONTEND ENGINEERING.md` (v2.1) y `.standard/09. CONCIENCIA DEL PROYECTO (PAP)`.

## 1. Reutilización y Alcance Mínimo (CRÍTICA)
- **FE-001 / PAP-002 (Búsqueda previa obligatoria):** Antes de crear cualquier componente, helper, store, estilo o función, explorar la base de código existente. La creación desde cero es el último recurso.
- **FE-002 (Preservación de identidad):** El código debe ser indistinguible de la base existente. Respetar la sintaxis nativa de Svelte 5 (runes: `$state`, `$derived`, `$props`), las clases utilitarias de Tailwind configuradas en el proyecto y los primitivos de `bits-ui`.
- **FE-005 / PAP-003 (Alcance mínimo):** Modificar estrictamente lo necesario para cumplir el requerimiento. Queda prohibido refactorizar módulos no solicitados, reorganizar carpetas o introducir capas arquitectónicas colaterales.
- **FE-011 (Sin suposiciones de incompletitud):** La ausencia de evidencia inicial no justifica declarar que una función falta; se debe agotar la inspección antes de recrear código.

## 2. Estructura y Modularidad
- **FE-003 (División por responsabilidad, no por líneas):** Un componente se separa solo cuando acumula múltiples responsabilidades desacoplables, nunca por un límite arbitrario de líneas.
- **FE-004 (Evitar componentes espejo):** Si dos piezas comparten más del 70% de su lógica o presentación, deben unificarse mediante parametrización o slots/snippets.
- **FE-007 (Función vs Componente):** Si una necesidad se resuelve con una función pura de pocas líneas, no encapsularla en un componente ni en un store.
- **FE-012 (Cero carpetas y archivos ceremoniales):** Prohibido dividir componentes pequeños en múltiples archivos satélite (`.types`, `.styles`, `.constants`, `.utils`) sin una complejidad de dominio que lo exija.
- **FE-015 (Proximidad del código):** Los elementos que evolucionan juntos deben permanecer en la misma carpeta o archivo para reducir la fricción cognitiva.

## 3. Estado, Reactividad y Nomenclatura
- **FE-008 (Cero abstracciones innecesarias):** No crear custom stores o helpers para lógica de 3 líneas que pertenece a un único componente.
- **FE-009 (Uso estricto de reactividad y efectos):** Reservar los efectos (`$effect`) exclusivamente para interacción con APIs del navegador o sincronización externa. El cálculo de datos se realiza mediante derivación reactiva pura (`$derived`).
- **FE-010 (Cero optimizaciones prematuras):** No introducir memorizaciones complejas sin métricas concretas de cuello de botella.
- **FE-013 (Nomenclatura con intención):** Prohibidos nombres comodín sin dominio (*Manager*, *Helper*, *Common*, *Generic*, *Shared*). Toda función o archivo debe expresar directamente su propósito en el negocio.
- **FE-016 (Cero arquitectura ceremonial):** Prohibido importar patrones hipercomplejos de otros frameworks o tutoriales que añadan indirecciones innecesarias al proyecto.
