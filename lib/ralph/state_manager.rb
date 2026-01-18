# frozen_string_literal: true

module Ralph
  class StateManager
    class << self
      def write_state(requirements)
        ErrorHandler.with_error_handling('State file update') do
          prd_file = Ralph::Config.get(:prd_file)
          File.write(prd_file, JSON.pretty_generate(requirements))
        end
      end
    end
  end
end
