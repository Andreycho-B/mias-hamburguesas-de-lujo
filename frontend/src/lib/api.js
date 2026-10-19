function resolveApiBase() {
  let url = import.meta.env.VITE_API_URL;
  if (url) {
    if (!url.startsWith('http://') && !url.startsWith('https://') && !url.startsWith('/')) {
      url = `https://${url}`;
    }
    if (!url.endsWith('/api/v1')) {
      url = url.replace(/\/+$/, '') + '/api/v1';
    }
    return url;
  }

  if (typeof window !== 'undefined' && window.location.hostname) {
    if (window.location.hostname.endsWith('.pages.dev')) {
      return '/api/v1';
    }
    // Si se accede desde un celular en la red Wi-Fi local (ej. 192.168.x.x),
    // apuntar al backend Go en esa misma IP de la máquina host.
    if (window.location.hostname !== 'localhost' && window.location.hostname !== '127.0.0.1') {
      return `http://${window.location.hostname}:8080/api/v1`;
    }
  }

  return 'http://localhost:8080/api/v1';
}

export const API_BASE = resolveApiBase();

function getAuthHeaders() {
  const token = typeof window !== 'undefined' ? localStorage.getItem('mias_access_token') : null;
  const headers = {
    'Content-Type': 'application/json',
    'Accept': 'application/json'
  };
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  return headers;
}

/**
 * Renueva la sesión usando el refresh token almacenado
 */
export async function refreshSessionAPI() {
  const refreshToken = typeof window !== 'undefined' ? localStorage.getItem('mias_refresh_token') : null;
  const res = await fetch(`${API_BASE}/auth/refresh`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ refresh_token: refreshToken || '' })
  });
  if (!res.ok) throw new Error('Error renovando sesión');
  const data = await res.json();
  if (data?.access_token) {
    localStorage.setItem('mias_access_token', data.access_token);
    if (data.refresh_token) {
      localStorage.setItem('mias_refresh_token', data.refresh_token);
    }
  }
  return data;
}

/**
 * Wrapper de fetch autenticado con reintento automático tras renovar token si responde 401
 */
export async function fetchWithAuth(url, options = {}) {
  const headers = options.headers ? { ...options.headers } : {};
  const token = typeof window !== 'undefined' ? localStorage.getItem('mias_access_token') : null;
  if (token && !headers['Authorization']) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  if (!headers['Content-Type'] && !(options.body instanceof FormData)) {
    headers['Content-Type'] = 'application/json';
  }
  if (!headers['Accept']) {
    headers['Accept'] = 'application/json';
  }

  let res = await fetch(url, { ...options, headers });

  // Si el access token expiró (401), intentar renovación automática una vez
  if (res.status === 401 && typeof window !== 'undefined' && localStorage.getItem('mias_refresh_token')) {
    try {
      const refreshed = await refreshSessionAPI();
      if (refreshed?.access_token) {
        headers['Authorization'] = `Bearer ${refreshed.access_token}`;
        res = await fetch(url, { ...options, headers });
      }
    } catch {
      // Sesión expirada definitivamente
    }
  }

  return res;
}

/**
 * Obtiene el catálogo gastronómico desde PostgreSQL con fallback al catálogo local
 */
export async function getCatalogAPI() {
  try {
    const res = await fetch(`${API_BASE}/catalog`, {
      headers: { 'Accept': 'application/json' }
    });
    if (!res.ok) throw new Error(`HTTP error ${res.status}`);
    const data = await res.json();
    return data;
  } catch (err) {
    console.warn('[API] Servidor Go offline o inalcanzable. Usando catálogo embebido Anillo 0.', err);
    return null;
  }
}

/**
 * Autentica cualquier usuario (cliente o personal) con correo y contraseña
 */
export async function loginAPI(email, password) {
  const res = await fetch(`${API_BASE}/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error || 'Credenciales incorrectas');
  }
  return data;
}

/**
 * Registra un nuevo cliente con correo, contraseña y nombre
 */
export async function registerAPI({ email, password, full_name, phone = '' }) {
  const res = await fetch(`${API_BASE}/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      email,
      password,
      full_name,
      phone,
      habeas_data_accepted: true
    })
  });
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error || 'Error al registrar usuario');
  }
  return data;
}

/**
 * Autentica personal con Argon2id y JWT (específico)
 */
