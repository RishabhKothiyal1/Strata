import React from 'react';
import { motion } from 'framer-motion';

export const FeatureCard: React.FC = () => {
  const features = [
    {
      title: 'PostgreSQL compatible',
      description: '100% compliant PostgreSQL database. Fully accessible with tables, foreign keys, views, triggers, and full SQL compiler access.',
      icon: 'database',
      illustration: (
        <svg className="w-full h-full text-[#f26500]/25 group-hover:text-[#f26500]/50 transition-colors" viewBox="0 0 100 100" fill="none" stroke="currentColor" strokeWidth="2">
          <rect x="25" y="15" width="50" height="20" rx="3" />
          <rect x="25" y="40" width="50" height="20" rx="3" />
          <rect x="25" y="65" width="50" height="20" rx="3" />
          <circle cx="35" cy="25" r="2" fill="currentColor" />
          <circle cx="35" cy="50" r="2" fill="currentColor" />
          <circle cx="35" cy="75" r="2" fill="currentColor" />
          <line x1="45" y1="25" x2="65" y2="25" />
          <line x1="45" y1="50" x2="65" y2="50" />
          <line x1="45" y1="75" x2="65" y2="75" />
        </svg>
      )
    },
    {
      title: 'Authentication',
      description: 'Robust authentication layer supporting email logins, social OAuth (Google, GitHub), magic link gateway access, and JWT token issuance.',
      icon: 'lock',
      illustration: (
        <svg className="w-full h-full text-[#f26500]/25 group-hover:text-[#f26500]/50 transition-colors" viewBox="0 0 100 100" fill="none" stroke="currentColor" strokeWidth="2">
          <rect x="30" y="40" width="40" height="35" rx="4" />
          <path d="M40 40V30C40 23 45 20 50 20C55 20 60 23 60 30V40" />
          <circle cx="50" cy="55" r="4" fill="currentColor" />
          <line x1="50" y1="59" x2="50" y2="67" />
        </svg>
      )
    },
    {
      title: 'Object Storage',
      description: 'Scale file assets upload. Upload profiles, static images, archives, or large attachments directly to bucket catalogs backed by MinIO S3.',
      icon: 'inventory_2',
      illustration: (
        <svg className="w-full h-full text-[#f26500]/25 group-hover:text-[#f26500]/50 transition-colors" viewBox="0 0 100 100" fill="none" stroke="currentColor" strokeWidth="2">
          <rect x="25" y="30" width="50" height="45" rx="4" />
          <line x1="25" y1="42" x2="75" y2="42" />
          <rect x="42" y="34" width="16" height="4" rx="1" fill="currentColor" />
          <circle cx="50" cy="58" r="5" />
          <line x1="50" y1="53" x2="50" y2="63" />
          <line x1="45" y1="58" x2="55" y2="58" />
        </svg>
      )
    },
    {
      title: 'Edge Functions',
      description: 'Deploy serverless script environments. Write clean ES5.1 JavaScript routines compiled on Goja virtual sandboxes inside secure runtimes.',
      icon: 'code',
      illustration: (
        <svg className="w-full h-full text-[#f26500]/25 group-hover:text-[#f26500]/50 transition-colors" viewBox="0 0 100 100" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M35 30L20 50L35 70" strokeLinecap="round" strokeLinejoin="round" />
          <path d="M65 30L80 50L65 70" strokeLinecap="round" strokeLinejoin="round" />
          <line x1="55" y1="25" x2="45" y2="75" strokeLinecap="round" />
        </svg>
      )
    },
    {
      title: 'Realtime Channels',
      description: 'WebSocket listener network. Listen to database changes, broadcast events to clients, and capture connected user presence state instantly.',
      icon: 'bolt',
      illustration: (
        <svg className="w-full h-full text-[#f26500]/25 group-hover:text-[#f26500]/50 transition-colors" viewBox="0 0 100 100" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M55 15L30 50H50L45 85L70 50H50L55 15Z" strokeLinejoin="round" />
        </svg>
      )
    },
    {
      title: 'Vector Database',
      description: 'Integrate Vector engines. Query pgvector AI embeddings, conduct semantic search logic, and store contexts for OpenAI, Gemini, or Claude models.',
      icon: 'auto_awesome',
      illustration: (
        <svg className="w-full h-full text-[#f26500]/25 group-hover:text-[#f26500]/50 transition-colors" viewBox="0 0 100 100" fill="none" stroke="currentColor" strokeWidth="2">
          <path d="M50 15L55 35L75 40L55 45L50 65L45 45L25 40L45 35L50 15Z" fill="currentColor" fillOpacity="0.1" />
          <path d="M75 60L78 70L88 73L78 76L75 86L72 76L62 73L72 70L75 60Z" fill="currentColor" fillOpacity="0.1" />
        </svg>
      )
    }
  ];

  return (
    <section id="features" className="py-24 max-w-7xl mx-auto px-md md:px-lg relative">
      
      {/* Grid Title */}
      <div className="text-center max-w-2xl mx-auto mb-20 space-y-sm">
        <h2 className="font-headline-lg text-3xl md:text-4xl font-extrabold text-white">
          All the backend tools, <span className="text-[#f26500]">fully integrated.</span>
        </h2>
        <p className="font-body-md text-zinc-400 text-sm md:text-base leading-relaxed">
          Strata is designed for production speed, offering robust abstractions over enterprise developer stacks out of the box.
        </p>
      </div>

      {/* Grid container */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-lg">
        {features.map((feature, i) => (
          <motion.div
            key={feature.title}
            initial={{ opacity: 0, y: 30 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true, margin: '-50px' }}
            transition={{ duration: 0.5, delay: i * 0.1 }}
            whileHover={{ y: -6 }}
            className="group relative bg-zinc-950/40 hover:bg-zinc-900/30 border border-white/5 hover:border-[#f26500]/25 rounded-2xl p-lg flex flex-col justify-between overflow-hidden transition-all duration-350 cursor-pointer shadow-md"
          >
            {/* Glowing Corner Effect */}
            <div className="absolute top-0 right-0 w-24 h-24 bg-gradient-to-br from-[#f26500]/5 to-transparent blur-md pointer-events-none group-hover:scale-125 transition-transform" />

            <div>
              {/* Header Icon */}
              <div className="w-10 h-10 rounded-lg bg-zinc-900 border border-white/5 text-[#f26500] flex items-center justify-center mb-lg shadow-sm">
                <span className="material-symbols-outlined text-xl">{feature.icon}</span>
              </div>

              {/* Title */}
              <h3 className="font-headline-md text-lg text-white font-bold mb-sm group-hover:text-[#f26500] transition-colors">
                {feature.title}
              </h3>

              {/* Description */}
              <p className="font-body-md text-zinc-400 text-xs md:text-sm leading-relaxed mb-lg">
                {feature.description}
              </p>
            </div>

            {/* Micro Illustration Canvas */}
            <div className="w-full h-32 bg-zinc-950/60 border border-white/5 rounded-xl overflow-hidden p-md relative flex items-center justify-center">
              {feature.illustration}
            </div>
          </motion.div>
        ))}
      </div>
    </section>
  );
};
