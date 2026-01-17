# Ralph Configuration
# frozen_string_literal: true

module Ralph
  class Config
    DEFAULTS = {
      opencode_timeout: 300, # seconds
      git_timeout: 30,               # seconds
      test_timeout: 120,             # seconds
      max_iterations: 50,            # maximum iterations before stopping
      log_level: :info,              # debug, info, warn, error
      log_file: 'ralph.log',         # log filename
      progress_file: 'progress.txt', # progress tracking file
      prd_file: 'prd.json',          # PRD state file
      agents_file: 'AGENTS.md',      # patterns file
      retry_attempts: 3, # number of retries for failed operations
      retry_delay: 5 # seconds between retries
    }.freeze

    attr_reader :settings

    def self.load
      @settings ||= DEFAULTS.merge(load_from_file)
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
