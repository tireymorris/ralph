# frozen_string_literal: true

require 'json'
require 'shellwords'
require 'open3'

module Ralph
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

      def capture_command_output(prompt, operation)
        model = Ralph::Config.get(:model)
        cmd = %w[opencode run]
        cmd += ['--model', model] if model

        output_lines = []

        Open3.popen3(*cmd) do |stdin, stdout, stderr, wait_thr|
          stdin.write(prompt)
          stdin.close

          stdout_thread = Thread.new do
            stdout.each_line do |line|
              puts line.encode('UTF-8', invalid: :replace, undef: :replace).strip
              output_lines << line
            end
          end

          stderr_thread = Thread.new do
            stderr.each_line do |line|
              puts line.encode('UTF-8', invalid: :replace, undef: :replace).strip
              # Also collect stderr if needed, but for now just stream
            end
          end

          stdout_thread.join
          stderr_thread.join

          exit_status = wait_thr.value
          if exit_status.success?
            puts "✅ Completed (exit #{exit_status.exitstatus})"
          else
            puts "❌ Failed (exit #{exit_status.exitstatus})"
          end
        end

        output = output_lines.join
        cleaned = clean_opencode_output(output)

        Logger.debug('Command output captured', { operation: operation, output_length: cleaned.length })
        cleaned
      rescue StandardError => e
        puts "❌ Error: #{e.message}"
        puts e.backtrace.first(3).join("\n")
        Logger.error('Command capture exception', { operation: operation, error: e.message })
        nil
      end

      def clean_opencode_output(output)
        return '' if output.nil? || output.strip.empty?

        output
          .gsub(/\x1b\[[0-9;]*[a-zA-Z]/, '')
          .gsub(/\n{3,}/, "\n\n")
          .strip
      end

      def safe_system_command(command, operation)
        Logger.debug("Executing command: #{command}", { operation: operation })

        result = system(command)

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
        return nil if json_string.nil?

        # Encode to UTF-8 first to handle binary data safely
        safe_string = json_string.encode('UTF-8', 'binary', invalid: :replace, undef: :replace, replace: '')
        return nil if safe_string.strip.empty?

        with_error_handling(context) do
          cleaned = safe_string.strip

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
