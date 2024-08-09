package conf_test

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ardanlabs/conf/v3"
	"github.com/ardanlabs/conf/v3/yaml"
	"github.com/google/go-cmp/cmp"
)

const (
	success = "\u2713"
	failed  = "\u2717"
)

// =============================================================================

// CustomValue provides support for testing a custom value.
type CustomValue struct {
	something string
}

// Set implements the Setter interface
func (c *CustomValue) Set(data string) error {
	*c = CustomValue{something: fmt.Sprintf("@%s@", data)}
	return nil
}

// String implements the Stringer interface
func (c CustomValue) String() string {
	return c.something
}

// Equal implements the Equal "interface" for go-cmp
func (c CustomValue) Equal(o CustomValue) bool {
	return c.something == o.something
}

// =============================================================================

type ip struct {
	Name      string   `conf:"default:localhost,env:IP_NAME_VAR"`
	IP        string   `conf:"default:127.0.0.0"`
	Endpoints []string `conf:"default:127.0.0.1:200;127.0.0.1:829"`
}

type Embed struct {
	Name     string        `conf:"default:bill"`
	Duration time.Duration `conf:"default:1s,flag:e-dur,short:d"`
}

type config struct {
	AnInt     int    `conf:"default:9"`
	AString   string `conf:"default:B,short:s"`
	Bool      bool
	Skip      string `conf:"-"`
	IP        ip
	DebugHost string      `conf:"default:http://user:password@0.0.0.0:4000,mask"`
	Password  string      `conf:"default:password,mask"`
	Custom    CustomValue `conf:"default:hello"`
	Embed
}

// =============================================================================
func TestRequired(t *testing.T) {
	t.Logf("\tTest: %d\tWhen required values are missing.", 1)
	{
		f := func(t *testing.T) {
			os.Args = []string{"conf.test"}
			var cfg struct {
				TestInt    int `conf:"required, default:1"`
				TestString string
				TestBool   bool
			}
			_, err := conf.Parse("TEST", &cfg)
			if err == nil {
				t.Fatalf("\t%s\tShould fail for missing required value.", failed)
			}
			t.Logf("\t%s\tShould fail for missing required value : %s", success, err)
		}
		t.Run("required-missing-value", f)
	}

	t.Logf("\tTest: %d\tWhen required env integer is zero.", 1)
	{
		f := func(t *testing.T) {
			os.Args = []string{"conf.test"}
			os.Setenv("TEST_TEST_INT", "0")

			var cfg struct {
				TestInt    int `conf:"required"`
				TestString string
				TestBool   bool
			}
			_, err := conf.Parse("TEST", &cfg)
			if err != nil {
				t.Fatalf("\t%s\tShould have parsed the required zero env integer : %s", failed, err)
			}
			t.Logf("\t%s\tShould have parsed the required zero env integer.", success)
		}
		t.Run("required-env-integer-zero", f)
	}

	t.Logf("\tTest: %d\tWhen required env string is empty.", 1)
	{
		f := func(t *testing.T) {
			os.Args = []string{"conf.test"}
			os.Setenv("TEST_TEST_STRING", "")

			var cfg struct {
				TestInt    int
				TestString string `conf:"required"`
				TestBool   bool
			}
			_, err := conf.Parse("TEST", &cfg)
			if err != nil {
				t.Fatalf("\t%s\tShould have parsed the required empty env string : %s", failed, err)
			}
			t.Logf("\t%s\tShould have parsed the required empty env string.", success)
		}
		t.Run("required-env-string-empty", f)
	}

	t.Logf("\tTest: %d\tWhen required env boolean is false.", 1)
	{
		f := func(t *testing.T) {
			os.Args = []string{"conf.test"}
			os.Setenv("TEST_TEST_BOOL", "false")

			var cfg struct {
				TestInt    int
				TestString string
				TestBool   bool `conf:"required"`
			}
			_, err := conf.Parse("TEST", &cfg)
			if err != nil {
				t.Fatalf("\t%s\tShould have parsed the required false env boolean : %s", failed, err)
			}
			t.Logf("\t%s\tShould have parsed the required false env boolean.", success)
		}
		t.Run("required-env-boolean-false", f)
	}

	t.Logf("\tTest: %d\tWhen struct has no fields.", 2)
	{
		f := func(t *testing.T) {
			os.Args = []string{"conf.test"}
			var cfg struct {
				testInt    int `conf:"required, default:1"`
				testString string
				testBool   bool
			}
			_, err := conf.Parse("TEST", &cfg)
			if err == nil {
				t.Fatalf("\t%s\tShould fail for struct with no exported fields.", failed)
			}
			t.Logf("\t%s\tShould fail for struct with no exported fields : %s", success, err)
		}
		t.Run("struct-missing-fields", f)
	}

	t.Logf("\tTest: %d\tWhen required values exist and are passed on args.", 3)
	{
		f := func(t *testing.T) {
			os.Args = []string{"conf.test", "--test-int", "1"}

			var cfg struct {
				TestInt    int `conf:"required, default:1"`
				TestString string
				TestBool   bool
			}
			_, err := conf.Parse("TEST", &cfg)
			if err != nil {
				t.Fatalf("\t%s\tShould have parsed the required field on args : %s", failed, err)
			}
			t.Logf("\t%s\tShould have parsed the required field on args.", success)
		}
		t.Run("required-existing-fields-args", f)
	}

	t.Logf("\tTest: %d\tWhen required values exist and are passed on env.", 4)
	{
		f := func(t *testing.T) {
			os.Args = []string{"conf.test"}
			os.Setenv("TEST_TEST_INT", "1")

			var cfg struct {
				TestInt    int `conf:"required, default:1"`
				TestString string
				TestBool   bool
			}
			_, err := conf.Parse("TEST", &cfg)
			if err != nil {
				t.Fatalf("\t%s\tShould have parsed the required field on Env : %s", failed, err)
			}
			t.Logf("\t%s\tShould have parsed the required field on Env.", success)
		}
		t.Run("required-existing-fields-args", f)
	}
}

