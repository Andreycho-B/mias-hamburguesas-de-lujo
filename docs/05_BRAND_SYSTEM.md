# 05. Sistema de Diseño e Identidad Visual (BRAND SYSTEM)

**Proyecto:** Mías Hamburguesas de Lujo  
**Versión:** 1.0.0  
**Estado:** Normativo / Aprobado  

---

## 1. Paleta de Colores y Tokens de Diseño

La paleta se inspira en el espacio físico real del restaurante (sillas de terciopelo escarlata, lámparas de cristal con luz ámbar cálida, paredes de carbón ahumado y el horno de latón «Rayo»):

| Token Tailwind | Nombre Semántico | Valor Hex / RGB | Aplicación Principal |
| :--- | :--- | :--- | :--- |
| `cream` | Crema Suave | `#F8F5EE` | Fondo principal de la Bienvenida, tarjetas de contraste y lino gastronómico. |
| `bone` | Hueso / Lino | `#EDE8DC` | Fondos secundarios, divisores suaves y etiquetas sutiles. |
| `char` | Carbón Profundo | `#0C0C0E` | Fondo de alto impacto, modo oscuro editorial y texto de alto contraste. |
| `char-card` | Carbón Elevado | `#161619` | Superficies de tarjetas oscuras y modales de caja. |
| `ember` / `scarlet` | Rojo Escarlata | `#C1121F` | El color del candado oficial y el horno Rayo. Botones primarios, badges de alerta y números de énfasis. |
| `gold` | Latón / Oro Pulido | `#C5A059` | Acentos de lujo, insignias de precio, remaches visuales y cartilla de sellos. |
| `ash` | Ceniza | `#8E8E93` | Textos secundarios de ingredientes, metadatos y notas técnicas. |

---

## 2. Tipografía y Jerarquía

* **Titulares Editoriales de Lujo (Headings)**:
  * Familia: `Cormorant Garamond` / `Playfair Display` (Google Fonts).
  * Aplicación: *"BIENVENIDO"*, nombres de hamburguesas (*Dorada*, *Suprema*, *Brava*), y títulos de sección (*Lo Íntimo Hecho para Compartir*).
* **Cuerpo de Texto y Controles de Interfaz (UI & Body)**:
  * Familia: `Plus Jakarta Sans` / `Outfit` (Google Fonts).
  * Aplicación: Descripciones de ingredientes, precios, botones táctiles de 44px, formularios del wizard y etiquetas de estado.
* **Comandas y Tickets Numéricos**:
  * Familia: `JetBrains Mono` / `Space Mono` (Google Fonts / Monospace del sistema).
  * Aplicación: Códigos `ORD-A101`, `DOM-B201`, fecha/hora de comanda y formato de impresión térmica.

---

## 3. Retícula Suiza y Reglas Anti-AI Slop

1. **Retícula Base Móvil**: 4 columnas con márgenes de `16px`, medianiles de `12px` y saltos rítmicos de `8px`.
2. **Líneas Divisoras (Hairlines Continuos)**: Separadores estructurales de `1px` en `rgba(0, 0, 0, 0.08)` sobre fondo claro y `rgba(255, 255, 255, 0.08)` sobre fondo oscuro.
3. **Prohibición de Estilos Genéricos de IA**:
   * Cero sombras gigantescas difusas (`box-shadow: 0 25px 50px -12px rgba(0,0,0,0.25)`). En su lugar se usan bordes de precisión de 1px con sombras de contacto rasantes (`0 2px 8px rgba(0,0,0,0.06)`).
   * Cero degradados sintéticos morado-azul.
   * Cero ilustraciones abstractas de relleno: solo fotografía real de los activos del local.

---

## 4. Mapeo Oficial de Activos (`imagenes/`)

* **Bienvenida (Intro)**: `marca/isotipo.webp` sobre fondo `cream` suave con tipografía *"BIENVENIDO"*.
* **Hamburguesas**: PNGs transparentes de alta resolución (`hamburguesas/Golden.png`, `Criolla.png`, `Crispy.png`, `Brava.png`, `Suprema.png`, `Orignal.png`).
* **Perros y Especialidades**: `perros/perro.png`, `de_la_casa/salchi_papa.png`.
* **Bebidas**: `bebidas/cocacola.webp`, `bretana-vidrio.png`, `agua_con_gas.webp`.
* **Fidelización y Sellos**: `historia/candado_branding_historia.webp` como fotografía de cabecera de la cartilla.
* **Historia**: `local/restaurante_mias_desde_dentro.webp`, `historia/horno_a_carbon.webp`, `fundadores/fundadores.webp`.
