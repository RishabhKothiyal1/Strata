import React, { useState } from 'react';
import { useAuth } from '../context/AuthContext';
import {
  useRecords,
  useCreateRecord,
  useUpdateRecord,
  useDeleteRecord,
} from '../hooks/useRecords';
import type { DatabaseRecord } from '../hooks/useRecords';
import { useRealtimeTable } from '../hooks/useRealtimeTable';
import { strata } from '../lib/strata';

type TabType = 'dashboard' | 'data' | 'storage' | 'functions' | 'settings';

export const Dashboard: React.FC = () => {
  const { user, signOut } = useAuth();
  const tableName = 'records';

  // State
  const [activeTab, setActiveTab] = useState<TabType>('dashboard');
  const [searchQuery, setSearchQuery] = useState('');
  const [isPanelOpen, setIsPanelOpen] = useState(false);
  const [editingRecord, setEditingRecord] = useState<DatabaseRecord | null>(null);
  const [isProfileOpen, setIsProfileOpen] = useState(false);

  // Form State
  const [formTitle, setFormTitle] = useState('');
  const [formBody, setFormBody] = useState('');
  const [formActive, setFormActive] = useState(true);
  const [isSaving, setIsSaving] = useState(false);

  // CSV Import state
  const [isCsvImportOpen, setIsCsvImportOpen] = useState(false);
  const [csvRows, setCsvRows] = useState<string[][]>([]);
  const [csvHeaders, setCsvHeaders] = useState<string[]>([]);
  const [isImporting, setIsImporting] = useState(false);

  // AI Schema state
  const [isAiSchemaOpen, setIsAiSchemaOpen] = useState(false);
  const [aiPrompt, setAiPrompt] = useState('');
  const [aiResponse, setAiResponse] = useState('');
  const [isGenerating, setIsGenerating] = useState(false);

  // Fetching data for DB records
  const { data: records = [], isLoading, isError, error } = useRecords(tableName);

  // Realtime subscription
  useRealtimeTable(tableName, ['records', tableName, user?.id]);

  // Mutations
  const createRecord = useCreateRecord(tableName);
  const updateRecord = useUpdateRecord(tableName);
  const deleteRecord = useDeleteRecord(tableName);

  // Filter records based on search
  const filteredRecords = records.filter((r) => {
    const titleMatch = r.title?.toLowerCase().includes(searchQuery.toLowerCase());
    const bodyMatch = r.body?.toLowerCase().includes(searchQuery.toLowerCase());
    const idMatch = String(r.id).includes(searchQuery);
    return titleMatch || bodyMatch || idMatch;
  });

  const openNewRecordPanel = () => {
    setEditingRecord(null);
    setFormTitle('');
    setFormBody('');
    setFormActive(true);
    setIsPanelOpen(true);
  };

  const openEditRecordPanel = (record: DatabaseRecord) => {
    setEditingRecord(record);
    setFormTitle(record.title);
    setFormBody(record.body || '');
    setFormActive(record.active);
    setIsPanelOpen(true);
  };

  const closePanel = () => {
    setIsPanelOpen(false);
    setEditingRecord(null);
  };

  const handleSave = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formTitle.trim()) {
      alert('Title is required');
      return;
    }

    setIsSaving(true);
    try {
      if (editingRecord) {
        await updateRecord.mutateAsync({
          id: editingRecord.id,
          title: formTitle,
          body: formBody,
          active: formActive,
        });
      } else {
        await createRecord.mutateAsync({
          title: formTitle,
          body: formBody,
          active: formActive,
        });
      }
      closePanel();
    } catch (err) {
      console.error(err);
      alert('Failed to save record. Check server logs.');
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (window.confirm('Are you sure you want to delete this record?')) {
      try {
        await deleteRecord.mutateAsync(id);
      } catch (err) {
        console.error(err);
        alert('Failed to delete record.');
      }
    }
  };

  const handleToggleActive = async (record: DatabaseRecord) => {
    try {
      await updateRecord.mutateAsync({
        id: record.id,
        active: !record.active,
      });
    } catch (err) {
      console.error(err);
    }
  };

  // --- CSV Import ---
  const parseCSV = (text: string): string[][] => {
    const lines = text.split('\n').filter((l) => l.trim());
    return lines.map((line) => {
      const result: string[] = [];
      let current = '';
      let inQuotes = false;
      for (let i = 0; i < line.length; i++) {
        const char = line[i];
        if (char === '"') {
          inQuotes = !inQuotes;
        } else if (char === ',' && !inQuotes) {
          result.push(current.trim());
          current = '';
        } else {
          current += char;
        }
      }
      result.push(current.trim());
      return result;
    });
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (event) => {
      const text = event.target?.result as string;
      const parsed = parseCSV(text);
      if (parsed.length > 0) {
        setCsvHeaders(parsed[0]);
        setCsvRows(parsed.slice(1));
      }
    };
    reader.readAsText(file);
  };

  const handleCsvImport = async () => {
    if (csvRows.length === 0) return;
    setIsImporting(true);
    try {
      for (const row of csvRows) {
        const title = row[0] || 'Imported Record';
        const body = row.length > 1 ? row.slice(1).join(', ') : '';
        await createRecord.mutateAsync({ title, body, active: true });
      }
      setIsCsvImportOpen(false);
      setCsvRows([]);
      setCsvHeaders([]);
    } catch (err) {
      console.error(err);
      alert('CSV import failed. Check server logs.');
    } finally {
      setIsImporting(false);
    }
  };

  // --- AI Schema Generation ---
  const handleAiGenerate = async () => {
    if (!aiPrompt.trim()) return;
    setIsGenerating(true);
    setAiResponse('');
    try {
      const result = await strata.ai.chat.chat({
        model: 'gpt-4o',
        messages: [
          {
            role: 'system',
            content:
              'You are a database schema assistant. Generate a JSON object representing a database record with title, body, and active fields. Respond ONLY with valid JSON, no markdown.',
          },
          { role: 'user', content: aiPrompt },
        ],
      });
      const content = result.choices?.[0]?.message?.content || '';
      setAiResponse(content);
    } catch (err) {
      console.error(err);
      setAiResponse('');
      alert('AI generation failed. Ensure an AI provider is configured.');
    } finally {
      setIsGenerating(false);
    }
  };

  const handleAiApply = async () => {
    try {
      let parsed: any;
      try {
        parsed = JSON.parse(aiResponse.replace(/```json|```/g, '').trim());
      } catch {
        alert('Invalid JSON from AI response.');
        return;
      }
      await createRecord.mutateAsync({
        title: parsed.title || 'AI Generated',
        body: parsed.body || aiResponse,
        active: parsed.active !== undefined ? parsed.active : true,
      });
      setIsAiSchemaOpen(false);
      setAiPrompt('');
      setAiResponse('');
    } catch (err) {
      console.error(err);
      alert('Failed to create record from AI schema.');
    }
  };

  return (
    <div className="bg-[#f8f9fa] text-[#191c1d] font-body-md min-h-screen overflow-x-hidden flex w-full">
      {/* SideNavBar (Desktop Only) */}
      <nav className="hidden md:flex bg-[#f3f4f5] border-r border-[#c7c4d7]/30 fixed left-0 top-0 h-full w-64 z-40 flex-col py-xl px-md space-y-sm">
        {/* Header */}
        <div className="flex items-center space-x-md px-md mb-xl">
          <div className="w-10 h-10 rounded-lg bg-[#4648d4] flex items-center justify-center text-white">
            <span className="material-symbols-outlined font-headline-md">layers</span>
          </div>
          <div>
            <h1 className="font-headline-md text-headline-md font-bold text-[#4648d4]">Strata Studio</h1>
            <p className="font-label-md text-label-md text-[#464554]">Backend Engine</p>
          </div>
        </div>

        {/* CTA */}
        <button
          onClick={openNewRecordPanel}
          className="w-full bg-[#4648d4] hover:bg-[#6063ee] text-white py-sm px-md rounded-lg font-label-md text-label-md shadow-sm transition-all duration-200 mb-lg flex items-center justify-center space-x-sm"
        >
          <span className="material-symbols-outlined text-md">add</span>
          <span>New Record</span>
        </button>

        {/* Main Navigation Tabs */}
        <div className="flex-1 space-y-xs">
          <button
            onClick={() => setActiveTab('dashboard')}
            className={`w-full group flex items-center p-md cursor-pointer rounded-lg border-l-4 transition-all duration-200 ${
              activeTab === 'dashboard'
                ? 'bg-[#e1e0ff] text-[#07006c] border-[#4648d4] font-semibold'
                : 'text-[#464554] border-transparent hover:bg-[#e1e3e4]'
            }`}
          >
            <span className="material-symbols-outlined mr-md">dashboard</span>
            <span className="font-label-md text-label-md">Dashboard</span>
          </button>
          
          <button
            onClick={() => setActiveTab('data')}
            className={`w-full group flex items-center p-md cursor-pointer rounded-lg border-l-4 transition-all duration-200 ${
              activeTab === 'data'
                ? 'bg-[#e1e0ff] text-[#07006c] border-[#4648d4] font-semibold'
                : 'text-[#464554] border-transparent hover:bg-[#e1e3e4]'
            }`}
          >
            <span className="material-symbols-outlined mr-md" style={{ fontVariationSettings: "'FILL' 1" }}>
              database
            </span>
            <span className="font-label-md text-label-md">Data</span>
          </button>

          <button
            onClick={() => setActiveTab('storage')}
            className={`w-full group flex items-center p-md cursor-pointer rounded-lg border-l-4 transition-all duration-200 ${
              activeTab === 'storage'
                ? 'bg-[#e1e0ff] text-[#07006c] border-[#4648d4] font-semibold'
                : 'text-[#464554] border-transparent hover:bg-[#e1e3e4]'
            }`}
          >
            <span className="material-symbols-outlined mr-md">inventory_2</span>
            <span className="font-label-md text-label-md">Storage</span>
          </button>

          <button
            onClick={() => setActiveTab('functions')}
            className={`w-full group flex items-center p-md cursor-pointer rounded-lg border-l-4 transition-all duration-200 ${
              activeTab === 'functions'
                ? 'bg-[#e1e0ff] text-[#07006c] border-[#4648d4] font-semibold'
                : 'text-[#464554] border-transparent hover:bg-[#e1e3e4]'
            }`}
          >
            <span className="material-symbols-outlined mr-md">code</span>
            <span className="font-label-md text-label-md">Functions</span>
          </button>

          <button
            onClick={() => setActiveTab('settings')}
            className={`w-full group flex items-center p-md cursor-pointer rounded-lg border-l-4 transition-all duration-200 ${
              activeTab === 'settings'
                ? 'bg-[#e1e0ff] text-[#07006c] border-[#4648d4] font-semibold'
                : 'text-[#464554] border-transparent hover:bg-[#e1e3e4]'
            }`}
          >
            <span className="material-symbols-outlined mr-md">settings</span>
            <span className="font-label-md text-label-md">Settings</span>
          </button>
        </div>

        {/* Footer Navigation Tabs */}
        <div className="pt-lg border-t border-[#c7c4d7]/30 space-y-xs">
          <a
            className="group flex items-center p-md cursor-pointer text-[#464554] hover:bg-[#e1e3e4] rounded-lg transition-all duration-200"
            href="#"
          >
            <span className="material-symbols-outlined mr-md">contact_support</span>
            <span className="font-label-md text-label-md">Support</span>
          </a>
          <a
            className="group flex items-center p-md cursor-pointer text-[#464554] hover:bg-[#e1e3e4] rounded-lg transition-all duration-200"
            href="#"
          >
            <span className="material-symbols-outlined mr-md">vpn_key</span>
            <span className="font-label-md text-label-md">API Keys</span>
          </a>
          <button
            onClick={signOut}
            className="w-full group flex items-center p-md cursor-pointer text-[#ba1a1a] hover:bg-[#ffdad6]/40 rounded-lg transition-all duration-200 text-left"
          >
            <span className="material-symbols-outlined mr-md">logout</span>
            <span className="font-label-md text-label-md">Sign Out</span>
          </button>
        </div>
      </nav>

      {/* Main Content Area Wrapper */}
      <div className="flex-1 flex flex-col md:ml-64 w-full min-h-screen">
        {/* TopNavBar (Mobile & Desktop) */}
        <header className="bg-white sticky top-0 z-30 shadow-sm flex justify-between items-center h-16 px-margin-desktop w-full border-b border-[#c7c4d7]/30">
          {/* Mobile Brand */}
          <div className="md:hidden flex items-center space-x-md">
            <span className="material-symbols-outlined text-[#4648d4] font-headline-md">menu</span>
            <span className="font-headline-md text-headline-md font-bold text-[#4648d4]">Strata Studio</span>
          </div>

          {/* Search Bar */}
          <div className="hidden md:flex items-center bg-[#f3f4f5] rounded-full px-md py-sm w-96 border border-[#c7c4d7]/60 focus-within:border-[#4648d4] focus-within:ring-2 focus-within:ring-[#4648d4]/10 transition-all">
            <span className="material-symbols-outlined text-[#464554] mr-sm">search</span>
            <input
              className="bg-transparent border-none outline-none w-full font-body-md text-body-md text-[#191c1d] placeholder:text-[#464554]/70"
              placeholder="Search database records..."
              type="text"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
            />
          </div>

          {/* Actions Right */}
          <div className="flex items-center space-x-md">
            <button className="text-[#464554] hover:text-[#191c1d] hover:bg-[#f3f4f5] transition-colors duration-200 p-sm rounded-full flex items-center justify-center">
              <span className="material-symbols-outlined">notifications</span>
            </button>
            <button className="text-[#464554] hover:text-[#191c1d] hover:bg-[#f3f4f5] transition-colors duration-200 p-sm rounded-full flex items-center justify-center">
              <span className="material-symbols-outlined">help</span>
            </button>
            <a
              target="_blank"
              rel="noreferrer"
              className="hidden md:flex font-label-md text-label-md text-[#4648d4] hover:bg-[#e1e0ff] px-md py-sm rounded-full transition-colors"
              href="https://strata.dev/docs"
            >
              Docs
            </a>

            {/* User Profile Dropdown Menu */}
            <div className="relative">
              <div
                onClick={() => setIsProfileOpen(!isProfileOpen)}
                className="w-8 h-8 rounded-full overflow-hidden border border-[#c7c4d7] ml-sm cursor-pointer hover:opacity-80 transition-opacity"
              >
                <div className="w-full h-full bg-[#e1e0ff] text-[#07006c] font-semibold flex items-center justify-center text-xs">
                  {user?.email?.substring(0, 2).toUpperCase() || 'DE'}
                </div>
              </div>
              
              {isProfileOpen && (
                <>
                  <div className="fixed inset-0 z-10" onClick={() => setIsProfileOpen(false)}></div>
                  <div className="absolute right-0 mt-xs w-48 bg-white rounded-lg border border-[#c7c4d7]/30 shadow-md py-sm z-20">
                    <div className="px-md py-xs border-b border-[#c7c4d7]/20 mb-xs">
                      <p className="font-label-md text-label-md font-bold truncate text-[#191c1d]">{user?.email}</p>
                      <p className="text-code-sm font-code-sm text-[#464554]/75 text-[10px] uppercase">
                        {user?.role || 'Developer'}
                      </p>
                    </div>
                    <button
                      onClick={() => {
                        setIsProfileOpen(false);
                        signOut();
                      }}
                      className="w-full text-left px-md py-sm font-body-md text-body-md text-[#ba1a1a] hover:bg-[#ffdad6]/20 transition-colors flex items-center gap-xs"
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

        {/* Page Content Canvas */}
        <main className="flex-1 p-margin-mobile md:p-margin-desktop overflow-y-auto">
          {/* TAB 1 & 2: DASHBOARD or DATA RECORDS */}
          {(activeTab === 'dashboard' || activeTab === 'data') && (
            <>
              {/* Page Header */}
              <div className="flex flex-col md:flex-row md:items-center justify-between mb-xl">
                <div>
                  <h2 className="font-headline-lg-mobile md:font-headline-lg text-headline-lg-mobile md:text-headline-lg font-bold text-[#191c1d] mb-xs">
                    {activeTab === 'dashboard' ? 'Studio Dashboard' : 'Data Records Collection'}
                  </h2>
                  <p className="font-body-md text-body-md text-[#464554]">
                    Manage database entries, edit variables, and inspect raw JSON outputs.
                  </p>
                </div>
                <div className="mt-md md:mt-0">
                  <button
                    onClick={openNewRecordPanel}
                    className="bg-[#4648d4] hover:bg-[#6063ee] text-white py-sm px-lg rounded-lg font-label-md text-label-md shadow-sm transition-colors flex items-center space-x-sm"
                  >
                    <span className="material-symbols-outlined text-md">add</span>
                    <span>New Record</span>
                  </button>
                </div>
              </div>

              {/* Database Content States */}
              {isLoading ? (
                <div className="bg-white rounded-xl shadow-sm p-xl flex flex-col items-center justify-center min-h-[400px] border border-[#e1e3e4]/50">
                  <div className="w-12 h-12 border-4 border-[#4648d4] border-t-transparent rounded-full animate-spin mb-md"></div>
                  <p className="font-body-lg text-body-lg text-[#464554]">Fetching collections records...</p>
                </div>
              ) : isError ? (
                <div className="bg-white rounded-xl shadow-sm p-xl flex flex-col items-center justify-center min-h-[400px] border border-[#ba1a1a]/30 text-center">
                  <span className="material-symbols-outlined text-[#ba1a1a] text-5xl mb-md">error_outline</span>
                  <h3 className="font-headline-md text-headline-md text-[#191c1d] mb-sm">Query Execution Failed</h3>
                  <p className="font-body-lg text-body-lg text-[#464554] max-w-md mb-xl">
                    {error instanceof Error ? error.message : 'An error occurred while calling the Strata REST API.'}
                  </p>
                  <button
                    onClick={() => window.location.reload()}
                    className="bg-[#4648d4] text-white px-lg py-sm rounded-lg font-label-md text-label-md transition-colors"
                  >
                    Retry Request
                  </button>
                </div>
              ) : records.length === 0 ? (
                /* Empty State Card */
                <div className="bg-white rounded-xl shadow-sm p-xl flex flex-col items-center justify-center min-h-[500px] border border-[#e1e3e4]/40">
                  <div className="w-64 h-64 mb-lg flex items-center justify-center">
                    <svg
                      className="w-full h-full text-[#c0c1ff]/50"
                      viewBox="0 0 200 200"
                      fill="none"
                      xmlns="http://www.w3.org/2000/svg"
                    >
                      <circle cx="100" cy="100" r="80" fill="currentColor" fillOpacity="0.1" />
                      <rect x="60" y="70" width="80" height="60" rx="8" stroke="currentColor" strokeWidth="3" />
                      <line x1="75" y1="90" x2="125" y2="90" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
                      <line x1="75" y1="105" x2="110" y2="105" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
                      <circle cx="130" cy="130" r="25" fill="#e1e0ff" stroke="#4648d4" strokeWidth="3" />
                      <path d="M124 130H136M130 124V136" stroke="#4648d4" strokeWidth="3" strokeLinecap="round" />
                    </svg>
                  </div>
                  <h3 className="font-headline-md text-headline-md text-[#191c1d] mb-sm text-center">No records found</h3>
                  <p className="font-body-lg text-body-lg text-[#464554] text-center max-w-sm mb-xl leading-relaxed">
                    Your database is currently empty. Create a new record or import data from CSV to get started.
                  </p>
                  <button
                    onClick={openNewRecordPanel}
                    className="bg-[#e2dfff] text-[#0f0069] py-md px-xl rounded-lg font-label-md text-label-md shadow-sm hover:opacity-90 transition-opacity flex items-center space-x-sm"
                  >
                    <span className="material-symbols-outlined">add_circle</span>
                    <span>Create your first record</span>
                  </button>
                  <div className="mt-xl pt-lg border-t border-[#c7c4d7]/40 w-full max-w-md flex flex-col items-center">
                    <p className="font-label-md text-label-md text-[#464554] mb-md uppercase tracking-wider">
                      Or start with
                    </p>
                    <div className="flex space-x-md">
                      <button
                        onClick={() => setIsCsvImportOpen(true)}
                        className="flex items-center space-x-xs text-[#4b41e1] hover:text-[#4648d4] transition-colors font-body-md text-body-md"
                      >
                        <span className="material-symbols-outlined text-sm">upload_file</span>
                        <span>Import CSV</span>
                      </button>
                      <span className="text-[#c7c4d7]">•</span>
                      <button
                        onClick={() => setIsAiSchemaOpen(true)}
                        className="flex items-center space-x-xs text-[#4b41e1] hover:text-[#4648d4] transition-colors font-body-md text-body-md"
                      >
                        <span className="material-symbols-outlined text-sm">auto_awesome</span>
                        <span>Generate AI Schema</span>
                      </button>
                    </div>
                  </div>
                </div>
              ) : (
                /* Table Data State */
                <div className="bg-white rounded-xl shadow-sm border border-[#c7c4d7]/30 overflow-hidden">
                  {searchQuery && filteredRecords.length === 0 ? (
                    <div className="p-xl text-center">
                      <span className="material-symbols-outlined text-[#464554]/40 text-4xl mb-sm">search_off</span>
                      <p className="font-body-lg text-body-lg text-[#464554]">
                        No records match "{searchQuery}"
                      </p>
                    </div>
                  ) : (
                    <div className="overflow-x-auto w-full">
                      <table className="w-full border-collapse text-left">
                        <thead>
                          <tr className="bg-[#f3f4f5] border-b border-[#c7c4d7]/30 text-[#464554] uppercase font-label-md text-label-md tracking-wider">
                            <th className="px-lg py-md w-20">ID</th>
                            <th className="px-lg py-md w-1/4">Title</th>
                            <th className="px-lg py-md">Description / Details</th>
                            <th className="px-lg py-md w-32 text-center">Status</th>
                            <th className="px-lg py-md w-44">Created At</th>
                            <th className="px-lg py-md w-28 text-center">Actions</th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-[#c7c4d7]/20 font-body-md text-body-md text-[#191c1d]">
                          {filteredRecords.map((record) => (
                            <tr key={record.id} className="hover:bg-[#f8f9fa] transition-colors">
                              <td className="px-lg py-md font-code-sm text-code-sm text-[#464554]">
                                #{record.id}
                              </td>
                              <td className="px-lg py-md font-semibold truncate max-w-xs">{record.title}</td>
                              <td className="px-lg py-md max-w-md truncate">
                                {record.body ? (
                                  <span className="font-code-sm text-code-sm bg-[#f3f4f5] px-sm py-[2px] rounded border border-[#c7c4d7]/10 text-[#464554] block truncate">
                                    {record.body}
                                  </span>
                                ) : (
                                  <span className="text-[#464554]/40 italic">No details</span>
                                )}
                              </td>
                              <td className="px-lg py-md text-center">
                                <button
                                  onClick={() => handleToggleActive(record)}
                                  title={record.active ? 'Click to deactivate' : 'Click to activate'}
                                  className={`inline-flex items-center gap-xs px-md py-[2px] rounded-full text-xs font-semibold select-none border transition-colors ${
                                    record.active
                                      ? 'bg-[#e1e0ff] text-[#07006c] border-[#c0c1ff]'
                                      : 'bg-[#d9dadb]/40 text-[#464554]/60 border-[#767586]/20'
                                  }`}
                                >
                                  <span
                                    className={`w-1.5 h-1.5 rounded-full ${
                                      record.active ? 'bg-[#4648d4]' : 'bg-[#464554]/60'
                                    }`}
                                  ></span>
                                  <span>{record.active ? 'Active' : 'Draft'}</span>
                                </button>
                              </td>
                              <td className="px-lg py-md text-[#464554] text-sm whitespace-nowrap">
                                {new Date(record.created_at).toLocaleString([], {
                                  year: 'numeric',
                                  month: '2-digit',
                                  day: '2-digit',
                                  hour: '2-digit',
                                  minute: '2-digit',
                                })}
                              </td>
                              <td className="px-lg py-md text-center">
                                <div className="flex items-center justify-center space-x-sm">
                                  <button
                                    onClick={() => openEditRecordPanel(record)}
                                    className="text-[#4648d4] hover:bg-[#e1e0ff]/40 p-xs rounded transition-colors"
                                    title="Edit Record"
                                  >
                                    <span className="material-symbols-outlined text-lg">edit</span>
                                  </button>
                                  <button
                                    onClick={() => handleDelete(record.id)}
                                    className="text-[#ba1a1a] hover:bg-[#ffdad6]/40 p-xs rounded transition-colors"
                                    title="Delete Record"
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
                  )}
                </div>
              )}
            </>
          )}

          {/* TAB 3: STORAGE */}
          {activeTab === 'storage' && (
            <div className="space-y-lg">
              <div className="flex flex-col md:flex-row md:items-center justify-between mb-xl">
                <div>
                  <h2 className="font-headline-lg text-headline-lg font-bold text-[#191c1d]">Object Storage</h2>
                  <p className="font-body-md text-body-md text-[#464554]">
                    Store, manage, and process assets. Includes automatic Lanczos3 resizing operations.
                  </p>
                </div>
                <button
                  onClick={() => alert('Storage upload integration is ready. Add files to see them.')}
                  className="bg-[#4648d4] text-white py-sm px-lg rounded-lg font-label-md text-label-md flex items-center space-x-sm"
                >
                  <span className="material-symbols-outlined">upload_file</span>
                  <span>Upload File</span>
                </button>
              </div>

              <div className="bg-white rounded-xl shadow-sm p-lg border border-[#c7c4d7]/30 flex flex-col items-center justify-center min-h-[350px]">
                <span className="material-symbols-outlined text-6xl text-[#4648d4]/30 mb-md">folder_open</span>
                <h3 className="font-headline-md text-headline-md text-[#191c1d] mb-xs font-semibold">No assets found</h3>
                <p className="font-body-md text-body-md text-[#464554] text-center max-w-sm mb-lg">
                  Drag and drop files here or click upload to store media assets in your MinIO S3 bucket.
                </p>
                <div className="w-full max-w-md border-2 border-dashed border-[#c7c4d7] rounded-lg p-lg text-center cursor-pointer hover:bg-[#f8f9fa] transition-colors">
                  <span className="material-symbols-outlined text-3xl text-[#464554]/50 mb-xs">cloud_upload</span>
                  <p className="font-label-md text-label-md text-[#464554]">Select file or drop it here</p>
                </div>
              </div>
            </div>
          )}

          {/* TAB 4: FUNCTIONS */}
          {activeTab === 'functions' && (
            <div className="space-y-lg">
              <div className="flex flex-col md:flex-row md:items-center justify-between mb-xl">
                <div>
                  <h2 className="font-headline-lg text-headline-lg font-bold text-[#191c1d]">Serverless Edge Functions</h2>
                  <p className="font-body-md text-body-md text-[#464554]">
                    Isolated JavaScript execution runtime powered by pure Goja JS interpreter.
                  </p>
                </div>
                <button
                  onClick={() => alert('Edge Function Deployer is ready.')}
                  className="bg-[#4648d4] text-white py-sm px-lg rounded-lg font-label-md text-label-md flex items-center space-x-sm"
                >
                  <span className="material-symbols-outlined">bolt</span>
                  <span>New Function</span>
                </button>
              </div>

              <div className="grid grid-cols-1 lg:grid-cols-3 gap-lg">
                <div className="lg:col-span-2 bg-white rounded-xl shadow-sm p-lg border border-[#c7c4d7]/30">
                  <h3 className="font-headline-md text-headline-md text-[#191c1d] font-bold mb-md">JavaScript Code Editor</h3>
                  <div className="bg-[#1e1e1e] rounded-lg p-md font-code-sm text-code-sm text-white overflow-x-auto min-h-[250px]">
                    <pre className="text-green-400">
{`// Example ES5.1 Serverless Edge Function
function handleRequest(req) {
  var payload = JSON.parse(req.body);
  console.log("Processing request for: " + payload.name);
  
  return {
    status: 200,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({
      message: "Hello " + payload.name + " from Strata Edge!",
      timestamp: new Date().toISOString()
    })
  };
}`}
                    </pre>
                  </div>
                </div>

                <div className="bg-white rounded-xl shadow-sm p-lg border border-[#c7c4d7]/30 space-y-md">
                  <h3 className="font-headline-md text-headline-md text-[#191c1d] font-bold">Runtime Properties</h3>
                  <div className="space-y-sm text-sm">
                    <div className="flex justify-between border-b pb-xs">
                      <span className="text-[#464554]">JS Engine</span>
                      <span className="font-semibold font-code-sm">Goja (ES5.1)</span>
                    </div>
                    <div className="flex justify-between border-b pb-xs">
                      <span className="text-[#464554]">Execution Timeout</span>
                      <span className="font-semibold font-code-sm">10.0 seconds</span>
                    </div>
                    <div className="flex justify-between border-b pb-xs">
                      <span className="text-[#464554]">Logs Integration</span>
                      <span className="font-semibold font-code-sm">Go slog binding</span>
                    </div>
                  </div>
                  <button
                    onClick={() => alert('Execution triggered.')}
                    className="w-full bg-[#e2dfff] text-[#0f0069] py-sm px-md rounded-lg font-label-md text-label-md"
                  >
                    Run Test Execution
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* TAB 5: SETTINGS */}
          {activeTab === 'settings' && (
            <div className="space-y-lg">
              <div className="mb-xl">
                <h2 className="font-headline-lg text-headline-lg font-bold text-[#191c1d]">Studio Settings</h2>
                <p className="font-body-md text-body-md text-[#464554]">
                  Manage configuration properties, security tokens, and API Gateways.
                </p>
              </div>

              <div className="bg-white rounded-xl shadow-sm p-lg border border-[#c7c4d7]/30 max-w-2xl space-y-md">
                <h3 className="font-headline-md text-headline-md text-[#191c1d] font-bold">Gateway Connection</h3>
                <div className="space-y-md">
                  <div>
                    <label className="block font-label-md text-label-md text-[#464554] mb-xs">Strata Server URL</label>
                    <input
                      disabled
                      className="w-full bg-[#f3f4f5] border border-[#c7c4d7] rounded-lg px-md py-sm font-code-sm text-code-sm text-[#464554]"
                      value={import.meta.env.VITE_STRATA_URL || 'http://localhost:8000'}
                    />
                  </div>
                  <div>
                    <label className="block font-label-md text-label-md text-[#464554] mb-xs">CORS Origin State</label>
                    <div className="p-sm bg-[#e1e0ff]/20 text-[#07006c] rounded-lg border border-[#c0c1ff]/30 text-sm">
                      Origin <strong>{window.location.origin}</strong> is allowed by API Gateway CORS filters.
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}
        </main>
      </div>

      {/* Slide-out Panel Overlay & Modal */}
      {isPanelOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-end">
          {/* Backdrop blur overlay */}
          <div
            onClick={closePanel}
            className="fixed inset-0 bg-[#191c1d]/30 backdrop-blur-sm transition-opacity"
          ></div>

          {/* Slide-over Panel Container */}
          <div className="h-full w-full max-w-md bg-white shadow-2xl relative z-10 flex flex-col border-l border-[#c7c4d7]/30 transform transition-transform duration-300">
            {/* Panel Header */}
            <div className="flex items-center justify-between px-lg py-md border-b border-[#c7c4d7]/30 bg-white">
              <div className="flex items-center gap-sm">
                <span className="material-symbols-outlined text-[#4648d4] text-[24px]">
                  {editingRecord ? 'edit_square' : 'add_box'}
                </span>
                <h3 className="font-headline-md text-headline-md font-bold text-[#191c1d]">
                  {editingRecord ? 'Edit Record' : 'Create New Record'}
                </h3>
              </div>
              <button
                onClick={closePanel}
                className="text-[#464554] hover:text-[#191c1d] hover:bg-[#f3f4f5] rounded-full p-xs transition-colors"
              >
                <span className="material-symbols-outlined">close</span>
              </button>
            </div>

            {/* Panel Body / Form (Scrollable area) */}
            <form onSubmit={handleSave} className="flex-1 overflow-y-auto p-lg space-y-lg pb-24">
              {/* Status Banner */}
              <div className="bg-[#e1e0ff]/30 text-[#07006c] border border-[#c0c1ff]/60 p-sm rounded-lg flex items-start gap-sm">
                <span className="material-symbols-outlined text-[20px] mt-0.5">info</span>
                <p className="font-body-md text-body-md text-sm">
                  {editingRecord ? (
                    <>
                      You are updating record <strong>#{editingRecord.id}</strong> in the <strong>{tableName}</strong> collection.
                    </>
                  ) : (
                    <>
                      You are adding a new record to the <strong>{tableName}</strong> collection. Ensure all fields are valid.
                    </>
                  )}
                </p>
              </div>

              {/* Title Input */}
              <div className="space-y-sm">
                <label className="block font-label-md text-label-md text-[#191c1d] font-semibold" htmlFor="record-title">
                  Title <span className="text-[#ba1a1a]">*</span>
                </label>
                <input
                  className="w-full bg-white border border-[#c7c4d7] rounded-lg px-md py-sm text-[#191c1d] font-body-md text-body-md placeholder-[#464554]/50 focus:outline-none focus:border-[#4648d4] focus:ring-4 focus:ring-[#4648d4]/10 transition-all shadow-sm"
                  id="record-title"
                  name="record-title"
                  placeholder="e.g., Primary Server Node"
                  required
                  type="text"
                  value={formTitle}
                  onChange={(e) => setFormTitle(e.target.value)}
                />
                <p className="font-code-sm text-code-sm text-[#464554]/70">A unique label for this record.</p>
              </div>

              {/* Description/Body Textarea */}
              <div className="space-y-sm">
                <label className="block font-label-md text-label-md text-[#191c1d] font-semibold" htmlFor="record-body">
                  Description / Body Data
                </label>
                <div className="relative group">
                  <textarea
                    className="w-full bg-white border border-[#c7c4d7] rounded-lg px-md py-sm text-[#191c1d] font-code-sm text-code-sm placeholder-[#464554]/50 focus:outline-none focus:border-[#4648d4] focus:ring-4 focus:ring-[#4648d4]/10 transition-all shadow-sm resize-y"
                    id="record-body"
                    name="record-body"
                    placeholder="Enter JSON or plain text details..."
                    rows={6}
                    value={formBody}
                    onChange={(e) => setFormBody(e.target.value)}
                  ></textarea>
                  <div className="absolute top-2 right-2 text-[#464554]/30 pointer-events-none flex gap-1">
                    <span className="material-symbols-outlined text-[16px]">data_object</span>
                  </div>
                </div>
              </div>

              {/* Toggle/Switch */}
              <div className="flex items-center justify-between py-sm border-t border-[#c7c4d7]/40">
                <div>
                  <span className="block font-label-md text-label-md text-[#191c1d] font-semibold">Active Status</span>
                  <span className="block font-code-sm text-code-sm text-[#464554]/70">
                    Publish immediately to subscribers
                  </span>
                </div>
                <label className="relative inline-flex items-center cursor-pointer">
                  <input
                    type="checkbox"
                    checked={formActive}
                    onChange={(e) => setFormActive(e.target.checked)}
                    className="sr-only peer"
                  />
                  <div className="w-11 h-6 bg-[#e1e3e4] peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-[#4648d4]/20 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-[#c7c4d7] after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-[#4648d4]"></div>
                </label>
              </div>

              {/* Panel Footer / Actions (Sticky inside the slide-over panel container but outside form scroll) */}
              <div className="absolute bottom-0 left-0 w-full px-lg py-md border-t border-[#c7c4d7]/30 bg-[#f3f4f5] flex justify-end gap-md z-20">
                <button
                  onClick={closePanel}
                  className="px-md py-sm rounded-lg border border-[#c7c4d7] bg-white text-[#191c1d] font-label-md text-label-md hover:bg-[#f3f4f5] transition-colors focus:outline-none focus:ring-2 focus:ring-[#4648d4]/50"
                  type="button"
                >
                  Cancel
                </button>
                <button
                  className="px-md py-sm rounded-lg bg-[#4648d4] text-white font-label-md text-label-md shadow-sm hover:bg-[#6063ee] transition-colors focus:outline-none focus:ring-4 focus:ring-[#4648d4]/20 flex items-center gap-xs disabled:opacity-50 disabled:cursor-not-allowed"
                  type="submit"
                  disabled={isSaving}
                >
                  {isSaving ? (
                    <span className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  ) : (
                    <span className="material-symbols-outlined text-[18px]">save</span>
                  )}
                  <span>{isSaving ? 'Saving...' : editingRecord ? 'Update Record' : 'Save Record'}</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* CSV Import Modal */}
      {isCsvImportOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div
            onClick={() => setIsCsvImportOpen(false)}
            className="fixed inset-0 bg-[#191c1d]/30 backdrop-blur-sm"
          ></div>
          <div className="relative bg-white rounded-xl shadow-2xl w-full max-w-lg mx-md z-10 flex flex-col max-h-[80vh]">
            <div className="flex items-center justify-between px-lg py-md border-b border-[#c7c4d7]/30">
              <div className="flex items-center gap-sm">
                <span className="material-symbols-outlined text-[#4648d4]">upload_file</span>
                <h3 className="font-headline-md text-headline-md font-bold text-[#191c1d]">Import CSV</h3>
              </div>
              <button
                onClick={() => setIsCsvImportOpen(false)}
                className="text-[#464554] hover:text-[#191c1d] hover:bg-[#f3f4f5] rounded-full p-xs transition-colors"
              >
                <span className="material-symbols-outlined">close</span>
              </button>
            </div>
            <div className="flex-1 overflow-y-auto p-lg space-y-lg">
              {csvRows.length === 0 ? (
                <div className="border-2 border-dashed border-[#c7c4d7] rounded-lg p-xl text-center">
                  <span className="material-symbols-outlined text-4xl text-[#464554]/50 mb-md">cloud_upload</span>
                  <p className="font-body-md text-body-md text-[#464554] mb-md">Select a CSV file to import</p>
                  <input
                    type="file"
                    accept=".csv"
                    onChange={handleFileChange}
                    className="block w-full text-sm text-[#464554] file:mr-4 file:py-sm file:px-md file:rounded-lg file:border-0 file:text-sm file:bg-[#4648d4] file:text-white hover:file:bg-[#6063ee] cursor-pointer"
                  />
                </div>
              ) : (
                <div className="space-y-md">
                  <div className="bg-[#e1e0ff]/20 text-[#07006c] p-sm rounded-lg text-sm">
                    Found {csvRows.length} row{csvRows.length !== 1 ? 's' : ''} with {csvHeaders.length} column{csvHeaders.length !== 1 ? 's' : ''}.
                  </div>
                  <div className="overflow-x-auto border border-[#c7c4d7]/30 rounded-lg max-h-60">
                    <table className="w-full text-left text-sm">
                      <thead>
                        <tr className="bg-[#f3f4f5] text-[#464554]">
                          {csvHeaders.map((h, i) => (
                            <th key={i} className="px-md py-xs font-semibold whitespace-nowrap">{h}</th>
                          ))}
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-[#c7c4d7]/20">
                        {csvRows.slice(0, 10).map((row, ri) => (
                          <tr key={ri} className="hover:bg-[#f8f9fa]">
                            {row.map((cell, ci) => (
                              <td key={ci} className="px-md py-xs truncate max-w-[150px]">{cell}</td>
                            ))}
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                  {csvRows.length > 10 && (
                    <p className="text-sm text-[#464554]/70">Showing first 10 of {csvRows.length} rows.</p>
                  )}
                  <div className="bg-[#f3f4f5] p-sm rounded-lg text-sm text-[#464554]">
                    The first column will be used as the record <strong>title</strong>. Remaining columns will be combined as the record <strong>body</strong>.
                  </div>
                </div>
              )}
            </div>
            <div className="px-lg py-md border-t border-[#c7c4d7]/30 bg-[#f3f4f5] flex justify-end gap-md rounded-b-xl">
              <button
                onClick={() => {
                  setIsCsvImportOpen(false);
                  setCsvRows([]);
                  setCsvHeaders([]);
                }}
                className="px-md py-sm rounded-lg border border-[#c7c4d7] bg-white text-[#191c1d] font-label-md text-label-md hover:bg-[#f3f4f5] transition-colors"
              >
                Cancel
              </button>
              <button
                onClick={handleCsvImport}
                disabled={csvRows.length === 0 || isImporting}
                className="px-md py-sm rounded-lg bg-[#4648d4] text-white font-label-md text-label-md shadow-sm hover:bg-[#6063ee] transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-xs"
              >
                {isImporting ? (
                  <span className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                ) : (
                  <span className="material-symbols-outlined text-[18px]">upload_file</span>
                )}
                <span>{isImporting ? 'Importing...' : `Import ${csvRows.length > 0 ? csvRows.length : ''} Records`}</span>
              </button>
            </div>
          </div>
        </div>
      )}

      {/* AI Schema Generation Modal */}
      {isAiSchemaOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center">
          <div
            onClick={() => setIsAiSchemaOpen(false)}
            className="fixed inset-0 bg-[#191c1d]/30 backdrop-blur-sm"
          ></div>
          <div className="relative bg-white rounded-xl shadow-2xl w-full max-w-lg mx-md z-10 flex flex-col max-h-[80vh]">
            <div className="flex items-center justify-between px-lg py-md border-b border-[#c7c4d7]/30">
              <div className="flex items-center gap-sm">
                <span className="material-symbols-outlined text-[#4648d4]">auto_awesome</span>
                <h3 className="font-headline-md text-headline-md font-bold text-[#191c1d]">Generate AI Schema</h3>
              </div>
              <button
                onClick={() => {
                  setIsAiSchemaOpen(false);
                  setAiPrompt('');
                  setAiResponse('');
                }}
                className="text-[#464554] hover:text-[#191c1d] hover:bg-[#f3f4f5] rounded-full p-xs transition-colors"
              >
                <span className="material-symbols-outlined">close</span>
              </button>
            </div>
            <div className="flex-1 overflow-y-auto p-lg space-y-lg">
              <div className="space-y-sm">
                <label className="block font-label-md text-label-md text-[#191c1d] font-semibold">
                  Describe your data
                </label>
                <textarea
                  className="w-full bg-white border border-[#c7c4d7] rounded-lg px-md py-sm text-[#191c1d] font-body-md text-body-md placeholder-[#464554]/50 focus:outline-none focus:border-[#4648d4] focus:ring-4 focus:ring-[#4648d4]/10 transition-all resize-y"
                  rows={4}
                  placeholder="e.g., A collection of blog posts with titles and content"
                  value={aiPrompt}
                  onChange={(e) => setAiPrompt(e.target.value)}
                />
                <p className="font-code-sm text-code-sm text-[#464554]/70">Describe the kind of records you want the AI to generate.</p>
              </div>

              {aiResponse && (
                <div className="space-y-sm">
                  <label className="block font-label-md text-label-md text-[#191c1d] font-semibold">Generated Schema</label>
                  <div className="bg-[#f3f4f5] rounded-lg p-md border border-[#c7c4d7]/30 font-code-sm text-code-sm text-[#191c1d] whitespace-pre-wrap break-words max-h-48 overflow-y-auto">
                    {aiResponse}
                  </div>
                </div>
              )}

              {isGenerating && (
                <div className="flex items-center justify-center py-lg space-x-md">
                  <span className="w-6 h-6 border-3 border-[#4648d4] border-t-transparent rounded-full animate-spin"></span>
                  <span className="font-body-md text-body-md text-[#464554]">Generating schema...</span>
                </div>
              )}
            </div>
            <div className="px-lg py-md border-t border-[#c7c4d7]/30 bg-[#f3f4f5] flex justify-end gap-md rounded-b-xl">
              <button
                onClick={() => {
                  setIsAiSchemaOpen(false);
                  setAiPrompt('');
                  setAiResponse('');
                }}
                className="px-md py-sm rounded-lg border border-[#c7c4d7] bg-white text-[#191c1d] font-label-md text-label-md hover:bg-[#f3f4f5] transition-colors"
              >
                Cancel
              </button>
              {aiResponse ? (
                <button
                  onClick={handleAiApply}
                  className="px-md py-sm rounded-lg bg-[#4648d4] text-white font-label-md text-label-md shadow-sm hover:bg-[#6063ee] transition-colors flex items-center gap-xs"
                >
                  <span className="material-symbols-outlined text-[18px]">add</span>
                  <span>Apply & Create Record</span>
                </button>
              ) : (
                <button
                  onClick={handleAiGenerate}
                  disabled={!aiPrompt.trim() || isGenerating}
                  className="px-md py-sm rounded-lg bg-[#4648d4] text-white font-label-md text-label-md shadow-sm hover:bg-[#6063ee] transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-xs"
                >
                  {isGenerating ? (
                    <span className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                  ) : (
                    <span className="material-symbols-outlined text-[18px]">auto_awesome</span>
                  )}
                  <span>{isGenerating ? 'Generating...' : 'Generate'}</span>
                </button>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
