# 01. Documento de Requisitos de Producto (PRD)

**Proyecto:** Mías.8 Hamburguesas de Lujo  
**Versión:** 1.0.0  
**Estado:** Normativo / Aprobado  

---

## 1. Alcance y Objetivos

Diseñar y construir una aplicación web progresiva (PWA) de nivel senior que unifique la experiencia gastronómica de los clientes de **Mías.8 Hamburguesas de Lujo**, agilice la toma y cobro de comandas en sala, canalice pedidos a domicilio estructurados vía WhatsApp y opere un programa de fidelización digital blindado contra fraudes.

---

## 2. Requisitos Funcionales (FR)

### FR-001: Pantalla de Bienvenida (Intro Screen)
* **Visual:** Fondo color crema suave (`#F8F5EE`), isotipo del candado de corazón rojo (`marca/isotipo.webp`), titular en Serif de lujo *"BIENVENIDO"*, subtítulo *"a mías hamburguesas de lujo"* y línea divisoria de 1px.
* **Comportamiento:** Duración de 3.5 segundos con salida cinemática suave (fade + scale) o interrupción inmediata al pulsar en cualquier parte ("clica en cualquier parte para continuar").
* **Persistencia:** Almacena timestamp en `localStorage` (`mias_intro_seen_at`). Si el usuario vuelve a entrar antes de 60 minutos, la intro se omite síncronamente y entra directo al Home.

### FR-002: Catálogo Gastronómico y Retícula Suiza
* **Categorías:** `[01] HAMBURGUESAS`, `[02] PERROS CALIENTES`, `[03] DE LA CASA`, `[04] TEMPORADA` (destacada).
* **Especificaciones Estrictas de Pan:**
  * Hamburguesas: Exclusivamente en **Pan Pretzel 160g**.
  * Perros Calientes: Exclusivamente en **Pan Brioche Artesanal**.
* **Ficha de Producto (Pantalla Completa):**
  * Al tocar un producto, se expande una vista completa con fotografía recortada en alta definición.
  * Selector de acompañamientos en **tarjetas cuadradas compactas** (ej. Combo Papa Francesa + Gaseosa por $11.000, Papa Nativa por $9.000, Tocineta por $5.000).
  * Botón inferior tipo cápsula anclado en la zona del pulgar: *"Pedir y pagar • $XX.XXX"*.

### FR-003: Carrito Flotante No Intrusivo
* **Disparador:** Botón circular flotante de 56×56 px en la esquina inferior derecha con badge numérico escarlata.
* **Drawer Inferior (Bits UI):** Al pulsar el círculo, se despliega un cajón accesible desde abajo con la lista de productos, controles de incremento y decremento `[ - 1 + ]` de 44px de área táctil, cálculo de subtotal y botón *"Continuar con la orden"*.

### FR-004: Habeas Data y Autenticación Cero-Pérdidas
* **Consentimiento Ley 1581:** Banner inferior antes de iniciar sesión que expone la política de tratamiento de datos con botones claros: *Aceptar y Continuar* o *Rechazar*.
* **Google OAuth en 1 Clic:** Autenticación rápida para clientes.
* **Migración de Carrito:** Si un cliente anónimo configuró productos en su carrito antes de identificarse, el sistema migra y asocia ese carrito a su nueva cuenta de usuario en base de datos sin pérdida de datos.

### FR-005: Selector de Modalidad de Pedido
* **Modalidad A: Consumo en el Punto (`ORD-XXXX`)**:
  * Asigna identificador alfanumérico secuencial (ej. `ORD-A101`).
  * Envía la comanda al panel del cajero en estado **`POR COBRAR`** (tarjeta roja + alerta sonora).
  * El cajero cobra presencialmente, entrega la factura física correlativa (ej. `#00124`) y un localizador físico para la mesa (acrílico/madera).
  * El cajero pulsa **`ORDEN PAGA`** (verde), habilitando la comanda para solicitud de sellos.
* **Modalidad B: Domicilio (`DOM-XXXX`)**:
  * **Tarifa Base:** Suma automáticamente **$6.000 COP** de envío al subtotal.
  * **Wizard Conversacional:**
    1. Paso 1: Nombre completo y teléfono móvil.
    2. Paso 2: Dirección exacta, barrio e indicaciones para el domiciliario.
    3. Paso 3: Método de pago (Efectivo con indicación de billete para cambio exacto, Nequi o Transferencia Bre-B).
  * **Despacho:** Registra `DOM-B201` en estado **`EN PREPARACIÓN`** en el panel de caja y abre automáticamente WhatsApp (`321 252 0881`) con el mensaje pre-estructurado.
  * **Trazabilidad:** Cajero pasa a **`EN CAMINO`**. Cliente marca **`Recibido`** y califica de 1 a 5 estrellas con comentario.

