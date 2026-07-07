import { StrataAuthClient } from './auth';
import { StrataRestClient } from './rest';
import { StrataStorageClient } from './storage';
import { StrataFunctionsClient } from './functions';
import { StrataAIClient } from './ai';
import { StrataRealtimeClient } from './realtime';

export interface StrataClientOptions {
  apiKey?: string;
}

export class StrataClient {
  public auth: StrataAuthClient;
  public storage: StrataStorageClient;
  public functions: StrataFunctionsClient;
  public ai: StrataAIClient;
  public realtime: StrataRealtimeClient;

  constructor(private url: string, private options: StrataClientOptions = {}) {
    this.url = url.replace(/\/+$/, '');
    this.auth = new StrataAuthClient(this.url);
    this.storage = new StrataStorageClient(this.url, this.auth);
    this.functions = new StrataFunctionsClient(this.url, this.auth);
    this.ai = new StrataAIClient(this.url, this.auth);
    this.realtime = new StrataRealtimeClient(this.url, this.auth);
  }

  /**
   * Perform REST actions on database tables.
   */
  public from<T = any>(table: string): StrataRestClient<T> {
    return new StrataRestClient<T>(this.url, table, this.auth);
  }
}
