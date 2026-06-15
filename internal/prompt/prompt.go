package prompt

// QuestionAnswer holds a clarifying question and the user's answer.
type QuestionAnswer struct {
	Question string
	Answer   string
}

// ClarifyingQuestions tells the AI to write question strings to questionsFile and stop.
func ClarifyingQuestions(userPrompt, questionsFile string, isEmptyCodebase bool) string {
	codebaseNote := "an existing codebase"
	if isEmptyCodebase {
		codebaseNote = "a new project (no existing source code)"
	}
	return mustRender("clarify", ClarifyData{
		CodebaseNote:  codebaseNote,
		UserPrompt:    userPrompt,
		QuestionsFile: questionsFile,
	})
}

func PRDGeneration(userPrompt, prdFile, branchPrefix string, isEmptyCodebase bool) string {
	return PRDGenerationWithAnswers(userPrompt, prdFile, branchPrefix, isEmptyCodebase, nil)
}

func PRDGenerationWithAnswers(userPrompt, prdFile, branchPrefix string, isEmptyCodebase bool, qas []QuestionAnswer) string {
	contextGuidance := "- context: describe ONLY the tech stack and patterns you ACTUALLY observe in the codebase"
	if isEmptyCodebase {
		contextGuidance = `Note: The working directory has no existing source code. This is a new project.
- context: describe ONLY the tech stack specified in the user's request, or state "New project - no existing codebase"
- Do NOT assume or invent a tech stack the user did not mention`
	}
	return mustRender("prd-generate", PRDGenerateData{
		UserPrompt:      userPrompt,
		PRDFile:         prdFile,
		BranchPrefix:    branchPrefix,
		ContextGuidance: contextGuidance,
		Clarifications:  qas,
	})
}

func PRDCritiqueRevision(userPrompt, prdFile, critique string) string {
	return mustRender("prd-critique-revision", PRDCritiqueRevisionData{
		UserPrompt: userPrompt,
		PRDFile:    prdFile,
		Critique:   critique,
	})
}

func PRDClarificationRevision(userPrompt, prdFile string, qas []QuestionAnswer) string {
	return mustRender("prd-clarification-revision", PRDClarificationRevisionData{
		UserPrompt:     userPrompt,
		PRDFile:        prdFile,
		Clarifications: qas,
	})
}

func Cleanup(codebaseContext, prdFile string, changedFiles []string) string {
	return mustRender("cleanup", CleanupData{
		Context:      codebaseContext,
		PRDFile:      prdFile,
		ChangedFiles: changedFiles,
	})
}

func StoryImplementation(storyID, title, description string, slices []SliceData, featureTestSpec, codebaseContext, prdFile string, completed, total int, dependsOn []string) string {
	pending := make([]SliceData, 0, len(slices))
	for _, slice := range slices {
		if slice.Passes {
			continue
		}
		pending = append(pending, slice)
	}
	return mustRender("story-implement", StoryImplementData{
		StoryID:         storyID,
		Title:           title,
		Description:     description,
		Slices:          pending,
		FeatureTestSpec: featureTestSpec,
		Context:         codebaseContext,
		PRDFile:         prdFile,
		Completed:       completed,
		Total:           total,
		DependsOn:       dependsOn,
	})
}
