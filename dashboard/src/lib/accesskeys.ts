import axios from 'axios'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api/v1'

interface AccessKey {
  id: string
  name: string
  description: string
  key?: string // Only returned on creation
  permissions: string[]
  expires_at?: string
  last_used?: string
  active: boolean
  created_at: string
  updated_at: string
}

interface CreateAccessKeyRequest {
  name: string
  description?: string
  permissions: string[]
  expires_in?: number // Hours until expiration
}

export const accessKeysApi = {
  async create(data: CreateAccessKeyRequest): Promise<AccessKey> {
    const token = localStorage.getItem('token')
    const response = await axios.post(`${API_URL}/access-keys`, data, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    })
    return response.data
  },

  async list(): Promise<{ access_keys: AccessKey[] }> {
    const token = localStorage.getItem('token')
    const response = await axios.get(`${API_URL}/access-keys`, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    })
    return response.data
  },

  async get(id: string): Promise<AccessKey> {
    const token = localStorage.getItem('token')
    const response = await axios.get(`${API_URL}/access-keys/${id}`, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    })
    return response.data
  },

  async update(id: string, data: Partial<AccessKey>): Promise<void> {
    const token = localStorage.getItem('token')
    await axios.put(`${API_URL}/access-keys/${id}`, data, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    })
  },

  async regenerate(id: string): Promise<{ key: string; message: string }> {
    const token = localStorage.getItem('token')
    const response = await axios.post(`${API_URL}/access-keys/${id}/regenerate`, {}, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    })
    return response.data
  },

  async delete(id: string): Promise<void> {
    const token = localStorage.getItem('token')
    await axios.delete(`${API_URL}/access-keys/${id}`, {
      headers: {
        Authorization: `Bearer ${token}`
      }
    })
  }
}