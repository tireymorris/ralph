# frozen_string_literal: true

require_relative 'spec_helper'

RSpec.describe Ralph::StateManager do
  describe '.write_state' do
    let(:requirements) do
      {
        'project' => 'Test Project',
        'stories' => [
          {
            'title' => 'Test Story',
            'acceptance_criteria' => ['criteria 1'],
            'status' => 'pending'
          }
        ]
      }
    end

    before do
      allow(Ralph::Config).to receive(:get).with(:prd_file).and_return('test_prd.json')
    end

    after do
      File.delete('test_prd.json') if File.exist?('test_prd.json')
    end

    it 'writes requirements to PRD file' do
      described_class.write_state(requirements)
      expect(File.exist?('test_prd.json')).to be true
    end

    it 'writes pretty JSON format' do
      described_class.write_state(requirements)
      content = File.read('test_prd.json')
      parsed = JSON.parse(content)
      expect(parsed).to eq(requirements)
    end

    it 'uses configured PRD file path' do
      expect(Ralph::Config).to receive(:get).with(:prd_file).and_return('custom_prd.json')
      described_class.write_state(requirements)
    end
  end
end
