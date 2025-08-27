import React, { useState } from 'react';
import { Helmet } from 'react-helmet-async';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { FiBarChart2, FiSettings, FiRefreshCw, FiDatabase, FiActivity } from 'react-icons/fi';
import toast from 'react-hot-toast';

import { apiService } from '../../services/apiService';

const AdminDashboardPage = () => {
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState('overview');
  const [cleanupParams, setCleanupParams] = useState({
    olderThan: 30,
    maxMemories: 1000
  });
  const [pruneParams, setPruneParams] = useState({
    orgId: '',
    tau: 0.1,
    maxEdges: 1000
  });

  // Queries
  const { data: adminStats, isLoading: statsLoading, refetch: refetchStats } = useQuery(
    ['admin-stats'],
    () => apiService.getAdminStats(),
    {
      refetchInterval: 60000, // Refresh every minute
    }
  );

  const { data: adminFeatures, isLoading: featuresLoading } = useQuery(
    ['admin-features'],
    () => apiService.getAdminFeatures(),
    {
      refetchInterval: 300000, // Refresh every 5 minutes
    }
  );

  const { data: securityHealth } = useQuery(
    ['security-health'],
    () => apiService.getSecurityHealth(),
    {
      refetchInterval: 30000,
    }
  );

  const { data: securityMetrics } = useQuery(
    ['security-metrics'],
    () => apiService.getSecurityMetrics(),
    {
      refetchInterval: 60000,
    }
  );

  // Mutations
  const cleanupMutation = useMutation({
    mutationFn: apiService.cleanupMemories,
    onSuccess: (data) => {
      toast.success(`Cleanup completed: ${data.data?.cleaned_blocks || 0} blocks, ${data.data?.cleaned_tags || 0} tags`);
      queryClient.invalidateQueries(['admin-stats']);
    },
    onError: (error) => {
      toast.error(`Cleanup failed: ${error.message}`);
    },
  });

  const pruneMutation = useMutation({
    mutationFn: apiService.pruneEdges,
    onSuccess: (data) => {
      toast.success(`Edge pruning completed: ${data.data?.pruned_count || 0} edges pruned`);
      queryClient.invalidateQueries(['admin-stats']);
    },
    onError: (error) => {
      toast.error(`Edge pruning failed: ${error.message}`);
    },
  });

  const setFeaturesMutation = useMutation({
    mutationFn: apiService.setAdminFeatures,
    onSuccess: () => {
      toast.success('Features updated successfully');
      queryClient.invalidateQueries(['admin-features']);
    },
    onError: (error) => {
      toast.error(`Failed to update features: ${error.message}`);
    },
  });

  const backfillTagsMutation = useMutation({
    mutationFn: apiService.backfillTags,
    onSuccess: (data) => {
      toast.success(`Backfill completed: ${data.data?.backfilled || 0} tags processed`);
      queryClient.invalidateQueries(['admin-stats']);
    },
    onError: (error) => {
      toast.error(`Backfill failed: ${error.message}`);
    },
  });

  // Handlers
  const handleCleanup = () => {
    if (window.confirm('Are you sure you want to perform cleanup? This action cannot be undone.')) {
      cleanupMutation.mutate(cleanupParams);
    }
  };

  const handlePrune = () => {
    if (!pruneParams.orgId) {
      toast.error('Organization ID is required');
      return;
    }
    if (window.confirm('Are you sure you want to prune edges? This action cannot be undone.')) {
      pruneMutation.mutate(pruneParams);
    }
  };

  const handleBackfill = () => {
    if (window.confirm('Are you sure you want to backfill tags? This may take some time.')) {
      backfillTagsMutation.mutate({});
    }
  };

  const renderStatCard = (title, value, icon, color = 'blue') => (
    <div className={`bg-white rounded-lg shadow-md p-6 border-l-4 border-${color}-500`}>
      <div className="flex items-center">
        <div className={`p-3 rounded-full bg-${color}-100 text-${color}-600`}>
          {icon}
        </div>
        <div className="ml-4">
          <p className="text-sm font-medium text-gray-600">{title}</p>
          <p className="text-2xl font-semibold text-gray-900">{value}</p>
        </div>
      </div>
    </div>
  );

  const renderFeatureToggle = (feature, enabled) => (
    <div key={feature} className="flex items-center justify-between p-4 bg-white rounded-lg shadow-sm border">
      <div>
        <h3 className="font-medium text-gray-900">{feature}</h3>
        <p className="text-sm text-gray-500">Feature toggle for {feature}</p>
      </div>
      <button
        onClick={() => setFeaturesMutation.mutate({ [feature]: !enabled })}
        disabled={setFeaturesMutation.isLoading}
        className={`relative inline-flex h-6 w-11 items-center rounded-full transition-colors ${
          enabled ? 'bg-blue-600' : 'bg-gray-200'
        }`}
      >
        <span
          className={`inline-block h-4 w-4 transform rounded-full bg-white transition-transform ${
            enabled ? 'translate-x-6' : 'translate-x-1'
          }`}
        />
      </button>
    </div>
  );

  return (
    <>
      <Helmet>
        <title>Admin Dashboard - MME</title>
      </Helmet>
      
      <div className="p-6">
        <div className="mb-6">
          <h1 className="text-2xl font-bold text-gray-900">Admin Dashboard</h1>
          <p className="text-gray-600 mt-1">System administration and monitoring</p>
        </div>

        {/* Tab Navigation */}
        <div className="mb-6">
          <nav className="flex space-x-8">
            <button
              onClick={() => setActiveTab('overview')}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'overview'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Overview
            </button>
            <button
              onClick={() => setActiveTab('maintenance')}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'maintenance'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Maintenance
            </button>
            <button
              onClick={() => setActiveTab('security')}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'security'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Security
            </button>
            <button
              onClick={() => setActiveTab('features')}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'features'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Features
            </button>
          </nav>
        </div>
        
        {/* Tab Content */}
        {activeTab === 'overview' && (
          <div>
            {/* Stats Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
              {statsLoading ? (
                <div className="col-span-full text-center py-8">
                  <div className="loading-spinner mx-auto mb-4"></div>
                  <p className="text-gray-600">Loading statistics...</p>
                </div>
              ) : adminStats?.data ? (
                <>
                  {renderStatCard(
                    'Total Memory Blocks',
                    adminStats.data.total_memory_blocks?.toLocaleString() || '0',
                    <FiDatabase size={24} />,
                    'blue'
                  )}
                  {renderStatCard(
                    'Total Tags',
                    adminStats.data.total_tags?.toLocaleString() || '0',
                    <FiActivity size={24} />,
                    'green'
                  )}
                                     {renderStatCard(
                     'Uptime',
                     `${Math.floor((adminStats.data.uptime_seconds || 0) / 3600)}h`,
                     <FiBarChart2 size={24} />,
                     'purple'
                   )}
                  {renderStatCard(
                    'Memory Usage',
                    `${adminStats.data.memory_usage_mb?.toFixed(1) || '0'} MB`,
                    <FiSettings size={24} />,
                    'orange'
                  )}
                </>
              ) : (
                <div className="col-span-full text-center py-8">
                  <FiBarChart2 size={48} className="mx-auto text-gray-400 mb-4" />
                  <p className="text-gray-600">No statistics available</p>
                </div>
              )}
            </div>

            {/* Recent Activity */}
            <div className="bg-white rounded-lg shadow-md p-6">
              <div className="flex justify-between items-center mb-4">
                <h3 className="text-lg font-semibold">Recent Activity</h3>
                <button
                  onClick={() => refetchStats()}
                  className="p-2 text-gray-400 hover:text-gray-600"
                >
                  <FiRefreshCw size={20} />
                </button>
              </div>
              
              {adminStats?.data ? (
                <div className="space-y-3">
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-600">Last Cleanup:</span>
                    <span className="text-gray-900">
                      {adminStats.data.last_cleanup_time 
                        ? new Date(adminStats.data.last_cleanup_time).toLocaleString()
                        : 'Never'
                      }
                    </span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-600">Requests per Minute:</span>
                    <span className="text-gray-900">{adminStats.data.requests_per_minute || '0'}</span>
                  </div>
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-600">Timestamp:</span>
                    <span className="text-gray-900">
                      {adminStats.data.timestamp 
                        ? new Date(adminStats.data.timestamp).toLocaleString()
                        : 'Unknown'
                      }
                    </span>
                  </div>
                </div>
              ) : (
                <p className="text-gray-600">No activity data available</p>
              )}
            </div>
          </div>
        )}

        {activeTab === 'maintenance' && (
          <div className="space-y-6">
            {/* Memory Cleanup */}
            <div className="bg-white rounded-lg shadow-md p-6">
              <h3 className="text-lg font-semibold mb-4">Memory Cleanup</h3>
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 mb-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Older Than (days)
                  </label>
                  <input
                    type="number"
                    value={cleanupParams.olderThan}
                    onChange={(e) => setCleanupParams(prev => ({ ...prev, olderThan: parseInt(e.target.value) }))}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Max Memories
                  </label>
                  <input
                    type="number"
                    value={cleanupParams.maxMemories}
                    onChange={(e) => setCleanupParams(prev => ({ ...prev, maxMemories: parseInt(e.target.value) }))}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
              </div>
              <button
                onClick={handleCleanup}
                disabled={cleanupMutation.isLoading}
                className="bg-red-600 text-white py-2 px-4 rounded-lg hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {cleanupMutation.isLoading ? 'Cleaning...' : 'Perform Cleanup'}
              </button>
            </div>

            {/* Edge Pruning */}
            <div className="bg-white rounded-lg shadow-md p-6">
              <h3 className="text-lg font-semibold mb-4">Edge Pruning</h3>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Organization ID
                  </label>
                  <input
                    type="text"
                    value={pruneParams.orgId}
                    onChange={(e) => setPruneParams(prev => ({ ...prev, orgId: e.target.value }))}
                    placeholder="Enter org ID"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Tau Threshold
                  </label>
                  <input
                    type="number"
                    step="0.01"
                    value={pruneParams.tau}
                    onChange={(e) => setPruneParams(prev => ({ ...prev, tau: parseFloat(e.target.value) }))}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Max Edges
                  </label>
                  <input
                    type="number"
                    value={pruneParams.maxEdges}
                    onChange={(e) => setPruneParams(prev => ({ ...prev, maxEdges: parseInt(e.target.value) }))}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>
              </div>
              <button
                onClick={handlePrune}
                disabled={pruneMutation.isLoading}
                className="bg-orange-600 text-white py-2 px-4 rounded-lg hover:bg-orange-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {pruneMutation.isLoading ? 'Pruning...' : 'Prune Edges'}
              </button>
            </div>

            {/* Tag Backfill */}
            <div className="bg-white rounded-lg shadow-md p-6">
              <h3 className="text-lg font-semibold mb-4">Tag Backfill</h3>
              <p className="text-gray-600 mb-4">
                Process memories that don't have tags and generate tags for them.
              </p>
              <button
                onClick={handleBackfill}
                disabled={backfillTagsMutation.isLoading}
                className="bg-blue-600 text-white py-2 px-4 rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
              >
                {backfillTagsMutation.isLoading ? 'Processing...' : 'Start Backfill'}
              </button>
            </div>
          </div>
        )}

        {activeTab === 'security' && (
          <div className="space-y-6">
            {/* Security Health */}
            <div className="bg-white rounded-lg shadow-md p-6">
              <h3 className="text-lg font-semibold mb-4">Security Health</h3>
              {securityHealth?.data ? (
                <div className="space-y-3">
                  <div className="flex justify-between text-sm">
                    <span className="text-gray-600">Status:</span>
                    <span className={`font-medium ${securityHealth.data.status === 'ok' ? 'text-green-600' : 'text-red-600'}`}>
                      {securityHealth.data.status}
                    </span>
                  </div>
                  {securityHealth.data.checks && Object.entries(securityHealth.data.checks).map(([check, status]) => (
                    <div key={check} className="flex justify-between text-sm">
                      <span className="text-gray-600">{check}:</span>
                      <span className={`font-medium ${status ? 'text-green-600' : 'text-red-600'}`}>
                        {status ? 'Pass' : 'Fail'}
                      </span>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-gray-600">Security health data not available</p>
              )}
            </div>

            {/* Security Metrics */}
            <div className="bg-white rounded-lg shadow-md p-6">
              <h3 className="text-lg font-semibold mb-4">Security Metrics</h3>
              {securityMetrics?.data ? (
                <div className="space-y-3">
                  {Object.entries(securityMetrics.data).map(([metric, value]) => (
                    <div key={metric} className="flex justify-between text-sm">
                      <span className="text-gray-600">{metric}:</span>
                      <span className="text-gray-900">{value}</span>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-gray-600">Security metrics not available</p>
              )}
            </div>
          </div>
        )}

        {activeTab === 'features' && (
          <div>
            <div className="bg-white rounded-lg shadow-md p-6">
              <h3 className="text-lg font-semibold mb-4">Feature Management</h3>
              {featuresLoading ? (
                <div className="text-center py-8">
                  <div className="loading-spinner mx-auto mb-4"></div>
                  <p className="text-gray-600">Loading features...</p>
                </div>
              ) : adminFeatures?.data?.features ? (
                <div className="space-y-4">
                  {Object.entries(adminFeatures.data.features).map(([feature, enabled]) => 
                    renderFeatureToggle(feature, enabled)
                  )}
                </div>
              ) : (
            <div className="text-center py-8">
                  <FiSettings size={48} className="mx-auto text-gray-400 mb-4" />
                  <p className="text-gray-600">No features available</p>
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </>
  );
};

export default AdminDashboardPage;
