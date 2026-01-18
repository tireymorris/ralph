# frozen_string_literal: true

module Ralph
  class CLI
    class << self
      def run(args)
        if args.empty? || args.include?('--help') || args.include?('-h')
          show_help
          return
        end

        dry_run = args.include?('--dry-run')
        prompt_args = args.reject { |arg| arg == '--dry-run' }

        if prompt_args.empty?
          puts '❌ Error: Please provide a prompt when using --dry-run'
          show_help
          return
        end

        prompt = prompt_args.join(' ')

        puts "📝 Request: #{prompt}"
        puts "📁 Working in: #{Dir.pwd}"
        puts "🎯 Mode: #{dry_run ? 'Dry run (PRD only)' : 'Full implementation'}"

        Agent.run(prompt, dry_run: dry_run)
      end

      def show_help
        puts <<~HELP
          Ralph - Autonomous Software Development Agent

          Usage:
            ./ralph "your feature description"           # Full implementation
            ./ralph "your feature description" --dry-run # Generate PRD only
        HELP
      end
    end
  end
end
