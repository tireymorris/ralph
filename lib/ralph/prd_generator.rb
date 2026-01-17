# frozen_string_literal: true

module Ralph
  # PRD (Product Requirements Document) generator
  class PrdGenerator
    class << self
      def generate(prompt)
        Logger.info('Generating PRD for prompt', { prompt: prompt })

        prd_prompt = build_prd_prompt(prompt)

        success = ErrorHandler.with_error_handling('PRD creation') do
          response = ErrorHandler.safe_system_command("opencode run \"#{prd_prompt}\" 2>/dev/null", 'Generate PRD')
          return nil unless response

          response = response.encode('UTF-8', 'binary', invalid: :replace, undef: :replace, replace: '').strip
          Logger.debug('OpenCode response received', { length: response.length })

          requirements = ErrorHandler.parse_json_safely(response, 'PRD requirements')
          return nil unless requirements

          validate_requirements(requirements)
          create_state_files(requirements)

          Logger.info('PRD analysis complete', {
                        project: requirements['project_name'],
                        stories: requirements['stories'].length
                      })

          requirements
        end

        unless success
          Logger.error('Failed to create PRD')
          return nil
        end

        success
      end

      private

      def build_prd_prompt(prompt)
        <<~PROMPT
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
      end

      def validate_requirements(requirements)
        # Validate required structure
        required_fields = %w[project_name branch_name stories]
        missing_fields = required_fields.select { |field| requirements[field].nil? || requirements[field].empty? }
        raise ArgumentError, "Missing required fields: #{missing_fields.join(', ')}" if missing_fields.any?

        unless requirements['stories'].is_a?(Array) && requirements['stories'].any?
          raise ArgumentError, 'Invalid stories format: expected non-empty array'
        end

        # Validate each story structure
        requirements['stories'].each_with_index do |story, index|
          story_fields = %w[id title description acceptance_criteria priority]
          missing_story_fields = story_fields.select { |field| story[field].nil? || story[field].to_s.strip.empty? }

          if missing_story_fields.any?
            raise ArgumentError, "Story #{index + 1} missing fields: #{missing_story_fields.join(', ')}"
          end

          unless story['acceptance_criteria'].is_a?(Array) && story['acceptance_criteria'].any?
            raise ArgumentError, "Story #{index + 1} has invalid acceptance criteria"
          end
        end
      end

      def create_state_files(requirements)
        ErrorHandler.with_error_handling('State file creation') do
          prd_file = Ralph::Config.get(:prd_file)
          agents_file = Ralph::Config.get(:agents_file)

          File.write(prd_file, JSON.pretty_generate(requirements))

          agents_content = "# Ralph Agent Patterns\n\n## Project Context\n- Technology: #{requirements['project_name']}\n- Stories: #{requirements['stories'].length} items\n- Started: #{Time.now.strftime('%Y-%m-%d %H:%M:%S')}\n\n"
          File.write(agents_file, agents_content)
        end
      end
    end
  end
end
