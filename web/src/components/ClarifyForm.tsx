import { useState } from "react";
import type { QuestionAnswer } from "../api/types";

interface ClarifyFormProps {
  questions: string[];
  onSubmit: (answers: QuestionAnswer[]) => void;
  submitting?: boolean;
}

export default function ClarifyForm({
  questions,
  onSubmit,
  submitting = false,
}: ClarifyFormProps) {
  const [answers, setAnswers] = useState<string[]>(() =>
    questions.map(() => ""),
  );

  const allFilled =
    questions.length > 0 &&
    questions.every((_, i) => answers[i]?.trim().length > 0);

  function handleChange(index: number, value: string) {
    setAnswers((prev) => {
      const next = [...prev];
      next[index] = value;
      return next;
    });
  }

  function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    if (!allFilled || submitting) return;
    onSubmit(
      questions.map((question, i) => ({
        question,
        answer: answers[i].trim(),
      })),
    );
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if ((e.metaKey || e.ctrlKey) && e.key === "Enter") {
      e.preventDefault();
      const form = (e.target as HTMLElement).closest("form");
      form?.requestSubmit();
    }
  }

  return (
    <form className="clarify-form" onSubmit={handleSubmit}>
      <h2 className="content-heading clarify-form-heading">
        A few questions first
      </h2>
      {questions.map((question, i) => (
        <label key={question} className="field">
          <span className="field-label">{question}</span>
          <textarea
            className="composer-input"
            aria-label={question}
            value={answers[i] ?? ""}
            onChange={(e) => handleChange(i, e.target.value)}
            onKeyDown={handleKeyDown}
            disabled={submitting}
            rows={2}
          />
        </label>
      ))}
      <button
        type="submit"
        className="btn btn--primary"
        disabled={!allFilled || submitting}
      >
        Submit answers
      </button>
    </form>
  );
}
