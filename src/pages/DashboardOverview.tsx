import React from 'react';
import { Link } from 'react-router-dom';
import { useRecords } from '../hooks/useRecords';
import { useAuth } from '../context/AuthContext';

export const DashboardOverview: React.FC = () => {
  const { user } = useAuth();
  const { data: records = [], isLoading } = useRecords('records');

  // Compute metrics
  const totalRecords = records.length;
  const activeRecords = records.filter(r => r.active).length;
  const draftRecords = totalRecords - activeRecords;

  // Recent 3 records
  const recentRecords = records.slice(0, 3);

  const stats = [
    {
      title: 'Total Records',
      value: isLoading ? '...' : totalRecords,
      subtitle: 'Database records',
      icon: 'database',
      color: 'bg-primary/10 text-primary border-primary/20',
    },
    {
      title: 'Active Collection',
      value: isLoading ? '...' : activeRecords,
      subtitle: `${draftRecords} draft records`,
      icon: 'check_circle',
      color: 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20',
    },
    {
      title: 'Storage Used',
      value: '12.4 MB',
      subtitle: 'of 1.0 GB total storage',
      icon: 'inventory_2',
      color: 'bg-tertiary/10 text-tertiary border-tertiary/20',
    },
    {
      title: 'Edge Runtime',
      value: '100%',
      subtitle: 'Isolated engine running (slog)',
      icon: 'bolt',
      color: 'bg-secondary/10 text-secondary border-secondary/20',
    },
  ];

  return (
    <div className="space-y-xl animate-fadeIn">
      {/* Page Header */}
      <div>
        <h2 className="font-headline-lg text-headline-lg font-bold text-[#191c1d] mb-xs">
          Studio Overview
        </h2>
        <p className="font-body-md text-body-md text-[#464554]">
          Inspect database health, execution contexts, storage, and API settings for {user?.email}.
        </p>
      </div>

      {/* Metrics Row */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-md">
        {stats.map((stat, i) => (
          <div
            key={i}
            className="bg-surface-container-lowest border border-outline-variant/30 p-lg rounded-xl shadow-sm flex items-start justify-between"
          >
            <div className="space-y-xs min-w-0">
              <span className="font-label-md text-label-md text-on-surface-variant block">{stat.title}</span>
              <span className="font-headline-lg text-3xl font-extrabold text-on-surface block">
                {stat.value}
              </span>
              <span className="font-body-sm text-body-sm text-on-surface-variant/75 truncate block">
                {stat.subtitle}
              </span>
            </div>
            <div className={`w-10 h-10 rounded-lg flex items-center justify-center shrink-0 border ${stat.color}`}>
              <span className="material-symbols-outlined text-xl">{stat.icon}</span>
            </div>
          </div>
        ))}
      </div>

      {/* Main Grid Content */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-lg">
        
        {/* Left Side: Recent Database Records */}
        <div className="lg:col-span-2 space-y-md">
          <div className="flex justify-between items-center">
            <h3 className="font-headline-md text-headline-md font-bold text-on-surface">
              Recent Database Records
            </h3>
            <Link
              to="/data"
              className="text-primary hover:text-primary-container hover:underline font-semibold font-label-md text-label-md flex items-center gap-xs"
            >
              <span>View Collection</span>
              <span className="material-symbols-outlined text-sm">arrow_forward</span>
            </Link>
          </div>

          {isLoading ? (
            <div className="bg-surface-container-lowest border border-outline-variant/30 rounded-xl p-xl flex items-center justify-center h-48">
              <div className="w-8 h-8 border-4 border-primary border-t-transparent rounded-full animate-spin" />
            </div>
          ) : recentRecords.length === 0 ? (
            /* Perfectly Centered Empty State */
            <div className="bg-surface-container-lowest border border-outline-variant/30 rounded-xl p-xl flex flex-col items-center justify-center text-center">
              <span className="material-symbols-outlined text-4xl text-on-surface-variant/40 block mb-sm">database_off</span>
              <h4 className="font-headline-md text-base text-on-surface font-bold mb-xs">No records available</h4>

              <Link
                to="/data?new=true"
                className="bg-primary-fixed text-on-primary-fixed hover:bg-primary-fixed-dim font-semibold text-xs py-xs px-md rounded-lg flex items-center gap-xs transition-colors"
              >
                <span className="material-symbols-outlined text-sm">add</span>
                <span>Create Record</span>
              </Link>
            </div>
          ) : (
            <div className="bg-surface-container-lowest border border-outline-variant/30 rounded-xl shadow-sm overflow-hidden">
              <div className="overflow-x-auto">
                <table className="w-full text-left border-collapse">
                  <thead>
                    <tr className="bg-surface-container-low border-b border-outline-variant/30 text-on-surface-variant uppercase font-label-md text-xs tracking-wider">
                      <th className="px-lg py-sm">ID</th>
                      <th className="px-lg py-sm">Title</th>
                      <th className="px-lg py-sm">Status</th>
                      <th className="px-lg py-sm">Created</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-outline-variant/20 font-body-md text-body-md text-on-surface">
                    {recentRecords.map((r) => (
                      <tr key={r.id} className="hover:bg-surface-container-low transition-colors">
                        <td className="px-lg py-sm font-code-sm text-code-sm text-on-surface-variant">#{r.id}</td>
                        <td className="px-lg py-sm font-semibold truncate max-w-xs">{r.title}</td>
                        <td className="px-lg py-sm">
                          <span className={`inline-flex items-center gap-xs px-md py-[1px] rounded-full text-[10px] font-bold border ${
                            r.active 
                              ? 'bg-emerald-500/10 text-emerald-400 border-emerald-500/20' 
                              : 'bg-surface-container-high text-on-surface-variant border-outline-variant/30'
                          }`}>
                            <span className={`w-1 h-1 rounded-full ${r.active ? 'bg-emerald-400' : 'bg-on-surface-variant'}`} />
                            <span>{r.active ? 'Active' : 'Draft'}</span>
                          </span>
                        </td>
                        <td className="px-lg py-sm text-on-surface-variant text-xs">
                          {new Date(r.created_at).toLocaleDateString()}
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>

        {/* Right Side: Quick Links Panel */}
        <div className="space-y-md">
          <h3 className="font-headline-md text-headline-md font-bold text-on-surface">
            Quick Actions
          </h3>

          <div className="bg-surface-container border border-outline-variant/30 rounded-xl shadow-sm p-lg space-y-md">
            <Link
              to="/data"
              className="flex items-start gap-md p-md rounded-lg hover:bg-surface-variant/30 transition-all group"
            >
              <div className="w-9 h-9 rounded-lg bg-primary/10 text-primary flex items-center justify-center shrink-0">
                <span className="material-symbols-outlined">database</span>
              </div>
              <div className="min-w-0">
                <span className="font-label-md text-label-md font-bold text-on-surface group-hover:text-primary flex items-center gap-xs">
                  <span>Browse Collections</span>
                  <span className="material-symbols-outlined text-xs opacity-0 group-hover:opacity-100 transition-opacity">arrow_forward</span>
                </span>
                <p className="font-body-sm text-xs text-on-surface-variant mt-xs">
                  Add, edit, filter, and import CSV tables.
                </p>
              </div>
            </Link>

            <Link
              to="/storage"
              className="flex items-start gap-md p-md rounded-lg hover:bg-surface-variant/30 transition-all group"
            >
              <div className="w-9 h-9 rounded-lg bg-tertiary/10 text-tertiary flex items-center justify-center shrink-0">
                <span className="material-symbols-outlined">inventory_2</span>
              </div>
              <div className="min-w-0">
                <span className="font-label-md text-label-md font-bold text-on-surface group-hover:text-tertiary flex items-center gap-xs">
                  <span>Object Storage</span>
                  <span className="material-symbols-outlined text-xs opacity-0 group-hover:opacity-100 transition-opacity">arrow_forward</span>
                </span>
                <p className="font-body-sm text-xs text-on-surface-variant mt-xs">
                  Manage static assets and media files.
                </p>
              </div>
            </Link>

            <Link
              to="/functions"
              className="flex items-start gap-md p-md rounded-lg hover:bg-surface-variant/30 transition-all group"
            >
              <div className="w-9 h-9 rounded-lg bg-secondary/10 text-secondary flex items-center justify-center shrink-0">
                <span className="material-symbols-outlined">code</span>
              </div>
              <div className="min-w-0">
                <span className="font-label-md text-label-md font-bold text-on-surface group-hover:text-secondary flex items-center gap-xs">
                  <span>Edge Functions</span>
                  <span className="material-symbols-outlined text-xs opacity-0 group-hover:opacity-100 transition-opacity">arrow_forward</span>
                </span>
                <p className="font-body-sm text-xs text-on-surface-variant mt-xs">
                  Deploy serverless Goja code runtime.
                </p>
              </div>
            </Link>

            <Link
              to="/api-keys"
              className="flex items-start gap-md p-md rounded-lg hover:bg-surface-variant/30 transition-all group"
            >
              <div className="w-9 h-9 rounded-lg bg-tertiary/10 text-tertiary-container flex items-center justify-center shrink-0">
                <span className="material-symbols-outlined">vpn_key</span>
              </div>
              <div className="min-w-0">
                <span className="font-label-md text-label-md font-bold text-on-surface group-hover:text-tertiary-container flex items-center gap-xs">
                  <span>API Credentials</span>
                  <span className="material-symbols-outlined text-xs opacity-0 group-hover:opacity-100 transition-opacity">arrow_forward</span>
                </span>
                <p className="font-body-sm text-xs text-on-surface-variant mt-xs">
                  Secure project credentials and token management.
                </p>
              </div>
            </Link>
          </div>
        </div>
      </div>
    </div>
  );
};
