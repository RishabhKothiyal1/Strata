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
} from './ai';
export {
  StrataRealtimeClient,
  RealtimeChannel,
  SubscriptionCallback,
} from './realtime';
