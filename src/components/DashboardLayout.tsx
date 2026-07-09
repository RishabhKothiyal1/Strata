import React, { useState, useEffect } from 'react';
import { NavLink, Outlet, useLocation, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export interface LayoutContextType {
  searchQuery: string;
  setSearchQuery: (query: string) => void;
  showToast: (message: string, type?: 'success' | 'error') => void;
}

export const DashboardLayout: React.FC = () => {
  const { user, signOut } = useAuth();
  const location = useLocation();
  const navigate = useNavigate();

  // State management
  const [searchQuery, setSearchQuery] = useState('');
  const [isProfileOpen, setIsProfileOpen] = useState(false);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);
  
  // Toast state
  const [toast, setToast] = useState<{ message: string; type: 'success' | 'error' } | null>(null);

  const showToast = (message: string, type: 'success' | 'error' = 'success') => {
    setToast({ message, type });
  };

  useEffect(() => {
    if (!toast) return;
    const timer = setTimeout(() => {
      setToast(null);
    }, 4000);
    return () => clearTimeout(timer);
  }, [toast]);

  // Sidebar navigation items
  const menuItems = [
    { to: '/dashboard', label: 'Dashboard', icon: 'dashboard' },
    { to: '/data', label: 'Data', icon: 'database', filled: true },
    { to: '/storage', label: 'Storage', icon: 'inventory_2' },
    { to: '/functions', label: 'Functions', icon: 'code' },
    { to: '/settings', label: 'Settings', icon: 'settings' },
  ];

  // Footer navigation items
  const footerItems = [
    { to: '/support', label: 'Support', icon: 'contact_support' },
    { to: '/api-keys', label: 'API Keys', icon: 'vpn_key' },
  ];

  // Quick Action Handler for sidebar CTA
  const handleNewRecordClick = () => {
    setIsMobileMenuOpen(false);
    navigate('/data?new=true');
  };

  return (
    <div className="bg-background text-on-background font-body-md min-h-screen flex w-full relative">
      
      {/* ── Mobile menu backdrop ── */}
      {isMobileMenuOpen && (
        <div
          className="fixed inset-0 bg-background/50 backdrop-blur-sm z-40 md:hidden animate-fadeIn"
          onClick={() => setIsMobileMenuOpen(false)}
        />
      )}

      {/* ── Sidebar Navigation ── */}
      <nav
        aria-label="Main Navigation"
        className={`
          bg-surface-container-low border-r border-outline-variant/30 fixed left-0 top-0 h-full w-64 z-40 flex flex-col py-xl px-md space-y-sm
          transition-transform duration-200
          ${isMobileMenuOpen ? 'translate-x-0' : '-translate-x-full'}
          md:translate-x-0
        `}
      >
        {/* Header Logo */}
        <div className="flex items-center space-x-md px-md mb-xl">
          <div className="w-10 h-10 rounded-lg bg-primary flex items-center justify-center text-on-primary shrink-0">
            <span className="material-symbols-outlined">layers</span>
          </div>
          <div className="min-w-0">
            <h1 className="font-headline-md text-headline-md font-bold text-primary truncate">Strata Studio</h1>
            <p className="font-label-md text-label-md text-on-surface-variant">Backend Engine</p>
          </div>
        </div>

        {/* Create Record CTA */}
        <button
          onClick={handleNewRecordClick}
          className="w-full bg-primary hover:bg-primary-container active:scale-[0.98] text-on-primary py-sm px-md rounded-lg font-label-md text-label-md shadow-sm transition-all duration-200 mb-lg flex items-center justify-center space-x-sm cursor-pointer focus:outline-none focus:ring-2 focus:ring-primary/50"
          aria-label="Create a new database record"
        >
          <span className="material-symbols-outlined text-md">add</span>
          <span>New Record</span>
        </button>

        {/* Navigation Tabs */}
        <div className="flex-1 space-y-xs overflow-y-auto">
          {menuItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              onClick={() => setIsMobileMenuOpen(false)}
              className={({ isActive }) => `
                w-full group flex items-center p-md rounded-lg border-l-4 transition-all duration-200 text-left
                ${isActive
                  ? 'bg-primary-fixed text-on-primary-fixed border-primary font-semibold'
                  : 'text-on-surface-variant border-transparent hover:bg-surface-variant hover:text-on-surface'
                }
              `}
            >
              <span
                className="material-symbols-outlined mr-md"
                style={item.filled ? { fontVariationSettings: "'FILL' 1" } : undefined}
              >
                {item.icon}
              </span>
              <span className="font-label-md text-label-md">{item.label}</span>
            </NavLink>
          ))}
        </div>

        {/* Sidebar Footer Support/API Keys */}
        <div className="pt-lg border-t border-outline-variant/30 space-y-xs shrink-0">
          {footerItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              onClick={() => setIsMobileMenuOpen(false)}
              className={({ isActive }) => `
                w-full group flex items-center p-md rounded-lg border-l-4 transition-all duration-200 text-left
                ${isActive
                  ? 'bg-primary-fixed text-on-primary-fixed border-primary font-semibold'
                  : 'text-on-surface-variant border-transparent hover:bg-surface-variant hover:text-on-surface'
                }
              `}
            >
              <span className="material-symbols-outlined mr-md">{item.icon}</span>
              <span className="font-label-md text-label-md">{item.label}</span>
            </NavLink>
          ))}
          
          <button
            onClick={signOut}
            className="w-full group flex items-center p-md text-error hover:bg-error-container/20 rounded-lg border-l-4 border-transparent transition-all duration-200 text-left cursor-pointer focus:outline-none focus:ring-2 focus:ring-error/30"
          >
            <span className="material-symbols-outlined mr-md">logout</span>
            <span className="font-label-md text-label-md">Sign Out</span>
          </button>
        </div>
      </nav>

      {/* ── Main Layout Wrapper ── */}
      <div className="flex-1 min-w-0 flex flex-col md:ml-64 min-h-screen">
        
        {/* Top Header Bar */}
        <header className="bg-surface-container-lowest sticky top-0 z-30 shadow-sm flex justify-between items-center h-16 px-md md:px-margin-desktop w-full border-b border-outline-variant/30">
          
          {/* Mobile Hamburguer and App Title */}
          <div className="md:hidden flex items-center space-x-sm">
            <button
              onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
              className="p-xs rounded-lg hover:bg-surface-container-low transition-colors focus:outline-none focus:ring-2 focus:ring-primary/30"
              aria-label="Toggle Navigation Drawer"
            >
              <span className="material-symbols-outlined text-primary">menu</span>
            </button>
            <span className="font-headline-md text-headline-md font-bold text-primary">Strata Studio</span>
          </div>

          {/* Search bar - Only visible on the /data route */}
          <div className={`
            md:flex items-center bg-surface-container-low rounded-full px-md py-sm w-96 border border-outline-variant/60 
            focus-within:border-primary focus-within:ring-2 focus-within:ring-primary/10 transition-all
            ${location.pathname === '/data' ? 'flex' : 'hidden'}
          `}>
            <span className="material-symbols-outlined text-on-surface-variant mr-sm">search</span>
            <input
              className="bg-transparent border-none outline-none w-full font-body-md text-body-md text-on-surface placeholder:text-on-surface-variant/70"
              placeholder="Search database records..."
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              aria-label="Search database records"
            />
          </div>
          
          {/* Placeholder search spacer to maintain alignment on other routes */}
          {location.pathname !== '/data' && <div className="hidden md:block w-96" />}

          {/* Topbar Actions */}
          <div className="flex items-center space-x-sm md:space-x-md">
            <button className="text-on-surface-variant hover:text-on-surface hover:bg-surface-container-low transition-colors duration-200 p-sm rounded-full focus:outline-none focus:ring-2 focus:ring-primary/30" title="Notifications">
              <span className="material-symbols-outlined">notifications</span>
            </button>
            <button className="text-on-surface-variant hover:text-on-surface hover:bg-surface-container-low transition-colors duration-200 p-sm rounded-full focus:outline-none focus:ring-2 focus:ring-primary/30" title="Help">
              <span className="material-symbols-outlined">help</span>
            </button>
            <a 
              target="_blank" 
              rel="noreferrer" 
              className="hidden md:flex font-label-md text-label-md text-primary hover:bg-primary-fixed px-md py-sm rounded-full transition-colors focus:outline-none focus:ring-2 focus:ring-primary/50" 
              href="https://strata.dev/docs"
            >
              Docs
            </a>

            {/* Profile Menu Dropdown */}
            <div className="relative">
              <button
                onClick={() => setIsProfileOpen(!isProfileOpen)}
                className="w-8 h-8 rounded-full overflow-hidden border border-outline-variant ml-xs cursor-pointer hover:opacity-80 transition-opacity focus:outline-none focus:ring-2 focus:ring-primary/50"
                aria-label="User profile menu"
                aria-expanded={isProfileOpen}
              >
                <div className="w-full h-full bg-primary-fixed text-on-primary-fixed font-semibold flex items-center justify-center text-xs">
                  {user?.email?.substring(0, 2).toUpperCase() || 'DE'}
                </div>
              </button>

              {isProfileOpen && (
                <>
                  <div className="fixed inset-0 z-40" onClick={() => setIsProfileOpen(false)} />
                  <div className="absolute right-0 mt-xs w-48 bg-surface-container-lowest rounded-lg border border-outline-variant/30 shadow-md py-sm z-50 animate-scaleIn origin-top-right">
                    <div className="px-md py-xs border-b border-outline-variant/20 mb-xs">
                      <p className="font-label-md text-label-md font-bold truncate text-on-surface">{user?.email}</p>
                      <p className="font-code-sm text-code-sm text-on-surface-variant/75 text-[10px] uppercase">
                        {user?.role || 'Developer'}
                      </p>
                    </div>
                    <button
                      onClick={() => { setIsProfileOpen(false); signOut(); }}
                      className="w-full text-left px-md py-sm font-body-md text-body-md text-error hover:bg-error-container/20 transition-colors flex items-center gap-xs cursor-pointer focus:outline-none"
                    >
                      <span className="material-symbols-outlined text-[18px]">logout</span>
                      <span>Sign Out</span>
                    </button>
                  </div>
                </>
              )}
            </div>
          </div>
        </header>

        {/* Global Toast Alert */}
        {toast && (
          <div className="fixed top-20 right-8 z-[100] p-sm bg-surface-container-lowest text-on-surface border-l-4 border-[#4caf50] rounded-r-lg flex items-start gap-sm text-sm animate-fadeIn shadow-lg max-w-sm">
            <span className="material-symbols-outlined text-[#4caf50] text-[20px] mt-0.5 shrink-0">check_circle</span>
            <span className="flex-1 font-body-md font-semibold text-on-surface">{toast.message}</span>
            <button onClick={() => setToast(null)} className="text-on-surface-variant/60 hover:text-on-surface cursor-pointer" type="button" aria-label="Dismiss toast notification">
              <span className="material-symbols-outlined text-[18px]">close</span>
            </button>
          </div>
        )}

        {/* ── Sub Page Content Outlet ── */}
        <main className="flex-1 p-margin-mobile md:p-margin-desktop overflow-y-auto">
          <Outlet context={{ searchQuery, setSearchQuery, showToast } as LayoutContextType} />
        </main>
      </div>
    </div>
  );
};
