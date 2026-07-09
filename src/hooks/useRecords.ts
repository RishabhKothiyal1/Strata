import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { strata } from '../lib/strata';
import { useAuth } from '../context/AuthContext';
import { withAuthRetry } from '../lib/withAuthRetry';

export interface DatabaseRecord {
  id: number;
  user_id: string;
  title: string;
  body: string;
  active: boolean;
  created_at: string;
}

let localIdCounter = 1;
const localStore = new Map<string, DatabaseRecord[]>();

function getLocalRecords(tableName: string, userId: string): DatabaseRecord[] {
  const key = `${tableName}:${userId}`;
  return localStore.get(key) || [];
}

function setLocalRecords(tableName: string, userId: string, records: DatabaseRecord[]) {
  const key = `${tableName}:${userId}`;
  localStore.set(key, records);
}

export function useRecords(tableName: string) {
  const { user } = useAuth();
  const userId = user?.id ? String(user.id) : 'anonymous';
  const queryKey = ['records', tableName, userId];

  return useQuery<DatabaseRecord[]>({
    queryKey,
    queryFn: async () => {
      if (!user) return getLocalRecords(tableName, 'anonymous');
      try {
        return await withAuthRetry(async () => {
          return await strata
            .from<DatabaseRecord[]>(tableName)
            .select('*')
            .eq('user_id', user.id)
            .order('created_at', 'desc')
            .execute();
        });
      } catch (err) {
        console.warn('API unavailable, using local storage:', err);
        return getLocalRecords(tableName, String(user.id));
      }
    },
    enabled: true,
  });
}

export function useCreateRecord(tableName: string) {
  const queryClient = useQueryClient();
  const { user } = useAuth();

  function getQueryKey() {
    return ['records', tableName, user?.id ? String(user.id) : 'anonymous'];
  }

  function getUserId() {
    return user?.id ? String(user.id) : 'anonymous';
  }

  return useMutation({
    mutationFn: async (record: Omit<DatabaseRecord, 'id' | 'user_id' | 'created_at'>) => {
      const userId = getUserId();
      try {
        return await withAuthRetry(async () => {
          return await strata
            .from(tableName)
            .insert({
              ...record,
              user_id: userId,
            })
            .execute();
        });
      } catch (err: any) {
        const isNetworkError =
          err instanceof TypeError ||
          err?.message?.includes('Failed to fetch') ||
          err?.message?.includes('NetworkError') ||
          err?.message?.includes('ERR_CONNECTION_REFUSED') ||
          err?.message?.includes('ERR_CONNECTION_RESET') ||
          err?.message?.includes('load failed');
        if (isNetworkError) {
          console.warn('API unavailable, saving locally:', err);
          const newRecord: DatabaseRecord = {
            id: localIdCounter++,
            user_id: userId,
            title: record.title || 'Untitled',
            body: record.body || '',
            active: record.active ?? true,
            created_at: new Date().toISOString(),
          };
          const existing = getLocalRecords(tableName, userId);
          setLocalRecords(tableName, userId, [newRecord, ...existing]);
          return newRecord;
        }
        throw err;
      }
    },
    onSuccess: (data: any) => {
      const queryKey = getQueryKey();
      const userId = getUserId();
      const record = data as DatabaseRecord;

      // Save to local storage so fallback queries can find it
      if (record) {
        const existing = getLocalRecords(tableName, userId);
        const updated = [record, ...existing.filter((r) => r.id !== record.id)];
        setLocalRecords(tableName, userId, updated);
      }

      // Update query cache immediately so the record is visible
      queryClient.setQueryData(queryKey, (old: DatabaseRecord[] | undefined) => {
        if (!record) return old;
        if (!old) return [record];
        return [record, ...old.filter((r) => r.id !== record.id)];
      });

      // Background refetch to ensure consistency
      queryClient.invalidateQueries({ queryKey });
      try {
        strata.realtime.channel(tableName).broadcast({
          event: 'mutation',
          type: 'INSERT',
        });
      } catch (err) {
        // broadcast failure is non-critical
      }
    },
  });
}

