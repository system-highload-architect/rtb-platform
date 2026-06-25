const BASE_URL = '/api';
const RPC_URL = '/rpc';

function getToken(): string | null {
  return localStorage.getItem('token');
}

async function rpcRequest(method: string, params: any): Promise<any> {
  const res = await fetch(RPC_URL, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      jsonrpc: '2.0',
      method,
      params,
      id: 1,
    }),
  });
  if (!res.ok) throw new Error('RPC error');
  const data = await res.json();
  if (data.error) throw new Error(data.error.message);
  return data.result;
}

export async function login(email: string, password: string) {
  const result = await rpcRequest('auth.login', { email, password });
  localStorage.setItem('token', result.access_token);
  return result;
}

export async function register(email: string, password: string, role = 'advertiser') {
  return rpcRequest('auth.register', { email, password, role });
}

async function authFetch(url: string, options: RequestInit = {}) {
  const token = getToken();
  const headers = { ...options.headers } as Record<string, string>;
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  const res = await fetch(url, { ...options, headers });
  if (!res.ok) throw new Error('Request failed');
  return res;
}

export async function fetchReport(start: string, end: string) {
  const res = await authFetch(`${BASE_URL}/report?start_date=${start}&end_date=${end}`);
  return res.json();
}

export async function fetchForecast(history: number[], horizon: number) {
  const params = new URLSearchParams({ history: history.join(','), horizon: String(horizon) });
  const res = await authFetch(`${BASE_URL}/forecast?${params}`);
  return res.json();
}

export async function fetchFactorAnalysis() {
  const res = await authFetch(`${BASE_URL}/factor-analysis`);
  return res.json();
}

export function getExportUrl(start: string, end: string) {
  const token = getToken();
  const url = `/export/report?start_date=${start}&end_date=${end}`;
  // для скачивания проще открыть ссылку, браузер сам добавит куки/заголовки, но можно и через fetch с заголовком
  return url;
}