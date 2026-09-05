# 13. Registro de Decisiones de Arquitectura y Producto (DECISION LOG)

**Proyecto:** Mías Hamburguesas de Lujo  
**Versión:** 1.0.0  
**Estado:** Normativo / Aprobado  

---

## Decisiones Clave Registradas

| ID | Fecha | Tema | Decisión Tomada | Justificación Técnica / Operativa |
| :--- | :--- | :--- | :--- | :--- |
| **ADR-001** | 2026-09-04 | Validación de Sellos | Digitación de número de factura física sin fotos ni códigos QR complejos. | Evita colapsar la caja en horas pico ("voleo"). El mesero o cajero aprueba en 1 clic verificando la comanda en su pantalla. |
| **ADR-002** | 2026-09-04 | Antifraude de Facturas | Restricción `UNIQUE` en base de datos sobre el número de factura física. | Impide matemáticamente que un número de factura se registre más de una vez por cualquier usuario. |
| **ADR-003** | 2026-09-04 | Identificación de Comandas | Prefijos alfanuméricos (`ORD-XXXX` en punto, `DOM-XXXX` en domicilio). | Diferencia de inmediato los números de pedido web de los números correlativos de factura física de la caja registradora. |
| **ADR-004** | 2026-09-05 | Redención de Hamburguesa | Aprobación presencial de caja sin códigos innecesarios. | Al completar 10 sellos, el cliente solicita en caja física. El cajero valida en su pantalla y aprueba. Desbloquea adiciones pagando solo el excedente. |
| **ADR-005** | 2026-09-05 | Pantalla de Bienvenida | Fondo crema claro, 3.5s, clic para saltar y cooldown de 60 minutos en `localStorage`. | Mantiene la elegancia editorial gastronómica y evita fricción a comensales recurrentes o que consultan su pedido. |
| **ADR-006** | 2026-09-05 | Alertas de Notificación | Web Audio API con síntesis nativa sin dependencias externas. | Garantiza cero latencia y reproduce un chime armónico y audible en sala y caja ante nuevos pedidos y solicitudes de sellos. |
| **ADR-007** | 2026-09-05 | Impresión de Comandas | Plantilla CSS `@media print` para formato térmico de 80mm/58mm. | Permite al cajero imprimir comandas de cocina directamente desde su pantalla web sin necesidad de controladores propietarios complejos. |
| **ADR-008** | 2026-09-05 | Caducidad de Sellos | 6 meses desde la última compra registrada. | Fomenta la recurrencia periódica protegiendo la viabilidad financiera del restaurante. |
| **ADR-009** | 2026-09-05 | Seguridad de Consultas | Inmunidad total a inyección SQL usando compilador SQLC sobre PostgreSQL 16. | Genera código Go fuertemente tipado; imposibilita la concatenación dinámica de strings en SQL. |
| **ADR-010** | 2026-09-05 | Persistencia de Carrito | Migración automática del carrito anónimo al iniciar sesión con Google. | Previene la tasa de abandono: el cliente no pierde los productos seleccionados al momento de registrarse. |
| **ADR-011** | 2026-09-05 | Estandarización de Marca | Eliminación absoluta de cualquier referencia a ".8" o "8". | El isotipo oficial es un candado de hierro en forma de corazón (que visualmente recordaba a un 8). El nombre oficial de la marca es estrictamente «Mías Hamburguesas de Lujo». |
| **ADR-012** | 2026-09-05 | Lenguaje Visual de Autor | Transición de Brasa (Bienvenida) a Crema Editorial (Home). | Reemplazo del diseño oscuro tipo SaaS por el lenguaje de alta autoría aprobado: gradiente de carbón a brasa viva con retícula de estrella cóncava en splash, e interior crema cálido (#F5F0EB). |
