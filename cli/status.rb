# frozen_string_literal: true

module Ralph
  class Status
    def self.show
      puts '🚀 Ralph Status'
      puts '=' * 50

      unless File.exist?('prd.json')
        puts "❌ Project not initialized. Run 'ralph init' first."
        return
      end

      begin
        prd = JSON.parse(File.read('prd.json'))
        puts "\n📋 Project: #{prd['projectName']}"
        puts "🌿 Branch: #{prd['branchName']}"

        stories = prd['userStories'] || []
        total_stories = stories.length
        completed_stories = stories.count { |s| s['passes'] == true }

        puts "\n📊 Progress:"
        puts "  Total Stories: #{total_stories}"
        puts "  Completed: #{completed_stories}"
        puts "  Remaining: #{total_stories - completed_stories}"
        puts "  Progress: #{((completed_stories.to_f / total_stories) * 100).round(1)}%"

        # Show remaining stories
        remaining = stories.reject { |s| s['passes'] == true }.sort_by { |s| s['priority'] }
        if remaining.any?
          puts "\n🔄 Remaining Stories (by priority):"
          remaining.each do |story|
            status = story['passes'] ? '✅' : '⏳'
            puts "  #{status} [P#{story['priority']}] #{story['title']}"
          end
        end
      rescue JSON::ParserError => e
        puts "❌ Error reading prd.json: #{e.message}"
      end

      if File.exist?('progress.txt')
        puts "\n📝 Recent Progress (last 5 lines):"
        File.readlines('progress.txt').last(5).each do |line|
          puts "  #{line}" unless line.strip.empty?
        end
      end

      puts "\n🔧 Git Status:"
      system("git status --porcelain | head -10 || echo 'Not a git repository'")
    end
  end

  class Debug
    def self.show
      puts '🔍 Ralph Debug Information'
      puts '=' * 50

      # Project structure
      puts "\n📁 Project Structure:"
      files = ['prd.json', 'prd.json.example', 'progress.txt', 'AGENTS.md']
      files.each do |file|
        exists = File.exist?(file) ? '✅' : '❌'
        size = File.exist?(file) ? "#{File.size(file)} bytes" : 'N/A'
        puts "  #{exists} #{file} (#{size})"
      end

      # PRD details
      if File.exist?('prd.json')
        puts "\n📋 PRD Details:"
        begin
          prd = JSON.parse(File.read('prd.json'))
          puts "  Project: #{prd['projectName']}"
          puts "  Branch: #{prd['branchName']}"
          puts "  Stories: #{prd['userStories']&.length || 0}"

          if prd['userStories']
            puts "\n📖 All Stories:"
            prd['userStories'].each do |story|
              status = story['passes'] ? '✅' : '❌'
              puts "  #{status} [#{story['id']}] #{story['title']}"
              puts "    Priority: #{story['priority']}"
            end
          end
        rescue StandardError => e
          puts "  ❌ Error parsing PRD: #{e.message}"
        end
      end

      # Git info
      puts "\n🔧 Git Information:"
      puts "  Current branch: #{`git branch --show-current 2>/dev/null || echo 'N/A'`.strip}"
      puts "  Last commit: #{`git log -1 --oneline 2>/dev/null || echo 'N/A'`.strip}"

      # Progress log
      if File.exist?('progress.txt') && !File.empty?('progress.txt')
        puts "\n📝 Full Progress Log:"
        puts File.read('progress.txt')
      else
        puts "\n📝 No progress log found"
      end
    end
  end
end
