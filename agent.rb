# frozen_string_literal: true

require 'fileutils'
require 'open3'
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

      temp_file = Tempfile.new(['ralph_prompt', '.txt'])
      begin
        temp_file.write(prompt)
        temp_file.close

        ['timeout', timeout_seconds.to_s, 'opencode', 'run', "opencode run $(cat #{temp_file.path})"]

        cmd_parts = if timeout_seconds
                      ['timeout', timeout_seconds.to_s, 'opencode', 'run', "$(cat #{temp_file.path})"]
                    else
                      ['opencode', 'run', "$(cat #{temp_file.path})"]
                    end

        stdout, stderr, = Open3.capture3(*cmd_parts)
        output = stdout + stderr
      ensure
        temp_file.close
        temp_file.unlink
      end

      Logger.debug('Command output captured', { operation: operation, output_length: output.length })
      output
    rescue StandardError => e
      Logger.error('Command capture exception', { prompt: prompt[0..100], operation: operation, error: e.message })
      nil
    end
  end

  class Agent
    def self.run(prompt)
      # Initialize logger
      Logger.configure(:info)
      Logger.info('Starting Ralph', { prompt: prompt, working_dir: Dir.pwd })

      ErrorHandler.with_error_handling('Directory change') do
        Dir.chdir(ENV['PWD'] || Dir.pwd)
      end

      puts '🤖 Ralph - Autonomous Software Development'
      puts "📝 Request: #{prompt}"
      puts "📁 Working in: #{Dir.pwd}"

      # Phase 1: Complete PRD and story analysis
      puts "\n📋 Phase 1: Creating PRD and analyzing project..."

      prd_prompt = <<~PROMPT
        You are Ralph, an autonomous software development agent.

        Your task: #{prompt}

        Step 1: Analyze the current codebase
        - Read main files to understand the project
        - Identify the technology stack and patterns

        Step 2: Create PRD.md file with this structure:
        ```
        # PRD - [Project Name]

        ## Project Branch
        feature/[branch-name]

        ## User Stories
        - [List of story IDs]
        ```

        Step 3: Create individual story files (story-1.md, story-2.md, etc.)
        Each story file must contain:
        ```
        # Story [ID]: [Title]

        ## Description
        [Detailed description]

        ## Acceptance Criteria
        - [Criterion 1]
        - [Criterion 2]

        ## Priority
        [1-5]

        ## Implementation Notes
        [Any specific guidance for implementing this story]
        ```

        CRITICAL: You MUST create ALL files (PRD.md and all story-*.md files). Do not respond with text only - actually create the files.

        When complete, respond with: "DONE: Created [N] story files"
      PROMPT

      success = ErrorHandler.with_error_handling('PRD creation') do
        Logger.info('Calling opencode to create PRD and story files...')
        response = ErrorHandler.capture_command_output(prd_prompt, 'Generate PRD')
        Logger.info("Response received: #{response ? 'Yes' : 'No'}")

        unless response
          Logger.error('No response from opencode')
          return false
        end

        # Wait a moment for file system to sync
        sleep 2

        # Check if PRD.md was created
        unless File.exist?('PRD.md')
          Logger.error('PRD.md file was not created')
          Logger.error("Response was: #{response[0..500]}")
          return false
        end

        Logger.info('PRD.md created successfully')

        story_files = Dir.glob('story-*.md').sort
        if story_files.empty?
          Logger.error('No story files were created')
          return false
        end

        Logger.info("Created #{story_files.length} story files: #{story_files.join(', ')}")

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
      loop do
        iteration += 1

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

        # Implement story by feeding story file to opencode
        if implement_story_from_file(story_file, iteration)
          mark_story_complete(story_id)
          puts '✅ Story completed successfully'
        else
          puts '❌ Story failed - will retry in next iteration'
        end

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

    def self.log_progress(iteration, story_file, success)
      Logger.log(success ? :info : :error, "Iteration #{iteration} completed", {
                   story: File.basename(story_file),
                   success: success
                 })

      ErrorHandler.with_error_handling('Progress logging') do
        log = [
          "## Iteration #{iteration} - #{Time.now.strftime('%Y-%m-%d %H:%M:%S')}",
          "Story: #{File.basename(story_file)}",
          "Status: #{success ? 'Success' : 'Failed'}",
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
      stories_dir = Dir.glob('story-*.md').sort
      stories_dir.find { |file| !file.include?('.completed.') }
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
