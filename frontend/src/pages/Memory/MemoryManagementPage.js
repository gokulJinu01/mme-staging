import React, { useState } from 'react';
import { Helmet } from 'react-helmet-async';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { FiDatabase, FiSearch, FiTrash2, FiZap, FiUpload } from 'react-icons/fi';
import toast from 'react-hot-toast';

import { apiService } from '../../services/apiService';

const MemoryManagementPage = () => {
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState('view');
  const [searchQuery, setSearchQuery] = useState('');
  const [newMemory, setNewMemory] = useState({
    content: '',
    tags: [],
    importance: 0.5,
    metadata: {}
  });

  // Queries
  const { data: memories, isLoading: memoriesLoading } = useQuery(
    ['memories', searchQuery],
    () => apiService.queryMemories({ tags: searchQuery, limit: 50 }),
    {
      enabled: activeTab === 'view',
      refetchInterval: 30000,
    }
  );

  const { data: recentMemories, isLoading: recentLoading } = useQuery(
    ['recent-memories'],
    () => apiService.getRecentMemories({ limit: 10 }),
    {
      enabled: activeTab === 'recent',
      refetchInterval: 30000,
    }
  );

  const { data: tokenizerHealth } = useQuery(
    ['tokenizer-health'],
    () => apiService.getTokenizerHealth(),
    {
      refetchInterval: 60000,
    }
  );

  // Mutations
  const saveMemoryMutation = useMutation({
    mutationFn: apiService.saveMemory,
    onSuccess: () => {
      toast.success('Memory saved successfully');
      queryClient.invalidateQueries(['memories']);
      queryClient.invalidateQueries(['recent-memories']);
      setNewMemory({ content: '', tags: [], importance: 0.5, metadata: {} });
    },
    onError: (error) => {
      toast.error(`Failed to save memory: ${error.message}`);
    },
  });

  const deleteMemoryMutation = useMutation({
    mutationFn: apiService.deleteMemory,
    onSuccess: () => {
      toast.success('Memory deleted successfully');
      queryClient.invalidateQueries(['memories']);
      queryClient.invalidateQueries(['recent-memories']);
    },
    onError: (error) => {
      toast.error(`Failed to delete memory: ${error.message}`);
    },
  });

  const promoteMemoryMutation = useMutation({
    mutationFn: apiService.promoteMemories,
    onSuccess: () => {
      toast.success('Memory promoted successfully');
      queryClient.invalidateQueries(['memories']);
    },
    onError: (error) => {
      toast.error(`Failed to promote memory: ${error.message}`);
    },
  });

  const injectMemoryMutation = useMutation({
    mutationFn: apiService.injectMemory,
    onSuccess: () => {
      toast.success('Memory injected successfully');
      queryClient.invalidateQueries(['memories']);
    },
    onError: (error) => {
      toast.error(`Failed to inject memory: ${error.message}`);
    },
  });

  // Handlers
  const handleSaveMemory = () => {
    if (!newMemory.content.trim()) {
      toast.error('Memory content is required');
      return;
    }
    saveMemoryMutation.mutate(newMemory);
  };

  const handleDeleteMemory = (id) => {
    if (window.confirm('Are you sure you want to delete this memory?')) {
      deleteMemoryMutation.mutate(id);
    }
  };

  const handlePromoteMemory = (memoryId) => {
    promoteMemoryMutation.mutate({ memoryId });
  };

  const handleInjectMemory = (memoryId) => {
    injectMemoryMutation.mutate({ memoryId });
  };

  const handleTagInput = (e) => {
    const tags = e.target.value.split(',').map(tag => tag.trim()).filter(tag => tag);
    setNewMemory(prev => ({ ...prev, tags }));
  };

  const renderMemoryCard = (memory) => (
    <div key={memory.id} className="bg-white rounded-lg shadow-md p-4 border border-gray-200">
      <div className="flex justify-between items-start mb-3">
        <div className="flex-1">
          <h3 className="font-semibold text-gray-900 mb-2">
            {memory.content.substring(0, 100)}{memory.content.length > 100 ? '...' : ''}
          </h3>
          <div className="flex flex-wrap gap-1 mb-2">
            {memory.tags?.map((tag, index) => (
              <span key={index} className="px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded">
                {typeof tag === 'string' ? tag : tag.label}
              </span>
            ))}
          </div>
          <div className="text-sm text-gray-500">
            <span>Importance: {memory.importance || 'N/A'}</span>
            {memory.createdAt && (
              <span className="ml-4">Created: {new Date(memory.createdAt).toLocaleDateString()}</span>
            )}
          </div>
        </div>
        <div className="flex gap-2">
          <button
            onClick={() => handlePromoteMemory(memory.id)}
            className="p-2 text-blue-600 hover:bg-blue-50 rounded"
            title="Promote Memory"
          >
            <FiZap size={16} />
          </button>
          <button
            onClick={() => handleInjectMemory(memory.id)}
            className="p-2 text-green-600 hover:bg-green-50 rounded"
            title="Inject Memory"
          >
            <FiUpload size={16} />
          </button>
          <button
            onClick={() => handleDeleteMemory(memory.id)}
            className="p-2 text-red-600 hover:bg-red-50 rounded"
            title="Delete Memory"
          >
            <FiTrash2 size={16} />
          </button>
        </div>
      </div>
    </div>
  );

  return (
    <>
      <Helmet>
        <title>Memory Management - MME</title>
      </Helmet>
      
      <div className="p-6">
        <div className="mb-6">
          <h1 className="text-2xl font-bold text-gray-900">Memory Management</h1>
          <p className="text-gray-600 mt-1">Store, organize, and manage your AI memories</p>
        </div>

        {/* Tokenizer Health Status */}
        {tokenizerHealth && (
          <div className="mb-6 p-4 bg-blue-50 border border-blue-200 rounded-lg">
            <h3 className="font-semibold text-blue-900 mb-2">Tokenizer Status</h3>
            <div className="text-sm text-blue-800">
              <p>Mode: {tokenizerHealth.data?.mode || 'Unknown'}</p>
              <p>Model: {tokenizerHealth.data?.model || 'Unknown'}</p>
              <p>Status: {tokenizerHealth.data?.status || 'Unknown'}</p>
            </div>
          </div>
        )}

        {/* Tab Navigation */}
        <div className="mb-6">
          <nav className="flex space-x-8">
            <button
              onClick={() => setActiveTab('view')}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'view'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              View Memories
            </button>
            <button
              onClick={() => setActiveTab('add')}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'add'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Add Memory
            </button>
            <button
              onClick={() => setActiveTab('recent')}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'recent'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Recent Memories
            </button>
          </nav>
        </div>

        {/* Tab Content */}
        {activeTab === 'view' && (
          <div>
            {/* Search Bar */}
            <div className="mb-6">
              <div className="relative">
                <FiSearch className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" />
                <input
                  type="text"
                  placeholder="Search memories by tags..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                />
              </div>
            </div>

            {/* Memories List */}
            <div className="space-y-4">
              {memoriesLoading ? (
                <div className="text-center py-8">
                  <div className="loading-spinner mx-auto mb-4"></div>
                  <p className="text-gray-600">Loading memories...</p>
                </div>
              ) : memories?.data?.results?.length > 0 ? (
                memories.data.results.map(renderMemoryCard)
              ) : (
                <div className="text-center py-8">
                  <FiDatabase size={48} className="mx-auto text-gray-400 mb-4" />
                  <p className="text-gray-600">No memories found</p>
                </div>
              )}
            </div>
          </div>
        )}

        {activeTab === 'add' && (
          <div className="max-w-2xl">
            <div className="bg-white rounded-lg shadow-md p-6">
              <h3 className="text-lg font-semibold mb-4">Add New Memory</h3>
              
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Memory Content
                  </label>
                  <textarea
                    value={newMemory.content}
                    onChange={(e) => setNewMemory(prev => ({ ...prev, content: e.target.value }))}
                    placeholder="Enter your memory content..."
                    rows={4}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Tags (comma-separated)
                  </label>
                  <input
                    type="text"
                    value={newMemory.tags.join(', ')}
                    onChange={handleTagInput}
                    placeholder="tag1, tag2, tag3"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Importance (0.0 - 1.0)
                  </label>
                  <input
                    type="range"
                    min="0"
                    max="1"
                    step="0.1"
                    value={newMemory.importance}
                    onChange={(e) => setNewMemory(prev => ({ ...prev, importance: parseFloat(e.target.value) }))}
                    className="w-full"
                  />
                  <div className="text-sm text-gray-500 mt-1">
                    Current value: {newMemory.importance}
                  </div>
                </div>

                <button
                  onClick={handleSaveMemory}
                  disabled={saveMemoryMutation.isLoading}
                  className="w-full bg-blue-600 text-white py-2 px-4 rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {saveMemoryMutation.isLoading ? 'Saving...' : 'Save Memory'}
                </button>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'recent' && (
          <div>
            <h3 className="text-lg font-semibold mb-4">Recent Memories</h3>
            <div className="space-y-4">
              {recentLoading ? (
                <div className="text-center py-8">
                  <div className="loading-spinner mx-auto mb-4"></div>
                  <p className="text-gray-600">Loading recent memories...</p>
                </div>
              ) : recentMemories?.data?.results?.length > 0 ? (
                recentMemories.data.results.map(renderMemoryCard)
              ) : (
                <div className="text-center py-8">
                  <FiDatabase size={48} className="mx-auto text-gray-400 mb-4" />
                  <p className="text-gray-600">No recent memories found</p>
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </>
  );
};

export default MemoryManagementPage;