func TestParse(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		args []string
		want config
	}{
		{
			"default",
			nil,
			nil,
			config{9, "B", false, "", ip{"localhost", "127.0.0.0", []string{"127.0.0.1:200", "127.0.0.1:829"}}, "http://user:password@0.0.0.0:4000", "password", CustomValue{something: "@hello@"}, Embed{"bill", time.Second}},
		},
		{
			"env",
			map[string]string{"TEST_AN_INT": "1", "TEST_A_STRING": "s", "TEST_BOOL": "TRUE", "TEST_SKIP": "SKIP", "TEST_IP_NAME_VAR": "local", "TEST_DEBUG_HOST": "http://bill:gopher@0.0.0.0:4000", "TEST_PASSWORD": "gopher", "TEST_NAME": "andy", "TEST_DURATION": "1m"},
			nil,
			config{1, "s", true, "", ip{"local", "127.0.0.0", []string{"127.0.0.1:200", "127.0.0.1:829"}}, "http://bill:gopher@0.0.0.0:4000", "gopher", CustomValue{something: "@hello@"}, Embed{"andy", time.Minute}},
		},
		{
			"flag",
			nil,
			[]string{"conf.test", "--an-int", "1", "-s", "s", "--bool", "--skip", "skip", "--ip-name", "local", "--debug-host", "http://bill:gopher@0.0.0.0:4000", "--password", "gopher", "--name", "andy", "--e-dur", "1m"},
			config{1, "s", true, "", ip{"local", "127.0.0.0", []string{"127.0.0.1:200", "127.0.0.1:829"}}, "http://bill:gopher@0.0.0.0:4000", "gopher", CustomValue{something: "@hello@"}, Embed{"andy", time.Minute}},
		},
		{
			"multi",
			map[string]string{"TEST_A_STRING": "s", "TEST_BOOL": "TRUE", "TEST_IP_NAME_VAR": "local", "TEST_DEBUG_HOST": "http://bill:gopher@0.0.0.0:4000", "TEST_PASSWORD": "gopher", "TEST_NAME": "andy", "TEST_DURATION": "1m"},
			[]string{"conf.test", "--an-int", "2", "--bool", "--skip", "skip", "--name", "jack", "-d", "1ms"},
			config{2, "s", true, "", ip{"local", "127.0.0.0", []string{"127.0.0.1:200", "127.0.0.1:829"}}, "http://bill:gopher@0.0.0.0:4000", "gopher", CustomValue{something: "@hello@"}, Embed{"jack", time.Millisecond}},
		},
	}

	t.Log("Given the need to parse basic configuration.")
	{
		for i, tt := range tests {
			t.Logf("\tTest: %d\tWhen checking with arguments %v", i, tt.args)
			{
				os.Clearenv()
				for k, v := range tt.envs {
					os.Setenv(k, v)
				}

				f := func(t *testing.T) {
					os.Args = tt.args

					var cfg config
					if _, err := conf.Parse("TEST", &cfg); err != nil {
						t.Fatalf("\t%s\tShould be able to Parse arguments : %s.", failed, err)
					}
					t.Logf("\t%s\tShould be able to Parse arguments.", success)

					if diff := cmp.Diff(tt.want, cfg); diff != "" {
						t.Fatalf("\t%s\tShould have properly initialized struct value\n%s", failed, diff)
					}
					t.Logf("\t%s\tShould have properly initialized struct value.", success)
				}

				t.Run(tt.name, f)
			}
		}
	}
}

