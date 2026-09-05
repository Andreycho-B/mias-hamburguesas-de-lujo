# 17. Glosario del Proyecto (PROJECT GLOSSARY)

**Proyecto:** Mías.8 Hamburguesas de Lujo  
**Versión:** 1.0.0  
**Estado:** Normativo / Aprobado  

---

## Términos del Dominio

* **Comanda (`ORD-XXXX` / `DOM-XXXX`)**: Identificador alfanumérico generado por la plataforma web para representar el pedido de un cliente, diferenciado de la factura fiscal o recibo físico de caja.
* **Factura Física (Recibo de Caja)**: Documento físico impreso por la caja registradora del restaurante con un número correlativo secuencial (ej. `#00124`), utilizado por el cliente para certificar su compra y solicitar sellos.
* **Localizador de Mesa**: Objeto físico (número de madera o acrílico) entregado por el cajero al cliente al momento de pagar en caja, que permite al mesero ubicar al comensal en el salón sin requerir geolocalización o selección manual de mesa en la app.
* **Sello (Stamp)**: Unidad de fidelización otorgada al cliente por cada hamburguesa calificada comprada en una orden web pagada.
* **Libro Mayor Inmutable (`stamp_ledger`)**: Tabla relacional contable donde cada movimiento de sellos (+N por compra, -10 por redención) se almacena como una transacción de solo inserción, calculando el saldo real como la suma acumulada de transacciones.
* **Redención**: Proceso mediante el cual un cliente que completó 10 sellos solicita presencialmente en caja su Hamburguesa Original ($17.000 COP) sin costo, pagando únicamente las adiciones o combos excedentes si decide agregarlos.
* **Pan Pretzel**: Tipo de pan artesanal brillante de corteza dorada y densa, exclusivo de todas las hamburguesas de Mías.
* **Pan Brioche Artesanal**: Tipo de pan suave y enriquecido, exclusivo de los perros calientes de la casa.
* **Horno Rayo**: Horno a carbón con acabado en rojo escarlata y latón dorado donde se asan las carnes premium del restaurante.
* **Candado de Amor**: Isotipo oficial y tradición del restaurante, donde los comensales sellan su visita colocando un candado en la reja decorativa del establecimiento.
