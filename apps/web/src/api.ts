export interface Readiness {
  status: 'ok';
  database: 'ok';
}

export async function getReadiness(signal?: AbortSignal): Promise<Readiness> {
  const response = await fetch('/api/ready', { signal });
  if (!response.ok) {
    throw new Error(`API request failed (${response.status})`);
  }
  return response.json() as Promise<Readiness>;
}
