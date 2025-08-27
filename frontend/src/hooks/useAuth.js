import { useState, useEffect, createContext, useContext } from 'react';
import { apiService } from '../services/apiService';

const AuthContext = createContext();

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

export const AuthProvider = ({ children }) => {
  const [user, setUser] = useState(null);
  const [isLoading, setIsLoading] = useState(true);
  const [isAuthenticated, setIsAuthenticated] = useState(false);

  // Check if user is authenticated on mount
  useEffect(() => {
    const checkAuth = async () => {
      try {
        const token = localStorage.getItem('mme_token');
        const userId = localStorage.getItem('mme_user_id');
        
        if (token && userId) {
          // Verify token is still valid by making a health check
          await apiService.getHealth();
          
          setUser({
            id: userId,
            name: localStorage.getItem('mme_user_name') || 'User',
            email: localStorage.getItem('mme_user_email') || '',
            role: localStorage.getItem('mme_user_role') || 'user',
            orgId: localStorage.getItem('mme_org_id') || '',
          });
          setIsAuthenticated(true);
        }
      } catch (error) {
        console.error('Auth check failed:', error);
        // Clear invalid auth data
        localStorage.removeItem('mme_token');
        localStorage.removeItem('mme_user_id');
        localStorage.removeItem('mme_user_name');
        localStorage.removeItem('mme_user_email');
        localStorage.removeItem('mme_user_role');
        localStorage.removeItem('mme_org_id');
      } finally {
        setIsLoading(false);
      }
    };

    checkAuth();
  }, []);

  const login = async (credentials) => {
    try {
      setIsLoading(true);
      
      // For demo purposes, we'll simulate login with the provided credentials
      // In production, this would call the actual auth endpoint
      const response = await apiService.login(credentials);
      
      // Store auth data
      localStorage.setItem('mme_token', response.token || 'demo-token');
      localStorage.setItem('mme_user_id', credentials.userId || 'demo-user');
      localStorage.setItem('mme_user_name', credentials.name || 'Demo User');
      localStorage.setItem('mme_user_email', credentials.email || 'demo@example.com');
      localStorage.setItem('mme_user_role', credentials.role || 'user');
      localStorage.setItem('mme_org_id', credentials.orgId || 'demo-org');
      
      setUser({
        id: credentials.userId || 'demo-user',
        name: credentials.name || 'Demo User',
        email: credentials.email || 'demo@example.com',
        role: credentials.role || 'user',
        orgId: credentials.orgId || 'demo-org',
      });
      
      setIsAuthenticated(true);
      return response;
    } catch (error) {
      throw error;
    } finally {
      setIsLoading(false);
    }
  };

  const logout = async () => {
    try {
      await apiService.logout();
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      // Clear auth data
      localStorage.removeItem('mme_token');
      localStorage.removeItem('mme_user_id');
      localStorage.removeItem('mme_user_name');
      localStorage.removeItem('mme_user_email');
      localStorage.removeItem('mme_user_role');
      localStorage.removeItem('mme_org_id');
      
      setUser(null);
      setIsAuthenticated(false);
    }
  };

  const updateUser = (updates) => {
    setUser(prev => ({ ...prev, ...updates }));
  };

  const hasRole = (role) => {
    return user?.role === role;
  };

  const hasPermission = (permission) => {
    // Simple permission check based on role
    const permissions = {
      user: ['read', 'write'],
      admin: ['read', 'write', 'delete', 'manage'],
      system: ['read', 'write', 'delete', 'manage', 'system'],
    };
    
    return permissions[user?.role]?.includes(permission) || false;
  };

  const value = {
    user,
    isAuthenticated,
    isLoading,
    login,
    logout,
    updateUser,
    hasRole,
    hasPermission,
  };

  return (
    <AuthContext.Provider value={value}>
      {children}
    </AuthContext.Provider>
  );
};

export default useAuth;
