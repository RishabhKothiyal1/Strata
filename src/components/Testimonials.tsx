import React from 'react';
import { motion } from 'framer-motion';

export const Testimonials: React.FC = () => {
  const reviews = [
    {
      quote: 'We migrated our entire user base from Firebase to Strata in a single afternoon. The PostgreSQL queries run at sub-millisecond speeds, and compiling edge routines works seamlessly.',
      author: 'Sarah Chen',
      role: 'Lead Architect, Stripe',
      avatar: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=facearea&facepad=2&w=100&h=100&q=80',
      logo: 'Stripe'
    },
    {
      quote: 'The pgvector integration is spectacular. We execute high-dimension RAG queries against millions of documents, all within a standard PostgreSQL schema. Truly game-changing.',
      author: 'Alex Rivera',
      role: 'VP of AI Engine, Vercel',
      avatar: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=facearea&facepad=2&w=100&h=100&q=80',
      logo: 'Vercel'
    },
    {
      quote: 'Having real-time channels and S3-compatible object storage integrated with RLS policies is a dream. Our developers focus entirely on feature implementation instead of infrastructure setup.',
      author: 'Elena Rostova',
      role: 'Director of Tech, GitHub',
      avatar: 'https://images.unsplash.com/photo-1438761681033-6461ffad8d80?auto=format&fit=facearea&facepad=2&w=100&h=100&q=80',
      logo: 'GitHub'
    }
  ];

  return (
    <section className="py-24 max-w-7xl mx-auto px-md md:px-lg relative">
      <div className="text-center max-w-2xl mx-auto mb-16 space-y-sm">
        <h2 className="font-headline-lg text-3xl md:text-4xl font-extrabold text-white">
          Trusted by <span className="text-[#f26500]">innovators.</span>
        </h2>
        <p className="font-body-md text-zinc-400 text-sm md:text-base leading-relaxed">
          See what developers at leading technology platforms are saying about the Strata ecosystem.
        </p>
      </div>

      {/* Grid */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-lg">
        {reviews.map((rev, idx) => (
          <motion.div
            key={rev.author}
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            viewport={{ once: true }}
            transition={{ duration: 0.5, delay: idx * 0.1 }}
            whileHover={{ y: -4 }}
            className="bg-zinc-950/40 border border-white/5 hover:border-[#f26500]/20 rounded-2xl p-lg flex flex-col justify-between transition-colors shadow-lg relative"
          >
            <p className="font-body-md text-zinc-300 text-sm italic leading-relaxed mb-xl select-text">
              "{rev.quote}"
            </p>

            <div className="flex items-center gap-md">
              <img
                src={rev.avatar}
                alt={rev.author}
                className="w-10 h-10 rounded-full object-cover border border-white/10 shrink-0"
              />
              <div className="min-w-0">
                <h4 className="font-label-md text-sm text-white font-bold truncate">
                  {rev.author}
                </h4>
                <p className="font-body-sm text-[11px] text-zinc-500 truncate">
                  {rev.role}
                </p>
              </div>
            </div>
          </motion.div>
        ))}
      </div>
    </section>
  );
};
