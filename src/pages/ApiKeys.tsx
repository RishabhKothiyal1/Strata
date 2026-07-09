import React, { useState, useEffect } from 'react';
import { useOutletContext } from 'react-router-dom';
import { Dialog } from '../components/Dialog';
import type { LayoutContextType } from '../components/DashboardLayout';

interface ApiKey {
  id: string;
  name: string;
  key: string; // The full key (stored securely in state/localstorage for demo purposes)
  prefix: string; // strata_sk_live_xxxx
  created: string;
  lastUsed: string;
  permissions: string[]; // ['Read', 'Write', 'Admin']
  status: 'Active' | 'Revoked';
}

const DEFAULT_KEYS: ApiKey[] = [
  {
    id: '1',
    name: 'Production Web App Client',
    key: 'strata_sk_live_4a1f9e830c29b846e01a',
    prefix: 'strata_sk_live_4a1f...',
    created: new Date(Date.now() - 3600000 * 24 * 30).toISOString(), // 30 days ago
    lastUsed: new Date(Date.now() - 300000).toISOString(), // 5 mins ago
    permissions: ['Read', 'Write'],
    status: 'Active',
  },
  {
    id: '2',
    name: 'CLI Deployment Tool',
    key: 'strata_sk_live_9b0c2e912c9f801e0a2b',
    prefix: 'strata_sk_live_9b0c...',
    created: new Date(Date.now() - 3600000 * 24 * 15).toISOString(), // 15 days ago
    lastUsed: new Date(Date.now() - 3600000 * 24 * 2).toISOString(), // 2 days ago
    permissions: ['Read', 'Write', 'Admin'],
    status: 'Active',
  },
  {
    id: '3',
    name: 'Analytics Logger (Deprecated)',
    key: 'strata_sk_live_1d2e830aef48a9b2c01d',
    prefix: 'strata_sk_live_1d2e...',
    created: new Date(Date.now() - 3600000 * 24 * 60).toISOString(), // 60 days ago
    lastUsed: new Date(Date.now() - 3600000 * 24 * 12).toISOString(), // 12 days ago
    permissions: ['Read'],
    status: 'Revoked',
  },
];

