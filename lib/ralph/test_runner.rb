# frozen_string_literal: true

module Ralph
  # Test runner for various project types
  class TestRunner
    TEST_COMMANDS = [
      'npm test',
      'yarn test',
      'pytest',
      'python -m pytest',
      'cargo test',
      'go test'
    ].freeze

    class << self
      def run
        TEST_COMMANDS.each do |cmd|
          next unless system("which #{cmd.split.first} > /dev/null 2>&1")

          print '🧪 Running tests... '
          result = system(cmd)
          puts result ? '✅' : '❌'
          Logger.info('Tests executed', { command: cmd, result: result })
          return result
        end

        puts '⚠️ No test framework detected'
        Logger.warn('No test framework detected')
        true # Continue without tests
      end
    end
  end
end
