# CLI Builder

A minimal Ruby CLI framework with command registration, interactive mode, and autocomplete.

## Usage

```bash
./cli.rb --help              # show commands
./cli.rb -i                  # interactive mode
./cli.rb namespace:command   # run command
```

## Adding Commands

Create files in `commands/` that register commands:

```ruby
CLI::Registry.register('myapp:greet', 'Say hello') do |name = 'World'|
  puts "Hello, #{name}!"
end
```

Commands are auto-loaded from `commands/*.rb`.