func TestParseEmptyNamespace(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		args []string
		want config
	}{
		{
			"env",
			map[string]string{"AN_INT": "1", "A_STRING": "s", "BOOL": "TRUE", "SKIP": "SKIP", "IP_NAME_VAR": "local", "DEBUG_HOST": "http://bill:gopher@0.0.0.0:4000", "PASSWORD": "gopher", "NAME": "andy", "DURATION": "1m"},
			nil,
			config{1, "s", true, "", ip{"local", "127.0.0.0", []string{"127.0.0.1:200", "127.0.0.1:829"}}, "http://bill:gopher@0.0.0.0:4000", "gopher", CustomValue{something: "@hello@"}, Embed{"andy", time.Minute}},
		},
	}

	t.Log("Given the need to parse basic configuration.")
	{
		for i, tt := range tests {
			t.Logf("\tTest: %d\tWhen checking with arguments %v", i, tt.args)
			{
				os.Clearenv()
				for k, v := range tt.envs {
					os.Setenv(k, v)
				}

				f := func(t *testing.T) {
					os.Args = tt.args

					var cfg config
					if _, err := conf.Parse("", &cfg); err != nil {
						t.Fatalf("\t%s\tShould be able to Parse arguments : %s.", failed, err)
					}
					t.Logf("\t%s\tShould be able to Parse arguments.", success)

					if diff := cmp.Diff(tt.want, cfg); diff != "" {
						t.Fatalf("\t%s\tShould have properly initialized struct value\n%s", failed, diff)
					}
					t.Logf("\t%s\tShould have properly initialized struct value.", success)
				}

				t.Run(tt.name, f)
			}
		}
	}
}

func TestParse_Args(t *testing.T) {
	t.Log("Given the need to capture remaining command line arguments after flags.")
	{
		type configArgs struct {
			Port int
			Args conf.Args
		}

		want := configArgs{
			Port: 9000,
			Args: conf.Args{"migrate", "seed"},
		}

		os.Args = []string{"conf.test", "--port", "9000", "migrate", "seed"}

		var cfg configArgs
		if _, err := conf.Parse("TEST", &cfg); err != nil {
			t.Fatalf("\t%s\tShould be able to Parse arguments : %s.", failed, err)
		}
		t.Logf("\t%s\tShould be able to Parse arguments.", success)

		if diff := cmp.Diff(want, cfg); diff != "" {
			t.Fatalf("\t%s\tShould have properly initialized struct value\n%s", failed, diff)
		}
		t.Logf("\t%s\tShould have properly initialized struct value.", success)
	}
}

