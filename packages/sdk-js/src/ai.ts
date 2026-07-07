import { StrataAuthClient } from './auth';

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

export class AICollectionClient {
  constructor(
    private url: string,
    private collectionName: string,
    private auth: StrataAuthClient
  ) {}

  public async addDocument(content: string, metadata: Record<string, any> = {}): Promise<AIDocument> {
    const finalUrl = `${this.url}/v1/ai/collections/${this.collectionName}/documents`;

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
      body: JSON.stringify({ content, metadata }),
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || `Failed to add document to collection ${this.collectionName}`);
    }

    return data as AIDocument;
  }

  public async listDocuments(): Promise<AIDocument[]> {
    const finalUrl = `${this.url}/v1/ai/collections/${this.collectionName}/documents`;

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
      throw new Error(data.error || `Failed to list documents in collection ${this.collectionName}`);
    }

    return data as AIDocument[];
  }

  public async deleteDocument(docId: string): Promise<{ message: string }> {
    const finalUrl = `${this.url}/v1/ai/collections/${this.collectionName}/documents/${docId}`;

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
      throw new Error(data.error || `Failed to delete document ${docId} from collection ${this.collectionName}`);
    }

    return data as { message: string };
  }

  public async search(query: string, topK: number = 5): Promise<AISearchResponse> {
    const finalUrl = `${this.url}/v1/ai/collections/${this.collectionName}/search`;

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
      body: JSON.stringify({ query, top_k: topK }),
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || `Semantic search in collection ${this.collectionName} failed`);
    }

    return data as AISearchResponse;
  }
}

export class StrataAIClient {
  constructor(private url: string, private auth: StrataAuthClient) {}

  public collection(name: string): AICollectionClient {
    return new AICollectionClient(this.url, name, this.auth);
  }

  public async listCollections(): Promise<AICollection[]> {
    const finalUrl = `${this.url}/v1/ai/collections`;

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
      throw new Error(data.error || 'Failed to list AI collections');
    }

    return data as AICollection[];
  }

  public async createCollection(name: string, description: string = ''): Promise<AICollection> {
    const finalUrl = `${this.url}/v1/ai/collections`;

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
      body: JSON.stringify({ name, description }),
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || `Failed to create AI collection ${name}`);
    }

    return data as AICollection;
  }

  public async deleteCollection(name: string): Promise<{ message: string }> {
    const finalUrl = `${this.url}/v1/ai/collections/${name}`;

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
      throw new Error(data.error || `Failed to delete AI collection ${name}`);
    }

    return data as { message: string };
  }
}
