# frozen_string_literal: true

require 'spec_helper'

RSpec.describe Ralph::Logger do
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

    it 'handles nil argument explicitly' do
      Ralph::Config.set(:log_level, :warn)
      described_class.configure(nil)
      expect(described_class.level).to eq(2)
    end

    it 'falls back to info when both level and config are nil' do
      Ralph::Config.set(:log_level, nil)
      described_class.configure(nil)
      expect(described_class.level).to eq(1) # info level
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

    it 'respects log level filtering' do
      described_class.configure(:error)
      expect { described_class.log(:debug, 'Should not appear') }
        .not_to output.to_stdout
    end

    it 'handles nil level gracefully' do
      described_class.instance_variable_set(:@level, nil)
      expect { described_class.log(:info, 'Test') }.not_to raise_error
    end

    it 'handles errors during logging gracefully' do
      described_class.configure(:debug)
      allow(Time).to receive(:now).and_raise(StandardError.new('time error'))

      expect { described_class.log(:info, 'Test') }
        .to output(/Logger error: time error/).to_stdout
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
end
