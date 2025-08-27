import React from 'react';
import { Helmet } from 'react-helmet-async';
import { FiSettings } from 'react-icons/fi';

const SystemManagementPage = () => {
  return (
    <>
      <Helmet>
        <title>System Management - MME</title>
      </Helmet>
      
      <div className="p-6">
        <div className="mb-6">
          <h1 className="text-2xl font-bold text-gray-900">System Management</h1>
          <p className="text-gray-600 mt-1">Memory cleanup, edge pruning, and feature flags</p>
        </div>
        
        <div className="card">
          <div className="card-body">
            <div className="text-center py-8">
              <FiSettings size={48} className="mx-auto text-gray-400 mb-4" />
              <p className="text-gray-600">System Management features coming soon</p>
            </div>
          </div>
        </div>
      </div>
    </>
  );
};

export default SystemManagementPage;
