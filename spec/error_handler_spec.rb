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
    let(:mock_wait_thr) { instance_double(Process::Waiter, value: mock_status) }
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
  end
end
