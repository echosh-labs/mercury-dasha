// Sovereign Type-Safe API Client for Mercury Dasha

export class ApiError extends Error {
  status: number;
  data?: any;

  constructor(message: string, status: number, data?: any) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.data = data;
  }
}

async function handleResponse<T>(res: Response): Promise<T> {
  if (!res.ok) {
    let errorMsg = `HTTP ${res.status}: ${res.statusText}`;
    let errorData: any = null;
    try {
      errorData = await res.json();
      if (errorData?.error) {
        errorMsg = errorData.error;
      }
    } catch {
      // Non-JSON response body
    }
    throw new ApiError(errorMsg, res.status, errorData);
  }

  // Handle 204 No Content
  if (res.status === 204) {
    return {} as T;
  }

  const contentType = res.headers.get("content-type");
  if (contentType && contentType.includes("application/json")) {
    return (await res.json()) as T;
  }

  return (await res.text()) as unknown as T;
}

export const api = {
  async get<T>(url: string, init?: RequestInit): Promise<T> {
    const res = await fetch(url, {
      ...init,
      method: "GET",
      headers: {
        Accept: "application/json",
        ...init?.headers,
      },
    });
    return handleResponse<T>(res);
  },

  async post<T>(url: string, body?: any, init?: RequestInit): Promise<T> {
    const isFormData = typeof FormData !== "undefined" && body instanceof FormData;
    const headers: Record<string, string> = {
      Accept: "application/json",
      ...(init?.headers as Record<string, string>),
    };

    if (!isFormData && body !== undefined) {
      headers["Content-Type"] = "application/json";
    }

    const res = await fetch(url, {
      ...init,
      method: "POST",
      headers,
      body: isFormData ? body : body !== undefined ? JSON.stringify(body) : undefined,
    });
    return handleResponse<T>(res);
  },

  async delete<T>(url: string, init?: RequestInit): Promise<T> {
    const res = await fetch(url, {
      ...init,
      method: "DELETE",
      headers: {
        Accept: "application/json",
        ...init?.headers,
      },
    });
    return handleResponse<T>(res);
  },
};
