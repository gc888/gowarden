package logger

import (
	"errors"
	"github.com/404cn/gowarden/utils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
)

const logFile = "gowarden.log"

func New(level int) (*zap.SugaredLogger, error) {
	var sugar *zap.SugaredLogger

	loglevel, err := getZapLogLevel(level)
	if err != nil {
		return sugar, err
	}

	encoder := getEncoder()
	err = checkLogPath()
	if err != nil {
		return sugar, err
	}
	writer := getLogWriter()

	core := zapcore.NewTee(
		zapcore.NewCore(encoder, zapcore.AddSync(os.Stdout), loglevel),
		zapcore.NewCore(encoder, writer, loglevel),
	)
	sugar = zap.New(core, zap.AddCaller()).Sugar()

	return sugar, nil
}

func getZapLogLevel(l int) (zapcore.Level, error) {
	switch l {
	case -1:
		return zapcore.DebugLevel, nil
	case 0:
		return zapcore.InfoLevel, nil
	case 1:
		return zapcore.WarnLevel, nil
	case 2:
		return zapcore.ErrorLevel, nil
	case 3:
		return zapcore.DPanicLevel, nil
	case 4:
		return zapcore.PanicLevel, nil
	case 5:
		return zapcore.FatalLevel, nil
	default:
		return zapcore.InfoLevel, errors.New("please set correct log level")
	}
}

func checkLogPath() error {
	if utils.PathExist(logFile) {
		return nil
	} else {
		log.Println("Didn't find log file, try to create ... ")
		_, err := os.Create(logFile)
		return err
	}
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getLogWriter() zapcore.WriteSyncer {
	file, _ := os.Create(logFile)
	return zapcore.AddSync(file)
}