func TestErrors(t *testing.T) {
	t.Log("Given the need to validate errors that can occur with Parse.")
	{
		t.Logf("\tTest: %d\tWhen passing bad values to Parse.", 0)
		{
			f := func(t *testing.T) {
				os.Args = []string{"conf.test"}

				var cfg struct {
					TestInt    int
					TestString string
					TestBool   bool
				}
				_, err := conf.Parse("TEST", cfg)
				if err == nil {
					t.Fatalf("\t%s\tShould NOT be able to accept a value by value.", failed)
				}
				t.Logf("\t%s\tShould NOT be able to accept a value by value : %s", success, err)
			}
			t.Run("not-by-ref", f)

			f = func(t *testing.T) {
				os.Args = []string{"conf.test"}

				var cfg []string
				_, err := conf.Parse("TEST", &cfg)
				if err == nil {
					t.Fatalf("\t%s\tShould NOT be able to pass anything but a struct value.", failed)
				}
				t.Logf("\t%s\tShould NOT be able to pass anything but a struct value : %s", success, err)
			}
			t.Run("not-struct-value", f)
		}

		t.Logf("\tTest: %d\tWhen bad tags to Parse.", 1)
		{
			f := func(t *testing.T) {
				os.Args = []string{"conf.test"}

				var cfg struct {
					TestInt    int `conf:"default:"`
					TestString string
					TestBool   bool
				}
				_, err := conf.Parse("TEST", &cfg)
				if err == nil {
					t.Fatalf("\t%s\tShould NOT be able to accept tag missing value.", failed)
				}
				t.Logf("\t%s\tShould NOT be able to accept tag missing value : %s", success, err)
			}
			t.Run("tag-missing-value", f)

			f = func(t *testing.T) {
				os.Args = []string{"conf.test"}

				var cfg struct {
					TestInt    int `conf:"short:ab"`
					TestString string
					TestBool   bool
				}
				_, err := conf.Parse("TEST", &cfg)
				if err == nil {
					t.Fatalf("\t%s\tShould NOT be able to accept invalid short tag.", failed)
				}
				t.Logf("\t%s\tShould NOT be able to accept invalid short tag : %s", success, err)
			}
			t.Run("tag-bad-short", f)
		}
	}
}

var withNamespace = `Usage: conf.test [options] [arguments]

OPTIONS
  --an-int/$TEST_AN_INT              <int>                 (default: 9)
  --a-string/-s/$TEST_A_STRING       <string>              (default: B)
  --bool/$TEST_BOOL                  <bool>                
  --ip-name/$TEST_IP_NAME_VAR        <string>              (default: localhost)
  --ip-ip/$TEST_IP_IP                <string>              (default: 127.0.0.0)
  --ip-endpoints/$TEST_IP_ENDPOINTS  <string>,[string...]  (default: 127.0.0.1:200;127.0.0.1:829)
  --debug-host/$TEST_DEBUG_HOST      <string>              (default: http://xxxxxx:xxxxxx@0.0.0.0:4000)
  --password/$TEST_PASSWORD          <string>              (default: xxxxxx)
  --custom/$TEST_CUSTOM              <value>               (default: hello)
  --name/$TEST_NAME                  <string>              (default: bill)
  --e-dur/-d/$TEST_DURATION          <duration>            (default: 1s)
  --help/-h                          
  display this help message`

var emptyNamespace = `Usage: conf.test [options] [arguments]

OPTIONS
  --an-int/$AN_INT              <int>                 (default: 9)
  --a-string/-s/$A_STRING       <string>              (default: B)
  --bool/$BOOL                  <bool>                
  --ip-name/$IP_NAME_VAR        <string>              (default: localhost)
  --ip-ip/$IP_IP                <string>              (default: 127.0.0.0)
  --ip-endpoints/$IP_ENDPOINTS  <string>,[string...]  (default: 127.0.0.1:200;127.0.0.1:829)
  --debug-host/$DEBUG_HOST      <string>              (default: http://xxxxxx:xxxxxx@0.0.0.0:4000)
  --password/$PASSWORD          <string>              (default: xxxxxx)
  --custom/$CUSTOM              <value>               (default: hello)
  --name/$NAME                  <string>              (default: bill)
  --e-dur/-d/$DURATION          <duration>            (default: 1s)
  --help/-h                     
  display this help message`

