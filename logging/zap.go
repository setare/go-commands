package logging

import (
	"fmt"
	"os"

	"go.uber.org/zap"
)

var Logger *zap.Logger

func init() {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed starting logger")
		panic(err)
	}
	Logger = logger
}
