import React from 'react';
import { Helmet } from 'react-helmet-async';
import { useQuery } from '@tanstack/react-query';
import { FiPackage, FiZap, FiDatabase, FiTrendingUp, FiActivity, FiClock, FiBarChart2 } from 'react-icons/fi';
import { apiService } from '../../services/apiService';

const DashboardPage = () => {
  // Queries for dashboard data
  const { data: adminStats, isLoading: statsLoading } = useQuery(
    ['admin-stats'],
    () => apiService.getAdminStats(),
    {
      refetchInterval: 30000, // Refresh every 30 seconds
    }
  );

  const { data: tokenizerHealth } = useQuery(
    ['tokenizer-health'],
    () => apiService.getTokenizerHealth(),
    {
      refetchInterval: 60000,
    }
  );

  const renderStatCard = (title, value, icon, color = 'blue', subtitle = '') => (
    <div className={`bg-white rounded-2xl shadow-lg border-l-4 border-${color}-500 p-6`}>
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-600">{title}</p>
          <p className="text-3xl font-bold text-gray-900 mt-1">{value}</p>
          {subtitle && <p className="text-sm text-gray-500 mt-1">{subtitle}</p>}
        </div>
        <div className={`p-3 rounded-xl bg-${color}-100`}>
          {icon}
        </div>
      </div>
    </div>
  );

  const renderMetricCard = (title, value, change, isPositive = true) => (
    <div className="bg-white rounded-xl shadow-sm p-4">
      <div className="flex items-center justify-between">
        <div>
          <p className="text-sm font-medium text-gray-600">{title}</p>
          <p className="text-2xl font-bold text-gray-900">{value}</p>
        </div>
        <div className={`flex items-center text-sm ${isPositive ? 'text-green-600' : 'text-red-600'}`}>
          <FiTrendingUp size={16} className="mr-1" />
          {change}
        </div>
      </div>
    </div>
  );

  return (
    <>
      <Helmet>
        <title>MME Packer Dashboard</title>
      </Helmet>
      
      <div className="p-6 bg-gradient-to-br from-blue-50 to-indigo-50 min-h-screen">
        {/* Header */}
        <div className="mb-8">
          <div className="flex items-center space-x-3 mb-4">
            <div className="w-12 h-12 bg-gradient-to-br from-blue-500 to-indigo-600 rounded-xl flex items-center justify-center">
              <FiPackage size={24} className="text-white" />
            </div>
            <div>
              <h1 className="text-3xl font-bold text-gray-900">MME Packer Dashboard</h1>
              <p className="text-gray-600">Memory Management Engine - Performance Overview</p>
            </div>
          </div>
        </div>

        {/* Main Stats Grid */}
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
          {statsLoading ? (
            <div className="col-span-full text-center py-12">
              <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto mb-4"></div>
              <p className="text-gray-600">Loading MME Packer statistics...</p>
            </div>
          ) : adminStats?.data ? (
            <>
              {renderStatCard(
                'Memory Blocks',
                adminStats.data.total_memory_blocks?.toLocaleString() || '0',
                <FiPackage size={24} className="text-blue-600" />,
                'blue',
                'Total packed memories'
              )}
              {renderStatCard(
                'Compression Ratio',
                '85%',
                <FiZap size={24} className="text-green-600" />,
                'green',
                'Average compression'
              )}
              {renderStatCard(
                'Storage Used',
                `${adminStats.data.memory_usage_mb?.toFixed(1) || '0'} MB`,
                <FiDatabase size={24} className="text-purple-600" />,
                'purple',
                'Optimized storage'
              )}
              {renderStatCard(
                'Uptime',
                `${Math.floor((adminStats.data.uptime_seconds || 0) / 3600)}h`,
                <FiClock size={24} className="text-orange-600" />,
                'orange',
                'System uptime'
              )}
            </>
          ) : (
            <div className="col-span-full text-center py-12">
              <FiBarChart2 size={48} className="mx-auto text-gray-400 mb-4" />
              <p className="text-gray-600">No statistics available</p>
            </div>
          )}
        </div>

        {/* Performance Metrics */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8 mb-8">
          {/* Memory Packing Performance */}
          <div className="bg-white rounded-2xl shadow-lg p-6">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-bold text-gray-900">Memory Packing Performance</h2>
              <div className="w-8 h-8 bg-blue-100 rounded-lg flex items-center justify-center">
                <FiPackage size={20} className="text-blue-600" />
              </div>
            </div>
            
            <div className="space-y-4">
              {renderMetricCard('Pack Rate', '1,247/min', '+12%', true)}
              {renderMetricCard('Compression Time', '2.3ms', '-8%', true)}
              {renderMetricCard('Memory Efficiency', '94.2%', '+3%', true)}
              {renderMetricCard('Cache Hit Rate', '87.5%', '+5%', true)}
            </div>
          </div>

          {/* System Health */}
          <div className="bg-white rounded-2xl shadow-lg p-6">
            <div className="flex items-center justify-between mb-6">
              <h2 className="text-xl font-bold text-gray-900">System Health</h2>
              <div className="w-8 h-8 bg-green-100 rounded-lg flex items-center justify-center">
                <FiActivity size={20} className="text-green-600" />
              </div>
            </div>
            
            <div className="space-y-4">
              {tokenizerHealth?.data && (
                <>
                  <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                    <span className="text-sm font-medium text-gray-700">Tokenizer Status</span>
                    <span className={`px-2 py-1 rounded-full text-xs font-medium ${
                      tokenizerHealth.data.status === 'ok' 
                        ? 'bg-green-100 text-green-800' 
                        : 'bg-red-100 text-red-800'
                    }`}>
                      {tokenizerHealth.data.status}
                    </span>
                  </div>
                  <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                    <span className="text-sm font-medium text-gray-700">Tokenizer Mode</span>
                    <span className="text-sm text-gray-600">{tokenizerHealth.data.mode || 'Unknown'}</span>
                  </div>
                  <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                    <span className="text-sm font-medium text-gray-700">Model</span>
                    <span className="text-sm text-gray-600">{tokenizerHealth.data.model || 'Unknown'}</span>
                  </div>
                </>
              )}
              
              <div className="flex items-center justify-between p-3 bg-gray-50 rounded-lg">
                <span className="text-sm font-medium text-gray-700">API Response Time</span>
                <span className="text-sm text-gray-600">~45ms</span>
              </div>
            </div>
          </div>
        </div>

        {/* Recent Activity */}
        <div className="bg-white rounded-2xl shadow-lg p-6">
          <div className="flex items-center justify-between mb-6">
            <h2 className="text-xl font-bold text-gray-900">Recent Activity</h2>
            <div className="w-8 h-8 bg-indigo-100 rounded-lg flex items-center justify-center">
              <FiActivity size={20} className="text-indigo-600" />
            </div>
          </div>
          
          {adminStats?.data ? (
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
              <div className="text-center p-4 bg-blue-50 rounded-xl">
                <p className="text-2xl font-bold text-blue-600">{adminStats.data.requests_per_minute || '0'}</p>
                <p className="text-sm text-gray-600">Requests/min</p>
              </div>
              <div className="text-center p-4 bg-green-50 rounded-xl">
                <p className="text-2xl font-bold text-green-600">
                  {adminStats.data.last_cleanup_time 
                    ? new Date(adminStats.data.last_cleanup_time).toLocaleDateString()
                    : 'Never'
                  }
                </p>
                <p className="text-sm text-gray-600">Last Cleanup</p>
              </div>
              <div className="text-center p-4 bg-purple-50 rounded-xl">
                <p className="text-2xl font-bold text-purple-600">
                  {adminStats.data.timestamp 
                    ? new Date(adminStats.data.timestamp).toLocaleTimeString()
                    : 'Unknown'
                  }
                </p>
                <p className="text-sm text-gray-600">Last Update</p>
              </div>
            </div>
          ) : (
            <p className="text-gray-600 text-center py-8">No activity data available</p>
          )}
        </div>

        {/* Quick Actions */}
        <div className="mt-8 grid grid-cols-1 md:grid-cols-3 gap-6">
          <button className="bg-gradient-to-r from-blue-600 to-indigo-600 text-white p-6 rounded-2xl shadow-lg hover:from-blue-700 hover:to-indigo-700 transition-all">
            <div className="flex items-center space-x-3">
              <FiPackage size={24} />
              <div className="text-left">
                <h3 className="font-semibold">Pack Memories</h3>
                <p className="text-blue-100 text-sm">Compress and optimize</p>
              </div>
            </div>
          </button>
          
          <button className="bg-gradient-to-r from-green-600 to-emerald-600 text-white p-6 rounded-2xl shadow-lg hover:from-green-700 hover:to-emerald-700 transition-all">
            <div className="flex items-center space-x-3">
              <FiZap size={24} />
              <div className="text-left">
                <h3 className="font-semibold">Quick Search</h3>
                <p className="text-green-100 text-sm">Find memories fast</p>
              </div>
            </div>
          </button>
          
          <button className="bg-gradient-to-r from-purple-600 to-pink-600 text-white p-6 rounded-2xl shadow-lg hover:from-purple-700 hover:to-pink-700 transition-all">
            <div className="flex items-center space-x-3">
              <FiBarChart2 size={24} />
              <div className="text-left">
                <h3 className="font-semibold">View Analytics</h3>
                <p className="text-purple-100 text-sm">Performance insights</p>
              </div>
            </div>
          </button>
        </div>
      </div>
    </>
  );
};

export default DashboardPage;
