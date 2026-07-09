import React, { useState } from 'react';
import { motion } from 'framer-motion';

export const DashboardPreview: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'editor' | 'table' | 'keys' | 'metrics'>('editor');

  return (
    <section className="py-24 max-w-7xl mx-auto px-md md:px-lg relative">
      <div className="text-center max-w-2xl mx-auto mb-16 space-y-sm">
        <h2 className="font-headline-lg text-3xl md:text-4xl font-extrabold text-white">
          Powerful administrative console, <span className="text-[#f26500]">built in.</span>
        </h2>
        <p className="font-body-md text-zinc-400 text-sm md:text-base leading-relaxed">
          Manage database schema tables, execute quick SQL scripts, generate API key pairs, and review execution logs instantly.
        </p>
      </div>

      {/* ── Browser Mockup Window ── */}
      <motion.div
        initial={{ opacity: 0, y: 40 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true, margin: '-100px' }}
        transition={{ duration: 0.8 }}
        className="w-full bg-[#0d0d0f] border border-white/5 rounded-2xl overflow-hidden shadow-2xl flex flex-col relative"
      >
        {/* Browser Top Bar */}
        <div className="flex items-center justify-between px-md py-sm bg-zinc-900/60 border-b border-white/5 select-none shrink-0">
          <div className="flex items-center space-x-xs">
            <span className="w-3 h-3 rounded-full bg-red-500/50 inline-block" />
            <span className="w-3 h-3 rounded-full bg-yellow-500/50 inline-block" />
            <span className="w-3 h-3 rounded-full bg-green-500/50 inline-block" />
          </div>
          
          {/* Address Bar */}
          <div className="bg-zinc-950/80 border border-white/5 px-md py-[4px] rounded-lg text-[10px] text-zinc-400 flex items-center justify-center space-x-xs w-64 md:w-96">
            <span className="material-symbols-outlined text-[10px] text-[#f26500]">lock</span>
            <span className="truncate">console.strata.dev/project/prod-strata</span>
          </div>

          <div className="w-12" /> {/* Spacer */}
        </div>

        {/* Dashboard Shell Grid */}
        <div className="flex flex-1 min-h-[500px] text-left">
          
          {/* Console Sidebar */}
          <aside className="w-48 bg-zinc-950 border-r border-white/5 p-md flex flex-col justify-between shrink-0 hidden md:flex">
            <div className="space-y-lg">
              <div className="flex items-center gap-xs px-xs">
                <span className="material-symbols-outlined text-md text-[#f26500]">layers</span>
                <span className="font-semibold text-xs text-white">Strata Studio</span>
              </div>

              <div className="space-y-xs">
                {[
                  { label: 'SQL Editor', icon: 'terminal', key: 'editor' },
                  { label: 'Table Editor', icon: 'table_chart', key: 'table' },
                  { label: 'API Keys', icon: 'vpn_key', key: 'keys' },
                  { label: 'Performance', icon: 'insights', key: 'metrics' },
                ].map((tab) => (
                  <button
                    key={tab.key}
                    onClick={() => setActiveTab(tab.key as any)}
                    className={`w-full flex items-center gap-xs p-xs rounded-md text-[11px] font-semibold transition-colors text-left ${
                      activeTab === tab.key
                        ? 'bg-white/5 text-white'
                        : 'text-zinc-400 hover:text-white hover:bg-white/5'
                    }`}
                  >
                    <span className="material-symbols-outlined text-[14px]">{tab.icon}</span>
                    <span>{tab.label}</span>
                  </button>
                ))}
              </div>
            </div>

            <div className="space-y-sm border-t border-white/5 pt-md">
              <div className="flex items-center gap-xs px-xs text-[10px] text-zinc-500">
                <span className="w-1.5 h-1.5 rounded-full bg-green-500" />
                <span>Connected to DB</span>
              </div>
            </div>
          </aside>

          {/* Console Main Content Panel */}
          <main className="flex-1 bg-zinc-900/10 p-lg overflow-x-auto min-w-0">
            
            {/* Tab: SQL Editor */}
            {activeTab === 'editor' && (
              <div className="space-y-md h-full flex flex-col">
                <div className="flex items-center justify-between border-b border-white/5 pb-sm">
                  <span className="font-headline-md text-xs font-bold text-white flex items-center gap-xs">
                    <span className="material-symbols-outlined text-sm text-[#f26500]">terminal</span>
                    <span>query_users.sql</span>
                  </span>
                  <button className="bg-[#f26500] hover:bg-[#ff7d26] text-white py-xs px-sm rounded text-[10px] font-semibold shadow flex items-center gap-xs">
                    <span className="material-symbols-outlined text-xs">play_arrow</span>
                    <span>Run Query</span>
                  </button>
                </div>
                
                {/* Editor code area */}
                <div className="flex-1 font-code-sm text-xs bg-zinc-950 p-md rounded-lg border border-white/5 min-h-[220px]">
                  <p className="text-zinc-500 mb-xs">-- Fetch all verified users active in the last 24 hours</p>
                  <p className="text-[#ffa35a]"><span className="text-[#f26500]">SELECT</span> id, email, created_at, role</p>
                  <p className="text-[#ffa35a]"><span className="text-[#f26500]">FROM</span> strata_users</p>
                  <p className="text-[#ffa35a]"><span className="text-[#f26500]">WHERE</span> is_verified = <span className="text-emerald-400">true</span></p>
                  <p className="text-[#ffa35a]"><span className="text-[#f26500]">ORDER BY</span> created_at <span className="text-[#f26500]">DESC</span></p>
                  <p className="text-[#ffa35a]"><span className="text-[#f26500]">LIMIT</span> <span className="text-purple-400">3</span>;</p>
                </div>

                {/* Console output log */}
                <div className="bg-zinc-950/80 border border-white/5 rounded-lg p-md font-code-sm text-[10px] text-zinc-400 space-y-xs">
                  <p className="text-emerald-400">✔ Query executed successfully in 12ms</p>
                  <p className="text-zinc-500">Output: 3 rows returned.</p>
                </div>
              </div>
            )}

            {/* Tab: Table Editor */}
            {activeTab === 'table' && (
              <div className="space-y-md">
                <div className="flex items-center justify-between border-b border-white/5 pb-sm">
                  <span className="font-headline-md text-xs font-bold text-white flex items-center gap-xs">
                    <span className="material-symbols-outlined text-sm text-[#f26500]">table_chart</span>
                    <span>strata_users</span>
                  </span>
                  <div className="flex items-center gap-xs">
                    <button className="border border-white/5 bg-white/5 text-zinc-300 py-xs px-sm rounded text-[10px] hover:bg-white/10 transition-colors">
                      + Insert Row
                    </button>
                  </div>
                </div>

                <div className="bg-zinc-950 border border-white/5 rounded-lg overflow-hidden">
                  <table className="w-full border-collapse text-left text-[11px]">
                    <thead>
                      <tr className="bg-zinc-900/80 border-b border-white/5 text-zinc-400 uppercase tracking-wider font-semibold">
                        <th className="px-md py-sm">id</th>
                        <th className="px-md py-sm">email</th>
                        <th className="px-md py-sm">role</th>
                        <th className="px-md py-sm">is_verified</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-white/5 text-zinc-300">
                      {[
                        { id: '1', email: 'alice@strata.dev', role: 'admin', verified: 'true' },
                        { id: '2', email: 'bob@example.com', role: 'developer', verified: 'true' },
                        { id: '3', email: 'charlie@gmail.com', role: 'user', verified: 'false' },
                      ].map((u) => (
                        <tr key={u.id} className="hover:bg-white/5">
                          <td className="px-md py-sm font-code-sm text-[#f26500]">#{u.id}</td>
                          <td className="px-md py-sm font-semibold">{u.email}</td>
                          <td className="px-md py-sm">{u.role}</td>
                          <td className="px-md py-sm">
                            <span className={`px-sm py-[1px] rounded-full text-[9px] font-bold ${
                              u.verified === 'true' ? 'bg-emerald-500/10 text-emerald-400' : 'bg-zinc-800 text-zinc-500'
                            }`}>{u.verified}</span>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            )}

            {/* Tab: API Keys */}
            {activeTab === 'keys' && (
              <div className="space-y-md">
                <div className="flex items-center justify-between border-b border-white/5 pb-sm">
                  <span className="font-headline-md text-xs font-bold text-white flex items-center gap-xs">
                    <span className="material-symbols-outlined text-sm text-[#f26500]">vpn_key</span>
                    <span>API Credentials</span>
                  </span>
                  <button className="bg-[#f26500] hover:bg-[#ff7d26] text-white py-xs px-sm rounded text-[10px] font-semibold shadow">
                    + Generate API Key
                  </button>
                </div>

                <div className="space-y-sm">
                  {[
                    { name: 'Production Frontend Client Key', key: 'strata_pk_live_f893d...j492a', role: 'Read-only' },
                    { name: 'Admin Goja Function Exec', key: 'strata_sk_live_2348b...k384a', role: 'Full Access' },
                  ].map((k) => (
                    <div key={k.name} className="bg-zinc-950 border border-white/5 rounded-lg p-md flex items-center justify-between text-xs">
                      <div>
                        <p className="font-bold text-white mb-xs">{k.name}</p>
                        <p className="font-code-sm text-zinc-500 text-[10px]">{k.key}</p>
                      </div>
                      <span className="bg-zinc-900 border border-white/5 text-[#f26500] text-[10px] font-bold py-xs px-sm rounded-full">
                        {k.role}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Tab: Performance */}
            {activeTab === 'metrics' && (
              <div className="space-y-md">
                <div className="flex items-center justify-between border-b border-white/5 pb-sm">
                  <span className="font-headline-md text-xs font-bold text-white flex items-center gap-xs">
                    <span className="material-symbols-outlined text-sm text-[#f26500]">insights</span>
                    <span>Performance Analytics</span>
                  </span>
                </div>

                {/* Simulated Chart */}
                <div className="grid grid-cols-3 gap-md">
                  <div className="bg-zinc-950 border border-white/5 p-md rounded-lg text-center space-y-xs">
                    <span className="text-[10px] text-zinc-500 block uppercase font-bold">API Gateway Response</span>
                    <span className="font-extrabold text-white text-md block">14ms</span>
                  </div>
                  <div className="bg-zinc-950 border border-white/5 p-md rounded-lg text-center space-y-xs">
                    <span className="text-[10px] text-zinc-500 block uppercase font-bold">Active WebSockets</span>
                    <span className="font-extrabold text-white text-md block">4,289</span>
                  </div>
                  <div className="bg-zinc-950 border border-white/5 p-md rounded-lg text-center space-y-xs">
                    <span className="text-[10px] text-zinc-500 block uppercase font-bold">DB CPU Load</span>
                    <span className="font-extrabold text-white text-md block">1.8%</span>
                  </div>
                </div>
                
                {/* Mock Chart Area */}
                <div className="w-full h-32 bg-zinc-950 border border-white/5 rounded-lg p-md flex items-end justify-between relative overflow-hidden">
                  <div className="absolute top-sm left-sm text-[9px] text-zinc-600">Requests per minute (last 24h)</div>
                  {[20, 30, 25, 45, 55, 40, 60, 80, 75, 90, 85, 95, 100, 90, 110, 120, 100].map((h, idx) => (
                    <div
                      key={idx}
                      className="bg-[#f26500]/60 hover:bg-[#f26500] rounded-t-sm w-[4%] transition-all"
                      style={{ height: `${h}%` }}
                    />
                  ))}
                </div>
              </div>
            )}

          </main>
        </div>

      </motion.div>
    </section>
  );
};
