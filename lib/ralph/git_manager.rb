# frozen_string_literal: true

module Ralph
  # Git operations manager
  class GitManager
    class << self
      def validate_repository
        ErrorHandler.with_error_handling('Git repository validation') do
          raise StandardError, 'Not in a git repository' unless system('git rev-parse --git-dir > /dev/null 2>&1')
        end
      end

      def commit_changes(story)
        puts '💾 Committing changes...'

        ErrorHandler.with_error_handling('Git commit', { story: story['id'] }) do
          if system('git diff --quiet --exit-code 2>/dev/null') && system('git diff --staged --quiet --exit-code 2>/dev/null')
            Logger.info('No changes to commit', { story: story['id'] })
            return true
          end

          ErrorHandler.safe_system_command('git add .', 'Stage changes')

          commit_title = story['title']&.to_s&.gsub("'", "''") || 'Story implementation'
          commit_desc = story['description']&.to_s&.gsub("'", "''") || ''
          story_id = story['id']&.to_s || 'unknown'

          commit_message = "feat: #{commit_title}

#{commit_desc}

Story: #{story_id}"

          ErrorHandler.safe_system_command("git commit -m '#{commit_message}'", 'Commit changes')
        end
      end
    end
  end
end
