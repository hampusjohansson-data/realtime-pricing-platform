const API_BASE = "http://localhost:8088";

export async function fetchLatestPrice(symbol) {
  const res = await fetch(`${API_BASE}/prices/${encodeURIComponent(symbol)}`);
  if (!res.ok) {
    throw new Error(`Failed to fetch latest price: ${res.status}`);
  }
  return res.json();
}

export async function fetchHistory(symbol, limit = 50) {
  const res = await fetch(
    `${API_BASE}/prices/${encodeURIComponent(symbol)}/history?limit=${limit}`
  );
  if (!res.ok) {
    throw new Error(`Failed to fetch history: ${res.status}`);
  }
  return res.json();
}
