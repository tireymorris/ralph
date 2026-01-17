# frozen_string_literal: true

module Ralph
  class Agent
    def self.run(max_iterations = nil)
      puts '🤖 Starting Ralph autonomous agent...'

      unless File.exist?('prd.json')
        puts "❌ No prd.json found. Please run 'ralph prd:convert' first."
        return
      end

      prd = JSON.parse(File.read('prd.json'))
      stories = prd['userStories'] || []

      puts "📋 Found #{stories.length} stories in #{prd['projectName']}"
      if max_iterations
        puts "🔄 Running up to #{max_iterations} iterations...\n"
      else
        puts "🔄 Running until completion...\n"
      end

      create_feature_branch(prd['branchName'])

      iteration_count = 0
      loop do
        iteration_count += 1
        max_reached = max_iterations && iteration_count > max_iterations

        puts '=' * 60
        puts "🔄 Iteration #{iteration_count}#{" / #{max_iterations}" if max_iterations}"
        puts '=' * 60

        completed_stories = stories.select { |s| s['passes'] == true }
        if completed_stories.length == stories.length
          puts "\n🎉 All stories completed!"
          puts '<promise>COMPLETE</promise>'
          break
        end

        if max_reached
          puts "\n🏁 Maximum iterations reached"
          puts "Run 'ralph status' to see current progress"
          break
        end

        next_story = stories.reject { |s| s['passes'] == true }.min_by { |s| s['priority'] }

        puts "\n📖 Working on: #{next_story['title']}"
        puts "🎯 Priority: #{next_story['priority']}"
        puts "📝 Description: #{next_story['description']}"

        success = implement_story(next_story)

        if success
          next_story['passes'] = true

          if run_quality_checks
            commit_changes(next_story)
            File.write('prd.json', JSON.pretty_generate(prd))
            log_progress(iteration_count, next_story, success)
          else
            puts '❌ Quality checks failed, skipping commit'
          end
        else
          puts '❌ Story implementation failed'
          log_progress(iteration_count, next_story, false)
        end

        puts "✅ Iteration #{iteration_count} completed"
        if max_iterations
          puts 'Press Enter to continue...'
          begin
            gets
          rescue StandardError
          end
        else
          puts 'Continuing to next iteration...'
          sleep(1)
        end
      end
    end

    def self.create_feature_branch(branch_name)
      puts "🌿 Creating feature branch: #{branch_name}"
      system("git checkout -b #{branch_name} 2>/dev/null || git checkout #{branch_name}")
      puts '  ✓ Branch ready'
    end

    def self.implement_story(story)
      puts "\n🔧 Implementing story: #{story['title']}"

      llm = Ralph::LLM.new
      project_context = build_project_context

      response = llm.implement_story(story, project_context)

      if response && response.include?('COMPLETED:')
        puts "   ✓ #{response}"
        true
      else
        puts "   ❌ Response: #{response}"
        false
      end
    rescue StandardError => e
      puts "   ❌ Implementation failed: #{e.message}"
      false
    end

    def self.run_quality_checks
      puts "\n🧪 Running quality checks..."

      if File.exist?('quality-checks.sh')
        # Use LLM-generated quality script
        puts '   Running LLM-generated quality checks...'
        result = system('./quality-checks.sh')
        puts result ? '   ✓ All quality checks passed' : '   ❌ Quality checks failed'
        result
      else
        # Fallback to basic checks
        llm = Ralph::LLM.new
        quality_commands = llm.analyze_project_for_quality_checks

        all_passed = true
        critical_checks = %w[typecheck test]

        critical_checks.each do |check_type|
          command = quality_commands[check_type]
          next unless command && !command.include?('No')

          print "   #{check_type.capitalize}... "
          result = system(command)
          if result
            puts '✅'
          else
            puts '❌'
            all_passed = false
          end
        end

        all_passed
      end
    end

    def self.commit_changes(story)
      puts "\n💾 Committing changes..."

      commit_message = "feat: #{story['title']}\n\n#{story['description']}\n\nStory: #{story['id']}"

      system('git add .')
      system("git commit -m \"#{commit_message}\"")

      puts '  ✓ Changes committed'
    end

    def self.log_progress(iteration, story, success)
      log_entry = [
        "## Iteration #{iteration} - #{Time.now.strftime('%Y-%m-%d %H:%M:%S')}",
        "Story: #{story['title']}",
        "Status: #{success ? '✅ Success' : '❌ Failed'}",
        "Description: #{story['description']}",
        'Acceptance Criteria:',
        story['acceptanceCriteria'].map { |c| "  - #{c}" }.join("\n"),
        '',
        '---'
      ].join("\n")

      File.open('progress.txt', 'a') do |f|
        f.puts log_entry
        f.puts
      end

      puts '  ✓ Progress logged'
    end

    # LLM handles AGENTS.md updates during story implementation
    def self.update_agents_md(story)
      # No-op - LLM updates AGENTS.md during implementation
    end

    def self.build_project_context
      agents_content = File.exist?('AGENTS.md') ? File.read('AGENTS.md') : ''
      previous_iterations = File.exist?('progress.txt') ? File.read('progress.txt').scan(/Iteration \d+/).length : 0

      {
        agents_content: agents_content,
        previous_iterations: previous_iterations,
        scan_files: true
      }
    end
  end
end
