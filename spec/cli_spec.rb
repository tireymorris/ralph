# frozen_string_literal: true

require 'spec_helper'

RSpec.describe Ralph::CLI do
  describe '.run' do
    before do
      allow(Ralph::Agent).to receive(:run).and_return(described_class::EXIT_SUCCESS)
    end

    context 'with no arguments' do
      it 'shows help and returns success' do
        result = nil
        expect { result = described_class.run([]) }
          .to output(/Usage:/).to_stdout
        expect(result).to eq(described_class::EXIT_SUCCESS)
      end

      it 'does not call Agent.run' do
        expect(Ralph::Agent).not_to receive(:run)
        described_class.run([])
      end
    end

    context 'with --help flag' do
      it 'shows help' do
        expect { described_class.run(['--help']) }
          .to output(/Usage:/).to_stdout
      end

      it 'returns success exit code' do
        result = described_class.run(['--help'])
        expect(result).to eq(described_class::EXIT_SUCCESS)
      end
    end

    context 'with -h flag' do
      it 'shows help' do
        expect { described_class.run(['-h']) }
          .to output(/Usage:/).to_stdout
      end
    end

    context 'with prompt' do
      it 'calls Agent.run with prompt' do
        expect(Ralph::Agent).to receive(:run).with('Add feature', dry_run: false)
        described_class.run(['Add feature'])
      end

      it 'joins multiple words into single prompt' do
        expect(Ralph::Agent).to receive(:run).with('Add user authentication', dry_run: false)
        described_class.run(%w[Add user authentication])
      end

      it 'prints request info' do
        expect { described_class.run(['Test']) }
          .to output(/Request: Test/).to_stdout
      end

      it 'prints working directory' do
        expect { described_class.run(['Test']) }
          .to output(/Working in:/).to_stdout
      end

      it 'prints mode as full implementation' do
        expect { described_class.run(['Test']) }
          .to output(/Mode: Full implementation/).to_stdout
      end
    end

    context 'with --dry-run flag' do
      it 'sets dry_run to true' do
        expect(Ralph::Agent).to receive(:run).with('Test prompt', dry_run: true)
        described_class.run(['Test', 'prompt', '--dry-run'])
      end

      it 'prints mode as dry run' do
        expect { described_class.run(['Test', '--dry-run']) }
          .to output(/Mode: Dry run/).to_stdout
      end

      it 'excludes --dry-run from prompt' do
        expect(Ralph::Agent).to receive(:run).with('Add feature', dry_run: true)
        described_class.run(['Add', 'feature', '--dry-run'])
      end
    end

    context 'with only --dry-run flag' do
      it 'shows error and help' do
        expect { described_class.run(['--dry-run']) }
          .to output(/Error:.*prompt.*Usage:/m).to_stdout
      end

      it 'returns failure exit code' do
        result = described_class.run(['--dry-run'])
        expect(result).to eq(described_class::EXIT_FAILURE)
      end

      it 'does not call Agent.run' do
        expect(Ralph::Agent).not_to receive(:run)
        described_class.run(['--dry-run'])
      end
    end

    context 'with --resume flag' do
      let(:prd_file) { Ralph::Config.get(:prd_file) }

      context 'when prd.json exists' do
        before do
          allow(Ralph::Agent).to receive(:resume).and_return(described_class::EXIT_SUCCESS)
        end

        after do
          File.delete(prd_file) if File.exist?(prd_file)
        end

        it 'calls Agent.resume' do
          File.write(prd_file, '{}')
          expect(Ralph::Agent).to receive(:resume)
          described_class.run(['--resume'])
        end

        it 'prints resume mode' do
          File.write(prd_file, '{}')
          expect { described_class.run(['--resume']) }
            .to output(/Resuming from:.*Mode: Resume implementation/m).to_stdout
        end
      end

      context 'when prd.json does not exist' do
        it 'shows error' do
          expect { described_class.run(['--resume']) }
            .to output(/Error:.*No.*prd.json.*found/).to_stdout
        end

        it 'returns failure exit code' do
          result = described_class.run(['--resume'])
          expect(result).to eq(described_class::EXIT_FAILURE)
        end

        it 'does not call Agent.resume' do
          expect(Ralph::Agent).not_to receive(:resume)
          described_class.run(['--resume'])
        end
      end
    end
  end

  describe '.show_help' do
    it 'includes usage examples' do
      expect { described_class.show_help }
        .to output(/Usage:/).to_stdout
    end

    it 'shows full implementation example' do
      expect { described_class.show_help }
        .to output(%r{\./bin/ralph "your feature description"}).to_stdout
    end

    it 'shows dry run example' do
      expect { described_class.show_help }
        .to output(/--dry-run/).to_stdout
    end

    it 'shows resume option' do
      expect { described_class.show_help }
        .to output(/--resume/).to_stdout
    end
  end

  describe 'exit codes' do
    it 'defines EXIT_SUCCESS as 0' do
      expect(described_class::EXIT_SUCCESS).to eq(0)
    end

    it 'defines EXIT_FAILURE as 1' do
      expect(described_class::EXIT_FAILURE).to eq(1)
    end

    it 'defines EXIT_PARTIAL as 2' do
      expect(described_class::EXIT_PARTIAL).to eq(2)
    end
  end
end
