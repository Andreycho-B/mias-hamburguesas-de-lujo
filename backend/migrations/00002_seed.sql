-- +goose Up
-- Datos Iniciales Oficiales del Menú (Extraídos de MIAS.pdf)

-- Categorías
INSERT INTO categories (id, slug, name, sort_order) VALUES
(1, 'hamburguesas', 'Hamburguesas', 1),
(2, 'perros', 'Perros Calientes', 2),
(3, 'de_la_casa', 'De la Casa', 3),
(4, 'bebidas', 'Bebidas', 4),
(5, 'adicionales', 'Adicionales', 5),
(6, 'temporada', 'Temporada', 6);

-- Productos: Hamburguesas (Pan Pretzel 160g)
INSERT INTO products (category_id, slug, name, description, price, bread_type, image_url, is_bestseller, is_chef_choice) VALUES
(1, 'original', 'Original', 'Pan pretzel, 160gr de carne de res premium, tocineta crujiente, queso mozzarella, cebolla caramelizada, tomate, salsa de la casa.', 17000, 'pretzel', '/imagenes/hamburguesas/Orignal.webp', TRUE, FALSE),
(1, 'crispy', 'Crispy', 'Pan pretzel, carne de pechuga apanada, queso mozzarella, ensalada coleslaw de repollo, pepinillos agridulces encurtidos, salsa miel mostaza artesanal.', 17000, 'pretzel', '/imagenes/hamburguesas/Crispy.webp', FALSE, FALSE),
(1, 'dorada', 'Dorada', 'Pan pretzel, 160gr de carne de res premium, queso Paipa fundido, piña glaseada al ron oscuro, tocineta crujiente, cebolla caramelizada, alioli de ajo rostizado artesanal.', 23000, 'pretzel', '/imagenes/hamburguesas/Golden.webp', TRUE, TRUE),
(1, 'criolla', 'Criolla', 'Pan pretzel, 160gr de carne de res premium, base de suero costeño, cubitos de plátano maduro, cebolla morada encurtida, lechuga en chiffonade.', 23000, 'pretzel', '/imagenes/hamburguesas/Criolla.webp', FALSE, FALSE),
(1, 'brava', 'Brava', 'Pan pretzel, 160gr de carne de res premium, queso colby jack, crema de aguacate, mermelada de cebolla al tequila, slaw de repollo y cilantro, mayonesa de chipotle artesanal ligeramente picante.', 26000, 'pretzel', '/imagenes/hamburguesas/Brava.webp', FALSE, TRUE),
(1, 'suprema', 'Suprema', 'Pan pretzel, doble carne premium (elección res, pechuga apanada o mixta), queso cheddar americano, cebolla caramelizada, tocineta crujiente, salsas de la casa.', 28000, 'pretzel', '/imagenes/hamburguesas/Suprema.webp', TRUE, FALSE);

-- Productos: Perros Calientes (Pan Brioche Artesanal)
INSERT INTO products (category_id, slug, name, description, price, bread_type, image_url) VALUES
(2, 'perro-clasico', 'Perro Clásico', 'Pan brioche artesanal, salchicha americana, cebolla caramelizada, queso mozzarella fundido, papa chips, salsas de la casa.', 15000, 'brioche', '/imagenes/perros/perro.png'),
(2, 'perro-especial', 'Perro Especial', 'Pan brioche artesanal, salchicha americana, cebolla caramelizada, queso cheddar americano fundido, tocineta, papa chips, salsas de la casa.', 17000, 'brioche', '/imagenes/perros/perro.png'),
(2, 'parrillero-choriperro', 'Parrillero (Choriperro)', 'Pan brioche artesanal, chorizo santarrosano, costra de queso paipa, crema de aguacate, cebolla encurtida, rúcula y alioli de ajo rostizado.', 20000, 'brioche', '/imagenes/perros/perro.png');

-- Productos: De la Casa
INSERT INTO products (category_id, slug, name, description, price, image_url) VALUES
(3, 'salchipapa', 'Salchipapa', 'Cortes de salchicha americana, papa francesa, queso mozzarella fundido, alioli de ajo rostizado, cebollín.', 22000, '/imagenes/de_la_casa/salchi_papa.png'),
(3, 'choripapa', 'Choripapa', 'Cortes de chorizo santarrosano, papa nativa, queso mozzarella, alioli de ajo rostizado, cebollín.', 24000, '/imagenes/de_la_casa/salchi_papa.png'),
(3, 'desgranado', 'Desgranado', 'Mazorca tierna, carne desmechada, papa nativa, queso mozzarella fundido, cubitos de plátano maduro, queso de costra dorada, queso costeño, salsa de suero costeño, alioli de ajo rostizado.', 30000, '/imagenes/de_la_casa/salchi_papa.png');

-- Productos: Bebidas
INSERT INTO products (category_id, slug, name, description, price, image_url) VALUES
(4, 'agua-con-gas', 'Agua con Gas', 'Botella de agua mineral con gas.', 3000, '/imagenes/bebidas/agua_con_gas.webp'),
(4, 'gaseosa', 'Gaseosa', 'Gaseosa Coca-Cola 400ml.', 4000, '/imagenes/bebidas/cocacola.webp'),
(4, 'bretana', 'Bretaña', 'Botella de Bretaña tradicional en vidrio.', 5000, '/imagenes/bebidas/bretana-vidrio.png');

-- Productos: Combos y Adicionales
INSERT INTO products (category_id, slug, name, description, price) VALUES
(5, 'combo-papa-gaseosa', 'Combo Papa a la Francesa + Gaseosa', 'Acompañamiento clásico para hamburguesa, perro o choriperro.', 11000),
(5, 'papa-francesa', 'Porción Papa Francesa', 'Porción individual de papa a la francesa crujiente.', 8000),
(5, 'papa-nativa', 'Porción Papa Nativa', 'Porción de papa nativa colombiana rústica al carbón.', 9000),
(5, 'tocineta-extra', 'Tocineta Crujiente Extra', 'Porción adicional de tocineta ahumada crujiente.', 5000),
(5, 'carne-extra', 'Carne de Hamburguesa Extra', '160gr adicionales de carne de res premium al carbón.', 8000),
(5, 'queso-extra', 'Queso Fundido Extra', 'Porción de queso mozzarella o cheddar fundido.', 5000),
(5, 'pepinillos-extra', 'Pepinillos Agridulces', 'Porción de pepinillos encurtidos artesanales.', 4000),
(5, 'salsa-de-la-casa', 'Salsa de la Casa', 'Tarro individual de salsa artesanal de la casa.', 4000);

-- Reset de secuencias
SELECT setval('categories_id_seq', (SELECT MAX(id) FROM categories));
SELECT setval('products_id_seq', (SELECT MAX(id) FROM products));

-- +goose Down
DELETE FROM products;
DELETE FROM categories;
