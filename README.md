# Ralph CLI Builder

A minimal Ruby CLI framework with command registration, interactive mode, and autocomplete.

## Usage

```bash
./ralph --help              # show commands
./ralph -i                  # interactive mode
./ralph namespace:command   # run command
```

## Adding Commands

Create files in `commands/` that register commands:

```ruby
Ralph::Registry.register('myapp:greet', 'Say hello') do |name = 'World'|
  puts "Hello, #{name}!"
end
```

Commands are auto-loaded from `commands/*.rb`.

