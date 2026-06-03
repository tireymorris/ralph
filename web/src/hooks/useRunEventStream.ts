import { useEffect } from "react";
import type { EventEnvelope } from "../api/types";
import { entryFromEnvelope, type TimelineEntry } from "../lib/timeline";

export const EVENT_STREAM_MAX_RETRIES = 3;
export const EVENT_STREAM_RETRY_DELAY_MS = 2000;

type OpenEventStream = (id: string) => EventSource;

export function useRunEventStream(
  id: string | undefined,
  openStream: OpenEventStream,
  onEntry: (entry: TimelineEntry) => void,
  onEnvelope?: (envelope: EventEnvelope) => void,
  streamGeneration = 0,
): void {
  useEffect(() => {
    if (!id) return;

    let retryCount = 0;
    let retryTimer: ReturnType<typeof setTimeout> | undefined;
    let source: EventSource | null = null;
    let stopped = false;
    const seenIds = new Set<string>();

    function attachHandlers(stream: EventSource) {
      stream.onmessage = (ev: MessageEvent) => {
        try {
          const envelope = JSON.parse(ev.data as string) as EventEnvelope;
          onEnvelope?.(envelope);
          const entry = entryFromEnvelope(envelope);
          if (entry) {
            if (seenIds.has(entry.id)) return;
            seenIds.add(entry.id);
            onEntry(entry);
          }
        } catch {
          // ignore malformed SSE payloads
        }
      };

      stream.onerror = () => {
        stream.close();
        if (stopped || retryCount >= EVENT_STREAM_MAX_RETRIES) {
          return;
        }
        retryCount += 1;
        retryTimer = setTimeout(connect, EVENT_STREAM_RETRY_DELAY_MS);
      };
    }

    function connect() {
      source = openStream(id!);
      attachHandlers(source);
    }

    connect();

    return () => {
      stopped = true;
      if (retryTimer !== undefined) {
        clearTimeout(retryTimer);
      }
      source?.close();
    };
  }, [id, openStream, onEntry, onEnvelope, streamGeneration]);
}
