# frozen_string_literal: true

require 'spec_helper'

RSpec.describe Ralph::StoryImplementer do
  let(:story) do
    {
      'id' => 'story-1',
      'title' => 'Test Story',
      'description' => 'Implement test feature',
      'acceptance_criteria' => ['Works correctly', 'Has tests'],
      'priority' => 1,
      'passes' => false
    }
  end

  let(:all_requirements) do
    {
      'project_name' => 'Test Project',
      'stories' => [story]
    }
  end

  describe '.implement' do
    before do
      allow(Ralph::TestRunner).to receive(:run).and_return(true)
      allow(Ralph::GitManager).to receive(:commit_changes)
      allow(Ralph::ProgressLogger).to receive(:log_iteration)
    end

    context 'when implementation succeeds' do
      before do
        allow(Ralph::ErrorHandler).to receive(:capture_command_output)
          .and_return('COMPLETED: Feature implemented successfully')
      end

      it 'returns true' do
        result = described_class.implement(story, 1, all_requirements)
        expect(result).to be true
      end

      it 'runs tests' do
        expect(Ralph::TestRunner).to receive(:run).and_return(true)
        described_class.implement(story, 1, all_requirements)
      end

      it 'commits changes after passing tests' do
        expect(Ralph::GitManager).to receive(:commit_changes).with(story)
        described_class.implement(story, 1, all_requirements)
      end

      it 'logs successful iteration' do
        expect(Ralph::ProgressLogger).to receive(:log_iteration).with(1, story, true)
        described_class.implement(story, 1, all_requirements)
      end
    end

    context 'when tests fail' do
      before do
        allow(Ralph::ErrorHandler).to receive(:capture_command_output)
          .and_return('COMPLETED: Done')
        allow(Ralph::TestRunner).to receive(:run).and_return(false)
      end

      it 'returns false' do
        result = described_class.implement(story, 1, all_requirements)
        expect(result).to be false
      end

      it 'does not commit changes' do
        expect(Ralph::GitManager).not_to receive(:commit_changes)
        described_class.implement(story, 1, all_requirements)
      end

      it 'logs failed iteration' do
        expect(Ralph::ProgressLogger).to receive(:log_iteration).with(1, story, false)
        described_class.implement(story, 1, all_requirements)
      end
    end

    context 'when implementation fails' do
      it 'returns false for nil response' do
        allow(Ralph::ErrorHandler).to receive(:capture_command_output).and_return(nil)

        result = described_class.implement(story, 1, all_requirements)
        expect(result).to be false
      end

      it 'returns false when COMPLETED not in response' do
        allow(Ralph::ErrorHandler).to receive(:capture_command_output)
          .and_return('Some other response')

        result = described_class.implement(story, 1, all_requirements)
        expect(result).to be false
      end

      it 'logs failed iteration' do
        allow(Ralph::ErrorHandler).to receive(:capture_command_output)
          .and_return('Failed response')

        expect(Ralph::ProgressLogger).to receive(:log_iteration).with(1, story, false)
        described_class.implement(story, 1, all_requirements)
      end
    end

    context 'prompt building' do
      it 'includes story details' do
        expect(Ralph::ErrorHandler).to receive(:capture_command_output) do |prompt, _op|
          expect(prompt).to include('Test Story')
          expect(prompt).to include('Implement test feature')
          expect(prompt).to include('Works correctly')
          'COMPLETED: done'
        end

        described_class.implement(story, 1, all_requirements)
      end

      it 'includes progress context' do
        completed_story = story.dup
        completed_story['passes'] = true
        requirements = {
          'stories' => [completed_story, { 'passes' => false }]
        }

        expect(Ralph::ErrorHandler).to receive(:capture_command_output) do |prompt, _op|
          expect(prompt).to include('1/2 stories done')
          'COMPLETED: done'
        end

        described_class.implement(story, 2, requirements)
      end

      it 'includes AGENTS.md context when present' do
        File.write('AGENTS.md', 'Previous patterns here')

        expect(Ralph::ErrorHandler).to receive(:capture_command_output) do |prompt, _op|
          expect(prompt).to include('Previous patterns here')
          'COMPLETED: done'
        end

        described_class.implement(story, 1, all_requirements)
      end
    end
  end
end
