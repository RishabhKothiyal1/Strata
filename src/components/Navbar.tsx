import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { motion, AnimatePresence } from 'framer-motion';

export const Navbar: React.FC = () => {
  const [isScrolled, setIsScrolled] = useState(false);
  const [isMobileMenuOpen, setIsMobileMenuOpen] = useState(false);

  useEffect(() => {
    const handleScroll = () => {
      setIsScrolled(window.scrollY > 20);
    };
    window.addEventListener('scroll', handleScroll);
    return () => window.removeEventListener('scroll', handleScroll);
  }, []);

  const navLinks = [
    { label: 'Product', href: '#features' },
    { label: 'Database', href: '#database' },
    { label: 'Auth', href: '#auth' },
    { label: 'Storage', href: '#storage' },
    { label: 'Edge Functions', href: '#functions' },
    { label: 'Realtime', href: '#realtime' },
    { label: 'Docs', href: 'https://strata.dev/docs' },
    { label: 'Pricing', href: '#pricing' },
  ];

  return (
    <>
      <header
        className={`fixed top-0 inset-x-0 z-50 transition-all duration-300 ${
          isScrolled 
            ? 'bg-zinc-950/80 backdrop-blur-md border-b border-white/5 py-sm' 
            : 'bg-transparent py-md'
        }`}
      >
        <div className="max-w-7xl mx-auto px-md md:px-lg flex items-center justify-between">
          {/* Logo */}
          <Link to="/" className="flex items-center gap-xs focus:outline-none">
            <div className="w-8 h-8 rounded-lg bg-[#f26500] flex items-center justify-center text-white shadow-sm shadow-[#f26500]/30 transition-transform hover:scale-105">
              <span className="material-symbols-outlined text-xl">layers</span>
            </div>
            <span className="font-headline-md text-lg text-white font-bold tracking-tight">
              Strata
            </span>
          </Link>

          {/* Desktop Navigation Links */}
          <nav className="hidden lg:flex items-center gap-md" aria-label="Desktop Nav">
            {navLinks.map((link) => (
              <a
                key={link.label}
                href={link.href}
                className="text-zinc-400 hover:text-white font-label-md text-xs transition-colors py-xs px-sm rounded-md hover:bg-white/5"
              >
                {link.label}
              </a>
            ))}
          </nav>

          {/* Desktop CTA actions */}
          <div className="hidden lg:flex items-center gap-md">
            {/* Github Link */}
            <a
              href="https://github.com/RishabhKothiyal1/Strata"
              target="_blank"
              rel="noreferrer"
              className="text-zinc-400 hover:text-white transition-colors"
              aria-label="GitHub Repository"
            >
              <svg className="w-5 h-5 fill-current" viewBox="0 0 24 24">
                <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
              </svg>
            </a>

            <Link
              to="/login"
              className="text-zinc-300 hover:text-white font-label-md text-xs transition-colors"
            >
              Sign In
            </Link>

            <Link
              to="/dashboard"
              className="bg-[#f26500] hover:bg-[#ff7d26] text-white py-[8px] px-md rounded-lg font-label-md text-xs font-semibold shadow-md shadow-[#f26500]/25 transition-all focus:outline-none focus:ring-2 focus:ring-[#f26500]/40"
            >
              Start Building
            </Link>
          </div>

          {/* Mobile hamburger menu toggle */}
          <button
            onClick={() => setIsMobileMenuOpen(!isMobileMenuOpen)}
            className="lg:hidden text-zinc-400 hover:text-white focus:outline-none p-xs"
            aria-label="Toggle Navigation Menu"
          >
            <span className="material-symbols-outlined text-2xl">
              {isMobileMenuOpen ? 'close' : 'menu'}
            </span>
          </button>
        </div>
      </header>

      {/* ── Mobile menu overlay list ── */}
      <AnimatePresence>
        {isMobileMenuOpen && (
          <motion.div
            initial={{ opacity: 0, y: -20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            transition={{ duration: 0.2 }}
            className="fixed inset-x-0 top-16 bg-zinc-950 border-b border-white/5 z-40 p-md flex flex-col gap-md lg:hidden"
          >
            <nav className="flex flex-col gap-xs" aria-label="Mobile Nav">
              {navLinks.map((link) => (
                <a
                  key={link.label}
                  href={link.href}
                  onClick={() => setIsMobileMenuOpen(false)}
                  className="text-zinc-400 hover:text-white hover:bg-white/5 font-label-md text-sm transition-colors py-sm px-md rounded-md"
                >
                  {link.label}
                </a>
              ))}
            </nav>

            <div className="h-px bg-white/5 w-full my-xs" />

            <div className="flex flex-col gap-md px-md">
              <a
                href="https://github.com/RishabhKothiyal1/Strata"
                target="_blank"
                rel="noreferrer"
                className="text-zinc-400 hover:text-white flex items-center gap-xs font-label-md text-sm transition-colors"
              >
                GitHub Repository
              </a>
              <Link
                to="/login"
                onClick={() => setIsMobileMenuOpen(false)}
                className="text-zinc-300 hover:text-white font-label-md text-sm transition-colors"
              >
                Sign In
              </Link>

              <Link
                to="/dashboard"
                onClick={() => setIsMobileMenuOpen(false)}
                className="bg-[#f26500] hover:bg-[#ff7d26] text-white py-sm text-center rounded-lg font-label-md text-sm font-semibold shadow-md shadow-[#f26500]/25 transition-all"
              >
                Start Building
              </Link>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </>
  );
};
