# frozen_string_literal: true

module Ralph
  class Logger
    LOG_LEVELS = { debug: 0, info: 1, warn: 2, error: 3 }.freeze

    class << self
      attr_reader :level

      def configure(level = nil)
        @level = LOG_LEVELS[level&.to_sym] ||
                 LOG_LEVELS[Ralph::Config.get(:log_level)&.to_sym] ||
                 LOG_LEVELS[:info]
      end

      def log(level, message, context = {})
        return unless @level && LOG_LEVELS[level] >= @level

        timestamp = Time.now.strftime('%Y-%m-%d %H:%M:%S')
        formatted = "[#{timestamp}] #{level.to_s.upcase}: #{message}"
        formatted += " | #{context}" unless context.empty?

        puts formatted
      rescue StandardError => e
        puts "❌ Logger error: #{e.message}"
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
    end
  end
end
