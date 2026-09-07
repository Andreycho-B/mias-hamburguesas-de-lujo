// Store reactivo con Runes de Svelte 5 (Anillo 0: Datos locales con sincronización)
import { getCatalogAPI, getStampBalanceAPI, getOrderByCodeAPI } from './api.js';

const SEED_CATEGORIES = [
  { id: 1, slug: 'hamburguesas', name: 'Hamburguesas', sort_order: 1 },
  { id: 2, slug: 'perros', name: 'Perros Calientes', sort_order: 2 },
  { id: 3, slug: 'de_la_casa', name: 'De la Casa', sort_order: 3 },
  { id: 4, slug: 'bebidas', name: 'Bebidas', sort_order: 4 },
  { id: 5, slug: 'adicionales', name: 'Adicionales', sort_order: 5 },
  { id: 6, slug: 'temporada', name: 'Temporada (Amor y Amistad)', sort_order: 6 }
];

const SEED_PRODUCTS = [
  {
    id: 1,
    category_id: 1,
    slug: 'original',
    name: 'Original',
    description: 'Pan pretzel, 160gr de carne de res premium, tocineta crujiente, queso mozzarella, cebolla caramelizada, tomate, salsa de la casa.',
    price: 17000,
    bread_type: 'pretzel',
    image_url: '/imagenes/hamburguesas/Orignal.webp',
    is_bestseller: true,
    is_chef_choice: false,
    is_in_stock: true
  },
  {
    id: 2,
    category_id: 1,
    slug: 'crispy',
    name: 'Crispy',
    description: 'Pan pretzel, carne de pechuga apanada, queso mozzarella, ensalada coleslaw de repollo, pepinillos agridulces encurtidos, salsa miel mostaza artesanal.',
    price: 17000,
    bread_type: 'pretzel',
    image_url: '/imagenes/hamburguesas/Crispy.webp',
    is_bestseller: false,
    is_chef_choice: false,
    is_in_stock: true
  },
  {
    id: 3,
    category_id: 1,
    slug: 'dorada',
    name: 'Dorada',
    description: 'Pan pretzel, 160gr de carne de res premium, queso Paipa fundido, piña glaseada al ron oscuro, tocineta crujiente, cebolla caramelizada, alioli de ajo rostizado artesanal.',
    price: 23000,
    bread_type: 'pretzel',
    image_url: '/imagenes/hamburguesas/Golden.webp',
    is_bestseller: true,
    is_chef_choice: true,
    is_in_stock: true
  },
  {
    id: 4,
    category_id: 1,
    slug: 'criolla',
    name: 'Criolla',
    description: 'Pan pretzel, 160gr de carne de res premium, base de suero costeño, cubitos de plátano maduro, cebolla morada encurtida, lechuga en chiffonade.',
    price: 23000,
    bread_type: 'pretzel',
    image_url: '/imagenes/hamburguesas/Criolla.webp',
    is_bestseller: false,
    is_chef_choice: false,
    is_in_stock: true
  },
  {
    id: 5,
    category_id: 1,
    slug: 'brava',
    name: 'Brava',
    description: 'Pan pretzel, 160gr de carne de res premium, queso colby jack, crema de aguacate, mermelada de cebolla al tequila, slaw de repollo y cilantro, mayonesa de chipotle artesanal.',
    price: 26000,
    bread_type: 'pretzel',
    image_url: '/imagenes/hamburguesas/Brava.webp',
    is_bestseller: false,
    is_chef_choice: true,
    is_in_stock: true
  },
  {
    id: 6,
    category_id: 1,
    slug: 'suprema',
    name: 'Suprema',
    description: 'Pan pretzel, doble carne premium (res, pechuga apanada o mixta), queso cheddar americano, cebolla caramelizada, tocineta crujiente, salsas de la casa.',
    price: 28000,
    bread_type: 'pretzel',
    image_url: '/imagenes/hamburguesas/Suprema.webp',
    is_bestseller: true,
    is_chef_choice: false,
    is_in_stock: true
  },
  {
    id: 7,
    category_id: 2,
    slug: 'perro-clasico',
    name: 'Perro Clásico',
    description: 'Pan brioche artesanal, salchicha americana, cebolla caramelizada, queso mozzarella fundido, papa chips, salsas de la casa.',
    price: 15000,
    bread_type: 'brioche',
    image_url: '/imagenes/perros/perro.png',
    is_bestseller: false,
    is_chef_choice: false,
    is_in_stock: true
  },
  {
    id: 8,
    category_id: 2,
    slug: 'perro-especial',
    name: 'Perro Especial',
    description: 'Pan brioche artesanal, salchicha americana, cebolla caramelizada, queso cheddar americano fundido, tocineta, papa chips, salsas de la casa.',
    price: 17000,
    bread_type: 'brioche',
    image_url: '/imagenes/perros/perro.png',
    is_bestseller: false,
    is_chef_choice: false,
    is_in_stock: true
  },
  {
    id: 9,
    category_id: 2,
    slug: 'parrillero-choriperro',
    name: 'Parrillero (Choriperro)',
    description: 'Pan brioche artesanal, chorizo santarrosano, costra de queso paipa, crema de aguacate, cebolla encurtida, rúcula y alioli de ajo rostizado.',
    price: 20000,
    bread_type: 'brioche',
    image_url: '/imagenes/perros/perro.png',
    is_bestseller: true,
    is_chef_choice: true,
    is_in_stock: true
  },
  {
    id: 10,
    category_id: 3,
    slug: 'salchipapa',
    name: 'Salchipapa',
    description: 'Cortes de salchicha americana, papa francesa, queso mozzarella fundido, alioli de ajo rostizado, cebollín.',
    price: 22000,
    image_url: '/imagenes/de_la_casa/salchi_papa.png',
    is_bestseller: false,
    is_chef_choice: false,
    is_in_stock: true
  },
  {
    id: 11,
    category_id: 3,
    slug: 'choripapa',
    name: 'Choripapa',
    description: 'Cortes de chorizo santarrosano, papa nativa, queso mozzarella, alioli de ajo rostizado, cebollín.',
    price: 24000,
    image_url: '/imagenes/de_la_casa/salchi_papa.png',
    is_bestseller: false,
    is_chef_choice: false,
    is_in_stock: true
  },
  {
    id: 12,
    category_id: 3,
    slug: 'desgranado',
    name: 'Desgranado',
    description: 'Mazorca tierna, carne desmechada, papa nativa, queso mozzarella fundido, cubitos de plátano maduro, queso de costra dorada, queso costeño, salsa de suero costeño, alioli de ajo rostizado.',
    price: 30000,
    image_url: '/imagenes/de_la_casa/salchi_papa.png',
    is_bestseller: true,
    is_chef_choice: false,
    is_in_stock: true
  },
  {
    id: 13,
    category_id: 4,
    slug: 'agua-con-gas',
    name: 'Agua con Gas',
    description: 'Botella de agua mineral con gas.',
    price: 3000,
    image_url: '/imagenes/bebidas/agua_con_gas.webp',
    is_bestseller: false,
    is_chef_choice: false,
    is_in_stock: true
  },
  {
    id: 14,
    category_id: 4,
    slug: 'gaseosa',
    name: 'Gaseosa',
    description: 'Gaseosa Coca-Cola tradicional 400ml.',
    price: 4000,
    image_url: '/imagenes/bebidas/cocacola.webp',
    is_bestseller: true,
    is_chef_choice: false,
    is_in_stock: true
  },
  {
    id: 15,
    category_id: 4,
    slug: 'bretana',
    name: 'Bretaña',
    description: 'Botella de Bretaña tradicional en vidrio.',
    price: 5000,
    image_url: '/imagenes/bebidas/bretana-vidrio.png',
    is_bestseller: false,
    is_chef_choice: false,
    is_in_stock: true
  },
  {
    id: 16,
    category_id: 5,
    slug: 'combo-papa-gaseosa',
    name: 'Combo Papa a la Francesa + Gaseosa',
    description: 'Acompañamiento clásico para hamburguesa, perro o choriperro con papas crujientes y refresco helado.',
    price: 11000,
    is_bestseller: true,
    is_chef_choice: false,
    is_in_stock: true
  },
  {
    id: 17,
    category_id: 5,
    slug: 'papa-francesa',
    name: 'Porción Papa Francesa',
    description: 'Porción individual de papa a la francesa crujiente dorada al punto de sal.',
    price: 8000,
    is_bestseller: false,
    is_chef_choice: false,
    is_in_stock: true
  },
  {
    id: 18,
    category_id: 5,
    slug: 'papa-nativa',
    name: 'Porción Papa Nativa',
    description: 'Porción de papa nativa colombiana rústica al carbón con alioli artesanal.',
    price: 9000,
    is_bestseller: true,
    is_chef_choice: true,
    is_in_stock: true
  },
  {
    id: 19,
    category_id: 5,
    slug: 'tocineta-extra',
    name: 'Tocineta Crujiente Extra',
    description: 'Porción adicional de tocineta ahumada crujiente cortada gruesa.',
    price: 5000,
    is_bestseller: false,
    is_chef_choice: false,
    is_in_stock: true
  },
  {
    id: 20,
    category_id: 5,
    slug: 'carne-extra',
    name: 'Carne de Hamburguesa Extra',
    description: '160gr adicionales de carne de res premium al carbón.',
    price: 8000,
    is_bestseller: false,
    is_chef_choice: false,
    is_in_stock: true
  },
  {
    id: 21,
    category_id: 5,
    slug: 'queso-extra',
    name: 'Queso Fundido Extra',
    description: 'Porción de queso mozzarella o cheddar fundido sobre tu plato.',
    price: 5000,
    is_bestseller: false,
    is_chef_choice: false,
    is_in_stock: true
  },
  {
    id: 22,
    category_id: 5,
    slug: 'pepinillos-extra',
    name: 'Pepinillos Agridulces',
    description: 'Porción de pepinillos encurtidos artesanales crocantes.',
    price: 4000,
    is_bestseller: false,
    is_chef_choice: false,
    is_in_stock: true
  },
  {
    id: 23,
    category_id: 5,
    slug: 'salsa-de-la-casa',
    name: 'Salsa de la Casa',
    description: 'Tarro individual de salsa artesanal de la casa.',
    price: 4000,
    is_bestseller: false,
    is_chef_choice: false,
    is_in_stock: true
  }
];

