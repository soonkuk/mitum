package main

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"runtime/trace"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/diode"
	"github.com/spf13/cobra"
)

var (
	sigc           chan os.Signal
	exitHooks      []func()
	exitCode       int = 0
	memProfileFile *os.File
	traceFile      *os.File
	log            zerolog.Logger
	logOutput      io.Writer
	stdoutLog      zerolog.Logger
)

var rootCmd = &cobra.Command{
	Use:   "contest",
	Short: "contest is the consensus tester of ISAAC+",
	Args:  cobra.NoArgs,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if FlagLogOut == "null" {
			logOutput = nil
		} else if len(FlagLogOut) < 1 {
			logOutput = os.Stdout
		} else {
			// check FlagLogOut is directory
			fi, err := os.Stat(FlagLogOut)
			if err != nil {
				cmd.Println("Error:", err.Error())
				os.Exit(1)
			}

			var dir string
			switch mode := fi.Mode(); {
			case mode.IsDir():
			case mode.IsRegular():
				dir = filepath.Base(FlagLogOut)
			}

			if f, err := os.Create(filepath.Join(dir, "all.log")); err != nil {
				cmd.Println("Error:", err.Error())
				os.Exit(1)
			} else {
				logOutput = f
			}
		}

		if flagLogLevel.lvl == zerolog.DebugLevel {
			logOutput = diode.NewWriter(
				logOutput,
				1000,
				0,
				func(missed int) {
					fmt.Fprintf(os.Stderr, "dropped %d messages", missed)
				},
			)
		}

		logContext := zerolog.
			New(os.Stdout).
			With().
			Timestamp()

		if flagLogLevel.lvl == zerolog.DebugLevel {
			logContext = logContext.
				Caller().
				Stack()
		}

		stdoutLog = logContext.Logger().Level(flagLogLevel.lvl)
		log = stdoutLog.Output(logOutput)

		log.Debug().
			RawJSON("flags", printFlagsJSON(cmd)).
			Msg("parsed flags")

		startProfile()

		sigc = make(chan os.Signal, 1)
		signal.Notify(sigc,
			syscall.SIGTERM,
			syscall.SIGQUIT,
		)

		go func() {
			s := <-sigc

			for _, h := range exitHooks {
				h()
			}

			log.Info().
				Str("sig", s.String()).
				Int("exit", exitCode).
				Msg("contest stopped by force")

			if l, ok := logOutput.(diode.Writer); ok {
				_ = l.Close()
			}

			os.Exit(exitCode)
		}()
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		for _, h := range exitHooks {
			h()
		}

		log.Info().
			Int("exit", exitCode).
			Msg("contest stopped")
		os.Exit(exitCode)
	},
}

func startProfile() {
	if len(flagTrace) > 0 {
		f, err := os.Create(flagTrace)
		if err != nil {
			panic(err)
		}
		if err := trace.Start(f); err != nil {
			panic(err)
		}
		traceFile = f
		log.Debug().Msg("trace enabled")
	}

	if len(flagCPUProfile) > 0 {
		f, err := os.Create(flagCPUProfile)
		if err != nil {
			panic(err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			panic(err)
		}
		log.Debug().Msg("cpuprofile enabled")
	}

	if len(flagMemProfile) > 0 {
		f, err := os.Create(flagMemProfile)
		if err != nil {
			panic(err)
		}
		if err := pprof.WriteHeapProfile(f); err != nil {
			panic(err)
		}
		memProfileFile = f
		log.Debug().Msg("memprofile enabled")
	}
}

func closeProfile() {
	if len(flagCPUProfile) > 0 {
		pprof.StopCPUProfile()
		log.Debug().Msg("cpu profile closed")
	}

	if len(flagMemProfile) > 0 {
		if err := memProfileFile.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close mem profile file")
		}

		log.Debug().Msg("mem profile closed")
	}

	if len(flagTrace) > 0 {
		trace.Stop()
		if err := traceFile.Close(); err != nil {
			log.Error().Err(err).Msg("failed to close trace file")
		}

		log.Debug().Msg("trace closed")
	}
}

func init() {
	exitHooks = append(exitHooks, closeProfile)
}

func main() {
	rootCmd.PersistentFlags().Var(&flagLogLevel, "log-level", "log level: {debug error warn info crit}")
	rootCmd.PersistentFlags().Var(&flagLogFormat, "log-format", "log format: {json terminal}")
	rootCmd.PersistentFlags().StringVar(&FlagLogOut, "log", FlagLogOut, "log output directory")
	rootCmd.PersistentFlags().StringVar(&flagCPUProfile, "cpuprofile", flagCPUProfile, "write cpu profile to file")
	rootCmd.PersistentFlags().StringVar(&flagMemProfile, "memprofile", flagMemProfile, "write memory profile to file")
	rootCmd.PersistentFlags().StringVar(&flagTrace, "trace", flagTrace, "write trace to file")
	rootCmd.PersistentFlags().BoolVar(&flagQuiet, "quiet", flagQuiet, "quiet")

	if err := rootCmd.Execute(); err != nil {
		rootCmd.Println("Error:", err.Error())
		os.Exit(1)
	}
}
