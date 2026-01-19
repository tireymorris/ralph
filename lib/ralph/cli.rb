# frozen_string_literal: true

module Ralph
  class CLI
    EXIT_SUCCESS = 0
    EXIT_FAILURE = 1
    EXIT_PARTIAL = 2

    class << self
      def run(args)
        if args.empty? || args.include?('--help') || args.include?('-h')
          show_help
          return EXIT_SUCCESS
        end

        options = parse_options(args)

        if options[:resume]
          return run_resume
        end

        if options[:prompt].empty?
          puts '❌ Error: Please provide a prompt'
          show_help
          return EXIT_FAILURE
        end

        puts "📝 Request: #{options[:prompt]}"
        puts "📁 Working in: #{Dir.pwd}"
        puts "🎯 Mode: #{options[:dry_run] ? 'Dry run (PRD only)' : 'Full implementation'}"

        Agent.run(options[:prompt], dry_run: options[:dry_run])
      end

      def show_help
        puts <<~HELP
          Ralph - Autonomous Software Development Agent

          Usage:
            ./bin/ralph "your feature description"           # Full implementation
            ./bin/ralph "your feature description" --dry-run # Generate PRD only
            ./bin/ralph --resume                             # Resume from existing prd.json

          Options:
            --dry-run    Generate PRD only, don't implement
            --resume     Resume implementation from existing prd.json
            --help, -h   Show this help message
        HELP
      end

      private

      def parse_options(args)
        {
          dry_run: args.include?('--dry-run'),
          resume: args.include?('--resume'),
          prompt: args.reject { |arg| arg.start_with?('--') || arg == '-h' }.join(' ')
        }
      end

      def run_resume
        prd_file = Config.get(:prd_file)

        unless File.exist?(prd_file)
          puts "❌ Error: No #{prd_file} found to resume from"
          puts '💡 Run ralph with a prompt first to generate a PRD'
          return EXIT_FAILURE
        end

        puts "📂 Resuming from: #{prd_file}"
        puts "📁 Working in: #{Dir.pwd}"
        puts '🎯 Mode: Resume implementation'

        Agent.resume
      end
    end
  end
end
