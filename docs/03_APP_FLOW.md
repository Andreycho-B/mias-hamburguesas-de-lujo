# 03. Flujos de la Aplicación y Diagramas de Navegación (APP FLOW)

**Proyecto:** Mías.8 Hamburguesas de Lujo  
**Versión:** 1.0.0  
**Estado:** Normativo / Aprobado  

---

## 1. Mapa de Navegación General

```mermaid
graph TD
    A[Entrada / URL] --> B{¿Intro vista < 60 min?}
    B -- No --> C[Intro Screen: 3.5s o Clic / Fondo Crema]
    B -- Sí --> D[Home: Menú Gastronómico Suizo]
    C --> D
    
    D --> E[Ficha de Producto a Pantalla Completa]
    E --> F[Selección de Combos / Acompañamientos]
    F --> G[Carrito Flotante Inferior 56px]
    
    G --> H{¿Usuario Autenticado?}
    H -- No --> I[Tarjeta Consentimiento Habeas Data ➔ Google OAuth]
    I --> J[Migración Automática de Carrito]
    H -- Sí --> K[Selector: ¿En el Punto o Domicilio?]
    J --> K
    
    K -- En el Punto --> L[Genera ORD-XXXX ➔ Tarjeta 'POR COBRAR' en Caja]
    L --> M[Cliente Paga en Caja ➔ Recibe Factura Física y Localizador Mesa]
    M --> N[Cajero marca 'ORDEN PAGA' ➔ Imprime Comanda Térmica]
    N --> O[Cliente en 'Mis Sellos' pulsa 'Solicitar Sello' e ingresa N° Factura]
    O --> P[Alerta SSE a Mesero/Caja ➔ Personal Aprueba ➔ Sellos Acreditados]
    
    K -- Domicilio --> Q[Wizard Conversacional Paso a Paso: Datos + Dirección + Pago]
    Q --> R[Genera DOM-XXXX ➔ Enlace WhatsApp 321 252 0881 + Registro en Caja]
    R --> S[Cajero cambia a 'EN CAMINO' ➔ Cliente marca 'Recibido' + Reseña 5★]

    P --> T{¿Completó 10 Sellos?}
    T -- Sí --> U[Botón: 'Redimir Hamburguesa Original']
    U --> V[Validación Presencial en Caja ➔ Cajero Aprueba ➔ Celebración GSAP]
```

---

## 2. Flujo del Personal (Mesero, Cajero, Admin)

```mermaid
graph TD
    A1[Login Staff: Usuario / Contraseña Argon2id] --> B1{Rol del Usuario}
    
    B1 -- Mesero --> C1[Panel Mesero Móvil: Bandeja de Solicitudes de Sellos]
    C1 --> D1[Revisión de N° Factura vs Orden ➔ Botones Aprobar / Rechazar]
    
    B1 -- Cajero --> E1[Panel Caja Central: Órdenes Locales + Domicilios + Sellos + Redenciones]
    E1 --> F1[Órdenes Locales: Botón Rojo 'Por Cobrar' ➔ Botón Verde 'Orden Paga' + Imprimir]
    E1 --> G1[Domicilios: 'En Preparación' ➔ 'En Camino']
    E1 --> H1[Redenciones: Validación presencial de 10 sellos ➔ Aprobar Hamburguesa Gratis]
    
    B1 -- Admin --> I1[Dashboard Ejecutivo + CMS Completo]
    I1 --> J1[Métricas BI: Top Ventas, Top Calificaciones, Retención]
    I1 --> K1[CMS: Precios, Stock Inmediato, Categoría de Temporada]
    I1 --> L1[Auditoría Forense Inmutable: admin_audit_logs]
```