export const ApiKeys: React.FC = () => {
  const { showToast } = useOutletContext<LayoutContextType>();

  // State management
  const [keys, setKeys] = useState<ApiKey[]>([]);
  const [isModalOpen, setIsModalOpen] = useState(false);
  
  // Form state
  const [keyName, setKeyName] = useState('');
  const [permRead, setPermRead] = useState(true);
  const [permWrite, setPermWrite] = useState(false);
  const [permAdmin, setPermAdmin] = useState(false);
  
  // Single-time display state
  const [generatedKey, setGeneratedKey] = useState<string | null>(null);
  
  // Reveal row state (tracks which key IDs are revealed)
  const [revealedIds, setRevealedIds] = useState<Record<string, boolean>>({});

  // Initialize and load from localstorage
  useEffect(() => {
    const stored = localStorage.getItem('strata.apikeys');
    if (stored) {
      try {
        setKeys(JSON.parse(stored));
      } catch {
        setKeys(DEFAULT_KEYS);
      }
    } else {
      setKeys(DEFAULT_KEYS);
      localStorage.setItem('strata.apikeys', JSON.stringify(DEFAULT_KEYS));
    }
  }, []);

  // Sync back to localstorage
  const saveKeys = (newKeys: ApiKey[]) => {
    setKeys(newKeys);
    localStorage.setItem('strata.apikeys', JSON.stringify(newKeys));
  };

  // Generate random hash
  const generateRandomKeyString = () => {
    const chars = 'abcdef0123456789';
    let result = '';
    for (let i = 0; i < 20; i++) {
      result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    return `strata_sk_live_${result}`;
  };

  // Handle key creation
  const handleGenerateKey = (e: React.FormEvent) => {
    e.preventDefault();
    if (!keyName.trim()) {
      showToast('Key name is required.', 'error');
      return;
    }

    const perms: string[] = [];
    if (permRead) perms.push('Read');
    if (permWrite) perms.push('Write');
    if (permAdmin) perms.push('Admin');

    if (perms.length === 0) {
      showToast('At least one permission must be selected.', 'error');
      return;
    }

    const fullKey = generateRandomKeyString();
    
    const newKey: ApiKey = {
      id: Math.random().toString(36).substring(2, 9),
      name: keyName,
      key: fullKey,
      prefix: `${fullKey.substring(0, 19)}...`,
      created: new Date().toISOString(),
      lastUsed: 'Never Used',
      permissions: perms,
      status: 'Active',
    };

    saveKeys([newKey, ...keys]);
    setGeneratedKey(fullKey); // Displays once
    showToast(`API Key "${keyName}" generated successfully.`, 'success');
  };

  // Close and reset modal
  const handleCloseModal = () => {
    setIsModalOpen(false);
    setKeyName('');
    setPermRead(true);
    setPermWrite(false);
    setPermAdmin(false);
    setGeneratedKey(null);
  };

  // Copy full key to clipboard helper
  const handleCopyToClipboard = (text: string, label: string) => {
    navigator.clipboard.writeText(text);
    showToast(`${label} copied to clipboard!`, 'success');
  };

  // Revoke / Activate key
  const handleToggleRevoke = (id: string, name: string, currentStatus: 'Active' | 'Revoked') => {
    const targetStatus: 'Active' | 'Revoked' = currentStatus === 'Active' ? 'Revoked' : 'Active';
    const updated = keys.map((k) => {
      if (k.id === id) {
        return { ...k, status: targetStatus };
      }
      return k;
    });
    saveKeys(updated);
    showToast(`API Key "${name}" is now ${targetStatus}.`, 'success');
  };

  // Delete key
  const handleDeleteKey = (id: string, name: string) => {
    const updated = keys.filter((k) => k.id !== id);
    saveKeys(updated);
    showToast(`Deleted API Key "${name}".`, 'success');
  };

  // Toggle reveal row state
  const handleToggleReveal = (id: string) => {
    setRevealedIds((prev) => ({
      ...prev,
      [id]: !prev[id],
    }));
  };

  return (
    <div className="space-y-lg animate-fadeIn w-full">
      {/* Page Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-md mb-md">
        <div>
          <h2 className="font-headline-lg text-headline-lg font-bold text-[#191c1d] mb-xs">
            API Credentials
          </h2>
          <p className="font-body-md text-body-md text-[#464554]">
            Manage your API credentials securely. Authenticate client SDK connections to Strata Engine.
          </p>
        </div>

        <button
          onClick={() => setIsModalOpen(true)}
          className="bg-[#4648d4] hover:bg-[#6063ee] text-white py-sm px-lg rounded-lg font-label-md text-label-md flex items-center justify-center space-x-sm shrink-0 shadow-sm cursor-pointer transition-colors focus:outline-none focus:ring-2 focus:ring-[#4648d4]/50"
        >
          <span className="material-symbols-outlined text-md">add_moderator</span>
          <span>Generate API Key</span>
        </button>
      </div>

      {/* API Keys Table */}
      {keys.length === 0 ? (
        /* Centers and aligns empty key state */
        <div className="bg-white rounded-xl shadow-sm border border-[#c7c4d7]/30 p-xl flex flex-col items-center justify-center text-center min-h-[300px]">
          <span className="material-symbols-outlined text-5xl text-[#464554]/30 mb-sm">vpn_key_off</span>
          <h3 className="font-headline-md text-lg text-[#191c1d] font-bold mb-xs">No credentials found</h3>
          <p className="font-body-md text-body-md text-[#464554] max-w-sm mb-lg">
            Create an API key credentials token to connect web clients or backends.
          </p>
          <button
            onClick={() => setIsModalOpen(true)}
            className="bg-[#e2dfff] hover:bg-[#c3c0ff] text-[#0f0069] py-sm px-lg rounded-lg text-sm font-semibold transition-colors cursor-pointer"
          >
            Generate first key
          </button>
        </div>
      ) : (
        <div className="bg-white rounded-xl shadow-sm border border-[#c7c4d7]/30 overflow-hidden w-full">
          <div className="overflow-x-auto w-full">
            <table className="w-full border-collapse text-left">
              <thead>
                <tr className="bg-[#f3f4f5] border-b border-[#c7c4d7]/30 text-[#464554] uppercase font-label-md text-label-md tracking-wider">
                  <th className="px-lg py-md">Key Name</th>
                  <th className="px-lg py-md w-60">Secret Key Prefix</th>
                  <th className="px-lg py-md w-36">Permissions</th>
                  <th className="px-lg py-md w-36 text-center">Status</th>
                  <th className="px-lg py-md w-44">Created At</th>
                  <th className="px-lg py-md w-44">Last Used</th>
                  <th className="px-lg py-md w-44 text-center">Actions</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-[#c7c4d7]/20 font-body-md text-body-md text-[#191c1d]">
                {keys.map((k) => (
                  <tr key={k.id} className="hover:bg-[#f8f9fa] transition-colors">
                    <td className="px-lg py-md font-semibold truncate max-w-xs">{k.name}</td>
                    
                    {/* Key Masking and Copy */}
                    <td className="px-lg py-md font-code-sm text-code-sm text-[#464554] whitespace-nowrap">
                      <div className="flex items-center space-x-sm">
                        <span className="font-mono">
                          {revealedIds[k.id] ? k.key : k.prefix}
                        </span>
                        <button
                          onClick={() => handleToggleReveal(k.id)}
                          className="text-[#464554]/60 hover:text-[#191c1d]"
                          title={revealedIds[k.id] ? 'Hide Key' : 'Reveal Key'}
                        >
                          <span className="material-symbols-outlined text-[16px]">
                            {revealedIds[k.id] ? 'visibility_off' : 'visibility'}
                          </span>
                        </button>
                        <button
                          onClick={() => handleCopyToClipboard(k.key, 'Secret key')}
                          className="text-[#464554]/60 hover:text-[#191c1d]"
                          title="Copy Full Key"
                        >
                          <span className="material-symbols-outlined text-[16px]">content_copy</span>
                        </button>
                      </div>
                    </td>

                    {/* Permissions Badges */}
                    <td className="px-lg py-md">
                      <div className="flex flex-wrap gap-xs">
                        {k.permissions.map((p) => {
                          let style = 'bg-gray-100 text-gray-700';
                          if (p === 'Read') style = 'bg-[#e0f7fa] text-[#006064]';
                          if (p === 'Write') style = 'bg-[#ffeecb] text-[#704200]';
                          if (p === 'Admin') style = 'bg-[#e1e0ff] text-[#07006c]';
                          return (
                            <span key={p} className={`text-[10px] font-bold px-sm py-[1px] rounded ${style}`}>
                              {p}
                            </span>
                          );
                        })}
                      </div>
                    </td>

                    {/* Status Badge */}
                    <td className="px-lg py-md text-center">
                      <span className={`inline-flex items-center gap-xs px-md py-[2px] rounded-full text-xs font-semibold border ${
                        k.status === 'Active'
                          ? 'bg-[#d1f7d1] text-[#0a5c0a] border-[#a5d6a7]/30'
                          : 'bg-[#ffdad6] text-[#93000a] border-[#ba1a1a]/20'
                      }`}>
                        <span className={`w-1.5 h-1.5 rounded-full ${k.status === 'Active' ? 'bg-[#0a5c0a]' : 'bg-[#ba1a1a]'}`} />
                        <span>{k.status}</span>
                      </span>
                    </td>

                    {/* Dates */}
                    <td className="px-lg py-md text-[#464554] text-xs whitespace-nowrap">
                      {new Date(k.created).toLocaleDateString()}
                    </td>
                    <td className="px-lg py-md text-[#464554] text-xs whitespace-nowrap">
                      {k.lastUsed === 'Never Used' ? 'Never Used' : new Date(k.lastUsed).toLocaleString()}
                    </td>

                    {/* Action buttons */}
                    <td className="px-lg py-md text-center">
                      <div className="flex items-center justify-center space-x-sm">
                        <button
                          onClick={() => handleToggleRevoke(k.id, k.name, k.status)}
                          className={`p-xs rounded transition-colors cursor-pointer focus:outline-none focus:ring-2 ${
                            k.status === 'Active'
                              ? 'text-[#704200] hover:bg-[#ffeecb]/40'
                              : 'text-[#0a5c0a] hover:bg-[#d1f7d1]/40'
                          }`}
                          title={k.status === 'Active' ? 'Revoke Key' : 'Activate Key'}
                        >
                          <span className="material-symbols-outlined text-lg">
                            {k.status === 'Active' ? 'block' : 'check_circle'}
                          </span>
                        </button>
                        <button
                          onClick={() => handleDeleteKey(k.id, k.name)}
                          className="text-[#ba1a1a] hover:bg-[#ffdad6]/40 p-xs rounded transition-colors cursor-pointer focus:outline-none focus:ring-2 focus:ring-[#ba1a1a]/30"
                          title="Delete Key"
                        >
                          <span className="material-symbols-outlined text-lg">delete</span>
                        </button>
                      </div>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* ── Generate API Key Dialog Modal ── */}
      <Dialog
        open={isModalOpen}
        onClose={handleCloseModal}
        title="Generate API Key Credentials"
        description={generatedKey ? 'Your key has been created. Copy it now.' : 'Create a credentials secret token to authorize database transactions.'}
        icon="vpn_key"
        size="small"
        footer={
          <button
            onClick={handleCloseModal}
            className="px-md py-sm bg-[#4648d4] hover:bg-[#6063ee] text-white rounded-lg font-label-md text-label-md shadow-sm transition-colors cursor-pointer focus:outline-none focus:ring-2 focus:ring-[#4648d4]/50"
          >
            {generatedKey ? 'Done & Close' : 'Close'}
          </button>
        }
      >
        {generatedKey ? (
          /* Single Time Secret Display (Aesthetics & Security) */
          <div className="space-y-lg animate-fadeIn">
            <div className="p-md bg-[#ffeecb]/40 text-[#704200] border border-[#ffe0b2]/60 rounded-lg flex items-start gap-sm">
              <span className="material-symbols-outlined text-xl mt-0.5 shrink-0">warning</span>
              <p className="font-body-md text-xs leading-normal">
                <strong>Attention:</strong> For security reasons, this key will be displayed <strong>ONLY ONCE</strong>. Copy it and save it in a safe password manager now. You cannot access it again once closed!
              </p>
            </div>

            <div className="space-y-sm">
              <span className="block font-label-md text-xs text-[#191c1d] font-bold">API Secret Key</span>
              <div className="flex bg-[#1e1e1e] border border-[#c7c4d7]/30 rounded-lg p-md justify-between items-center shadow-inner">
                <span className="font-mono text-xs font-semibold text-green-400 select-all break-all pr-sm">
                  {generatedKey}
                </span>
                <button
                  onClick={() => handleCopyToClipboard(generatedKey, 'API Secret key')}
                  className="bg-[#4648d4] hover:bg-[#6063ee] text-white p-xs rounded transition-colors flex items-center justify-center shrink-0 cursor-pointer focus:outline-none focus:ring-2 focus:ring-[#4648d4]/30"
                  title="Copy Key Token"
                >
                  <span className="material-symbols-outlined text-sm">content_copy</span>
                </button>
              </div>
            </div>
          </div>
        ) : (
          /* Key Generation Config Form */
          <form onSubmit={handleGenerateKey} className="space-y-lg">
            
            {/* Key Name */}
            <div className="space-y-sm">
              <label className="block font-label-md text-label-md text-[#191c1d] font-semibold" htmlFor="apikey-name">
                Key Name <span className="text-[#ba1a1a]">*</span>
              </label>
              <input
                className="w-full bg-white border border-[#c7c4d7] rounded-lg px-md py-sm text-[#191c1d] font-body-md text-body-md placeholder-[#464554]/50 focus:outline-none focus:border-[#4648d4] focus:ring-4 focus:ring-[#4648d4]/10 transition-all shadow-sm"
                id="apikey-name"
                type="text"
                placeholder="e.g. Production Mobile App"
                value={keyName}
                onChange={(e) => setKeyName(e.target.value)}
                required
              />
            </div>

            {/* Permissions Checkboxes */}
            <div className="space-y-sm">
              <span className="block font-label-md text-label-md text-[#191c1d] font-semibold">
                Permissions Scope
              </span>
              
              <div className="space-y-sm bg-[#f3f4f5] border border-[#c7c4d7]/20 p-md rounded-lg">
                {/* Read */}
                <label className="flex items-start gap-md cursor-pointer select-none">
                  <input
                    type="checkbox"
                    checked={permRead}
                    onChange={(e) => setPermRead(e.target.checked)}
                    className="w-4 h-4 rounded text-[#4648d4] focus:ring-[#4648d4]/20 mt-[2px] cursor-pointer"
                  />
                  <div>
                    <span className="block text-xs font-semibold text-[#191c1d]">Read Scope</span>
                    <span className="block text-[10px] text-[#464554]">Allows fetching database records and storage structures.</span>
                  </div>
                </label>

                {/* Write */}
                <label className="flex items-start gap-md cursor-pointer select-none border-t border-[#c7c4d7]/20 pt-xs mt-xs">
                  <input
                    type="checkbox"
                    checked={permWrite}
                    onChange={(e) => setPermWrite(e.target.checked)}
                    className="w-4 h-4 rounded text-[#4648d4] focus:ring-[#4648d4]/20 mt-[2px] cursor-pointer"
                  />
                  <div>
                    <span className="block text-xs font-semibold text-[#191c1d]">Write Scope</span>
                    <span className="block text-[10px] text-[#464554]">Allows creating, updating, and deleting database records.</span>
                  </div>
                </label>

                {/* Admin */}
                <label className="flex items-start gap-md cursor-pointer select-none border-t border-[#c7c4d7]/20 pt-xs mt-xs">
                  <input
                    type="checkbox"
                    checked={permAdmin}
                    onChange={(e) => setPermAdmin(e.target.checked)}
                    className="w-4 h-4 rounded text-[#4648d4] focus:ring-[#4648d4]/20 mt-[2px] cursor-pointer"
                  />
                  <div>
                    <span className="block text-xs font-semibold text-[#191c1d]">Admin Scope</span>
                    <span className="block text-[10px] text-[#464554]">Full permissions, including server config adjustments.</span>
                  </div>
                </label>
              </div>
            </div>

            <button
              type="submit"
              className="w-full bg-[#4648d4] hover:bg-[#6063ee] text-white py-sm rounded-lg font-label-md text-label-md font-bold shadow-sm transition-all duration-200 cursor-pointer focus:outline-none focus:ring-2 focus:ring-[#4648d4]/50"
            >
              Generate Credential Key
            </button>
          </form>
        )}
      </Dialog>
    </div>
  );
};
