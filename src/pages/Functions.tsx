import React, { useState } from 'react';
import { useOutletContext } from 'react-router-dom';
import type { LayoutContextType } from '../components/DashboardLayout';

interface Template {
  name: string;
  code: string;
  mockLogs: string[];
  mockResponse: string;
}

const TEMPLATES: Record<string, Template> = {
  jsonProcessor: {
    name: 'JSON Request Processor',
    code: `// Example ES5.1 Serverless Edge Function
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
}`,
    mockLogs: [
      'INFO: [runtime] Initializing Goja Javascript runtime environment.',
      'INFO: [runtime] Successfully compiled handleRequest (ES5.1 standard).',
      'INFO: [execution] Triggering handleRequest with payload { name: "Developer Partner" }.',
      'INFO: [user-code] Processing request for: Developer Partner',
      'SUCCESS: [runtime] Script successfully finished. Execution time: 1.42ms.',
    ],
    mockResponse: `{
  "status": 200,
  "headers": {
    "Content-Type": "application/json"
  },
  "body": {
    "message": "Hello Developer Partner from Strata Edge!",
    "timestamp": "2026-07-09T17:49:50.124Z"
  }
}`,
  },
  authGatekeeper: {
    name: 'Authorization Gatekeeper',
    code: `// Example Header Authorization Filter
function handleRequest(req) {
  var authHeader = req.headers["Authorization"];
  console.log("Validating security bearer header token...");
  
  if (!authHeader || authHeader.indexOf("Bearer ") !== 0) {
    console.warn("Security Alert: Invalid or missing authorization header!");
    return {
      status: 401,
      body: JSON.stringify({ error: "Unauthorized access token requested." })
    };
  }
  
  console.log("Token authentication approved by API Gateway CORS.");
  return {
    status: 200,
    body: JSON.stringify({ authorized: true, user: "authenticated-token-context" })
  };
}`,
    mockLogs: [
      'INFO: [runtime] Initializing Goja Javascript runtime environment.',
      'INFO: [runtime] Successfully compiled handleRequest (ES5.1 standard).',
      'INFO: [execution] Triggering handleRequest with security headers.',
      'WARN: [user-code] Security Alert: Invalid or missing authorization header!',
      'SUCCESS: [runtime] Security gatekeeper rejected token context. Status: 401.',
    ],
    mockResponse: `{
  "status": 401,
  "body": {
    "error": "Unauthorized access token requested."
  }
}`,
  },
  imageMetaExtractor: {
    name: 'Image Metadata Extractor',
    code: `// Process file upload payloads
function handleRequest(req) {
  var filename = req.query.filename || "upload.png";
  console.log("Analyzing file metadata attributes for: " + filename);
  
  var fileMeta = {
    original_name: filename,
    mime_type: "image/png",
    dimensions: "1920x1080",
    filter_preset: "Lanczos3-optimized",
    compressed: true
  };
  
  console.log("Compression applied, metadata extracted.");
  return {
    status: 200,
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(fileMeta)
  };
}`,
    mockLogs: [
      'INFO: [runtime] Initializing Goja Javascript runtime environment.',
      'INFO: [runtime] Successfully compiled handleRequest (ES5.1 standard).',
      'INFO: [execution] Triggering handleRequest with filename upload.png.',
      'INFO: [user-code] Analyzing file metadata attributes for: upload.png',
      'INFO: [user-code] Compression applied, metadata extracted.',
      'SUCCESS: [runtime] Image metadata extraction complete. Execution time: 2.11ms.',
    ],
    mockResponse: `{
  "status": 200,
  "headers": {
    "Content-Type": "application/json"
  },
  "body": {
    "original_name": "upload.png",
    "mime_type": "image/png",
    "dimensions": "1920x1080",
    "filter_preset": "Lanczos3-optimized",
    "compressed": true
  }
}`,
  },
};

