import React from 'react';
import { Link, useLocation } from 'react-router-dom';
import { 
  FiHome, 
  FiPackage, 
  FiZap, 
  FiDatabase, 
  FiSearch, 
  FiBarChart2, 
  FiShield, 
  FiSettings, 
  FiActivity,
  FiX,
  FiMenu
} from 'react-icons/fi';
import { useAuth } from '../../hooks/useAuth';

const Sidebar = ({ isOpen, onClose, user }) => {
  const location = useLocation();
  const { hasRole } = useAuth();

  const navigation = [
    {
      name: 'Dashboard',
      href: '/dashboard',
      icon: FiHome,
      current: location.pathname === '/dashboard' || location.pathname === '/',
    },
    {
      name: 'Memory Packing',
      href: '/memory',
      icon: FiPackage,
      current: location.pathname === '/memory',
    },
    {
      name: 'Memory Promotion',
      href: '/promote',
      icon: FiZap,
      current: location.pathname === '/promote',
    },
    {
      name: 'Tag Management',
      href: '/tags',
      icon: FiDatabase,
      current: location.pathname === '/tags',
    },
    {
      name: 'Search & Query',
      href: '/search',
      icon: FiSearch,
      current: location.pathname === '/search',
    },
    {
      name: 'Analytics',
      href: '/analytics',
      icon: FiBarChart2,
      current: location.pathname === '/analytics',
    },
  ];

  const adminNavigation = [
    {
      name: 'Admin Dashboard',
      href: '/admin',
      icon: FiShield,
      current: location.pathname === '/admin',
      requiresAdmin: true,
    },
    {
      name: 'System Management',
      href: '/admin/system',
      icon: FiSettings,
      current: location.pathname === '/admin/system',
      requiresAdmin: true,
    },
    {
      name: 'Monitoring',
      href: '/admin/monitoring',
      icon: FiActivity,
      current: location.pathname === '/admin/monitoring',
      requiresAdmin: true,
    },
  ];

  const settingsNavigation = [
    {
      name: 'Settings',
      href: '/settings',
      icon: FiSettings,
      current: location.pathname === '/settings',
    },
  ];

  const isActive = (href) => {
    if (href === '/dashboard' && location.pathname === '/') return true;
    return location.pathname === href;
  };

  return (
    <>
      {/* Overlay */}
      {isOpen && (
        <div
          className="fixed inset-0 z-30 bg-black bg-opacity-50 lg:hidden"
          onClick={onClose}
        />
      )}

      {/* Sidebar */}
      <div
        className={`fixed left-0 top-0 z-40 h-screen w-64 transform bg-gradient-to-b from-blue-900 to-indigo-900 border-r border-blue-800 transition-transform duration-300 ease-in-out lg:translate-x-0 lg:static lg:inset-0 ${
          isOpen ? 'translate-x-0' : '-translate-x-full'
        }`}
      >
        <div className="flex h-full flex-col">
          {/* Header */}
          <div className="flex items-center justify-between p-6 border-b border-blue-800">
            <div className="flex items-center space-x-3">
              <div className="w-10 h-10 bg-gradient-to-br from-blue-400 to-indigo-500 rounded-xl flex items-center justify-center">
                <FiPackage size={20} className="text-white" />
              </div>
              <div>
                <h1 className="text-xl font-bold text-white">MME Packer</h1>
                <p className="text-blue-200 text-xs">Memory Engine</p>
              </div>
            </div>
            <button
              onClick={onClose}
              className="lg:hidden p-2 text-blue-300 hover:text-white hover:bg-blue-800 rounded-lg"
            >
              <FiX size={20} />
            </button>
          </div>

          {/* Navigation */}
          <nav className="flex-1 px-4 py-6 space-y-2">
            {/* Main Navigation */}
            <div>
              <h3 className="text-xs font-semibold text-blue-300 uppercase tracking-wider mb-3">
                Main
              </h3>
              <div className="space-y-1">
                {navigation.map((item) => (
                  <Link
                    key={item.name}
                    to={item.href}
                    className={`flex items-center px-3 py-2 text-sm font-medium rounded-xl transition-colors ${
                      item.current
                        ? 'bg-blue-700 text-white shadow-lg'
                        : 'text-blue-200 hover:text-white hover:bg-blue-800'
                    }`}
                    onClick={onClose}
                  >
                    <item.icon size={18} className="mr-3" />
                    {item.name}
                  </Link>
                ))}
              </div>
            </div>

            {/* Admin Navigation */}
            {hasRole && hasRole('admin') && (
              <div className="mt-8">
                <h3 className="text-xs font-semibold text-blue-300 uppercase tracking-wider mb-3">
                  Administration
                </h3>
                <div className="space-y-1">
                  {adminNavigation.map((item) => (
                    <Link
                      key={item.name}
                      to={item.href}
                      className={`flex items-center px-3 py-2 text-sm font-medium rounded-xl transition-colors ${
                        item.current
                          ? 'bg-blue-700 text-white shadow-lg'
                          : 'text-blue-200 hover:text-white hover:bg-blue-800'
                      }`}
                      onClick={onClose}
                    >
                      <item.icon size={18} className="mr-3" />
                      {item.name}
                    </Link>
                  ))}
                </div>
              </div>
            )}

            {/* Settings Navigation */}
            <div className="mt-8">
              <h3 className="text-xs font-semibold text-blue-300 uppercase tracking-wider mb-3">
                System
              </h3>
              <div className="space-y-1">
                {settingsNavigation.map((item) => (
                  <Link
                    key={item.name}
                    to={item.href}
                    className={`flex items-center px-3 py-2 text-sm font-medium rounded-xl transition-colors ${
                      item.current
                        ? 'bg-blue-700 text-white shadow-lg'
                        : 'text-blue-200 hover:text-white hover:bg-blue-800'
                    }`}
                    onClick={onClose}
                  >
                    <item.icon size={18} className="mr-3" />
                    {item.name}
                  </Link>
                ))}
              </div>
            </div>
          </nav>

          {/* User Profile */}
          <div className="p-4 border-t border-blue-800">
            <div className="flex items-center space-x-3">
              <div className="w-10 h-10 bg-gradient-to-br from-blue-400 to-indigo-500 rounded-xl flex items-center justify-center">
                <span className="text-white font-semibold text-sm">
                  {user?.name?.charAt(0) || 'U'}
                </span>
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-white truncate">
                  {user?.name || 'Demo User'}
                </p>
                <p className="text-xs text-blue-300 truncate">
                  {user?.role || 'User'}
                </p>
              </div>
            </div>
          </div>
        </div>
      </div>
    </>
  );
};

export default Sidebar;
