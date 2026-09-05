# 00. Contexto del Proyecto: Mías Hamburguesas de Lujo

**Versión:** 1.0.0  
**Estado:** Normativo / Aprobado  
**Fecha de Creación:** 2026-09-05  

---

## 1. Visión y Propósito del Negocio

**Mías Hamburguesas de Lujo** es una propuesta gastronómica boutique de alta gama centrada en hamburguesas artesanales al carbón preparadas en horno «Rayo» (acabado en rojo escarlata y latón dorado). El restaurante combina un ambiente íntimo y sofisticado (*"Lo Íntimo Hecho para Compartir"*), con sillas de terciopelo escarlata, iluminación cálida con candelabros de cristal y la tradición simbólica del **candado de amor** (los comensales sellan su vínculo dejando un candado en la reja del local).

El objetivo de esta plataforma web es doble:
1. **Fidelización de Alta Gama**: Crear un programa de sellos digitales transparente, ultra-seguro y sin fricción (10 sellos = 1 hamburguesa Original gratis).
2. **Gestión Ágil de Comandas**: Facilitar tanto la toma de pedidos presenciales en el punto como el despacho estructurado de domicilios coordinados vía WhatsApp (`321 252 0881`) y reflejados en el panel de caja.

---

## 2. Stakeholders y Roles del Sistema

| Rol | Actor / Usuario | Responsabilidades y Objetivos |
| :--- | :--- | :--- |
| **Cliente** | Comensal en punto o domicilio | Consulta la carta, arma pedidos, despacha a WhatsApp, acumula sellos mediante su factura física y redime hamburguesas gratuitas. |
| **Mesero** | Personal de sala | Atiende las mesas guiado por localizadores físicos (números de madera/acrílico), promueve el programa de fidelización y valida/aprueba solicitudes de sellos en tiempo real desde su móvil. |
| **Cajero** | Personal de caja / despacho | Gestiona cobros presenciales (`Por cobrar` ➔ `Orden paga`), imprime comandas térmicas (80mm/58mm), despacha pedidos a domicilio (`En preparación` ➔ `En camino`), aprueba sellos y valida redenciones presenciales de hamburguesas gratis. |
| **Admin** | Propietario / Gerente | Control total del CMS (productos, precios, disponibilidad, temporada), visualización de métricas BI (top ventas, mejores calificaciones, retención de clientes) y auditoría forense inmutable. |

---

## 3. Identidad de Marca y Ubicación

* **Lema Oficial:** *"Lo Íntimo Hecho para Compartir"* / *"Amor para toda la vida"*.
* **Isotipo:** Candado de latón en forma de corazón con cerradura central (`marca/isotipo.webp`).
* **Canal Oficial de Domicilios:** WhatsApp Business `+57 321 252 0881`.
* **Horario de Servicio:** Sábado a Jueves de 3:00 PM a 9:30 PM (Viernes cerrado).
* **Dirección Estética:** Diseño editorial suizo gastronómico (retículas milimétricas, líneas hairlines continuas, paleta carbón profundo, crema suave, rojo escarlata y latón dorado). Prohibido el uso de estética genérica de IA (AI Slop).

---

## 4. Principios Rectores de Ingeniería

1. **Precisión de Reloj Suizo**: Cero ambigüedades, transacciones atómicas con bloqueo pesimista en PostgreSQL y libro mayor inmutable para los sellos (`stamp_ledger`).
2. **Degradación Controlada (Anillo 0)**: Si el backend en Go o la base de datos se desconectan temporalmente, el catálogo del menú y el generador de pedidos a WhatsApp funcionan 100% de forma autónoma desde el cliente.
3. **Seguridad Nivel Bancario (Zero-Trust)**:
   * Inmunidad total a inyecciones SQL mediante compilador SQLC.
   * Contraseñas con Argon2id para el personal.
   * Access Token JWT de 15 minutos en memoria + Refresh Token rotativo en cookie `HttpOnly; Secure; SameSite=Strict`.
4. **Ergonomía Móvil (Mobile-First)**: El 80% de las interacciones críticas se ubican en la zona inferior del pulgar con áreas táctiles mínimas de 44×44 px.
