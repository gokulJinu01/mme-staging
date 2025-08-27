import React from 'react';
import { Helmet } from 'react-helmet-async';
import { FiActivity } from 'react-icons/fi';

const MonitoringPage = () => {
  return (
    <>
      <Helmet>
        <title>Monitoring - MME</title>
      </Helmet>
      
      <div className="p-6">
        <div className="mb-6">
          <h1 className="text-2xl font-bold text-gray-900">Monitoring</h1>
          <p className="text-gray-600 mt-1">Real-time metrics, log viewing, and performance monitoring</p>
        </div>
        
        <div className="card">
          <div className="card-body">
            <div className="text-center py-8">
              <FiActivity size={48} className="mx-auto text-gray-400 mb-4" />
              <p className="text-gray-600">Monitoring features coming soon</p>
            </div>
          </div>
        </div>
      </div>
    </>
  );
};

export default MonitoringPage;
