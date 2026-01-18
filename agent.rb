# frozen_string_literal: true

require 'fileutils'
require 'open3'
require 'shellwords'
require_relative 'config'

module Ralph
  class Logger
    LOG_LEVELS = { debug: 0, info: 1, warn: 2, error: 3 }.freeze

    def self.configure(level = nil)
      @level = LOG_LEVELS[level&.to_sym] || LOG_LEVELS[Ralph::Config.get(:log_level)&.to_sym] || LOG_LEVELS[:info]
    end

    def self.log(level, message, context = {})
      return unless @level && LOG_LEVELS[level] >= @level

      timestamp = Time.now.strftime('%Y-%m-%d %H:%M:%S')
      {
        timestamp: timestamp,
        level: level.to_s.upcase,
        message: message,
        context: context
      }

      formatted = "[#{timestamp}] #{level.to_s.upcase}: #{message}"
      formatted += " | #{context}" unless context.empty?

      puts formatted
      write_to_file(formatted)
    rescue StandardError => e
      puts "❌ Logger error: #{e.message}"
    end

    def self.debug(message, context = {})
      log(:debug, message, context)
    end

    def self.info(message, context = {})
      log(:info, message, context)
    end

    def self.warn(message, context = {})
      log(:warn, message, context)
    end

    def self.error(message, context = {})
      log(:error, message, context)
    end

    def self.write_to_file(message)
      log_file = Ralph::Config.get(:log_file)
      File.open(log_file, 'a') { |f| f.puts(message) }
    end
  end

  class ErrorHandler
    def self.log_error(operation, error, context = {})
      Logger.error("Error in #{operation}", {
                     error_class: error.class.name,
                     error_message: error.message,
                     backtrace: error.backtrace&.first(3),
                     context: context
                   })
    end

    def self.with_error_handling(operation, context = {})
      result = yield
      Logger.debug("Completed #{operation}", context)
      result
    rescue StandardError => e
      log_error(operation, e, context)
      nil
    end

    def self.safe_system_command(command, operation, timeout_seconds = nil)
      Logger.debug("Executing command: #{command}", { operation: operation })

      timeout_seconds ||= Ralph::Config.get(:opencode_timeout)
      full_command = timeout_seconds ? "timeout #{timeout_seconds} #{command}" : command
      result = system(full_command)

      if result.nil?
        Logger.error('Command execution failed', { command: command, operation: operation })
        false
      elsif !result
        Logger.warn('Command returned non-zero', { command: command, operation: operation })
        false
      else
        Logger.debug('Command succeeded', { command: command, operation: operation })
        true
      end
    rescue StandardError => e
      Logger.error('System command exception', { command: command, operation: operation, error: e.message })
      false
    end

    def self.capture_command_output(prompt, operation, timeout_seconds = nil)
      Logger.debug("Capturing output for: #{prompt[0..100]}...", { operation: operation })

      timeout_seconds ||= Ralph::Config.get(:opencode_timeout)

      # Write prompt to file in current directory
      prompt_file = ".ralph_prompt_#{$$}.txt"
      begin
        File.write(prompt_file, prompt)

        cmd = "timeout #{timeout_seconds} bash -c 'cat #{prompt_file.shellescape} | opencode run --format default /dev/stdin' 2>&1"
        output = `#{cmd}`

        # Clean output - remove ANSI codes and JSON artifacts
        cleaned = clean_opencode_output(output)

        Logger.debug('Command output captured', { operation: operation, output_length: cleaned.length })
        cleaned
      ensure
        File.delete(prompt_file) if File.exist?(prompt_file)
      end
    rescue StandardError => e
      Logger.error('Command capture exception', { prompt: prompt[0..100], operation: operation, error: e.message })
      nil
    end

    def self.clean_opencode_output(output)
      # Remove ANSI color codes
      output.gsub(/\x1b\[[0-9;]*[a-zA-Z]/, '')
            # Remove JSON-like artifacts (lines starting with { or })
            .gsub(/^[{"].*$/m, '')
            # Clean up extra whitespace
            .gsub(/\n{3,}/, "\n\n")
            .strip
    end
  end

  class Agent
    def self.run(prompt, dry_run: false)
      # Initialize logger
      Logger.configure(:info)
      Logger.info('Starting Ralph', { prompt: prompt, dry_run: dry_run })

      ErrorHandler.with_error_handling('Directory change') do
        Dir.chdir(ENV['PWD'] || Dir.pwd)
      end

      puts '🤖 Ralph - Autonomous Software Development'
      puts "📝 Request: #{prompt}"
      puts "📁 Working in: #{Dir.pwd}"

      # Phase 1: Complete PRD and story analysis
      puts "\n📋 Phase 1: Creating PRD and analyzing project..."

      prd_prompt = "Task: Add simple scoring

Step 1: Read source files to understand the project.

Step 2: Create PRD.md with:
# PRD - Simple Scoring
## Branch: feature/scoring
## Stories: story-1, story-2

Step 3: Create story-1.md with:
# Story story-1: Score Component
## Priority: 1
## Description: Create Score component
## Acceptance Criteria:
- Score component added
- Score initializes to 0

Step 4: Create story-2.md with:
# Story story-2: Score Display
## Priority: 1
## Description: Display score in UI
## Acceptance Criteria:
- Score shown in game
- Score updates in real-time

Use Write tool. NO JSON. NO MARKDOWN CODE BLOCKS."

      success = ErrorHandler.with_error_handling('PRD creation') do
        Logger.info('Calling opencode to create files...')
        response = ErrorHandler.capture_command_output(prd_prompt, 'Generate PRD')
        Logger.info("Response: #{response ? response[0..200] : 'nil'}")

        sleep 2

        return false unless File.exist?('PRD.md')

        story_files = Dir.glob('story-*.md').sort
        if story_files.empty?
          Logger.error('No story files created')
          return false
        end

        Logger.info("Created PRD.md and #{story_files.length} story files")

        agents_content = "# Ralph Agent Patterns\n\n## Project Context\n- PRD: PRD.md\n- Stories: #{story_files.length} items\n- Started: #{Time.now.strftime('%Y-%m-%d %H:%M:%S')}\n\n"
        File.write('AGENTS.md', agents_content)

        true
      end

      unless success
        Logger.error('Failed to create PRD')
        return
      end

      # Phase 2: Autonomous implementation loop
      puts "\n🔄 Phase 2: Implementing all stories..."

      branch_name = extract_branch_name_from_prd
      create_feature_branch(branch_name)

      iteration = 0
      max_retries = Ralph::Config.get(:retry_attempts) || 3
      loop do
        iteration += 1
        retry_count = 0

        puts "\n" + '=' * 60
        puts "🔄 Iteration #{iteration}"
        puts '=' * 60

        # Find next incomplete story file
        story_file = find_next_story_file

        if story_file.nil?
          puts "\n🎉 All stories completed!"
          puts '<promise>COMPLETE</promise>'
          break
        end

        story_id = File.basename(story_file, '.md')
        puts "\n📖 Implementing: #{story_id}"

        # Implement with retry logic
        success = false
        max_retries.times do |attempt|
          puts "  Attempt #{attempt + 1}/#{max_retries}"

          if implement_story_from_file(story_file, iteration)
            success = true
            retry_count = attempt
            mark_story_complete(story_id)
            puts '✅ Story completed successfully'
            break
          else
            puts "❌ Attempt #{attempt + 1} failed"
            sleep 2 if attempt < max_retries - 1
          end
        end

        log_progress(iteration, story_file, success, retries: retry_count)

        puts '❌ Story failed after all retries - skipping to next story' unless success

        sleep 1
      end
    end

    def self.create_feature_branch(branch_name)
      puts "🌿 Creating branch: #{branch_name}"

      ErrorHandler.with_error_handling('Git repository validation') do
        raise StandardError, 'Not in a git repository' unless system('git rev-parse --git-dir > /dev/null 2>&1')
      end

      ErrorHandler.with_error_handling('Git branch creation', { branch: branch_name }) do
        if system("git show-ref --verify --quiet refs/heads/#{branch_name}")
          ErrorHandler.safe_system_command("git checkout #{branch_name}", 'Checkout existing branch')
        else
          ErrorHandler.safe_system_command("git checkout -b #{branch_name}", 'Create new branch')
        end
      end
    end

    def self.implement_story_from_file(story_file, iteration)
      story_content = ErrorHandler.with_error_handling('Reading story file', { file: story_file }) do
        File.read(story_file)
      end || ''

      context = ErrorHandler.with_error_handling('Reading AGENTS.md') do
        File.exist?('AGENTS.md') ? File.read('AGENTS.md') : ''
      end || ''

      implementation_prompt = <<~PROMPT
        #{story_content}

        Context:
        - Iteration: #{iteration}
        - Previous patterns: #{context || 'None yet'}
        - AGENTS.md exists: #{File.exist?('AGENTS.md')}

        Process:
        1. Read the story requirements above
        2. Implement complete solution
        3. Run tests if available
        4. Update AGENTS.md with new patterns discovered
        5. Commit changes with a descriptive message

        Work systematically. When complete, respond: "COMPLETED: [summary]"

        CRITICAL: Respond ONLY with the completion message, nothing else.
      PROMPT

      response = ErrorHandler.with_error_handling('Story implementation', { file: story_file }) do
        response = ErrorHandler.capture_command_output(implementation_prompt, "Implement #{story_file}")
        return nil unless response

        response.encode('UTF-8', 'binary', invalid: :replace, undef: :replace, replace: '').strip
      end

      unless response
        Logger.error('Story implementation failed', { file: story_file })
        return false
      end

      if response&.include?('COMPLETED:')
        puts "✓ #{response}"

        if run_tests
          log_progress(iteration, story_file, true)
          true
        else
          log_progress(iteration, story_file, false)
          false
        end
      else
        puts '❌ Implementation failed'
        log_progress(iteration, story_file, false)
        false
      end
    end

    def self.run_tests
      # Try to detect and run tests for this project
      test_commands = [
        'npm test',
        'yarn test',
        'pytest',
        'python -m pytest',
        'cargo test',
        'go test'
      ]

      test_commands.each do |cmd|
        next unless system("which #{cmd.split.first} > /dev/null 2>&1")

        print '🧪 Running tests... '
        result = system(cmd)
        puts result ? '✅' : '❌'
        return result
      end

      puts '⚠️ No test framework detected'
      true # Continue without tests
    end

    def self.commit_changes
      puts '💾 Committing changes...'

      ErrorHandler.with_error_handling('Git commit') do
        status_output = `git status --porcelain 2>/dev/null`
        if status_output.nil? || status_output.strip.empty?
          Logger.info('No changes to commit')
          return true
        end

        ErrorHandler.safe_system_command('git add .', 'Stage changes')
        ErrorHandler.safe_system_command("git commit -m 'Story implementation'", 'Commit changes')
      end
    end

    def self.log_progress(iteration, story_file, success, retries: 0)
      Logger.log(success ? :info : :error, "Iteration #{iteration} completed", {
                   story: File.basename(story_file),
                   success: success,
                   retries: retries
                 })

      ErrorHandler.with_error_handling('Progress logging') do
        log = [
          "## Iteration #{iteration} - #{Time.now.strftime('%Y-%m-%d %H:%M:%S')}",
          "Story: #{File.basename(story_file)}",
          "Status: #{success ? 'Success' : 'Failed'}",
          "Retries: #{retries}",
          '---'
        ].join("\n")

        File.open('progress.txt', 'a') { |f| f.puts(log + "\n") }
      end
    end

    def self.extract_branch_name_from_prd
      ErrorHandler.with_error_handling('Reading PRD') do
        prd_content = File.read('PRD.md')
        return ::Regexp.last_match(1).strip if prd_content =~ /## Project Branch\s*\n\s*([^\n]+)/

        'feature/implementation'
      end
    end

    def self.find_next_story_file
      # First check for .completed.md files to see what's already done
      completed = Dir.glob('story-*.completed.md').map { |f| File.basename(f, '.completed.md') }
      remaining = Dir.glob('story-*.md').sort.reject { |f| completed.include?(File.basename(f, '.md')) }

      remaining.find { |file| story_not_yet_committed?(file) }
    end

    def self.story_not_yet_committed?(story_file)
      story_id = File.basename(story_file, '.md')
      # Check if this story ID appears in recent git log
      log = `git log --oneline -20 2>/dev/null`
      !log.include?("story-#{story_id.gsub('story-', '')}:") && !log.include?(story_id)
    end

    def self.mark_story_complete(story_id)
      ErrorHandler.with_error_handling('Mark story complete', { story: story_id }) do
        original_file = "#{story_id}.md"
        completed_file = "#{story_id}.completed.md"

        if File.exist?(original_file)
          File.rename(original_file, completed_file)
          Logger.info("Marked #{story_id} as complete")
        end
      end
    end
  end
end
