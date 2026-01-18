# frozen_string_literal: true

require_relative '../ralph/config'
require_relative '../ralph/logger'
require_relative '../ralph/error_handler'
require_relative '../ralph/git_manager'
require_relative '../ralph/prd_generator'
require_relative '../ralph/test_runner'
require_relative '../ralph/story_implementer'
require_relative '../ralph/progress_logger'

module Ralph
  # Main autonomous agent implementation
  class Agent
    class << self
      def run(prompt, dry_run: false)
        initialize_environment
        Logger.info('Starting Ralph', { prompt: prompt, dry_run: dry_run })

        # Generate PRD
        requirements = PrdGenerator.generate(prompt)
        return unless requirements

        if dry_run
          puts '🎯 Dry run mode: PRD generated successfully'
          puts "📁 Files created: #{Ralph::Config.get(:prd_file)}, #{Ralph::Config.get(:agents_file)}"
          return
        end

        # Autonomous implementation loop
        run_implementation_loop(requirements)
      end

      private

      def initialize_environment
        ErrorHandler.with_error_handling('Directory change') do
          Dir.chdir(ENV['PWD'] || Dir.pwd)
        end

        puts '🤖 Ralph - Autonomous Software Development'
      end

      def run_implementation_loop(requirements)
        puts "\n🔄 Phase 2: Implementing all stories..."

        GitManager.create_branch(requirements['branch_name'])

        iteration = 0
        loop do
          iteration += 1

          puts "\n#{'=' * 60}"
          puts "🔄 Iteration #{iteration}"
          puts '=' * 60

          # Find next incomplete story
          next_story = requirements['stories'].find { |s| s['passes'] != true }

          if next_story.nil?
            puts "\n🎉 All stories completed!"
            puts '<promise>COMPLETE</promise>'
            break
          end

          puts "\n📖 Implementing: #{next_story['title']}"
          puts "🎯 Priority: #{next_story['priority']}"

          # Implement story
          if StoryImplementer.implement(next_story, iteration, requirements)
            next_story['passes'] = true
            ProgressLogger.update_state(requirements)
            puts '✅ Story completed successfully'
          else
            puts '❌ Story failed - will retry in next iteration'
          end
        end
      end
    end
  end
end
