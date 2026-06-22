package constants

import "time"

const MaxRecoveryAttempts = 2

const MaxImplementationReviewRounds = 8

const MaxCleanupRounds = 3

const RunnerRecoveryCooldown = 750 * time.Millisecond

const MinRunnerInvokeDuration = 500 * time.Millisecond

const RunnerFastFailRetryDelay = 1 * time.Second
