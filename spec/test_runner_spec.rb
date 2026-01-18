# frozen_string_literal: true

require 'spec_helper'

RSpec.describe Ralph::TestRunner do
  describe '.run' do
    before do
      Ralph::TestRunner::TEST_COMMANDS.each do |cmd|
        binary = cmd.split.first
        allow_any_instance_of(Kernel).to receive(:system)
          .with("which #{binary} > /dev/null 2>&1")
          .and_return(false)
      end
    end

    context 'when no test framework is detected' do
      it 'returns true to allow continuation' do
        expect(described_class.run).to be true
      end

      it 'logs warning' do
        expect(Ralph::Logger).to receive(:warn).with('No test framework detected')
        described_class.run
      end
    end

    context 'when npm is available' do
      before do
        allow_any_instance_of(Kernel).to receive(:system)
          .with('which npm > /dev/null 2>&1')
          .and_return(true)
      end

      it 'runs npm test and returns result' do
        allow_any_instance_of(Kernel).to receive(:system)
          .with('npm test')
          .and_return(true)

        expect(described_class.run).to be true
      end

      it 'logs test execution' do
        allow_any_instance_of(Kernel).to receive(:system)
          .with('npm test')
          .and_return(true)

        expect(Ralph::Logger).to receive(:info)
          .with('Tests executed', { command: 'npm test', result: true })

        described_class.run
      end

      it 'returns false when tests fail' do
        allow_any_instance_of(Kernel).to receive(:system)
          .with('npm test')
          .and_return(false)

        expect(described_class.run).to be false
      end
    end

    context 'when pytest is available' do
      before do
        allow_any_instance_of(Kernel).to receive(:system)
          .with('which pytest > /dev/null 2>&1')
          .and_return(true)
      end

      it 'runs pytest' do
        expect_any_instance_of(Kernel).to receive(:system)
          .with('pytest')
          .and_return(true)

        described_class.run
      end
    end

    context 'when multiple frameworks are available' do
      before do
        allow_any_instance_of(Kernel).to receive(:system)
          .with('which npm > /dev/null 2>&1')
          .and_return(true)
        allow_any_instance_of(Kernel).to receive(:system)
          .with('which pytest > /dev/null 2>&1')
          .and_return(true)
      end

      it 'uses the first detected framework' do
        expect_any_instance_of(Kernel).to receive(:system)
          .with('npm test')
          .and_return(true)

        expect_any_instance_of(Kernel).not_to receive(:system)
          .with('pytest')

        described_class.run
      end
    end
  end

  describe 'TEST_COMMANDS' do
    it 'includes common test runners' do
      expect(described_class::TEST_COMMANDS).to include('npm test')
      expect(described_class::TEST_COMMANDS).to include('pytest')
      expect(described_class::TEST_COMMANDS).to include('cargo test')
      expect(described_class::TEST_COMMANDS).to include('go test')
    end
  end
end
