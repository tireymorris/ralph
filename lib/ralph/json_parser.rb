# frozen_string_literal: true

require 'json'

module Ralph
  class JsonParser
    class << self
      def parse_safely(json_string, context = 'JSON parsing')
        return nil if json_string.nil?

        safe_string = encode_utf8(json_string)
        return nil if safe_string.strip.empty?

        ErrorHandler.with_error_handling(context) do
          cleaned = extract_json_object(safe_string)
          parsed = JSON.parse(cleaned)
          validate_hash!(parsed)
          parsed
        end
      end

      private

      def encode_utf8(string)
        string.encode('UTF-8', 'binary', invalid: :replace, undef: :replace, replace: '')
      end

      def extract_json_object(string)
        cleaned = string.strip
        json_match = cleaned.match(/\{[\s\S]*\}/)
        json_match ? json_match[0] : cleaned
      end

      def validate_hash!(parsed)
        raise ArgumentError, 'Invalid JSON structure: expected Hash' unless parsed.is_a?(Hash)
      end
    end
  end
end
