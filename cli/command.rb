# frozen_string_literal: true

module CLI
  # Represents a registered CLI command with name, description, and handler
  class Command
    attr_reader :name, :description, :handler

    def initialize(name, description, &handler)
      @name = name
      @description = description
      @handler = handler
    end

    def call(*args)
      handler.call(*args)
    end
  end
end

