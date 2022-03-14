package log

import (
	"os"

	"github.com/mattn/go-colorable"
	LR "github.com/sirupsen/logrus"
)

func New() LR.Logger {
	res := *LR.New()
	res.SetOutput(os.Stdout)
	res.SetFormatter(&LR.TextFormatter{ForceColors: true})
	res.SetOutput(colorable.NewColorableStdout())
	return res
}
