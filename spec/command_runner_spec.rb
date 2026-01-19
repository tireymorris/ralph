# frozen_string_literal: true

require 'spec_helper'

RSpec.describe Ralph::CommandRunner do
  describe '.build_opencode_command' do
    it 'returns base command without model' do
      allow(Ralph::Config).to receive(:get).with(:model).and_return(nil)
      expect(described_class.send(:build_opencode_command)).to eq(%w[opencode run])
    end

    it 'includes model when configured' do
      allow(Ralph::Config).to receive(:get).with(:model).and_return('opencode/grok-code')
      expect(described_class.send(:build_opencode_command)).to eq(%w[opencode run --model opencode/grok-code])
    end
  end

  describe '.clean_output' do
    it 'returns empty string for nil' do
      expect(described_class.send(:clean_output, nil)).to eq('')
    end

    it 'returns empty string for whitespace only' do
      expect(described_class.send(:clean_output, '   ')).to eq('')
    end

    it 'removes ANSI escape codes' do
      input = "\e[32mGreen text\e[0m"
      expect(described_class.send(:clean_output, input)).to eq('Green text')
    end

    it 'collapses multiple newlines' do
      input = "line1\n\n\n\nline2"
      expect(described_class.send(:clean_output, input)).to eq("line1\n\nline2")
    end

    it 'strips leading and trailing whitespace' do
      input = "  \n  content  \n  "
      expect(described_class.send(:clean_output, input)).to eq('content')
    end
  end

  describe '.print_exit_status' do
    it 'prints success message for successful exit' do
      status = double('status', success?: true, exitstatus: 0)
      expect { described_class.send(:print_exit_status, status) }
        .to output(/Completed \(exit 0\)/).to_stdout
    end

    it 'prints failure message for failed exit' do
      status = double('status', success?: false, exitstatus: 1)
      expect { described_class.send(:print_exit_status, status) }
        .to output(/Failed \(exit 1\)/).to_stdout
    end
  end

  describe '.safe_system' do
    it 'returns true when command succeeds' do
      allow(Kernel).to receive(:system).and_return(true)
      expect(described_class.safe_system('echo test', 'test op')).to be true
    end

    it 'returns false when command fails' do
      allow(Kernel).to receive(:system).and_return(false)
      expect(described_class.safe_system('false', 'test op')).to be false
    end

    it 'returns false when command not found' do
      allow(Kernel).to receive(:system).and_return(nil)
      expect(described_class.safe_system('nonexistent', 'test op')).to be false
    end

    it 'returns false when exception raised' do
      allow(Kernel).to receive(:system).and_raise(StandardError.new('error'))
      expect(described_class.safe_system('cmd', 'test op')).to be false
    end
  end
end
