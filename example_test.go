package conf_test

import (
	"errors"
	"fmt"
	"os"

	"github.com/ardanlabs/conf/v3"
)

// Demonstrates parsing configuration from command-line flags, environment variables, and
// displaying version and help information.
func Example() {
	// Define a configuration struct with version information
	// and a sample configuration option
	cfg := struct {
		conf.Version
		SomeOption string `conf:"help:an example option"`
	}{
		Version: conf.Version{
			Build: "v1.0.0",
			Desc:  "Example Application",
		},
	}

	// Simulate command-line arguments for demonstration
	os.Args = []string{"example", "--help"}

	// Parse configuration with strict flag validation
	info, err := conf.ParseWithOptions("", &cfg, conf.WithStrictFlags())
	if err != nil {
		switch {
		case errors.Is(err, conf.ErrVersionWanted):
			// Version was requested, display version information
			fmt.Println(info)
			return
		case errors.Is(err, conf.ErrHelpWanted):
			// When help is requested, display usage information; customize as needed.
			fmt.Println(info)
			return
		}

		fmt.Printf("ERROR: %s\n", err)
		return
	}

	fmt.Printf("Application (build: %s) started\n", cfg.Version.Build)
	fmt.Println("Configuration loaded successfully")

	// Output:
	// Usage: example [options...] [arguments...]
	//
	// OPTIONS
	//   -h, --help                     display this help message
	//       --some-option  <string>    an example option
	//   -v, --version                  display version
	//
	// ENVIRONMENT
	//   SOME_OPTION  <string>    an example option
}

// Demonstrates parsing configuration from command-line flags and environment variables.
func ExampleParse() {
	// Define a configuration struct with version information
	// and a sample configuration option
	cfg := struct {
		conf.Version
		SomeOption string `conf:"help:an example option"`
	}{
		Version: conf.Version{
			Build: "v1.0.0",
			Desc:  "Example Application",
		},
	}

	// Simulate command-line arguments for demonstration
	os.Args = []string{"example", "--some-option", "value"}

	// Parse configuration without strict flag validation
	info, err := conf.Parse("", &cfg)
	if err != nil {
		switch {
		case errors.Is(err, conf.ErrVersionWanted):
			// Version was requested, display version information
			fmt.Println(info)
			return
		case errors.Is(err, conf.ErrHelpWanted):
			// When help is requested, display usage information; customize as needed.
			fmt.Println(info)
			return
		}

		fmt.Printf("ERROR: %s\n", err)
		return
	}

	fmt.Printf("Application (build: %s) started\n", cfg.Version.Build)
	fmt.Println("SomeOption:", cfg.SomeOption)

	// Output:
	// Application (build: v1.0.0) started
	// SomeOption: value
}

// Demonstrates using ParseWithOptions and WithStrictFlags for strict flag validation, to catch
// unrecognized command-line options.
func ExampleWithStrictFlags() {
	cfg := struct {
		Port    int  `conf:"default:8080,help:server port"`
		Timeout int  `conf:"default:30,help:timeout in seconds"`
		Debug   bool `conf:"default:false,help:enable debug mode"`
	}{}

	// Simulate valid command-line arguments
	os.Args = []string{"app", "--port", "9000", "--debug",
		// The unknown flag will trigger an error with strict validation
		"--unknown-flag",
	}

	info, err := conf.ParseWithOptions("", &cfg, conf.WithStrictFlags())
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(info)
			return
		}

		fmt.Printf("ERROR: %s\n", err)
		return
	}

	fmt.Printf("Port: %d\n", cfg.Port)
	fmt.Printf("Timeout: %d\n", cfg.Timeout)
	fmt.Printf("Debug: %v\n", cfg.Debug)

	// Output:
	// ERROR: parsing config: unrecognized flag: --unknown-flag
}

// Demonstrates how to use the Version struct to add build and description information to your application.
func ExampleVersion() {
	cfg := struct {
		conf.Version
		AppName string `conf:"default:myapp,help:application name"`
	}{
		Version: conf.Version{
			Build: "v2.1.0",
			Desc:  "My Application",
		},
	}

	// Simulate version request
	os.Args = []string{"app", "--version"}

	info, err := conf.ParseWithOptions("", &cfg, conf.WithStrictFlags())
	if err != nil {
		if errors.Is(err, conf.ErrVersionWanted) {
			// Version information is returned in info string
			fmt.Println(info)
			return
		}
	}

	// Output:
	// Version: v2.1.0
	// My Application
}

// Demonstrates how to capture remaining command-line arguments after flag processing using the Args type.
func ExampleArgs() {
	cfg := struct {
		Port      int `conf:"default:8080,help:server port"`
		conf.Args     // Capture remaining command-line arguments after flags
	}{}

	// Simulate command-line arguments
	os.Args = []string{"app", "--port", "8081", "start", "verbose"}

	_, err := conf.ParseWithOptions("", &cfg, conf.WithStrictFlags())
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			return
		}
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	fmt.Printf("Port: %d\n", cfg.Port)
	fmt.Printf("Args: %v\n", cfg.Args)

	// Output:
	// Port: 8081
	// Args: [start verbose]
}
