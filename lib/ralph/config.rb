# frozen_string_literal: true

module Ralph
  # Configuration management for Ralph agent
  class Config
    DEFAULTS = {
      opencode_timeout: nil, # no timeout - let it cook
      git_timeout: nil,       # no timeout
      test_timeout: nil,      # no timeout
      max_iterations: 50,            # maximum iterations before stopping
      log_level: :info,              # debug, info, warn, error
      log_file: 'ralph.log',         # log filename
      progress_file: 'progress.txt', # progress tracking file
      prd_file: 'prd.json',          # PRD state file
      agents_file: 'AGENTS.md',      # patterns file
      retry_attempts: 3, # number of retries for failed operations
      retry_delay: 5 # seconds between retries
    }.freeze

    class << self
      attr_reader :settings

        # Load configuration from defaults and file
        @load ||= DEFAULTS.merge(load_from_file)
      end

      def get(key)
        load[key]
      end

      def set(key, value)
        @settings = load
        @settings[key] = value
      end

      private

        # Load configuration from defaults and file_from_file
        config_file = 'ralph.config.json'
        return {} unless File.exist?(config_file)

        begin
          JSON.parse(File.read(config_file))
        rescue JSON::ParserError => e
          puts "⚠️ Invalid config file: #{e.message}"
          {}
        end
      end
    end
  end
end
