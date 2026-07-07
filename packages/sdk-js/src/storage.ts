import { NovaBaseAuthClient } from './auth';

export interface BucketInfo {
  name: string;
  created_at: string;
}

export interface UploadResult {
  filepath: string;
  url: string;
  size: number;
}

export class StorageBucketClient {
  constructor(
    private url: string,
    private bucket: string,
    private auth: NovaBaseAuthClient
  ) {}

  public async upload(
    filePath: string,
    fileBody: any,
    contentType: string = 'application/octet-stream'
  ): Promise<UploadResult> {
    const finalUrl = `${this.url}/v1/storage/buckets/${this.bucket}/upload`;

    const headers: Record<string, string> = {};
    const session = this.auth.getSession();
    if (session && session.access_token) {
      headers['Authorization'] = `Bearer ${session.access_token}`;
    }

    const formData = new FormData();
    
    // Check if it's running in Node.js and fileBody is a Buffer/ArrayBuffer
    if (typeof window === 'undefined') {
      // In Node.js environment
      if (Buffer.isBuffer(fileBody) || fileBody instanceof ArrayBuffer) {
        const buffer = Buffer.isBuffer(fileBody) ? fileBody : Buffer.from(fileBody);
        const blob = new Blob([buffer as any], { type: contentType });
        formData.append('file', blob, filePath);
      } else {
        formData.append('file', fileBody, filePath);
      }
    } else {
      // In Browser environment
      formData.append('file', fileBody, filePath);
    }

    const res = await fetch(finalUrl, {
      method: 'POST',
      headers,
      body: formData,
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || `Upload to bucket ${this.bucket} failed`);
    }

    return data as UploadResult;
  }

  public async download(
    filePath: string,
    options?: { width?: number; height?: number }
  ): Promise<Response> {
    const queryParams = new URLSearchParams();
    if (options?.width !== undefined) {
      queryParams.set('width', options.width.toString());
    }
    if (options?.height !== undefined) {
      queryParams.set('height', options.height.toString());
    }

    const queryString = queryParams.toString();
    const finalUrl = `${this.url}/v1/storage/buckets/${this.bucket}/download/${filePath}${
      queryString ? `?${queryString}` : ''
    }`;

    const headers: Record<string, string> = {};
    const session = this.auth.getSession();
    if (session && session.access_token) {
      headers['Authorization'] = `Bearer ${session.access_token}`;
    }

    const res = await fetch(finalUrl, {
      method: 'GET',
      headers,
    });

    if (!res.ok) {
      throw new Error(`Download of file ${filePath} failed`);
    }

    return res;
  }
}

export class NovaBaseStorageClient {
  constructor(private url: string, private auth: NovaBaseAuthClient) {}

  public from(bucket: string): StorageBucketClient {
    return new StorageBucketClient(this.url, bucket, this.auth);
  }

  public async listBuckets(): Promise<BucketInfo[]> {
    const finalUrl = `${this.url}/v1/storage/buckets`;

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
      throw new Error(data.error || 'Failed to list buckets');
    }

    return data as BucketInfo[];
  }

  public async createBucket(name: string): Promise<{ message: string; bucket: string }> {
    const finalUrl = `${this.url}/v1/storage/buckets`;

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
      body: JSON.stringify({ name }),
    });

    const data = await res.json();
    if (!res.ok) {
      throw new Error(data.error || `Failed to create bucket ${name}`);
    }

    return data as { message: string; bucket: string };
  }

  public async deleteBucket(name: string): Promise<{ message: string }> {
    const finalUrl = `${this.url}/v1/storage/buckets/${name}`;

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
      throw new Error(data.error || `Failed to delete bucket ${name}`);
    }

    return data as { message: string };
  }
}
