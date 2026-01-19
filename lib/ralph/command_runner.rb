# frozen_string_literal: true

require 'open3'

module Ralph
  class CommandRunner
    class << self
      def capture_opencode_output(prompt, operation, timeout: nil)
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
        handle_error(e, operation)
      end

      def safe_system(command, operation)
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

      private

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

      def wait_with_timeout(wait_thr, stdout_thread, stderr_thread, timeout)
        deadline = Time.now + timeout

        loop do
          break if wait_thr.join(0.1)
          return timeout_threads(stdout_thread, stderr_thread) if Time.now > deadline
        end

        stdout_thread.join(1)
        stderr_thread.join(1)
        true
      end

      def timeout_threads(stdout_thread, stderr_thread)
        safe_kill(stdout_thread)
        safe_kill(stderr_thread)
        false
      end

      def safe_kill(thread)
        thread.kill
      rescue StandardError
        nil
      end

      def handle_timeout(wait_thr, timeout, operation)
        puts "⏱️ Command timed out after #{timeout}s"
        safe_process_kill(wait_thr.pid)
        Logger.error('Command timeout', { operation: operation, timeout: timeout })
      end

      def safe_process_kill(pid)
        Process.kill('TERM', pid)
      rescue StandardError
        nil
      end

      def print_exit_status(exit_status)
        if exit_status.success?
          puts "✅ Completed (exit #{exit_status.exitstatus})"
        else
          puts "❌ Failed (exit #{exit_status.exitstatus})"
        end
      end

      def finalize_output(output_lines, operation)
        cleaned = clean_output(output_lines.join)
        Logger.debug('Command output captured', { operation: operation, output_length: cleaned.length })
        cleaned
      end

      def clean_output(output)
        return '' if output.nil?

        encoded = output.encode('UTF-8', 'binary', invalid: :replace, undef: :replace, replace: '')
        return '' if encoded.strip.empty?

        encoded
          .gsub(/\x1b\[[0-9;]*[a-zA-Z]/, '')
          .gsub(/\n{3,}/, "\n\n")
          .strip
      end

      def handle_error(error, operation)
        puts "❌ Error: #{error.message}"
        puts error.backtrace.first(3).join("\n")
        Logger.error('Command capture exception', { operation: operation, error: error.message })
        nil
      end
    end
  end
end
