# Ralph Configuration
# frozen_string_literal: true

module Ralph
  class Config
    DEFAULTS = {
      opencode_timeout: nil, # no timeout
      git_timeout: nil,       # no timeout
      test_timeout: nil,      # no timeout
      max_iterations: 50,            # maximum iterations before stopping
      log_level: :info,              # debug, info, warn, error
      prd_file: 'prd.json',          # PRD state file
      retry_attempts: 3, # number of retries for failed operations
      retry_delay: 5 # seconds between retries
    }.freeze

    attr_reader :settings

    def self.load
      @load ||= DEFAULTS.merge(load_from_file)
    end

    def self.get(key)
      load[key]
    end

    def self.set(key, value)
      @settings = load
      @settings[key] = value
    end

    def self.load_from_file
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
