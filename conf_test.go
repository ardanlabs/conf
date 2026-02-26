package conf_test

import (
	"errors"
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
	Immutable string      `conf:"default:mydefaultvalue,immutable"`
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
			config{9, "B", false, "", ip{"localhost", "127.0.0.0", []string{"127.0.0.1:200", "127.0.0.1:829"}}, "http://user:password@0.0.0.0:4000", "password", "mydefaultvalue", CustomValue{something: "@hello@"}, Embed{"bill", time.Second}},
		},
		{
			"env",
			map[string]string{"TEST_AN_INT": "1", "TEST_A_STRING": "s", "TEST_BOOL": "TRUE", "TEST_SKIP": "SKIP", "TEST_IP_NAME_VAR": "local", "TEST_DEBUG_HOST": "http://bill:gopher@0.0.0.0:4000", "TEST_PASSWORD": "gopher", "TEST_NAME": "andy", "TEST_DURATION": "1m"},
			nil,
			config{1, "s", true, "", ip{"local", "127.0.0.0", []string{"127.0.0.1:200", "127.0.0.1:829"}}, "http://bill:gopher@0.0.0.0:4000", "gopher", "mydefaultvalue", CustomValue{something: "@hello@"}, Embed{"andy", time.Minute}},
		},
		{
			"flag",
			nil,
			[]string{"conf.test", "--an-int", "1", "-s", "s", "--bool", "--skip", "skip", "--ip-name", "local", "--debug-host", "http://bill:gopher@0.0.0.0:4000", "--password", "gopher", "--name", "andy", "--e-dur", "1m"},
			config{1, "s", true, "", ip{"local", "127.0.0.0", []string{"127.0.0.1:200", "127.0.0.1:829"}}, "http://bill:gopher@0.0.0.0:4000", "gopher", "mydefaultvalue", CustomValue{something: "@hello@"}, Embed{"andy", time.Minute}},
		},
		{
			"multi",
			map[string]string{"TEST_A_STRING": "s", "TEST_BOOL": "TRUE", "TEST_IP_NAME_VAR": "local", "TEST_DEBUG_HOST": "http://bill:gopher@0.0.0.0:4000", "TEST_PASSWORD": "gopher", "TEST_NAME": "andy", "TEST_DURATION": "1m"},
			[]string{"conf.test", "--an-int", "2", "--bool", "--skip", "skip", "--name", "jack", "-d", "1ms"},
			config{2, "s", true, "", ip{"local", "127.0.0.0", []string{"127.0.0.1:200", "127.0.0.1:829"}}, "http://bill:gopher@0.0.0.0:4000", "gopher", "mydefaultvalue", CustomValue{something: "@hello@"}, Embed{"jack", time.Millisecond}},
		},
		{
			"immutable-env",
			map[string]string{"TEST_IMMUTABLE": "change"},
			nil,
			config{9, "B", false, "", ip{"localhost", "127.0.0.0", []string{"127.0.0.1:200", "127.0.0.1:829"}}, "http://user:password@0.0.0.0:4000", "password", "mydefaultvalue", CustomValue{something: "@hello@"}, Embed{"bill", time.Second}},
		},
		{
			"immutable-args",
			nil,
			[]string{"conf.test", "--immutable", "change"},
			config{9, "B", false, "", ip{"localhost", "127.0.0.0", []string{"127.0.0.1:200", "127.0.0.1:829"}}, "http://user:password@0.0.0.0:4000", "password", "mydefaultvalue", CustomValue{something: "@hello@"}, Embed{"bill", time.Second}},
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
			config{1, "s", true, "", ip{"local", "127.0.0.0", []string{"127.0.0.1:200", "127.0.0.1:829"}}, "http://bill:gopher@0.0.0.0:4000", "gopher", "mydefaultvalue", CustomValue{something: "@hello@"}, Embed{"andy", time.Minute}},
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

var withNamespace = `Usage: conf.test [options...] [arguments...]

OPTIONS
  -s, --a-string      <string>              (default: B)                                  
      --an-int        <int>                 (default: 9)                                  
      --bool          <bool>                                                              
      --custom        <value>               (default: hello)                              
      --debug-host    <string>              (default: http://xxxxxx:xxxxxx@0.0.0.0:4000)  
  -d, --e-dur         <duration>            (default: 1s)                                 
  -h, --help                                                                              display this help message
      --immutable     <string>              (immutable,default: mydefaultvalue)           
      --ip-endpoints  <string>,[string...]  (default: 127.0.0.1:200;127.0.0.1:829)        
      --ip-ip         <string>              (default: 127.0.0.0)                          
      --ip-name       <string>              (default: localhost)                          
      --name          <string>              (default: bill)                               
      --password      <string>              (default: xxxxxx)                             

ENVIRONMENT
  TEST_A_STRING      <string>              (default: B)                                  
  TEST_AN_INT        <int>                 (default: 9)                                  
  TEST_BOOL          <bool>                                                              
  TEST_CUSTOM        <value>               (default: hello)                              
  TEST_DEBUG_HOST    <string>              (default: http://xxxxxx:xxxxxx@0.0.0.0:4000)  
  TEST_DURATION      <duration>            (default: 1s)                                 
  TEST_IMMUTABLE     <string>              (immutable,default: mydefaultvalue)           
  TEST_IP_ENDPOINTS  <string>,[string...]  (default: 127.0.0.1:200;127.0.0.1:829)        
  TEST_IP_IP         <string>              (default: 127.0.0.0)                          
  TEST_IP_NAME_VAR   <string>              (default: localhost)                          
  TEST_NAME          <string>              (default: bill)                               
  TEST_PASSWORD      <string>              (default: xxxxxx)                             `

var emptyNamespace = `Usage: conf.test [options...] [arguments...]

OPTIONS
  -s, --a-string      <string>              (default: B)                                  
      --an-int        <int>                 (default: 9)                                  
      --bool          <bool>                                                              
      --custom        <value>               (default: hello)                              
      --debug-host    <string>              (default: http://xxxxxx:xxxxxx@0.0.0.0:4000)  
  -d, --e-dur         <duration>            (default: 1s)                                 
  -h, --help                                                                              display this help message
      --immutable     <string>              (immutable,default: mydefaultvalue)           
      --ip-endpoints  <string>,[string...]  (default: 127.0.0.1:200;127.0.0.1:829)        
      --ip-ip         <string>              (default: 127.0.0.0)                          
      --ip-name       <string>              (default: localhost)                          
      --name          <string>              (default: bill)                               
      --password      <string>              (default: xxxxxx)                             

ENVIRONMENT
  A_STRING      <string>              (default: B)                                  
  AN_INT        <int>                 (default: 9)                                  
  BOOL          <bool>                                                              
  CUSTOM        <value>               (default: hello)                              
  DEBUG_HOST    <string>              (default: http://xxxxxx:xxxxxx@0.0.0.0:4000)  
  DURATION      <duration>            (default: 1s)                                 
  IMMUTABLE     <string>              (immutable,default: mydefaultvalue)           
  IP_ENDPOINTS  <string>,[string...]  (default: 127.0.0.1:200;127.0.0.1:829)        
  IP_IP         <string>              (default: 127.0.0.0)                          
  IP_NAME_VAR   <string>              (default: localhost)                          
  NAME          <string>              (default: bill)                               
  PASSWORD      <string>              (default: xxxxxx)                             `

var withNamespaceOptions = `Usage: conf.test [options...] [arguments...]

OPTIONS
  -h, --help           display this help message
      --port  <int>    

ENVIRONMENT
  TEST_PORT  <int>`

var emptyNamespaceOptions = `Usage: conf.test [options...] [arguments...]

OPTIONS
  -h, --help           display this help message
      --port  <int>    

ENVIRONMENT
  PORT  <int>`

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
			envs: map[string]string{
				"TEST_AN_INT":      "1",
				"TEST_A_STRING":    "s",
				"TEST_BOOL":        "TRUE",
				"TEST_SKIP":        "SKIP",
				"TEST_IP_NAME_VAR": "local",
				"TEST_NAME":        "andy",
				"TEST_DURATION":    "1m",
			},
			want:    withNamespace,
			options: withNamespaceOptions,
		},
		{
			name:      "empty-namespace",
			namespace: "",
			envs: map[string]string{
				"AN_INT":      "1",
				"A_STRING":    "s",
				"BOOL":        "TRUE",
				"SKIP":        "SKIP",
				"IP_NAME_VAR": "local",
				"NAME":        "andy",
				"DURATION":    "1m",
			},
			want:    emptyNamespace,
			options: emptyNamespaceOptions,
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
						t.Log(err)
						return
					}

					got, err := conf.UsageInfo(tt.namespace, &cfg)
					if err != nil {
						t.Log(err)
						return
					}

					got = strings.TrimRight(got, "\n")
					gotS := strings.Split(got, "\n")
					wantS := strings.Split(tt.want, "\n")
					if diff := cmp.Diff(gotS, wantS); diff != "" {
						t.Errorf("\t%s\tShould match the output byte for byte. See diff:", failed)
						t.Log(diff)
						t.Log("GOT:\n", got)
						t.Log("EXP:\n", tt.want)
						return
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
						t.Log(err)
						return
					}

					got = strings.TrimRight(got, " \n")
					gotS := strings.Split(got, "\n")
					wantS := strings.Split(tt.options, "\n")
					if diff := cmp.Diff(gotS, wantS); diff != "" {
						t.Errorf("\t%s\tShould match the output byte for byte. See diff:", failed)
						t.Log(diff)
						t.Log("GOT:\n", got)
						t.Log("EXP:\n", tt.options)
						return
					}
					t.Logf("\t%s\tShould match byte for byte the output.", success)
				}
			}

			t.Run(tt.name, f)
		}
	}
}

