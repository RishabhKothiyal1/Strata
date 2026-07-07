import { NovaBaseAuthClient } from './auth';

let WS: any;
if (typeof WebSocket !== 'undefined') {
  WS = WebSocket;
} else {
  try {
    // Dynamic import to satisfy browser bundlers
    const wsModule = require('ws');
    WS = wsModule.default || wsModule;
  } catch (e) {
    // fallback
  }
}

export type SubscriptionCallback = (payload: { channel: string; data: any; sender_id: number }) => void;

export class RealtimeChannel {
  private isSubscribed = false;

  constructor(
    private channelName: string,
    private client: NovaBaseRealtimeClient
  ) {}

  public subscribe(callback: SubscriptionCallback): { unsubscribe: () => void } {
    this.client.registerCallback(this.channelName, callback);
    this.client.ensureConnected();

    if (!this.isSubscribed) {
      this.client.send({ event: 'subscribe', channel: this.channelName });
      this.isSubscribed = true;
    }

    return {
      unsubscribe: () => {
        this.unsubscribe(callback);
      },
    };
  }

  public broadcast(data: any): void {
    this.client.ensureConnected();
    this.client.send({ event: 'broadcast', channel: this.channelName, data });
  }

  private unsubscribe(callback: SubscriptionCallback): void {
    this.client.deregisterCallback(this.channelName, callback);
    if (this.client.getCallbackCount(this.channelName) === 0) {
      this.client.send({ event: 'unsubscribe', channel: this.channelName });
      this.isSubscribed = false;
    }
  }
}

export class NovaBaseRealtimeClient {
  private ws: any = null;
  private callbacks: Map<string, Set<SubscriptionCallback>> = new Map();
  private sendQueue: any[] = [];
  private isConnecting = false;

  constructor(private url: string, private auth: NovaBaseAuthClient) {}

  public channel(channelName: string): RealtimeChannel {
    return new RealtimeChannel(channelName, this);
  }

  public registerCallback(channel: string, callback: SubscriptionCallback) {
    if (!this.callbacks.has(channel)) {
      this.callbacks.set(channel, new Set());
    }
    this.callbacks.get(channel)!.add(callback);
  }

  public deregisterCallback(channel: string, callback: SubscriptionCallback) {
    if (this.callbacks.has(channel)) {
      this.callbacks.get(channel)!.delete(callback);
      if (this.callbacks.get(channel)!.size === 0) {
        this.callbacks.delete(channel);
      }
    }
  }

  public getCallbackCount(channel: string): number {
    return this.callbacks.has(channel) ? this.callbacks.get(channel)!.size : 0;
  }

  public ensureConnected() {
    if (this.ws && (this.ws.readyState === 0 || this.ws.readyState === 1)) {
      return;
    }
    if (this.isConnecting) {
      return;
    }

    this.isConnecting = true;
    // Map HTTP/HTTPS url to WS/WSS
    let wsUrl = this.url.replace(/^http/, 'ws');
    // Direct endpoint or gateway endpoint
    wsUrl = `${wsUrl}/v1/realtime`;

    if (!WS) {
      throw new Error('WebSocket client is not available. Please install "ws" dependency if running in Node.js.');
    }

    try {
      this.ws = new WS(wsUrl);
    } catch (e) {
      this.isConnecting = false;
      throw e;
    }

    this.ws.onopen = () => {
      this.isConnecting = false;
      // Resend subscriptions
      for (const channel of this.callbacks.keys()) {
        this.send({ event: 'subscribe', channel });
      }
      // Empty queue
      while (this.sendQueue.length > 0) {
        const msg = this.sendQueue.shift();
        this.send(msg);
      }
    };

    this.ws.onmessage = (event: any) => {
      try {
        const payload = JSON.parse(event.data);
        const channel = payload.channel;
        if (channel && this.callbacks.has(channel)) {
          const list = this.callbacks.get(channel)!;
          for (const cb of list) {
            try {
              cb(payload);
            } catch (err) {
              console.error('Error in subscription callback:', err);
            }
          }
        }
      } catch (err) {
        // ignore unparseable messages
      }
    };

    this.ws.onerror = (error: any) => {
      console.error('Realtime WebSocket error:', error);
    };

    this.ws.onclose = () => {
      this.isConnecting = false;
      this.ws = null;
      // Auto-reconnect after 3 seconds
      setTimeout(() => {
        if (this.callbacks.size > 0) {
          try {
            this.ensureConnected();
          } catch (e) {
            // silent retry
          }
        }
      }, 3000);
    };
  }

  public send(msg: any) {
    if (this.ws && this.ws.readyState === 1) {
      this.ws.send(JSON.stringify(msg));
    } else {
      this.sendQueue.push(msg);
    }
  }

  public disconnect() {
    if (this.ws) {
      this.ws.close();
      this.ws = null;
    }
    this.isConnecting = false;
    this.callbacks.clear();
    this.sendQueue = [];
  }
}
