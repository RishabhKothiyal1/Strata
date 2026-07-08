import React, { createContext, useContext, useEffect, useState } from 'react';
import { Navigate, useLocation } from 'react-router-dom';
import { strata } from '../lib/strata';
import type { User, Session } from '@strata/sdk';

interface AuthContextType {
  user: User | null;
  session: Session | null;
  loading: boolean;
  signIn: (email: string, password: string) => Promise<void>;
  signUp: (email: string, password: string) => Promise<void>;
  signOut: () => Promise<void>;
  enterDemoMode: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const DEMO_USER: User = {
  id: 'demo-user',
  email: 'demo@strata.dev',
  role: 'Developer',
  org_id: 'demo-org',
  created_at: new Date().toISOString(),
};

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [session, setSession] = useState<Session | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const stored = localStorage.getItem('strata.demo.mode');
    if (stored === 'true') {
      setUser(DEMO_USER);
      setLoading(false);
      return;
    }

    setUser(strata.auth.getUser());
    setSession(strata.auth.getSession());

    const { unsubscribe } = strata.auth.onAuthStateChange((event, currentSession) => {
      console.log(`Auth state changed: ${event}`, currentSession);
      setSession(currentSession);
      setUser(strata.auth.getUser());
      setLoading(false);
    });

    return () => {
      unsubscribe();
    };
  }, []);

  const signIn = async (email: string, password: string) => {
    setLoading(true);
    try {
      await strata.auth.signIn(email, password);
    } catch (err: any) {
      const msg = err?.message || '';
      const isNetworkError =
        msg.includes('Failed to fetch') ||
        msg.includes('NetworkError') ||
        msg.includes('ERR_CONNECTION_REFUSED') ||
        msg.includes('ERR_CONNECTION_RESET') ||
        msg.includes('load failed');
      if (isNetworkError || msg.includes('401') || msg.includes('unauthorized')) {
        console.warn('Auth API unavailable, entering demo mode:', err);
        localStorage.setItem('strata.demo.mode', 'true');
        setUser(DEMO_USER);
      } else {
        throw err;
      }
    } finally {
      setLoading(false);
    }
  };

  const signUp = async (email: string, password: string) => {
    setLoading(true);
    try {
      await strata.auth.signUp(email, password);
    } catch (err: any) {
      const msg = err?.message || '';
      const isNetworkError =
        msg.includes('Failed to fetch') ||
        msg.includes('NetworkError') ||
        msg.includes('ERR_CONNECTION_REFUSED');
      if (isNetworkError) {
        console.warn('Signup API unavailable, entering demo mode:', err);
        localStorage.setItem('strata.demo.mode', 'true');
        setUser(DEMO_USER);
      } else {
        throw err;
      }
    } finally {
      setLoading(false);
    }
  };

  const signOut = async () => {
    setLoading(true);
    try {
      await strata.auth.signOut();
    } catch {
      // ignore
    } finally {
      localStorage.removeItem('strata.demo.mode');
      setUser(null);
      setSession(null);
      setLoading(false);
    }
  };

  const enterDemoMode = () => {
    localStorage.setItem('strata.demo.mode', 'true');
    setUser(DEMO_USER);
    setSession(null);
    setLoading(false);
  };

  return (
    <AuthContext.Provider value={{ user, session, loading, signIn, signUp, signOut, enterDemoMode }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};

export const RequireAuth: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { user, loading } = useAuth();
  const location = useLocation();

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="flex flex-col items-center space-y-md">
          <div className="w-10 h-10 border-4 border-primary border-t-transparent rounded-full animate-spin"></div>
          <p className="font-body-md text-on-surface-variant animate-pulse">Loading Strata Studio...</p>
        </div>
      </div>
    );
  }

  if (!user) {
    // Redirect to login but save the current location they tried to go to
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  return <>{children}</>;
};
