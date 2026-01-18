# frozen_string_literal: true

require 'spec_helper'

RSpec.describe Ralph::ProgressLogger do
  let(:progress_file) { 'progress.txt' }
  let(:prd_file) { 'prd.json' }

  let(:story) do
    {
      'id' => 'story-1',
      'title' => 'Test Story',
      'priority' => 1
    }
  end

  after(:each) do
    File.delete(progress_file) if File.exist?(progress_file)
    File.delete(prd_file) if File.exist?(prd_file)
  end

  describe '.log_iteration' do
    context 'on success' do
      it 'logs info level message' do
        allow(Ralph::Logger).to receive(:log)
        allow(Ralph::Logger).to receive(:debug)

        expect(Ralph::Logger).to receive(:log).with(
          :info,
          'Iteration 1 completed',
          hash_including(story: 'Test Story', success: true)
        )

        described_class.log_iteration(1, story, true)
      end

      it 'writes to progress file' do
        described_class.log_iteration(1, story, true)

        content = File.read(progress_file)
        expect(content).to include('Iteration 1')
        expect(content).to include('Test Story')
        expect(content).to include('Success')
      end
    end

    context 'on failure' do
      it 'logs error level message' do
        allow(Ralph::Logger).to receive(:log)
        allow(Ralph::Logger).to receive(:debug)

        expect(Ralph::Logger).to receive(:log).with(
          :error,
          'Iteration 1 completed',
          hash_including(success: false)
        )

        described_class.log_iteration(1, story, false)
      end

      it 'writes failure status to progress file' do
        described_class.log_iteration(1, story, false)

        content = File.read(progress_file)
        expect(content).to include('Failed')
      end
    end

    it 'appends to existing progress file' do
      File.write(progress_file, "Previous content\n")

      described_class.log_iteration(1, story, true)

      content = File.read(progress_file)
      expect(content).to include('Previous content')
      expect(content).to include('Iteration 1')
    end

    it 'includes timestamp' do
      described_class.log_iteration(1, story, true)

      content = File.read(progress_file)
      expect(content).to match(/\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}/)
    end
  end

  describe '.update_state' do
    let(:requirements) do
      {
        'project_name' => 'Test Project',
        'stories' => [
          { 'id' => 'story-1', 'passes' => true }
        ]
      }
    end

    it 'writes requirements to PRD file' do
      described_class.update_state(requirements)

      content = File.read(prd_file)
      parsed = JSON.parse(content)

      expect(parsed['project_name']).to eq('Test Project')
      expect(parsed['stories'].first['passes']).to be true
    end

    it 'formats JSON with pretty print' do
      described_class.update_state(requirements)

      content = File.read(prd_file)
      expect(content).to include("\n")
      expect(content.lines.length).to be > 1
    end

    it 'handles write errors gracefully' do
      allow(File).to receive(:write).and_raise(Errno::EACCES)

      expect { described_class.update_state(requirements) }.not_to raise_error
    end
  end
end
