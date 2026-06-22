package runstate

type ReviewLoopState struct {
	Checkpoint                 string `json:"checkpoint,omitempty"`
	ReviewIteration            int    `json:"review_iteration,omitempty"`
	ReviewFingerprint          string `json:"review_fingerprint,omitempty"`
	ReviewElapsedMs            int64  `json:"review_elapsed_ms,omitempty"`
	StopReason                 string `json:"stop_reason,omitempty"`
	LastReviewTranscriptPath   string `json:"last_review_transcript_path,omitempty"`
	LastReviewChangedFilesHash string `json:"last_review_changed_files_hash,omitempty"`
	RecoveryAttempts           int    `json:"recovery_attempts,omitempty"`
}

func reviewLoopScalarsPresent(u ReviewLoopUpdate) bool {
	return u.ReviewIteration > 0 ||
		u.ReviewFingerprint != "" ||
		u.ReviewElapsedMs > 0 ||
		u.LastReviewTranscriptPath != "" ||
		u.LastReviewChangedFilesHash != ""
}

func ApplyReviewLoopUpdate(dst *ReviewLoopState, u ReviewLoopUpdate) {
	if u.Checkpoint != "" {
		dst.Checkpoint = u.Checkpoint
	}
	if reviewLoopScalarsPresent(u) {
		dst.ReviewIteration = u.ReviewIteration
		dst.ReviewFingerprint = u.ReviewFingerprint
		dst.ReviewElapsedMs = u.ReviewElapsedMs
		if u.LastReviewTranscriptPath != "" {
			dst.LastReviewTranscriptPath = u.LastReviewTranscriptPath
		}
		dst.LastReviewChangedFilesHash = u.LastReviewChangedFilesHash
		if u.StopReason != "" {
			dst.StopReason = u.StopReason
		} else {
			dst.StopReason = ""
		}
	} else if u.Checkpoint == CheckpointImplReview && u.ReviewFingerprint == "" && dst.ReviewFingerprint != "" {
		dst.ReviewFingerprint = ""
	} else if u.StopReason != "" {
		dst.StopReason = u.StopReason
	}
	if u.ClearRecoveryAttempts {
		dst.RecoveryAttempts = 0
	} else if u.RecoveryAttempts > 0 {
		dst.RecoveryAttempts = u.RecoveryAttempts
	}
}
