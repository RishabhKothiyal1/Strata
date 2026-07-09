import React, { useEffect, useRef, useCallback } from 'react';
import { createPortal } from 'react-dom';
import { motion, AnimatePresence } from 'framer-motion';

export interface DialogProps {
  open: boolean;
  onClose: () => void;
  title: string;
  description?: string;
  icon?: string;
  size?: 'small' | 'medium' | 'large' | 'fullscreen';
  children: React.ReactNode;
  footer?: React.ReactNode;
  maxHeight?: string; // e.g. "85vh", "80vh"
}

export const Dialog: React.FC<DialogProps> = ({
  open,
  onClose,
  title,
  description,
  icon,
  size = 'medium',
  children,
  footer,
  maxHeight = '85vh',
}) => {
  const containerRef = useRef<HTMLDivElement>(null);

  // ESC key handler
  const handleKeyDown = useCallback(
    (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    },
    [onClose]
  );

  // Focus trap + Keyboard listeners + Scroll lock
  useEffect(() => {
    if (!open) return;

    // Store original body style to restore
    const originalOverflow = document.body.style.overflow;
    document.body.style.overflow = 'hidden';

    document.addEventListener('keydown', handleKeyDown);

    // Focus first focusable element inside the dialog
    const focusableElements = containerRef.current?.querySelectorAll<HTMLElement>(
      'a[href], area[href], input:not([disabled]), select:not([disabled]), textarea:not([disabled]), button:not([disabled]), iframe, object, embed, [tabindex="0"], [contenteditable]'
    );
    
    const firstFocusable = focusableElements ? Array.from(focusableElements).find(el => el.tabIndex !== -1) : null;
    
    const timer = setTimeout(() => {
      if (firstFocusable) {
        firstFocusable.focus();
      } else {
        containerRef.current?.focus();
      }
    }, 50);

    // Tab key listener for trapping focus
    const handleTabKey = (e: KeyboardEvent) => {
      if (e.key !== 'Tab' || !focusableElements) return;
      const elements = Array.from(focusableElements).filter(el => el.tabIndex !== -1);
      if (elements.length === 0) {
        e.preventDefault();
        return;
      }
      const first = elements[0];
      const last = elements[elements.length - 1];

      if (e.shiftKey) {
        if (document.activeElement === first) {
          last.focus();
          e.preventDefault();
        }
      } else {
        if (document.activeElement === last) {
          first.focus();
          e.preventDefault();
        }
      }
    };

    document.addEventListener('keydown', handleTabKey);

    return () => {
      document.body.style.overflow = originalOverflow;
      document.removeEventListener('keydown', handleKeyDown);
      document.removeEventListener('keydown', handleTabKey);
      clearTimeout(timer);
    };
  }, [open, handleKeyDown]);

  // Determine size classes
  let sizeClass = 'w-full max-w-[720px]';
  if (size === 'small') sizeClass = 'w-full max-w-[500px]';
  if (size === 'large') sizeClass = 'w-full max-w-[960px]';
  if (size === 'fullscreen') sizeClass = 'w-screen h-screen max-w-none max-h-none rounded-none';

  return createPortal(
    <AnimatePresence>
      {open && (
        <div
          className="fixed inset-0 z-[60] flex items-center justify-center p-md md:p-lg"
          role="dialog"
          aria-modal="true"
          aria-labelledby="dialog-title"
        >
          {/* Backdrop Blur Overlay */}
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: 0.18, ease: 'easeInOut' }}
            onClick={onClose}
            className="fixed inset-0 bg-[#191c1d]/30 backdrop-blur-sm cursor-pointer"
          />

          {/* Dialog Container */}
          <motion.div
            ref={containerRef}
            tabIndex={-1}
            initial={{ opacity: 0, scale: 0.95 }}
            animate={{ opacity: 1, scale: 1 }}
            exit={{ opacity: 0, scale: 0.95 }}
            transition={{ duration: 0.18, ease: 'easeInOut' }}
            style={size !== 'fullscreen' ? { maxHeight } : undefined}
            className={`
              relative z-10 bg-white rounded-xl shadow-2xl flex flex-col border border-[#c7c4d7]/30 outline-none overflow-hidden
              ${sizeClass}
            `}
          >
            {/* Sticky Header */}
            <div className="flex items-start justify-between px-lg py-md border-b border-[#c7c4d7]/30 bg-white shrink-0">
              <div className="flex items-start gap-sm min-w-0 pr-md">
                {icon && (
                  <span className="material-symbols-outlined text-[#4648d4] text-[24px] mt-0.5 shrink-0">
                    {icon}
                  </span>
                )}
                <div className="min-w-0">
                  <h3 id="dialog-title" className="font-headline-md text-headline-md font-bold text-[#191c1d] truncate">
                    {title}
                  </h3>
                  {description && (
                    <p className="font-body-sm text-body-sm text-[#464554] mt-xs line-clamp-2">
                      {description}
                    </p>
                  )}
                </div>
              </div>
              <button
                onClick={onClose}
                className="text-[#464554] hover:text-[#191c1d] hover:bg-[#f3f4f5] rounded-full p-xs transition-colors shrink-0 cursor-pointer"
                aria-label="Close dialog"
              >
                <span className="material-symbols-outlined text-[20px]">close</span>
              </button>
            </div>

            {/* Scrollable Body Container */}
            <div className="flex-1 overflow-y-auto p-lg">
              {children}
            </div>

            {/* Sticky Footer */}
            {footer && (
              <div className="px-lg py-md border-t border-[#c7c4d7]/30 bg-[#f3f4f5] flex justify-end gap-md shrink-0 rounded-b-xl">
                {footer}
              </div>
            )}
          </motion.div>
        </div>
      )}
    </AnimatePresence>,
    document.body
  );
};
