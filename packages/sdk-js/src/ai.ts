import { StrataAuthClient } from './auth';

// ---------------------------------------------------------------------------
// Embedding types
// ---------------------------------------------------------------------------

export interface EmbeddingsRequest {
  provider?: string;
  model?: string;
  input: string[];
}

export interface EmbeddingsResponse {
  model: string;
  embeddings: number[][];
  usage?: { prompt_tokens: number; total_tokens: number };
}

// ---------------------------------------------------------------------------
// Hub types
// ---------------------------------------------------------------------------

export interface HubPrompt {
  id: string;
  user_id: string;
  name: string;
  description: string;
  category: string;
  content: string;
  variables: number;
  version: number;
  author: string;
  tags: string[];
  created_at: string;
  updated_at: string;
}

export interface HubAgent {
  id: string;
  user_id: string;
  name: string;
  description: string;
  system_prompt: string;
  model: string;
  temperature: number;
  memory: boolean;
  allowed_models: string[];
  functions: string[];
  knowledge_base: string;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface AgentChatRequest {
  message: string;
  history?: ChatMessage[];
}

export interface AgentChatResponse {
  response: string;
  provider: string;
  model: string;
  latency_ms: number;
  usage?: ChatUsage;
}

export interface HubWorkflow {
  id: string;
  user_id: string;
  name: string;
  nodes: any[];
  edges: any[];
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface WorkflowExecuteResponse {
  message: string;
  latency_ms: number;
  nodes: number;
  total_tokens: number;
  outputs: Record<string, string>;
}

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
// Embeddings Client
// ---------------------------------------------------------------------------

export class StrataEmbeddingsClient {
  constructor(private url: string, private auth: StrataAuthClient) {}

  async generate(req: EmbeddingsRequest): Promise<EmbeddingsResponse> {
    return api(`${this.url}/v1/ai/embeddings`, 'POST', req, this.auth);
  }
}

// ---------------------------------------------------------------------------
// Hub Clients (Prompts, Agents, Workflows)
// ---------------------------------------------------------------------------

export class StrataPromptsClient {
  constructor(private url: string, private auth: StrataAuthClient) {}

  async list(): Promise<HubPrompt[]> {
    return api(`${this.url}/v1/ai/hub/prompts`, 'GET', null, this.auth);
  }

  async create(prompt: Partial<HubPrompt>): Promise<HubPrompt> {
    return api(`${this.url}/v1/ai/hub/prompts`, 'POST', prompt, this.auth);
  }

  async update(id: string, prompt: Partial<HubPrompt>): Promise<HubPrompt> {
    return api(`${this.url}/v1/ai/hub/prompts/${id}`, 'PUT', prompt, this.auth);
  }

  async delete(id: string): Promise<void> {
    await api(`${this.url}/v1/ai/hub/prompts/${id}`, 'DELETE', null, this.auth);
  }

  async fork(id: string): Promise<HubPrompt> {
    return api(`${this.url}/v1/ai/hub/prompts/${id}/fork`, 'POST', null, this.auth);
  }
}

export class StrataAgentsClient {
  constructor(private url: string, private auth: StrataAuthClient) {}

  async list(): Promise<HubAgent[]> {
    return api(`${this.url}/v1/ai/hub/agents`, 'GET', null, this.auth);
  }

  async create(agent: Partial<HubAgent>): Promise<HubAgent> {
    return api(`${this.url}/v1/ai/hub/agents`, 'POST', agent, this.auth);
  }

  async update(id: string, agent: Partial<HubAgent>): Promise<void> {
    await api(`${this.url}/v1/ai/hub/agents/${id}`, 'PUT', agent, this.auth);
  }

  async delete(id: string): Promise<void> {
    await api(`${this.url}/v1/ai/hub/agents/${id}`, 'DELETE', null, this.auth);
  }

  async chat(id: string, req: AgentChatRequest): Promise<AgentChatResponse> {
    return api(`${this.url}/v1/ai/hub/agents/${id}/chat`, 'POST', req, this.auth);
  }
}

export class StrataWorkflowsClient {
  constructor(private url: string, private auth: StrataAuthClient) {}

  async list(): Promise<HubWorkflow[]> {
    return api(`${this.url}/v1/ai/hub/workflows`, 'GET', null, this.auth);
  }

  async create(workflow: Partial<HubWorkflow>): Promise<HubWorkflow> {
    return api(`${this.url}/v1/ai/hub/workflows`, 'POST', workflow, this.auth);
  }

  async update(id: string, workflow: Partial<HubWorkflow>): Promise<void> {
    await api(`${this.url}/v1/ai/hub/workflows/${id}`, 'PUT', workflow, this.auth);
  }

  async delete(id: string): Promise<void> {
    await api(`${this.url}/v1/ai/hub/workflows/${id}`, 'DELETE', null, this.auth);
  }

  async execute(id: string): Promise<WorkflowExecuteResponse> {
    return api(`${this.url}/v1/ai/hub/workflows/${id}/execute`, 'POST', null, this.auth);
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
  public embed: StrataEmbeddingsClient;
  public prompts: StrataPromptsClient;
  public agents: StrataAgentsClient;
  public workflows: StrataWorkflowsClient;

  constructor(private url: string, private auth: StrataAuthClient) {
    this.providers = new StrataProviderManager(url, auth);
    this.chat = new StrataChatClient(url, auth);
    this.models = new StrataModelRegistry(url, auth);
    this.usage = new StrataUsageClient(url, auth);
    this.embed = new StrataEmbeddingsClient(url, auth);
    this.prompts = new StrataPromptsClient(url, auth);
    this.agents = new StrataAgentsClient(url, auth);
    this.workflows = new StrataWorkflowsClient(url, auth);
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
