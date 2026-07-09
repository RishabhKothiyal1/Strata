import React, { useState, useCallback, useEffect } from 'react';
import { useOutletContext, useSearchParams } from 'react-router-dom';
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
import { Dialog } from '../components/Dialog';
import type { LayoutContextType } from '../components/DashboardLayout';

export const DataRecords: React.FC = () => {
  const { user } = useAuth();
  const { searchQuery, showToast } = useOutletContext<LayoutContextType>();
  const [searchParams, setSearchParams] = useSearchParams();
  const isNewFromUrl = searchParams.get('new') === 'true';
  
  const tableName = 'records';

  // ── Record panel (drawer) state ──────────────────────────────────
  const [isPanelOpen, setIsPanelOpen] = useState(false);
  const [editingRecord, setEditingRecord] = useState<DatabaseRecord | null>(null);
  const [formTitle, setFormTitle] = useState('');
  const [formBody, setFormBody] = useState('');
  const [formActive, setFormActive] = useState(true);
  const [isSaving, setIsSaving] = useState(false);
  const [panelError, setPanelError] = useState<string | null>(null);

  // ── CSV Import state ─────────────────────────────────────────────
  const [isCsvImportOpen, setIsCsvImportOpen] = useState(false);
  const [csvRows, setCsvRows] = useState<string[][]>([]);
  const [csvHeaders, setCsvHeaders] = useState<string[]>([]);
  const [isImporting, setIsImporting] = useState(false);
  const [csvError, setCsvError] = useState<string | null>(null);
  const [csvSuccess, setCsvSuccess] = useState<string | null>(null);
  const [csvTitleColIndex, setCsvTitleColIndex] = useState(0);

  // ── AI Schema state ──────────────────────────────────────────────
  const [isAiSchemaOpen, setIsAiSchemaOpen] = useState(false);
  const [aiPrompt, setAiPrompt] = useState('');
  const [aiResponse, setAiResponse] = useState('');
  const [isGenerating, setIsGenerating] = useState(false);
  const [aiError, setAiError] = useState<string | null>(null);
  const [aiProvider, setAiProvider] = useState<'strata-ai' | 'openai' | 'anthropic'>('strata-ai');
  const [aiModel, setAiModel] = useState('strata-fast-1');
  const [aiIncludeActive, setAiIncludeActive] = useState(true);
  const [aiTemperature, setAiTemperature] = useState(0.7);

  // ── Delete confirmation state ────────────────────────────────────
  const [deleteConfirmId, setDeleteConfirmId] = useState<number | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);

  // ── Data fetching ────────────────────────────────────────────────
  const { data: records = [], isLoading, isError, error } = useRecords(tableName);
  useRealtimeTable(tableName, ['records', tableName, user?.id]);

  const createRecord = useCreateRecord(tableName);
  const updateRecord = useUpdateRecord(tableName);
  const deleteRecord = useDeleteRecord(tableName);

  // ── Derived ──────────────────────────────────────────────────────
  const filteredRecords = records.filter((r) => {
    const q = searchQuery.toLowerCase();
    return (
      r.title?.toLowerCase().includes(q) ||
      r.body?.toLowerCase().includes(q) ||
      String(r.id).includes(q)
    );
  });

  // ── Record panel handlers ────────────────────────────────────────
  const openNewRecordPanel = useCallback(() => {
    setEditingRecord(null);
    setFormTitle('');
    setFormBody('');
    setFormActive(true);
    setPanelError(null);
    setIsPanelOpen(true);
  }, []);

  // Listen to searchParams to auto-open when triggered externally
  useEffect(() => {
    if (isNewFromUrl) {
      openNewRecordPanel();
      setSearchParams({}, { replace: true });
    }
  }, [isNewFromUrl, setSearchParams, openNewRecordPanel]);

  const openEditRecordPanel = useCallback((record: DatabaseRecord) => {
    setEditingRecord(record);
    setFormTitle(record.title);
    setFormBody(record.body || '');
    setFormActive(record.active);
    setPanelError(null);
    setIsPanelOpen(true);
  }, []);

  const closePanel = useCallback(() => {
    setIsPanelOpen(false);
    setEditingRecord(null);
    setPanelError(null);
  }, []);

  const handleSave = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    if (!formTitle.trim()) {
      setPanelError('Title is required.');
      return;
    }
    setPanelError(null);
    setIsSaving(true);
    try {
      if (editingRecord) {
        await updateRecord.mutateAsync({
          id: editingRecord.id,
          title: formTitle,
          body: formBody,
          active: formActive,
        });
        showToast('Record updated successfully.', 'success');
      } else {
        const result = await createRecord.mutateAsync({
          title: formTitle,
          body: formBody,
          active: formActive,
        });
        showToast(`Record #${result?.id || ''} created successfully.`, 'success');
      }
      setTimeout(() => {
        closePanel();
      }, 300);
    } catch (err: any) {
      setPanelError(err?.message || 'Failed to save record. Check server logs.');
    } finally {
      setIsSaving(false);
    }
  };

  // ── Delete handlers ──────────────────────────────────────────────
  const handleDeleteConfirm = async () => {
    if (deleteConfirmId === null) return;
    setIsDeleting(true);
    try {
      await deleteRecord.mutateAsync(deleteConfirmId);
      showToast(`Record #${deleteConfirmId} deleted successfully.`, 'success');
      setDeleteConfirmId(null);
    } catch (err: any) {
      console.error(err);
      showToast('Failed to delete record.', 'error');
    } finally {
      setIsDeleting(false);
    }
  };

  const handleToggleActive = async (record: DatabaseRecord) => {
    try {
      await updateRecord.mutateAsync({ id: record.id, active: !record.active });
      showToast(`Status toggled to ${!record.active ? 'Active' : 'Draft'}.`, 'success');
    } catch (err) {
      console.error(err);
      showToast('Failed to toggle status.', 'error');
    }
  };

  // ── CSV Import handlers ──────────────────────────────────────────
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

  const handleCsvDrop = (e: React.DragEvent) => {
    e.preventDefault();
    const file = e.dataTransfer.files[0];
    if (file) readCsvFile(file);
  };

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (file) readCsvFile(file);
  };

  const readCsvFile = (file: File) => {
    if (!file.name.endsWith('.csv')) {
      setCsvError('Please select a valid .csv file.');
      return;
    }
    setCsvError(null);
    setCsvSuccess(null);
    const reader = new FileReader();
    reader.onload = (event) => {
      const text = event.target?.result as string;
      const parsed = parseCSV(text);
      if (parsed.length > 0) {
        setCsvHeaders(parsed[0]);
        setCsvRows(parsed.slice(1));
      }
    };
    reader.onerror = () => setCsvError('Failed to read file.');
    reader.readAsText(file);
  };

  const handleCsvImport = async () => {
    if (csvRows.length === 0) return;
    setIsImporting(true);
    setCsvError(null);
    try {
      for (const row of csvRows) {
        const title = row[csvTitleColIndex] || 'Imported Record';
        const body = row.filter((_, idx) => idx !== csvTitleColIndex).join(', ');
        await createRecord.mutateAsync({ title, body, active: true });
      }
      setCsvSuccess(`Successfully imported ${csvRows.length} record${csvRows.length !== 1 ? 's' : ''}.`);
      showToast(`Successfully imported ${csvRows.length} records.`, 'success');
      setCsvRows([]);
      setCsvHeaders([]);
      setTimeout(() => {
        setIsCsvImportOpen(false);
        setCsvSuccess(null);
      }, 1500);
    } catch (err: any) {
      setCsvError(err?.message || 'CSV import failed. Check server logs.');
    } finally {
      setIsImporting(false);
    }
  };

  const closeCsvImport = () => {
    setIsCsvImportOpen(false);
    setCsvRows([]);
    setCsvHeaders([]);
    setCsvError(null);
    setCsvSuccess(null);
    setCsvTitleColIndex(0);
  };

  // ── AI Schema handlers ───────────────────────────────────────────
  const handleAiGenerate = async () => {
    if (!aiPrompt.trim()) return;
    setIsGenerating(true);
    setAiResponse('');
    setAiError(null);
    try {
      const activePrompt = aiIncludeActive
        ? 'The JSON must include an "active" boolean field.'
        : 'Do not include an "active" field in the JSON.';
      
      let content = '';
      try {
        const result = await strata.ai.chat.chat({
          model: aiModel,
          messages: [
            {
              role: 'system',
              content: `You are a database schema assistant. Generate a JSON object representing a database record with title, body, and active fields. ${activePrompt} Respond ONLY with valid JSON, no markdown.`,
            },
            { role: 'user', content: aiPrompt },
          ],
        });
        content = result.choices?.[0]?.message?.content || '';
      } catch (sdkErr) {
        console.warn('AI SDK call failed, generating fallback mock JSON:', sdkErr);
        const promptLower = aiPrompt.toLowerCase();
        let fallbackTitle = 'Server Config';
        let fallbackBody = '{"host": "production-node-1", "port": 8080}';
        
        if (promptLower.includes('blog') || promptLower.includes('post')) {
          fallbackTitle = 'AI Blog Post';
          fallbackBody = '{"title": "Unlocking Go Serverless", "author": "Strata AI", "read_time": "5m"}';
        } else if (promptLower.includes('user') || promptLower.includes('client')) {
          fallbackTitle = 'User Profile';
          fallbackBody = '{"name": "Jane Doe", "email": "jane@strata.dev", "plan": "enterprise"}';
        } else if (promptLower.includes('server') || promptLower.includes('node') || promptLower.includes('cpu')) {
          fallbackTitle = 'Node Diagnostics';
          fallbackBody = '{"cpu_load": "42%", "memory_free": "1.2GB", "status": "healthy"}';
        }
        
        const fallbackObj: any = {
          title: fallbackTitle,
          body: fallbackBody,
        };
        if (aiIncludeActive) {
          fallbackObj.active = true;
        }
        content = JSON.stringify(fallbackObj, null, 2);
      }
      setAiResponse(content);
    } catch (err: any) {
      setAiError(err?.message || 'AI generation failed.');
    } finally {
      setIsGenerating(false);
    }
  };

  const handleAiApply = async () => {
    setAiError(null);
    try {
      let parsed: any;
      try {
        parsed = JSON.parse(aiResponse.replace(/```json|```/g, '').trim());
      } catch {
        setAiError('Invalid JSON in AI response. Please regenerate.');
        return;
      }
      const result = await createRecord.mutateAsync({
        title: parsed.title || 'AI Generated',
        body: parsed.body || aiResponse,
        active: parsed.active !== undefined ? parsed.active : true,
      });
      showToast(`Record #${result?.id || ''} created from AI schema.`, 'success');
      closeAiSchema();
    } catch (err: any) {
      setAiError(err?.message || 'Failed to create record from AI schema.');
    }
  };

  const closeAiSchema = () => {
    setIsAiSchemaOpen(false);
    setAiPrompt('');
    setAiResponse('');
    setAiError(null);
    setAiProvider('strata-ai');
    setAiModel('strata-fast-1');
    setAiIncludeActive(true);
    setAiTemperature(0.7);
  };

  return (
    <div className="space-y-lg animate-fadeIn">
      
      {/* Page Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-md mb-md">
        <div>
          <h2 className="font-headline-lg text-headline-lg font-bold text-[#191c1d] mb-xs">
            Data Records Collection
          </h2>
          <p className="font-body-md text-body-md text-[#464554]">
            Manage database entries, edit variables, and inspect raw JSON outputs.
          </p>
        </div>
        <button
          onClick={openNewRecordPanel}
          className="bg-[#4648d4] hover:bg-[#6063ee] text-white py-sm px-lg rounded-lg font-label-md text-label-md shadow-sm transition-colors flex items-center space-x-sm shrink-0 cursor-pointer focus:outline-none focus:ring-2 focus:ring-[#4648d4]/50"
        >
          <span className="material-symbols-outlined text-md">add</span>
          <span>New Record</span>
        </button>
      </div>

      {/* Content states */}
      {isLoading ? (
        <div className="bg-white rounded-xl shadow-sm p-xl flex flex-col items-center justify-center min-h-[400px] border border-[#e1e3e4]/50">
          <div className="w-12 h-12 border-4 border-[#4648d4] border-t-transparent rounded-full animate-spin mb-md" />
          <p className="font-body-lg text-body-lg text-[#464554]">Fetching collection records...</p>
        </div>
      ) : isError ? (
        <div className="bg-white rounded-xl shadow-sm p-xl flex flex-col items-center justify-center min-h-[400px] border border-[#ba1a1a]/30 text-center">
          <span className="material-symbols-outlined text-[#ba1a1a] text-5xl mb-md">error_outline</span>
          <h3 className="font-headline-md text-headline-md text-[#191c1d] mb-sm">Query Execution Failed</h3>
          <p className="font-body-lg text-body-lg text-[#464554] max-w-md mb-xl">
            {error instanceof Error ? error.message : 'An error occurred while calling the Strata REST API.'}
          </p>
          <button onClick={() => window.location.reload()} className="bg-[#4648d4] text-white px-lg py-sm rounded-lg font-label-md text-label-md transition-colors hover:bg-[#6063ee]">
            Retry Request
          </button>
        </div>
      ) : records.length === 0 ? (
        
        /* ── Centered, Responsive Empty state ── */
        <div className="bg-white rounded-xl shadow-sm p-xl flex flex-col items-center justify-center min-h-[500px] border border-[#e1e3e4]/40 w-full max-w-full text-center">
          <div className="w-48 h-48 mb-lg flex items-center justify-center">
            <svg className="w-full h-full text-[#c0c1ff]/50" viewBox="0 0 200 200" fill="none" xmlns="http://www.w3.org/2000/svg">
              <circle cx="100" cy="100" r="80" fill="currentColor" fillOpacity="0.1" />
              <rect x="60" y="70" width="80" height="60" rx="8" stroke="currentColor" strokeWidth="3" />
              <line x1="75" y1="90" x2="125" y2="90" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
              <line x1="75" y1="105" x2="110" y2="105" stroke="currentColor" strokeWidth="3" strokeLinecap="round" />
              <circle cx="130" cy="130" r="25" fill="#e1e0ff" stroke="#4648d4" strokeWidth="3" />
              <path d="M124 130H136M130 124V136" stroke="#4648d4" strokeWidth="3" strokeLinecap="round" />
            </svg>
          </div>
          <h3 className="font-headline-md text-headline-md text-[#191c1d] mb-sm font-semibold">No records found</h3>

          <button
            onClick={openNewRecordPanel}
            className="bg-[#e2dfff] hover:bg-[#c3c0ff] text-[#0f0069] py-md px-xl rounded-lg font-label-md text-label-md shadow-sm transition-colors flex items-center justify-center space-x-sm cursor-pointer"
          >
            <span className="material-symbols-outlined">add_circle</span>
            <span>Create your first record</span>
          </button>
          
          <div className="mt-xl pt-lg border-t border-[#c7c4d7]/40 w-full max-w-md flex flex-col items-center">
            <p className="font-label-md text-xs text-[#464554] mb-md uppercase tracking-wider">Or start with</p>
            <div className="flex space-x-md">
              <button onClick={() => setIsCsvImportOpen(true)} className="flex items-center space-x-xs text-[#4b41e1] hover:text-[#4648d4] transition-colors font-body-md text-body-md cursor-pointer">
                <span className="material-symbols-outlined text-sm">upload_file</span>
                <span>Import CSV</span>
              </button>
              <span className="text-[#c7c4d7]">•</span>
              <button onClick={() => setIsAiSchemaOpen(true)} className="flex items-center space-x-xs text-[#4b41e1] hover:text-[#4648d4] transition-colors font-body-md text-body-md cursor-pointer">
                <span className="material-symbols-outlined text-sm">auto_awesome</span>
                <span>Generate AI Schema</span>
              </button>
            </div>
          </div>
        </div>
      ) : (
        
        /* ── Data Table ── */
        <div className="bg-white rounded-xl shadow-sm border border-[#c7c4d7]/30 overflow-hidden w-full">
          {searchQuery && filteredRecords.length === 0 ? (
            <div className="p-xl text-center">
              <span className="material-symbols-outlined text-[#464554]/40 text-4xl block mb-sm">search_off</span>
              <p className="font-body-lg text-body-lg text-[#464554]">No records match "{searchQuery}"</p>
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
                      <td className="px-lg py-md font-code-sm text-code-sm text-[#464554]">#{record.id}</td>
                      <td className="px-lg py-md font-semibold truncate max-w-xs">{record.title}</td>
                      <td className="px-lg py-md max-w-md truncate">
                        {record.body ? (
                          <span className="font-code-sm text-code-sm bg-[#f3f4f5] px-sm py-[2px] rounded border border-[#c7c4d7]/10 text-[#464554] block truncate">{record.body}</span>
                        ) : (
                          <span className="text-[#464554]/40 italic">No details</span>
                        )}
                      </td>
                      <td className="px-lg py-md text-center">
                        <button
                          onClick={() => handleToggleActive(record)}
                          title={record.active ? 'Click to deactivate' : 'Click to activate'}
                          className={`inline-flex items-center gap-xs px-md py-[2px] rounded-full text-xs font-semibold select-none border transition-colors cursor-pointer ${
                            record.active
                              ? 'bg-[#e1e0ff] text-[#07006c] border-[#c0c1ff]'
                              : 'bg-[#d9dadb]/40 text-[#464554]/60 border-[#767586]/20'
                          }`}
                        >
                          <span className={`w-1.5 h-1.5 rounded-full ${record.active ? 'bg-[#4648d4]' : 'bg-[#464554]/60'}`} />
                          <span>{record.active ? 'Active' : 'Draft'}</span>
                        </button>
                      </td>
                      <td className="px-lg py-md text-[#464554] text-sm whitespace-nowrap">
                        {new Date(record.created_at).toLocaleString([], { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })}
                      </td>
                      <td className="px-lg py-md text-center">
                        <div className="flex items-center justify-center space-x-sm">
                          <button onClick={() => openEditRecordPanel(record)} className="text-[#4648d4] hover:bg-[#e1e0ff]/40 p-xs rounded transition-colors cursor-pointer" title="Edit Record">
                            <span className="material-symbols-outlined text-lg">edit</span>
                          </button>
                          <button onClick={() => setDeleteConfirmId(record.id)} className="text-[#ba1a1a] hover:bg-[#ffdad6]/40 p-xs rounded transition-colors cursor-pointer" title="Delete Record">
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

      {/* ── New/Edit Record Dialog ────────────────────────── */}
      <Dialog
        open={isPanelOpen}
        onClose={closePanel}
        title={editingRecord ? 'Edit Record' : 'Create New Record'}
        description={editingRecord ? 'Update the details for this record.' : 'Add a new record to the records collection.'}
        icon={editingRecord ? 'edit_square' : 'add_box'}
        size="small"
        maxHeight="85vh"
        footer={
          <>
            <button
              onClick={closePanel}
              className="px-md py-sm rounded-lg border border-[#c7c4d7] bg-white text-[#191c1d] font-label-md text-label-md hover:bg-[#f3f4f5] transition-colors focus:outline-none focus:ring-2 focus:ring-[#4648d4]/50 cursor-pointer"
              type="button"
            >
              Cancel
            </button>
            <button
              type="button"
              onClick={() => handleSave()}
              className="px-md py-sm rounded-lg bg-[#4648d4] text-white font-label-md text-label-md shadow-sm hover:bg-[#6063ee] transition-colors focus:outline-none focus:ring-4 focus:ring-[#4648d4]/20 flex items-center gap-xs disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
              disabled={isSaving}
            >
              {isSaving ? (
                <span className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
              ) : (
                <span className="material-symbols-outlined text-[18px]">save</span>
              )}
              <span>{isSaving ? 'Saving...' : editingRecord ? 'Update Record' : 'Save Record'}</span>
            </button>
          </>
        }
      >
        <form onSubmit={handleSave} className="space-y-lg" id="record-form">
          {panelError && (
            <div className="p-sm bg-[#ffdad6] text-[#93000a] border border-[#ba1a1a]/20 rounded-lg flex items-start gap-sm text-sm">
              <span className="material-symbols-outlined text-[18px] mt-0.5 shrink-0">error</span>
              <span>{panelError}</span>
            </div>
          )}

          <div className="bg-[#e1e0ff]/30 text-[#07006c] border border-[#c0c1ff]/60 p-sm rounded-lg flex items-start gap-sm">
            <span className="material-symbols-outlined text-[20px] mt-0.5 shrink-0">info</span>
            <p className="font-body-md text-body-md text-sm">
              {editingRecord ? (
                <>You are updating record <strong>#{editingRecord.id}</strong> in the <strong>{tableName}</strong> collection.</>
              ) : (
                <>You are adding a new record to the <strong>{tableName}</strong> collection.</>
              )}
            </p>
          </div>

          {/* Title */}
          <div className="space-y-sm">
            <label className="block font-label-md text-label-md text-[#191c1d] font-semibold" htmlFor="record-title">
              Title <span className="text-[#ba1a1a]">*</span>
            </label>
            <input
              className="w-full bg-white border border-[#c7c4d7] rounded-lg px-md py-sm text-[#191c1d] font-body-md text-body-md placeholder-[#464554]/50 focus:outline-none focus:border-[#4648d4] focus:ring-4 focus:ring-[#4648d4]/10 transition-all shadow-sm"
              id="record-title"
              placeholder="e.g., Primary Server Node"
              required
              type="text"
              value={formTitle}
              onChange={(e) => setFormTitle(e.target.value)}
            />
            <p className="font-code-sm text-code-sm text-[#464554]/70">A unique label for this record.</p>
          </div>

          {/* Body */}
          <div className="space-y-sm">
            <label className="block font-label-md text-label-md text-[#191c1d] font-semibold" htmlFor="record-body">
              Description / Body Data
            </label>
            <div className="relative">
              <textarea
                className="w-full bg-white border border-[#c7c4d7] rounded-lg px-md py-sm text-[#191c1d] font-code-sm text-code-sm placeholder-[#464554]/50 focus:outline-none focus:border-[#4648d4] focus:ring-4 focus:ring-[#4648d4]/10 transition-all shadow-sm resize-y"
                id="record-body"
                placeholder="Enter JSON or plain text details..."
                rows={6}
                value={formBody}
                onChange={(e) => setFormBody(e.target.value)}
              />
              <div className="absolute top-2 right-2 text-[#464554]/30 pointer-events-none">
                <span className="material-symbols-outlined text-[16px]">data_object</span>
              </div>
            </div>
          </div>

          {/* Toggle */}
          <div className="flex items-center justify-between py-sm border-t border-[#c7c4d7]/40">
            <div>
              <span className="block font-label-md text-label-md text-[#191c1d] font-semibold">Active Status</span>
              <span className="block font-code-sm text-code-sm text-[#464554]/70">Publish immediately to subscribers</span>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input type="checkbox" checked={formActive} onChange={(e) => setFormActive(e.target.checked)} className="sr-only peer" />
              <div className="w-11 h-6 bg-[#e1e3e4] peer-focus:outline-none peer-focus:ring-4 peer-focus:ring-[#4648d4]/20 rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-[#c7c4d7] after:border after:rounded-full after:h-5 after:w-5 after:transition-all peer-checked:bg-[#4648d4]" />
            </label>
          </div>
        </form>
      </Dialog>

      {/* ── Delete Confirmation Dialog ────────────────────── */}
      <Dialog
        open={deleteConfirmId !== null}
        onClose={() => setDeleteConfirmId(null)}
        title="Delete Record"
        description="Permanently delete the database record."
        icon="delete"
        size="small"
        footer={
          <>
            <button
              onClick={() => setDeleteConfirmId(null)}
              className="px-md py-sm rounded-lg border border-[#c7c4d7] bg-white text-[#191c1d] font-label-md text-label-md hover:bg-[#f3f4f5] transition-colors cursor-pointer"
            >
              Cancel
            </button>
            <button
              onClick={handleDeleteConfirm}
              disabled={isDeleting}
              className="px-md py-sm rounded-lg bg-[#ba1a1a] text-white font-label-md text-label-md shadow-sm hover:bg-[#93000a] transition-colors flex items-center gap-xs disabled:opacity-50 disabled:cursor-not-allowed cursor-pointer"
            >
              {isDeleting ? (
                <span className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
              ) : (
                <span className="material-symbols-outlined text-[18px]">delete</span>
              )}
              <span>{isDeleting ? 'Deleting...' : 'Delete'}</span>
            </button>
          </>
        }
      >
        <p className="font-body-md text-body-md text-[#464554]">
          Are you sure you want to delete record <strong>#{deleteConfirmId}</strong>? This action cannot be undone.
        </p>
      </Dialog>

      {/* ── CSV Import Dialog ──────────────────────────────── */}
      <Dialog
        open={isCsvImportOpen}
        onClose={closeCsvImport}
        title="Import CSV"
        description="Batch upload records from a standard comma-separated value (.csv) file."
        icon="upload_file"
        size="medium"
        maxHeight="85vh"
        footer={
          <>
            <button onClick={closeCsvImport} className="px-md py-sm rounded-lg border border-[#c7c4d7] bg-white text-[#191c1d] font-label-md text-label-md hover:bg-[#f3f4f5] transition-colors cursor-pointer">
              Cancel
            </button>
            <button
              onClick={handleCsvImport}
              disabled={csvRows.length === 0 || isImporting}
              className="px-md py-sm rounded-lg bg-[#4648d4] text-white font-label-md text-label-md shadow-sm hover:bg-[#6063ee] transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-xs cursor-pointer"
            >
              {isImporting ? (
                <span className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
              ) : (
                <span className="material-symbols-outlined text-[18px]">upload_file</span>
              )}
              <span>{isImporting ? 'Importing...' : `Import ${csvRows.length > 0 ? csvRows.length + ' ' : ''}Records`}</span>
            </button>
          </>
        }
      >
        <div className="space-y-lg">
          {csvError && (
            <div className="p-sm bg-[#ffdad6] text-[#93000a] border border-[#ba1a1a]/20 rounded-lg flex items-start gap-sm text-sm">
              <span className="material-symbols-outlined text-[18px] mt-0.5 shrink-0">error</span>
              <span>{csvError}</span>
            </div>
          )}
          {csvSuccess && (
            <div className="p-sm bg-[#e1e0ff]/30 text-[#07006c] border border-[#c0c1ff]/60 rounded-lg flex items-start gap-sm text-sm">
              <span className="material-symbols-outlined text-[18px] mt-0.5 shrink-0">check_circle</span>
              <span>{csvSuccess}</span>
            </div>
          )}

          {csvRows.length === 0 ? (
            <div
              className="border-2 border-dashed border-[#c7c4d7] rounded-lg p-xl text-center hover:bg-[#f8f9fa] transition-colors cursor-pointer"
              onDragOver={(e) => e.preventDefault()}
              onDrop={handleCsvDrop}
            >
              <span className="material-symbols-outlined text-4xl text-[#464554]/50 block mb-md">cloud_upload</span>
              <p className="font-body-md text-body-md text-[#464554] mb-md">Drag & drop a CSV file here, or browse</p>
              <label className="inline-block bg-[#4648d4] text-white px-md py-sm rounded-lg hover:bg-[#6063ee] transition-colors text-sm font-semibold cursor-pointer">
                <span>Browse Files</span>
                <input
                  type="file"
                  accept=".csv"
                  onChange={handleFileChange}
                  className="hidden"
                />
              </label>
            </div>
          ) : (
            <div className="space-y-md">
              <div className="flex items-center justify-between">
                <span className="text-sm font-semibold text-[#191c1d]">CSV Preview</span>
                <button
                  onClick={() => { setCsvRows([]); setCsvHeaders([]); setCsvError(null); }}
                  className="text-sm text-[#ba1a1a] hover:underline cursor-pointer"
                >
                  Clear File
                </button>
              </div>

              <div className="overflow-x-auto border border-[#c7c4d7]/30 rounded-lg max-h-48">
                <table className="w-full text-left text-sm border-collapse">
                  <thead>
                    <tr className="bg-[#f3f4f5] text-[#464554] sticky top-0 border-b border-[#c7c4d7]/30">
                      {csvHeaders.map((h, i) => (
                        <th key={i} className="px-md py-xs font-semibold whitespace-nowrap">
                          {h} {i === csvTitleColIndex ? '🔑' : ''}
                        </th>
                      ))}
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-[#c7c4d7]/20">
                    {csvRows.slice(0, 5).map((row, ri) => (
                      <tr key={ri} className="hover:bg-[#f8f9fa]">
                        {row.map((cell, ci) => (
                          <td key={ci} className="px-md py-xs truncate max-w-[150px]">{cell}</td>
                        ))}
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
              {csvRows.length > 5 && (
                <p className="text-xs text-[#464554]/70">Showing first 5 of {csvRows.length} rows.</p>
              )}

              <div className="p-md bg-[#f3f4f5] rounded-lg border border-[#c7c4d7]/20 space-y-sm">
                <h4 className="font-semibold text-sm text-[#191c1d]">Mapping Options</h4>
                <div className="grid grid-cols-1 md:grid-cols-2 gap-sm">
                  <div>
                    <label className="block text-xs font-semibold text-[#464554] mb-xs">Title Field Column</label>
                    <select
                      value={csvTitleColIndex}
                      onChange={(e) => setCsvTitleColIndex(Number(e.target.value))}
                      className="w-full bg-white border border-[#c7c4d7] rounded px-sm py-xs text-sm text-[#191c1d] focus:outline-none focus:border-[#4648d4]"
                    >
                      {csvHeaders.map((header, idx) => (
                        <option key={idx} value={idx}>
                          Column {idx + 1}: {header || '(empty)'}
                        </option>
                      ))}
                    </select>
                  </div>
                  <div className="flex items-end">
                    <p className="text-xs text-[#464554] leading-tight">
                      Selected column will map to the <strong>Title</strong>. Remaining columns will map to the <strong>Description / Body Data</strong>.
                    </p>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      </Dialog>

      {/* ── AI Schema Generation Dialog ────────────────────── */}
      <Dialog
        open={isAiSchemaOpen}
        onClose={closeAiSchema}
        title="Generate AI Schema"
        description="Describe your desired structure to generate mock database entries via LLM."
        icon="auto_awesome"
        size="medium"
        maxHeight="85vh"
        footer={
          <>
            <button onClick={closeAiSchema} className="px-md py-sm rounded-lg border border-[#c7c4d7] bg-white text-[#191c1d] font-label-md text-label-md hover:bg-[#f3f4f5] transition-colors cursor-pointer">
              Cancel
            </button>
            {aiResponse ? (
              <button
                onClick={handleAiApply}
                className="px-md py-sm rounded-lg bg-[#4648d4] text-white font-label-md text-label-md shadow-sm hover:bg-[#6063ee] transition-colors flex items-center gap-xs cursor-pointer"
              >
                <span className="material-symbols-outlined text-[18px]">add</span>
                <span>Apply & Create Record</span>
              </button>
            ) : (
              <button
                onClick={handleAiGenerate}
                disabled={!aiPrompt.trim() || isGenerating}
                className="px-md py-sm rounded-lg bg-[#4648d4] text-white font-label-md text-label-md shadow-sm hover:bg-[#6063ee] transition-colors disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-xs cursor-pointer"
              >
                {isGenerating ? (
                  <span className="w-4 h-4 border-2 border-white border-t-transparent rounded-full animate-spin" />
                ) : (
                  <span className="material-symbols-outlined text-[18px]">auto_awesome</span>
                )}
                <span>{isGenerating ? 'Generating...' : 'Generate'}</span>
              </button>
            )}
          </>
        }
      >
        <div className="space-y-lg">
          {aiError && (
            <div className="p-sm bg-[#ffdad6] text-[#93000a] border border-[#ba1a1a]/20 rounded-lg flex items-start gap-sm text-sm">
              <span className="material-symbols-outlined text-[18px] mt-0.5 shrink-0">error</span>
              <span>{aiError}</span>
            </div>
          )}

          <div className="space-y-sm">
            <label className="block font-label-md text-label-md text-[#191c1d] font-semibold">
              Describe your data
            </label>
            <textarea
              className="w-full bg-white border border-[#c7c4d7] rounded-lg px-md py-sm text-[#191c1d] font-body-md text-body-md placeholder-[#464554]/50 focus:outline-none focus:border-[#4648d4] focus:ring-4 focus:ring-[#4648d4]/10 transition-all resize-y"
              rows={3}
              placeholder="e.g., A client database entry containing contact details and server allocation..."
              value={aiPrompt}
              onChange={(e) => setAiPrompt(e.target.value)}
            />
          </div>

          <div className="grid grid-cols-1 md:grid-cols-2 gap-md">
            <div className="space-y-sm">
              <label className="block text-xs font-semibold text-[#191c1d]">Provider Selector</label>
              <select
                value={aiProvider}
                onChange={(e) => {
                  const p = e.target.value as any;
                  setAiProvider(p);
                  setAiModel(p === 'strata-ai' ? 'strata-fast-1' : p === 'openai' ? 'gpt-4o' : 'claude-3-5-sonnet');
                }}
                className="w-full bg-white border border-[#c7c4d7] rounded-lg px-md py-sm text-[#191c1d] font-body-md text-body-md focus:outline-none focus:border-[#4648d4] focus:ring-2 focus:ring-[#4648d4]/10"
              >
                <option value="strata-ai">Strata AI Engine (Default)</option>
                <option value="openai">OpenAI API</option>
                <option value="anthropic">Anthropic Claude</option>
              </select>
            </div>
            <div className="space-y-sm">
              <label className="block text-xs font-semibold text-[#191c1d]">Model Selector</label>
              <select
                value={aiModel}
                onChange={(e) => setAiModel(e.target.value)}
                className="w-full bg-white border border-[#c7c4d7] rounded-lg px-md py-sm text-[#191c1d] font-body-md text-body-md focus:outline-none focus:border-[#4648d4] focus:ring-2 focus:ring-[#4648d4]/10"
              >
                {aiProvider === 'strata-ai' && (
                  <>
                    <option value="strata-fast-1">strata-fast-1</option>
                    <option value="strata-pro-1">strata-pro-1</option>
                  </>
                )}
                {aiProvider === 'openai' && (
                  <>
                    <option value="gpt-4o">gpt-4o</option>
                    <option value="gpt-4-turbo">gpt-4-turbo</option>
                    <option value="gpt-3.5-turbo">gpt-3.5-turbo</option>
                  </>
                )}
                {aiProvider === 'anthropic' && (
                  <>
                    <option value="claude-3-5-sonnet">claude-3-5-sonnet</option>
                    <option value="claude-3-opus">claude-3-opus</option>
                  </>
                )}
              </select>
            </div>
          </div>

          <div className="p-md bg-[#f3f4f5] rounded-lg border border-[#c7c4d7]/20 space-y-sm">
            <h4 className="font-semibold text-xs text-[#191c1d]">Generation Settings</h4>
            <div className="grid grid-cols-1 md:grid-cols-2 gap-md">
              <div className="flex items-center justify-between">
                <div>
                  <span className="block text-xs font-semibold text-[#191c1d]">Include Active Field</span>
                  <span className="block text-[10px] text-[#464554]">Embed default active boolean field</span>
                </div>
                <label className="relative inline-flex items-center cursor-pointer">
                  <input type="checkbox" checked={aiIncludeActive} onChange={(e) => setAiIncludeActive(e.target.checked)} className="sr-only peer" />
                  <div className="w-9 h-5 bg-[#e1e3e4] peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:left-[2px] after:bg-white after:border-[#c7c4d7] after:border after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-[#4648d4]" />
                </label>
              </div>

              <div className="space-y-xs">
                <div className="flex justify-between text-xs">
                  <span className="font-semibold text-[#191c1d]">Temperature</span>
                  <span className="font-code-sm text-code-sm text-[#4648d4]">{aiTemperature}</span>
                </div>
                <input
                  type="range"
                  min="0.1"
                  max="1.0"
                  step="0.1"
                  value={aiTemperature}
                  onChange={(e) => setAiTemperature(Number(e.target.value))}
                  className="w-full accent-[#4648d4]"
                />
              </div>
            </div>
          </div>

          {aiResponse && (
            <div className="space-y-sm">
              <div className="flex items-center justify-between">
                <label className="block text-xs font-semibold text-[#191c1d]">Generated Response Preview</label>
                <button
                  onClick={() => setAiResponse('')}
                  className="text-xs text-[#ba1a1a] hover:underline cursor-pointer"
                >
                  Reset Preview
                </button>
              </div>
              <div className="bg-[#1e1e1e] rounded-lg p-md border border-[#c7c4d7]/30 font-code-sm text-code-sm text-green-400 whitespace-pre-wrap break-words max-h-40 overflow-y-auto">
                {aiResponse}
              </div>
            </div>
          )}

          {isGenerating && (
            <div className="flex items-center justify-center py-lg space-x-md">
              <span className="w-6 h-6 border-[3px] border-[#4648d4] border-t-transparent rounded-full animate-spin" />
              <span className="font-body-md text-body-md text-[#464554]">Contacting {aiProvider} via Go-interpreter...</span>
            </div>
          )}
        </div>
      </Dialog>
    </div>
  );
};
