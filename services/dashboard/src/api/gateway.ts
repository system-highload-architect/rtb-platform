const BASE_URL = '/api';

export async function fetchReport(start: string, end: string) {
  const res = await fetch(`${BASE_URL}/report?start_date=${start}&end_date=${end}`);
  if (!res.ok) throw new Error('Report fetch failed');
  return res.json();
}

export async function fetchForecast(history: number[], horizon: number) {
  const params = new URLSearchParams({
    history: history.join(','),
    horizon: String(horizon),
  });
  const res = await fetch(`${BASE_URL}/forecast?${params}`);
  if (!res.ok) throw new Error('Forecast fetch failed');
  return res.json();
}

export async function fetchFactorAnalysis() {
  const res = await fetch(`${BASE_URL}/factor-analysis`);
  if (!res.ok) throw new Error('Factor analysis failed');
  return res.json();
}

export function getExportUrl(start: string, end: string) {
  return `/export/report?start_date=${start}&end_date=${end}`;
}