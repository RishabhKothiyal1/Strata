import React from 'react';
import { Link } from 'react-router-dom';

export const Footer: React.FC = () => {
  const footerLinks = [
    {
      title: 'Product',
      links: [
        { label: 'Database', href: '#database' },
        { label: 'Authentication', href: '#auth' },
        { label: 'Object Storage', href: '#storage' },
        { label: 'Edge Functions', href: '#functions' },
        { label: 'Realtime Channels', href: '#realtime' },
      ]
    },
    {
      title: 'Resources',
      links: [
        { label: 'Documentation', href: 'https://strata.dev/docs' },
        { label: 'System Status', href: '#' },
        { label: 'CORS Configuration', href: '#' },
        { label: 'Goja Playground', href: '#' },
      ]
    },
    {
      title: 'Developers',
      links: [
        { label: 'GitHub Repository', href: 'https://github.com/RishabhKothiyal1/Strata' },
        { label: 'API Reference', href: '#' },
        { label: 'SDK Downloads', href: '#' },
        { label: 'Open Source', href: '#' },
      ]
    },
    {
      title: 'Company',
      links: [
        { label: 'About Us', href: '#' },
        { label: 'Brand Guidelines', href: '#' },
        { label: 'Careers', href: '#' },
        { label: 'Security Policies', href: '#' },
      ]
    }
  ];

  return (
    <footer className="bg-zinc-950 border-t border-white/5 pt-20 pb-10 mt-24">
      <div className="max-w-7xl mx-auto px-md md:px-lg grid grid-cols-2 md:grid-cols-6 gap-xl mb-16">
        
        {/* Logo and Brand tagline */}
        <div className="col-span-2 space-y-md text-left">
          <Link to="/" className="flex items-center gap-xs">
            <div className="w-8 h-8 rounded-lg bg-[#f26500] flex items-center justify-center text-white shadow-sm shadow-[#f26500]/30">
              <span className="material-symbols-outlined text-xl">layers</span>
            </div>
            <span className="font-headline-md text-lg text-white font-bold tracking-tight">
              Strata
            </span>
          </Link>
          <p className="font-body-md text-zinc-500 text-xs md:text-sm leading-relaxed max-w-xs">
            Open-source Backend-as-a-Service powered by PostgreSQL. Deploy fast, scale infinitely, maintain control.
          </p>
          
          {/* Social Icons */}
          <div className="flex items-center space-x-md text-zinc-500 pt-sm">
            <a href="https://github.com/RishabhKothiyal1/Strata" target="_blank" rel="noreferrer" className="hover:text-white transition-colors" aria-label="GitHub">
              <span className="material-symbols-outlined text-md">code</span>
            </a>
            <a href="#" className="hover:text-white transition-colors" aria-label="Discord">
              <span className="material-symbols-outlined text-md">forum</span>
            </a>
            <a href="#" className="hover:text-white transition-colors" aria-label="X / Twitter">
              <span className="material-symbols-outlined text-md">alternate_email</span>
            </a>
            <a href="#" className="hover:text-white transition-colors" aria-label="YouTube">
              <span className="material-symbols-outlined text-md">smart_display</span>
            </a>
          </div>
        </div>

        {/* Dynamic Navigation Columns */}
        {footerLinks.map((col) => (
          <div key={col.title} className="text-left space-y-md">
            <h4 className="font-label-md text-xs text-white font-bold uppercase tracking-wider">
              {col.title}
            </h4>
            <ul className="space-y-xs text-xs" aria-label={`${col.title} navigation`}>
              {col.links.map((link) => (
                <li key={link.label}>
                  <a
                    href={link.href}
                    className="text-zinc-500 hover:text-white transition-colors"
                  >
                    {link.label}
                  </a>
                </li>
              ))}
            </ul>
          </div>
        ))}

      </div>

      {/* Copyright Notice */}
      <div className="max-w-7xl mx-auto px-md md:px-lg border-t border-white/5 pt-md flex flex-col md:flex-row items-center justify-between gap-md text-xs text-zinc-600">
        <span>© {new Date().getFullYear()} Strata Studio Inc. All rights reserved.</span>
        <div className="flex items-center space-x-md">
          <a href="#" className="hover:text-white transition-colors">Privacy Policy</a>
          <span>•</span>
          <a href="#" className="hover:text-white transition-colors">Terms of Service</a>
        </div>
      </div>
    </footer>
  );
};
