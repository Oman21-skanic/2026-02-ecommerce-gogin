import { API_BASE_URL } from "../env";

type RequestOptions = {
  method?: string;
  body?: unknown;
  token?: string;
};

export async function backendRequest<T>(path: string, options: RequestOptions = {}): Promise<{ data: T | null; error: string | null; status: number }> {
  const headers: Record<string, string> = {
    Accept: "application/json",
  };

  if (options.body !== undefined) {
    headers["Content-Type"] = "application/json";
  }

  if (options.token) {
    headers.Authorization = `Bearer ${options.token}`;
  }

  const response = await fetch(`${API_BASE_URL}${path}`, {
    method: options.method || "GET",
    headers,
    body: options.body !== undefined ? JSON.stringify(options.body) : undefined,
  });

  let payload: any = null;
  const isJSON = response.headers.get("content-type")?.includes("application/json");
  if (isJSON) {
    payload = await response.json();
  }

  if (!response.ok) {
    return {
      data: null,
      error: payload?.error || `Request failed (${response.status})`,
      status: response.status,
    };
  }

  return {
    data: payload as T,
    error: null,
    status: response.status,
  };
}
