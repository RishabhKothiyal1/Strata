import React, { useState } from 'react';
import { useOutletContext } from 'react-router-dom';
import type { LayoutContextType } from '../components/DashboardLayout';

export const Support: React.FC = () => {
  const { showToast } = useOutletContext<LayoutContextType>();

  // Bug form state
  const [bugTitle, setBugTitle] = useState('');
  const [bugDesc, setBugDesc] = useState('');
  const [bugFile, setBugFile] = useState<File | null>(null);
  const [isSubmittingBug, setIsSubmittingBug] = useState(false);
  const [bugError, setBugError] = useState<string | null>(null);

  // Feature form state
  const [featTitle, setFeatTitle] = useState('');
  const [featDesc, setFeatDesc] = useState('');
  const [featCategory, setFeatCategory] = useState('database');
  const [isSubmittingFeat, setIsSubmittingFeat] = useState(false);
  const [featError, setFeatError] = useState<string | null>(null);

  // Card links
  const resourceCards = [
    {
      title: 'Documentation',
      desc: 'Browse user guides, authentication tutorials, and edge function schemas.',
      href: 'https://strata.dev/docs',
      icon: 'menu_book',
    },
    {
      title: 'API Reference',
      desc: 'Inspect REST API endpoints, Go wrappers, and payload options.',
      href: 'https://strata.dev/docs/api',
      icon: 'api',
    },
    {
      title: 'Tutorials',
      desc: 'Learn how to deploy serverless JS and set up S3 object storage folders.',
      href: 'https://strata.dev/docs/tutorials',
      icon: 'school',
    },
    {
      title: 'GitHub Repository',
      desc: 'Star our repository, review open source scripts, and inspect Go compilers.',
      href: 'https://github.com/strata-dev/strata',
      icon: 'code',
    },
    {
      title: 'Issues List',
      desc: 'Report core bugs or track upcoming features on the GitHub repository.',
      href: 'https://github.com/strata-dev/strata/issues',
      icon: 'bug_report',
    },
    {
      title: 'Discussions',
      desc: 'Interact with community developers, ask setup questions, and share plugins.',
      href: 'https://github.com/strata-dev/strata/discussions',
      icon: 'forum',
    },
  ];

  // Submit Bug Handler
  const handleBugSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setBugError(null);
    if (!bugTitle.trim() || !bugDesc.trim()) {
      setBugError('Please fill in all required fields.');
      return;
    }
    setIsSubmittingBug(true);
    setTimeout(() => {
      setIsSubmittingBug(false);
      showToast('Bug report submitted successfully.', 'success');
      setBugTitle('');
      setBugDesc('');
      setBugFile(null);
      // Reset file input
      const fileInput = document.getElementById('bug-screenshot') as HTMLInputElement;
      if (fileInput) fileInput.value = '';
    }, 1000);
  };

  // Submit Feature Request Handler
  const handleFeatSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    setFeatError(null);
    if (!featTitle.trim() || !featDesc.trim()) {
      setFeatError('Please fill in all required fields.');
      return;
    }
    setIsSubmittingFeat(true);
    setTimeout(() => {
      setIsSubmittingFeat(false);
      showToast('Feature request submitted successfully.', 'success');
      setFeatTitle('');
      setFeatDesc('');
      setFeatCategory('database');
    }, 1000);
  };

  return (
    <div className="space-y-xl animate-fadeIn w-full">
      {/* Page Header */}
      <div>
        <h2 className="font-headline-lg text-headline-lg font-bold text-[#191c1d] mb-xs">
          Support Hub
        </h2>
        <p className="font-body-md text-body-md text-[#464554]">
          Need help using Strata Studio? Reach out to us or explore the developer resources below.
        </p>
      </div>

      {/* Grid: Help Cards */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-md">
        {resourceCards.map((card, i) => (
          <a
            key={i}
            href={card.href}
            target="_blank"
            rel="noopener noreferrer"
            className="bg-white border border-[#c7c4d7]/30 hover:border-[#4648d4] p-lg rounded-xl shadow-sm hover:shadow-md transition-all duration-200 flex items-start gap-md group"
          >
            <div className="w-10 h-10 rounded-lg bg-[#e1e0ff] text-[#4648d4] flex items-center justify-center shrink-0 border border-[#c0c1ff]">
              <span className="material-symbols-outlined text-xl">{card.icon}</span>
            </div>
            <div className="space-y-xs min-w-0">
              <h4 className="font-label-md text-label-md font-bold text-[#191c1d] group-hover:text-[#4648d4] flex items-center gap-xs">
                <span>{card.title}</span>
                <span className="material-symbols-outlined text-xs opacity-0 group-hover:opacity-100 transition-opacity">open_in_new</span>
              </h4>
              <p className="font-body-sm text-xs text-[#464554] leading-relaxed">
                {card.desc}
              </p>
            </div>
          </a>
        ))}
      </div>

      {/* Grid: Forms */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-lg pt-md">
        
        {/* Form 1: Report a Bug */}
        <div className="bg-white rounded-xl shadow-sm p-lg border border-[#c7c4d7]/30 space-y-md">
          <div>
            <h3 className="font-headline-md text-headline-md font-bold text-[#191c1d] mb-xs flex items-center gap-xs">
              <span className="material-symbols-outlined text-[#ba1a1a]">bug_report</span>
              <span>Report a Bug</span>
            </h3>
            <p className="font-body-sm text-xs text-[#464554]">
              Found a bug in Strata Studio? Let us know and our engineers will investigate.
            </p>
          </div>

          <form onSubmit={handleBugSubmit} className="space-y-md">
            {bugError && (
              <div className="p-sm bg-[#ffdad6] text-[#93000a] border border-[#ba1a1a]/20 rounded-lg text-xs font-semibold">
                {bugError}
              </div>
            )}

            <div className="space-y-xs">
              <label className="block font-label-md text-xs text-[#191c1d] font-semibold" htmlFor="bug-title">
                Summary Title <span className="text-[#ba1a1a]">*</span>
              </label>
              <input
                className="w-full bg-white border border-[#c7c4d7] rounded-lg px-md py-sm text-[#191c1d] font-body-md text-sm placeholder-[#464554]/50 focus:outline-none focus:border-[#4648d4] focus:ring-4 focus:ring-[#4648d4]/10 transition-all shadow-sm"
                id="bug-title"
                type="text"
                placeholder="e.g. S3 file upload fails on large archives"
                value={bugTitle}
                onChange={(e) => setBugTitle(e.target.value)}
                required
              />
            </div>

            <div className="space-y-xs">
              <label className="block font-label-md text-xs text-[#191c1d] font-semibold" htmlFor="bug-desc">
                Description & Steps <span className="text-[#ba1a1a]">*</span>
              </label>
              <textarea
                className="w-full bg-white border border-[#c7c4d7] rounded-lg px-md py-sm text-[#191c1d] font-body-md text-sm placeholder-[#464554]/50 focus:outline-none focus:border-[#4648d4] focus:ring-4 focus:ring-[#4648d4]/10 transition-all shadow-sm resize-y"
                id="bug-desc"
                rows={4}
                placeholder="Describe what happened, error codes, and instructions to reproduce..."
                value={bugDesc}
                onChange={(e) => setBugDesc(e.target.value)}
                required
              />
            </div>

            <div className="space-y-xs">
              <label className="block font-label-md text-xs text-[#191c1d] font-semibold" htmlFor="bug-screenshot">
                Screenshot / Log Upload
              </label>
              <input
                className="w-full text-xs text-[#464554] file:mr-md file:py-xs file:px-md file:rounded-lg file:border-0 file:text-xs file:font-semibold file:bg-[#e1e0ff] file:text-[#07006c] file:hover:bg-[#c0c1ff] file:cursor-pointer"
                id="bug-screenshot"
                type="file"
                accept="image/*,.txt,.log"
                onChange={(e) => setBugFile(e.target.files?.[0] || null)}
              />
              {bugFile && (
                <p className="text-[11px] text-[#0a5c0a] font-semibold mt-xs flex items-center gap-xs">
                  <span className="material-symbols-outlined text-[14px]">check_circle</span>
                  <span>Selected: {bugFile.name} ({Math.round(bugFile.size / 1024)} KB)</span>
                </p>
              )}
            </div>

            <button
              type="submit"
              disabled={isSubmittingBug}
              className="bg-[#4648d4] hover:bg-[#6063ee] disabled:opacity-50 disabled:cursor-not-allowed text-white py-xs px-lg rounded-lg font-label-md text-xs font-semibold shadow-sm transition-colors cursor-pointer focus:outline-none focus:ring-2 focus:ring-[#4648d4]/50"
            >
              {isSubmittingBug ? 'Submitting...' : 'Submit Bug Report'}
            </button>
          </form>
        </div>

        {/* Form 2: Feature Request */}
        <div className="bg-white rounded-xl shadow-sm p-lg border border-[#c7c4d7]/30 space-y-md">
          <div>
            <h3 className="font-headline-md text-headline-md font-bold text-[#191c1d] mb-xs flex items-center gap-xs">
              <span className="material-symbols-outlined text-[#4648d4]">auto_awesome</span>
              <span>Request a Feature</span>
            </h3>
            <p className="font-body-sm text-xs text-[#464554]">
              Want a new feature or core engine optimization? Tell us what you want to see built.
            </p>
          </div>

          <form onSubmit={handleFeatSubmit} className="space-y-md">
            {featError && (
              <div className="p-sm bg-[#ffdad6] text-[#93000a] border border-[#ba1a1a]/20 rounded-lg text-xs font-semibold">
                {featError}
              </div>
            )}

            <div className="space-y-xs">
              <label className="block font-label-md text-xs text-[#191c1d] font-semibold" htmlFor="feat-title">
                Feature Title <span className="text-[#ba1a1a]">*</span>
              </label>
              <input
                className="w-full bg-white border border-[#c7c4d7] rounded-lg px-md py-sm text-[#191c1d] font-body-md text-sm placeholder-[#464554]/50 focus:outline-none focus:border-[#4648d4] focus:ring-4 focus:ring-[#4648d4]/10 transition-all shadow-sm"
                id="feat-title"
                type="text"
                placeholder="e.g. Add TypeScript compilation for serverless edge functions"
                value={featTitle}
                onChange={(e) => setFeatTitle(e.target.value)}
                required
              />
            </div>

            <div className="space-y-xs">
              <label className="block font-label-md text-xs text-[#191c1d] font-semibold" htmlFor="feat-category">
                Category Area
              </label>
              <select
                id="feat-category"
                value={featCategory}
                onChange={(e) => setFeatCategory(e.target.value)}
                className="w-full bg-white border border-[#c7c4d7] rounded-lg px-md py-sm text-[#191c1d] font-body-md text-sm focus:outline-none focus:border-[#4648d4] focus:ring-4 focus:ring-[#4648d4]/10 transition-all shadow-sm"
              >
                <option value="database">Database Collection CRUD</option>
                <option value="storage">Object S3 Storage</option>
                <option value="functions">Serverless Edge Runtime</option>
                <option value="security">Security & API Keys</option>
                <option value="other">General UI / Other</option>
              </select>
            </div>

            <div className="space-y-xs">
              <label className="block font-label-md text-xs text-[#191c1d] font-semibold" htmlFor="feat-desc">
                Feature Description <span className="text-[#ba1a1a]">*</span>
              </label>
              <textarea
                className="w-full bg-white border border-[#c7c4d7] rounded-lg px-md py-sm text-[#191c1d] font-body-md text-sm placeholder-[#464554]/50 focus:outline-none focus:border-[#4648d4] focus:ring-4 focus:ring-[#4648d4]/10 transition-all shadow-sm resize-y"
                id="feat-desc"
                rows={3}
                placeholder="Describe your feature request, its target scope, and how it will benefit developers..."
                value={featDesc}
                onChange={(e) => setFeatDesc(e.target.value)}
                required
              />
            </div>

            <button
              type="submit"
              disabled={isSubmittingFeat}
              className="bg-[#4648d4] hover:bg-[#6063ee] disabled:opacity-50 disabled:cursor-not-allowed text-white py-xs px-lg rounded-lg font-label-md text-xs font-semibold shadow-sm transition-colors cursor-pointer focus:outline-none focus:ring-2 focus:ring-[#4648d4]/50"
            >
              {isSubmittingFeat ? 'Submitting...' : 'Submit Feature Request'}
            </button>
          </form>
        </div>
      </div>
    </div>
  );
};