var withNamespaceOptions = `Usage: conf.test [options] [arguments]

OPTIONS
  --port/$TEST_PORT  <int>  
  --help/-h          
  display this help message`

var emptyNamespaceOptions = `Usage: conf.test [options] [arguments]

OPTIONS
  --port/$PORT  <int>  
  --help/-h     
  display this help message`

func TestUsage(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		envs      map[string]string
		want      string
		options   string
	}{
		{
			name:      "with-namespace",
			namespace: "TEST",
			envs:      map[string]string{"TEST_AN_INT": "1", "TEST_A_STRING": "s", "TEST_BOOL": "TRUE", "TEST_SKIP": "SKIP", "TEST_IP_NAME_VAR": "local", "TEST_NAME": "andy", "TEST_DURATION": "1m"},
			want:      withNamespace,
			options:   withNamespaceOptions,
		},
		{
			name:      "empty-namespace",
			namespace: "",
			envs:      map[string]string{"AN_INT": "1", "A_STRING": "s", "BOOL": "TRUE", "SKIP": "SKIP", "IP_NAME_VAR": "local", "NAME": "andy", "DURATION": "1m"},
			want:      emptyNamespace,
			options:   emptyNamespaceOptions,
		},
	}

	t.Log("Given the need validate usage output.")
	{
		for testID, tt := range tests {
			f := func(t *testing.T) {
				t.Logf("\tTest: %d\tWhen testing %s", testID, tt.name)
				{
					os.Clearenv()
					for k, v := range tt.envs {
						os.Setenv(k, v)
					}

					os.Args = []string{"conf.test"}

					var cfg config
					if _, err := conf.Parse(tt.namespace, &cfg); err != nil {
						fmt.Print(err)
						return
					}

					got, err := conf.UsageInfo(tt.namespace, &cfg)
					if err != nil {
						fmt.Print(err)
						return
					}

					got = strings.TrimRight(got, " \n")
					t.Log(got)
					gotS := strings.Split(got, "\n")
					wantS := strings.Split(tt.want, "\n")
					if diff := cmp.Diff(gotS, wantS); diff != "" {
						t.Errorf("\t%s\tShould match the output byte for byte. See diff:", failed)
						t.Log(diff)
					}
					t.Logf("\t%s\tShould match byte for byte the output.", success)
				}

				t.Logf("\tTest: %d\tWhen using a struct with arguments.", 1)
				{
					var cfg struct {
						Port int
						Args conf.Args
					}

					got, err := conf.UsageInfo(tt.namespace, &cfg)
					if err != nil {
						fmt.Print(err)
						return
					}

					got = strings.TrimRight(got, " \n")
					gotS := strings.Split(got, "\n")
					wantS := strings.Split(tt.options, "\n")
					if diff := cmp.Diff(gotS, wantS); diff != "" {
						t.Errorf("\t%s\tShould match the output byte for byte. See diff:", failed)
						t.Log(diff)
					}
					t.Logf("\t%s\tShould match byte for byte the output.", success)
				}
			}

			t.Run(tt.name, f)
		}
	}
}

func ExampleString() {
	tt := struct {
		envs map[string]string
	}{
		envs: map[string]string{"TEST_AN_INT": "1", "TEST_S": "s", "TEST_BOOL": "TRUE", "TEST_SKIP": "SKIP", "TEST_IP_NAME": "local", "TEST_NAME": "andy", "TEST_DURATION": "1m"},
	}

	os.Clearenv()
	for k, v := range tt.envs {
		os.Setenv(k, v)
	}

	os.Args = []string{"conf.test"}

	var cfg config
	if _, err := conf.Parse("TEST", &cfg); err != nil {
		fmt.Print(err)
		return
	}

	out, err := conf.String(&cfg)
	if err != nil {
		fmt.Print(err)
		return
	}

	fmt.Print(out)

	// Output:
	// --an-int=1
	// --a-string/-s=B
	// --bool=true
	// --ip-name=localhost
	// --ip-ip=127.0.0.0
	// --ip-endpoints=[127.0.0.1:200 127.0.0.1:829]
	// --debug-host=http://xxxxxx:xxxxxx@0.0.0.0:4000
	// --password=xxxxxx
	// --custom=@hello@
	// --name=andy
	// --e-dur/-d=1m0s
}

