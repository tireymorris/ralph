# frozen_string_literal: true

require 'English'
require 'fileutils'
require 'open3'
require 'shellwords'
require_relative 'config'
require_relative 'logger'
require_relative 'error_handler'

module Ralph
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

      prd_prompt = "Task: #{prompt}

Step 1: THOROUGHLY analyze the project by reading ALL relevant files:
- Use Glob to find all source files (*.rb, *.js, *.ts, *.py, etc.)
- Read key files to understand the project structure, existing code, dependencies, and current state
- Check for existing test files, package.json, Gemfile, or other config files
- Understand what frameworks, libraries, and patterns are currently used

Step 2: Based on your thorough analysis of the codebase and the user's request \"#{prompt}\", create PRD.md with:
# PRD - #{prompt.split(' ').first.capitalize} #{prompt.split(' ')[1..3].join(' ')}
## Branch: feature/implementation

## Overview
[Brief description of what needs to be done based on both the user request AND your codebase analysis]

## Goals
[List appropriate number of goals based on the actual project state and user request - could be 2-8+ goals as needed]

## Technical Approach
[Describe the approach considering the existing codebase, frameworks, and patterns discovered]

Step 3: Create appropriate number of story files (story-1.md, story-2.md, etc.) based on the complexity of the user's request \"#{prompt}\" and your codebase analysis. Don't limit to 2 stories if the request needs more breakdown. Each story should have:
# Story story-X: [Relevant Title]
## Priority: 1
## Description: [Specific description based on both user request and actual codebase analysis]
## Acceptance Criteria:
- [Specific, actionable criteria that address the user's needs in the context of the existing codebase]

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

        puts "\n#{'=' * 60}"
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

      # Commit any changes before creating branch to avoid carrying modifications
      ErrorHandler.with_error_handling('Git status check before branch') do
        status_output = ErrorHandler.capture_command_output('git status --porcelain', 'Check git status')
        if status_output && !status_output.strip.empty?
          puts '📝 Committing changes before creating branch...'
          ErrorHandler.safe_system_command('git add .', 'Stage changes before branch')
          ErrorHandler.safe_system_command('git commit -m "Commit changes before branch creation"',
                                           'Commit changes before branch')
        end
      end

      ErrorHandler.with_error_handling('Git branch creation', { branch: branch_name }) do
        # Switch to main/master first to ensure clean branch creation
        main_branch = system('git rev-parse --verify main >/dev/null 2>&1') ? 'main' : 'master'
        ErrorHandler.safe_system_command("git checkout #{main_branch}", "Switch to #{main_branch} branch")

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

        File.open('progress.txt', 'a') { |f| f.puts("#{log}\n") }
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
