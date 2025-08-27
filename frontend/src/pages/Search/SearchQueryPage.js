import React from 'react';
import { Helmet } from 'react-helmet-async';
import { FiSearch } from 'react-icons/fi';

const SearchQueryPage = () => {
  return (
    <>
      <Helmet>
        <title>Search & Query - MME</title>
      </Helmet>
      
      <div className="p-6">
        <div className="mb-6">
          <h1 className="text-2xl font-bold text-gray-900">Search & Query</h1>
          <p className="text-gray-600 mt-1">Advanced search across your memory graph</p>
        </div>
        
        <div className="card">
          <div className="card-body">
            <div className="text-center py-8">
              <FiSearch size={48} className="mx-auto text-gray-400 mb-4" />
              <p className="text-gray-600">Search & Query features coming soon</p>
            </div>
          </div>
        </div>
      </div>
    </>
  );
};

export default SearchQueryPage;
