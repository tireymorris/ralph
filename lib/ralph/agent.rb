# frozen_string_literal: true

require_relative 'config'
require_relative 'logger'
require_relative 'error_handler'
require_relative 'command_runner'
require_relative 'json_parser'
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
        print_phase_header('PHASE 2: Autonomous Implementation Loop')

        total_stories = requirements['stories'].length
        iteration = 0
        max_iterations = Config.get(:max_iterations)
        max_retries = Config.get(:retry_attempts)

        loop do
          iteration += 1
          if iteration > max_iterations
            return handle_max_iterations(requirements, total_stories, iteration, max_iterations)
          end

          result = run_single_iteration(iteration, requirements, total_stories, max_retries)
          return CLI::EXIT_SUCCESS if result == :completed
          return CLI::EXIT_FAILURE if result == :all_failed
        end
      end

      def handle_max_iterations(requirements, total_stories, iteration, max_iterations)
        print_phase_header('MAX ITERATIONS REACHED', prefix: '⚠️')
        completed = requirements['stories'].count { |s| s['passes'] == true }
        puts "📊 Completed: #{completed}/#{total_stories} stories"
        puts "📝 Max Iterations: #{max_iterations}"
        Logger.error('Max iterations exceeded', { iteration: iteration, max: max_iterations })
        completed == total_stories ? CLI::EXIT_SUCCESS : CLI::EXIT_PARTIAL
      end

      def print_phase_header(title, prefix: '🚀')
        puts "\n#{'=' * 60}"
        puts "#{prefix} #{title}"
        puts '=' * 60
      end

      def run_single_iteration(iteration, requirements, total_stories, max_retries)
        puts "\n#{'=' * 60}"
        puts "🔄 ITERATION #{iteration} - #{Time.now.strftime('%H:%M:%S')}"
        puts '=' * 60

        # Find next story that hasn't passed and hasn't exceeded retry limit (sorted by priority)
        next_story = requirements['stories']
                     .select { |s| s['passes'] != true && (s['retry_count'] || 0) < max_retries }
                     .min_by { |s| s['priority'] || Float::INFINITY }

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
        retry_count = next_story['retry_count'] || 0

        print_story_info(next_story, completed_stories, total_stories, retry_count)

        puts "\n⚡ Starting implementation..."
        if StoryImplementer.implement(next_story, iteration, requirements)
          handle_story_success(next_story, retry_count, requirements, completed_stories, total_stories)
        else
          handle_story_failure(next_story, retry_count, requirements)
        end

        :continue
      end

      def print_story_info(story, completed, total, retry_count)
        progress_percentage = (completed.to_f / total * 100).round(1)
        description = story['description'][0..80]
        description += '...' if story['description'].length > 80

        puts "\n📈 Progress: #{completed}/#{total} stories (#{progress_percentage}%)"
        puts "\n📖 Current Story: #{story['title']}"
        puts "🎯 Priority: #{story['priority']}"
        puts "🔄 Attempt: #{retry_count + 1}/#{Config.get(:retry_attempts)}"
        puts "📝 Description: #{description}"
      end

      def handle_story_success(story, retry_count, requirements, completed_stories, total_stories)
        story['passes'] = true
        story['retry_count'] = retry_count
        ProgressLogger.update_state(requirements)
        puts "\n✅ Story completed successfully!"
        puts "📊 Progress: #{completed_stories + 1}/#{total_stories} stories"
      end

      def handle_story_failure(story, retry_count, requirements)
        story['retry_count'] = retry_count + 1
        ProgressLogger.update_state(requirements)
        remaining = Config.get(:retry_attempts) - story['retry_count']
        puts "\n❌ Story failed - #{remaining} retries remaining"
        puts '⏳ Waiting before retry...'
        sleep Config.get(:retry_delay)
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
