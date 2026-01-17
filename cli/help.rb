# frozen_string_literal: true

module Ralph
  module Help
    def self.show(commands)
      puts '╔═══════════════════════════════════════════════════════════════════════════════╗'
      puts '║                                 🚀 Ralph                                      ║'
      puts '╚═══════════════════════════════════════════════════════════════════════════════╝'
      puts
      puts 'Usage:'
      puts '  ./ralph <command> [args...]'
      puts '  ./ralph -i                    # Interactive mode'
      puts
      puts 'Available Commands:'
      puts

      commands.keys.sort.group_by { |k| k.split(':').first }.each do |namespace, cmds|
        puts "  #{namespace.upcase}:"
        cmds.each do |cmd_name|
          command = commands[cmd_name]
          puts "    #{cmd_name.ljust(40)} #{command.description}"
        end
        puts
      end

      puts 'Examples:'
      puts '  ./ralph -i'
      puts
    end
  end
end
