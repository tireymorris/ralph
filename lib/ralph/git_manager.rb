# frozen_string_literal: true

require 'shellwords'

module Ralph
  class GitManager
    class << self
      def validate_repository
        ErrorHandler.with_error_handling('Git repository validation') do
          raise StandardError, 'Not in a git repository' unless system('git rev-parse --git-dir > /dev/null 2>&1')
        end
      end

      def create_branch(branch_name)
        return unless branch_name

        ErrorHandler.with_error_handling('Git branch creation', { branch: branch_name }) do
          # Check if branch already exists
          if system("git show-ref --verify --quiet refs/heads/#{Shellwords.escape(branch_name)}")
            puts "  📌 Branch '#{branch_name}' already exists, switching to it"
            ErrorHandler.safe_system_command("git checkout #{Shellwords.escape(branch_name)}", 'Switch branch')
          else
            puts "  🌱 Creating new branch '#{branch_name}'"
            ErrorHandler.safe_system_command("git checkout -b #{Shellwords.escape(branch_name)}", 'Create branch')
          end
        end
      end

      def current_branch
        `git rev-parse --abbrev-ref HEAD 2>/dev/null`.strip
      end

      def no_unstaged_changes?
        system('git diff --quiet --exit-code 2>/dev/null')
      end

      def no_staged_changes?
        system('git diff --staged --quiet --exit-code 2>/dev/null')
      end

      def commit_changes(story)
        puts '💾 Committing changes...'

        ErrorHandler.with_error_handling('Git commit', { story: story['id'] }) do
          if no_unstaged_changes? && no_staged_changes?
            Logger.info('No changes to commit', { story: story['id'] })
            return true
          end

          ErrorHandler.safe_system_command('git add .', 'Stage changes')

          commit_title = story['title']&.to_s || 'Story implementation'
          commit_desc = story['description'].to_s
          story_id = story['id']&.to_s || 'unknown'

          commit_message = "feat: #{commit_title}\n\n#{commit_desc}\n\nStory: #{story_id}"

          ErrorHandler.safe_system_command("git commit -m #{Shellwords.escape(commit_message)}", 'Commit changes')
        end
      end
    end
  end
end
