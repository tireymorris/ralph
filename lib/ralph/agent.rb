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
        puts "\n" + '=' * 60
        puts '🤖 RALPH - Autonomous Software Development Agent'
        puts '=' * 60

        ErrorHandler.with_error_handling('Directory change') do
          Dir.chdir(ENV['PWD'] || Dir.pwd)
        end

        puts "📍 Working Directory: #{Dir.pwd}"
        puts "⏰ Started: #{Time.now.strftime('%Y-%m-%d %H:%M:%S')}"
      end

      def run_implementation_loop(requirements)
        puts "\n" + '=' * 60
        puts '🚀 PHASE 2: Autonomous Implementation Loop'
        puts '=' * 60

        total_stories = requirements['stories'].length
        completed_stories = 0

        iteration = 0
        loop do
          iteration += 1

          puts "\n#{'=' * 60}"
          puts "🔄 ITERATION #{iteration} - #{Time.now.strftime('%H:%M:%S')}"
          puts '=' * 60

          # Find next incomplete story
          next_story = requirements['stories'].find { |s| s['passes'] != true }

          if next_story.nil?
            puts "\n" + '=' * 60
            puts '🎉 ALL STORIES COMPLETED!'
            puts '=' * 60
            puts "📊 Total Stories: #{total_stories}"
            puts "📝 Total Iterations: #{iteration}"
            puts "⏰ Completed: #{Time.now.strftime('%Y-%m-%d %H:%M:%S')}"
            puts '<promise>COMPLETE</promise>'
            break
          end

          completed_stories = requirements['stories'].count { |s| s['passes'] == true }
          progress_percentage = (completed_stories.to_f / total_stories * 100).round(1)

          puts "\n📈 Progress: #{completed_stories}/#{total_stories} stories (#{progress_percentage}%)"
          puts "\n📖 Current Story: #{next_story['title']}"
          puts "🎯 Priority: #{next_story['priority']}"
          puts "📝 Description: #{next_story['description'][0..80]}#{'...' if next_story['description'].length > 80}"

          puts "\n⚡ Starting implementation..."
          # Implement story
          if StoryImplementer.implement(next_story, iteration, requirements)
            next_story['passes'] = true
            ProgressLogger.update_state(requirements)
            puts "\n✅ Story completed successfully!"
            puts "📊 Progress: #{completed_stories + 1}/#{total_stories} stories"
          else
            puts "\n❌ Story failed - will retry in next iteration"
            puts '⏳ Waiting before retry...'
            sleep 2
          end
        end
      end
    end
  end
end