### FR-006: Programa de Fidelización (Sellos Digitales)
* **Regla:** 1 sello por cada hamburguesa pedida en una orden web pagada.
* **Solicitud:** El cliente va a "Mis Sellos", selecciona su orden paga e ingresa el número de la factura física.
* **Validación de Personal:**
  * Solicitud visible en tiempo real para Meseros y Cajero vía SSE.
  * Mesero o Cajero revisan y pulsan **Aprobar** o **Rechazar**.
  * Al aprobar, la base de datos bloquea el número de factura con restricción `UNIQUE` (impide reutilización) y registra los sellos en el `stamp_ledger`.
* **Límites Temporales:**
  * Plazo de solicitud: Máximo 24 horas desde la orden.
  * Tiempo de revisión: 60 minutos (auto-cancelación si no se atiende, permitiendo reintentar).
  * Caducidad de sellos acumulados: 6 meses desde la última compra registrada.

### FR-007: Redención de Hamburguesa Original Gratis
* **Capacidad:** Cartilla de 10 sellos.
* **Regla de Congelación:** Al completar 10 sellos, el cliente no acumula sellos adicionales en compras posteriores hasta que redima su hamburguesa pendiente.
* **Redención Presencial:** Cliente pulsa *"Redimir Hamburguesa Original"* (muestra estado: *"A la espera de aprobación en caja"*).
* **Adiciones:** Si el cliente añade extras o combo, el sistema descuenta los $17.000 base de la Original y solo cobra el excedente en caja.
* **Aprobación:** Cajero valida presencialmente y pulsa "Aprobar" en 1 clic. Salta animación de felicitación (GSAP) en el teléfono del cliente y se desbloquea la cartilla a 0/10.

### FR-008: Alertas Sonoras (Web Audio API)
* Síntesis nativa de un chime armónico sin latencia que suena en el panel de caja y meseros cada vez que entra un pedido o una solicitud de sello.

### FR-009: Impresión Térmica de Comandas
* Plantilla CSS `@media print` en el panel de caja optimizada para rollos de 80mm y 58mm (fuente monoespaciada, código de orden, tabla de productos, totales y notas).

### FR-010: PWA y Notificaciones Web Push
* Configuración completa de PWA instalable (`manifest.webmanifest`, `display: standalone`).
* Web Push con VAPID: Notificación amistosa cuando el cliente acumula 8 o 9 sellos para incentivar la visita final.

### FR-011: Panel Ejecutivo de Administración (Admin Dashboard)
* **Métricas BI:** Top ventas (volumen y facturación), Top calificaciones de 1 a 5 estrellas, Clientes a punto de redimir, Ventas punto vs domicilio.
* **CMS:** Interruptor *"En Stock / Agotado"* en 1 clic por producto, gestión de categorías, precios y temporada.
* **Gestión de Personal y Auditoría:** Creación de cuentas (mesero/cajero) y libro inmutable de auditoría forense (`admin_audit_logs`).

---

## 3. Requisitos No Funcionales (NFR)

* **NFR-001 (Inmunidad SQL):** 100% de consultas compiladas con SQLC en Go. Cero concatenación de cadenas en SQL.
* **NFR-002 (Cero Inconsistencias):** Bloqueo pesimista `SELECT ... FOR UPDATE` en PostgreSQL para transacciones de sellos y redenciones.
* **NFR-003 (Degradación Controlada):** Catálogo y generación de pedidos a WhatsApp funcionan incluso si el servidor Go o PostgreSQL están caídos.
* **NFR-004 (Rendimiento Web):** Cero CLS (Cumulative Layout Shift = 0), LCP < 1.5s en conexiones 4G móviles.
* **NFR-005 (Seguridad de Sesiones):** Access Token en memoria volátil + Refresh Token en cookie `HttpOnly; Secure; SameSite=Strict`.
* **NFR-006 (Accesibilidad):** Cumplimiento WCAG 2.1 AA, áreas de toque mínimas de 44×44 px.
