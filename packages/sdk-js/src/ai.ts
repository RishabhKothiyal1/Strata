import { StrataAuthClient } from './auth';

// ---------------------------------------------------------------------------
// Provider types
// ---------------------------------------------------------------------------

export interface AIProvider {
  id: string;
  user_id: string;
  provider: string;
  base_url?: string;
  default_model: string;
  enabled: boolean;
  is_primary: boolean;
  fallback_order: number;
  created_at: string;
  updated_at: string;
}

export interface CreateProviderRequest {
  provider: string;
  api_key: string;
  base_url?: string;
  default_model?: string;
}

export interface UpdateProviderRequest {
  api_key?: string;
  base_url?: string;
  default_model?: string;
  enabled?: boolean;
  is_primary?: boolean;
}

export interface TestProviderResponse {
  status: 'ok' | 'error';
  message: string;
  latency_ms: number;
}

// ---------------------------------------------------------------------------
// Collection types
// ---------------------------------------------------------------------------

export interface AICollection {
  id: string;
  name: string;
  description: string;
  doc_count: number;
  created_at: string;
}

export interface AIDocument {
  id: string;
  collection_id: string;
  content: string;
  metadata: Record<string, any>;
  created_at: string;
}

export interface AISearchResult {
  id: string;
  collection_id: string;
  content: string;
  metadata: Record<string, any>;
  created_at: string;
  score: number;
}

export interface AISearchResponse {
  query: string;
  top_k: number;
  results: AISearchResult[];
}

// ---------------------------------------------------------------------------
// Chat types
// ---------------------------------------------------------------------------

export interface ChatMessage {
  role: 'system' | 'user' | 'assistant';
  content: string;
}

export interface ChatRequest {
  model: string;
  messages: ChatMessage[];
  temperature?: number;
  max_tokens?: number;
  provider?: string;
}

export interface ChatChoice {
  index: number;
  message: ChatMessage;
  finish_reason: string;
}

export interface ChatUsage {
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
}

export interface ChatResponse {
  id: string;
  model: string;
  choices: ChatChoice[];
  usage?: ChatUsage;
  provider: string;
  latency_ms: number;
}

// ---------------------------------------------------------------------------
// Model types
// ---------------------------------------------------------------------------

export interface ModelInfo {
  id: string;
  provider: string;
  name: string;
  category: string;
}

// ---------------------------------------------------------------------------
// Usage types
// ---------------------------------------------------------------------------

export interface UsageStats {
  provider: string;
  model: string;
  total_requests: number;
  successful_requests: number;
  failed_requests: number;
  total_tokens: number;
  avg_latency_ms: number;
  estimated_cost: number;
}

// ---------------------------------------------------------------------------
// HTTP helper
// ---------------------------------------------------------------------------

function headers(auth: StrataAuthClient): Record<string, string> {
  const h: Record<string, string> = { 'Content-Type': 'application/json' };
  const session = auth.getSession();
  if (session?.access_token) {
    h['Authorization'] = `Bearer ${session.access_token}`;
  }
  return h;
}

async function api(url: string, method: string, body: any, auth: StrataAuthClient): Promise<any> {
  const res = await fetch(url, {
    method,
    headers: headers(auth),
    body: body ? JSON.stringify(body) : undefined,
  });
  const data = await res.json();
  if (!res.ok) {
    throw new Error(data.error || `HTTP ${res.status}`);
  }
  return data;
}

// ---------------------------------------------------------------------------
// Provider Manager
// ---------------------------------------------------------------------------

export class StrataProviderManager {
  constructor(private url: string, private auth: StrataAuthClient) {}

  async list(): Promise<AIProvider[]> {
    return api(`${this.url}/v1/ai/providers`, 'GET', null, this.auth);
  }

  async create(req: CreateProviderRequest): Promise<AIProvider> {
    return api(`${this.url}/v1/ai/providers`, 'POST', req, this.auth);
  }

  async update(provider: string, req: UpdateProviderRequest): Promise<void> {
    await api(`${this.url}/v1/ai/providers/${provider}`, 'PUT', req, this.auth);
  }

  async delete(provider: string): Promise<void> {
    await api(`${this.url}/v1/ai/providers/${provider}`, 'DELETE', null, this.auth);
  }

  async test(provider: string): Promise<TestProviderResponse> {
    return api(`${this.url}/v1/ai/providers/${provider}/test`, 'POST', null, this.auth);
  }