class AppStore {
  categories = $state(SEED_CATEGORIES);
  products = $state(SEED_PRODUCTS);
  cart = $state([]);
  user = $state(null);
  stampBalance = $state(0);
  activeOrders = $state([]);
  selectedProduct = $state(null); // Producto abierto en pantalla completa
  isCartOpen = $state(false);
  isSidebarOpen = $state(false);
  isOrderModeModalOpen = $state(false);
  isDeliveryWizardOpen = $state(false);
  isAuthModalOpen = $state(false);
  currentView = $state('menu'); // 'menu', 'stamps', 'history', 'staff'
  staffRole = $state(null); // 'waiter', 'cashier', 'admin'

  cartCount = $derived(
    this.cart.reduce((sum, item) => sum + item.quantity, 0)
  );

  cartSubtotal = $derived(
    this.cart.reduce((sum, item) => sum + item.totalPrice, 0)
  );

  constructor() {
    this.loadPersistedData();
    this.initOrderTracking();
  }

  loadPersistedData() {
    if (typeof window === 'undefined') return;
    try {
      const savedCart = localStorage.getItem('mias_cart');
      if (savedCart) this.cart = JSON.parse(savedCart);

      const savedUser = localStorage.getItem('mias_user');
      if (savedUser) {
        this.user = JSON.parse(savedUser);
        if (this.user && ['waiter', 'cashier', 'admin'].includes(this.user.role)) {
          this.staffRole = this.user.role;
          this.currentView = 'staff';
        }
      }

      const savedOrders = localStorage.getItem('mias_orders');
      if (savedOrders) this.activeOrders = JSON.parse(savedOrders);

      const savedStamps = localStorage.getItem('mias_stamps');
      if (savedStamps) this.stampBalance = parseInt(savedStamps, 10);
    } catch {}

    // Sincronizar dinámicamente con PostgreSQL manteniendo Anillo 0 como respaldo
    this.refreshCatalog();
    if (this.user) {
      this.refreshStampBalance();
    }
  }