type ConfExplicit struct {
	Version conf.Version
	Address string
}

type ConfImplicit struct {
	conf.Version
	Address string
}

func TestVersionExplicit(t *testing.T) {
	tests := []struct {
		name    string
		config  ConfExplicit
		args    []string
		want    string
		wantErr bool
	}{
		{
			name: "version",
			args: []string{"--version"},
			config: ConfExplicit{
				Version: conf.Version{
					Build: "v1.0.0",
				},
			},
			wantErr: false,
			want:    "Version: v1.0.0",
		},
		{
			name: "vershort",
			args: []string{"conf.test", "-v"},
			config: ConfExplicit{
				Version: conf.Version{
					Build: "v1.0.0",
				},
			},
			wantErr: false,
			want:    "Version: v1.0.0",
		},
		{
			name: "verdes",
			args: []string{"conf.test", "-version"},
			config: ConfExplicit{
				Version: conf.Version{
					Build: "v1.0.0",
					Desc:  "Service Description",
				},
			},
			wantErr: false,
			want:    "Version: v1.0.0\nService Description",
		},
		{
			name: "verdesshort",
			args: []string{"conf.test", "-v"},
			config: ConfExplicit{
				Version: conf.Version{
					Build: "v1.0.0",
					Desc:  "Service Description",
				},
			},
			wantErr: false,
			want:    "Version: v1.0.0\nService Description",
		},
		{
			name: "desshort",
			args: []string{"conf.test", "-v"},
			config: ConfExplicit{
				Version: conf.Version{
					Desc: "Service Description",
				},
			},
			wantErr: false,
			want:    "Service Description",
		},
		{
			name:    "none",
			args:    []string{"conf.test", "-v"},
			config:  ConfExplicit{},
			want:    "",
			wantErr: false,
		},
	}

	t.Log("Given the need validate version output.")
	{
		for i, tt := range tests {
			t.Logf("\tTest: %d\tWhen using an explicit struct.", i)
			{
				f := func(t *testing.T) {
					os.Args = tt.args
					if help, err := conf.Parse("APP", &tt.config); err != nil {
						if err == conf.ErrHelpWanted {
							if diff := cmp.Diff(tt.want, help); diff != "" {
								t.Errorf("\t%s\tShould match the output byte for byte. See diff:", failed)
								t.Log(diff)
							}
							t.Logf("\t%s\tShould match byte for byte the output.", success)
						}
					}
				}

				t.Run(tt.name, f)
			}
		}
	}
}

func TestVersionImplicit(t *testing.T) {
	tests := []struct {
		name    string
		config  ConfImplicit
		args    []string
		want    string
		wantErr bool
	}{
		{
			name: "only version",
			args: []string{"conf.test", "--version"},
			config: ConfImplicit{
				Version: conf.Version{
					Build: "v1.0.0",
				},
			},
			wantErr: false,
			want:    "Version: v1.0.0",
		},
		{
			name: "only version shortcut",
			args: []string{"conf.test", "-v"},
			config: ConfImplicit{
				Version: conf.Version{
					Build: "v1.0.0",
				},
			},
			wantErr: false,
			want:    "Version: v1.0.0",
		},
		{
			name: "version and description",
			args: []string{"conf.test", "-version"},
			config: ConfImplicit{
				Version: conf.Version{
					Build: "v1.0.0",
					Desc:  "Service Description",
				},
			},
			wantErr: false,
			want:    "Version: v1.0.0\nService Description",
		},
		{
			name: "version and description shortcut",
			args: []string{"conf.test", "-v"},
			config: ConfImplicit{
				Version: conf.Version{
					Build: "v1.0.0",
					Desc:  "Service Description",
				},
			},
			wantErr: false,
			want:    "Version: v1.0.0\nService Description",
		},
		{
			name: "only description shortcut",
			args: []string{"conf.test", "-v"},
			config: ConfImplicit{
				Version: conf.Version{
					Desc: "Service Description",
				},
			},
			wantErr: false,
			want:    "Service Description",
		},
		{
			name:    "no version",
			args:    []string{"conf.test", "-v"},
			config:  ConfImplicit{},
			want:    "",
			wantErr: false,
		},
	}

	t.Log("Given the need validate version output.")
	{
		for i, tt := range tests {
			t.Logf("\tTest: %d\tWhen using an emplicit struct with.", i)
			{
				f := func(t *testing.T) {
					os.Args = tt.args
					if help, err := conf.Parse("APP", &tt.config); err != nil {
						if err == conf.ErrHelpWanted {
							if diff := cmp.Diff(tt.want, help); diff != "" {
								t.Errorf("\t%s\tShould match the output byte for byte. See diff:", failed)
								t.Log(diff)
							}
							t.Logf("\t%s\tShould match byte for byte the output.", success)
						}
					}
				}

				t.Run(tt.name, f)
			}
		}
	}
}