  async models(provider: string): Promise<string[]> {
    return api(`${this.url}/v1/ai/providers/${provider}/models`, 'GET', null, this.auth);
  }
}

// ---------------------------------------------------------------------------
// AI Chat
// ---------------------------------------------------------------------------

export class StrataChatClient {
  constructor(private url: string, private auth: StrataAuthClient) {}

  async chat(req: ChatRequest): Promise<ChatResponse> {
    return api(`${this.url}/v1/ai/chat`, 'POST', req, this.auth);
  }

  async stream(req: ChatRequest, onChunk: (content: string) => void): Promise<void> {
    const h = headers(this.auth);
    const res = await fetch(`${this.url}/v1/ai/chat/stream`, {
      method: 'POST',
      headers: { ...h, Accept: 'text/event-stream' },
      body: JSON.stringify(req),
    });

    if (!res.ok) {
      const data = await res.json();
      throw new Error(data.error || `HTTP ${res.status}`);
    }

    const reader = res.body?.getReader();
    if (!reader) return;

    const decoder = new TextDecoder();
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      const text = decoder.decode(value);
      const lines = text.split('\n');
      for (const line of lines) {
        if (line.startsWith('data: ')) {
          try {
            const data = JSON.parse(line.slice(6));
            if (data.done) return;
            if (data.content) onChunk(data.content);
          } catch {}
        }
      }
    }
  }
}

// ---------------------------------------------------------------------------
// Model Registry
// ---------------------------------------------------------------------------

export class StrataModelRegistry {
  constructor(private url: string, private auth: StrataAuthClient) {}

  async list(provider?: string): Promise<ModelInfo[]> {
    const qs = provider ? `?provider=${provider}` : '';
    return api(`${this.url}/v1/ai/models${qs}`, 'GET', null, this.auth);
  }
}

// ---------------------------------------------------------------------------
// Usage
// ---------------------------------------------------------------------------

export class StrataUsageClient {
  constructor(private url: string, private auth: StrataAuthClient) {}

  async stats(limit: number = 50): Promise<UsageStats[]> {
    return api(`${this.url}/v1/ai/usage?limit=${limit}`, 'GET', null, this.auth);
  }
}

// ---------------------------------------------------------------------------
// Collection Client (existing with minor updates)
// ---------------------------------------------------------------------------

export class AICollectionClient {
  constructor(
    private url: string,
    private collectionName: string,
    private auth: StrataAuthClient
  ) {}

  public async addDocument(content: string, metadata: Record<string, any> = {}): Promise<AIDocument> {
    return api(`${this.url}/v1/ai/collections/${this.collectionName}/documents`, 'POST', { content, metadata }, this.auth);
  }

  public async listDocuments(): Promise<AIDocument[]> {
    return api(`${this.url}/v1/ai/collections/${this.collectionName}/documents`, 'GET', null, this.auth);
  }

  public async deleteDocument(docId: string): Promise<{ message: string }> {
    return api(`${this.url}/v1/ai/collections/${this.collectionName}/documents/${docId}`, 'DELETE', null, this.auth);
  }

  public async search(query: string, topK: number = 5): Promise<AISearchResponse> {
    return api(`${this.url}/v1/ai/collections/${this.collectionName}/search`, 'POST', { query, top_k: topK }, this.auth);
  }
}

// ---------------------------------------------------------------------------
// StrataAIClient (main entry point)
// ---------------------------------------------------------------------------

export class StrataAIClient {
  public providers: StrataProviderManager;
  public chat: StrataChatClient;
  public models: StrataModelRegistry;
  public usage: StrataUsageClient;

  constructor(private url: string, private auth: StrataAuthClient) {
    this.providers = new StrataProviderManager(url, auth);
    this.chat = new StrataChatClient(url, auth);
    this.models = new StrataModelRegistry(url, auth);
    this.usage = new StrataUsageClient(url, auth);
  }

  // Legacy collection methods
  public collection(name: string): AICollectionClient {
    return new AICollectionClient(this.url, name, this.auth);
  }

  public async listCollections(): Promise<AICollection[]> {
    return api(`${this.url}/v1/ai/collections`, 'GET', null, this.auth);
  }

  public async createCollection(name: string, description: string = ''): Promise<AICollection> {
    return api(`${this.url}/v1/ai/collections`, 'POST', { name, description }, this.auth);
  }

  public async deleteCollection(name: string): Promise<{ message: string }> {
    return api(`${this.url}/v1/ai/collections/${name}`, 'DELETE', null, this.auth);
  }
}
