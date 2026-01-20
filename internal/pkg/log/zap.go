/*
Copyright 2025 The Application Catalog Manager contributors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package log

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	ctrlruntimelzap "sigs.k8s.io/controller-runtime/pkg/log/zap"
)

type Format string

func (f *Format) Type() string {
	return "string"
}

func (f *Format) String() string {
	return string(*f)
}

func (f *Format) Set(s string) error {
	switch strings.ToLower(s) {
	case "json":
		*f = FormatJSON
		return nil
	case "console":
		*f = FormatConsole
		return nil
	default:
		return fmt.Errorf("invalid format '%s'", s)
	}
}

const (
	FormatJSON    Format = "JSON"
	FormatConsole Format = "Console"
)

var AvailableFormats = []Format{FormatJSON, FormatConsole}

// Options exports an options struct to be used by the command-line as flag.
type Options struct {
	// Debug configures the debug log level.
	Debug bool
	// Format corresponds to the log format (JSON or plain text).
	Format Format
}

func NewDefaultOptions() Options {
	return Options{
		Debug:  false,
		Format: FormatJSON,
	}
}

func (o *Options) AddFlags(fs *flag.FlagSet) {
	fs.BoolVar(&o.Debug, "log-debug", o.Debug, "Enables more verbose logging")
	fs.Var(&o.Format, "log-format", "Log format, one of JSON or Console")
}

func (o *Options) AddPFlags(fs *pflag.FlagSet) {
	fs.BoolVar(&o.Debug, "log-debug", o.Debug, "Enables more verbose logging")
	fs.Var(&o.Format, "log-format", "Log format, one of JSON or Console")
}

func (o *Options) Validate() error {
	for i := range AvailableFormats {
		if o.Format == AvailableFormats[i] {
			return nil
		}
	}

	return fmt.Errorf("invalid log-format specified %q; available: %+v", o.Format, AvailableFormats)
}

func NewFromOptions(o Options) *zap.Logger {
	return New(o.Debug, o.Format)
}

func New(debug bool, format Format) *zap.Logger {
	sink := zapcore.AddSync(os.Stderr)

	lvl := zap.NewAtomicLevelAt(zap.InfoLevel)
	if debug {
		lvl = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	encCfg := zap.NewProductionEncoderConfig()

	encCfg.TimeKey = "time"
	encCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encCfg.EncodeDuration = zapcore.StringDurationEncoder

	var enc zapcore.Encoder
	if format == FormatJSON {
		enc = zapcore.NewJSONEncoder(encCfg)
	} else {
		enc = zapcore.NewConsoleEncoder(encCfg)
	}

	opts := []zap.Option{
		zap.AddCaller(),
		zap.ErrorOutput(sink),
	}

	coreLog := zapcore.NewCore(&ctrlruntimelzap.KubeAwareEncoder{Encoder: enc}, sink, lvl)
	return zap.New(coreLog, opts...)
}

// NewDefault creates new default logger.
func NewDefault() *zap.Logger {
	return New(false, FormatJSON)
}
