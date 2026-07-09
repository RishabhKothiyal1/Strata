import { useEffect } from 'react';
import { useQueryClient } from '@tanstack/react-query';
import { strata } from '../lib/strata';

export function useRealtimeTable(tableName: string, queryKey: any[]) {
  const queryClient = useQueryClient();

  const queryKeyString = JSON.stringify(queryKey);

  useEffect(() => {
    console.log(`Subscribing to realtime channel: ${tableName}`);
    const channel = strata.realtime.channel(tableName);
    
    const subscription = channel.subscribe((payload) => {
      console.log(`Realtime event received on channel ${tableName}:`, payload);
      // Invalidate the query key so React Query refetches the latest data
      queryClient.invalidateQueries({ queryKey: JSON.parse(queryKeyString) });
    });

    return () => {
      console.log(`Unsubscribing from realtime channel: ${tableName}`);
      subscription.unsubscribe();
    };
  }, [tableName, queryClient, queryKeyString]);
}
