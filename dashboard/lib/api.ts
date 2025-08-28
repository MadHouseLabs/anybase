import axios from 'axios';
import Cookies from 'js-cookie';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL ? `${process.env.NEXT_PUBLIC_API_URL}/api/v1` : '/api/v1';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Add auth token to requests
api.interceptors.request.use((config) => {
  const token = Cookies.get('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// Handle auth errors
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      Cookies.remove('token');
      Cookies.remove('user');
      if (typeof window !== 'undefined') {
        window.location.href = '/login';
      }
    }
    return Promise.reject(error);
  }
);

// Auth APIs
export const authApi = {
  login: async (email: string, password: string) => {
    const response = await api.post('/auth/login', { email, password });
    return response.data;
  },
  register: async (data: any) => {
    const response = await api.post('/auth/register', data);
    return response.data;
  },
  logout: async () => {
    await api.post('/auth/logout');
    Cookies.remove('token');
    Cookies.remove('user');
  },
};

// Collections APIs
export const collectionsApi = {
  list: async () => {
    const response = await api.get('/collections');
    return response.data;
  },
  get: async (name: string) => {
    const response = await api.get(`/collections/${name}`);
    return response.data;
  },
  create: async (data: any) => {
    const response = await api.post('/collections', data);
    return response.data;
  },
  update: async (name: string, data: any) => {
    const response = await api.put(`/collections/${name}`, data);
    return response.data;
  },
  delete: async (name: string) => {
    const response = await api.delete(`/collections/${name}`);
    return response.data;
  },
  getIndexes: async (name: string) => {
    const response = await api.get(`/collections/${name}/indexes`);
    return response.data;
  },
  createIndex: async (name: string, data: any) => {
    const response = await api.post(`/collections/${name}/indexes`, data);
    return response.data;
  },
  deleteIndex: async (collectionName: string, indexName: string) => {
    const response = await api.delete(`/collections/${collectionName}/indexes/${indexName}`);
    return response.data;
  },
};

// Data APIs
export const dataApi = {
  query: async (collection: string, params?: any) => {
    const response = await api.get(`/data/${collection}`, { params });
    return response.data;
  },
  get: async (collection: string, id: string) => {
    const response = await api.get(`/data/${collection}/${id}`);
    return response.data;
  },
  create: async (collection: string, data: any) => {
    const response = await api.post(`/data/${collection}`, data);
    return response.data;
  },
  update: async (collection: string, id: string, data: any) => {
    const response = await api.put(`/data/${collection}/${id}`, data);
    return response.data;
  },
  delete: async (collection: string, id: string) => {
    const response = await api.delete(`/data/${collection}/${id}`);
    return response.data;
  },
};

// Documents API (alias for dataApi with better naming)
export const documentsApi = {
  list: async (collection: string, params?: any) => {
    const response = await api.get(`/data/${collection}`, { params });
    return response.data;
  },
  get: async (collection: string, id: string) => {
    return dataApi.get(collection, id);
  },
  create: async (collection: string, data: any) => {
    return dataApi.create(collection, data);
  },
  update: async (collection: string, id: string, data: any) => {
    return dataApi.update(collection, id, data);
  },
  delete: async (collection: string, id: string) => {
    return dataApi.delete(collection, id);
  },
};

// Views APIs
export const viewsApi = {
  list: async () => {
    const response = await api.get('/views');
    return response.data;
  },
  create: async (data: any) => {
    const response = await api.post('/views', data);
    return response.data;
  },
  query: async (name: string) => {
    const response = await api.get(`/views/${name}/query`);
    return response.data;
  },
};


// Users APIs
export const usersApi = {
  list: async () => {
    const response = await api.get('/admin/users');
    return response.data;
  },
  create: async (data: any) => {
    const response = await api.post('/admin/users', data);
    return response.data;
  },
  get: async (id: string) => {
    const response = await api.get(`/admin/users/${id}`);
    return response.data;
  },
  update: async (id: string, data: any) => {
    const response = await api.put(`/admin/users/${id}`, data);
    return response.data;
  },
  delete: async (id: string) => {
    const response = await api.delete(`/admin/users/${id}`);
    return response.data;
  },
};

export default api;