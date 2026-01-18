# frozen_string_literal: true

require 'spec_helper'

RSpec.describe Ralph::Config do
  before(:each) do
    described_class.reset!
  end

  describe '.load' do
    it 'returns default configuration' do
      config = described_class.load
      expect(config).to be_a(Hash)
      expect(config[:max_iterations]).to eq(50)
      expect(config[:log_level]).to eq(:info)
    end

    it 'memoizes the configuration' do
      first_load = described_class.load
      second_load = described_class.load
      expect(first_load).to be(second_load)
    end

    context 'with config file' do
      let(:config_file) { 'ralph.config.json' }

      after { File.delete(config_file) if File.exist?(config_file) }

      it 'merges file configuration with defaults' do
        File.write(config_file, '{"max_iterations": 100}')
        described_class.reset!

        config = described_class.load
        expect(config[:max_iterations]).to eq(100)
        expect(config[:log_level]).to eq(:info)
      end

      it 'handles invalid JSON gracefully' do
        File.write(config_file, 'not valid json')
        described_class.reset!

        expect { described_class.load }.not_to raise_error
        expect(described_class.get(:max_iterations)).to eq(50)
      end
    end
  end

  describe '.get' do
    it 'returns value for existing key' do
      expect(described_class.get(:log_file)).to eq('ralph.log')
    end

    it 'returns nil for non-existent key' do
      expect(described_class.get(:nonexistent)).to be_nil
    end
  end

  describe '.set' do
    it 'sets a configuration value' do
      described_class.set(:max_iterations, 25)
      expect(described_class.get(:max_iterations)).to eq(25)
    end

    it 'allows setting new keys' do
      described_class.set(:custom_key, 'custom_value')
      expect(described_class.get(:custom_key)).to eq('custom_value')
    end
  end

  describe '.reset!' do
    it 'clears cached configuration' do
      described_class.set(:max_iterations, 999)
      described_class.reset!
      expect(described_class.get(:max_iterations)).to eq(50)
    end
  end

  describe 'DEFAULTS' do
    it 'contains all required keys' do
      required_keys = %i[
        model git_timeout test_timeout max_iterations
        log_level log_file progress_file prd_file agents_file
        retry_attempts retry_delay
      ]

      required_keys.each do |key|
        expect(described_class::DEFAULTS).to have_key(key)
      end
    end

    it 'is frozen' do
      expect(described_class::DEFAULTS).to be_frozen
    end
  end
end
