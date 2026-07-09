import React, { useState } from 'react';
import { useOutletContext } from 'react-router-dom';
import type { LayoutContextType } from '../components/DashboardLayout';

export const Settings: React.FC = () => {
  const { showToast } = useOutletContext<LayoutContextType>();

  // Settings state variables
  const [projectName, setProjectName] = useState('Strata Studio Production');
  const [projectDesc, setProjectDesc] = useState('Production database instance for the application.');
  const [backupSchedule, setBackupSchedule] = useState('daily');
  const [selectedRegion, setSelectedRegion] = useState('us-east-1');
  const [isSaving, setIsSaving] = useState(false);

  const handleSaveSettings = (e: React.FormEvent) => {
    e.preventDefault();
    setIsSaving(true);
    setTimeout(() => {
      setIsSaving(false);
      showToast('Settings saved successfully.', 'success');
    }, 800);
  };

  return (
    <div className="space-y-lg animate-fadeIn max-w-4xl">
      {/* Page Header */}
      <div>
        <h2 className="font-headline-lg text-headline-lg font-bold text-[#191c1d] mb-xs">
          Studio Settings
        </h2>
        <p className="font-body-md text-body-md text-[#464554]">
          Manage configuration properties, backups schedules, security tokens, and API Gateways.
        </p>
      </div>

      <form onSubmit={handleSaveSettings} className="space-y-lg">
        
        {/* Section 1: Project Metadata */}
        <div className="bg-white rounded-xl shadow-sm p-lg border border-[#c7c4d7]/30 space-y-md">
          <h3 className="font-headline-md text-headline-md font-bold text-[#191c1d] pb-sm border-b border-[#c7c4d7]/20">
            Project Metadata
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-md">
            <div className="space-y-sm">
              <label className="block font-label-md text-label-md text-[#191c1d] font-semibold" htmlFor="proj-name">
                Project Name
              </label>
              <input
                className="w-full bg-white border border-[#c7c4d7] rounded-lg px-md py-sm text-[#191c1d] font-body-md text-body-md placeholder-[#464554]/50 focus:outline-none focus:border-[#4648d4] focus:ring-4 focus:ring-[#4648d4]/10 transition-all shadow-sm"
                id="proj-name"
                type="text"
                value={projectName}
                onChange={(e) => setProjectName(e.target.value)}
              />
            </div>
            
            <div className="space-y-sm">
              <label className="block font-label-md text-label-md text-[#191c1d] font-semibold" htmlFor="proj-region">
                Data Region
              </label>
              <select
                id="proj-region"
                value={selectedRegion}
                onChange={(e) => setSelectedRegion(e.target.value)}
                className="w-full bg-white border border-[#c7c4d7] rounded-lg px-md py-sm text-[#191c1d] font-body-md text-body-md focus:outline-none focus:border-[#4648d4] focus:ring-4 focus:ring-[#4648d4]/10 transition-all shadow-sm"
              >
                <option value="us-east-1">us-east-1 (N. Virginia)</option>
                <option value="eu-west-1">eu-west-1 (Ireland)</option>
                <option value="ap-southeast-1">ap-southeast-1 (Singapore)</option>
                <option value="us-west-2">us-west-2 (Oregon)</option>
              </select>
            </div>
          </div>

          <div className="space-y-sm">
            <label className="block font-label-md text-label-md text-[#191c1d] font-semibold" htmlFor="proj-desc">
              Description
            </label>
            <textarea
              className="w-full bg-white border border-[#c7c4d7] rounded-lg px-md py-sm text-[#191c1d] font-body-md text-body-md placeholder-[#464554]/50 focus:outline-none focus:border-[#4648d4] focus:ring-4 focus:ring-[#4648d4]/10 transition-all shadow-sm resize-y"
              id="proj-desc"
              rows={3}
              value={projectDesc}
              onChange={(e) => setProjectDesc(e.target.value)}
            />
          </div>
        </div>

        {/* Section 2: Gateway Configuration */}
        <div className="bg-white rounded-xl shadow-sm p-lg border border-[#c7c4d7]/30 space-y-md">
          <h3 className="font-headline-md text-headline-md font-bold text-[#191c1d] pb-sm border-b border-[#c7c4d7]/20">
            Gateway Connection
          </h3>
          <div className="space-y-md">
            <div>
              <label className="block font-label-md text-label-md text-[#464554] mb-xs">Strata Server URL</label>
              <input
                disabled
                className="w-full bg-[#f3f4f5] border border-[#c7c4d7] rounded-lg px-md py-sm font-code-sm text-code-sm text-[#464554]"
                value={import.meta.env.VITE_STRATA_URL || 'http://localhost:8000'}
              />
              <p className="font-code-sm text-[10px] text-[#464554]/70 mt-xs">Connection url is read-only from environment variables.</p>
            </div>
            
            <div>
              <label className="block font-label-md text-label-md text-[#464554] mb-xs">CORS Origin State</label>
              <div className="p-sm bg-[#e1e0ff]/20 text-[#07006c] rounded-lg border border-[#c0c1ff]/30 text-sm">
                Origin <strong>{window.location.origin}</strong> is allowed by API Gateway CORS filters.
              </div>
            </div>
          </div>
        </div>

        {/* Section 3: Backups Scheduling */}
        <div className="bg-white rounded-xl shadow-sm p-lg border border-[#c7c4d7]/30 space-y-md">
          <h3 className="font-headline-md text-headline-md font-bold text-[#191c1d] pb-sm border-b border-[#c7c4d7]/20">
            Automated Backups
          </h3>
          
          <div className="space-y-sm">
            <span className="block font-label-md text-label-md text-[#191c1d] font-semibold">Backup Frequency</span>
            <span className="block text-xs text-[#464554] mb-sm">Choose how often your database snapshot backup is performed.</span>
            
            <div className="grid grid-cols-3 gap-md">
              {['daily', 'weekly', 'off'].map((freq) => (
                <button
                  key={freq}
                  type="button"
                  onClick={() => setBackupSchedule(freq)}
                  className={`
                    py-md px-lg rounded-lg border text-center font-label-md text-label-md font-bold transition-all duration-200 cursor-pointer focus:outline-none
                    ${backupSchedule === freq
                      ? 'bg-[#e1e0ff] text-[#07006c] border-[#4648d4] shadow-sm'
                      : 'bg-white border-[#c7c4d7] text-[#464554] hover:bg-[#f3f4f5]'
                    }
                  `}
                >
                  <span className="uppercase">{freq}</span>
                </button>
              ))}
            </div>
          </div>
        </div>

        {/* Action Form Footer */}
        <div className="flex justify-end pt-md">
          <button
            type="submit"
            disabled={isSaving}
            className="bg-[#4648d4] hover:bg-[#6063ee] disabled:opacity-50 disabled:cursor-not-allowed text-white py-sm px-xl rounded-lg font-label-md text-label-md font-bold shadow-sm transition-all duration-200 flex items-center gap-xs cursor-pointer focus:outline-none focus:ring-2 focus:ring-[#4648d4]/50"
          >
            {isSaving ? (
              <span className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
            ) : (
              <span className="material-symbols-outlined text-[18px]">save</span>
            )}
            <span>{isSaving ? 'Saving Settings...' : 'Save Settings'}</span>
          </button>
        </div>
      </form>
    </div>
  );
};
