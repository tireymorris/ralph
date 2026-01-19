# frozen_string_literal: true

module Ralph
  class ProgressLogger
    class << self
      def update_state(requirements)
        ErrorHandler.with_error_handling('State update', { operation: 'write_state' }) do
          StateManager.write_state(requirements)
        end
      end
    end
  end
end
