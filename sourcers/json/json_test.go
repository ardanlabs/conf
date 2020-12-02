package json_test

import (
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ardanlabs/conf"
	"github.com/ardanlabs/conf/sourcers/json"
	"github.com/google/go-cmp/cmp"
)

const (
	success = "\u2713"
	failed  = "\u2717"
)

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
	AnInt   int    `conf:"default:9"`
	AString string `conf:"default:B,short:s"`
	Bool    bool
	Skip    string `conf:"-"`
	IP      ip
	Embed
}

func TestJSONParse(t *testing.T) {
	tests := []struct {
		name string
		envs map[string]string
		args []string
		json io.Reader
		want config
	}{
		{
			"default",
			nil,
			nil,
			strings.NewReader(`{}`),
			config{9, "B", false, "", ip{"localhost", "127.0.0.0", []string{"127.0.0.1:200", "127.0.0.1:829"}}, Embed{"bill", time.Second}},
		},
		{
			"json",
			nil,
			nil,
			strings.NewReader(`{"a_string": "s", "d": "1m", "ignored": true, "bool": true}`),
			config{9, "s", true, "", ip{"localhost", "127.0.0.0", []string{"127.0.0.1:200", "127.0.0.1:829"}}, Embed{"bill", time.Minute}},
		},
		{
			"env",
			map[string]string{"TEST_AN_INT": "1", "TEST_A_STRING": "s", "TEST_BOOL": "TRUE", "TEST_SKIP": "SKIP", "TEST_IP_NAME_VAR": "local", "TEST_NAME": "andy", "TEST_DURATION": "1m"},
			nil,
			strings.NewReader(`{}`),
			config{1, "s", true, "", ip{"local", "127.0.0.0", []string{"127.0.0.1:200", "127.0.0.1:829"}}, Embed{"andy", time.Minute}},
		},
		{
			"flag",
			nil,
			[]string{"--an-int", "1", "-s", "s", "--bool", "--skip", "skip", "--ip-name", "local", "--name", "andy", "--e-dur", "1m"},
			strings.NewReader(`{}`),
			config{1, "s", true, "", ip{"local", "127.0.0.0", []string{"127.0.0.1:200", "127.0.0.1:829"}}, Embed{"andy", time.Minute}},
		},
		{
			"multi",
			map[string]string{"TEST_A_STRING": "s", "TEST_BOOL": "TRUE", "TEST_IP_NAME_VAR": "local", "TEST_NAME": "andy", "TEST_DURATION": "1m"},
			[]string{"--an-int", "2", "--bool", "--skip", "skip", "--name", "jack", "-d", "1ms"},
			strings.NewReader(`{}`),
			config{2, "s", true, "", ip{"local", "127.0.0.0", []string{"127.0.0.1:200", "127.0.0.1:829"}}, Embed{"jack", time.Millisecond}},
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
				jsonSourcer, _ := json.NewSource(tt.json)

				f := func(t *testing.T) {
					var cfg config
					if err := conf.Parse(tt.args, "TEST", &cfg, jsonSourcer); err != nil {
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
