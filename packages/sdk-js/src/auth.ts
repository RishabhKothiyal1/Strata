export interface User {
  id: string | number;
  email: string;
  role: string;
  org_id: string;
  created_at: string;
}

export interface Session {
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export type AuthStateListener = (event: 'SIGNED_IN' | 'SIGNED_OUT' | 'TOKEN_REFRESHED', session: Session | null) => void;

export class NovaBaseAuthClient {
  private session: Session | null = null;
  private user: User | null = null;
  private listeners: Set<AuthStateListener> = new Set();

  constructor(private url: string) {
    this.loadSession();
  }

  private loadSession() {
    if (typeof window !== 'undefined' && window.localStorage) {
      const stored = window.localStorage.getItem('novabase.session');
      if (stored) {
        try {
          this.session = JSON.parse(stored);
          // Simple base64 decoding of JWT payload to get user
          if (this.session && this.session.access_token) {
            const parts = this.session.access_token.split('.');
            if (parts.length === 3) {
              const payload = JSON.parse(Buffer ? Buffer.from(parts[1], 'base64').toString() : atob(parts[1]));
              this.user = {
                id: payload.sub,
                email: payload.email,
                role: payload.role,
                org_id: payload.org_id,
                created_at: new Date(payload.iat * 1000).toISOString(),
              };
            }
          }
        } catch (e) {
          this.clearSession();
        }
      }
    }
  }

  private saveSession(session: Session) {
    this.session = session;
    if (typeof window !== 'undefined' && window.localStorage) {
      window.localStorage.setItem('novabase.session', JSON.stringify(session));
    }
    // Decode user
    try {
      const parts = session.access_token.split('.');
      const payload = JSON.parse(
        typeof Buffer !== 'undefined'
          ? Buffer.from(parts[1], 'base64').toString()
          : atob(parts[1])
      );
      this.user = {
        id: payload.sub,
        email: payload.email,
        role: payload.role,
        org_id: payload.org_id,
        created_at: new Date(payload.iat * 1000).toISOString(),
      };
    } catch (e) {
      // ignore
    }
  }

  private clearSession() {
    this.session = null;
    this.user = null;
    if (typeof window !== 'undefined' && window.localStorage) {
      window.localStorage.removeItem('novabase.session');
    }
  }

  public getSession(): Session | null {
    return this.session;
  }

  public getUser(): User | null {
    return this.user;
  }

  public onAuthStateChange(listener: AuthStateListener): { unsubscribe: () => void } {
    this.listeners.add(listener);
    // Emit initial state
    listener(this.session ? 'SIGNED_IN' : 'SIGNED_OUT', this.session);
    return {
      unsubscribe: () => {
        this.listeners.delete(listener);
      },
    };
  }

  private emit(event: 'SIGNED_IN' | 'SIGNED_OUT' | 'TOKEN_REFRESHED', session: Session | null) {
    for (const listener of this.listeners) {
      try {
        listener(event, session);
      } catch (e) {
        console.error(e);
      }
    }
  }

  public async signUp(email: string, password: string, options?: { role?: string; org_id?: string }): Promise<User> {
    const res = await fetch(`${this.url}/v1/auth/register`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password, ...options }),
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || 'Registration failed');
    }

    return data as User;
  }

  public async signIn(email: string, password: string): Promise<{ session: Session; user: User }> {
    const res = await fetch(`${this.url}/v1/auth/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password }),
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || 'Login failed');
    }

    const session: Session = {
      access_token: data.access_token,
      refresh_token: data.refresh_token,
      expires_in: data.expires_in,
    };

    this.saveSession(session);
    this.emit('SIGNED_IN', session);

    return { session, user: this.user! };
  }

  public async signOut(): Promise<void> {
    if (this.session) {
      try {
        await fetch(`${this.url}/v1/auth/logout`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ refresh_token: this.session.refresh_token }),
        });
      } catch (e) {
        // Log out locally even if server call fails
      }
    }
    this.clearSession();
    this.emit('SIGNED_OUT', null);
  }

  public async refreshSession(): Promise<Session> {
    if (!this.session) {
      throw new Error('No active session to refresh');
    }

    const res = await fetch(`${this.url}/v1/auth/refresh`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: this.session.refresh_token }),
    });

    const data = await res.json();
    if (!res.ok) {
      this.clearSession();
      this.emit('SIGNED_OUT', null);
      throw new Error(data.error || 'Token refresh failed');
    }

    const session: Session = {
      access_token: data.access_token,
      refresh_token: data.refresh_token,
      expires_in: data.expires_in,
    };

    this.saveSession(session);
    this.emit('TOKEN_REFRESHED', session);

    return session;
  }
}