export const Functions: React.FC = () => {
  const { showToast } = useOutletContext<LayoutContextType>();
  const [selectedTemplateKey, setSelectedTemplateKey] = useState<string>('jsonProcessor');
  const [code, setCode] = useState<string>(TEMPLATES.jsonProcessor.code);
  const [isRunning, setIsRunning] = useState<boolean>(false);
  const [executionOutput, setExecutionOutput] = useState<{ logs: string[]; response: string } | null>(null);

  // Template change handler
  const handleTemplateChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    const key = e.target.value;
    setSelectedTemplateKey(key);
    setCode(TEMPLATES[key].code);
    setExecutionOutput(null);
  };

  // Run Code simulation
  const handleRunExecution = () => {
    setIsRunning(true);
    setExecutionOutput(null);
    setTimeout(() => {
      setIsRunning(false);
      const template = TEMPLATES[selectedTemplateKey];
      setExecutionOutput({
        logs: template.mockLogs,
        response: template.mockResponse,
      });
      showToast('Function executed successfully.', 'success');
    }, 1200);
  };

  return (
    <div className="space-y-lg animate-fadeIn w-full">
      {/* Page Header */}
      <div className="flex flex-col md:flex-row md:items-center justify-between gap-md mb-md">
        <div>
          <h2 className="font-headline-lg text-headline-lg font-bold text-[#191c1d] mb-xs">
            Serverless Edge Functions
          </h2>
          <p className="font-body-md text-body-md text-[#464554]">
            Isolated JavaScript execution runtime powered by the pure Goja JS interpreter engine.
          </p>
        </div>

        {/* Action Header Button */}
        <button
          onClick={handleRunExecution}
          disabled={isRunning}
          className="bg-[#4648d4] hover:bg-[#6063ee] disabled:opacity-50 disabled:cursor-not-allowed text-white py-sm px-lg rounded-lg font-label-md text-label-md flex items-center space-x-sm shrink-0 shadow-sm cursor-pointer focus:outline-none focus:ring-2 focus:ring-[#4648d4]/50"
        >
          <span className="material-symbols-outlined text-md">bolt</span>
          <span>{isRunning ? 'Running...' : 'Run Test Execution'}</span>
        </button>
      </div>

      {/* Code Editor Area & Config */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-lg">
        
        {/* Code Editor Panel */}
        <div className="lg:col-span-2 bg-white rounded-xl shadow-sm p-lg border border-[#c7c4d7]/30 space-y-md flex flex-col">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-sm">
            <h3 className="font-headline-md text-headline-md font-bold text-[#191c1d]">
              JavaScript Code Editor
            </h3>
            
            {/* Template Selector dropdown */}
            <div className="flex items-center space-x-xs">
              <label className="text-xs font-semibold text-[#464554]" htmlFor="template-selector">
                Template:
              </label>
              <select
                id="template-selector"
                value={selectedTemplateKey}
                onChange={handleTemplateChange}
                className="bg-[#f3f4f5] border border-[#c7c4d7]/50 rounded px-sm py-xs text-xs text-[#191c1d] focus:outline-none focus:border-[#4648d4]"
              >
                {Object.entries(TEMPLATES).map(([key, val]) => (
                  <option key={key} value={key}>
                    {val.name}
                  </option>
                ))}
              </select>
            </div>
          </div>

          {/* Textarea code editor */}
          <div className="relative flex-1 min-h-[300px]">
            <textarea
              className="w-full h-full min-h-[300px] bg-[#1e1e1e] rounded-lg p-md font-code-sm text-code-sm text-green-400 focus:outline-none focus:ring-2 focus:ring-[#4648d4]/50 font-mono resize-y shadow-inner"
              value={code}
              onChange={(e) => setCode(e.target.value)}
              aria-label="JavaScript editor"
              spellCheck={false}
            />
          </div>
        </div>

        {/* Right Side: Runtime Context */}
        <div className="bg-white rounded-xl shadow-sm p-lg border border-[#c7c4d7]/30 flex flex-col justify-between space-y-md">
          <div>
            <h3 className="font-headline-md text-headline-md font-bold text-[#191c1d] mb-sm">
              Runtime Properties
            </h3>
            <p className="font-body-sm text-xs text-[#464554] mb-md leading-normal">
              Edge Runtime parameters and sandboxing limitations.
            </p>

            <div className="space-y-sm text-sm border-t border-[#c7c4d7]/20 pt-md">
              <div className="flex justify-between border-b border-[#c7c4d7]/10 pb-xs">
                <span className="text-[#464554]">JS Engine</span>
                <span className="font-semibold font-code-sm text-xs text-[#191c1d]">Goja (ES5.1 Engine)</span>
              </div>
              <div className="flex justify-between border-b border-[#c7c4d7]/10 pb-xs">
                <span className="text-[#464554]">Execution Timeout</span>
                <span className="font-semibold font-code-sm text-xs text-[#191c1d]">10.0 seconds</span>
              </div>
              <div className="flex justify-between border-b border-[#c7c4d7]/10 pb-xs">
                <span className="text-[#464554]">Memory Allocation</span>
                <span className="font-semibold text-[#191c1d]">64 MB Max RAM</span>
              </div>
              <div className="flex justify-between pb-xs">
                <span className="text-[#464554]">Logs Integration</span>
                <span className="font-semibold font-code-sm text-xs text-[#0a5c0a]">Go slog bindings</span>
              </div>
            </div>
          </div>

          <div className="space-y-sm">
            <button
              onClick={handleRunExecution}
              disabled={isRunning}
              className="w-full bg-[#e2dfff] hover:bg-[#c3c0ff] text-[#0f0069] py-sm px-md rounded-lg font-label-md text-label-md flex items-center justify-center space-x-sm cursor-pointer transition-colors"
            >
              <span className="material-symbols-outlined text-sm">bolt</span>
              <span>Execute Sandbox Test</span>
            </button>
            <div className="bg-[#ffeecb]/30 text-[#704200] border border-[#ffe0b2]/30 p-sm rounded-lg text-xs leading-normal">
              <strong>Notice:</strong> Network requests inside the interpreter are strictly routed through CORS proxies automatically.
            </div>
          </div>
        </div>
      </div>

      {/* Console output drawer */}
      {(isRunning || executionOutput) && (
        <div className="space-y-md animate-fadeIn">
          <h3 className="font-headline-md text-headline-md font-bold text-[#191c1d]">
            Execution Result
          </h3>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-md">
            {/* Logs console */}
            <div className="bg-[#18181b] border border-[#27272a] rounded-xl p-lg font-code-sm text-xs font-mono text-[#a1a1aa] space-y-xs overflow-y-auto max-h-[300px]">
              <span className="text-zinc-500 font-bold block mb-sm">SYSTEM LOGS:</span>
              {isRunning ? (
                <div className="flex items-center space-x-xs text-[#a1a1aa] py-sm">
                  <span className="w-3 h-3 border-2 border-[#a1a1aa] border-t-transparent rounded-full animate-spin shrink-0" />
                  <span>Invoking execution context...</span>
                </div>
              ) : (
                executionOutput?.logs.map((log, i) => {
                  let color = 'text-zinc-400';
                  if (log.includes('WARN')) color = 'text-yellow-500';
                  if (log.includes('SUCCESS')) color = 'text-green-500 font-bold';
                  return (
                    <span key={i} className={`block leading-relaxed ${color}`}>
                      {log}
                    </span>
                  );
                })
              )}
            </div>

            {/* Response Output JSON */}
            <div className="bg-[#1e1e1e] border border-[#333] rounded-xl p-lg font-code-sm text-xs font-mono text-green-400 overflow-y-auto max-h-[300px]">
              <span className="text-zinc-500 font-bold block mb-sm">RESPONSE BODY:</span>
              {isRunning ? (
                <span className="text-zinc-500">Awaiting runtime response payload...</span>
              ) : (
                <pre>{executionOutput?.response}</pre>
              )}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
