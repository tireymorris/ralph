# frozen_string_literal: true

require 'spec_helper'

RSpec.describe Ralph::JsonParser do
  describe '.parse_safely' do
    it 'returns nil for nil input' do
      expect(described_class.parse_safely(nil)).to be_nil
    end

    it 'returns nil for empty string' do
      expect(described_class.parse_safely('')).to be_nil
    end

    it 'parses valid JSON' do
      json = '{"key": "value"}'
      result = described_class.parse_safely(json)
      expect(result).to eq({ 'key' => 'value' })
    end

    it 'extracts JSON from surrounding text' do
      json = 'Some text {"key": "value"} more text'
      result = described_class.parse_safely(json)
      expect(result).to eq({ 'key' => 'value' })
    end

    it 'returns nil for invalid JSON' do
      expect(described_class.parse_safely('not json')).to be_nil
    end

    it 'returns nil for non-Hash JSON' do
      expect(described_class.parse_safely('[1, 2, 3]')).to be_nil
    end

    it 'handles nested JSON' do
      json = '{"outer": {"inner": "value"}}'
      result = described_class.parse_safely(json)
      expect(result).to eq({ 'outer' => { 'inner' => 'value' } })
    end

    it 'handles unicode characters' do
      json = '{"temp": "72F"}'
      result = described_class.parse_safely(json)
      expect(result).to eq({ 'temp' => '72F' })
    end

    it 'handles binary data gracefully' do
      json = "{\"key\": \"value\xC0\xC1\"}"
      result = described_class.parse_safely(json)
      expect(result).to be_a(Hash)
    end
  end
end
