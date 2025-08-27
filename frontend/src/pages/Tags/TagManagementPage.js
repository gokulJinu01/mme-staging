import React, { useState } from 'react';
import { Helmet } from 'react-helmet-async';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { FiTag, FiSearch } from 'react-icons/fi';
import toast from 'react-hot-toast';

import { apiService } from '../../services/apiService';

const TagManagementPage = () => {
  const queryClient = useQueryClient();
  const [activeTab, setActiveTab] = useState('extract');
  const [extractContent, setExtractContent] = useState('');
  const [queryPrompt, setQueryPrompt] = useState('');
  const [deltaData, setDeltaData] = useState({
    tagId: '',
    delta: 0,
    reason: ''
  });

  // Queries - These are not used since we use mutations instead
  // const { data: extractedTags, isLoading: extractLoading } = useQuery(
  //   ['extracted-tags', extractContent],
  //   () => apiService.extractTags({ content: extractContent }),
  //   {
  //     enabled: false, // Manual trigger only
  //   }
  // );

  // const { data: queryResults, isLoading: queryLoading } = useQuery(
  //   ['query-results', queryPrompt],
  //   () => apiService.queryByPrompt({ prompt: queryPrompt }),
  //   {
  //     enabled: false, // Manual trigger only
  //   }
  // );

  // Mutations
  const extractTagsMutation = useMutation({
    mutationFn: apiService.extractTags,
    onSuccess: (data) => {
      toast.success(`Extracted ${data.data?.tags?.length || 0} tags successfully`);
      queryClient.invalidateQueries(['extracted-tags']);
    },
    onError: (error) => {
      toast.error(`Failed to extract tags: ${error.message}`);
    },
  });

  const queryByPromptMutation = useMutation({
    mutationFn: apiService.queryByPrompt,
    onSuccess: (data) => {
      toast.success(`Found ${data.data?.results?.length || 0} results`);
      queryClient.invalidateQueries(['query-results']);
    },
    onError: (error) => {
      toast.error(`Failed to query by prompt: ${error.message}`);
    },
  });

  const updateTagDeltaMutation = useMutation({
    mutationFn: apiService.updateTagDelta,
    onSuccess: () => {
      toast.success('Tag delta updated successfully');
      setDeltaData({ tagId: '', delta: 0, reason: '' });
    },
    onError: (error) => {
      toast.error(`Failed to update tag delta: ${error.message}`);
    },
  });

  // Handlers
  const handleExtractTags = () => {
    if (!extractContent.trim()) {
      toast.error('Content is required for tag extraction');
      return;
    }
    extractTagsMutation.mutate({ content: extractContent });
  };

  const handleQueryByPrompt = () => {
    if (!queryPrompt.trim()) {
      toast.error('Prompt is required for querying');
      return;
    }
    queryByPromptMutation.mutate({ prompt: queryPrompt });
  };

  const handleUpdateDelta = () => {
    if (!deltaData.tagId || !deltaData.reason) {
      toast.error('Tag ID and reason are required');
      return;
    }
    updateTagDeltaMutation.mutate(deltaData);
  };

  const renderTagCard = (tag, index) => (
    <div key={index} className="bg-white rounded-lg shadow-md p-4 border border-gray-200">
      <div className="flex justify-between items-center">
        <div>
          <h3 className="font-semibold text-gray-900">
            {typeof tag === 'string' ? tag : tag.label}
          </h3>
          {tag.type && (
            <p className="text-sm text-gray-500">Type: {tag.type}</p>
          )}
          {tag.confidence && (
            <p className="text-sm text-gray-500">Confidence: {(tag.confidence * 100).toFixed(1)}%</p>
          )}
        </div>
        <span className="px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded">
          Tag
        </span>
      </div>
    </div>
  );

  const renderQueryResult = (result, index) => (
    <div key={index} className="bg-white rounded-lg shadow-md p-4 border border-gray-200">
      <div className="flex justify-between items-start mb-3">
        <div className="flex-1">
          <h3 className="font-semibold text-gray-900 mb-2">
            {result.content?.substring(0, 100)}{result.content?.length > 100 ? '...' : ''}
          </h3>
          <div className="flex flex-wrap gap-1 mb-2">
            {result.tags?.map((tag, tagIndex) => (
              <span key={tagIndex} className="px-2 py-1 bg-green-100 text-green-800 text-xs rounded">
                {typeof tag === 'string' ? tag : tag.label}
              </span>
            ))}
          </div>
          <div className="text-sm text-gray-500">
            <span>Score: {result.score?.toFixed(3) || 'N/A'}</span>
            {result.createdAt && (
              <span className="ml-4">Created: {new Date(result.createdAt).toLocaleDateString()}</span>
            )}
          </div>
        </div>
      </div>
    </div>
  );

  return (
    <>
      <Helmet>
        <title>Tag Management - MME</title>
      </Helmet>
      
      <div className="p-6">
        <div className="mb-6">
          <h1 className="text-2xl font-bold text-gray-900">Tag Management</h1>
          <p className="text-gray-600 mt-1">Extract, query, and manage tags for your memories</p>
        </div>

        {/* Tab Navigation */}
        <div className="mb-6">
          <nav className="flex space-x-8">
            <button
              onClick={() => setActiveTab('extract')}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'extract'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Extract Tags
            </button>
            <button
              onClick={() => setActiveTab('query')}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'query'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Query by Prompt
            </button>
            <button
              onClick={() => setActiveTab('delta')}
              className={`py-2 px-1 border-b-2 font-medium text-sm ${
                activeTab === 'delta'
                  ? 'border-blue-500 text-blue-600'
                  : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
              }`}
            >
              Update Tag Delta
            </button>
          </nav>
        </div>

        {/* Tab Content */}
        {activeTab === 'extract' && (
          <div className="max-w-4xl">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* Input Section */}
              <div className="bg-white rounded-lg shadow-md p-6">
                <h3 className="text-lg font-semibold mb-4">Extract Tags from Content</h3>
                
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Content
                    </label>
                    <textarea
                      value={extractContent}
                      onChange={(e) => setExtractContent(e.target.value)}
                      placeholder="Enter content to extract tags from..."
                      rows={6}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    />
                  </div>

                  <button
                    onClick={handleExtractTags}
                    disabled={extractTagsMutation.isLoading || !extractContent.trim()}
                    className="w-full bg-blue-600 text-white py-2 px-4 rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {extractTagsMutation.isLoading ? 'Extracting...' : 'Extract Tags'}
                  </button>
                </div>
              </div>

              {/* Results Section */}
              <div className="bg-white rounded-lg shadow-md p-6">
                <h3 className="text-lg font-semibold mb-4">Extracted Tags</h3>
                
                <div className="space-y-3">
                  {extractTagsMutation.isLoading ? (
                    <div className="text-center py-8">
                      <div className="loading-spinner mx-auto mb-4"></div>
                      <p className="text-gray-600">Extracting tags...</p>
                    </div>
                  ) : extractTagsMutation.data?.data?.tags?.length > 0 ? (
                    extractTagsMutation.data.data.tags.map(renderTagCard)
                  ) : extractTagsMutation.isSuccess ? (
                    <div className="text-center py-8">
                      <FiTag size={48} className="mx-auto text-gray-400 mb-4" />
                      <p className="text-gray-600">No tags extracted</p>
                    </div>
                  ) : (
                    <div className="text-center py-8">
                      <FiTag size={48} className="mx-auto text-gray-400 mb-4" />
                      <p className="text-gray-600">Enter content and extract tags to see results</p>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'query' && (
          <div className="max-w-4xl">
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              {/* Input Section */}
              <div className="bg-white rounded-lg shadow-md p-6">
                <h3 className="text-lg font-semibold mb-4">Query Memories by Prompt</h3>
                
                <div className="space-y-4">
                  <div>
                    <label className="block text-sm font-medium text-gray-700 mb-2">
                      Prompt
                    </label>
                    <textarea
                      value={queryPrompt}
                      onChange={(e) => setQueryPrompt(e.target.value)}
                      placeholder="Enter a prompt to search for related memories..."
                      rows={4}
                      className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    />
                  </div>

                  <button
                    onClick={handleQueryByPrompt}
                    disabled={queryByPromptMutation.isLoading || !queryPrompt.trim()}
                    className="w-full bg-green-600 text-white py-2 px-4 rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
                  >
                    {queryByPromptMutation.isLoading ? 'Searching...' : 'Search Memories'}
                  </button>
                </div>
              </div>

              {/* Results Section */}
              <div className="bg-white rounded-lg shadow-md p-6">
                <h3 className="text-lg font-semibold mb-4">Query Results</h3>
                
                <div className="space-y-3">
                  {queryByPromptMutation.isLoading ? (
                    <div className="text-center py-8">
                      <div className="loading-spinner mx-auto mb-4"></div>
                      <p className="text-gray-600">Searching memories...</p>
                    </div>
                  ) : queryByPromptMutation.data?.data?.results?.length > 0 ? (
                    queryByPromptMutation.data.data.results.map(renderQueryResult)
                  ) : queryByPromptMutation.isSuccess ? (
                    <div className="text-center py-8">
                      <FiSearch size={48} className="mx-auto text-gray-400 mb-4" />
                      <p className="text-gray-600">No results found</p>
                    </div>
                  ) : (
                    <div className="text-center py-8">
                      <FiSearch size={48} className="mx-auto text-gray-400 mb-4" />
                      <p className="text-gray-600">Enter a prompt to search for memories</p>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        )}

        {activeTab === 'delta' && (
          <div className="max-w-2xl">
            <div className="bg-white rounded-lg shadow-md p-6">
              <h3 className="text-lg font-semibold mb-4">Update Tag Delta</h3>
              
              <div className="space-y-4">
                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Tag ID
                  </label>
                  <input
                    type="text"
                    value={deltaData.tagId}
                    onChange={(e) => setDeltaData(prev => ({ ...prev, tagId: e.target.value }))}
                    placeholder="Enter tag ID"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Delta Value
                  </label>
                  <input
                    type="number"
                    step="0.1"
                    value={deltaData.delta}
                    onChange={(e) => setDeltaData(prev => ({ ...prev, delta: parseFloat(e.target.value) }))}
                    placeholder="Enter delta value"
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>

                <div>
                  <label className="block text-sm font-medium text-gray-700 mb-2">
                    Reason
                  </label>
                  <textarea
                    value={deltaData.reason}
                    onChange={(e) => setDeltaData(prev => ({ ...prev, reason: e.target.value }))}
                    placeholder="Enter reason for delta update..."
                    rows={3}
                    className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  />
                </div>

                <button
                  onClick={handleUpdateDelta}
                  disabled={updateTagDeltaMutation.isLoading || !deltaData.tagId || !deltaData.reason}
                  className="w-full bg-purple-600 text-white py-2 px-4 rounded-lg hover:bg-purple-700 disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {updateTagDeltaMutation.isLoading ? 'Updating...' : 'Update Tag Delta'}
                </button>
              </div>
            </div>
          </div>
        )}
      </div>
    </>
  );
};

export default TagManagementPage;
