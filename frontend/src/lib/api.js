const API_BASE = import.meta.env.VITE_API_URL || (typeof window !== 'undefined' && window.location.hostname ? `http://${window.location.hostname}:8080/api/v1` : 'http://localhost:8080/api/v1');

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
 * Autentica personal con Argon2id y JWT
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
 * Registra una comanda en el backend
 */
export async function createOrderAPI(orderPayload) {
  try {
    const res = await fetch(`${API_BASE}/orders/`, {
      method: 'POST',
      headers: getAuthHeaders(),
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
    const res = await fetch(`${API_BASE}/orders/${code}`, {
      headers: getAuthHeaders()
    });
    if (!res.ok) throw new Error(`HTTP error ${res.status}`);
    return await res.json();
  } catch (err) {
    return null;
  }
}

/**
 * Solicita registro de sello con número de orden y factura física
 */
export async function requestStampAPI(orderID, receiptNumber) {
  const res = await fetch(`${API_BASE}/stamps/request`, {
    method: 'POST',
    headers: getAuthHeaders(),
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
    const res = await fetch(`${API_BASE}/stamps/balance`, {
      headers: getAuthHeaders()
    });
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
  const res = await fetch(`${API_BASE}/stamps/redeem`, {
    method: 'POST',
    headers: getAuthHeaders()
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
    const res = await fetch(`${API_BASE}/orders/staff/pending-in-store`, {
      headers: getAuthHeaders()
    });
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
    const res = await fetch(`${API_BASE}/orders/staff/active-delivery`, {
      headers: getAuthHeaders()
    });
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
  const res = await fetch(`${API_BASE}/orders/${orderID}/pay`, {
    method: 'PATCH',
    headers: getAuthHeaders()
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Error al marcar orden como paga');
  return data;
}

/**
 * Endpoints Staff: Actualizar estado de despacho de domicilio
 */
export async function updateDeliveryStatusAPI(orderID, status) {
  const res = await fetch(`${API_BASE}/orders/${orderID}/delivery-status`, {
    method: 'PATCH',
    headers: getAuthHeaders(),
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
    const res = await fetch(`${API_BASE}/stamps/staff/pending`, {
      headers: getAuthHeaders()
    });
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
  const res = await fetch(`${API_BASE}/stamps/staff/${requestID}/review`, {
    method: 'PATCH',
    headers: getAuthHeaders(),
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
    const res = await fetch(`${API_BASE}/stamps/staff/pending-redemptions`, {
      headers: getAuthHeaders()
    });
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
  const res = await fetch(`${API_BASE}/stamps/staff/redemptions/${redemptionID}/approve`, {
    method: 'PATCH',
    headers: getAuthHeaders()
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Error al aprobar redención');
  return data;
}
