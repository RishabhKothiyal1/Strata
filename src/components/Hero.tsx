import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';

export const Hero: React.FC = () => {
  const [copied, setCopied] = useState(false);

  const handleCopy = () => {
    navigator.clipboard.writeText('npm install @strata/strata-js');
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  return (
    <section className="relative pt-32 pb-24 overflow-hidden flex flex-col items-center justify-center text-center px-md md:px-lg">
      <div className="max-w-4xl mx-auto space-y-xl z-10">
        
        {/* Badge Alert */}
        <motion.div
          initial={{ opacity: 0, y: 15 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5 }}
          className="inline-flex items-center gap-xs bg-white/5 border border-white/10 rounded-full px-sm py-xs text-xs font-semibold text-zinc-300"
        >
          <span className="w-1.5 h-1.5 rounded-full bg-[#f26500] animate-pulse" />
          <span>Strata Studio v1.0 Launching Soon</span>
          <span className="text-[#f26500] hover:underline cursor-pointer flex items-center">
            Read Docs <span className="material-symbols-outlined text-xs">arrow_forward</span>
          </span>
        </motion.div>

        {/* Headline */}
        <motion.h1
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.1 }}
          className="font-headline-lg text-4xl md:text-6xl font-extrabold text-white leading-tight tracking-tight max-w-3xl mx-auto"
        >
          Build in a weekend.{' '}
          <span className="bg-gradient-to-r from-[#f26500] to-[#ffa35a] bg-clip-text text-transparent block mt-sm md:inline md:mt-0">
            Scale to millions.
          </span>
        </motion.h1>

        {/* Subheading / Description */}
        <motion.p
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.2 }}
          className="font-body-lg text-zinc-400 max-w-2xl mx-auto text-base md:text-lg leading-relaxed"
        >
          Strata is the open-source Firebase alternative powered by PostgreSQL. 
          Deploy database schemas, handle authentication, manage object storage, 
          run serverless Goja edge runtimes, and listen to realtime WebSockets instantly.
        </motion.p>

        {/* Action Buttons */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, delay: 0.3 }}
          className="flex flex-col sm:flex-row items-center justify-center gap-md pt-xs"
        >
          <Link
            to="/dashboard"
            className="w-full sm:w-auto bg-[#f26500] hover:bg-[#ff7d26] text-white py-sm px-xl rounded-lg font-semibold shadow-md shadow-[#f26500]/25 transition-all text-center"
          >
            Start Building
          </Link>
          <a
            href="#pricing"
            className="w-full sm:w-auto border border-white/10 hover:border-white/20 bg-white/5 text-white py-sm px-xl rounded-lg font-semibold transition-all text-center hover:bg-white/10"
          >
            Book Demo
          </a>
        </motion.div>

        {/* Interactive Copyable Terminal Component */}
        <motion.div
          initial={{ opacity: 0, scale: 0.95 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.7, delay: 0.4 }}
          className="max-w-md mx-auto bg-zinc-950 border border-white/5 rounded-xl overflow-hidden shadow-2xl relative"
        >
          {/* Window Header */}
          <div className="flex items-center justify-between px-md py-sm bg-zinc-900/60 border-b border-white/5 select-none">
            <div className="flex space-x-xs">
              <span className="w-3 h-3 rounded-full bg-red-500/80 inline-block" />
              <span className="w-3 h-3 rounded-full bg-yellow-500/80 inline-block" />
              <span className="w-3 h-3 rounded-full bg-green-500/80 inline-block" />
            </div>
            <span className="font-code-sm text-[10px] text-zinc-500">bash</span>
            <div className="w-12" /> {/* Spacer */}
          </div>
          
          {/* Terminal Body */}
          <div className="p-lg flex items-center justify-between font-code-sm text-xs text-left">
            <div className="flex items-center gap-xs select-all">
              <span className="text-[#f26500] font-bold select-none">$</span>
              <span className="text-zinc-200">npm install @strata/strata-js</span>
            </div>
            
            <button
              onClick={handleCopy}
              className="text-zinc-400 hover:text-white transition-colors cursor-pointer p-xs hover:bg-white/5 rounded"
              title="Copy install command"
              aria-label="Copy code command"
            >
              <span className="material-symbols-outlined text-[18px]">
                {copied ? 'check' : 'content_copy'}
              </span>
            </button>
          </div>
        </motion.div>

      </div>
    </section>
  );
};
