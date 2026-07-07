import { NovaBaseAuthClient } from './auth';
import { NovaBaseRestClient } from './rest';
import { NovaBaseStorageClient } from './storage';
import { NovaBaseFunctionsClient } from './functions';
import { NovaBaseAIClient } from './ai';
import { NovaBaseRealtimeClient } from './realtime';

export interface NovaBaseClientOptions {
  apiKey?: string;
}

export class NovaBaseClient {
  public auth: NovaBaseAuthClient;
  public storage: NovaBaseStorageClient;
  public functions: NovaBaseFunctionsClient;
  public ai: NovaBaseAIClient;
  public realtime: NovaBaseRealtimeClient;

  constructor(private url: string, private options: NovaBaseClientOptions = {}) {
    // Normalise trailing slashes
    this.url = url.replace(/\/+$/, '');

    this.auth = new NovaBaseAuthClient(this.url);
    this.storage = new NovaBaseStorageClient(this.url, this.auth);
    this.functions = new NovaBaseFunctionsClient(this.url, this.auth);
    this.ai = new NovaBaseAIClient(this.url, this.auth);
    this.realtime = new NovaBaseRealtimeClient(this.url, this.auth);
  }

  /**
   * Perform REST actions on database tables.
   */
  public from<T = any>(table: string): NovaBaseRestClient<T> {
    return new NovaBaseRestClient<T>(this.url, table, this.auth);
  }
}
