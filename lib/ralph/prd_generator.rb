# frozen_string_literal: true

require_relative 'state_manager'

module Ralph
  class PrdGenerator
    class << self
      def generate(prompt)
        print_generation_header(prompt)
        Logger.info('Generating PRD for prompt', { prompt: prompt })

        requirements = fetch_and_parse_requirements(prompt)
        return handle_generation_failure unless requirements

        finalize_requirements(requirements)
      end

      def print_generation_header(prompt)
        puts "\n📋 PHASE 1: Generating Project Requirements Document"
        puts "🎯 Analyzing: #{prompt}"
      end

      def fetch_and_parse_requirements(prompt)
        puts "\n🔍 Building analysis prompt..."
        prd_prompt = build_prd_prompt(prompt)

        ErrorHandler.with_error_handling('PRD creation') do
          response = fetch_opencode_response(prd_prompt)
          return nil unless response

          parse_and_validate_response(response)
        end
      end

      def fetch_opencode_response(prd_prompt)
        puts "\n🚀 Sending request to OpenCode API..."
        response = CommandRunner.capture_opencode_output(prd_prompt, 'Generate PRD')
        return nil unless response

        puts "\n📝 Processing OpenCode response..."
        Logger.debug('OpenCode response received', { length: response.length })
        response
      end

      def parse_and_validate_response(response)
        puts "\n🔧 Parsing requirements..."
        requirements = JsonParser.parse_safely(response, 'PRD requirements')
        return nil unless requirements

        print_parsed_requirements(requirements)
        puts "\n🛡️ Validating requirements structure..."
        validate_requirements(requirements)
        requirements
      end

      def print_parsed_requirements(requirements)
        puts "\n✅ Requirements parsed successfully:"
        puts "  📁 Project: #{requirements['project_name']}"
        puts "  📖 Stories: #{requirements['stories'].length}"
      end

      def finalize_requirements(requirements)
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

      def handle_generation_failure
        puts "\n❌ Failed to create PRD"
        Logger.error('Failed to create PRD')
        nil
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
             - CRITICAL: Each story MUST include a test_spec with guidance for writing integration tests
          #{'   '}
          3. TEST SPECIFICATION REQUIREMENTS
             - The test_spec field provides GUIDANCE for writing actual integration test code
             - An actual test file will be created and run for EACH story before moving to the next
             - Tests must validate RUNTIME behavior, not just compilation
             - For UI features: describe interactions to automate (clicks, inputs, assertions on DOM)
             - For API integrations: describe requests to make and expected responses
             - For setup stories: describe how to verify the setup works (e.g., app starts, imports work)
             - Include specific assertions that can be coded (e.g., "element with class X should contain Y")
             - Each test builds on previous tests - later stories should verify previous functionality still works
          #{'   '}
          4. OUTPUT REQUIREMENTS
             - Respond ONLY with raw JSON (no markdown, no explanation)
          #{'   '}
           Required JSON format:
           {
             "project_name": "descriptive project name",
             "stories": [
              {
                "id": "story-1",
                "title": "Story title",
                "description": "Detailed description",
                "acceptance_criteria": ["criterion 1", "criterion 2"],
                "test_spec": "Integration test guidance: 1) Start app, 2) Navigate to X, 3) Assert element Y is visible, 4) Click Z, 5) Assert result. Be specific about selectors and expected values.",
                "priority": 1,
                "passes": false
              }
            ]
          }

          CRITICAL:#{' '}
          - Return only the JSON object, nothing else.
          - Every story MUST have a non-empty test_spec field with actionable test guidance.
          - Test specs should be specific enough to write automated tests (selectors, expected values, actions).
          - Tests are cumulative - each story's test should also verify previous stories still work.
        PROMPT
      end

      def validate_requirements(requirements)
        validate_top_level_fields(requirements)
        validate_stories_array(requirements['stories'])
        requirements['stories'].each_with_index { |story, index| validate_story(story, index) }
      end

      def validate_top_level_fields(requirements)
        required_fields = %w[project_name stories]
        missing = required_fields.select { |field| requirements[field].nil? || requirements[field].empty? }
        raise ArgumentError, "Missing required fields: #{missing.join(', ')}" if missing.any?
      end

      def validate_stories_array(stories)
        return if stories.is_a?(Array) && stories.any?

        raise ArgumentError, 'Invalid stories format: expected non-empty array'
      end

      def validate_story(story, index)
        validate_story_fields(story, index)
        validate_acceptance_criteria(story, index)
      end

      def validate_story_fields(story, index)
        required = %w[id title description acceptance_criteria priority]
        missing = required.select { |field| story[field].nil? || story[field].to_s.strip.empty? }
        raise ArgumentError, "Story #{index + 1} missing fields: #{missing.join(', ')}" if missing.any?
      end

      def validate_acceptance_criteria(story, index)
        criteria = story['acceptance_criteria']
        return if criteria.is_a?(Array) && criteria.any?

        raise ArgumentError, "Story #{index + 1} has invalid acceptance criteria"
      end

      def create_state_files(requirements)
        StateManager.write_state(requirements)
      end
    end
  end
end
