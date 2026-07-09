import React from 'react';
import { motion } from 'framer-motion';

export const Templates: React.FC = () => {
  const templates = [
    {
      title: 'SaaS Subscription Starter',
      description: 'A complete SaaS boilerplate with Stripe billing, PostgreSQL tables schema, and pre-built Auth flows.',
      tags: ['Next.js', 'Stripe', 'Postgres'],
      icon: 'point_of_sale'
    },
    {
      title: 'AI Chat Playground',
      description: 'Interact with OpenAI & Anthropic Claude models using pgvector embeddings and semantic search cataloging.',
      tags: ['React', 'Vector DB', 'LLM'],
      icon: 'auto_awesome'
    },
    {
      title: 'Stripe Payment Gateway',
      description: 'Checkout templates configured with webhooks, product sync routines, and subscription lifecycle triggers.',
      tags: ['Node.js', 'Stripe', 'Webhooks'],
      icon: 'credit_card'
    },
    {
      title: 'Next.js 15 Boilerplate',
      description: 'Clean starting setup utilizing Next.js app router, Tailwind styling, and Strata SDK client components.',
      tags: ['Next.js', 'Tailwind', 'SDK'],
      icon: 'layers'
    },
    {
      title: 'React Client SPA',
      description: 'Single Page Application template integrated with React Router, TanStack query caching, and Auth providers.',
      tags: ['React', 'Vite', 'Router'],
      icon: 'desktop_windows'
    },
    {
      title: 'Blog Portfolio Engine',
      description: 'Configure Markdown ingestion, static image caching via S3, and real-time review commentary forums.',
      tags: ['Markdown', 'Storage', 'CDN'],
      icon: 'article'
    }
  ];

  return (
    <section className="py-24 max-w-7xl mx-auto px-md md:px-lg relative">
      <div className="text-center max-w-2xl mx-auto mb-16 space-y-sm">
        <h2 className="font-headline-lg text-3xl md:text-4xl font-extrabold text-white">
          Jumpstart your next <span className="text-[#f26500]">creation.</span>
        </h2>
        <p className="font-body-md text-zinc-400 text-sm md:text-base leading-relaxed">
          Clone fully configured boilerplate templates directly into your Strata workspace and deploy to production in minutes.
        </p>
      </div>

      {/* Templates Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-lg">
        {templates.map((tpl, i) => (
          <motion.div
            key={tpl.title}
            initial={{ opacity: 0, scale: 0.96 }}
            whileInView={{ opacity: 1, scale: 1 }}
            viewport={{ once: true, margin: '-50px' }}
            transition={{ duration: 0.4, delay: i * 0.05 }}
            whileHover={{ scale: 1.02 }}
            className="bg-zinc-950/40 border border-white/5 hover:border-[#f26500]/20 rounded-xl p-lg flex flex-col justify-between transition-colors shadow-lg cursor-pointer"
          >
            <div>
              <div className="w-9 h-9 rounded-lg bg-zinc-900 border border-white/5 text-[#f26500] flex items-center justify-center mb-md">
                <span className="material-symbols-outlined text-lg">{tpl.icon}</span>
              </div>
              
              <h3 className="font-headline-md text-sm text-white font-bold mb-xs">
                {tpl.title}
              </h3>
              
              <p className="font-body-md text-zinc-400 text-xs leading-relaxed mb-md">
                {tpl.description}
              </p>
            </div>

            <div className="flex flex-wrap gap-xs pt-xs">
              {tpl.tags.map((tag) => (
                <span
                  key={tag}
                  className="bg-white/5 border border-white/5 text-zinc-400 text-[10px] font-bold px-sm py-[2px] rounded-full"
                >
                  {tag}
                </span>
              ))}
            </div>
          </motion.div>
        ))}
      </div>
    </section>
  );
};
