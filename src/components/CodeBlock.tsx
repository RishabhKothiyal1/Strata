import React, { useState } from 'react';

export const CodeBlock: React.FC = () => {
  const [activeTab, setActiveTab] = useState<'db' | 'auth' | 'storage' | 'realtime'>('db');
  const [copied, setCopied] = useState(false);

  const codeSnippets = {
    db: `import { createClient } from "@strata/strata-js"

const strata = createClient({
  url: "https://project.strata.dev",
  apiKey: "strata_pk_live_..."
})

// Query all verified todo items
const { data, error } = await strata
  .from("todos")
  .select("*")
  .eq("is_verified", true)
  .order("created_at", { ascending: false })`,
    auth: `import { createClient } from "@strata/strata-js"

const strata = createClient({
  url: "https://project.strata.dev",
  apiKey: "strata_pk_live_..."
})

// Sign up a developer account
const { session, error } = await strata.auth.signUp({
  email: "developer@example.com",
  password: "securepassword123",
  options: {
    redirectTo: "https://strata.dev/welcome"
  }
})`,
    storage: `import { createClient } from "@strata/strata-js"

const strata = createClient({
  url: "https://project.strata.dev",
  apiKey: "strata_pk_live_..."
})

// Upload an avatar image file
const file = avatarInput.files[0]
const { path, error } = await strata.storage
  .from("avatars")
  .upload("profiles/user-12.png", file, {
    cacheControl: "3600",
    upsert: true
  })`,
    realtime: `import { createClient } from "@strata/strata-js"

const strata = createClient({
  url: "https://project.strata.dev",
  apiKey: "strata_pk_live_..."
})

// Listen to realtime database table changes
const channel = strata
  .channel("room-1")
  .on("postgres_changes", { event: "INSERT", schema: "public", table: "messages" }, (payload) => {
    console.log("New message: ", payload.new)
  })
  .subscribe()`
  };

  const handleCopy = () => {
    navigator.clipboard.writeText(codeSnippets[activeTab]);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <section className="py-24 max-w-7xl mx-auto px-md md:px-lg relative">
      <div className="grid grid-cols-1 lg:grid-cols-5 gap-xl items-center">
        
        {/* Left Side: Text Description */}
        <div className="lg:col-span-2 space-y-md text-left">
          <h2 className="font-headline-lg text-3xl md:text-4xl font-extrabold text-white">
            Write code. <br />
            <span className="text-[#f26500]">Get data.</span>
          </h2>
          <p className="font-body-md text-zinc-400 text-sm md:text-base leading-relaxed">
            The Strata SDK is built to feel natural. Interact with database rows, subscribe to realtime websocket broadcasts, manage files, or call edge endpoints with a single lightweight client.
          </p>
          <div className="flex flex-col gap-sm pt-sm">
            {[
              { key: 'db', label: 'Query PostgreSQL tables' },
              { key: 'auth', label: 'Configure secure signup OAuth' },
              { key: 'storage', label: 'Upload objects to S3 CDN' },
              { key: 'realtime', label: 'Subscribe to WebSocket streams' },
            ].map((feature) => (
              <button
                key={feature.key}
                onClick={() => setActiveTab(feature.key as any)}
                className={`flex items-center gap-xs font-semibold text-xs py-xs text-left focus:outline-none transition-colors ${
                  activeTab === feature.key ? 'text-[#f26500]' : 'text-zinc-500 hover:text-zinc-300'
                }`}
              >
                <span className="material-symbols-outlined text-sm">
                  {activeTab === feature.key ? 'radio_button_checked' : 'radio_button_unchecked'}
                </span>
                <span>{feature.label}</span>
              </button>
            ))}
          </div>
        </div>

        {/* Right Side: Code Block Mockup */}
        <div className="lg:col-span-3">
          <div className="bg-zinc-950 border border-white/5 rounded-2xl overflow-hidden shadow-2xl relative flex flex-col">
            
            {/* Tabs Header */}
            <div className="flex items-center justify-between px-md py-sm bg-zinc-900/60 border-b border-white/5 shrink-0 select-none">
              <div className="flex space-x-xs">
                {['db', 'auth', 'storage', 'realtime'].map((t) => (
                  <button
                    key={t}
                    onClick={() => setActiveTab(t as any)}
                    className={`font-code-sm text-[11px] font-bold py-xs px-sm rounded transition-colors ${
                      activeTab === t
                        ? 'bg-white/5 text-white'
                        : 'text-zinc-500 hover:text-zinc-300'
                    }`}
                  >
                    {t === 'db' ? 'Database.ts' : t === 'auth' ? 'Auth.ts' : t === 'storage' ? 'Storage.ts' : 'Realtime.ts'}
                  </button>
                ))}
              </div>

              {/* Copy button */}
              <button
                onClick={handleCopy}
                className="text-zinc-400 hover:text-white transition-colors cursor-pointer p-xs hover:bg-white/5 rounded"
                title="Copy code"
                aria-label="Copy code block"
              >
                <span className="material-symbols-outlined text-[18px]">
                  {copied ? 'check' : 'content_copy'}
                </span>
              </button>
            </div>

            {/* Code Content */}
            <div className="p-lg font-code-sm text-xs text-left overflow-x-auto min-h-[300px] select-text">
              <pre className="text-zinc-300 leading-relaxed font-code-sm">
                <code>
                  {codeSnippets[activeTab]}
                </code>
              </pre>
            </div>
          </div>
        </div>

      </div>
    </section>
  );
};
