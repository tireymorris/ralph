#!/usr/bin/env ruby
# frozen_string_literal: true

require_relative 'cli/registry'
require_relative 'cli/help'
require_relative 'cli/interactive'

Dir.glob(File.join(__dir__, 'commands', '*.rb')).each do |file|
  require file
end

if __FILE__ == $PROGRAM_NAME || ARGV[0]&.start_with?('--')
  command_name = ARGV.shift

  case command_name
  when '-i', '--interactive'
    CLI::Interactive.run(CLI::Registry.commands)
  when nil, 'help', '--help', '-h'
    CLI::Help.show(CLI::Registry.commands)
  else
    CLI::Registry.run(command_name, *ARGV)
  end
end
