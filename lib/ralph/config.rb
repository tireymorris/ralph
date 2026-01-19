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
      command_timeout: nil,
      max_iterations: 50,
      log_level: :info,
      prd_file: 'prd.json',
      retry_attempts: 3,
      retry_delay: 5
    }.freeze

    class << self
      def load
        @load ||= DEFAULTS.merge(load_from_file)
      end

      def get(key)
        load[key]
      end

      def set(key, value)
        if key == :model
          validate_model!(value)
        end
        load[key] = value
      end

      def validate_model!(model)
        return if model.nil?
        return if SUPPORTED_MODELS.include?(model)

        raise ArgumentError, "Unsupported model: #{model}. Supported models: #{SUPPORTED_MODELS.join(', ')}"
      end

      def reset!
        @load = nil
      end

      private

      def load_from_file
        config_file = 'ralph.config.json'
        return {} unless File.exist?(config_file)

        begin
          config = JSON.parse(File.read(config_file), symbolize_names: true)
          validate_model!(config[:model]) if config[:model]
          config
        rescue JSON::ParserError => e
          puts "⚠️ Invalid config file: #{e.message}"
          {}
        rescue ArgumentError => e
          puts "⚠️ Invalid config: #{e.message}"
          {}
        end
      end
    end
  end
end
