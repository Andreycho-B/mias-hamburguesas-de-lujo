-- +goose Up
-- Modalidad "Para Llevar": el frontend la ofrece y el ticket la distingue
-- (TKA), pero el CHECK solo admitía in_store/delivery y el tipo se perdía
-- como in_store. takeout se comporta como in_store (sin tarifa, ORD-, pago
-- en punto) conservando su identidad para caja y reportes.
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_order_type_check;
ALTER TABLE orders ADD CONSTRAINT orders_order_type_check CHECK (order_type IN ('in_store', 'delivery', 'takeout'));

-- +goose Down
ALTER TABLE orders DROP CONSTRAINT IF EXISTS orders_order_type_check;
ALTER TABLE orders ADD CONSTRAINT orders_order_type_check CHECK (order_type IN ('in_store', 'delivery'));
