# frozen_string_literal: true

require 'json'

module Ralph
  class Config
    SUPPORTED_MODELS = %w[
      opencode/big-pickle
      opencode/glm-4.7-free
      opencode/gpt-5-nano
      opencode/grok-code
      opencode/minimax-m2.1-free
    ].freeze

    DEFAULT_MODEL = 'opencode/grok-code'

    DEFAULTS = {
      model: DEFAULT_MODEL,
      git_timeout: nil,
      test_timeout: nil,
      max_iterations: 50,
      log_level: :info,
      prd_file: 'prd.json',
      retry_attempts: 3,
      retry_delay: 5
    }.freeze

    class << self
      def load
        @settings ||= DEFAULTS.merge(load_from_file)
      end

      def get(key)
        load[key]
      end

      def set(key, value)
        load[key] = value
      end

      def reset!
        @settings = nil
      end

      private

      def load_from_file
        config_file = 'ralph.config.json'
        return {} unless File.exist?(config_file)

        begin
          JSON.parse(File.read(config_file), symbolize_names: true)
        rescue JSON::ParserError => e
          puts "⚠️ Invalid config file: #{e.message}"
          {}
        end
      end
    end
  end
end
