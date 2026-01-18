# frozen_string_literal: true

require 'spec_helper'

RSpec.describe Ralph::Logger do
  let(:log_file) { 'ralph.log' }

  before(:each) do
    File.delete(log_file) if File.exist?(log_file)
  end

  after(:each) do
    File.delete(log_file) if File.exist?(log_file)
  end

  describe '.configure' do
    it 'sets log level from symbol' do
      described_class.configure(:debug)
      expect(described_class.level).to eq(0)
    end

    it 'sets log level from string' do
      described_class.configure('warn')
      expect(described_class.level).to eq(2)
    end

    it 'defaults to info level when invalid' do
      described_class.configure(:invalid)
      expect(described_class.level).to eq(1)
    end

    it 'uses config value when no argument provided' do
      Ralph::Config.set(:log_level, :debug)
      described_class.configure
      expect(described_class.level).to eq(0)
    end
  end

  describe '.log' do
    before { described_class.configure(:debug) }

    it 'writes formatted message to stdout' do
      expect { described_class.log(:info, 'Test message') }
        .to output(/INFO: Test message/).to_stdout
    end

    it 'includes context in output' do
      expect { described_class.log(:info, 'Test', { key: 'value' }) }
        .to output(/key.*value/).to_stdout
    end

    it 'writes to log file after flush' do
      described_class.log(:info, 'File test')
      described_class.flush
      expect(File.read(log_file)).to include('File test')
    end

    it 'respects log level filtering' do
      described_class.configure(:error)
      expect { described_class.log(:debug, 'Should not appear') }
        .not_to output.to_stdout
    end

    it 'handles nil level gracefully' do
      described_class.instance_variable_set(:@level, nil)
      expect { described_class.log(:info, 'Test') }.not_to raise_error
    end
  end

  describe 'convenience methods' do
    before { described_class.configure(:debug) }

    describe '.debug' do
      it 'logs at debug level' do
        expect { described_class.debug('Debug msg') }
          .to output(/DEBUG: Debug msg/).to_stdout
      end
    end

    describe '.info' do
      it 'logs at info level' do
        expect { described_class.info('Info msg') }
          .to output(/INFO: Info msg/).to_stdout
      end
    end

    describe '.warn' do
      it 'logs at warn level' do
        expect { described_class.warn('Warn msg') }
          .to output(/WARN: Warn msg/).to_stdout
      end
    end

    describe '.error' do
      it 'logs at error level' do
        expect { described_class.error('Error msg') }
          .to output(/ERROR: Error msg/).to_stdout
      end
    end
  end

  describe 'LOG_LEVELS' do
    it 'defines all standard levels' do
      expect(described_class::LOG_LEVELS).to eq({
                                                  debug: 0, info: 1, warn: 2, error: 3
                                                })
    end
  end

  describe '.flush' do
    before { described_class.configure(:debug) }

    it 'writes buffered messages to file' do
      described_class.log(:info, 'Buffered message 1')
      described_class.log(:warn, 'Buffered message 2')
      described_class.flush
      content = File.read(log_file)
      expect(content).to include('Buffered message 1')
      expect(content).to include('Buffered message 2')
    end

    it 'clears the buffer after flushing' do
      described_class.log(:info, 'Test message')
      described_class.flush
      # Buffer should be empty, so another flush does nothing new
      initial_size = File.size(log_file)
      described_class.flush
      expect(File.size(log_file)).to eq(initial_size)
    end
  end
end
