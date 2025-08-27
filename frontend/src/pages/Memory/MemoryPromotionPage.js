import React from 'react';
import { Helmet } from 'react-helmet-async';
import { FiTrendingUp } from 'react-icons/fi';

const MemoryPromotionPage = () => {
  return (
    <>
      <Helmet>
        <title>Memory Promotion - MME</title>
      </Helmet>
      
      <div className="p-6">
        <div className="mb-6">
          <h1 className="text-2xl font-bold text-gray-900">Memory Promotion</h1>
          <p className="text-gray-600 mt-1">AI-powered memory retrieval with relevance scoring</p>
        </div>
        
        <div className="card">
          <div className="card-body">
            <div className="text-center py-8">
              <FiTrendingUp size={48} className="mx-auto text-gray-400 mb-4" />
              <p className="text-gray-600">Memory Promotion features coming soon</p>
            </div>
          </div>
        </div>
      </div>
    </>
  );
};

export default MemoryPromotionPage;
