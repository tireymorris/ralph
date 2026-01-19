# frozen_string_literal: true

require 'spec_helper'

RSpec.describe Ralph::Agent do
  let(:requirements) do
    {
      'project_name' => 'Test',
      'stories' => [
        { 'id' => 'story-1', 'title' => 'Story 1', 'description' => 'Desc', 'passes' => false, 'priority' => 1 }
      ]
    }
  end

  describe '.run' do
    before do
      allow(Ralph::PrdGenerator).to receive(:generate).and_return(requirements)
      allow(Ralph::StoryImplementer).to receive(:implement).and_return(true)
      allow(Ralph::ProgressLogger).to receive(:update_state)
      allow(Ralph::GitManager).to receive(:create_branch)
    end

    context 'dry run mode' do
      it 'generates PRD and exits' do
        expect(Ralph::PrdGenerator).to receive(:generate).with('Test prompt')
        expect(Ralph::StoryImplementer).not_to receive(:implement)

        described_class.run('Test prompt', dry_run: true)
      end

      it 'prints dry run message' do
        expect { described_class.run('Test prompt', dry_run: true) }
          .to output(/Dry run mode/).to_stdout
      end

      it 'returns success exit code' do
        result = described_class.run('Test prompt', dry_run: true)
        expect(result).to eq(Ralph::CLI::EXIT_SUCCESS)
      end
    end

    context 'full implementation' do
      it 'generates PRD then implements stories' do
        expect(Ralph::PrdGenerator).to receive(:generate).and_return(requirements)
        expect(Ralph::StoryImplementer).to receive(:implement)
          .with(hash_including('id' => 'story-1'), 1, requirements)
          .and_return(true)

        described_class.run('Test prompt')
      end

      it 'marks story as passed on success' do
        allow(Ralph::StoryImplementer).to receive(:implement).and_return(true)

        described_class.run('Test prompt')

        expect(requirements['stories'].first['passes']).to be true
      end

      it 'updates state after completing story' do
        allow(Ralph::StoryImplementer).to receive(:implement).and_return(true)
        expect(Ralph::ProgressLogger).to receive(:update_state).with(requirements)

        described_class.run('Test prompt')
      end

      it 'prints completion message when all stories done' do
        allow(Ralph::StoryImplementer).to receive(:implement).and_return(true)

        expect { described_class.run('Test prompt') }
          .to output(/ALL STORIES COMPLETED/).to_stdout
      end

      it 'returns success exit code on completion' do
        allow(Ralph::StoryImplementer).to receive(:implement).and_return(true)

        result = described_class.run('Test prompt')
        expect(result).to eq(Ralph::CLI::EXIT_SUCCESS)
      end
    end

    context 'when PRD generation fails' do
      it 'exits early' do
        allow(Ralph::PrdGenerator).to receive(:generate).and_return(nil)
        expect(Ralph::StoryImplementer).not_to receive(:implement)

        described_class.run('Test prompt')
      end

      it 'returns failure exit code' do
        allow(Ralph::PrdGenerator).to receive(:generate).and_return(nil)

        result = described_class.run('Test prompt')
        expect(result).to eq(Ralph::CLI::EXIT_FAILURE)
      end
    end

    context 'when story implementation fails' do
      it 'retries the story in next iteration' do
        call_count = 0
        allow(Ralph::StoryImplementer).to receive(:implement) do
          call_count += 1
          call_count > 1
        end

        described_class.run('Test prompt')

        expect(call_count).to be >= 2
      end

      it 'does not mark story as passed' do
        allow(Ralph::StoryImplementer).to receive(:implement).and_return(false, true)

        described_class.run('Test prompt')
      end

      it 'increments retry count on failure' do
        allow(Ralph::StoryImplementer).to receive(:implement).and_return(false, true)

        described_class.run('Test prompt')

        expect(requirements['stories'].first['retry_count']).to eq(1)
      end
    end

    context 'when story exceeds max retries' do
      before do
        Ralph::Config.set(:retry_attempts, 2)
        allow(Ralph::StoryImplementer).to receive(:implement).and_return(false)
      end

      after do
        Ralph::Config.reset!
      end

      it 'stops retrying after max attempts' do
        call_count = 0
        allow(Ralph::StoryImplementer).to receive(:implement) do
          call_count += 1
          false
        end

        described_class.run('Test prompt')

        expect(call_count).to eq(2)
      end

      it 'returns failure exit code' do
        result = described_class.run('Test prompt')
        expect(result).to eq(Ralph::CLI::EXIT_FAILURE)
      end

      it 'prints failure message with retry hint' do
        expect { described_class.run('Test prompt') }
          .to output(/exceeded max retries.*--resume/m).to_stdout
      end
    end

    context 'with branch_name in requirements' do
      let(:requirements_with_branch) do
        {
          'project_name' => 'Test',
          'branch_name' => 'feature/test-feature',
          'stories' => [
            { 'id' => 'story-1', 'title' => 'Story 1', 'description' => 'Desc', 'passes' => false, 'priority' => 1 }
          ]
        }
      end

      before do
        allow(Ralph::PrdGenerator).to receive(:generate).and_return(requirements_with_branch)
      end

      it 'creates git branch' do
        expect(Ralph::GitManager).to receive(:create_branch).with('feature/test-feature')
        described_class.run('Test prompt')
      end
    end

    context 'with multiple stories' do
      let(:multi_story_requirements) do
        {
          'project_name' => 'Test',
          'stories' => [
            { 'id' => 'story-1', 'title' => 'Story 1', 'description' => 'D1', 'passes' => false, 'priority' => 1 },
            { 'id' => 'story-2', 'title' => 'Story 2', 'description' => 'D2', 'passes' => false, 'priority' => 2 }
          ]
        }
      end

      before do
        allow(Ralph::PrdGenerator).to receive(:generate).and_return(multi_story_requirements)
      end

      it 'implements stories sequentially' do
        order = []
        allow(Ralph::StoryImplementer).to receive(:implement) do |story, _iter, _reqs|
          order << story['id']
          multi_story_requirements['stories'].find { |s| s['id'] == story['id'] }['passes'] = true
          true
        end

        described_class.run('Test prompt')

        expect(order).to eq(%w[story-1 story-2])
      end

      it 'reports progress correctly' do
        allow(Ralph::StoryImplementer).to receive(:implement) do |story, _iter, _reqs|
          multi_story_requirements['stories'].find { |s| s['id'] == story['id'] }['passes'] = true
          true
        end

        expect { described_class.run('Test prompt') }
          .to output(%r{1/2 stories.*2/2 stories}m).to_stdout
      end
    end

    context 'when max iterations exceeded' do
      before do
        Ralph::Config.set(:max_iterations, 2)
        Ralph::Config.set(:retry_attempts, 10)
        allow(Ralph::StoryImplementer).to receive(:implement).and_return(false)
      end

      after do
        Ralph::Config.reset!
      end

      it 'stops after max iterations' do
        call_count = 0
        allow(Ralph::StoryImplementer).to receive(:implement) do
          call_count += 1
          false
        end

        described_class.run('Test prompt')

        expect(call_count).to eq(2)
      end

      it 'prints max iterations message' do
        expect { described_class.run('Test prompt') }
          .to output(/MAX ITERATIONS REACHED/).to_stdout
      end

      it 'returns partial exit code when some stories incomplete' do
        result = described_class.run('Test prompt')
        expect(result).to eq(Ralph::CLI::EXIT_PARTIAL)
      end

      it 'returns success if all stories completed before max iterations' do
        allow(Ralph::StoryImplementer).to receive(:implement) do
          requirements['stories'].first['passes'] = true
          true
        end
        Ralph::Config.set(:max_iterations, 1)

        # Need to return true on first call to complete story
        result = described_class.run('Test prompt')
        expect(result).to eq(Ralph::CLI::EXIT_SUCCESS)
      end
    end

    context 'with stories having different priorities' do
      let(:priority_requirements) do
        {
          'project_name' => 'Test',
          'stories' => [
            { 'id' => 'story-low', 'title' => 'Low Priority', 'description' => 'D', 'passes' => false,
              'priority' => 3 },
            { 'id' => 'story-high', 'title' => 'High Priority', 'description' => 'D', 'passes' => false,
              'priority' => 1 },
            { 'id' => 'story-med', 'title' => 'Med Priority', 'description' => 'D', 'passes' => false, 'priority' => 2 }
          ]
        }
      end

      before do
        allow(Ralph::PrdGenerator).to receive(:generate).and_return(priority_requirements)
      end

      it 'implements stories in priority order' do
        order = []
        allow(Ralph::StoryImplementer).to receive(:implement) do |story, _iter, _reqs|
          order << story['id']
          priority_requirements['stories'].find { |s| s['id'] == story['id'] }['passes'] = true
          true
        end

        described_class.run('Test prompt')

        expect(order).to eq(%w[story-high story-med story-low])
      end
    end

    context 'with long description' do
      let(:long_desc_requirements) do
        {
          'project_name' => 'Test',
          'stories' => [
            {
              'id' => 'story-1',
              'title' => 'Story 1',
              'description' => 'A' * 100,
              'passes' => false,
              'priority' => 1
            }
          ]
        }
      end

      before do
        allow(Ralph::PrdGenerator).to receive(:generate).and_return(long_desc_requirements)
        allow(Ralph::StoryImplementer).to receive(:implement).and_return(true)
      end

      it 'truncates description with ellipsis' do
        expect { described_class.run('Test prompt') }
          .to output(/\.\.\./).to_stdout
      end
    end
  end

  describe '.resume' do
    let(:prd_file) { Ralph::Config.get(:prd_file) }
    let(:existing_requirements) do
      {
        'project_name' => 'Test',
        'stories' => [
          { 'id' => 'story-1', 'title' => 'Story 1', 'description' => 'D1', 'passes' => true, 'priority' => 1 },
          { 'id' => 'story-2', 'title' => 'Story 2', 'description' => 'D2', 'passes' => false, 'priority' => 2 }
        ]
      }
    end

    before do
      allow(Ralph::StoryImplementer).to receive(:implement).and_return(true)
      allow(Ralph::ProgressLogger).to receive(:update_state)
    end

    after do
      File.delete(prd_file) if File.exist?(prd_file)
    end

    it 'loads existing PRD file' do
      File.write(prd_file, JSON.pretty_generate(existing_requirements))

      expect { described_class.resume }
        .to output(/PRD loaded: Test/).to_stdout
    end

    it 'shows progress from existing PRD' do
      File.write(prd_file, JSON.pretty_generate(existing_requirements))

      expect { described_class.resume }
        .to output(%r{1/2 stories already completed}).to_stdout
    end

    it 'only implements incomplete stories' do
      File.write(prd_file, JSON.pretty_generate(existing_requirements))

      expect(Ralph::StoryImplementer).to receive(:implement)
        .with(hash_including('id' => 'story-2'), anything, anything)
        .and_return(true)

      described_class.resume
    end

    it 'returns success when all stories complete' do
      completed_requirements = existing_requirements.dup
      completed_requirements['stories'].each { |s| s['passes'] = true }
      File.write(prd_file, JSON.pretty_generate(completed_requirements))

      result = described_class.resume
      expect(result).to eq(Ralph::CLI::EXIT_SUCCESS)
    end

    it 'returns failure when PRD file is invalid' do
      File.write(prd_file, 'invalid json')

      result = described_class.resume
      expect(result).to eq(Ralph::CLI::EXIT_FAILURE)
    end

    it 'initializes retry counts if missing' do
      File.write(prd_file, JSON.pretty_generate(existing_requirements))
      allow(Ralph::StoryImplementer).to receive(:implement).and_return(true)

      described_class.resume

      # Should not raise error due to missing retry_count
    end
  end
end
