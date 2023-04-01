package configflag

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/benchttp/engine/configio"
)

// Bind reads arguments provided to flagset as config fields
// and binds their value to the appropriate fields of dst.
// The provided *flag.Flagset must not have been parsed yet, otherwise
// bindings its values would fail.
func Bind(flagset *flag.FlagSet, dst *configio.Builder) {
	for field, bind := range bindings {
		flagset.Func(field, flagsUsage[field], bind(dst))
	}
}

type setter = func(string) error

var bindings = map[string]func(*configio.Builder) setter{
	flagMethod: func(b *configio.Builder) setter {
		return func(in string) error {
			b.SetRequestMethod(in)
			return nil
		}
	},
	flagURL: func(b *configio.Builder) setter {
		return func(in string) error {
			u, err := url.ParseRequestURI(in)
			if err != nil {
				return err
			}
			b.SetRequestURL(u)
			return nil
		}
	},
	flagHeader: func(b *configio.Builder) setter {
		return func(in string) error {
			keyval := strings.SplitN(in, ":", 2)
			if len(keyval) != 2 {
				return errors.New(`-header: expect format "<key>:<value>"`)
			}
			key, val := keyval[0], keyval[1]
			b.SetRequestHeaderFunc(func(h http.Header) http.Header {
				if h == nil {
					h = http.Header{}
				}
				h[key] = append(h[key], val)
				return h
			})
			return nil
		}
	},
	flagBody: func(b *configio.Builder) setter {
		return func(in string) error {
			errFormat := fmt.Errorf(`expect format "<type>:<content>", got %q`, in)
			if in == "" {
				return errFormat
			}
			split := strings.SplitN(in, ":", 2)
			if len(split) != 2 {
				return errFormat
			}
			btype, bcontent := split[0], split[1]
			if bcontent == "" {
				return errFormat
			}
			switch btype {
			case "raw":
				b.SetRequestBody(io.NopCloser(bytes.NewBufferString(bcontent)))
			// case "file":
			// 	// TODO
			default:
				return fmt.Errorf(`unsupported type: %s (only "raw" accepted)`, btype)
			}
			return nil
		}
	},
	flagRequests: func(b *configio.Builder) setter {
		return func(in string) error {
			n, err := strconv.Atoi(in)
			if err != nil {
				return err
			}
			b.SetRequests(n)
			return nil
		}
	},
	flagConcurrency: func(b *configio.Builder) setter {
		return func(in string) error {
			n, err := strconv.Atoi(in)
			if err != nil {
				return err
			}
			b.SetConcurrency(n)
			return nil
		}
	},
	flagInterval: func(b *configio.Builder) setter {
		return func(in string) error {
			d, err := time.ParseDuration(in)
			if err != nil {
				return err
			}
			b.SetInterval(d)
			return nil
		}
	},
	flagRequestTimeout: func(b *configio.Builder) setter {
		return func(in string) error {
			d, err := time.ParseDuration(in)
			if err != nil {
				return err
			}
			b.SetRequestTimeout(d)
			return nil
		}
	},
	flagGlobalTimeout: func(b *configio.Builder) setter {
		return func(in string) error {
			d, err := time.ParseDuration(in)
			if err != nil {
				return err
			}
			b.SetGlobalTimeout(d)
			return nil
		}
	},
}
