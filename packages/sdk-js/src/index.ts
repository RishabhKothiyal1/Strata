export { StrataClient, StrataClientOptions } from './client';
export {
  StrataAuthClient,
  User,
  Session,
  AuthStateListener,
} from './auth';
export {
  StrataRestClient,
  FilterOperator,
  QueryFilter,
} from './rest';
export {
  StrataStorageClient,
  StorageBucketClient,
  BucketInfo,
  UploadResult,
} from './storage';
export {
  StrataFunctionsClient,
  FunctionInfo,
  InvokeResponse,
} from './functions';
export {
  StrataAIClient,
  AICollectionClient,
  AICollection,
  AIDocument,
  AISearchResult,
  AISearchResponse,
  StrataProviderManager,
  StrataChatClient,
  StrataModelRegistry,
  StrataUsageClient,
  AIProvider,
  CreateProviderRequest,
  UpdateProviderRequest,
  TestProviderResponse,
  ChatRequest,
  ChatMessage,
  ChatResponse,
  ChatChoice,
  ChatUsage,
  ModelInfo,
  UsageStats,
} from './ai';
export {
  StrataRealtimeClient,
  RealtimeChannel,
  SubscriptionCallback,
} from './realtime';