func TestParseBoolFlag(t *testing.T) {
	type Inner struct {
		Bool2 bool
	}

	type config struct {
		Bool bool `conf:"short:b"`
		Args conf.Args
	}

	type config2 struct {
		Bool  bool `conf:"short:b"`
		Args  conf.Args
		Inner Inner
	}

	type config3 struct {
		Bool bool `conf:"short:b"`
		Args conf.Args
		Inner
	}

	type args struct {
		prefix  string
		cfg     interface{}
		parsers []conf.Parsers
	}

	tests := []struct {
		name    string
		osags   []string
		args    args
		wantErr bool
		expect  interface{}
	}{
		{
			name:  "long w/o equals",
			osags: []string{"cmd", "--bool", "extra"},
			args: args{
				cfg: &config{},
			},
			wantErr: false,
			expect: &config{
				Bool: true,
				Args: conf.Args{"extra"},
			},
		},
		{
			name:  "short w/o equals",
			osags: []string{"cmd", "-b", "extra"},
			args: args{
				cfg: &config{},
			},
			wantErr: false,
			expect: &config{
				Bool: true,
				Args: conf.Args{"extra"},
			},
		},
		{
			name:  "long w/equals true",
			osags: []string{"cmd", "--bool=true", "extra"},
			args: args{
				cfg: &config{},
			},
			wantErr: false,
			expect: &config{
				Bool: true,
				Args: conf.Args{"extra"},
			},
		},
		{
			name:  "short w/equals true",
			osags: []string{"cmd", "-b=true", "extra"},
			args: args{
				cfg: &config{},
			},
			wantErr: false,
			expect: &config{
				Bool: true,
				Args: conf.Args{"extra"},
			},
		},
		{
			name:  "long w/equals false",
			osags: []string{"cmd", "--bool=false", "extra"},
			args: args{
				cfg: &config{},
			},
			wantErr: false,
			expect: &config{
				Bool: false,
				Args: conf.Args{"extra"},
			},
		},
		{
			name:  "short w/equals false",
			osags: []string{"cmd", "-b=false", "extra"},
			args: args{
				cfg: &config{},
			},
			wantErr: false,
			expect: &config{
				Bool: false,
				Args: conf.Args{"extra"},
			},
		},
		{
			name:  "just long flag",
			osags: []string{"cmd", "--bool"},
			args: args{
				cfg: &config{},
			},
			wantErr: false,
			expect: &config{
				Bool: true,
				Args: conf.Args{},
			},
		},
		{
			name:  "just short flag",
			osags: []string{"cmd", "-b"},
			args: args{
				cfg: &config{},
			},
			wantErr: false,
			expect: &config{
				Bool: true,
				Args: conf.Args{},
			},
		},
		{
			name:  "just extra",
			osags: []string{"cmd", "extra"},
			args: args{
				cfg: &config{},
			},
			wantErr: false,
			expect: &config{
				Bool: false,
				Args: conf.Args{"extra"},
			},
		},
		{
			name:  "inner",
			osags: []string{"cmd", "extra"},
			args: args{
				cfg: &config2{},
			},
			wantErr: false,
			expect: &config2{
				Bool: false,
				Args: conf.Args{"extra"},
			},
		},
		{
			name:  "embedded",
			osags: []string{"cmd", "extra"},
			args: args{
				cfg: &config3{},
			},
			wantErr: false,
			expect: &config3{
				Bool: false,
				Args: conf.Args{"extra"},
			},
		},
	}
	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			tmpArgs := os.Args
			t.Cleanup(func() {
				os.Args = tmpArgs
			})

			os.Args = tt.osags

			prefix := tt.args.prefix
			if prefix == "" {
				prefix = "CONF_TEST_PARSE_BOOL_FLAG"
			}

			_, err := conf.Parse(prefix, tt.args.cfg, tt.args.parsers...)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parse bool flag with args: error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.expect, tt.args.cfg); diff != "" {
				t.Errorf("parse bool flag with args: cfg mismatch (-expect +got):\n%s", diff)
			}
		})
	}
}

