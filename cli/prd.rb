# frozen_string_literal: true

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

        Write the PRD to: #{filename}

        Focus on clarity and actionable requirements that can be implemented iteratively.
        Make user stories small enough for single-iteration implementation.

        Write the file and then respond with "OK" when complete.
      PROMPT

      response = llm.complete(prompt, { write_files: true })

      if File.exist?(filename)
        puts "  ✓ Created #{filename}"
        puts "\nNext: Run 'ralph prd:convert #{filename}' to extract stories"
      else
        puts '  ❌ Failed to create PRD file'
        puts "  LLM response: #{response}"
      end
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

        # Show preview
        puts "\nUser Stories Preview:"
        stories.each_with_index do |story, i|
          puts "  #{i + 1}. #{story['title']} (Priority: #{story['priority']})"
        end

        puts "\nReady to run: ralph run"
      else
        puts '  ❌ Failed to extract stories from PRD'
      end
    end
  end
end
