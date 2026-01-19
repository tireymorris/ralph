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
      allow(Ralph::GitManager).to receive(:commit_changes)
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

      it 'commits changes' do
        expect(Ralph::GitManager).to receive(:commit_changes).with(story)
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

      it 'does not commit changes' do
        allow(Ralph::ErrorHandler).to receive(:capture_command_output)
          .and_return('Failed response')

        expect(Ralph::GitManager).not_to receive(:commit_changes)
        described_class.implement(story, 1, all_requirements)
      end

      it 'logs error when response is nil' do
        allow(Ralph::ErrorHandler).to receive(:capture_command_output).and_return(nil)
        expect(Ralph::Logger).to receive(:error).with('Story implementation failed', { story: 'story-1' })

        described_class.implement(story, 1, all_requirements)
      end

      it 'prints failure message when COMPLETED not in response' do
        allow(Ralph::ErrorHandler).to receive(:capture_command_output)
          .and_return('Some other response')

        expect { described_class.implement(story, 1, all_requirements) }
          .to output(/Implementation did not complete/).to_stdout
      end
    end

    context 'process_implementation_response with nil' do
      it 'handles nil response safely' do
        # Directly test process_implementation_response with nil
        result = described_class.send(:process_implementation_response, story, 1, nil)
        expect(result).to be false
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

      it 'instructs OpenCode to run tests' do
        expect(Ralph::ErrorHandler).to receive(:capture_command_output) do |prompt, _op|
          expect(prompt).to include('Run tests')
          expect(prompt).to include('responsible for running tests')
          'COMPLETED: done'
        end

        described_class.implement(story, 1, all_requirements)
      end
    end
  end
end
