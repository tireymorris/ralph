package prompt

func FollowUpRevision(userMessage, prdFile, runTranscript string) string {
	return mustRender("followup", FollowUpData{
		UserMessage:   userMessage,
		PRDFile:       prdFile,
		RunTranscript: runTranscript,
	})
}
