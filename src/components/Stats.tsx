import React, { useEffect, useState, useRef } from 'react';
import { motion, useInView } from 'framer-motion';

export const Stats: React.FC = () => {
  const containerRef = useRef<HTMLDivElement>(null);
  const isInView = useInView(containerRef, { once: true, amount: 0.3 });
  
  const [requests, setRequests] = useState(0);
  const [projects, setProjects] = useState(0);
  const [uptime, setUptime] = useState(90.00);
  const [countries, setCountries] = useState(0);

  useEffect(() => {
    if (!isInView) return;

    // AnimateRequests: from 0 to 10
    let reqStart = 0;
    const reqTimer = setInterval(() => {
      reqStart += 1;
      if (reqStart >= 10) {
        setRequests(10);
        clearInterval(reqTimer);
      } else {
        setRequests(reqStart);
      }
    }, 100);

    // Animate Projects: from 0 to 500
    let projStart = 0;
    const projTimer = setInterval(() => {
      projStart += 25;
      if (projStart >= 500) {
        setProjects(500);
        clearInterval(projTimer);
      } else {
        setProjects(projStart);
      }
    }, 50);

    // Animate Uptime: from 90.00 to 99.99
    let uptimeStart = 90.00;
    const uptimeTimer = setInterval(() => {
      uptimeStart += 0.45;
      if (uptimeStart >= 99.99) {
        setUptime(99.99);
        clearInterval(uptimeTimer);
      } else {
        setUptime(parseFloat(uptimeStart.toFixed(2)));
      }
    }, 45);

    // Animate Countries: from 0 to 150
    let countryStart = 0;
    const countryTimer = setInterval(() => {
      countryStart += 10;
      if (countryStart >= 150) {
        setCountries(150);
        clearInterval(countryTimer);
      } else {
        setCountries(countryStart);
      }
    }, 60);

    return () => {
      clearInterval(reqTimer);
      clearInterval(projTimer);
      clearInterval(uptimeTimer);
      clearInterval(countryTimer);
    };
  }, [isInView]);

  const statItems = [
    { label: 'API Requests Served', value: `${requests}M+` },
    { label: 'Developer Projects', value: `${projects}K+` },
    { label: 'Platform Uptime SLA', value: `${uptime}%` },
    { label: 'Global Edge Locations', value: `${countries}+` },
  ];

  return (
    <section className="py-20 bg-zinc-950/20 border-t border-b border-white/5 relative select-none">
      <div ref={containerRef} className="max-w-7xl mx-auto px-md md:px-lg">
        <div className="grid grid-cols-2 lg:grid-cols-4 gap-xl">
          {statItems.map((item, idx) => (
            <motion.div
              key={item.label}
              initial={{ opacity: 0, y: 15 }}
              animate={isInView ? { opacity: 1, y: 0 } : {}}
              transition={{ duration: 0.5, delay: idx * 0.1 }}
              className="text-center space-y-xs"
            >
              <span className="font-headline-lg text-4xl md:text-5xl font-extrabold text-white bg-gradient-to-r from-white via-zinc-200 to-zinc-400 bg-clip-text text-transparent block">
                {item.value}
              </span>
              <span className="font-label-md text-[10px] md:text-xs text-zinc-500 uppercase tracking-widest font-bold block">
                {item.label}
              </span>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
};