// =============================================================================

var yamlData1 = `
a: Easy!
b:
  c: 2
  d: [3, 4]
c: 2000-01-01T10:17:00Z
`

var yamlData2 = `
a: Easy!
b:
  c: 2
  d: [3, 4]
c: 2000-01-01T10:17:00Z
d: 2000-01-01T10:17:00Z
`

type internal struct {
	RenamedC int   `yaml:"c"`
	D        []int `yaml:",flow"`
}

type yamlConfig1 struct {
	A string
	B internal
	E string    `conf:"default:postgres"`
	C time.Time `conf:"default:2023-06-16T10:17:00Z"`
}

type yamlConfig2 struct {
	A string
	B internal
	E string    `conf:"default:postgres"`
	C time.Time `conf:"default:2023-06-16T10:17:00Z"`
	D time.Time `conf:"required"`
}

func TestYAML(t *testing.T) {
	ts, _ := time.Parse(time.RFC3339, "2000-01-01T10:17:00Z")

	tests := []struct {
		name string
		yaml []byte
		envs map[string]string
		args []string
		got  any
		exp  any
	}{
		{
			"default",
			[]byte(yamlData1),
			nil,
			nil,
			&yamlConfig1{},
			&yamlConfig1{A: "Easy!", B: internal{RenamedC: 2, D: []int{3, 4}}, E: "postgres", C: ts},
		},
		{
			"env",
			[]byte(yamlData2),
			map[string]string{"TEST_A": "EnvEasy!", "TEST_D": "2000-01-01T10:17:00Z"},
			nil,
			&yamlConfig2{},
			&yamlConfig2{A: "EnvEasy!", B: internal{RenamedC: 2, D: []int{3, 4}}, E: "postgres", C: ts, D: ts},
		},
		{
			"flag",
			[]byte(yamlData2),
			nil,
			[]string{"conf.test", "--a", "FlagEasy!", "--d", "2000-01-01T10:17:00Z"},
			&yamlConfig2{},
			&yamlConfig2{A: "FlagEasy!", B: internal{RenamedC: 2, D: []int{3, 4}}, E: "postgres", C: ts, D: ts},
		},
	}

	t.Log("Given the need to parse basic yaml configuration.")
	{
		for i, tt := range tests {
			t.Logf("\tTest: %d\tWhen checking with arguments %v", i, tt.args)
			{
				os.Clearenv()
				for k, v := range tt.envs {
					os.Setenv(k, v)
				}

				f := func(t *testing.T) {
					os.Args = tt.args

					if _, err := conf.Parse("TEST", tt.got, yaml.WithData(tt.yaml)); err != nil {
						t.Fatalf("\t%s\tShould be able to Parse arguments : %s.", failed, err)
					}
					t.Logf("\t%s\tShould be able to Parse arguments.", success)

					if diff := cmp.Diff(tt.exp, tt.got); diff != "" {
						t.Fatalf("\t%s\tShould have properly initialized struct value\n%s", failed, diff)
					}
					t.Logf("\t%s\tShould have properly initialized struct value.", success)
				}

				t.Run(tt.name, f)
			}
		}
	}
}
