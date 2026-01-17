# frozen_string_literal: true

require 'readline'
require 'io/console'

module Ralph
  # Interactive REPL mode with autocomplete
  module Interactive
    def self.run(commands)
      cmd_list = commands.keys.sort
      setup_autocomplete(cmd_list)

      loop do
        clear_screen
        print_header
        print_commands(commands)
        print_tips

        input = Readline.readline("\e[1;32m❯\e[0m ", true)
        break if quit?(input)
        next if input.strip.empty?

        command_name, *args = input.strip.split(/\s+/)

        unless commands.key?(command_name)
          handle_unknown_command(command_name, cmd_list)
          next
        end

        execute_command(command_name, args, commands)
      end

      puts
      puts "\e[1;32m👋 Goodbye!\e[0m"
      puts
    end

    class << self
      private

      def setup_autocomplete(cmd_list)
        Readline.completion_append_character = nil
        Readline.completion_proc = proc do |prefix|
          cmd_list.grep(/^#{Regexp.escape(prefix)}/i)
        end
      end

      def clear_screen
        system('clear') || system('cls')
      end

      def print_header
        puts
        puts "\e[1;36m╔═══════════════════════════════════════════════════════════════════════════════╗\e[0m"
        puts "\e[1;36m║\e[0m                              \e[1;33m🚀 Ralph\e[0m                                        \e[1;36m║\e[0m"
        puts "\e[1;36m╚═══════════════════════════════════════════════════════════════════════════════╝\e[0m"
        puts
        puts "\e[1;32mAvailable Commands:\e[0m"
        puts
      end

      def print_commands(commands)
        commands.keys.sort.group_by { |k| k.split(':').first }.each do |namespace, cmds|
          puts "  \e[1;35m#{namespace.upcase}:\e[0m"
          cmds.each do |cmd_name|
            command = commands[cmd_name]
            puts "    \e[36m#{cmd_name.ljust(40)}\e[0m \e[90m│\e[0m #{command.description}"
          end
          puts
        end
      end

      def print_tips
        puts "\e[90m─\e[0m" * 80
        puts
        puts "\e[1;33m💡 Tips:\e[0m"
        puts '  • Type command name and press Tab for autocomplete'
        puts '  • Commands can accept arguments'
        puts "  • Type 'exit' or 'quit' to leave"
        puts '  • Press Ctrl+C to cancel'
        puts
        puts "\e[90m─\e[0m" * 80
        puts
      end

      def quit?(input)
        input.nil? || %w[exit quit].include?(input.strip.downcase)
      end

      def handle_unknown_command(command_name, cmd_list)
        fuzzy_matches = cmd_list.select { |cmd| cmd.include?(command_name) }

        puts
        if fuzzy_matches.any?
          puts "\e[1;33m❓ Did you mean one of these?\e[0m"
          fuzzy_matches.each { |match| puts "  • #{match}" }
        else
          puts "\e[1;31m❌ Unknown command: #{command_name}\e[0m"
        end
        puts
        puts 'Press any key to continue...'
        $stdin.getch
      end

      def execute_command(command_name, args, commands)
        puts
        puts "\e[90m─\e[0m" * 80
        puts "\e[1;36mRunning:\e[0m #{command_name} #{args.join(' ')}"
        puts "\e[90m─\e[0m" * 80
        puts

        begin
          commands[command_name].call(*args)
        rescue StandardError => e
          puts
          puts "\e[1;31m❌ Error: #{e.message}\e[0m"
          puts e.backtrace.first(5).map { |line| "   \e[90m#{line}\e[0m" }.join("\n")
        end

        puts
        puts "\e[90m─\e[0m" * 80
        puts
        puts 'Press any key to continue...'
        $stdin.getch
      end
    end
  end
end
