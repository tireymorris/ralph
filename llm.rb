# frozen_string_literal: true

require 'json'
require 'net/http'

module Ralph
  class LLM
    attr_reader :provider, :model

    def initialize(provider = nil)
      @provider = provider || ENV['RALPH_LLM_PROVIDER'] || 'openai'
      @model = ENV['RALPH_LLM_MODEL'] || default_model
    end

    def complete(prompt, context = {})
      system_prompt = build_system_prompt(context)
      complete_ollama(system_prompt, prompt)
    end

    def analyze_project_for_quality_checks
      prompt = <<~PROMPT
        You are a DevOps engineer. Analyze this project to determine quality check commands.

        Return JSON with these commands for this tech stack:
        {
          "typecheck": "command or null",
          "test": "command",#{' '}
          "lint": "command or null",
          "build": "command or null"
        }

        Use null if not applicable to this project.
      PROMPT

      response = complete(prompt, { scan_files: true })
      JSON.parse(response)
    rescue JSON::ParserError
      fallback_quality_checks
    rescue StandardError => e
      puts "⚠️  Failed to analyze quality checks: #{e.message}"
      fallback_quality_checks
    end

    def extract_stories_from_prd(prd_content)
      prompt = <<~PROMPT
        You are a software architect. Analyze this PRD and extract user stories.

        Create user stories small enough for one iteration. Assign priorities (1 = highest).

        Return JSON only:
        {
          "projectName": "project name",
          "branchName": "feature/name",
          "userStories": [
            {
              "id": "story-1",
              "title": "Story title",#{' '}
              "description": "Detailed description",
              "acceptanceCriteria": ["criterion 1", "criterion 2"],
              "priority": 1,
              "passes": false
            }
          ]
        }

        PRD content:
        #{prd_content}
      PROMPT

      response = complete(prompt)
      json_data = JSON.parse(response)

      # Write prd.json file
      File.write('prd.json', JSON.pretty_generate(json_data))
      puts '  ✓ Created prd.json'

      # Update AGENTS.md
      agents_insights = "# Ralph Agent Patterns\n\n## Architecture Insights\n- Project: #{json_data['projectName']}\n- Stories prioritized for iterative implementation\n"
      File.write('AGENTS.md', agents_insights)
      puts '  ✓ Updated AGENTS.md'

      json_data['userStories']
    rescue JSON::ParserError => e
      puts "⚠️  Failed to parse stories: #{e.message}"
      fallback_stories
    end

    def implement_story(story, project_context = {})
      prompt = <<~PROMPT
        You are an expert software developer implementing user stories iteratively.

        Story: #{story['title']}
        Description: #{story['description']}
        Acceptance Criteria:
        #{story['acceptanceCriteria'].map { |c| "- #{c}" }.join("\n")}

        Context:
        - Previous iterations: #{project_context[:previous_iterations] || 0}
        - AGENTS.md patterns: #{project_context[:agents_content] || 'None'}

        Your task:
        1. Read existing code to understand patterns
        2. Implement story completely
        3. Update any related files
        4. Update AGENTS.md with new patterns discovered
        5. Commit changes when everything works

        Work systematically. When complete, respond with "COMPLETED: [brief summary]".
      PROMPT

      complete(prompt, project_context)
    end

    private

    def build_system_prompt(context)
      base_prompt = 'You are an expert software developer working autonomously.'

      base_prompt += "\nYou can read files to understand the project." if context[:scan_files]

      base_prompt += "\nProject patterns:\n#{context[:agents_content]}" if context[:agents_content]

      base_prompt += "\nWork systematically and ensure tasks are completed properly."
      base_prompt
    end

    def complete_ollama(system_prompt, user_prompt)
      uri = URI('http://localhost:11434/api/generate')
      prompt = "#{system_prompt}\n\n#{user_prompt}"

      response = Net::HTTP.post(uri, {
        model: @model,
        prompt: prompt,
        stream: false
      }.to_json, 'Content-Type' => 'application/json')

      JSON.parse(response.body)['response']
    end

    def default_model
      @model || 'qwen3-coder:latest'
    end

    def fallback_quality_checks
      {
        'typecheck' => 'echo "No typecheck command configured"',
        'test' => 'echo "No test command configured"',
        'lint' => 'echo "No lint command configured"',
        'build' => 'echo "No build command configured"'
      }
    end

    def fallback_stories
      [{
        'id' => 'story-1',
        'title' => 'Implement feature',
        'description' => 'Implement the main feature described in the PRD',
        'acceptanceCriteria' => [
          'Feature works as expected',
          'Tests pass',
          'Code follows project conventions'
        ],
        'priority' => 1
      }]
    end
  end
end
