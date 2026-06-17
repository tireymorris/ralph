import type { PRDDocument, PRDStory } from "../api/types";

interface StoryProgressPanelProps {
  prd: PRDDocument;
  defaultOpen?: boolean;
}

function storyStatus(story: PRDStory): "done" | "pending" {
  return story.passes ? "done" : "pending";
}

function firstUnfinishedSliceIndex(story: PRDStory): number {
  return story.slices.findIndex((slice) => !slice.passes);
}

function completedSliceCount(story: PRDStory): number {
  if (story.passes) {
    return story.slices.length;
  }

  const firstUnfinishedIndex = firstUnfinishedSliceIndex(story);
  if (firstUnfinishedIndex === -1) {
    return story.slices.length;
  }

  return firstUnfinishedIndex;
}

function sliceStatus(
  story: PRDStory,
  sliceIndex: number,
  isCurrent: boolean,
): "completed" | "in progress" | "pending" {
  if (story.passes) {
    return "completed";
  }

  if (!isCurrent) {
    return "pending";
  }

  const firstUnfinishedIndex = firstUnfinishedSliceIndex(story);

  if (firstUnfinishedIndex === -1) {
    return "completed";
  }

  if (sliceIndex < firstUnfinishedIndex) {
    return "completed";
  }

  if (sliceIndex === firstUnfinishedIndex) {
    return "in progress";
  }

  return "pending";
}

export default function StoryProgressPanel({
  prd,
  defaultOpen = true,
}: StoryProgressPanelProps) {
  const completed = prd.stories.filter((s) => s.passes).length;
  const total = prd.stories.length;
  const currentIndex = prd.stories.findIndex((s) => !s.passes);

  return (
    <details className="panel panel--compact story-progress-panel" open={defaultOpen}>
      <summary className="story-progress-summary">
        <span className="story-progress-title">Stories</span>
        <span className="content-meta story-progress-count">
          {completed}/{total} done
        </span>
      </summary>
      <ol className="story-progress-list">
        {prd.stories.map((story, i) => {
          const status = storyStatus(story);
          const isCurrent = i === currentIndex;
          const completedSlices = completedSliceCount(story);
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
                  <span className="content-subheading story-progress-item-title">
                    <span className="content-meta story-progress-item-number">
                      {i + 1}
                    </span>
                    {story.title}
                  </span>
                </summary>
                <p className="content-body story-progress-item-description">
                  {story.description}
                </p>
                <p className="content-meta story-progress-slice-count">
                  {story.slices.length > 0
                    ? `${completedSlices}/${story.slices.length} slices done`
                    : "0/0 slices done"}
                </p>
                <ul className="slice-list story-progress-slice-list">
                  {story.slices.map((slice, sliceIndex) => {
                    const status = sliceStatus(story, sliceIndex, isCurrent);

                    return (
                      <li key={slice.id} className="slice-item story-progress-slice">
                        <p className="slice-item-status">
                          <strong>Status:</strong> {status}
                        </p>
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
                    );
                  })}
                </ul>
              </details>
            </li>
          );
        })}
      </ol>
    </details>
  );
}
