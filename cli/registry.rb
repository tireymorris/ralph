# frozen_string_literal: true

require_relative 'command'

module CLI
  # Stores and executes registered commands
  class Registry
    class << self
      def commands
        @commands ||= {}
      end

      def register(name, description, &block)
        commands[name] = Command.new(name, description, &block)
      end

      def run(command_name, *args)
        command = commands[command_name]

        if command.nil?
          puts "❌ Unknown command: #{command_name}"
          puts
          Help.show(commands)
          exit 1
        end

        command.call(*args)
      end
    end
  end
end

