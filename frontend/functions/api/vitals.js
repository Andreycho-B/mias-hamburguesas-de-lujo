export async function onRequestGet() {
  return new Response(JSON.stringify({ status: 'ok', service: 'vitals-telemetry-proxy' }), {
    headers: { 'Content-Type': 'application/json' },
  });
}

export async function onRequestPost(context) {
  const { request } = context;

  try {
    const body = await request.text();

    const headers = new Headers();
    headers.set('Content-Type', 'application/json');

    const clientIp = request.headers.get('CF-Connecting-IP');
    if (clientIp) {
      headers.set('CF-Connecting-IP', clientIp);
      headers.set('X-Forwarded-For', clientIp);
    }

    const userAgent = request.headers.get('User-Agent');
    if (userAgent) {
      headers.set('User-Agent', userAgent);
    }

    const referer = request.headers.get('Referer') || 'https://mias-hamburguesas-de-lujo.pages.dev/';
    headers.set('Referer', referer);

    // Reenvío server-to-server a Cloudflare Insights RUM
    await fetch('https://cloudflareinsights.com/cdn-cgi/rum', {
      method: 'POST',
      headers,
      body,
    });

    return new Response(null, {
      status: 204,
      headers: {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Methods': 'POST, OPTIONS',
      },
    });
  } catch (err) {
    return new Response(null, { status: 204 });
  }
}

export async function onRequestOptions() {
  return new Response(null, {
    status: 204,
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Methods': 'POST, OPTIONS',
      'Access-Control-Allow-Headers': 'Content-Type',
    },
  });
}