  async refreshCatalog() {
    try {
      const data = await getCatalogAPI();
      if (data && data.products && data.products.length > 0) {
        this.products = data.products;
        if (data.categories && data.categories.length > 0) {
          this.categories = data.categories;
        }
      }
    } catch (err) {
      console.warn('[Store] Operando con catálogo local Anillo 0', err);
    }
  }

  async refreshStampBalance() {
    try {
      const res = await getStampBalanceAPI();
      if (res && typeof res.active_stamps === 'number') {
        this.setStampBalance(res.active_stamps);
      }
    } catch {}
  }

  saveCart() {
    if (typeof window === 'undefined') return;
    localStorage.setItem('mias_cart', JSON.stringify(this.cart));
  }

  addToCart(product, combo = false, extras = []) {
    let unitPrice = product.price;
    if (combo) unitPrice += 11000;
    for (const ex of extras) {
      unitPrice += ex.price;
    }

    const existingIndex = this.cart.findIndex(
      (item) => item.product.id === product.id && item.combo === combo && JSON.stringify(item.extras) === JSON.stringify(extras)
    );

    if (existingIndex > -1) {
      this.cart[existingIndex].quantity += 1;
      this.cart[existingIndex].totalPrice = this.cart[existingIndex].quantity * unitPrice;
    } else {
      this.cart.push({
        id: Date.now() + Math.random(),
        product,
        quantity: 1,
        combo,
        extras,
        unitPrice,
        totalPrice: unitPrice
      });
    }

    this.saveCart();
  }

