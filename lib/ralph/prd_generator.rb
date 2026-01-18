# frozen_string_literal: true

require_relative 'state_manager'

module Ralph
  class PrdGenerator
    class << self
      def generate(prompt)
        puts "\n📋 PHASE 1: Generating Project Requirements Document"
        puts "🎯 Analyzing: #{prompt}"

        Logger.info('Generating PRD for prompt', { prompt: prompt })

        puts "\n🔍 Building analysis prompt..."
        prd_prompt = build_prd_prompt(prompt)

        success = ErrorHandler.with_error_handling('PRD creation') do
          puts "\n🚀 Sending request to OpenCode API..."
          response = ErrorHandler.capture_command_output(prd_prompt, 'Generate PRD')
          return nil unless response

          puts "\n📝 Processing OpenCode response..."
          Logger.debug('OpenCode response received', { length: response.length })

          puts "\n🔧 Parsing requirements..."
          requirements = ErrorHandler.parse_json_safely(response, 'PRD requirements')
          return nil unless requirements

          puts "\n✅ Requirements parsed successfully:"
          puts "  📁 Project: #{requirements['project_name']}"
          puts "  🌿 Branch: #{requirements['branch_name']}"
          puts "  📖 Stories: #{requirements['stories'].length}"

          puts "\n🛡️ Validating requirements structure..."
          validate_requirements(requirements)

          puts "\n💾 Creating state files..."
          create_state_files(requirements)

          puts "\n🎉 PRD Analysis Complete!"
          puts "  ✅ Project: #{requirements['project_name']}"
          puts "  ✅ Stories: #{requirements['stories'].length}"

          Logger.info('PRD analysis complete', {
                        project: requirements['project_name'],
                        stories: requirements['stories'].length
                      })

          requirements
        end

        unless success
          puts "\n❌ Failed to create PRD"
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
        StateManager.write_state(requirements)
      end
    end
  end
end
