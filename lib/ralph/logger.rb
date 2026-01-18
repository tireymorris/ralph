# frozen_string_literal: true

module Ralph
  # Logging system for Ralph agent
  # Provides structured logging with configurable levels and output
  class Logger
    LOG_LEVELS = { debug: 0, info: 1, warn: 2, error: 3 }.freeze

    class << self
      attr_reader :level

      def configure(level = nil)
        @level = LOG_LEVELS[level&.to_sym] ||
                 LOG_LEVELS[Ralph::Config.get(:log_level)&.to_sym] ||
                 LOG_LEVELS[:info]
        @@buffer ||= []
      end

      def log(level, message, context = {})
        return unless @level && LOG_LEVELS[level] >= @level

        timestamp = Time.now.strftime('%Y-%m-%d %H:%M:%S')
        {
          timestamp: timestamp,
          level: level.to_s.upcase,
          message: message,
          context: context
        }

        formatted = "[#{timestamp}] #{level.to_s.upcase}: #{message}"
        formatted += " | #{context}" unless context.empty?

        puts formatted
        @@buffer << formatted
      rescue StandardError => e
        puts "❌ Logger error: #{e.message}"
      end

      def flush
        @@buffer.each { |msg| write_to_file(msg) }
        @@buffer.clear
      end

      def debug(message, context = {})
        log(:debug, message, context)
      end

      def info(message, context = {})
        log(:info, message, context)
      end

      def warn(message, context = {})
        log(:warn, message, context)
      end

      def error(message, context = {})
        log(:error, message, context)
      end

      private

      def write_to_file(message)
        log_file = Ralph::Config.get(:log_file)
        File.open(log_file, 'a') { |f| f.puts(message) }
      end
    end
  end
end
