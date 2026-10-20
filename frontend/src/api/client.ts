// API Client for WhatsApp Reminder

export interface ApiResponse<T = any> {
  data?: T
  error?: string
  message?: string
}

let authToken: string = localStorage.getItem('token') || ''

export function setAuthToken(token: string) {
  authToken = token
  if (token) {
    localStorage.setItem('token', token)
  } else {
    localStorage.removeItem('token')
  }
}

export function getAuthToken(): string {
  return authToken
}

export async function request<T = any>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers || {})
  headers.set('Content-Type', 'application/json')
  
  if (authToken) {
    headers.set('Authorization', `Bearer ${authToken}`)
  }

  const response = await fetch(endpoint, {
    ...options,
    headers,
  })

  // Handle 401 Unauthorized
  if (response.status === 401 && !endpoint.includes('/auth/login') && !endpoint.includes('/auth/status')) {
    setAuthToken('')
    window.dispatchEvent(new CustomEvent('auth:expired'))
  }

  const data = await response.json().catch(() => null)

  if (!response.ok) {
    const errorMsg = data?.error || data?.message || `Request gagal (${response.status})`
    throw new Error(errorMsg)
  }

  return data as T
}

export const api = {
  get: <T>(url: string) => request<T>(url, { method: 'GET' }),
  post: <T>(url: string, body?: any) => request<T>(url, { method: 'POST', body: body ? JSON.stringify(body) : undefined }),
  put: <T>(url: string, body?: any) => request<T>(url, { method: 'PUT', body: body ? JSON.stringify(body) : undefined }),
  patch: <T>(url: string, body?: any) => request<T>(url, { method: 'PATCH', body: body ? JSON.stringify(body) : undefined }),
  delete: <T>(url: string) => request<T>(url, { method: 'DELETE' }),
}