  updateQuantity(itemId, delta) {
    const item = this.cart.find((i) => i.id === itemId);
    if (!item) return;

    item.quantity += delta;
    if (item.quantity <= 0) {
      this.cart = this.cart.filter((i) => i.id !== itemId);
    } else {
      item.totalPrice = item.quantity * item.unitPrice;
    }
    this.saveCart();
  }

  clearCart() {
    this.cart = [];
    this.saveCart();
  }

  setUser(user) {
    this.user = user;
    if (user && ['waiter', 'cashier', 'admin'].includes(user.role)) {
      this.staffRole = user.role;
      this.currentView = 'staff';
    } else {
      this.staffRole = null;
      if (this.currentView === 'staff') {
        this.currentView = 'menu';
      }
    }
    if (typeof window !== 'undefined') {
      localStorage.setItem('mias_user', JSON.stringify(user));
    }
    if (user) {
      this.refreshStampBalance();
    }
  }

  logout() {
    this.user = null;
    this.staffRole = null;
    this.currentView = 'menu';
    this.isSidebarOpen = false;
    this.isCartOpen = false;
    this.isAuthModalOpen = false;
    if (typeof window !== 'undefined') {
      localStorage.removeItem('mias_user');
      localStorage.removeItem('mias_access_token');
      localStorage.removeItem('mias_refresh_token');
    }
  }

  initOrderTracking() {
    if (typeof window === 'undefined') return;

    // Conexión SSE para eventos de órdenes en vivo
    try {
      const apiBase = import.meta.env.VITE_API_URL || (window.location.hostname ? `http://${window.location.hostname}:8080/api/v1` : 'http://localhost:8080/api/v1');
      const token = localStorage.getItem('mias_access_token') || '';
      const sseUrl = `${apiBase}/events/orders${token ? `?token=${token}` : ''}`;
      const es = new EventSource(sseUrl);

      es.addEventListener('order_paid', (e) => {
        try {
          const data = JSON.parse(e.data);
          const payload = data.data || data;
          if (payload?.code) {
            this.updateOrderStatus(payload.code, 'PAID');
            this.refreshStampBalance();
          }
        } catch {}
      });

      es.addEventListener('delivery_status_updated', (e) => {
        try {
          const data = JSON.parse(e.data);
          const payload = data.data || data;
          if (payload?.code && payload?.status) {
            this.updateOrderStatus(payload.code, payload.status);
          }
        } catch {}
      });

      es.addEventListener('stamp_request_reviewed', () => {
        this.refreshStampBalance();
      });
    } catch (err) {
      console.warn('[Store] SSE no iniciado', err);
    }

    // Polling resiliente cada 3.5 segundos para sincronizar el estado de órdenes activas del cliente
    setInterval(async () => {
      if (this.staffRole || ['waiter', 'cashier', 'admin'].includes(this.user?.role)) return;

      const pendingOrders = this.activeOrders.filter(o => 
        o.status !== 'PAID' && o.status !== 'DELIVERED' && o.status !== 'CANCELLED'
      );

      for (const ord of pendingOrders) {
        if (!ord.code) continue;
        try {
          const res = await getOrderByCodeAPI(ord.code);
          if (res && res.order && res.order.status && res.order.status !== ord.status) {
            this.updateOrderStatus(ord.code, res.order.status);
            if (res.order.status === 'PAID') {
              this.refreshStampBalance();
            }
          }
        } catch {}
      }
    }, 3500);
  }

  addOrder(order) {
    this.activeOrders = [order, ...this.activeOrders];
    if (typeof window !== 'undefined') {
      localStorage.setItem('mias_orders', JSON.stringify(this.activeOrders));
    }
  }

  updateOrderStatus(orderCode, newStatus) {
    const ord = this.activeOrders.find((o) => o.code === orderCode);
    if (ord) {
      ord.status = newStatus;
      if (typeof window !== 'undefined') {
        localStorage.setItem('mias_orders', JSON.stringify(this.activeOrders));
      }
    }
  }

  setStampBalance(stamps) {
    this.stampBalance = stamps;
    if (typeof window !== 'undefined') {
      localStorage.setItem('mias_stamps', stamps.toString());
    }
  }

  toggleSidebar() {
    this.isSidebarOpen = !this.isSidebarOpen;
  }

  openSidebar() {
    this.isSidebarOpen = true;
  }

  closeSidebar() {
    this.isSidebarOpen = false;
  }
}

export const store = new AppStore();
