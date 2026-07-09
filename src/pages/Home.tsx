import React, { useEffect, useState } from 'react';
import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';

// Import components
import { BackgroundEffects } from '../components/BackgroundEffects';
import { Navbar } from '../components/Navbar';
import { Hero } from '../components/Hero';
import { FeatureCard } from '../components/FeatureCard';
import { DashboardPreview } from '../components/DashboardPreview';
import { CodeBlock } from '../components/CodeBlock';
import { Templates } from '../components/Templates';
import { Comparison } from '../components/Comparison';
import { Stats } from '../components/Stats';
import { Testimonials } from '../components/Testimonials';
import { Footer } from '../components/Footer';

export const Home: React.FC = () => {
  // Update SEO metadata on mount
  useEffect(() => {
    document.title = 'Strata — The Open Source Backend-as-a-Service';
    
    // Update meta description
    const metaDescription = document.querySelector('meta[name="description"]');
    if (metaDescription) {
      metaDescription.setAttribute('content', 'Strata provides PostgreSQL, Auth, S3 Storage, Goja Edge Functions, and WebSockets Realtime instantly for developers. MIT licensed and fully self-hostable.');
    }
  }, []);

  // CLI terminal animated player states
  const [, setTerminalStep] = useState(0);
  const [terminalLines, setTerminalLines] = useState<string[]>([]);

  useEffect(() => {
    let active = true;
    const steps = [
      {
        cmd: 'strata init',
        output: [
          '⚡ Initializing Strata workspace...',
          '✔ Created strata.config.json',
          '✔ Created strata/schemas/schema.json',
          '✔ Created strata/functions/index.js',
          '✔ Workspace ready. Active database: strata-prod-db'
        ]
      },
      {
        cmd: 'strata db push',
        output: [
          '⚙ Compiling schema changes...',
          '✔ PostgreSQL migration strata_01_init compiled',
          '✔ Table strata_users created successfully',
          '✔ Row-level security (RLS) policies synchronized'
        ]
      },
      {
        cmd: 'strata deploy',
        output: [
          '📦 Bundling edge scripts...',
          '⚡ Deploying to global edge gateway routes...',
          '✔ Functions uploaded successfully (size: 42.4kb)',
          '✔ Active URL: https://functions.strata.dev/exec/prod-run'
        ]
      },
      {
        cmd: 'strata logs',
        output: [
          '监听 Streaming live edge server logs...',
          '[12:44:02 UTC] GET /api/v1/todos - Status: 200 OK (12ms)',
          '[12:44:05 UTC] POST /api/v1/auth/signup - Status: 201 Created (48ms)',
          '[12:44:12 UTC] WebSocket Connected - client_id: ws_842e'
        ]
      }
    ];

    let timer: any;

    const runScript = (index: number) => {
      if (!active) return;
      if (index >= steps.length) {
        // Reset cycle
        timer = setTimeout(() => {
          if (!active) return;
          setTerminalLines([]);
          setTerminalStep(0);
          runScript(0);
        }, 5000);
        return;
      }

      setTerminalLines(prev => [...prev, `strata_cli $ ${steps[index].cmd}`]);
      
      let lineIdx = 0;
      const printLines = () => {
        if (!active) return;
        if (lineIdx < steps[index].output.length) {
          const val = steps[index].output[lineIdx];
          if (val) {
            setTerminalLines(prev => [...prev, val]);
          }
          lineIdx++;
          timer = setTimeout(printLines, 600);
        } else {
          // Pause between commands
          setTerminalStep(index + 1);
          timer = setTimeout(() => {
            if (!active) return;
            runScript(index + 1);
          }, 2500);
        }
      };
      
      timer = setTimeout(printLines, 800);
    };

    runScript(0);

    return () => {
      active = false;
      clearTimeout(timer);
    };
  }, []);

  const marqueeLogos = [
    { name: 'GitHub', path: 'M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z' },
    { name: 'Stripe', path: 'M20 10.3c0-3.3-1.6-4.9-4.8-4.9-3.2 0-5.3 1.9-5.3 5 0 4.3 5.8 3.6 5.8 5.6 0 .7-.6 1.1-1.5 1.1-1.3 0-3-.5-4.2-1.2l-.7 2.8c1.3.7 3.3 1.2 5.1 1.2 3.3 0 5.1-1.7 5.1-5 0-4.3-5.8-3.7-5.8-5.6 0-.6.5-1 1.3-1 1 0 2.5.4 3.5.9l.6-2.9zm6.6.6c0-2 1.4-2.8 3.5-2.8 1.1 0 1.9.2 2.4.5V18c-.6.3-1.5.5-2.4.5-2.3 0-3.5-1-3.5-3.3v-4.3zm6-.1v-2c-.5-.2-1-.3-1.5-.3-1.2 0-1.7.5-1.7 1.6V18c0 .2.1.3.3.3h2.6v-7.4zm-14.7-6h-3v3h3v-3z' },
    { name: 'Vercel', path: 'M12 2L2 22h20L12 2z' },
    { name: 'Cloudflare', path: 'M19.3 12.8c.2.4.3.9.3 1.4 0 .9-.5 1.7-1.3 2.1-.8.4-1.8.3-2.5-.2-.7.6-1.7.8-2.6.5-.9-.3-1.5-1-1.7-1.9-.3.6-.9 1.1-1.6 1.2-.7.1-1.5-.1-2-.6-.6.6-1.5.8-2.3.5-.8-.3-1.4-1-1.6-1.9H3c-.6 0-1.1-.3-1.4-.8s-.3-1.1 0-1.6l1.2-2.1c.3-.5.9-.8 1.5-.8h2.6c.4-.9 1.2-1.6 2.2-1.8 1-.2 2.1 0 2.9.7.7-.6 1.6-.9 2.6-.7 1 .2 1.7.9 2.1 1.8 1-.9 2.5-1.1 3.7-.4 1.2.7 1.8 2.1 1.5 3.5h.3c.7.1 1.4.5 1.8 1.2.4.7.4 1.6 0 2.3l-1.2 2.1z' },
    { name: 'Figma', path: 'M12 24c3.314 0 6-2.686 6-6v-6c0-3.314-2.686-6-6-6s-6 2.686-6 6v6c0 3.314 2.686 6 6 6zm0-18c1.657 0 3 1.343 3 3v3H9V9c0-1.657 1.343-3 3-3zm-3 9v-3h6v3H9zm3 3c-1.657 0-3-1.343-3-3v-3h6v3c0 1.657-1.343 3-3 3z' },
    { name: 'Docker', path: 'M2 13h2v-2H2v2zm3 0h2v-2H5v2zm0-3h2V8H5v2zm3 3h2v-2H8v2zm0-3h2V8H8v2zm0-3h2V5H8v2zm3 6h2v-2h-2v2zm0-3h2V8h-2v2zm3 3h2v-12h-2v12z' },
    { name: 'Netlify', path: 'M20.2 12.8L12 21.2 3.8 12.8C2 11 2 8 3.8 6.2s4.8-1.8 6.6 0L12 7.8l1.6-1.6c1.8-1.8 4.8-1.8 6.6 0s1.8 4.8 0 6.6z' },
    { name: 'Railway', path: 'M2 2h20v20H2V2zm4 4v12h12V6H6z' }
  ];

  return (
    <div className="bg-[#030303] text-zinc-200 min-h-screen relative overflow-hidden select-none">
      {/* Dynamic Grid Background Canvas */}
      <BackgroundEffects />

      {/* Header Sticky Navbar */}
      <Navbar />

      {/* Central Hero Block */}
      <div className="relative z-10">
        <Hero />

        {/* ── Infinite Marquee Brand Section ── */}
        <section className="py-12 border-t border-b border-white/5 bg-zinc-950/40 relative overflow-hidden select-none">
          <div className="max-w-7xl mx-auto px-md md:px-lg mb-sm text-center">
            <p className="font-label-md text-[10px] text-zinc-500 uppercase tracking-widest font-bold">
              Trusted by world-class developer organizations
            </p>
          </div>
          
          <div className="w-full relative overflow-hidden flex mt-md">
            {/* Fade overlays on edges */}
            <div className="absolute left-0 top-0 bottom-0 w-16 md:w-32 bg-gradient-to-r from-[#030303] to-transparent z-20 pointer-events-none" />
            <div className="absolute right-0 top-0 bottom-0 w-16 md:w-32 bg-gradient-to-l from-[#030303] to-transparent z-20 pointer-events-none" />

            <div className="animate-marquee flex items-center space-x-xl md:space-x-2xl">
              {/* Double array to handle loop transitions */}
              {[...marqueeLogos, ...marqueeLogos, ...marqueeLogos].map((logo, idx) => (
                <div
                  key={`${logo.name}-${idx}`}
                  className="flex items-center gap-xs text-zinc-500 hover:text-white transition-colors duration-250 cursor-pointer select-none shrink-0 group"
                >
                  <svg className="w-5 h-5 fill-current" viewBox="0 0 24 24">
                    <path d={logo.path} />
                  </svg>
                  <span className="font-label-md text-xs font-semibold select-none">{logo.name}</span>
                </div>
              ))}
            </div>
          </div>
        </section>

        {/* Modular Feature Cards Grid */}
        <FeatureCard />

        {/* Modular fake console showcase mockup */}
        <DashboardPreview />

        {/* Tabbed SDK Code playground Block */}
        <CodeBlock />

        {/* Starter boilerplates Starters Grid */}
        <Templates />

        {/* Feature comparison Matrix Table */}
        <Comparison />

        {/* ── Developer Experience CLI Terminal split section ── */}
        <section className="py-24 max-w-7xl mx-auto px-md md:px-lg relative">
          <div className="grid grid-cols-1 lg:grid-cols-5 gap-xl items-center">
            
            {/* Left Side: Developer workflow benefits */}
            <div className="lg:col-span-2 text-left space-y-md">
              <h2 className="font-headline-lg text-3xl md:text-4xl font-extrabold text-white">
                Developer workflow, <br />
                <span className="text-[#f26500]">accelerated.</span>
              </h2>
              <p className="font-body-md text-zinc-400 text-sm md:text-base leading-relaxed">
                Strata features an intuitive Command Line Interface (CLI). Initialize environments locally, execute secure migrations, deploy serverless code runtime bundles, and stream console container logs.
              </p>
              
              <div className="space-y-sm pt-sm font-body-md text-sm text-zinc-300">
                <div className="flex items-start gap-xs">
                  <span className="material-symbols-outlined text-emerald-400 text-sm mt-0.5">check_circle</span>
                  <div>
                    <strong className="text-white">Declarative config</strong>
                    <p className="text-xs text-zinc-500">Configure CORS rules, Backups, and Gateway mappings directly via JSON files.</p>
                  </div>
                </div>
                <div className="flex items-start gap-xs">
                  <span className="material-symbols-outlined text-emerald-400 text-sm mt-0.5">check_circle</span>
                  <div>
                    <strong className="text-white">Zero-friction migrations</strong>
                    <p className="text-xs text-zinc-500">Synchronize database collection schema definitions automatically without SQL compile headaches.</p>
                  </div>
                </div>
              </div>
            </div>

            {/* Right Side: Animated CLI terminal playing script */}
            <div className="lg:col-span-3">
              <div className="bg-zinc-950 border border-white/5 rounded-2xl overflow-hidden shadow-2xl flex flex-col relative h-[360px]">
                {/* Window header */}
                <div className="flex items-center justify-between px-md py-sm bg-zinc-900/60 border-b border-white/5 select-none shrink-0">
                  <div className="flex space-x-xs">
                    <span className="w-3 h-3 rounded-full bg-red-500/50 inline-block" />
                    <span className="w-3 h-3 rounded-full bg-yellow-500/50 inline-block" />
                    <span className="w-3 h-3 rounded-full bg-green-500/50 inline-block" />
                  </div>
                  <span className="font-code-sm text-[10px] text-zinc-500">strata-cli-simulator</span>
                  <div className="w-12" />
                </div>
                
                {/* Simulated CLI stdout screen */}
                <div className="p-lg font-code-sm text-xs text-left overflow-y-auto flex-1 space-y-xs select-text text-zinc-300">
                  {terminalLines.map((line, idx) => {
                    if (!line) return null;
                    return (
                      <p key={idx} className={
                        line.startsWith('strata_cli $') 
                          ? 'text-white font-semibold pt-sm' 
                          : line.startsWith('✔') || line.startsWith('⚡')
                          ? 'text-emerald-400'
                          : 'text-zinc-400 pl-xs'
                      }>
                        {line}
                      </p>
                    );
                  })}
                  {/* Blinking cursor */}
                  <span className="inline-block w-1.5 h-3.5 bg-zinc-400 animate-pulse ml-xs mt-xs align-middle" />
                </div>
              </div>
            </div>

          </div>
        </section>

        {/* Count performance stats counters */}
        <Stats />

        {/* Developer Testimonials Grid */}
        <Testimonials />

        {/* ── Ready to build CTA Section ── */}
        <section className="py-24 max-w-5xl mx-auto px-md md:px-lg relative">
          <motion.div
            initial={{ opacity: 0, scale: 0.98 }}
            whileInView={{ opacity: 1, scale: 1 }}
            viewport={{ once: true }}
            transition={{ duration: 0.6 }}
            className="bg-zinc-950 border border-white/5 hover:border-[#f26500]/25 rounded-3xl p-xl md:p-3xl text-center space-y-lg relative overflow-hidden shadow-2xl"
          >
            {/* Background glowing blob */}
            <div className="absolute -bottom-20 left-1/2 -translate-x-1/2 w-[350px] h-[350px] rounded-full bg-[#f26500]/5 blur-[80px]" />
            
            <h2 className="font-headline-lg text-3xl md:text-5xl font-extrabold text-white leading-tight">
              Ready to accelerate your <br />
              <span className="bg-gradient-to-r from-[#f26500] to-[#ffa35a] bg-clip-text text-transparent">backend engineering?</span>
            </h2>
            
            <p className="font-body-md text-zinc-400 text-sm md:text-base max-w-xl mx-auto leading-relaxed w-full">
              Start building with PostgreSQL schema tables, instant REST APIs, JWT authentication, and S3 file dropzones today.
            </p>

            <div className="flex flex-col sm:flex-row items-center justify-center gap-md pt-sm">
              <Link
                to="/dashboard"
                className="w-full sm:w-auto bg-[#f26500] hover:bg-[#ff7d26] text-white py-sm px-xl rounded-lg font-semibold shadow-md shadow-[#f26500]/25 transition-colors text-center"
              >
                Start Building Now
              </Link>
              <a
                href="https://strata.dev/docs"
                className="w-full sm:w-auto border border-white/10 hover:border-white/20 bg-white/5 text-white py-sm px-xl rounded-lg font-semibold transition-colors text-center"
              >
                Read Documentation
              </a>
            </div>
          </motion.div>
        </section>

        {/* Global Product Footer Column Grid */}
        <Footer />
      </div>
    </div>
  );
};
