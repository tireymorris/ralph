import { useEffect, useRef, useState } from "react";
import { isTerminalRunStatus } from "../lib/format";
import {
  isRunAwaitingUser,
  isRunStalled,
  RUN_STALL_CHECK_INTERVAL_MS,
  RUN_STALL_THRESHOLD_MS,
} from "../lib/stall";

export function useRunStall(
  runStatus: string | undefined,
  clarifyQuestionCount: number,
  activityKey: string | number,
  enabled: boolean,
): boolean {
  const lastActivityRef = useRef(Date.now());
  const [stalled, setStalled] = useState(false);

  useEffect(() => {
    lastActivityRef.current = Date.now();
    setStalled(false);
  }, [activityKey]);

  useEffect(() => {
    if (!enabled || !runStatus || isTerminalRunStatus(runStatus)) {
      setStalled(false);
      return;
    }
    if (isRunAwaitingUser(runStatus, clarifyQuestionCount)) {
      setStalled(false);
      return;
    }

    function check() {
      setStalled(
        isRunStalled(
          lastActivityRef.current,
          Date.now(),
          RUN_STALL_THRESHOLD_MS,
        ),
      );
    }

    check();
    const timer = setInterval(check, RUN_STALL_CHECK_INTERVAL_MS);
    return () => clearInterval(timer);
  }, [enabled, runStatus, clarifyQuestionCount]);

  return stalled;
}
