import React from 'react';
import { Helmet } from 'react-helmet-async';
import { FiBarChart } from 'react-icons/fi';

const AnalyticsPage = () => {
  return (
    <>
      <Helmet>
        <title>Analytics - MME</title>
      </Helmet>
      
      <div className="p-6">
        <div className="mb-6">
          <h1 className="text-2xl font-bold text-gray-900">Analytics</h1>
          <p className="text-gray-600 mt-1">Memory growth trends and performance metrics</p>
        </div>
        
        <div className="card">
          <div className="card-body">
            <div className="text-center py-8">
              <FiBarChart size={48} className="mx-auto text-gray-400 mb-4" />
              <p className="text-gray-600">Analytics features coming soon</p>
            </div>
          </div>
        </div>
      </div>
    </>
  );
};

export default AnalyticsPage;
