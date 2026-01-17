# frozen_string_literal: true

require 'fileutils'

module Ralph
  class PRD
    def self.create(description)
      timestamp = Time.now.strftime('%Y%m%d-%H%M%S')
      filename = "tasks/prd-#{timestamp}.md"

      FileUtils.mkdir_p('tasks')

      llm = Ralph::LLM.new
      prompt = <<~PROMPT
        You are a product manager creating a detailed PRD for: #{description}

        Create a comprehensive Product Requirements Document with:
        1. Executive Summary
        2. Feature Description#{'  '}
        3. User Stories (detailed, with acceptance criteria)
        4. Technical Requirements
        5. Success Metrics

        Focus on clarity and actionable requirements that can be implemented iteratively.
        Make user stories small enough for single-iteration implementation.

        Return the complete PRD as markdown content only.
      PROMPT

      response = llm.complete(prompt)
      markdown_content = extract_markdown_content(response)

      File.write(filename, markdown_content)

      puts "  ✓ Created #{filename}"
      puts "\nNext: Run 'ralph prd:convert #{filename}' to extract stories"
    end

    def self.convert(file_path)
      unless File.exist?(file_path)
        puts "❌ Error: PRD file not found: #{file_path}"
        return
      end

      puts '  🤖 LLM analyzing PRD and extracting stories...'

      llm = Ralph::LLM.new
      prd_content = File.read(file_path)
      stories = llm.extract_stories_from_prd(prd_content)

      if stories && !stories.empty?
        puts "  ✓ Extracted #{stories.length} user stories"
        puts '  ✓ Created prd.json with stories prioritized for implementation'

        puts "\nUser Stories Preview:"
        stories.each_with_index do |story, i|
          puts "  #{i + 1}. #{story['title']} (Priority: #{story['priority']})"
        end

        puts "\nReady to run: ralph run"
      else
        puts '  ❌ Failed to extract stories from PRD'
      end
    end

    def self.extract_markdown_content(response)
      if response.match(/```markdown\s*\n(.*?)\n```/m)
        ::Regexp.last_match(1)
      elsif response.match(/```\s*\n(.*?)\n```/m)
        ::Regexp.last_match(1)
      else
        response
      end
    end
  end
end
