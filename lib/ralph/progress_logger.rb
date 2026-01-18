# frozen_string_literal: true

module Ralph
  class ProgressLogger
    class << self
      def update_state(requirements)
        StateManager.write_state(requirements)
      end
    end
  end
end
