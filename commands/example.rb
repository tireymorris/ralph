# frozen_string_literal: true

Ralph::Registry.register('example:hello', 'Print a greeting') do |name = 'World'|
  puts "Hello, #{name}!"
end

Ralph::Registry.register('example:echo', 'Echo back arguments') do |*args|
  puts args.join(' ')
end
Ralph::Registry.register('init', 'Initialize Ralph project structure') do
  puts 'Initializing Ralph project...'
  Ralph::Project.init
end

Ralph::Registry.register('prd:create', 'Create a new PRD from feature description') do |description = nil|
  if description.nil?
    puts 'Error: Please provide a feature description'
    puts 'Usage: ralph prd:create "Build a user authentication system"'
    return
  end

  puts "Creating PRD for: #{description}"
  Ralph::PRD.create(description)
end

Ralph::Registry.register('prd:convert', 'Convert PRD markdown to JSON format') do |file_path = nil|
  if file_path.nil?
    puts 'Error: Please provide PRD file path'
    puts 'Usage: ralph prd:convert tasks/prd-feature.md'
    return
  end

  puts "Converting PRD: #{file_path}"
  Ralph::PRD.convert(file_path)
end

Ralph::Registry.register('run', 'Start Ralph autonomous agent loop') do |max_iterations = nil|
  if max_iterations
    puts "Starting Ralph autonomous agent (max #{max_iterations} iterations)..."
    Ralph::Agent.run(max_iterations.to_i)
  else
    puts 'Starting Ralph autonomous agent (run until completion)...'
    Ralph::Agent.run
  end
end

Ralph::Registry.register('status', 'Show current Ralph progress and status') do
  Ralph::Status.show
end

Ralph::Registry.register('debug', 'Show debug information and state') do
  Ralph::Debug.show
end
