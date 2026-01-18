# frozen_string_literal: true

module Ralph
  # Error handling utilities for Ralph agent
  class ErrorHandler
    class << self
      def log_error(operation, error, context = {})
        Logger.error("Error in #{operation}", {
                       error_class: error.class.name,
                       error_message: error.message,
                       backtrace: error.backtrace&.first(3),
                       context: context
                     })
      end

      def with_error_handling(operation, context = {})
        result = yield
        Logger.debug("Completed #{operation}", context)
        result
      rescue StandardError => e
        log_error(operation, e, context)
        nil
      end

      def safe_system_command(command, operation)
        Logger.debug("Executing command: #{command}", { operation: operation })

        # No timeouts - let it cook
        full_command = command
        result = system(full_command)

        if result.nil?
          Logger.error('Command execution failed', { command: command, operation: operation })
          false
        elsif !result
          Logger.warn('Command returned non-zero', { command: command, operation: operation })
          false
        else
          Logger.debug('Command succeeded', { command: command, operation: operation })
          true
        end
      rescue StandardError => e
        Logger.error('System command exception', { command: command, operation: operation, error: e.message })
        false
      end

      def parse_json_safely(json_string, context = 'JSON parsing')
        return nil if json_string.nil? || json_string.strip.empty?

        with_error_handling(context) do
          cleaned = json_string.encode('UTF-8', 'binary', invalid: :replace, undef: :replace, replace: '').strip

          # Try to extract JSON if it's wrapped in markdown or other text
          json_match = cleaned.match(/\{[\s\S]*\}/)
          cleaned = json_match[0] if json_match

          parsed = JSON.parse(cleaned)
          raise ArgumentError, 'Invalid JSON structure: expected Hash' unless parsed.is_a?(Hash)

          parsed
        end
      end
    end
  end
end
