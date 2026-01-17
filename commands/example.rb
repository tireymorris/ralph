CLI::Registry.register('example:hello', 'Print a greeting') do |name = 'World'|
  puts "Hello, #{name}!"
end

CLI::Registry.register('example:echo', 'Echo back arguments') do |*args|
  puts args.join(' ')
end
