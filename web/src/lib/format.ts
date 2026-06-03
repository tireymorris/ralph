const statusLabels: Record<string, string> = {
  running: "Running",
  waiting_clarify: "Needs Answers",
  waiting_review: "Needs Review",
  completed: "Completed",
  failed: "Failed",
  cancelled: "Cancelled",
};

export function formatStatus(status: string): string {
  return statusLabels[status] ?? status;
}

export function statusBadgeClass(status: string): string {
  switch (status) {
    case "running":
      return "run-status-badge--running";
    case "waiting_clarify":
    case "waiting_review":
      return "run-status-badge--waiting";
    case "completed":
      return "run-status-badge--completed";
    case "failed":
      return "run-status-badge--failed";
    case "cancelled":
      return "run-status-badge--cancelled";
    default:
      return "run-status-badge--default";
  }
}

export function relativeTime(iso: string): string {
  const seconds = Math.floor(
    (Date.now() - new Date(iso).getTime()) / 1000,
  );
  if (seconds < 5) return "just now";
  if (seconds < 60) return `${seconds}s ago`;
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}
