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
        <<~PROMPT
          You are Ralph implementing story: #{story['title']}

          Story: #{story['description']}
          Acceptance Criteria: #{story['acceptance_criteria'].join(', ')}

          Context: Iteration #{iteration} (#{completed}/#{total} stories done)

          Process:
          1. Read existing code to understand patterns
          2. Implement complete solution
          3. Run tests and fix any issues
          4. Commit changes with descriptive message

          IMPORTANT: You are responsible for running tests and ensuring they pass before completing.

          Work systematically. When ALL tests pass and changes are committed, respond: "COMPLETED: [summary]"

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
