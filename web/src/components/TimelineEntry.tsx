import { useState } from "react";

const MAX_DISPLAY_CHARS = 5000;

export interface TimelineEntryProps {
  variant: "assistant" | "system" | "error";
  text: string;
}

const variantAriaLabel: Record<TimelineEntryProps["variant"], string> = {
  assistant: "Assistant message",
  system: "System message",
  error: "Error message",
};

export default function TimelineEntryBubble({
  variant,
  text,
}: TimelineEntryProps) {
  const [expanded, setExpanded] = useState(false);
  const truncated = text.length > MAX_DISPLAY_CHARS;
  const displayText =
    truncated && !expanded ? text.slice(0, MAX_DISPLAY_CHARS) + "…" : text;

  return (
    <li
      className={`timeline-entry timeline-entry--${variant}`}
      aria-label={variantAriaLabel[variant]}
    >
      <div className="timeline-entry-body">
        <pre className="timeline-entry-text">{displayText}</pre>
        {truncated && (
          <button
            type="button"
            className="timeline-show-more"
            onClick={() => setExpanded(!expanded)}
          >
            {expanded ? "Show less" : "Show more"}
          </button>
        )}
      </div>
    </li>
  );
}
