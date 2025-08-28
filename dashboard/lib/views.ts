import axios from 'axios'
import Cookies from 'js-cookie'

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'

const getAuthHeaders = () => {
  const token = Cookies.get('token')
  return token ? { Authorization: `Bearer ${token}` } : {}
}

export const viewsApi = {
  // List all views
  list: async () => {
    const response = await axios.get(`${API_BASE_URL}/api/v1/views`, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Create a new view
  create: async (data: {
    name: string
    description?: string
    collection: string
    filter?: Record<string, any>
    fields?: string[]
  }) => {
    const response = await axios.post(`${API_BASE_URL}/api/v1/views`, data, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Get a specific view
  get: async (name: string) => {
    const response = await axios.get(`${API_BASE_URL}/api/v1/views/${name}`, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Query/execute a view with runtime filters and sort
  query: async (name: string, params?: { 
    limit?: number; 
    skip?: number; 
    filter?: Record<string, any>; // Extra runtime filters
    sort?: Record<string, number>; // Runtime sort
  }) => {
    const queryParams: any = {}
    if (params?.limit) queryParams.limit = params.limit
    if (params?.skip) queryParams.skip = params.skip
    if (params?.filter) queryParams.filter = JSON.stringify(params.filter)
    if (params?.sort) queryParams.sort = JSON.stringify(params.sort)
    
    const response = await axios.get(`${API_BASE_URL}/api/v1/views/${name}/query`, {
      headers: getAuthHeaders(),
      params: queryParams,
    })
    return response.data
  },

  // Update a view
  update: async (name: string, data: {
    description?: string
    filter?: Record<string, any>
    fields?: string[]
  }) => {
    const response = await axios.put(`${API_BASE_URL}/api/v1/views/${name}`, data, {
      headers: getAuthHeaders(),
    })
    return response.data
  },

  // Delete a view
  delete: async (name: string) => {
    const response = await axios.delete(`${API_BASE_URL}/api/v1/views/${name}`, {
      headers: getAuthHeaders(),
    })
    return response.data
  },
}