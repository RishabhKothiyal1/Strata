import { strata } from './strata';

/**
 * Executes a Strata SDK operation, catching auth errors (401),
 * attempting to refresh the session once, and retrying.
 */
export async function withAuthRetry<T>(operation: () => Promise<T>): Promise<T> {
  try {
    return await operation();
  } catch (error: any) {
    const isAuthError =
      error instanceof Error &&
      (error.message.includes('401') ||
        error.message.toLowerCase().includes('unauthorized') ||
        error.message.toLowerCase().includes('token expired') ||
        error.message.toLowerCase().includes('jwt'));

    if (isAuthError) {
      console.warn('Auth token expired or unauthorized. Attempting token refresh...', error);
      try {
        await strata.auth.refreshSession();
        // Retry the operation once with the refreshed token
        return await operation();
      } catch (refreshError) {
        console.error('Session refresh failed:', refreshError);
        // If session refresh fails, propagate original error or sign out
        throw error;
      }
    }
    
    // Propagate non-auth errors
    throw error;
  }
}
