package common

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"sync"
	"time"

	"boscoin.io/sebak/lib/errors"
	"github.com/inconshreveable/log15"
	"github.com/mattn/go-isatty"
)

type JSONLog interface {
	JSONLog()
}

func LogFormatter(f string) log15.Format {
	var logFormatter log15.Format
	switch f {
	case "terminal":
		if InTest || isatty.IsTerminal(os.Stdout.Fd()) {
			logFormatter = log15.TerminalFormat()
		} else {
			logFormatter = log15.LogfmtFormat()
		}
	case "", "json":
		logFormatter = JsonFormatEx(false, true)
	}

	return logFormatter
}

func LogHandler(format log15.Format, f string) (log15.Handler, error) {
	if len(f) < 1 {
		return log15.StreamHandler(os.Stdout, format), nil
	}

	return log15.FileHandler(f, format)
}

func SetLogger(logger log15.Logger, level log15.Lvl, handler log15.Handler) {
	logger.SetHandler(log15.LvlFilterHandler(level, handler))
}

// `formatJSONValue` and `JsonFormatEx` was derived from
// https://github.com/inconshreveable/log15/blob/199fca55789248e0520a3bd33e9045799738e793/format.go#L131
// .
const errorKey = "LOG15_ERROR"

func formatLogJSONValue(value interface{}) (result interface{}) {
	defer func() {
		if err := recover(); err != nil {
			if v := reflect.ValueOf(value); v.Kind() == reflect.Ptr && v.IsNil() {
				result = "nil"
			} else {
				panic(err)
			}
		}
	}()

	switch v := value.(type) {
	case JSONLog:
		return v
	case json.Marshaler:
		return v
	case *errors.Error:
		return v
	case Time:
		return v.String()
	case time.Time:
		return v.Format(TIMEFORMAT_ISO8601)
	case error:
		return v.Error()
	case fmt.Stringer:
		return v.String()
	default:
		return v
	}
}

func JsonFormatEx(pretty, lineSeparated bool) log15.Format {
	jsonMarshal := func(v interface{}) ([]byte, error) {
		return encodeJSON(v, false, false)
	}

	if pretty {
		jsonMarshal = func(v interface{}) ([]byte, error) {
			return json.MarshalIndent(v, "", "    ")
		}
	}

	return log15.FormatFunc(func(r *log15.Record) []byte {
		props := make(map[string]interface{})

		props[r.KeyNames.Time] = r.Time
		props[r.KeyNames.Lvl] = r.Lvl.String()
		props[r.KeyNames.Msg] = r.Msg

		for i := 0; i < len(r.Ctx); i += 2 {
			k, ok := r.Ctx[i].(string)
			if !ok {
				props[errorKey] = fmt.Sprintf("%+v is not a string key", r.Ctx[i])
			}
			props[k] = formatLogJSONValue(r.Ctx[i+1])
		}

		b, err := jsonMarshal(props)
		if err != nil {
			b, _ = jsonMarshal(map[string]string{
				errorKey: err.Error(),
			})
			return b
		}

		if lineSeparated {
			b = append(b, '\n')
		}

		return b
	})
}

func TerminalLogString(s string) string {
	return strings.TrimSpace(strings.Replace(s, "\"", "'", -1))
}

type Logger struct {
	sync.RWMutex
	logCtx log15.Ctx
	log    log15.Logger
}

func NewLogger(log log15.Logger, args ...interface{}) *Logger {
	l := &Logger{log: log}
	l.SetLogContext(args...)

	return l
}

func (l *Logger) LogContext() []interface{} {
	l.RLock()
	defer l.RUnlock()

	var args []interface{}
	for k, v := range l.logCtx {
		args = append(args, k, v)
	}

	return args
}

func (l *Logger) SetLogContext(args ...interface{}) {
	if len(args)%2 != 0 {
		panic(fmt.Errorf("invalid number of args: %v", len(args)))
	}

	l.Lock()
	defer l.Unlock()

	if l.logCtx == nil {
		l.logCtx = log15.Ctx{}
	}

	for i := 0; i < len(args); i += 2 {
		k, ok := args[i].(string)
		if !ok {
			panic(fmt.Errorf("key should be string: %T found", args[i]))
		}
		l.logCtx[k] = args[i+1]
	}

	l.log = l.log.New(l.logCtx)
}

func (l *Logger) SetLogger(log log15.Logger) {
	l.Lock()
	defer l.Unlock()

	l.log = log
}

func (l *Logger) Log() log15.Logger {
	l.RLock()
	defer l.RUnlock()

	return l.log
}
