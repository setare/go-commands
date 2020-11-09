package commands

import (
	"context"
	"errors"
	"os"

	signals "github.com/setare/go-os-signals"
	"github.com/setare/services"
	zapreporter "github.com/setare/services-reporter-zap"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

type CommandFunc = func(*cobra.Command, []string)
type CommandEFunc = func(*cobra.Command, []string) error

type CmdBuilder interface {
	Use(string) CmdBuilder
	Short(string) CmdBuilder
	Long(string) CmdBuilder
	WithServices(...services.Service) CmdBuilder
	WithLogger(*zap.Logger) CmdBuilder
	WithRetrierBuilder(*services.RetrierBuilder) CmdBuilder
	WithSignalListener(signals.Listener) CmdBuilder
	DisableSignalListener(bool) CmdBuilder
	Run(CommandFunc) CmdBuilder
	Build() *cobra.Command
}

type cmdBuilder struct {
	use                    string
	short                  string
	long                   string
	beforeRun              CommandEFunc
	run                    CommandFunc
	logger                 *zap.Logger
	retrierBuilder         *services.RetrierBuilder
	signalListenerDisabled bool
	signalListener         signals.Listener
	services               []services.Service
}

func CommandBuilder() CmdBuilder {
	return &cmdBuilder{}
}

func (builder *cmdBuilder) Use(use string) CmdBuilder {
	builder.use = use
	return builder
}

func (builder *cmdBuilder) Short(short string) CmdBuilder {
	builder.short = short
	return builder
}

func (builder *cmdBuilder) Long(long string) CmdBuilder {
	builder.long = long
	return builder
}

func (builder *cmdBuilder) BeforeRun(beforeRun CommandEFunc) CmdBuilder {
	builder.beforeRun = beforeRun
	return builder
}

func (builder *cmdBuilder) Run(run CommandFunc) CmdBuilder {
	builder.run = run
	return builder
}

func (builder *cmdBuilder) WithServices(services ...services.Service) CmdBuilder {
	builder.services = services
	return builder
}

func (builder *cmdBuilder) WithLogger(logger *zap.Logger) CmdBuilder {
	builder.logger = logger
	return builder
}

func (builder *cmdBuilder) WithRetrierBuilder(retrierBuilder *services.RetrierBuilder) CmdBuilder {
	builder.retrierBuilder = retrierBuilder
	return builder
}

func (builder *cmdBuilder) DisableSignalListener(disabled bool) CmdBuilder {
	builder.signalListenerDisabled = disabled
	return builder
}

func (builder *cmdBuilder) WithSignalListener(listener signals.Listener) CmdBuilder {
	builder.signalListener = listener
	return builder
}

func (builder *cmdBuilder) Build() *cobra.Command {
	b := *builder
	cmd := &cobra.Command{}
	cmd.Use = b.use
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		logger := b.logger
		reporter := zapreporter.NewReporter(logger)

		if b.retrierBuilder != nil {
			for i, service := range b.services {
				b.services[i] = b.retrierBuilder.Build(service)
			}
		}

		if b.beforeRun != nil {
			err := b.beforeRun(cmd, args)
			if err != nil {
				return err
			}
		}

		starter := services.NewStarter(
			b.services...,
		).WithReporter(reporter)
		err := starter.Start()
		if err != nil && errors.Is(err, context.Canceled) {
			logger.Warn("initialization canceled")
			os.Exit(2)
		} else if err != nil {
			logger.Error("could not start the service:", zap.Error(err))
			os.Exit(1)
		}

		if b.run != nil {
			b.run(cmd, args)
		}

		// Skip listening signal
		if b.signalListenerDisabled {
			return nil
		}

		listener := builder.signalListener
		if listener == nil {
			listener = signals.NewListener(os.Interrupt)
		}

		err = starter.ListenSignals(listener)
		if err != nil && err != context.Canceled {
			logger.Error("error listening signals:", zap.Error(err))
			os.Exit(3)
		}

		return nil
	}
	return cmd
}
