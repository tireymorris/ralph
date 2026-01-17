# frozen_string_literal: true

module Ralph
  class LLM
    attr_reader :provider, :model, :client

    def initialize(provider = nil)
      @provider = provider || ENV['RALPH_LLM_PROVIDER'] || 'openai'
      @model = ENV['RALPH_LLM_MODEL'] || default_model
      @client = setup_client
    end

    def complete(prompt, context = {})
      system_prompt = build_system_prompt(context)

      case @provider
      when 'openai'
        complete_openai(system_prompt, prompt)
      when 'anthropic'
        complete_anthropic(system_prompt, prompt)
      when 'ollama'
        complete_ollama(system_prompt, prompt)
      else
        raise "Unsupported provider: #{@provider}"
      end
    end

    def analyze_project_for_quality_checks
      prompt = <<~PROMPT
        Analyze this project and determine the appropriate quality check commands.
        Look at package.json, requirements.txt, Cargo.toml, etc. to understand the tech stack.

        Return JSON with:
        {
          "typecheck": "command to run type checking",
          "test": "command to run tests",#{' '}
          "lint": "command to run linting",
          "build": "command to build project"
        }

        If a command doesn't exist for this stack, return null for that field.
      PROMPT

      response = complete(prompt, { scan_files: true })
      JSON.parse(response)
    rescue JSON::ParserError
      fallback_quality_checks
    end

    def extract_stories_from_prd(prd_content)
      prompt = <<~PROMPT
        Extract user stories from this PRD content. Return JSON array with:
        [
          {
            "id": "story-1",
            "title": "Story title",
            "description": "Detailed description",
            "acceptanceCriteria": ["criterion 1", "criterion 2"],
            "priority": 1
          }
        ]

        Each story should be small enough to complete in one iteration.
        Assign priorities (1 = highest priority).

        PRD content:
        #{prd_content}
      PROMPT

      response = complete(prompt)
      JSON.parse(response)
    rescue JSON::ParserError => e
      puts "⚠️  Failed to parse stories from LLM response: #{e.message}"
      fallback_stories
    end

    def implement_story(story, project_context = {})
      prompt = <<~PROMPT
        Implement this user story. Return only the completed code changes.

        Story: #{story['title']}
        Description: #{story['description']}
        Acceptance Criteria:
        #{story['acceptanceCriteria'].map { |c| "- #{c}" }.join("\n")}

        Context:
        - Previous iterations: #{project_context[:previous_iterations] || 0}
        - AGENTS.md insights: #{project_context[:agents_content] || 'None'}

        Follow project conventions and ensure all acceptance criteria are met.
        Write complete, working code that can be committed.
      PROMPT

      complete(prompt, project_context)
    end

    private

    def build_system_prompt(context)
      base_prompt = 'You are an expert software developer implementing user stories iteratively.'

      base_prompt += "\nScan the project files to understand the tech stack and structure." if context[:scan_files]

      base_prompt += "\nProject patterns and insights:\n#{context[:agents_content]}" if context[:agents_content]

      base_prompt
    end

    def setup_client
      case @provider
      when 'openai'
        require 'openai'
        OpenAI::Client.new(access_token: ENV['OPENAI_API_KEY'])
      when 'anthropic'
        require 'anthropic'
        Anthropic::Client.new(api_key: ENV['ANTHROPIC_API_KEY'])
      when 'ollama'
        # Ollama uses HTTP calls directly
        nil
      else
        raise "Unsupported provider: #{@provider}"
      end
    end

    def complete_openai(system_prompt, user_prompt)
      response = @client.chat(
        model: @model,
        messages: [
          { role: 'system', content: system_prompt },
          { role: 'user', content: user_prompt }
        ]
      )
      response.dig('choices', 0, 'message', 'content')
    end

    def complete_anthropic(system_prompt, user_prompt)
      response = @client.messages(
        model: @model,
        system: system_prompt,
        messages: [
          { role: 'user', content: user_prompt }
        ]
      )
      response.content[0].text
    end

    def complete_ollama(system_prompt, user_prompt)
      require 'net/http'
      require 'json'

      uri = URI('http://localhost:11434/api/generate')
      response = Net::HTTP.post(uri, {
        model: @model,
        prompt: "#{system_prompt}\n\n#{user_prompt}",
        stream: false
      }.to_json, 'Content-Type' => 'application/json')

      JSON.parse(response.body)['response']
    end

    def default_model
      case @provider
      when 'openai' then 'gpt-4'
      when 'anthropic' then 'claude-3-sonnet-20241022'
      when 'ollama' then 'codellama'
      else 'gpt-4'
      end
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
