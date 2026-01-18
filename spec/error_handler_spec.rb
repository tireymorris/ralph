# frozen_string_literal: true

require 'spec_helper'

RSpec.describe Ralph::ErrorHandler do
  before { Ralph::Logger.configure(:error) }

  describe '.log_error' do
    it 'logs error details' do
      error = StandardError.new('Test error')
      error.set_backtrace(%w[line1 line2 line3 line4])

      expect(Ralph::Logger).to receive(:error).with(
        'Error in test operation',
        hash_including(
          error_class: 'StandardError',
          error_message: 'Test error'
        )
      )

      described_class.log_error('test operation', error, { extra: 'context' })
    end

    it 'handles errors without backtrace' do
      error = StandardError.new('No backtrace')

      expect { described_class.log_error('op', error) }.not_to raise_error
    end
  end

  describe '.with_error_handling' do
    it 'returns block result on success' do
      result = described_class.with_error_handling('test') { 'success' }
      expect(result).to eq('success')
    end

    it 'returns nil on exception' do
      result = described_class.with_error_handling('test') { raise 'error' }
      expect(result).to be_nil
    end

    it 'logs the error on exception' do
      expect(Ralph::Logger).to receive(:error)
      described_class.with_error_handling('test') { raise 'error' }
    end

    it 'passes context through' do
      expect(Ralph::Logger).to receive(:debug).with('Completed test', { key: 'val' })
      described_class.with_error_handling('test', { key: 'val' }) { 'ok' }
    end
  end

  describe '.clean_opencode_output' do
    it 'returns empty string for nil input' do
      expect(described_class.clean_opencode_output(nil)).to eq('')
    end

    it 'returns empty string for whitespace-only input' do
      expect(described_class.clean_opencode_output('   ')).to eq('')
    end

    it 'removes ANSI color codes' do
      input = "\e[32mGreen text\e[0m"
      expect(described_class.clean_opencode_output(input)).to eq('Green text')
    end

    it 'preserves JSON content' do
      input = '{"key": "value"}'
      result = described_class.clean_opencode_output(input)
      expect(result).to include('{')
    end

    it 'reduces multiple newlines' do
      input = "Line1\n\n\n\n\nLine2"
      result = described_class.clean_opencode_output(input)
      expect(result).to eq("Line1\n\nLine2")
    end

    it 'strips whitespace' do
      input = "  content  \n"
      expect(described_class.clean_opencode_output(input)).to eq('content')
    end
  end

  describe '.safe_system_command' do
    it 'returns true for successful command' do
      result = described_class.safe_system_command('true', 'test')
      expect(result).to be true
    end

    it 'returns false for failed command' do
      result = described_class.safe_system_command('false', 'test')
      expect(result).to be false
    end

    it 'returns false for non-existent command' do
      result = described_class.safe_system_command('nonexistent_command_xyz', 'test')
      expect(result).to be false
    end

    it 'logs debug on success' do
      expect(Ralph::Logger).to receive(:debug).at_least(:twice)
      described_class.safe_system_command('true', 'test')
    end
  end

  describe '.parse_json_safely' do
    it 'returns nil for nil input' do
      expect(described_class.parse_json_safely(nil)).to be_nil
    end

    it 'returns nil for empty string' do
      expect(described_class.parse_json_safely('')).to be_nil
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

    it 'returns nil for non-hash JSON' do
      result = described_class.parse_json_safely('[1, 2, 3]')
      expect(result).to be_nil
    end

    it 'returns nil for invalid JSON' do
      result = described_class.parse_json_safely('not json at all')
      expect(result).to be_nil
    end

    it 'handles UTF-8 encoding issues' do
      input = String.new('{"key": "value"}', encoding: 'ASCII-8BIT')
      result = described_class.parse_json_safely(input)
      expect(result).to eq({ 'key' => 'value' })
    end
  end

  describe '.capture_command_output' do
    it 'writes prompt to temp file and cleans up' do
      allow(Open3).to receive(:popen3).and_yield(
        StringIO.new, StringIO.new("output\n"), StringIO.new,
        double(value: double(exitstatus: 0, success?: true))
      )

      described_class.capture_command_output('test prompt', 'test op')

      temp_files = Dir.glob('.ralph_prompt_*.txt')
      expect(temp_files).to be_empty
    end

    it 'returns nil on error' do
      allow(Open3).to receive(:popen3).and_raise(StandardError.new('test'))

      result = described_class.capture_command_output('prompt', 'op')
      expect(result).to be_nil
    end
  end
end
