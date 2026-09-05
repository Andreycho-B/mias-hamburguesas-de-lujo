// Cliente API para Mías Hamburguesas de Lujo (Zero-Trust & Ring 0 Fallback)

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1';

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
  try {
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
  } catch (err) {
    throw err;
  }
}

/**
 * Registra una comanda en el backend
 */
export async function createOrderAPI(orderPayload) {
  try {
    const token = localStorage.getItem('mias_access_token');
    const headers = {
      'Content-Type': 'application/json',
      'Accept': 'application/json'
    };
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    const res = await fetch(`${API_BASE}/orders/`, {
      method: 'POST',
      headers,
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
 * Solicita registro de sello con número de factura
 */
export async function requestStampAPI(receiptNumber, customerName) {
  try {
    const token = localStorage.getItem('mias_access_token');
    const headers = {
      'Content-Type': 'application/json',
      'Accept': 'application/json'
    };
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }

    const res = await fetch(`${API_BASE}/stamps/request`, {
      method: 'POST',
      headers,
      body: JSON.stringify({
        receipt_number: receiptNumber,
        customer_name: customerName
      })
    });
    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || 'Factura ya registrada o no válida');
    }
    return data;
  } catch (err) {
    throw err;
  }
}
