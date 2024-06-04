# Usage

## Importing the Package 📦

First, import the package in your Go file:

```go
import "github.com/catalystgo/xro-log/logger"
```

## Setting Up the Logger 🛠️

The logger can be initialized with a specific log level and options. By default, the log level is set to `ErrorLevel`.

```go
logger.SetLogger(logger.New(zapcore.DebugLevel))
```

## Logging Messages 📝

You can log messages at different levels using the following functions:

- `Debug(ctx context.Context, args ...interface{})`
- `Debugf(ctx context.Context, format string, args ...interface{})`
- `DebugKV(ctx context.Context, message string, kvs ...interface{})`
- `Info(ctx context.Context, args ...interface{})`
- `Infof(ctx context.Context, format string, args ...interface{})`
- `InfoKV(ctx context.Context, message string, kvs ...interface{})`
- `Warn(ctx context.Context, args ...interface{})`
- `Warnf(ctx context.Context, format string, args ...interface{})`
- `WarnKV(ctx context.Context, message string, kvs ...interface{})`
- `Error(ctx context.Context, args ...interface{})`
- `Errorf(ctx context.Context, format string, args ...interface{})`
- `ErrorKV(ctx context.Context, message string, kvs ...interface{})`
- `Fatal(ctx context.Context, args ...interface{})`
- `Fatalf(ctx context.Context, format string, args ...interface{})`
- `FatalKV(ctx context.Context, message string, kvs ...interface{})`
- `Panic(ctx context.Context, args ...interface{})`
- `Panicf(ctx context.Context, format string, args ...interface{})`
- `PanicKV(ctx context.Context, message string, kvs ...interface{})`

## Examples 🚀

Here are some examples of how to use the logging functions:

```go
ctx := context.Background()

logger.Debug(ctx, "This is a debug message")
logger.Debugf(ctx, "This is a debug message with a variable: %d", 42)
logger.DebugKV(ctx, "This is a debug message with key-value pairs", "key1", "value1", "key2", "value2")

logger.Info(ctx, "This is an info message")
logger.Infof(ctx, "This is an info message with a variable: %d", 42)
logger.InfoKV(ctx, "This is an info message with key-value pairs", "key1", "value1", "key2", "value2")

logger.Warn(ctx, "This is a warning message")
logger.Warnf(ctx, "This is a warning message with a variable: %d", 42)
logger.WarnKV(ctx, "This is a warning message with key-value pairs", "key1", "value1", "key2", "value2")

logger.Error(ctx, "This is an error message")
logger.Errorf(ctx, "This is an error message with a variable: %d", 42)
logger.ErrorKV(ctx, "This is an error message with key-value pairs", "key1", "value1", "key2", "value2")

logger.Fatal(ctx, "This is a fatal message")
logger.Fatalf(ctx, "This is a fatal message with a variable: %d", 42)
logger.FatalKV(ctx, "This is a fatal message with key-value pairs", "key1", "value1", "key2", "value2")

logger.Panic(ctx, "This is a panic message")
logger.Panicf(ctx, "This is a panic message with a variable: %d", 42)
logger.PanicKV(ctx, "This is a panic message with key-value pairs", "key1", "value1", "key2", "value2")
```

## Getting and Setting the Log Level 📏

You can get and set the current log level using the `Level` and `SetLevel` functions.

```go
currentLevel := logger.Level()
logger.SetLevel(zapcore.InfoLevel)
```

## Global Logger 🌐

You can get and set the global logger using the `Logger` and `SetLogger` functions.

```go
globalLogger := logger.Logger()
logger.SetLogger(globalLogger)
```

## License 📑

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

## Contributing 🤝

Contributions are welcome! Please feel free to submit a pull request or open an issue.
