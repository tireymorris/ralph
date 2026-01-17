# frozen_string_literal: true

module Ralph
  class Project
    def self.init
      puts '🚀 Initializing Ralph project...'

      dirs = %w[tasks archive scripts/ralph skills/prd skills/ralph]
      dirs.each do |dir|
        FileUtils.mkdir_p(dir) unless Dir.exist?(dir)
        puts "  ✓ Created #{dir}/"
      end

      create_files unless File.exist?('prd.json.example')

      puts "\n✅ Ralph project initialized successfully!"
      puts "\nNext steps:"
      puts '1. ralph prd:create "your feature description"'
      puts '2. ralph prd:convert tasks/prd-your-feature.md'
      puts '3. ralph run'
    end

    def self.create_files
      # Example PRD JSON
      File.write('prd.json.example', <<~JSON)
        {
          "projectName": "Example Project",
          "branchName": "feature/example-feature",
          "userStories": [
            {
              "id": "story-1",
              "title": "Example story title",
              "description": "Detailed description of what needs to be implemented",
              "acceptanceCriteria": [
                "Criterion 1",
                "Criterion 2"
              ],
              "priority": 1,
              "passes": false
            }
          ]
        }
      JSON

      File.write('progress.txt', "# Ralph Progress Log\n\n")

      File.write('AGENTS.md', <<~MARKDOWN)
        File.write('AGENTS.md', <<~MARKDOWN)
          # Project Agents Documentation
        #{'  '}
          ## Patterns Discovered
          <!-- Add patterns discovered during development -->
        #{'  '}
          ## Gotchas
          <!-- Add common gotchas and issues encountered -->
        #{'  '}
          ## Useful Context
          <!-- Add project-specific context and conventions -->
      MARKDOWN

      puts '  ✓ Created prd.json.example'
      puts '  ✓ Created progress.txt'
      puts '  ✓ Created AGENTS.md'
    end
  end
end
