import React, { useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

export const Signup: React.FC = () => {
  const { signUp } = useAuth();
  const navigate = useNavigate();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (password !== confirmPassword) {
      setError("Passwords do not match");
      return;
    }

    if (password.length < 6) {
      setError("Password must be at least 6 characters long");
      return;
    }

    setIsSubmitting(true);
    try {
      await signUp(email, password);
      setSuccess(true);
      setTimeout(() => {
        navigate('/login');
      }, 2000);
    } catch (err: any) {
      setError(err.message || 'Registration failed. Please try again.');
    } finally {
      setIsSubmitting(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center p-lg bg-[#f8f9fa] text-[#191c1d] font-body-md text-body-md antialiased">
      <main className="w-full max-w-[420px]">
        {/* Brand Header */}
        <div className="flex flex-col items-center mb-lg">
          <div className="w-12 h-12 bg-[#4648d4] rounded-xl flex items-center justify-center shadow-sm mb-sm text-white">
            <span className="material-symbols-outlined text-2xl">layers</span>
          </div>
          <h1 className="font-headline-lg text-headline-lg text-[#191c1d] font-bold text-center">Strata Studio</h1>
          <p className="font-body-md text-body-md text-[#464554] text-center mt-xs">Create your developer account</p>
        </div>

        {/* Card */}
        <div className="bg-white rounded-xl shadow-sm p-lg relative overflow-hidden border border-[#e1e3e4]">
          <div className="absolute top-0 inset-x-0 h-1 bg-gradient-to-r from-transparent via-[#4648d4]/20 to-transparent"></div>

          {error && (
            <div className="mb-md p-sm bg-[#ffdad6] text-[#93000a] border border-[#ba1a1a]/20 rounded-lg flex items-start gap-xs text-sm">
              <span className="material-symbols-outlined text-md mt-0.5">error</span>
              <span>{error}</span>
            </div>
          )}

          {success && (
            <div className="mb-md p-sm bg-[#e1e0ff]/30 text-[#07006c] border border-[#c0c1ff]/60 rounded-lg flex items-start gap-xs text-sm">
              <span className="material-symbols-outlined text-md mt-0.5">check_circle</span>
              <span>Account created! Redirecting to login...</span>
            </div>
          )}

          <form onSubmit={handleSubmit} className="space-y-md">
            {/* Email */}
            <div>
              <label className="block font-label-md text-label-md text-[#191c1d] mb-xs" htmlFor="email">
                Email address
              </label>
              <div className="relative rounded-md input-glow transition-shadow duration-200">
                <div className="absolute inset-y-0 left-0 pl-sm flex items-center pointer-events-none">
                  <span className="material-symbols-outlined text-[#464554] text-lg">mail</span>
                </div>
                <input
                  className="block w-full pl-xl pr-md py-[10px] bg-[#f8f9fa] border border-[#c7c4d7] rounded-lg font-body-md text-body-md text-[#191c1d] placeholder-[#464554]/60 focus:outline-none focus:border-[#4648d4] focus:ring-[3px] focus:ring-[#4648d4]/10 transition-all duration-200"
                  id="email"
                  name="email"
                  placeholder="developer@strata.dev"
                  required
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                />
              </div>
            </div>

            {/* Password */}
            <div>
              <label className="block font-label-md text-label-md text-[#191c1d] mb-xs" htmlFor="password">
                Password
              </label>
              <div className="relative rounded-md input-glow transition-shadow duration-200">
                <div className="absolute inset-y-0 left-0 pl-sm flex items-center pointer-events-none">
                  <span className="material-symbols-outlined text-[#464554] text-lg">lock</span>
                </div>
                <input
                  className="block w-full pl-xl pr-md py-[10px] bg-[#f8f9fa] border border-[#c7c4d7] rounded-lg font-body-md text-body-md text-[#191c1d] placeholder-[#464554]/60 focus:outline-none focus:border-[#4648d4] focus:ring-[3px] focus:ring-[#4648d4]/10 transition-all duration-200"
                  id="password"
                  name="password"
                  placeholder="••••••••"
                  required
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                />
              </div>
            </div>

            {/* Confirm Password */}
            <div>
              <label className="block font-label-md text-label-md text-[#191c1d] mb-xs" htmlFor="confirmPassword">
                Confirm Password
              </label>
              <div className="relative rounded-md input-glow transition-shadow duration-200">
                <div className="absolute inset-y-0 left-0 pl-sm flex items-center pointer-events-none">
                  <span className="material-symbols-outlined text-[#464554] text-lg">lock_clock</span>
                </div>
                <input
                  className="block w-full pl-xl pr-md py-[10px] bg-[#f8f9fa] border border-[#c7c4d7] rounded-lg font-body-md text-body-md text-[#191c1d] placeholder-[#464554]/60 focus:outline-none focus:border-[#4648d4] focus:ring-[3px] focus:ring-[#4648d4]/10 transition-all duration-200"
                  id="confirmPassword"
                  name="confirmPassword"
                  placeholder="••••••••"
                  required
                  type="password"
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                />
              </div>
            </div>

            {/* Submit Button */}
            <button
              className="w-full flex justify-center items-center gap-sm bg-[#4648d4] hover:bg-[#6063ee] text-white font-label-md text-label-md py-[12px] px-lg rounded-lg shadow-sm transition-all duration-200 mt-lg group disabled:opacity-55"
              type="submit"
              disabled={isSubmitting}
            >
              <span>{isSubmitting ? 'Creating account...' : 'Create Account'}</span>
              <span className="material-symbols-outlined text-[18px] group-hover:translate-x-1 transition-transform duration-200">
                arrow_forward
              </span>
            </button>
          </form>
        </div>

        {/* Footer Link */}
        <p className="mt-lg text-center font-body-md text-body-md text-[#464554]">
          Already have an account?{' '}
          <Link className="font-label-md text-label-md text-[#4648d4] hover:text-[#6063ee] transition-colors ml-xs" to="/login">
            Sign in here
          </Link>
        </p>
      </main>
    </div>
  );
};
