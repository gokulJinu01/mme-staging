import React, { useState, useEffect } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { Helmet } from 'react-helmet-async';
import { useQuery } from '@tanstack/react-query';
import toast from 'react-hot-toast';

// Layout Components
import Sidebar from './components/Layout/Sidebar';
import Header from './components/Layout/Header';

// Authentication
import LoginPage from './pages/Auth/LoginPage';
import { useAuth } from './hooks/useAuth';

// Core User Pages
import DashboardPage from './pages/Dashboard/DashboardPage';
import MemoryManagementPage from './pages/Memory/MemoryManagementPage';
import MemoryPromotionPage from './pages/Memory/MemoryPromotionPage';
import TagManagementPage from './pages/Tags/TagManagementPage';
import SearchQueryPage from './pages/Search/SearchQueryPage';
import AnalyticsPage from './pages/Analytics/AnalyticsPage';

// Admin Pages
import AdminDashboardPage from './pages/Admin/AdminDashboardPage';
import SystemManagementPage from './pages/Admin/SystemManagementPage';
import MonitoringPage from './pages/Admin/MonitoringPage';

// Settings
import SettingsPage from './pages/Settings/SettingsPage';

// API Service
import { apiService } from './services/apiService';

// Loading Component
const LoadingSpinner = () => (
  <div className="flex items-center justify-center min-h-screen">
    <div className="loading-spinner"></div>
    <span className="ml-2 text-gray-600">Loading MME...</span>
  </div>
);

function App() {
  const { user, isAuthenticated, isLoading } = useAuth();
  const [sidebarOpen, setSidebarOpen] = useState(false);

  // Health check
  const { data: healthStatus } = useQuery(
    'health',
    () => apiService.getHealth(),
    {
      refetchInterval: 30000, // Check every 30 seconds
      retry: 3,
      onError: (error) => {
        console.error('Health check failed:', error);
        toast.error('Connection to MME service lost');
      },
    }
  );

  // Close sidebar on route change
  useEffect(() => {
    setSidebarOpen(false);
  }, []);

  if (isLoading) {
    return <LoadingSpinner />;
  }

  // Public routes (no authentication required)
  if (!isAuthenticated) {
    return (
      <>
        <Helmet>
          <title>Login - MME</title>
        </Helmet>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
          <Route path="*" element={<Navigate to="/login" replace />} />
        </Routes>
      </>
    );
  }

  // Protected routes (authentication required)
  return (
    <>
      <Helmet>
        <title>MME Packer - Memory Management Engine</title>
      </Helmet>
      
      <div className="flex h-screen bg-gray-50">
        {/* Sidebar */}
        <Sidebar 
          isOpen={sidebarOpen} 
          onClose={() => setSidebarOpen(false)}
          user={user}
        />
        
        {/* Main Content */}
        <div className="flex-1 flex flex-col overflow-hidden">
          <Header onMenuClick={() => setSidebarOpen(true)} user={user} healthStatus={healthStatus} />
          <main className="flex-1 overflow-x-hidden overflow-y-auto bg-gray-50">
            <Routes>
              {/* Core User Routes */}
              <Route path="/" element={<DashboardPage />} />
              <Route path="/dashboard" element={<DashboardPage />} />
              <Route path="/memory" element={<MemoryManagementPage />} />
              <Route path="/promote" element={<MemoryPromotionPage />} />
              <Route path="/tags" element={<TagManagementPage />} />
              <Route path="/search" element={<SearchQueryPage />} />
              <Route path="/analytics" element={<AnalyticsPage />} />
              
              {/* Admin Routes */}
              <Route path="/admin" element={<AdminDashboardPage />} />
              <Route path="/admin/system" element={<SystemManagementPage />} />
              <Route path="/admin/monitoring" element={<MonitoringPage />} />
              
              {/* Settings */}
              <Route path="/settings" element={<SettingsPage />} />
              
              {/* Fallback */}
              <Route path="*" element={<Navigate to="/" replace />} />
            </Routes>
          </main>
        </div>
      </div>
    </>
  );
}

export default App;
