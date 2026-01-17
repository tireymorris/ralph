# frozen_string_literal: true

require 'json'
require 'fileutils'

module Ralph
  class Agent
    def self.run(prompt)
      # Change to the directory where ralph was invoked
      Dir.chdir(ENV['PWD'] || Dir.pwd)

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

      begin
        response = `opencode run "#{prd_prompt}" 2>/dev/null`
        response = response.encode('UTF-8', 'binary', invalid: :replace, undef: :replace, replace: '').strip
        requirements = JSON.parse(response)

        # Create state files
        File.write('prd.json', JSON.pretty_generate(requirements))

        agents_content = "# Ralph Agent Patterns\n\n## Project Context\n- Technology: #{requirements['project_name']}\n- Stories: #{requirements['stories'].length} items\n- Started: #{Time.now.strftime('%Y-%m-%d %H:%M:%S')}\n\n"
        File.write('AGENTS.md', agents_content)

        puts "✅ Analyzed project: #{requirements['project_name']}"
        puts "📊 Found #{requirements['stories'].length} user stories"
      rescue JSON::ParserError => e
        puts "❌ Failed to parse requirements: #{e.message}"
        puts "Response was: #{response[0..200]}..."
        return
      end

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
      system("git checkout -b #{branch_name} 2>/dev/null || git checkout #{branch_name}")
    end

    def self.implement_story(story, iteration, all_requirements)
      completed = all_requirements['stories'].count { |s| s['passes'] == true }
      total = all_requirements['stories'].length

      context = File.exist?('AGENTS.md') ? File.read('AGENTS.md') : ''

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

      response = `opencode run "#{implementation_prompt}" 2>/dev/null`
      response = response.encode('UTF-8', 'binary', invalid: :replace, undef: :replace, replace: '').strip

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

      system('git add .')
      system("git commit -m 'feat: #{story['title']}\n\n#{story['description']}\n\nStory: #{story['id']}'")

      puts '✓ Changes committed'
    end

    def self.log_progress(iteration, story, success)
      log = [
        "## Iteration #{iteration} - #{Time.now.strftime('%Y-%m-%d %H:%M:%S')}",
        "Story: #{story['title']}",
        "Status: #{success ? 'Success' : 'Failed'}",
        '---'
      ].join("\n")

      File.open('progress.txt', 'a') { |f| f.puts(log + "\n") }
    end

    def self.update_state(requirements)
      File.write('prd.json', JSON.pretty_generate(requirements))
    end
  end
end