export function useUpdateRecord(tableName: string) {
  const queryClient = useQueryClient();
  const { user } = useAuth();

  function getQueryKey() {
    return ['records', tableName, user?.id ? String(user.id) : 'anonymous'];
  }

  function getUserId() {
    return user?.id ? String(user.id) : 'anonymous';
  }

  return useMutation({
    mutationFn: async ({ id, ...updates }: Partial<DatabaseRecord> & { id: number }) => {
      const userId = getUserId();
      try {
        return await withAuthRetry(async () => {
          return await strata
            .from(tableName)
            .update(updates)
            .eq('id', id)
            .eq('user_id', userId)
            .execute();
        });
      } catch (err: any) {
        const isNetworkError =
          err instanceof TypeError ||
          err?.message?.includes('Failed to fetch') ||
          err?.message?.includes('NetworkError') ||
          err?.message?.includes('ERR_CONNECTION_REFUSED') ||
          err?.message?.includes('ERR_CONNECTION_RESET') ||
          err?.message?.includes('load failed');
        if (isNetworkError) {
          console.warn('API unavailable, updating locally:', err);
          const existing = getLocalRecords(tableName, userId);
          const updated = existing.map((r) =>
            r.id === id ? { ...r, ...updates } : r
          );
          setLocalRecords(tableName, userId, updated);
          return { id, ...updates };
        }
        throw err;
      }
    },
    onSuccess: (_data: any, variables) => {
      const queryKey = getQueryKey();
      const userId = getUserId();

      // Save to local storage for fallback queries
      const existing = getLocalRecords(tableName, userId);
      const updated = existing.map((r) =>
        r.id === variables.id ? { ...r, ...variables } : r
      );
      setLocalRecords(tableName, userId, updated);

      // Update query cache immediately
      queryClient.setQueryData(queryKey, (old: DatabaseRecord[] | undefined) => {
        if (!old) return old;
        return old.map((r) =>
          r.id === variables.id ? { ...r, ...variables } : r
        );
      });

      // Background refetch
      queryClient.invalidateQueries({ queryKey });
      try {
        strata.realtime.channel(tableName).broadcast({
          event: 'mutation',
          type: 'UPDATE',
        });
      } catch (err) {
        // broadcast failure is non-critical
      }
    },
  });
}

export function useDeleteRecord(tableName: string) {
  const queryClient = useQueryClient();
  const { user } = useAuth();

  function getQueryKey() {
    return ['records', tableName, user?.id ? String(user.id) : 'anonymous'];
  }

  function getUserId() {
    return user?.id ? String(user.id) : 'anonymous';
  }

  return useMutation({
    mutationFn: async (id: number) => {
      const userId = getUserId();
      try {
        return await withAuthRetry(async () => {
          return await strata
            .from(tableName)
            .delete()
            .eq('id', id)
            .eq('user_id', userId)
            .execute();
        });
      } catch (err: any) {
        const isNetworkError =
          err instanceof TypeError ||
          err?.message?.includes('Failed to fetch') ||
          err?.message?.includes('NetworkError') ||
          err?.message?.includes('ERR_CONNECTION_REFUSED') ||
          err?.message?.includes('ERR_CONNECTION_RESET') ||
          err?.message?.includes('load failed');
        if (isNetworkError) {
          console.warn('API unavailable, deleting locally:', err);
          const existing = getLocalRecords(tableName, userId);
          setLocalRecords(tableName, userId, existing.filter((r) => r.id !== id));
          return { success: true };
        }
        throw err;
      }
    },
    onSuccess: (_data: any, id: number) => {
      const queryKey = getQueryKey();
      const userId = getUserId();

      // Remove from local storage
      const existing = getLocalRecords(tableName, userId);
      setLocalRecords(tableName, userId, existing.filter((r) => r.id !== id));

      // Update query cache immediately
      queryClient.setQueryData(queryKey, (old: DatabaseRecord[] | undefined) => {
        if (!old) return old;
        return old.filter((r) => r.id !== id);
      });

      // Background refetch
      queryClient.invalidateQueries({ queryKey });
      try {
        strata.realtime.channel(tableName).broadcast({
          event: 'mutation',
          type: 'DELETE',
        });
      } catch (err) {
        // broadcast failure is non-critical
      }
    },
  });
}
