# frozen_string_literal: true

require 'json'

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

        # Safely execute a block with error handling(operation, context = {})
        result = yield
        Logger.debug("Completed #{operation}", context)
        result
      rescue StandardError => e
        log_error(operation, e, context)
        nil
      end

      def capture_command_output(prompt, operation)
        Logger.debug("Capturing output for: #{prompt[0..100]}...", { operation: operation })

        # Write prompt to file in current directory
        prompt_file = ".ralph_prompt_#{Process.pid}.txt"
        begin
          File.write(prompt_file, prompt)

          cmd = "bash -c 'cat #{prompt_file.shellescape} | opencode run --format default /dev/stdin' 2>&1"
          output = `#{cmd}`

          # Clean output - remove ANSI codes and JSON artifacts
          cleaned = clean_opencode_output(output)

          Logger.debug('Command output captured', { operation: operation, output_length: cleaned.length })
          cleaned
        ensure
          File.delete(prompt_file) if File.exist?(prompt_file)
        end
      rescue StandardError => e
        Logger.error('Command capture exception', { prompt: prompt[0..100], operation: operation, error: e.message })
        nil
      end

      def clean_opencode_output(output)
        return '' if output.nil? || output.strip.empty?

        output
          .gsub(/\x1b\[[0-9;]*[a-zA-Z]/, '') # Remove ANSI color codes
          .gsub(/^[{"].*$/m, '') # Remove JSON artifacts from beginning
          .gsub(/\n{3,}/, "\n\n") # Reduce multiple newlines to max 2
          .strip
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
