import React from 'react';
import { motion } from 'framer-motion';

export const BackgroundEffects: React.FC = () => {
  return (
    <div className="fixed inset-0 pointer-events-none z-0 overflow-hidden bg-[#030303]">
      {/* ── Grid Pattern Overlay ── */}
      <div 
        className="absolute inset-0 opacity-[0.05]"
        style={{
          backgroundImage: `
            linear-gradient(to right, rgba(255, 255, 255, 0.1) 1px, transparent 1px),
            linear-gradient(to bottom, rgba(255, 255, 255, 0.1) 1px, transparent 1px)
          `,
          backgroundSize: '24px 24px',
        }}
      />

      {/* ── Radiant Orange Blob ── */}
      <motion.div
        className="absolute -top-[10%] left-[20%] w-[500px] h-[500px] rounded-full bg-[#f26500]/10 blur-[120px]"
        animate={{
          x: [0, 40, -20, 0],
          y: [0, -30, 20, 0],
        }}
        transition={{
          duration: 20,
          repeat: Infinity,
          ease: 'easeInOut',
        }}
      />

      {/* ── Radiant Purple Blob ── */}
      <motion.div
        className="absolute top-[40%] -right-[10%] w-[600px] h-[600px] rounded-full bg-[#4648d4]/10 blur-[150px]"
        animate={{
          x: [0, -50, 30, 0],
          y: [0, 40, -30, 0],
        }}
        transition={{
          duration: 25,
          repeat: Infinity,
          ease: 'easeInOut',
        }}
      />

      {/* ── Soft Ambient Central Glow ── */}
      <div className="absolute inset-0 bg-radial from-transparent via-[#030303]/40 to-[#030303] z-[1]" />
    </div>
  );
};
