# frozen_string_literal: true

module Ralph
  # Progress logging utilities
  class ProgressLogger
    class << self
      def log_iteration(iteration, story, success)
        Logger.log(success ? :info : :error, "Iteration #{iteration} completed", {
                     story: story['title'],
                     story_id: story['id'],
                     success: success,
                     priority: story['priority']
                   })

        ErrorHandler.with_error_handling('Progress logging') do
          progress_file = Ralph::Config.get(:progress_file)
          log = [
            "## Iteration #{iteration} - #{Time.now.strftime('%Y-%m-%d %H:%M:%S')}",
            "Story: #{story['title']}",
            "Status: #{success ? 'Success' : 'Failed'}",
            '---'
          ].join("\n")

          File.open(progress_file, 'a') { |f| f.puts(log + "\n") }
        end
      end

      def update_state(requirements)
        ErrorHandler.with_error_handling('State update') do
          prd_file = Ralph::Config.get(:prd_file)
          File.write(prd_file, JSON.pretty_generate(requirements))
        end
      end
    end
  end
end
