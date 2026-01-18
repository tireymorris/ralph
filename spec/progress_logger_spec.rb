# frozen_string_literal: true

require 'spec_helper'

RSpec.describe Ralph::ProgressLogger do
  let(:prd_file) { 'prd.json' }

  after(:each) do
    File.delete(prd_file) if File.exist?(prd_file)
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
