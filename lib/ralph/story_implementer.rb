# frozen_string_literal: true

module Ralph
  class StoryImplementer
    class << self
      def implement(story, iteration, all_requirements)
        completed = all_requirements['stories'].count { |s| s['passes'] == true }
        total = all_requirements['stories'].length

        implementation_prompt = build_implementation_prompt(story, iteration, completed, total)

        response = ErrorHandler.with_error_handling('Story implementation', { story: story['id'] }) do
          ErrorHandler.capture_command_output(implementation_prompt, "Implement story: #{story['title']}")
        end

        unless response
          Logger.error('Story implementation failed', { story: story['id'] })
          return false
        end

        process_implementation_response(story, iteration, response)
      end

      private

      def build_implementation_prompt(story, iteration, completed, total)
        test_spec = story['test_spec'] || 'No test spec provided - create and run appropriate tests'

        <<~PROMPT
          You are Ralph implementing story: #{story['title']}

          Story: #{story['description']}
          Acceptance Criteria: #{story['acceptance_criteria'].join(', ')}

          VALIDATION TEST SPEC (MUST PASS):
          #{test_spec}

          Context: Iteration #{iteration} (#{completed}/#{total} stories done)

          Process:
          1. Read existing code to understand patterns
          2. Implement complete solution
          3. CRITICAL: Run the VALIDATION TEST SPEC steps to verify the feature actually works at RUNTIME
             - Start the dev server if needed (npm run dev)
             - Use browser automation, curl, or manual verification commands
             - Check for runtime errors, not just compilation
             - Verify the feature behaves correctly, not just that code exists
          4. Fix any runtime issues discovered during validation
          5. Commit changes with descriptive message

          IMPORTANT:#{' '}
          - Do NOT mark as complete if you only ran lint/build checks
          - You MUST verify the feature works at runtime per the test spec
          - If the test spec requires UI verification, start the dev server and test it
          - If API calls are involved, verify they return expected data
          - Check browser console for runtime errors

          Work systematically. When the VALIDATION TEST SPEC passes and changes are committed, respond: "COMPLETED: [summary of what was validated]"

          CRITICAL: Respond ONLY with the completion message, nothing else.
        PROMPT
      end

      def process_implementation_response(story, _iteration, response)
        if response&.include?('COMPLETED:')
          puts "✓ #{response}"
          GitManager.commit_changes(story)
          true
        else
          puts '❌ Implementation did not complete'
          false
        end
      end
    end
  end
end
