import { StrataAuthClient } from './auth';

export interface FunctionInfo {
  id: string;
  name: string;
  description: string;
  code: string;
  created_at: string;
  updated_at: string;
}

export interface InvokeResponse<T = any> {
  status_code: number;
  body: T;
  duration_ms: string;
}

export class StrataFunctionsClient {
  constructor(private url: string, private auth: StrataAuthClient) {}

  public async list(): Promise<FunctionInfo[]> {
    const finalUrl = `${this.url}/v1/functions`;

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };
    const session = this.auth.getSession();
    if (session && session.access_token) {
      headers['Authorization'] = `Bearer ${session.access_token}`;
    }

    const res = await fetch(finalUrl, {
      method: 'GET',
      headers,
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || 'Failed to list functions');
    }

    return data as FunctionInfo[];
  }

  public async deploy(
    name: string,
    code: string,
    description: string = ''
  ): Promise<FunctionInfo> {
    const finalUrl = `${this.url}/v1/functions`;

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };
    const session = this.auth.getSession();
    if (session && session.access_token) {
      headers['Authorization'] = `Bearer ${session.access_token}`;
    }

    const res = await fetch(finalUrl, {
      method: 'POST',
      headers,
      body: JSON.stringify({ name, description, code }),
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || `Failed to deploy function ${name}`);
    }

    return data as FunctionInfo;
  }

  public async get(name: string): Promise<FunctionInfo> {
    const finalUrl = `${this.url}/v1/functions/${name}`;

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };
    const session = this.auth.getSession();
    if (session && session.access_token) {
      headers['Authorization'] = `Bearer ${session.access_token}`;
    }

    const res = await fetch(finalUrl, {
      method: 'GET',
      headers,
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || `Failed to get function ${name}`);
    }

    return data as FunctionInfo;
  }

  public async update(
    name: string,
    code: string,
    description: string = ''
  ): Promise<FunctionInfo> {
    const finalUrl = `${this.url}/v1/functions/${name}`;

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };
    const session = this.auth.getSession();
    if (session && session.access_token) {
      headers['Authorization'] = `Bearer ${session.access_token}`;
    }

    const res = await fetch(finalUrl, {
      method: 'PUT',
      headers,
      body: JSON.stringify({ code, description }),
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || `Failed to update function ${name}`);
    }

    return data as FunctionInfo;
  }

  public async delete(name: string): Promise<{ message: string }> {
    const finalUrl = `${this.url}/v1/functions/${name}`;

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };
    const session = this.auth.getSession();
    if (session && session.access_token) {
      headers['Authorization'] = `Bearer ${session.access_token}`;
    }

    const res = await fetch(finalUrl, {
      method: 'DELETE',
      headers,
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || `Failed to delete function ${name}`);
    }

    return data as { message: string };
  }

  public async invoke<T = any>(
    name: string,
    body?: any,
    customHeaders?: Record<string, string>
  ): Promise<InvokeResponse<T>> {
    const finalUrl = `${this.url}/v1/functions/${name}/invoke`;

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...customHeaders,
    };
    const session = this.auth.getSession();
    if (session && session.access_token) {
      headers['Authorization'] = `Bearer ${session.access_token}`;
    }

    const res = await fetch(finalUrl, {
      method: 'POST',
      headers,
      body: body ? JSON.stringify({ body }) : undefined,
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || `Invocation of function ${name} failed`);
    }

    return data as InvokeResponse<T>;
  }
}
