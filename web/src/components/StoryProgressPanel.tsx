import type { PRDDocument, PRDStory } from "../api/types";

interface StoryProgressPanelProps {
  prd: PRDDocument;
  defaultOpen?: boolean;
}

function storyStatus(story: PRDStory): "done" | "pending" {
  return story.passes ? "done" : "pending";
}

export default function StoryProgressPanel({
  prd,
  defaultOpen = true,
}: StoryProgressPanelProps) {
  const completed = prd.stories.filter((s) => s.passes).length;
  const total = prd.stories.length;
  const currentIndex = prd.stories.findIndex((s) => !s.passes);

  return (
    <details className="story-progress-panel" open={defaultOpen}>
      <summary className="story-progress-summary">
        <span className="story-progress-title">Stories</span>
        <span className="story-progress-count">
          {completed}/{total} done
        </span>
      </summary>
      <ol className="story-progress-list">
        {prd.stories.map((story, i) => {
          const status = storyStatus(story);
          const isCurrent = i === currentIndex;
          const completedSlices = story.slices.filter((slice) => slice.passes).length;
          return (
            <li key={story.id}>
              <details
                className="story-progress-item"
                open={isCurrent}
              >
                <summary className="story-progress-item-summary">
                  <span
                    className={`story-progress-status story-progress-status--${status}${isCurrent ? " story-progress-status--current" : ""}`}
                    aria-hidden
                  />
                  <span className="story-progress-item-title">
                    <span className="story-progress-item-number">{i + 1}</span>
                    {story.title}
                  </span>
                </summary>
                <p className="story-progress-item-description">
                  {story.description}
                </p>
                <p className="story-progress-item-description">
                  {story.slices.length > 0
                    ? `${completedSlices}/${story.slices.length} slices done`
                    : "0/0 slices done"}
                </p>
                <ul className="story-progress-slice-list">
                  {story.slices.map((slice) => (
                    <li key={slice.id} className="story-progress-slice">
                      <p>
                        <strong>Behavior:</strong> {slice.behavior}
                      </p>
                      <p>
                        <strong>Red hint:</strong> {slice.red_hint}
                      </p>
                      {slice.refactor_hint ? (
                        <p>
                          <strong>Refactor hint:</strong>{" "}
                          {slice.refactor_hint}
                        </p>
                      ) : null}
                    </li>
                  ))}
                </ul>
              </details>
            </li>
          );
        })}
      </ol>
    </details>
  );
}
