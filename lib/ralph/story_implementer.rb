# frozen_string_literal: true

module Ralph
  class StoryImplementer
    class << self
      def implement(story, iteration, all_requirements)
        completed = all_requirements['stories'].count { |s| s['passes'] == true }
        total = all_requirements['stories'].length

        implementation_prompt = build_implementation_prompt(story, iteration, completed, total)

        response = ErrorHandler.with_error_handling('Story implementation', { story: story['id'] }) do
          CommandRunner.capture_opencode_output(implementation_prompt, "Implement story: #{story['title']}")
        end

        unless response
          Logger.error('Story implementation failed', { story: story['id'] })
          return false
        end

        process_implementation_response(story, response)
      end

      private

      def build_implementation_prompt(story, iteration, completed, total)
        <<~PROMPT
          You are Ralph implementing story: #{story['title']}

          #{story_context(story)}

          Context: Iteration #{iteration} (#{completed}/#{total} stories done)

          #{implementation_process(story['id'])}

          #{critical_requirements}

          #{completion_format}
        PROMPT
      end

      def story_context(story)
        test_spec = story['test_spec'] || 'No test spec provided - create and run appropriate tests'
        <<~CONTEXT
          Story: #{story['description']}
          Acceptance Criteria: #{story['acceptance_criteria'].join(', ')}

          Test Spec Guidelines:
          #{test_spec}
        CONTEXT
      end

      def implementation_process(story_id)
        <<~PROCESS
          IMPLEMENTATION PROCESS:

          1. READ existing code to understand patterns and test setup
          2. IMPLEMENT the feature completely
          3. WRITE AN INTEGRATION TEST for this story:
             - Create/update test file: tests/#{story_id}.test.{js,ts,rb,py} (match project language)
             - Test MUST verify the feature works at RUNTIME, not just compilation
             - Use appropriate testing framework (Playwright, Puppeteer, Vitest, Jest, RSpec, pytest, etc.)
          4. RUN THE TEST and ensure it PASSES - do NOT proceed until tests pass
          5. RUN ALL PREVIOUS TESTS to ensure no regressions
          6. COMMIT changes including both implementation and test files
        PROCESS
      end

      def critical_requirements
        <<~REQUIREMENTS
          CRITICAL REQUIREMENTS:
          - You MUST write an actual test file, not just describe tests
          - You MUST run the test and see it pass in the output
          - Do NOT mark complete if you only ran lint/build - tests must pass
          - The test must verify RUNTIME behavior (e.g., app starts, UI renders, API responds)
        REQUIREMENTS
      end

      def completion_format
        <<~FORMAT
          When the integration test passes and changes are committed, respond:
          "COMPLETED: [summary] | TEST: [test file path] | RESULT: [pass/fail with brief output]"

          CRITICAL: Respond ONLY with the completion message, nothing else.
        FORMAT
      end

      def process_implementation_response(story, response)
        return handle_incomplete_response unless response&.include?('COMPLETED:')

        log_response(response)
        GitManager.commit_changes(story)
        true
      end

      def log_response(response)
        if response.include?('TEST:') && response.include?('RESULT:')
          puts "✓ #{response}"
        else
          puts "⚠️ #{response}"
          puts '⚠️ Warning: No test verification found in response, but marking as complete'
        end
      end

      def handle_incomplete_response
        puts '❌ Implementation did not complete'
        false
      end
    end
  end
end
