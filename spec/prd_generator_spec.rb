# frozen_string_literal: true

require 'spec_helper'

RSpec.describe Ralph::PrdGenerator do
  let(:valid_prd_response) do
    {
      'project_name' => 'Test Project',
      'branch_name' => 'feature/test',
      'stories' => [
        {
          'id' => 'story-1',
          'title' => 'First Story',
          'description' => 'Do something',
          'acceptance_criteria' => ['Criterion 1'],
          'priority' => 1,
          'passes' => false
        }
      ]
    }.to_json
  end

  describe '.generate' do
    context 'with successful response' do
      before do
        allow(Ralph::ErrorHandler).to receive(:capture_command_output)
          .and_return(valid_prd_response)
      end

      it 'returns parsed requirements' do
        result = described_class.generate('Add feature')

        expect(result['project_name']).to eq('Test Project')
        expect(result['stories'].length).to eq(1)
      end

      it 'creates prd.json file' do
        described_class.generate('Add feature')

        expect(File.exist?('prd.json')).to be true
        content = JSON.parse(File.read('prd.json'))
        expect(content['project_name']).to eq('Test Project')
      end

      it 'logs info on completion' do
        expect(Ralph::Logger).to receive(:info).with('Generating PRD for prompt', anything)
        expect(Ralph::Logger).to receive(:info).with('PRD analysis complete', anything)

        described_class.generate('Add feature')
      end
    end

    context 'with invalid response' do
      it 'returns nil for nil response' do
        allow(Ralph::ErrorHandler).to receive(:capture_command_output).and_return(nil)

        expect(described_class.generate('prompt')).to be_nil
      end

      it 'returns nil for invalid JSON' do
        allow(Ralph::ErrorHandler).to receive(:capture_command_output)
          .and_return('not json')

        expect(described_class.generate('prompt')).to be_nil
      end

      it 'returns nil for missing required fields' do
        incomplete = { 'project_name' => 'Test' }.to_json
        allow(Ralph::ErrorHandler).to receive(:capture_command_output)
          .and_return(incomplete)

        expect(described_class.generate('prompt')).to be_nil
      end
    end

    context 'validation' do
      it 'rejects empty stories array' do
        response = {
          'project_name' => 'Test',
          'branch_name' => 'test',
          'stories' => []
        }.to_json

        allow(Ralph::ErrorHandler).to receive(:capture_command_output)
          .and_return(response)

        expect(described_class.generate('prompt')).to be_nil
      end

      it 'rejects stories missing required fields' do
        response = {
          'project_name' => 'Test',
          'branch_name' => 'test',
          'stories' => [{ 'id' => 'story-1' }]
        }.to_json

        allow(Ralph::ErrorHandler).to receive(:capture_command_output)
          .and_return(response)

        expect(described_class.generate('prompt')).to be_nil
      end

      it 'rejects stories with empty acceptance criteria' do
        response = {
          'project_name' => 'Test',
          'branch_name' => 'test',
          'stories' => [{
            'id' => 'story-1',
            'title' => 'Title',
            'description' => 'Desc',
            'acceptance_criteria' => [],
            'priority' => 1
          }]
        }.to_json

        allow(Ralph::ErrorHandler).to receive(:capture_command_output)
          .and_return(response)

        expect(described_class.generate('prompt')).to be_nil
      end

      it 'rejects non-array acceptance criteria' do
        response = {
          'project_name' => 'Test',
          'branch_name' => 'test',
          'stories' => [{
            'id' => 'story-1',
            'title' => 'Title',
            'description' => 'Desc',
            'acceptance_criteria' => 'not an array',
            'priority' => 1
          }]
        }.to_json

        allow(Ralph::ErrorHandler).to receive(:capture_command_output)
          .and_return(response)

        expect(described_class.generate('prompt')).to be_nil
      end
    end
  end
end
