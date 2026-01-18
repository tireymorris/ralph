# frozen_string_literal: true

require 'simplecov'
SimpleCov.start do
  add_filter '/spec/'
  enable_coverage :branch
  minimum_coverage 90
end

require_relative '../lib/ralph/config'
require_relative '../lib/ralph/logger'
require_relative '../lib/ralph/error_handler'
require_relative '../lib/ralph/git_manager'
require_relative '../lib/ralph/progress_logger'
require_relative '../lib/ralph/prd_generator'
require_relative '../lib/ralph/story_implementer'
require_relative '../lib/ralph/agent'
require_relative '../lib/ralph/cli'

RSpec.configure do |config|
  config.expect_with :rspec do |expectations|
    expectations.include_chain_clauses_in_custom_matcher_descriptions = true
  end

  config.mock_with :rspec do |mocks|
    mocks.verify_partial_doubles = true
  end

  config.shared_context_metadata_behavior = :apply_to_host_groups
  config.filter_run_when_matching :focus
  config.example_status_persistence_file_path = 'spec/examples.txt'
  config.disable_monkey_patching!
  config.warnings = true
  config.order = :random
  Kernel.srand config.seed

  config.before(:each) do
    Ralph::Config.reset!
    Ralph::Logger.configure(:error)
  end

  config.after(:each) do
    %w[ralph.log progress.txt prd.json AGENTS.md].each do |file|
      File.delete(file) if File.exist?(file)
    end
  end
end
