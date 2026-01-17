# frozen_string_literal: true

module Ralph
  class PRD
    def self.create(description)
      timestamp = Time.now.strftime('%Y%m%d-%H%M%S')
      filename = "tasks/prd-#{timestamp}.md"

      FileUtils.mkdir_p('tasks')

      prd_content = <<~MARKDOWN
        # Product Requirements Document

        ## Feature Description
        #{description}

        ## User Stories

        <!-- User stories will be generated here -->

        ## Technical Requirements

        <!-- Technical requirements will be listed here -->

        ## Acceptance Criteria

        <!-- Acceptance criteria will be detailed here -->
      MARKDOWN

      File.write(filename, prd_content)
      puts "  ✓ Created #{filename}"
      puts "\nNext: Edit the PRD file with detailed requirements, then run:"
      puts "  ralph prd:convert #{filename}"
    end

    def self.convert(file_path)
      unless File.exist?(file_path)
        puts "❌ Error: PRD file not found: #{file_path}"
        return
      end

      content = File.read(file_path)

      stories = parse_user_stories(content)

      prd_json = {
        projectName: extract_project_name(file_path),
        branchName: "feature/#{extract_feature_name(file_path)}",
        userStories: stories
      }

      File.write('prd.json', JSON.pretty_generate(prd_json))
      puts '  ✓ Converted to prd.json'
      puts "  ✓ Found #{stories.length} user stories"

      # Show preview
      puts "\nUser Stories Preview:"
      stories.each_with_index do |story, i|
        puts "  #{i + 1}. #{story[:title]} (Priority: #{story[:priority]})"
      end

      puts "\nReady to run: ralph run"
    end

    def self.parse_user_stories(content)
      llm = Ralph::LLM.new
      stories = llm.extract_stories_from_prd(content)

      # Ensure required fields and default passes: false
      stories.map do |story|
        story.merge('passes' => false)
      end
    rescue StandardError => e
      puts "⚠️  Failed to extract stories with LLM: #{e.message}"
      puts '  Using fallback story generation'

      # Fallback: try basic regex extraction
      fallback_parse_stories(content)
    end

    def self.fallback_parse_stories(content)
      stories = []

      content.scan(/^###?\s*(.+)$/i) do |matches|
        title = matches[0]&.strip
        next if title.nil? || title.empty?

        stories << {
          id: "story-#{stories.length + 1}",
          title: title,
          description: "Description for #{title}",
          acceptanceCriteria: [
            "Criterion 1 for #{title}",
            "Criterion 2 for #{title}"
          ],
          priority: stories.length + 1,
          passes: false
        }
      end

      if stories.empty?
        stories << {
          id: 'story-1',
          title: 'Implement feature',
          description: 'Implement the main feature described in the PRD',
          acceptanceCriteria: [
            'Feature works as expected',
            'Tests pass',
            'Code follows project conventions'
          ],
          priority: 1,
          passes: false
        }
      end

      stories
    end

    def self.extract_project_name(file_path)
      File.basename(file_path, '.md').gsub(/^prd-/, '').split('-').map(&:capitalize).join(' ')
    end

    def self.extract_feature_name(file_path)
      File.basename(file_path, '.md').gsub(/^prd-/, '').downcase.gsub(/[^a-z0-9]/, '-')
    end
  end
end
