import axios from 'axios'

const API_BASE_URL = import.meta.env.VITE_API_URL || '/api/v1'

const apiClient = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
})

// Add auth token to requests
apiClient.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// Handle auth errors
apiClient.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token')
      localStorage.removeItem('user')
      window.location.href = '/login'
    }
    return Promise.reject(error)
  }
)

export interface UserSettings {
  theme: string
  language: string
  timezone: string
  date_format: string
  time_format: string
  email_notifications: boolean
  security_alerts: boolean
}

export interface SystemSettings {
  connection_pool_size: number
  query_timeout: number
  max_retries: number
  compression_enabled: boolean
  encryption_enabled: boolean
  session_timeout?: number
  password_policy?: string
  mfa_required?: boolean
  audit_log_enabled?: boolean
  rate_limit: number
  burst_limit?: number
  cors_enabled: boolean
}

export const settingsApi = {
  // User settings
  getUserSettings: async () => {
    const response = await apiClient.get<UserSettings>('/settings/user')
    return response.data
  },

  updateUserSettings: async (settings: Partial<UserSettings>) => {
    const response = await apiClient.put('/settings/user', settings)
    return response.data
  },

  // System settings
  getSystemSettings: async () => {
    const response = await apiClient.get<SystemSettings>('/settings/system')
    return response.data
  },

  updateSystemSettings: async (settings: Partial<SystemSettings>) => {
    const response = await apiClient.put('/settings/system', settings)
    return response.data
  }
}