export async function staffLoginAPI(email, password) {
  const res = await fetch(`${API_BASE}/auth/staff-login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error || 'Credenciales inválidas');
  }
  return data;
}

/**
 * Autentica o registra cliente mediante OAuth / Google
 */
export async function oauthLoginAPI({ email, full_name, avatar_url, phone = '', habeas_data_accepted = true, google_token = '' }) {
  const res = await fetch(`${API_BASE}/auth/oauth`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      email,
      full_name: full_name || email.split('@')[0],
      avatar_url: avatar_url || '',
      phone: phone || '',
      habeas_data_accepted: !!habeas_data_accepted,
      google_token: google_token || ''
    })
  });
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error || 'Error al autenticar con Google');
  }
  return data;
}

/**
 * Registra una comanda en el backend
 */
export async function createOrderAPI(orderPayload) {
  try {
    const res = await fetchWithAuth(`${API_BASE}/orders/`, {
      method: 'POST',
      body: JSON.stringify(orderPayload)
    });
    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || 'Error al procesar el pedido');
    }
    return data;
  } catch (err) {
    console.warn('[API] No se pudo enviar pedido al backend, operando localmente.', err);
    return null;
  }
}

/**
 * Consulta el estado y detalle de una comanda por su código alfanumérico
 */
export async function getOrderByCodeAPI(code) {
  try {
    const res = await fetchWithAuth(`${API_BASE}/orders/${code}`);
    if (!res.ok) throw new Error(`HTTP error ${res.status}`);
    return await res.json();
  } catch (err) {
    return null;
  }
}

/**
 * Consulta el historial completo de comandas del comensal autenticado o por códigos/teléfono desde PostgreSQL
 */
export async function getMyOrdersAPI(codes = [], phone = '') {
  try {
    const params = new URLSearchParams();
    if (codes && codes.length > 0) {
      params.append('codes', codes.join(','));
    }
    if (phone) {
      params.append('phone', phone);
    }
    const query = params.toString() ? `?${params.toString()}` : '';
    const res = await fetchWithAuth(`${API_BASE}/orders/my-orders${query}`);
    if (!res.ok) return null;
    return await res.json();
  } catch (err) {
    return null;
  }
}

/**
 * Solicita registro de sello con número de orden y factura física
 */
export async function requestStampAPI(orderID, receiptNumber) {
  const res = await fetchWithAuth(`${API_BASE}/stamps/request`, {
    method: 'POST',
    body: JSON.stringify({
      order_id: orderID,
      receipt_number: receiptNumber
    })
  });
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error || 'Factura ya registrada o no válida');
  }
  return data;
}

/**
 * Consulta el balance activo de sellos del comensal autenticado
 */
export async function getStampBalanceAPI() {
  try {
    const res = await fetchWithAuth(`${API_BASE}/stamps/balance`);
    if (!res.ok) throw new Error(`HTTP error ${res.status}`);
    return await res.json();
  } catch (err) {
    return null;
  }
}

/**
 * Solicita la redención presencial de la hamburguesa gratis al alcanzar 10 sellos
 */
export async function requestRedemptionAPI() {
  const res = await fetchWithAuth(`${API_BASE}/stamps/redeem`, {
    method: 'POST'
  });
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error || 'No fue posible solicitar la redención');
  }
  return data;
}

/**
 * Endpoints Staff: Pedidos pendientes en punto
 */
export async function getStaffPendingInStoreOrdersAPI() {
  try {
    const res = await fetchWithAuth(`${API_BASE}/orders/staff/pending-in-store`);
    if (!res.ok) throw new Error(`HTTP error ${res.status}`);
    return await res.json();
  } catch (err) {
    return null;
  }
}

/**
 * Endpoints Staff: Domicilios activos
 */
export async function getStaffActiveDeliveryOrdersAPI() {
  try {
    const res = await fetchWithAuth(`${API_BASE}/orders/staff/active-delivery`);
    if (!res.ok) throw new Error(`HTTP error ${res.status}`);
    return await res.json();
  } catch (err) {
    return null;
  }
}

/**
 * Endpoints Staff: Marcar orden como paga en caja
 */
export async function markOrderAsPaidAPI(orderID) {
  const res = await fetchWithAuth(`${API_BASE}/orders/${orderID}/pay`, {
    method: 'PATCH'
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Error al marcar orden como paga');
  return data;
}

/**
 * Endpoints Staff: Actualizar estado de despacho de domicilio
 */
export async function updateDeliveryStatusAPI(orderID, status) {
  const res = await fetchWithAuth(`${API_BASE}/orders/${orderID}/delivery-status`, {
    method: 'PATCH',
    body: JSON.stringify({ status })
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Error actualizando despacho');
  return data;
}

/**
 * Endpoints Staff: Listar solicitudes de sellos pendientes
 */
export async function getStaffPendingStampsAPI() {
  try {
    const res = await fetchWithAuth(`${API_BASE}/stamps/staff/pending`);
    if (!res.ok) throw new Error(`HTTP error ${res.status}`);
    return await res.json();
  } catch (err) {
    return null;
  }
}

/**
 * Endpoints Staff: Aprobar o rechazar solicitud de sellos
 */
export async function reviewStaffStampAPI(requestID, approve) {
  const res = await fetchWithAuth(`${API_BASE}/stamps/staff/${requestID}/review`, {
    method: 'PATCH',
    body: JSON.stringify({ approve })
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Error al procesar solicitud');
  return data;
}

/**
 * Endpoints Staff: Listar solicitudes de redención de hamburguesa gratis
 */
export async function getStaffPendingRedemptionsAPI() {
  try {
    const res = await fetchWithAuth(`${API_BASE}/stamps/staff/pending-redemptions`);
    if (!res.ok) throw new Error(`HTTP error ${res.status}`);
    return await res.json();
  } catch (err) {
    return null;
  }
}

/**
 * Endpoints Staff: Aprobar redención presencial en caja
 */
export async function approveRedemptionAPI(redemptionID) {
  const res = await fetchWithAuth(`${API_BASE}/stamps/staff/redemptions/${redemptionID}/approve`, {
    method: 'PATCH'
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Error al aprobar redención');
  return data;
}

/**
 * Endpoints Admin: Obtener métricas consolidadas de ventas, auditoría y rating
 */
export async function getAdminMetricsAPI() {
  try {
    const res = await fetchWithAuth(`${API_BASE}/admin/metrics`);
    if (!res.ok) throw new Error(`HTTP error ${res.status}`);
    return await res.json();
  } catch (err) {
    console.warn('[API] Error al cargar métricas de admin desde backend', err);
    return null;
  }
}

/**
 * Endpoints Admin: Conmutar disponibilidad de stock de producto en carta
 */
export async function toggleProductStockAPI(productId, inStock) {
  try {
    const res = await fetchWithAuth(`${API_BASE}/admin/catalog/products/${productId}/stock`, {
      method: 'PATCH',
      body: JSON.stringify({ is_in_stock: inStock })
    });
    if (!res.ok) throw new Error(`HTTP error ${res.status}`);
    return await res.json();
  } catch (err) {
    console.warn('[API] Error conmutando stock de producto en backend', err);
    return null;
  }
}
