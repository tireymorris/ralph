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

      def capture_command_output(prompt, operation, timeout: nil)
        output_lines = []
        timeout ||= Ralph::Config.get(:command_timeout)

        Open3.popen3(*build_opencode_command) do |stdin, stdout, stderr, wait_thr|
          stdin.write(prompt)
          stdin.close

          threads = start_output_threads(stdout, stderr, output_lines)
          return nil unless wait_for_completion(wait_thr, threads, timeout, operation)

          print_exit_status(wait_thr.value)
        end

        finalize_output(output_lines, operation)
      rescue StandardError => e
        handle_capture_error(e, operation)
      end

      def build_opencode_command
        cmd = %w[opencode run]
        model = Ralph::Config.get(:model)
        cmd += ['--model', model] if model
        cmd
      end

      def start_output_threads(stdout, stderr, output_lines)
        stdout_thread = Thread.new { stream_output(stdout, output_lines) }
        stderr_thread = Thread.new { stream_output(stderr) }
        { stdout: stdout_thread, stderr: stderr_thread }
      end

      def stream_output(io, collector = nil)
        $stdout.sync = true
        io.each_line do |line|
          cleaned_line = line.encode('UTF-8', invalid: :replace, undef: :replace).strip
          puts cleaned_line
          $stdout.flush
          collector&.push(line)
        end
      end

      def wait_for_completion(wait_thr, threads, timeout, operation)
        if timeout
          return true if wait_with_timeout(wait_thr, threads[:stdout], threads[:stderr], timeout)

          handle_timeout(wait_thr, timeout, operation)
          false
        else
          threads[:stdout].join
          threads[:stderr].join
          true
        end
      end

      def handle_timeout(wait_thr, timeout, operation)
        puts "⏱️ Command timed out after #{timeout}s"
        begin
          Process.kill('TERM', wait_thr.pid)
        rescue StandardError
          nil
        end
        Logger.error('Command timeout', { operation: operation, timeout: timeout })
      end

      def print_exit_status(exit_status)
        if exit_status.success?
          puts "✅ Completed (exit #{exit_status.exitstatus})"
        else
          puts "❌ Failed (exit #{exit_status.exitstatus})"
        end
      end

      def finalize_output(output_lines, operation)
        cleaned = clean_opencode_output(output_lines.join)
        Logger.debug('Command output captured', { operation: operation, output_length: cleaned.length })
        cleaned
      end

      def handle_capture_error(error, operation)
        puts "❌ Error: #{error.message}"
        puts error.backtrace.first(3).join("\n")
        Logger.error('Command capture exception', { operation: operation, error: error.message })
        nil
      end

      def wait_with_timeout(wait_thr, stdout_thread, stderr_thread, timeout)
        deadline = Time.now + timeout

        until wait_thr.join(0.1)
          next unless Time.now > deadline

          begin
            stdout_thread.kill
          rescue StandardError
            nil
          end
          begin
            stderr_thread.kill
          rescue StandardError
            nil
          end
          return false
        end

        stdout_thread.join(1)
        stderr_thread.join(1)
        true
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
