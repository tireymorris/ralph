# frozen_string_literal: true

require_relative 'config'
require_relative 'logger'
require_relative 'error_handler'
require_relative 'git_manager'
require_relative 'prd_generator'
require_relative 'story_implementer'
require_relative 'progress_logger'

module Ralph
  class Agent
    WORKING_FILES = %w[prd.json].freeze

    class << self
      def run(prompt, dry_run: false)
        initialize_environment
        Logger.info('Starting Ralph', { prompt: prompt, dry_run: dry_run })

        requirements = PrdGenerator.generate(prompt)
        return CLI::EXIT_FAILURE unless requirements

        if dry_run
          puts '🎯 Dry run mode: PRD generated successfully'
          puts "📁 Files created: #{Ralph::Config.get(:prd_file)}"
          return CLI::EXIT_SUCCESS
        end

        setup_git_branch(requirements)
        run_implementation_loop(requirements)
      end

      def resume
        initialize_environment
        Logger.info('Resuming Ralph from existing PRD')

        requirements = load_existing_prd
        return CLI::EXIT_FAILURE unless requirements

        completed = requirements['stories'].count { |s| s['passes'] == true }
        total = requirements['stories'].length

        puts "\n📊 PRD loaded: #{requirements['project_name']}"
        puts "📈 Progress: #{completed}/#{total} stories already completed"

        if completed == total
          puts "\n✅ All stories already completed!"
          cleanup_working_files
          return CLI::EXIT_SUCCESS
        end

        run_implementation_loop(requirements)
      end

      private

      def initialize_environment
        Logger.configure

        puts "\n#{'=' * 60}"
        puts '🤖 RALPH - Autonomous Software Development Agent'
        puts '=' * 60

        ErrorHandler.with_error_handling('Directory change') do
          Dir.chdir(ENV['PWD'] || Dir.pwd)
        end

        puts "📍 Working Directory: #{Dir.pwd}"
        puts "⏰ Started: #{Time.now.strftime('%Y-%m-%d %H:%M:%S')}"
      end

      def load_existing_prd
        prd_file = Config.get(:prd_file)

        ErrorHandler.with_error_handling('Load existing PRD') do
          content = File.read(prd_file)
          requirements = JSON.parse(content)

          # Initialize retry counts if not present
          requirements['stories'].each do |story|
            story['retry_count'] ||= 0
          end

          requirements
        end
      end

      def setup_git_branch(requirements)
        branch_name = requirements['branch_name']
        return unless branch_name

        puts "\n🌿 Setting up git branch: #{branch_name}"
        GitManager.create_branch(branch_name)
      end

      def run_implementation_loop(requirements)
        puts "\n#{'=' * 60}"
        puts '🚀 PHASE 2: Autonomous Implementation Loop'
        puts '=' * 60

        total_stories = requirements['stories'].length
        iteration = 0
        max_iterations = Config.get(:max_iterations)
        max_retries = Config.get(:retry_attempts)

        loop do
          iteration += 1

          if iteration > max_iterations
            puts "\n#{'=' * 60}"
            puts '⚠️ MAX ITERATIONS REACHED'
            puts '=' * 60
            completed = requirements['stories'].count { |s| s['passes'] == true }
            puts "📊 Completed: #{completed}/#{total_stories} stories"
            puts "📝 Max Iterations: #{max_iterations}"
            Logger.error('Max iterations exceeded', { iteration: iteration, max: max_iterations })
            return completed == total_stories ? CLI::EXIT_SUCCESS : CLI::EXIT_PARTIAL
          end

          result = run_single_iteration(iteration, requirements, total_stories, max_retries)
          case result
          when :completed
            return CLI::EXIT_SUCCESS
          when :all_failed
            return CLI::EXIT_FAILURE
          end
        end
      end

      def run_single_iteration(iteration, requirements, total_stories, max_retries)
        puts "\n#{'=' * 60}"
        puts "🔄 ITERATION #{iteration} - #{Time.now.strftime('%H:%M:%S')}"
        puts '=' * 60

        # Find next story that hasn't passed and hasn't exceeded retry limit
        next_story = requirements['stories'].find do |s|
          s['passes'] != true && (s['retry_count'] || 0) < max_retries
        end

        # Check if all remaining stories have exceeded retries
        if next_story.nil?
          failed_stories = requirements['stories'].reject { |s| s['passes'] == true }

          return handle_all_completed(iteration, total_stories, requirements) if failed_stories.empty?

          return handle_all_failed(failed_stories, max_retries, requirements)

        end

        run_story_implementation(next_story, iteration, requirements, total_stories)
      end

      def handle_all_completed(iteration, total_stories, _requirements)
        puts "\n#{'=' * 60}"
        puts '🎉 ALL STORIES COMPLETED!'
        puts '=' * 60
        puts "📊 Total Stories: #{total_stories}"
        puts "📝 Total Iterations: #{iteration}"
        puts "⏰ Completed: #{Time.now.strftime('%Y-%m-%d %H:%M:%S')}"
        cleanup_working_files
        puts '<promise>COMPLETE</promise>'
        :completed
      end

      def handle_all_failed(failed_stories, max_retries, requirements)
        puts "\n#{'=' * 60}"
        puts '❌ IMPLEMENTATION STOPPED - Stories exceeded max retries'
        puts '=' * 60
        puts "📊 Failed stories (#{failed_stories.length}):"
        failed_stories.each do |s|
          puts "  - #{s['title']} (#{s['retry_count']}/#{max_retries} attempts)"
        end
        puts "\n💡 Fix the issues and run: ./bin/ralph --resume"
        ProgressLogger.update_state(requirements)
        :all_failed
      end

      def run_story_implementation(next_story, iteration, requirements, total_stories)
        completed_stories = requirements['stories'].count { |s| s['passes'] == true }
        progress_percentage = (completed_stories.to_f / total_stories * 100).round(1)
        retry_count = next_story['retry_count'] || 0

        puts "\n📈 Progress: #{completed_stories}/#{total_stories} stories (#{progress_percentage}%)"
        puts "\n📖 Current Story: #{next_story['title']}"
        puts "🎯 Priority: #{next_story['priority']}"
        puts "🔄 Attempt: #{retry_count + 1}/#{Config.get(:retry_attempts)}"
        puts "📝 Description: #{next_story['description'][0..80]}#{'...' if next_story['description'].length > 80}"

        puts "\n⚡ Starting implementation..."
        if StoryImplementer.implement(next_story, iteration, requirements)
          next_story['passes'] = true
          next_story['retry_count'] = retry_count
          ProgressLogger.update_state(requirements)
          puts "\n✅ Story completed successfully!"
          puts "📊 Progress: #{completed_stories + 1}/#{total_stories} stories"
        else
          next_story['retry_count'] = retry_count + 1
          ProgressLogger.update_state(requirements)
          remaining = Config.get(:retry_attempts) - next_story['retry_count']
          puts "\n❌ Story failed - #{remaining} retries remaining"
          puts '⏳ Waiting before retry...'
          sleep Config.get(:retry_delay)
        end

        :continue
      end

      def cleanup_working_files
        puts "\n🧹 Cleaning up working files..."
        WORKING_FILES.each do |file|
          if File.exist?(file)
            File.delete(file)
            puts "  🗑️  Deleted #{file}"
          end
        end
      end
    end
  end
end
