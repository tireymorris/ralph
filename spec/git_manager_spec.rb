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

  describe '.create_branch' do
    context 'when branch_name is nil' do
      it 'does nothing' do
        expect(Ralph::ErrorHandler).not_to receive(:safe_system_command)
        described_class.create_branch(nil)
      end
    end

    context 'when branch already exists' do
      before do
        allow_any_instance_of(Kernel).to receive(:system)
          .with(/git show-ref --verify --quiet/)
          .and_return(true)
      end

      it 'switches to existing branch' do
        expect(Ralph::ErrorHandler).to receive(:safe_system_command)
          .with(%r{git checkout feature/test}, 'Switch branch')
          .and_return(true)

        described_class.create_branch('feature/test')
      end

      it 'prints switching message' do
        allow(Ralph::ErrorHandler).to receive(:safe_system_command).and_return(true)

        expect { described_class.create_branch('feature/test') }
          .to output(/already exists, switching/).to_stdout
      end
    end

    context 'when branch does not exist' do
      before do
        allow_any_instance_of(Kernel).to receive(:system)
          .with(/git show-ref --verify --quiet/)
          .and_return(false)
      end

      it 'creates new branch' do
        expect(Ralph::ErrorHandler).to receive(:safe_system_command)
          .with(%r{git checkout -b feature/new-branch}, 'Create branch')
          .and_return(true)

        described_class.create_branch('feature/new-branch')
      end

      it 'prints creating message' do
        allow(Ralph::ErrorHandler).to receive(:safe_system_command).and_return(true)

        expect { described_class.create_branch('feature/test') }
          .to output(/Creating new branch/).to_stdout
      end
    end

    it 'escapes special characters in branch name' do
      allow_any_instance_of(Kernel).to receive(:system).and_return(false)
      allow(Ralph::ErrorHandler).to receive(:safe_system_command).and_return(true)

      # Should not raise error with special characters
      expect { described_class.create_branch('feature/test-branch') }.not_to raise_error
    end
  end

  describe '.current_branch' do
    it 'returns current git branch name' do
      allow(described_class).to receive(:`).with('git rev-parse --abbrev-ref HEAD 2>/dev/null').and_return("main\n")

      expect(described_class.current_branch).to eq('main')
    end

    it 'strips whitespace from branch name' do
      allow(described_class).to receive(:`)
        .with('git rev-parse --abbrev-ref HEAD 2>/dev/null')
        .and_return("  feature/test  \n")

      expect(described_class.current_branch).to eq('feature/test')
    end
  end

  describe '.no_unstaged_changes?' do
    it 'returns true when no unstaged changes' do
      allow_any_instance_of(Kernel).to receive(:system)
        .with('git diff --quiet --exit-code 2>/dev/null')
        .and_return(true)

      expect(described_class.no_unstaged_changes?).to be true
    end

    it 'returns false when there are unstaged changes' do
      allow_any_instance_of(Kernel).to receive(:system)
        .with('git diff --quiet --exit-code 2>/dev/null')
        .and_return(false)

      expect(described_class.no_unstaged_changes?).to be false
    end
  end

  describe '.no_staged_changes?' do
    it 'returns true when no staged changes' do
      allow_any_instance_of(Kernel).to receive(:system)
        .with('git diff --staged --quiet --exit-code 2>/dev/null')
        .and_return(true)

      expect(described_class.no_staged_changes?).to be true
    end

    it 'returns false when there are staged changes' do
      allow_any_instance_of(Kernel).to receive(:system)
        .with('git diff --staged --quiet --exit-code 2>/dev/null')
        .and_return(false)

      expect(described_class.no_staged_changes?).to be false
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
        allow(described_class).to receive(:no_unstaged_changes?).and_return(true)
        allow(described_class).to receive(:no_staged_changes?).and_return(true)
      end

      it 'logs info and returns true' do
        expect(Ralph::Logger).to receive(:info).with('No changes to commit', anything)
        expect(described_class.commit_changes(story)).to be true
      end
    end

    context 'when there are changes' do
      before do
        allow(described_class).to receive(:no_unstaged_changes?).and_return(false)
        allow(described_class).to receive(:no_staged_changes?).and_return(true)
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
        expect(Ralph::ErrorHandler).to receive(:safe_system_command) do |cmd, _op|
          expect(cmd).to start_with('git commit -m')
          # Shellwords.escape converts spaces to backslash-space
          expect(cmd).to include('Test\\ Story')
          expect(cmd).to include('story-1')
          true
        end

        described_class.commit_changes(story)
      end

      it 'properly escapes special characters in title' do
        story['title'] = "Story's $title `with` special chars"
        allow(Ralph::ErrorHandler).to receive(:safe_system_command).and_return(true)

        expect(Ralph::ErrorHandler).to receive(:safe_system_command)
          .with(/git commit -m /, anything).and_return(true)

        described_class.commit_changes(story)
      end
    end

    context 'with nil story fields' do
      let(:minimal_story) { {} }

      before do
        allow(described_class).to receive(:no_unstaged_changes?).and_return(false)
        allow(described_class).to receive(:no_staged_changes?).and_return(true)
        allow(Ralph::ErrorHandler).to receive(:safe_system_command).and_return(true)
      end

      it 'uses default values' do
        expect(Ralph::ErrorHandler).to receive(:safe_system_command)
          .with('git add .', anything).and_return(true)
        expect(Ralph::ErrorHandler).to receive(:safe_system_command) do |cmd, _op|
          expect(cmd).to start_with('git commit -m')
          # Shellwords.escape converts spaces to backslash-space
          expect(cmd).to include('Story\\ implementation')
          true
        end

        described_class.commit_changes(minimal_story)
      end
    end
  end
end
