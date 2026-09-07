# Estándar Global de Curvatura y Tarjetas para Mías Hamburguesas de Lujo

Este documento establece la regla obligatoria de diseño estético para **todas las tarjetas, modales y elementos con esquinas redondeadas** del proyecto:

## 1. Curvatura Universal de Esquinas (Estándar Oficial)
Para todas las tarjetas, modales y contenedores destacados con esquinas redondeadas:
- **Móvil (`< 640px`)**: **`rounded-[42px]`** (disponible también como clase de Tailwind **`rounded-card`**).
  - *Fundamento:* Sigue la curvatura concéntrica continua (*squircle*) de la pantalla del iPhone con un marco perimetral simétrico de 12px.
- **Escritorio / Tablet (`sm:`, `md:`)**: **`sm:rounded-[48px]`** (disponible también como **`sm:rounded-card-lg`**).

## 2. Paleta de Color: Crema Luminoso y Refinado
- **Fondo de tarjeta**: `#FAF6EE` (o `#FAF6ED`): blanco marfil/crema cálido luminoso, limpio y editorial.
- **Borde perimetral**: `border border-black/5` para un contorno ultra nítido.
- **Sombra profunda**: `shadow-[0_25px_60px_-15px_rgba(0,0,0,0.6)]`.

## 3. Márgenes y Encuadre en Móvil
- **Encuadre Inner Screen**: Margen uniforme exacto de **12px** en los cuatro costados (`p-[12px]`).
- Sin franjas negras asimétricas en la parte superior o inferior.
- **Escritorio**: Centrado flotante con `sm:p-5 md:p-6 sm:max-w-2xl sm:max-h-[88vh]`.

## 4. Botones de Acción y Cierre
- En tarjetas con esquinas de 42px, los controles superiores se colocan a `top-5 right-5` (`20px`) para guardar una separación armónica con el arco de la esquina.
