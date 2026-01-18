# frozen_string_literal: true

require 'spec_helper'

RSpec.describe Ralph::GitManager do
  describe '.validate_repository' do
    context 'in a git repository' do
      it 'does not raise error' do
        allow_any_instance_of(Kernel).to receive(:system)
          .with('git rev-parse --git-dir > /dev/null 2>&1')
          .and_return(true)

        expect { described_class.validate_repository }.not_to raise_error
      end
    end

    context 'not in a git repository' do
      it 'logs error and returns nil' do
        allow_any_instance_of(Kernel).to receive(:system)
          .with('git rev-parse --git-dir > /dev/null 2>&1')
          .and_return(false)

        result = described_class.validate_repository
        expect(result).to be_nil
      end
    end
  end

  describe '.commit_changes' do
    let(:story) do
      {
        'id' => 'story-1',
        'title' => 'Test Story',
        'description' => 'Test description'
      }
    end

    context 'when there are no changes' do
      before do
        allow_any_instance_of(Kernel).to receive(:system).with('git diff --quiet --exit-code 2>/dev/null').and_return(true)
        allow_any_instance_of(Kernel).to receive(:system).with('git diff --staged --quiet --exit-code 2>/dev/null').and_return(true)
      end

      it 'logs info and returns true' do
        expect(Ralph::Logger).to receive(:info).with('No changes to commit', anything)
        expect(described_class.commit_changes(story)).to be true
      end
    end

    context 'when there are changes' do
      before do
        allow_any_instance_of(Kernel).to receive(:system).with('git diff --quiet --exit-code 2>/dev/null').and_return(false)
        allow_any_instance_of(Kernel).to receive(:system).with('git diff --staged --quiet --exit-code 2>/dev/null').and_return(true)
      end

      it 'stages and commits changes' do
        expect(Ralph::ErrorHandler).to receive(:safe_system_command)
          .with('git add .', 'Stage changes').and_return(true)
        expect(Ralph::ErrorHandler).to receive(:safe_system_command)
          .with(/git commit -m/, 'Commit changes').and_return(true)

        described_class.commit_changes(story)
      end

      it 'includes story details in commit message' do
        expect(Ralph::ErrorHandler).to receive(:safe_system_command)
          .with('git add .', anything).and_return(true)
        expect(Ralph::ErrorHandler).to receive(:safe_system_command)
          .with(/feat: Test Story.*Story: story-1/m, anything).and_return(true)

        described_class.commit_changes(story)
      end

      it 'escapes single quotes in title' do
        story['title'] = "Story's title"
        allow(Ralph::ErrorHandler).to receive(:safe_system_command).and_return(true)

        expect(Ralph::ErrorHandler).to receive(:safe_system_command)
          .with(/Story''s title/, anything).and_return(true)

        described_class.commit_changes(story)
      end
    end

    context 'with nil story fields' do
      let(:minimal_story) { {} }

      before do
        allow_any_instance_of(Kernel).to receive(:system).with('git diff --quiet --exit-code 2>/dev/null').and_return(false)
        allow_any_instance_of(Kernel).to receive(:system).with('git diff --staged --quiet --exit-code 2>/dev/null').and_return(true)
        allow(Ralph::ErrorHandler).to receive(:safe_system_command).and_return(true)
      end

      it 'uses default values' do
        expect(Ralph::ErrorHandler).to receive(:safe_system_command)
          .with(/feat: Story implementation/, anything).and_return(true)

        described_class.commit_changes(minimal_story)
      end
    end
  end
end
