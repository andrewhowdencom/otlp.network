package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/adrg/xdg"
	"github.com/andrewhowdencom/otlp.network/internal/collector"
	"github.com/andrewhowdencom/otlp.network/internal/server"
	"github.com/andrewhowdencom/otlp.network/internal/telemetry"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string
var Version = "dev"

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "otlp-network",
	Version: Version,
	Short:   "An exporter that exports Network statistics from the end (Linux) device",
	Long: `otlp-network is a CLI application that exports network statistics
from the Linux device it is running on.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Initialize telemetry
		shutdown, err := telemetry.Setup(cmd.Context(), telemetry.Config{
			ServiceName:    "otlp-network",
			ServiceVersion: Version,
			Endpoint:       viper.GetString("otel.endpoint"),
			Insecure:       viper.GetBool("otel.insecure"),
			Interval:       viper.GetDuration("otel.interval"),
		})
		if err != nil {
			return err
		}
		// Ensure telemetry is shut down when the application exits
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := shutdown(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "failed to shutdown telemetry: %v\n", err)
			}
		}()

		// Start Uptime Collector (Always enabled for now, or could match others)
		uptime := collector.NewUptime()
		if err := uptime.Start(cmd.Context()); err != nil {
			return err
		}

		// Device Collector
		if viper.GetBool("collector.device.enabled") {
			c, err := collector.NewDevice()
			if err != nil {
				return err
			}
			if err := c.Start(cmd.Context()); err != nil {
				return err
			}
		}

		// Wifi Collector
		if viper.GetBool("collector.wifi.enabled") {
			c, err := collector.NewWifi()
			if err != nil {
				return err
			}
			if err := c.Start(cmd.Context()); err != nil {
				return err
			}
		}

		// TCP Collector
		if viper.GetBool("collector.tcp.enabled") {
			c, err := collector.NewTCP()
			if err != nil {
				return err
			}
			if err := c.Start(cmd.Context()); err != nil {
				return err
			}
		}

		// UDP Collector
		if viper.GetBool("collector.udp.enabled") {
			c, err := collector.NewUDP()
			if err != nil {
				return err
			}
			if err := c.Start(cmd.Context()); err != nil {
				return err
			}
		}

		// Conntrack Collector
		if viper.GetBool("collector.conntrack.enabled") {
			c, err := collector.NewConntrack()
			if err != nil {
				return err
			}
			if err := c.Start(cmd.Context()); err != nil {
				return err
			}
		}

		// Softnet Collector
		if viper.GetBool("collector.softnet.enabled") {
			c, err := collector.NewSoftnet()
			if err != nil {
				return err
			}
			if err := c.Start(cmd.Context()); err != nil {
				return err
			}
		}

		// Sockstat Collector
		if viper.GetBool("collector.sockstat.enabled") {
			c, err := collector.NewSockstat()
			if err != nil {
				return err
			}
			if err := c.Start(cmd.Context()); err != nil {
				return err
			}
		}

		// Start Prometheus Metrics Server
		srv, err := server.New(viper.GetString("prometheus.host"), viper.GetInt("prometheus.port"))
		if err != nil {
			return err
		}
		if err := srv.Start(); err != nil {
			return err
		}

		// Wait for context cancellation (SIGINT/SIGTERM)
		<-cmd.Context().Done()
		fmt.Println("shutting down...")

		// Graceful shutdown for the server
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			return err
		}

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// Create a cancellable context for graceful shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	err := rootCmd.ExecuteContext(ctx)
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $XDG_CONFIG_HOME/otlp-network/otlp-network.yaml)")

	rootCmd.PersistentFlags().String("otel.endpoint", "", "OpenTelemetry exporter endpoint (HTTP)")
	rootCmd.PersistentFlags().Bool("otel.insecure", true, "Use insecure connection for OpenTelemetry")
	rootCmd.PersistentFlags().Duration("otel.interval", 60*time.Second, "OpenTelemetry export interval")
	rootCmd.PersistentFlags().String("prometheus.host", "", "Host to expose Prometheus metrics (empty for all interfaces)")
	rootCmd.PersistentFlags().Int("prometheus.port", 9464, "Port for Prometheus metrics")

	rootCmd.PersistentFlags().Bool("collector.device.enabled", false, "Enable device collector")
	rootCmd.PersistentFlags().Bool("collector.wifi.enabled", false, "Enable wifi collector")
	rootCmd.PersistentFlags().Bool("collector.tcp.enabled", false, "Enable tcp collector")
	rootCmd.PersistentFlags().Bool("collector.udp.enabled", false, "Enable udp collector")
	rootCmd.PersistentFlags().Bool("collector.conntrack.enabled", false, "Enable conntrack collector")
	rootCmd.PersistentFlags().Bool("collector.softnet.enabled", false, "Enable softnet collector")
	rootCmd.PersistentFlags().Bool("collector.sockstat.enabled", false, "Enable sockstat collector")

	viper.BindPFlag("otel.endpoint", rootCmd.PersistentFlags().Lookup("otel.endpoint"))
	viper.BindPFlag("otel.insecure", rootCmd.PersistentFlags().Lookup("otel.insecure"))
	viper.BindPFlag("otel.interval", rootCmd.PersistentFlags().Lookup("otel.interval"))
	viper.BindPFlag("prometheus.host", rootCmd.PersistentFlags().Lookup("prometheus.host"))
	viper.BindPFlag("prometheus.port", rootCmd.PersistentFlags().Lookup("prometheus.port"))

	viper.BindPFlag("collector.device.enabled", rootCmd.PersistentFlags().Lookup("collector.device.enabled"))
	viper.BindPFlag("collector.wifi.enabled", rootCmd.PersistentFlags().Lookup("collector.wifi.enabled"))
	viper.BindPFlag("collector.tcp.enabled", rootCmd.PersistentFlags().Lookup("collector.tcp.enabled"))
	viper.BindPFlag("collector.udp.enabled", rootCmd.PersistentFlags().Lookup("collector.udp.enabled"))
	viper.BindPFlag("collector.conntrack.enabled", rootCmd.PersistentFlags().Lookup("collector.conntrack.enabled"))
	viper.BindPFlag("collector.softnet.enabled", rootCmd.PersistentFlags().Lookup("collector.softnet.enabled"))
	viper.BindPFlag("collector.sockstat.enabled", rootCmd.PersistentFlags().Lookup("collector.sockstat.enabled"))

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Search for config in XDG directories and system directories.
		// 1. User config: $XDG_CONFIG_HOME/otlp-network/otlp-network.yaml
		// 2. System config: $XDG_CONFIG_DIRS/otlp-network/otlp-network.yaml
		// 3. Fallback: /etc/otlp-network/otlp-network.yaml
		// 4. Fallback: ./otlp-network.yaml

		viper.SetConfigName("otlp-network")
		viper.SetConfigType("yaml")

		// 1. User configuration directory
		viper.AddConfigPath(filepath.Join(xdg.ConfigHome, "otlp-network"))

		// 2. System configuration directories
		for _, dir := range xdg.ConfigDirs {
			viper.AddConfigPath(filepath.Join(dir, "otlp-network"))
		}

		// 3. Fallbacks
		viper.AddConfigPath("/etc/otlp-network")
		viper.AddConfigPath(".")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
}
