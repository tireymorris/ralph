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
end
