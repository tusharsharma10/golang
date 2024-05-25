package logger

import (
	"context"
	"encoding/json"
	"os"
	"sync"

	"restapi/util"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	TransactionIDKey = "TRANSACTION_ID"
)

type Z = map[string]interface{}

var (
	once      sync.Once
	singleton *zap.SugaredLogger
)

// Init initializes a thread-safe singleton logger
// This would be called from a main method when the application starts up
// This function would ideally, take zap configuration, but is left out
// in favor of simplicity using the example logger.
func Init(
	name string,
	logLevel string,
) {
	// once ensures the singleton is initialized only once
	once.Do(func() {
		if name == "" {
			name = "goofy"
		}

		// by default, this sets the minimum logging level to info
		cfg := zap.NewProductionConfig()
		cfg.Level.SetLevel(parseLogLevel(logLevel))

		logDir := os.Getenv("LOG_DIR")

		if logDir == "" {
			logDir = "logs"
		}

		util.MakeDir(logDir, 0755)

		outputFile := logDir + "/" + name + ".log"

		// make the logDir if it does not exist

		cfg.OutputPaths = []string{outputFile}

		cfg.EncoderConfig.TimeKey = "logTime"
		cfg.EncoderConfig.EncodeTime = zapcore.RFC3339TimeEncoder

		cfg.EncoderConfig.MessageKey = "message"

		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

		builtLogger, _ := cfg.Build(zap.AddCallerSkip(2), zap.AddStacktrace(zapcore.DebugLevel))

		singleton = builtLogger.Sugar()
	})
}

func Info(ctx context.Context, msg string, data Z) {
	log(ctx, zapcore.InfoLevel, msg, data)
}

func Debug(ctx context.Context, msg string, data Z) {
	log(ctx, zapcore.DebugLevel, msg, data)
}

func Error(ctx context.Context, msg string, data Z) {
	log(ctx, zapcore.ErrorLevel, msg, data)
}

func log(ctx context.Context, level zapcore.Level, message string, data Z) {
	if singleton == nil {
		Init("other", os.Getenv("LOG_LEVEL"))
	}

	if ctx == nil {
		ctx = context.Background()
	}

	byteData, err := json.Marshal(data)
	if err != nil {
		Error(ctx, "error in logging", Z{
			"error": err.Error(),
		})
	}

	modifiedArgs := ingestData(ctx, string(byteData))

	switch level {
	case zapcore.ErrorLevel:
		{
			singleton.Errorw(message, modifiedArgs...)
			break
		}
	case zapcore.WarnLevel:
		{
			singleton.Warnw(message, modifiedArgs...)
			break
		}
	case zapcore.InfoLevel:
		{
			singleton.Infow(message, modifiedArgs...)
			break
		}
	case zapcore.DebugLevel:
		{
			singleton.Debugw(message, modifiedArgs...)
			break
		}
	}
}

func ingestData(ctx context.Context, inputdata string) []interface{} {
	data := Z{
		"raw_data": inputdata,
	}

	if data[TransactionIDKey] == nil || data[TransactionIDKey] == "" {
		data[TransactionIDKey] = ctx.Value(TransactionIDKey)
	}

	if data[TransactionIDKey] == nil || data[TransactionIDKey] == "" {
		data[TransactionIDKey] = newTransactionID()
	}

	// organize Data in key, value, key, value... in an array of interface
	argsLen := len(data) * 2 // key + value
	args := make([]interface{}, argsLen)

	i := 0
	for k, v := range data {
		args[i] = k
		args[i+1] = v
		i += 2
	}

	return args
}

// NewTransactionID generates a random string.
func newTransactionID() string {
	return uuid.New().String()
}

func parseLogLevel(logLevel string) zapcore.Level {
	switch logLevel {
	case "debug":
		{
			return zapcore.DebugLevel
		}
	case "info":
		{
			return zapcore.InfoLevel
		}
	case "error":
		{
			return zapcore.ErrorLevel
		}
	default:
		{
			return zapcore.InfoLevel
		}
	}
}