var expectedStringOutput = `    --a-string=B
    --an-int=1
    --bool=true
    --custom=@hello@
    --debug-host=http://xxxxxx:xxxxxx@0.0.0.0:4000
    --e-dur=1m0s
    --immutable=mydefaultvalue
    --ip-endpoints=[127.0.0.1:200 127.0.0.1:829]
    --ip-ip=127.0.0.0
    --ip-name=localhost
    --name=andy
    --password=xxxxxx`

func TestExampleString(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		want      string
		envs      map[string]string
	}{
		{
			name:      "with-namespace",
			namespace: "TEST",
			envs: map[string]string{
				"TEST_AN_INT":   "1",
				"TEST_S":        "s",
				"TEST_BOOL":     "TRUE",
				"TEST_SKIP":     "SKIP",
				"TEST_IP_NAME":  "local",
				"TEST_NAME":     "andy",
				"TEST_DURATION": "1m",
			},
			want: expectedStringOutput,
		},
		{
			name:      "without-namespace",
			namespace: "",
			envs: map[string]string{
				"AN_INT":   "1",
				"S":        "s",
				"BOOL":     "TRUE",
				"SKIP":     "SKIP",
				"IP_NAME":  "local",
				"NAME":     "andy",
				"DURATION": "1m",
			},
			want: expectedStringOutput,
		},
	}

	t.Log("Given the need validate conf output.")
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
						t.Log(err)
						return
					}

					got, err := conf.String(&cfg)
					if err != nil {
						t.Log(err)
						return
					}

					got = strings.TrimRight(got, "\n")
					gotS := strings.Split(got, "\n")
					wantS := strings.Split(tt.want, "\n")
					if diff := cmp.Diff(gotS, wantS); diff != "" {
						t.Errorf("\t%s\tShould match the output byte for byte. See diff:", failed)
						t.Log(diff)
						t.Log("GOT:\n", got)
						t.Log("EXP:\n", tt.want)
						return
					}
					t.Logf("\t%s\tShould match byte for byte the output.", success)
				}
			}

			t.Run(tt.name, f)
		}
	}
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
						if err != conf.ErrHelpWanted {
							if diff := cmp.Diff(tt.want, help); diff != "" {
								t.Errorf("\t%s\tShould match the output byte for byte. See diff:", failed)
								t.Log(diff)
								return
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
								return
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
		cfg     any
		parsers []conf.Parsers
	}

	tests := []struct {
		name    string
		osags   []string
		args    args
		wantErr bool
		expect  any
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

type internal struct {
	RenamedC int   `yaml:"c"`
	D        []int `yaml:",flow"`
}

var yamlData1 = `
a: Easy!
b:
  c: 2
  d: [3, 4]
g: 2000-01-01T10:17:00Z
`

type yamlConfig1 struct {
	A string
	B internal
	E string    `conf:"default:postgres"`
	F time.Time `conf:"default:2023-06-16T10:17:00Z"`
	G time.Time `conf:"notzero"`
}

var yamlData2 = `
a: Easy!
b:
  c: 2
  d: [3, 4]
g: 2000-01-01T10:17:00Z
i: 2000-01-01T10:17:00Z
`

type yamlConfig2 struct {
	A string
	B internal
	E string    `conf:"default:postgres"`
	F time.Time `conf:"default:2023-06-16T10:17:00Z"`
	G time.Time `conf:"notzero"`
	I time.Time `conf:"required"`
}

var yamlData31 = `
a: Easy!
b:
  c: 2
  d: [3, 4]
i: 2000-01-01T10:17:00Z
`

var yamlData32 = `
a: Easy!
b:
  c: 2
  d: [3, 4]
g: 2000-01-01T10:17:00Z
`

type yamlConfig3 struct {
	A string
	B internal
	E string    `conf:"default:postgres"`
	F time.Time `conf:"default:2023-06-16T10:17:00Z"`
	G time.Time `conf:"notzero"`
	I time.Time `conf:"required"`
}

func TestYAML(t *testing.T) {
	dTS, _ := time.Parse(time.RFC3339, "2023-06-16T10:17:00Z")
	oTS, _ := time.Parse(time.RFC3339, "2000-01-01T10:17:00Z")

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
			&yamlConfig1{A: "Easy!", B: internal{RenamedC: 2, D: []int{3, 4}}, E: "postgres", F: dTS, G: oTS},
		},
		{
			"env",
			[]byte(yamlData2),
			map[string]string{"TEST_A": "EnvEasy!", "TEST_G": "2000-01-01T10:17:00Z", "TEST_I": "2000-01-01T10:17:00Z"},
			nil,
			&yamlConfig2{},
			&yamlConfig2{A: "EnvEasy!", B: internal{RenamedC: 2, D: []int{3, 4}}, E: "postgres", F: dTS, G: oTS, I: oTS},
		},
		{
			"flag",
			[]byte(yamlData2),
			nil,
			[]string{"conf.test", "--a", "FlagEasy!", "--g", "2000-01-01T10:17:00Z", "--i", "2000-01-01T10:17:00Z"},
			&yamlConfig2{},
			&yamlConfig2{A: "FlagEasy!", B: internal{RenamedC: 2, D: []int{3, 4}}, E: "postgres", F: dTS, G: oTS, I: oTS},
		},
		{
			"notzero",
			[]byte(yamlData31),
			nil,
			nil,
			&yamlConfig3{},
			errors.New("parsing config: field G is set to zero value"),
		},
		{
			"required",
			[]byte(yamlData32),
			nil,
			nil,
			&yamlConfig3{},
			errors.New("parsing config: required field I is missing value"),
		},
	}

	t.Log("Given the need to parse basic yaml configuration.")
	{
		for i, tt := range tests {
			t.Logf("\tTest: %d-%s\tWhen checking with arguments %v", i, tt.name, tt.args)
			{
				os.Clearenv()
				for k, v := range tt.envs {
					os.Setenv(k, v)
				}

				f := func(t *testing.T) {
					os.Args = tt.args

					if _, err := conf.Parse("TEST", tt.got, yaml.WithData(tt.yaml)); err != nil {
						errExp, ok := tt.exp.(error)
						if ok {
							if err.Error() == errExp.Error() {
								t.Logf("\t%s\tShould be able to get the correct error.", success)
								return
							}

							t.Fatalf("\t%s\tShould get the correct error : %s.", failed, err)
						}

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


func TestMapTraversal(t *testing.T) {
	yamlLabels := []byte(`
labels:
  env: staging
  region: us-east
`)

	tests := []struct {
		name    string
		envs    map[string]string
		args    []string
		initial map[string]string
		parsers []conf.Parsers
		want    map[string]string
	}{
		{
			name:    "env-override",
			envs:    map[string]string{"TEST_LABELS_ENV": "production"},
			initial: map[string]string{"env": "staging", "region": "us-east"},
			want:    map[string]string{"env": "production", "region": "us-east"},
		},
		{
			name:    "flag-override",
			args:    []string{"conf.test", "--labels-env", "production"},
			initial: map[string]string{"env": "staging", "region": "us-east"},
			want:    map[string]string{"env": "production", "region": "us-east"},
		},
		{
			name: "nil-map-whole-map-format",
			envs: map[string]string{"TEST_LABELS": "env:production;region:us-east"},
			want: map[string]string{"env": "production", "region": "us-east"},
		},
		{
			name:    "empty-map-falls-through-to-whole-map-format",
			envs:    map[string]string{"TEST_LABELS": "env:production"},
			initial: map[string]string{},
			want:    map[string]string{"env": "production"},
		},
		{
			name:    "flag-overrides-env-same-key",
			envs:    map[string]string{"TEST_LABELS_ENV": "from-env"},
			args:    []string{"conf.test", "--labels-env", "from-flag"},
			initial: map[string]string{"env": "staging"},
			want:    map[string]string{"env": "from-flag"},
		},
		{
			name:    "yaml-then-env-override",
			envs:    map[string]string{"TEST_LABELS_ENV": "production"},
			parsers: []conf.Parsers{yaml.WithData(yamlLabels)},
			want:    map[string]string{"env": "production", "region": "us-east"},
		},
		{
			name:    "yaml-then-flag-override",
			args:    []string{"conf.test", "--labels-env", "production"},
			parsers: []conf.Parsers{yaml.WithData(yamlLabels)},
			want:    map[string]string{"env": "production", "region": "us-east"},
		},
		{
			name:    "hyphen-key-env-override",
			envs:    map[string]string{"TEST_LABELS_ONE_TWO": "val"},
			initial: map[string]string{"one-two": "old"},
			want:    map[string]string{"one-two": "val"},
		},
		{
			name:    "underscore-key-env-override",
			envs:    map[string]string{"TEST_LABELS_ONE_TWO": "val"},
			initial: map[string]string{"one_two": "old"},
			want:    map[string]string{"one_two": "val"},
		},
		{
			name:    "hyphen-key-flag-override",
			args:    []string{"conf.test", "--labels-one-two", "val"},
			initial: map[string]string{"one-two": "old"},
			want:    map[string]string{"one-two": "val"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.envs {
				os.Setenv(k, v)
			}
			if tt.args != nil {
				os.Args = tt.args
			} else {
				os.Args = []string{"conf.test"}
			}

			var cfg struct {
				Labels map[string]string
			}
			cfg.Labels = tt.initial

			if _, err := conf.Parse("TEST", &cfg, tt.parsers...); err != nil {
				t.Fatalf("\t%s\tShould not error: %s", failed, err)
			}
			if diff := cmp.Diff(tt.want, cfg.Labels); diff != "" {
				t.Fatalf("\t%s\tMap mismatch (-want +got):\n%s", failed, diff)
			}
		})
	}
}

func TestMapTraversalEdgeCases(t *testing.T) {
	t.Run("map-in-nested-struct", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("TEST_INFRA_LABELS_ENV", "production")
		os.Args = []string{"conf.test"}

		var cfg struct {
			Infra struct {
				Labels map[string]string
			}
		}
		cfg.Infra.Labels = map[string]string{"env": "staging", "region": "us-east"}

		if _, err := conf.Parse("TEST", &cfg); err != nil {
			t.Fatalf("\t%s\tShould not error: %s", failed, err)
		}
		if diff := cmp.Diff(
			map[string]string{"env": "production", "region": "us-east"},
			cfg.Infra.Labels,
		); diff != "" {
			t.Fatalf("\t%s\tMap mismatch (-want +got):\n%s", failed, diff)
		}
	})

	t.Run("strict-flags-map-entry-consumed", func(t *testing.T) {
		os.Clearenv()
		os.Args = []string{"conf.test", "--labels-env", "production"}

		var cfg struct {
			Labels map[string]string
		}
		cfg.Labels = map[string]string{"env": "staging"}

		if _, err := conf.ParseWithOptions("TEST", &cfg, conf.WithStrictFlags()); err != nil {
			t.Fatalf("\t%s\tShould not error with strict flags: %s", failed, err)
		}
		if cfg.Labels["env"] != "production" {
			t.Errorf("\t%s\tExpected env=production, got %q", failed, cfg.Labels["env"])
		}
	})

	t.Run("new-key-not-added", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("TEST_LABELS_NEWKEY", "val")
		os.Args = []string{"conf.test"}

		var cfg struct {
			Labels map[string]string
		}
		cfg.Labels = map[string]string{"env": "staging"}

		if _, err := conf.Parse("TEST", &cfg); err != nil {
			t.Fatalf("\t%s\tShould not error: %s", failed, err)
		}
		if _, exists := cfg.Labels["newkey"]; exists {
			t.Errorf("\t%s\tKey absent at parse time should not be added", failed)
		}
	})

	t.Run("string-output-reflects-overrides", func(t *testing.T) {
		os.Clearenv()
		os.Setenv("TEST_LABELS_ENV", "production")
		os.Args = []string{"conf.test"}

		var cfg struct {
			Labels map[string]string
		}
		cfg.Labels = map[string]string{"env": "staging"}

		if _, err := conf.Parse("TEST", &cfg); err != nil {
			t.Fatalf("\t%s\tShould not error: %s", failed, err)
		}
		out, err := conf.String(&cfg)
		if err != nil {
			t.Fatalf("\t%s\tString() should not error: %s", failed, err)
		}
		if !strings.Contains(out, "production") {
			t.Errorf("\t%s\tString() should contain overridden value, got:\n%s", failed, out)
		}
		if strings.Contains(out, "staging") {
			t.Errorf("\t%s\tString() should not contain old value, got:\n%s", failed, out)
		}
	})

	t.Run("int-map-flag-override", func(t *testing.T) {
		os.Clearenv()
		os.Args = []string{"conf.test", "--ports-http", "9090"}

		var cfg struct {
			Ports map[string]int
		}
		cfg.Ports = map[string]int{"http": 8080, "https": 8443}

		if _, err := conf.Parse("TEST", &cfg); err != nil {
			t.Fatalf("\t%s\tShould not error: %s", failed, err)
		}
		if diff := cmp.Diff(map[string]int{"http": 9090, "https": 8443}, cfg.Ports); diff != "" {
			t.Fatalf("\t%s\tMap mismatch (-want +got):\n%s", failed, diff)
		}
	})

	boolTests := []struct {
		name string
		envs map[string]string
		args []string
		want map[string]bool
	}{
		{
			name: "bool-map-flag-without-value",
			args: []string{"conf.test", "--flags-debug"},
			want: map[string]bool{"debug": true, "verbose": false},
		},
		{
			name: "bool-map-env-override",
			envs: map[string]string{"TEST_FLAGS_DEBUG": "true"},
			want: map[string]bool{"debug": true, "verbose": false},
		},
	}
	for _, tt := range boolTests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.envs {
				os.Setenv(k, v)
			}
			if tt.args != nil {
				os.Args = tt.args
			} else {
				os.Args = []string{"conf.test"}
			}

			var cfg struct {
				Flags map[string]bool
			}
			cfg.Flags = map[string]bool{"debug": false, "verbose": false}

			if _, err := conf.Parse("TEST", &cfg); err != nil {
				t.Fatalf("\t%s\tShould not error: %s", failed, err)
			}
			if diff := cmp.Diff(tt.want, cfg.Flags); diff != "" {
				t.Fatalf("\t%s\tMap mismatch (-want +got):\n%s", failed, diff)
			}
		})
	}
}
