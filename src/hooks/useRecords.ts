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
  const userId = user?.id || 'anonymous';
  const queryKey = ['records', tableName, userId];

  return useQuery<DatabaseRecord[]>({
    queryKey,
    queryFn: async () => {
      if (!user) return getLocalRecords(tableName, 'anonymous');
      try {
        return await withAuthRetry(async () => {
          const result = await strata
            .from<DatabaseRecord[]>(tableName)
            .select('*')
            .eq('user_id', user.id)
            .order('created_at', 'desc')
            .execute();
          return result;
        });
      } catch (err) {
        console.warn('API unavailable, using local storage:', err);
        return getLocalRecords(tableName, user.id);
      }
    },
    enabled: true,
  });
}

export function useCreateRecord(tableName: string) {
  const queryClient = useQueryClient();
  const { user } = useAuth();

  return useMutation({
    mutationFn: async (record: Omit<DatabaseRecord, 'id' | 'user_id' | 'created_at'>) => {
      const userId = user?.id || 'anonymous';
      try {
        return await withAuthRetry(async () => {
          return await strata
            .from(tableName)
            .insert({
              ...record,
              user_id: user?.id || userId,
            })
            .execute();
        });
      } catch (err) {
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
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['records', tableName, user?.id || 'anonymous'] });
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

  return useMutation({
    mutationFn: async ({ id, ...updates }: Partial<DatabaseRecord> & { id: number }) => {
      const userId = user?.id || 'anonymous';
      try {
        return await withAuthRetry(async () => {
          return await strata
            .from(tableName)
            .update(updates)
            .eq('id', id)
            .eq('user_id', user?.id || userId)
            .execute();
        });
      } catch (err) {
        console.warn('API unavailable, updating locally:', err);
        const existing = getLocalRecords(tableName, userId);
        const updated = existing.map((r) =>
          r.id === id ? { ...r, ...updates } : r
        );
        setLocalRecords(tableName, userId, updated);
        return { id, ...updates };
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['records', tableName, user?.id || 'anonymous'] });
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

  return useMutation({
    mutationFn: async (id: number) => {
      const userId = user?.id || 'anonymous';
      try {
        return await withAuthRetry(async () => {
          return await strata
            .from(tableName)
            .delete()
            .eq('id', id)
            .eq('user_id', user?.id || userId)
            .execute();
        });
      } catch (err) {
        console.warn('API unavailable, deleting locally:', err);
        const existing = getLocalRecords(tableName, userId);
        setLocalRecords(tableName, userId, existing.filter((r) => r.id !== id));
        return { success: true };
      }
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['records', tableName, user?.id || 'anonymous'] });
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
