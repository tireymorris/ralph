import { useId, useState } from "react";

interface RunPromptProps {
  prompt: string;
}

export default function RunPrompt({ prompt }: RunPromptProps) {
  const [expanded, setExpanded] = useState(false);
  const contentId = useId();
  const canToggle = prompt.length > 120 || prompt.includes("\n");

  return (
    <div className={`run-detail-prompt-wrap${expanded ? " run-detail-prompt-wrap--expanded" : ""}`}>
      <p
        id={contentId}
        className="run-detail-prompt"
        title={!expanded ? prompt : undefined}
      >
        {prompt}
      </p>
      {canToggle && (
        <button
          type="button"
          className="run-detail-prompt-toggle"
          aria-expanded={expanded}
          aria-controls={contentId}
          onClick={() => setExpanded((v) => !v)}
        >
          {expanded ? "Show less" : "Show full prompt"}
        </button>
      )}
    </div>
  );
}
