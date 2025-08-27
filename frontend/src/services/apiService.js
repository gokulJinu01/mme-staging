import axios from 'axios';
import toast from 'react-hot-toast';

// Create axios instance for Tagging Service
const taggingApi = axios.create({
  baseURL: process.env.REACT_APP_TAGGING_API_URL || 'http://localhost:8081',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Create axios instance for Tagmaker Service
const tagmakerApi = axios.create({
  baseURL: process.env.REACT_APP_TAGMAKER_API_URL || 'http://localhost:8000',
  timeout: 10000,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Request interceptor to add auth headers
const addAuthHeaders = (config) => {
  const token = localStorage.getItem('mme_token');
  const userId = localStorage.getItem('mme_user_id');
  
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  
  if (userId) {
    config.headers['X-User-ID'] = userId;
  }
  
  return config;
};

taggingApi.interceptors.request.use(addAuthHeaders, (error) => Promise.reject(error));
tagmakerApi.interceptors.request.use(addAuthHeaders, (error) => Promise.reject(error));

// Response interceptor for error handling
const handleResponse = (response) => response;
const handleError = (error) => {
  if (error.response?.status === 401) {
    localStorage.removeItem('mme_token');
    localStorage.removeItem('mme_user_id');
    window.location.href = '/login';
  }
  
  const message = error.response?.data?.error?.message || error.response?.data?.message || error.message || 'An error occurred';
  toast.error(message);
  
  return Promise.reject(error);
};

taggingApi.interceptors.response.use(handleResponse, handleError);
tagmakerApi.interceptors.response.use(handleResponse, handleError);

// ===== TAGGING SERVICE ENDPOINTS =====

// System endpoints
export const getTaggingHealth = async () => {
  const response = await taggingApi.get('/health');
  return response.data;
};

export const getTaggingMetrics = async () => {
  const response = await taggingApi.get('/metrics');
  return response.data;
};

// Memory Management (9 endpoints)
export const saveMemory = async (memoryData) => {
  const response = await taggingApi.post('/memory/save', memoryData);
  return response.data;
};

export const queryMemories = async (params) => {
  const response = await taggingApi.get('/memory/query', { params });
  return response.data;
};

export const queryMemoriesDebug = async (params) => {
  const response = await taggingApi.get('/memory/query-debug', { params });
  return response.data;
};

export const getRecentMemories = async (params) => {
  const response = await taggingApi.get('/memory/recent', { params });
  return response.data;
};

export const deleteMemory = async (id) => {
  const response = await taggingApi.delete(`/memory/${id}`);
  return response.data;
};

export const promoteMemories = async (promoteData) => {
  const response = await taggingApi.post('/memory/promote', promoteData);
  return response.data;
};

export const injectMemory = async (memoryData) => {
  const response = await taggingApi.post('/memory/inject', memoryData);
  return response.data;
};

export const getTokenizerHealth = async () => {
  const response = await taggingApi.get('/memory/tokenizer-health');
  return response.data;
};

// Tag Operations (3 endpoints)
export const extractTags = async (extractData) => {
  const response = await taggingApi.post('/tags/extract', extractData);
  return response.data;
};

export const queryByPrompt = async (promptData) => {
  const response = await taggingApi.post('/tags/query', promptData);
  return response.data;
};

export const updateTagDelta = async (deltaData) => {
  const response = await taggingApi.post('/tags/delta', deltaData);
  return response.data;
};

// Search Operations (1 endpoint)
export const semanticSearch = async (searchData) => {
  const response = await taggingApi.post('/search/semantic', searchData);
  return response.data;
};

// Events Processing (1 endpoint)
export const handlePackEvent = async (eventData) => {
  const response = await taggingApi.post('/events/pack', eventData);
  return response.data;
};

// Background Processing (1 endpoint)
export const backfillTags = async (backfillData) => {
  const response = await taggingApi.post('/processing/backfill-tags', backfillData);
  return response.data;
};

// Admin Operations (5 endpoints)
export const getAdminStats = async () => {
  const response = await taggingApi.get('/admin/stats');
  return response.data;
};

export const cleanupMemories = async (cleanupData) => {
  const response = await taggingApi.post('/admin/cleanup', cleanupData);
  return response.data;
};

export const pruneEdges = async (pruneData) => {
  const response = await taggingApi.post('/admin/edges/prune', pruneData);
  return response.data;
};

export const getAdminFeatures = async () => {
  const response = await taggingApi.get('/admin/features');
  return response.data;
};

export const setAdminFeatures = async (featuresData) => {
  const response = await taggingApi.post('/admin/features', featuresData);
  return response.data;
};

// Security endpoints
export const getSecurityHealth = async () => {
  const response = await taggingApi.get('/security/health');
  return response.data;
};

export const getSecurityMetrics = async () => {
  const response = await taggingApi.get('/security/metrics');
  return response.data;
};

// ===== TAGMAKER SERVICE ENDPOINTS =====

// System endpoints
export const getTagmakerHealth = async () => {
  const response = await tagmakerApi.get('/health');
  return response.data;
};

export const getTagmakerMetrics = async () => {
  const response = await tagmakerApi.get('/metrics');
  return response.data;
};

export const getTagmakerDocs = async () => {
  const response = await tagmakerApi.get('/docs');
  return response.data;
};

export const getTagmakerReDoc = async () => {
  const response = await tagmakerApi.get('/redoc');
  return response.data;
};

export const getTagmakerOpenAPI = async () => {
  const response = await tagmakerApi.get('/openapi.json');
  return response.data;
};

// Business endpoints (7 endpoints)
export const getTagmakerStatus = async () => {
  const response = await tagmakerApi.get('/');
  return response.data;
};

export const getQueueStatus = async () => {
  const response = await tagmakerApi.get('/queue-status');
  return response.data;
};

export const getDatabaseStatus = async () => {
  const response = await tagmakerApi.get('/database-status');
  return response.data;
};

export const getDatabaseDebug = async () => {
  const response = await tagmakerApi.get('/database-debug');
  return response.data;
};

export const testDatabaseDirect = async () => {
  const response = await tagmakerApi.get('/database-direct-test');
  return response.data;
};

export const manualRebalance = async (rebalanceData) => {
  const response = await tagmakerApi.post('/manual-rebalance', rebalanceData);
  return response.data;
};

export const extractTagsFromContent = async (extractData) => {
  const response = await tagmakerApi.post('/extract-tags', extractData);
  return response.data;
};

export const generateAndSave = async (generateData) => {
  const response = await tagmakerApi.post('/generate-and-save', generateData);
  return response.data;
};

// Tagmaker Security endpoints
export const getTagmakerSecurityHealth = async () => {
  const response = await tagmakerApi.get('/security/health');
  return response.data;
};

export const getTagmakerSecurityMetrics = async () => {
  const response = await tagmakerApi.get('/security/metrics');
  return response.data;
};

// Tagmaker Admin endpoints
export const triggerEdgeLearning = async (learningData) => {
  const response = await tagmakerApi.post('/admin/edge-learn/replay', learningData);
  return response.data;
};

// Tagmaker Utility endpoints
export const submitFeedback = async (feedbackData) => {
  const response = await tagmakerApi.post('/feedback', feedbackData);
  return response.data;
};

export const getTagmakerVersion = async () => {
  const response = await tagmakerApi.get('/version');
  return response.data;
};

// ===== AUTHENTICATION =====

export const login = async (credentials) => {
  // For demo purposes, simulate successful login
  return Promise.resolve({ token: 'demo-token', user: credentials });
};

export const logout = async () => {
  // For demo purposes, simulate successful logout
  return Promise.resolve({ success: true });
};

// ===== API SERVICE OBJECT =====

export const apiService = {
  // Tagging Service
  getTaggingHealth,
  getTaggingMetrics,
  saveMemory,
  queryMemories,
  queryMemoriesDebug,
  getRecentMemories,
  deleteMemory,
  promoteMemories,
  injectMemory,
  getTokenizerHealth,
  extractTags,
  queryByPrompt,
  updateTagDelta,
  semanticSearch,
  handlePackEvent,
  backfillTags,
  getAdminStats,
  cleanupMemories,
  pruneEdges,
  getAdminFeatures,
  setAdminFeatures,
  getSecurityHealth,
  getSecurityMetrics,
  
  // Tagmaker Service
  getTagmakerHealth,
  getTagmakerMetrics,
  getTagmakerDocs,
  getTagmakerReDoc,
  getTagmakerOpenAPI,
  getTagmakerStatus,
  getQueueStatus,
  getDatabaseStatus,
  getDatabaseDebug,
  testDatabaseDirect,
  manualRebalance,
  extractTagsFromContent,
  generateAndSave,
  getTagmakerSecurityHealth,
  getTagmakerSecurityMetrics,
  triggerEdgeLearning,
  submitFeedback,
  getTagmakerVersion,
  
  // Authentication
  login,
  logout,
  
  // Legacy compatibility
  getHealth: getTaggingHealth,
  querySearch: semanticSearch,
  queryBlocksDebug: queryMemoriesDebug,
  getSystemInfo: getAdminStats,
};

export default apiService;
