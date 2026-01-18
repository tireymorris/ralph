# frozen_string_literal: true

module Ralph
  class Config
    DEFAULTS = {
      opencode_timeout: nil,
      git_timeout: nil,
      test_timeout: nil,
      max_iterations: 50,
      log_level: :info,
      prd_file: 'prd.json',
      retry_attempts: 3,
      retry_delay: 5
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
