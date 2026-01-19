# frozen_string_literal: true

require 'spec_helper'

RSpec.describe Ralph::ErrorHandler do
  describe '.log_error' do
    it 'logs error with context' do
      error = StandardError.new('Test error')
      error.set_backtrace(%w[line1 line2 line3 line4])

      expect(Ralph::Logger).to receive(:error).with(
        'Error in test operation',
        hash_including(
          error_class: 'StandardError',
          error_message: 'Test error',
          backtrace: %w[line1 line2 line3]
        )
      )

      described_class.log_error('test operation', error, { extra: 'context' })
    end

    it 'handles error with nil backtrace' do
      error = StandardError.new('Test error')
      # Don't set backtrace - it will be nil

      expect(Ralph::Logger).to receive(:error).with(
        'Error in test operation',
        hash_including(
          error_class: 'StandardError',
          error_message: 'Test error',
          backtrace: nil
        )
      )

      described_class.log_error('test operation', error, {})
    end
  end

  describe '.with_error_handling' do
    it 'returns block result on success' do
      result = described_class.with_error_handling('test') { 'success' }
      expect(result).to eq('success')
    end

    it 'logs debug message on success' do
      expect(Ralph::Logger).to receive(:debug).with('Completed test', { key: 'value' })
      described_class.with_error_handling('test', { key: 'value' }) { 'success' }
    end

    it 'returns nil on error' do
      result = described_class.with_error_handling('test') { raise 'boom' }
      expect(result).to be_nil
    end

    it 'logs error on exception' do
      expect(Ralph::Logger).to receive(:error).with(
        'Error in test',
        hash_including(error_message: 'boom')
      )

      described_class.with_error_handling('test') { raise 'boom' }
    end
  end

  describe '.clean_opencode_output' do
    it 'returns empty string for nil' do
      expect(described_class.clean_opencode_output(nil)).to eq('')
    end

    it 'returns empty string for whitespace only' do
      expect(described_class.clean_opencode_output('   ')).to eq('')
    end

    it 'removes ANSI escape codes' do
      input = "\e[32mGreen text\e[0m"
      expect(described_class.clean_opencode_output(input)).to eq('Green text')
    end

    it 'collapses multiple newlines' do
      input = "line1\n\n\n\nline2"
      expect(described_class.clean_opencode_output(input)).to eq("line1\n\nline2")
    end

    it 'strips leading and trailing whitespace' do
      input = "  content  \n"
      expect(described_class.clean_opencode_output(input)).to eq('content')
    end
  end

  describe '.safe_system_command' do
    it 'returns true on successful command' do
      allow_any_instance_of(Kernel).to receive(:system).with('echo test').and_return(true)

      result = described_class.safe_system_command('echo test', 'test op')
      expect(result).to be true
    end

    it 'returns false when command fails' do
      allow_any_instance_of(Kernel).to receive(:system).with('false').and_return(false)

      result = described_class.safe_system_command('false', 'test op')
      expect(result).to be false
    end

    it 'returns false when command not found' do
      allow_any_instance_of(Kernel).to receive(:system).with('nonexistent').and_return(nil)

      result = described_class.safe_system_command('nonexistent', 'test op')
      expect(result).to be false
    end

    it 'returns false on exception' do
      allow_any_instance_of(Kernel).to receive(:system).and_raise(StandardError.new('error'))

      result = described_class.safe_system_command('cmd', 'test op')
      expect(result).to be false
    end
  end

  describe '.parse_json_safely' do
    it 'returns nil for nil input' do
      expect(described_class.parse_json_safely(nil)).to be_nil
    end

    it 'returns nil for empty string' do
      expect(described_class.parse_json_safely('')).to be_nil
    end

    it 'returns nil for whitespace only' do
      expect(described_class.parse_json_safely('   ')).to be_nil
    end

    it 'parses valid JSON' do
      result = described_class.parse_json_safely('{"key": "value"}')
      expect(result).to eq({ 'key' => 'value' })
    end

    it 'extracts JSON from surrounding text' do
      input = 'Some text {"key": "value"} more text'
      result = described_class.parse_json_safely(input)
      expect(result).to eq({ 'key' => 'value' })
    end

    it 'returns nil for invalid JSON' do
      expect(described_class.parse_json_safely('not json')).to be_nil
    end

    it 'returns nil for JSON array (expects Hash)' do
      expect(described_class.parse_json_safely('[1, 2, 3]')).to be_nil
    end

    it 'handles UTF-8 encoding issues' do
      input = "{\"key\": \"value\xC0\xC1\"}"
      result = described_class.parse_json_safely(input)
      expect(result).to be_a(Hash)
    end
  end

  describe '.capture_command_output' do
    let(:mock_stdin) { instance_double(IO, write: nil, close: nil) }
    let(:mock_stdout) { instance_double(IO) }
    let(:mock_stderr) { instance_double(IO) }
    let(:mock_wait_thr) { instance_double(Process::Waiter, value: mock_status, pid: 12_345) }
    let(:mock_status) { instance_double(Process::Status, success?: true, exitstatus: 0) }

    before do
      allow(mock_stdout).to receive(:each_line).and_yield("output line\n")
      allow(mock_stderr).to receive(:each_line)
    end

    it 'sends prompt to stdin' do
      allow(Open3).to receive(:popen3).and_yield(mock_stdin, mock_stdout, mock_stderr, mock_wait_thr)

      expect(mock_stdin).to receive(:write).with('test prompt')
      expect(mock_stdin).to receive(:close)

      described_class.capture_command_output('test prompt', 'test op')
    end

    it 'returns cleaned output on success' do
      allow(Open3).to receive(:popen3).and_yield(mock_stdin, mock_stdout, mock_stderr, mock_wait_thr)

      result = described_class.capture_command_output('prompt', 'op')
      expect(result).to eq('output line')
    end

    it 'returns nil on exception' do
      allow(Open3).to receive(:popen3).and_raise(StandardError.new('connection failed'))

      result = described_class.capture_command_output('prompt', 'op')
      expect(result).to be_nil
    end

    it 'uses configured model' do
      Ralph::Config.set(:model, 'opencode/grok-code')

      expect(Open3).to receive(:popen3).with('opencode', 'run', '--model', 'opencode/grok-code')
                                       .and_yield(mock_stdin, mock_stdout, mock_stderr, mock_wait_thr)

      described_class.capture_command_output('prompt', 'op')
    end

    context 'with failed exit status' do
      let(:mock_status) { instance_double(Process::Status, success?: false, exitstatus: 1) }

      it 'prints failed message' do
        allow(Open3).to receive(:popen3).and_yield(mock_stdin, mock_stdout, mock_stderr, mock_wait_thr)

        expect { described_class.capture_command_output('prompt', 'op') }
          .to output(/Failed \(exit 1\)/).to_stdout
      end
    end

    context 'with timeout' do
      it 'returns nil when wait_for_completion fails' do
        allow(Open3).to receive(:popen3).and_yield(mock_stdin, mock_stdout, mock_stderr, mock_wait_thr)
        allow(described_class).to receive(:start_output_threads).and_return({ stdout: double, stderr: double })
        allow(described_class).to receive(:wait_for_completion).and_return(false)

        result = described_class.capture_command_output('prompt', 'op', timeout: 1)
        expect(result).to be_nil
      end
    end
  end

  describe '.build_opencode_command' do
    it 'returns base command without model' do
      Ralph::Config.set(:model, nil)
      expect(described_class.build_opencode_command).to eq(%w[opencode run])
    end

    it 'includes model when configured' do
      Ralph::Config.set(:model, 'opencode/grok-code')
      expect(described_class.build_opencode_command).to eq(%w[opencode run --model opencode/grok-code])
    end
  end

  describe '.wait_for_completion' do
    let(:mock_wait_thr) { instance_double(Process::Waiter, pid: 12_345) }
    let(:mock_stdout_thread) { instance_double(Thread) }
    let(:mock_stderr_thread) { instance_double(Thread) }
    let(:threads) { { stdout: mock_stdout_thread, stderr: mock_stderr_thread } }

    context 'without timeout' do
      it 'joins threads and returns true' do
        expect(mock_stdout_thread).to receive(:join)
        expect(mock_stderr_thread).to receive(:join)

        result = described_class.wait_for_completion(mock_wait_thr, threads, nil, 'op')
        expect(result).to be true
      end
    end

    context 'with timeout that succeeds' do
      it 'returns true when command completes in time' do
        allow(described_class).to receive(:wait_with_timeout).and_return(true)

        result = described_class.wait_for_completion(mock_wait_thr, threads, 30, 'op')
        expect(result).to be true
      end
    end

    context 'with timeout that fails' do
      it 'returns false and handles timeout' do
        allow(described_class).to receive(:wait_with_timeout).and_return(false)
        allow(Process).to receive(:kill)

        expect { described_class.wait_for_completion(mock_wait_thr, threads, 30, 'op') }
          .to output(/timed out/).to_stdout
      end

      it 'returns false' do
        allow(described_class).to receive(:wait_with_timeout).and_return(false)
        allow(Process).to receive(:kill)

        result = described_class.wait_for_completion(mock_wait_thr, threads, 30, 'op')
        expect(result).to be false
      end
    end
  end

  describe '.handle_timeout' do
    let(:mock_wait_thr) { instance_double(Process::Waiter, pid: 12_345) }

    it 'kills the process and logs error' do
      expect(Process).to receive(:kill).with('TERM', 12_345)

      expect { described_class.handle_timeout(mock_wait_thr, 30, 'op') }
        .to output(/timed out after 30s/).to_stdout
    end

    it 'handles kill errors gracefully' do
      allow(Process).to receive(:kill).and_raise(Errno::ESRCH)

      expect { described_class.handle_timeout(mock_wait_thr, 30, 'op') }
        .to output(/timed out/).to_stdout
    end
  end

  describe '.print_exit_status' do
    it 'prints success message for successful exit' do
      status = instance_double(Process::Status, success?: true, exitstatus: 0)
      expect { described_class.print_exit_status(status) }
        .to output(/Completed \(exit 0\)/).to_stdout
    end

    it 'prints failure message for failed exit' do
      status = instance_double(Process::Status, success?: false, exitstatus: 1)
      expect { described_class.print_exit_status(status) }
        .to output(/Failed \(exit 1\)/).to_stdout
    end
  end

  describe '.wait_with_timeout' do
    it 'returns true when process completes in time' do
      wait_thr = instance_double(Process::Waiter)
      stdout_thread = instance_double(Thread)
      stderr_thread = instance_double(Thread)

      allow(wait_thr).to receive(:join).with(0.1).and_return(wait_thr)
      allow(stdout_thread).to receive(:join).with(1)
      allow(stderr_thread).to receive(:join).with(1)

      result = described_class.wait_with_timeout(wait_thr, stdout_thread, stderr_thread, 30)
      expect(result).to be true
    end

    it 'returns false and kills threads when timeout exceeded' do
      wait_thr = instance_double(Process::Waiter)
      stdout_thread = instance_double(Thread)
      stderr_thread = instance_double(Thread)

      call_count = 0
      allow(wait_thr).to receive(:join).with(0.1) do
        call_count += 1
        nil
      end
      allow(Time).to receive(:now).and_return(Time.at(0), Time.at(0), Time.at(100))
      allow(stdout_thread).to receive(:kill)
      allow(stderr_thread).to receive(:kill)

      result = described_class.wait_with_timeout(wait_thr, stdout_thread, stderr_thread, 1)
      expect(result).to be false
    end

    it 'handles thread kill errors gracefully' do
      wait_thr = instance_double(Process::Waiter)
      stdout_thread = instance_double(Thread)
      stderr_thread = instance_double(Thread)

      allow(wait_thr).to receive(:join).with(0.1).and_return(nil)
      allow(Time).to receive(:now).and_return(Time.at(0), Time.at(100))
      allow(stdout_thread).to receive(:kill).and_raise(StandardError)
      allow(stderr_thread).to receive(:kill).and_raise(StandardError)

      result = described_class.wait_with_timeout(wait_thr, stdout_thread, stderr_thread, 1)
      expect(result).to be false
    end
  end

  describe '.stream_output' do
    it 'streams lines without collector' do
      io = instance_double(IO)
      allow(io).to receive(:each_line).and_yield("line1\n")

      expect { described_class.stream_output(io) }
        .to output(/line1/).to_stdout
    end

    it 'streams lines with collector' do
      io = instance_double(IO)
      collector = []
      allow(io).to receive(:each_line).and_yield("line1\n")

      described_class.stream_output(io, collector)
      expect(collector).to eq(["line1\n"])
    end
  end

  describe '.finalize_output' do
    it 'cleans and returns output' do
      result = described_class.finalize_output(%W[line1\n line2\n], 'op')
      expect(result).to eq("line1\nline2")
    end
  end

  describe '.handle_capture_error' do
    it 'prints error and returns nil' do
      error = StandardError.new('test error')
      error.set_backtrace(%w[line1 line2 line3 line4])

      result = nil
      expect { result = described_class.handle_capture_error(error, 'op') }
        .to output(/Error: test error/).to_stdout

      expect(result).to be_nil
    end
  end
end
