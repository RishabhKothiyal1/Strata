import React from 'react';
import { motion } from 'framer-motion';

export const Comparison: React.FC = () => {
  const rows = [
    { feature: 'Database backend', strata: 'PostgreSQL', supabase: 'PostgreSQL', firebase: 'NoSQL Firestore' },
    { feature: 'Open Source code', strata: 'Yes (MIT)', supabase: 'Yes (Apache 2.0)', firebase: 'No (Proprietary)' },
    { feature: 'Self-Hosting capable', strata: 'Yes', supabase: 'Yes (Complex)', firebase: 'No' },
    { feature: 'User Authentication', strata: 'Yes', supabase: 'Yes', firebase: 'Yes' },
    { feature: 'S3 Object Storage', strata: 'Yes', supabase: 'Yes', firebase: 'Yes' },
    { feature: 'WebSockets Realtime', strata: 'Yes', supabase: 'Yes', firebase: 'Yes' },
    { feature: 'Serverless Functions', strata: 'Yes (Goja JS)', supabase: 'Yes (Deno)', firebase: 'Yes (Node.js)' },
    { feature: 'Vector Search engine', strata: 'Yes (pgvector)', supabase: 'Yes (pgvector)', firebase: 'No' },
  ];

  return (
    <section className="py-24 max-w-7xl mx-auto px-md md:px-lg relative">
      <div className="text-center max-w-2xl mx-auto mb-16 space-y-sm">
        <h2 className="font-headline-lg text-3xl md:text-4xl font-extrabold text-white">
          Why choose <span className="text-[#f26500]">Strata?</span>
        </h2>
        <p className="font-body-md text-zinc-400 text-sm md:text-base leading-relaxed">
          See how Strata stacks up against other industry-leading Backend-as-a-Service platforms.
        </p>
      </div>

      {/* Comparison Table */}
      <motion.div
        initial={{ opacity: 0, y: 25 }}
        whileInView={{ opacity: 1, y: 0 }}
        viewport={{ once: true }}
        transition={{ duration: 0.6 }}
        className="w-full bg-zinc-950/40 border border-white/5 rounded-2xl overflow-hidden shadow-2xl"
      >
        <div className="overflow-x-auto w-full">
          <table className="w-full border-collapse text-left text-xs md:text-sm">
            <thead>
              <tr className="bg-zinc-900/40 border-b border-white/5 text-zinc-400 uppercase tracking-wider font-semibold">
                <th className="px-lg py-md">Feature</th>
                <th className="px-lg py-md text-[#f26500] font-bold bg-[#f26500]/5">Strata</th>
                <th className="px-lg py-md">Supabase</th>
                <th className="px-lg py-md">Firebase</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-white/5 text-zinc-300">
              {rows.map((row) => (
                <tr key={row.feature} className="hover:bg-white/5 transition-colors">
                  <td className="px-lg py-md font-semibold text-white">{row.feature}</td>
                  <td className="px-lg py-md bg-[#f26500]/5 font-semibold text-white border-l border-r border-[#f26500]/10">
                    <span className="flex items-center gap-xs">
                      <span className="material-symbols-outlined text-[#f26500] text-sm">check_circle</span>
                      <span>{row.strata}</span>
                    </span>
                  </td>
                  <td className="px-lg py-md">
                    <span className="flex items-center gap-xs">
                      <span className="material-symbols-outlined text-zinc-500 text-sm">check_circle</span>
                      <span>{row.supabase}</span>
                    </span>
                  </td>
                  <td className="px-lg py-md">
                    <span className="flex items-center gap-xs">
                      <span className={`material-symbols-outlined text-sm ${row.firebase === 'No' ? 'text-zinc-700' : 'text-zinc-500'}`}>
                        {row.firebase === 'No' ? 'cancel' : 'check_circle'}
                      </span>
                      <span className={row.firebase === 'No' ? 'text-zinc-600' : 'text-zinc-300'}>{row.firebase}</span>
                    </span>
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </motion.div>
    </section>
  );
};
