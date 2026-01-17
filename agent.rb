# frozen_string_literal: true

require 'json'
require 'fileutils'
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
        You are Ralph, an autonomous software development agent. Your task is to implement: #{prompt}

        Follow this process:

        1. PROJECT ANALYSIS
           - Scan current directory to understand existing codebase
           - Identify technology stack, patterns, conventions
           - Note dependencies, tests, build setup
        #{'   '}
        2. CREATE PRD
           - Generate comprehensive user stories
           - Each story must be implementable in one iteration
           - Include acceptance criteria and priorities (1=highest)
        #{'   '}
        3. OUTPUT REQUIREMENTS
           - Respond ONLY with raw JSON (no markdown, no explanation)
        #{'   '}
        Required JSON format:
        {
          "project_name": "descriptive project name",
          "branch_name": "feature/descriptive-name",#{' '}
          "stories": [
            {
              "id": "story-1",
              "title": "Story title",
              "description": "Detailed description",
              "acceptance_criteria": ["criterion 1", "criterion 2"],
              "priority": 1,
              "passes": false
            }
          ]
        }

        CRITICAL: Return only the JSON object, nothing else.
      PROMPT

      success = ErrorHandler.with_error_handling('PRD creation') do
        response = ErrorHandler.safe_system_command("opencode run \"#{prd_prompt}\" 2>/dev/null", 'Generate PRD')
        return false unless response

        response = response.encode('UTF-8', 'binary', invalid: :replace, undef: :replace, replace: '').strip
        Logger.debug('OpenCode response received', { length: response.length })

        requirements = parse_json_safely(response, 'PRD requirements')
        return false unless requirements

        # Validate required structure
        required_fields = %w[project_name branch_name stories]
        missing_fields = required_fields.select { |field| requirements[field].nil? || requirements[field].empty? }
        raise ArgumentError, "Missing required fields: #{missing_fields.join(', ')}" if missing_fields.any?

        unless requirements['stories'].is_a?(Array) && requirements['stories'].any?
          raise ArgumentError, 'Invalid stories format: expected non-empty array'
        end

        # Create state files
        ErrorHandler.with_error_handling('State file creation') do
          File.write('prd.json', JSON.pretty_generate(requirements))

          agents_content = "# Ralph Agent Patterns\n\n## Project Context\n- Technology: #{requirements['project_name']}\n- Stories: #{requirements['stories'].length} items\n- Started: #{Time.now.strftime('%Y-%m-%d %H:%M:%S')}\n\n"
          File.write('AGENTS.md', agents_content)
        end

        Logger.info('PRD analysis complete', {
                      project: requirements['project_name'],
                      stories: requirements['stories'].length
                    })

        requirements
      end

      unless success
        Logger.error('Failed to create PRD')
        return
      end

      requirements = success

      # Phase 2: Autonomous implementation loop
      puts "\n🔄 Phase 2: Implementing all stories..."

      create_feature_branch(requirements['branch_name'])

      iteration = 0
      loop do
        iteration += 1

        puts "\n" + '=' * 60
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
        if implement_story(next_story, iteration, requirements)
          next_story['passes'] = true
          update_state(requirements)
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
        # Check if branch already exists
        if system("git show-ref --verify --quiet refs/heads/#{branch_name}")
          # Checkout existing branch
          ErrorHandler.safe_system_command("git checkout #{branch_name}", 'Checkout existing branch')
        else
          # Create new branch
          ErrorHandler.safe_system_command("git checkout -b #{branch_name}", 'Create new branch')
        end
      end
    end

    def self.implement_story(story, iteration, all_requirements)
      completed = all_requirements['stories'].count { |s| s['passes'] == true }
      total = all_requirements['stories'].length

      context = ErrorHandler.with_error_handling('Reading AGENTS.md') do
        File.exist?('AGENTS.md') ? File.read('AGENTS.md') : ''
      end || ''

      implementation_prompt = <<~PROMPT
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

      response = ErrorHandler.with_error_handling('Story implementation', { story: story['id'] }) do
        success = ErrorHandler.safe_system_command("opencode run \"#{implementation_prompt}\" 2>/dev/null",
                                                   "Implement story: #{story['title']}")
        return nil unless success

        response = `opencode run "#{implementation_prompt}" 2>/dev/null`
        response.encode('UTF-8', 'binary', invalid: :replace, undef: :replace, replace: '').strip
      end

      unless response
        Logger.error('Story implementation failed', { story: story['id'] })
        return false
      end

      if response&.include?('COMPLETED:')
        puts "✓ #{response}"

        # Run tests if available
        if run_tests
          commit_changes(story)
          log_progress(iteration, story, true)
          true
        else
          log_progress(iteration, story, false)
          false
        end
      else
        puts '❌ Implementation failed'
        log_progress(iteration, story, false)
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

    def self.commit_changes(story)
      puts '💾 Committing changes...'

      ErrorHandler.with_error_handling('Git commit', { story: story['id'] }) do
        # Check if there are changes to commit
        status_output = `git status --porcelain 2>/dev/null`
        if status_output.nil? || status_output.strip.empty?
          Logger.info('No changes to commit', { story: story['id'] })
          return true
        end

        ErrorHandler.safe_system_command('git add .', 'Stage changes')

        commit_title = story['title']&.to_s&.gsub("'", "''") || 'Story implementation'
        commit_desc = story['description']&.to_s&.gsub("'", "''") || ''
        story_id = story['id']&.to_s || 'unknown'

        commit_message = "feat: #{commit_title}

#{commit_desc}

Story: #{story_id}"

        ErrorHandler.safe_system_command("git commit -m '#{commit_message}'", 'Commit changes')
      end
    end

    def self.log_progress(iteration, story, success)
      Logger.log(success ? :info : :error, "Iteration #{iteration} completed", {
                   story: story['title'],
                   story_id: story['id'],
                   success: success,
                   priority: story['priority']
                 })

      ErrorHandler.with_error_handling('Progress logging') do
        log = [
          "## Iteration #{iteration} - #{Time.now.strftime('%Y-%m-%d %H:%M:%S')}",
          "Story: #{story['title']}",
          "Status: #{success ? 'Success' : 'Failed'}",
          '---'
        ].join("\n")

        File.open('progress.txt', 'a') { |f| f.puts(log + "\n") }
      end
    end

    def self.update_state(requirements)
      ErrorHandler.with_error_handling('State update') do
        File.write('prd.json', JSON.pretty_generate(requirements))
      end
    end

    def self.parse_json_safely(json_string, context = 'JSON parsing')
      return nil if json_string.nil? || json_string.strip.empty?

      ErrorHandler.with_error_handling(context) do
        cleaned = json_string.encode('UTF-8', 'binary', invalid: :replace, undef: :replace, replace: '').strip

        # Try to extract JSON if it's wrapped in markdown or other text
        json_match = cleaned.match(/\{[\s\S]*\}/)
        cleaned = json_match[0] if json_match

        parsed = JSON.parse(cleaned)
        raise ArgumentError, 'Invalid JSON structure: expected Hash' unless parsed.is_a?(Hash)

        parsed
      end
    end
  end
end
