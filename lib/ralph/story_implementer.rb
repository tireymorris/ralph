# frozen_string_literal: true

module Ralph
  class StoryImplementer
    class << self
      def implement(story, iteration, all_requirements)
        completed = all_requirements['stories'].count { |s| s['passes'] == true }
        total = all_requirements['stories'].length

        context = read_context_file

        implementation_prompt = build_implementation_prompt(story, iteration, completed, total, context)

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

      def read_context_file
        ErrorHandler.with_error_handling('Reading AGENTS.md') do
          agents_file = Ralph::Config.get(:agents_file)
          File.exist?(agents_file) ? File.read(agents_file) : ''
        end || ''
      end

      def build_implementation_prompt(story, iteration, completed, total, context)
        <<~PROMPT
          You are Ralph implementing story: #{story['title']}

          Story: #{story['description']}
          Acceptance Criteria: #{story['acceptance_criteria'].join(', ')}

          Context: Iteration #{iteration} (#{completed}/#{total} stories done)
          Previous patterns: #{context || 'None yet'}

          Process:
          1. Read existing code to understand patterns
          2. Implement complete solution
          3. Run tests and fix issues
          4. Update AGENTS.md with new patterns
          5. Commit changes

          Work systematically. When complete, respond: "COMPLETED: [summary]"

          CRITICAL: Respond ONLY with the completion message, nothing else.
        PROMPT
      end

      def process_implementation_response(story, iteration, response)
        if response&.include?('COMPLETED:')
          puts "✓ #{response}"

          test_success = TestRunner.run
          if test_success
            GitManager.commit_changes(story)
            ProgressLogger.log_iteration(iteration, story, true)
            true
          else
            ProgressLogger.log_iteration(iteration, story, false)
            false
          end
        else
          puts '❌ Implementation failed'
          ProgressLogger.log_iteration(iteration, story, false)
          false
        end
      end
    end
  end
end
