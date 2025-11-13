import axios from 'axios';

const API_BASE_URL = 'http://localhost:8080';

const api = axios.create({
  baseURL: API_BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Добавляем токен к запросам
api.interceptors.request.use(
  (config) => {
    // Список публичных endpoints, которым не нужен токен
    const publicEndpoints = ['/password/reset', '/password/reset/confirm', '/login', '/register'];
    const isPublicEndpoint = publicEndpoints.some(endpoint => config.url?.includes(endpoint));
    
    // Добавляем токен только если endpoint не публичный и токен есть
    const token = localStorage.getItem('token');
    if (token && !isPublicEndpoint) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    
    console.log('📤 API Request:', config.method.toUpperCase(), config.url);
    console.log('   Headers:', config.headers);
    return config;
  },
  (error) => {
    console.error('❌ Request error:', error);
    return Promise.reject(error);
  }
);

// Обработка ошибок ответа
api.interceptors.response.use(
  (response) => {
    console.log('✅ API Response:', response.status, response.config.url);
    return response;
  },
  (error) => {
    console.error('❌ API Error:', error.response?.status, error.config?.url);
    console.error('   Message:', error.message);
    console.error('   Data:', error.response?.data);
    
    // Не делаем редирект на логин для публичных endpoints (восстановление пароля)
    const publicEndpoints = ['/password/reset', '/password/reset/confirm'];
    const isPublicEndpoint = publicEndpoints.some(endpoint => 
      error.config?.url?.includes(endpoint)
    );
    
    if (error.response?.status === 401 && !isPublicEndpoint) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
      window.location.href = '/login';
    }
    return Promise.reject(error);
  }
);

// Auth API
export const authAPI = {
  register: (data) => api.post('/register', data),
  login: (data) => api.post('/login', data),
};

// Products API
export const productsAPI = {
  getAll: (page = 1, limit = 20) => api.get('/products', { params: { page, limit } }),
  getById: (id) => api.get(`/products/${id}`),
  getSizes: (productId) => api.get(`/products/${productId}/sizes`),
};

// Brands API
export const brandsAPI = {
  getAll: () => api.get('/brands'),
  getById: (id) => api.get(`/brands/${id}`),
};

// Categories API
export const categoriesAPI = {
  getAll: () => api.get('/categories'),
  getById: (id) => api.get(`/categories/${id}`),
};

// Basket API
export const basketAPI = {
  getByUserId: (userId) => api.get(`/basket/${userId}`),
  add: (data) => api.post('/basket', data),
  update: (id, data) => api.put(`/basket/${id}`, data),
  delete: (id) => api.delete(`/basket/${id}`),
};

// Favorites API
export const favoritesAPI = {
  getByUserId: (userId) => api.get(`/favorites/${userId}`),
  add: (data) => api.post('/favorites', data),
  delete: (id) => api.delete(`/favorites/${id}`),
};

// Orders API
export const ordersAPI = {
  create: (data) => api.post('/orders', data),
  addProduct: (data) => api.post('/order-products', data),
  getByUserId: (userId, page = 1, limit = 20) => api.get(`/orders/user/${userId}`, { params: { page, limit } }),
  getProductsByOrderId: (orderId) => api.get(`/order-products/order/${orderId}`),
  getAll: (page = 1, limit = 20) => api.get('/orders', { params: { page, limit } }), // для менеджера (если доступно)
};

// User API
export const userAPI = {
  update: (id, data) => api.put(`/users/${id}`, data),
  getById: (id) => api.get(`/users/${id}`),
};

// Password API
export const passwordAPI = {
  change: (data) => api.post('/password/change', data),
  confirmChange: (data) => api.post('/password/change/confirm', data),
  reset: (data) => api.post('/password/reset', data),
  confirmReset: (data) => api.post('/password/reset/confirm', data),
};

// Reviews API
export const reviewsAPI = {
  getByProductId: (productId, page = 1, limit = 20) => api.get(`/reviews/product/${productId}`, { params: { page, limit } }),
  getByUserId: (userId, page = 1, limit = 20) => api.get(`/reviews/user/${userId}`, { params: { page, limit } }),
  create: (data) => api.post('/reviews', data),
  update: (id, data) => api.put(`/reviews/${id}`, data),
  delete: (id) => api.delete(`/reviews/${id}`),
};

// Reports API
export const reportsAPI = {
  getAll: () => api.get('/reports'),
  getById: (id) => api.get(`/reports/${id}`),
  create: (data) => api.post('/reports', data),
  // PDF Reports
  generateSalesPDF: () => api.get('/reports/sales/pdf', { responseType: 'blob' }),
  generateInventoryPDF: () => api.get('/reports/inventory/pdf', { responseType: 'blob' }),
  generateCustomersPDF: () => api.get('/reports/customers/pdf', { responseType: 'blob' }),
  generateCategoriesPDF: () => api.get('/reports/categories/pdf', { responseType: 'blob' }),
  // Excel Reports
  generateSalesExcel: () => api.get('/reports/sales/excel', { responseType: 'blob' }),
  generateInventoryExcel: () => api.get('/reports/inventory/excel', { responseType: 'blob' }),
  generateCustomersExcel: () => api.get('/reports/customers/excel', { responseType: 'blob' }),
  generateCategoriesExcel: () => api.get('/reports/categories/excel', { responseType: 'blob' }),
  // Text Reports
  generateCustomersText: () => api.get('/reports/customers/text', { responseType: 'blob' }),
  generateInventoryText: () => api.get('/reports/inventory/text', { responseType: 'blob' }),
};

// Admin API
export const adminAPI = {
  // Products
  products: {
    getAll: (page = 1, limit = 20) => api.get('/admin/products', { params: { page, limit } }),
    getById: (id) => api.get(`/admin/products/${id}`),
    create: (data) => api.post('/admin/products', data),
    update: (id, data) => api.put(`/admin/products/${id}`, data),
    delete: (id) => api.delete(`/admin/products/${id}`),
  },
  // Brands
  brands: {
    getAll: () => api.get('/admin/brands'),
    getById: (id) => api.get(`/admin/brands/${id}`),
    create: (data) => api.post('/admin/brands', data),
    update: (id, data) => api.put(`/admin/brands/${id}`, data),
    delete: (id) => api.delete(`/admin/brands/${id}`),
  },
  // Categories
  categories: {
    getAll: () => api.get('/admin/categories'),
    getById: (id) => api.get(`/admin/categories/${id}`),
    create: (data) => api.post('/admin/categories', data),
    update: (id, data) => api.put(`/admin/categories/${id}`, data),
    delete: (id) => api.delete(`/admin/categories/${id}`),
  },
  // Users
  users: {
    getAll: (page = 1, limit = 20) => api.get('/admin/users', { params: { page, limit } }),
    getById: (id) => api.get(`/admin/users/${id}`),
    create: (data) => api.post('/admin/users', data),
    update: (id, data) => api.put(`/admin/users/${id}`, data),
    delete: (id) => api.delete(`/admin/users/${id}`),
  },
  // Reviews
  reviews: {
    getAll: (page = 1, limit = 20) => api.get('/admin/reviews', { params: { page, limit } }),
    getById: (id) => api.get(`/admin/reviews/${id}`),
    update: (id, data) => api.put(`/admin/reviews/${id}`, data),
    delete: (id) => api.delete(`/admin/reviews/${id}`),
  },
  // Logs
  logs: {
    getAll: (page = 1, limit = 20) => api.get('/admin/logs', { params: { page, limit } }),
    getById: (id) => api.get(`/admin/logs/${id}`),
    delete: (id) => api.delete(`/admin/logs/${id}`),
  },
  // Backup
  backup: {
    create: () => api.post('/admin/backup'),
    getInfo: () => api.get('/admin/backup/info'),
    delete: (filename) => api.delete(`/admin/backup/${filename}`),
    download: (filename) => api.get(`/admin/backup/download/${filename}`, { responseType: 'blob' }),
    restore: (file) => {
      const formData = new FormData();
      formData.append('file', file);
      return api.post('/admin/backup/restore', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });
    },
  },
  // Orders
  orders: {
    getAll: (page = 1, limit = 20) => api.get('/orders', { params: { page, limit } }),
    getById: (id) => api.get(`/orders/${id}`),
  },
  // Order Products
  orderProducts: {
    getAll: () => api.get('/order-products'),
    getByOrderId: (orderId) => api.get(`/order-products/order/${orderId}`),
    update: (id, data) => api.put(`/order-products/${id}`, data),
  },
};

export default api;

