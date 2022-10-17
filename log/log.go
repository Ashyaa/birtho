package log

import (
	"io"
	"os"
	"path"
	FP "path/filepath"

	"github.com/mattn/go-colorable"
	LR "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func New() LR.Logger {
	res := *LR.New()
	res.SetOutput(colorable.NewColorableStdout())
	res.SetFormatter(&LR.TextFormatter{ForceColors: true, FullTimestamp: true})
	exePath, err := os.Executable()
	if err != nil {
		res.Fatal(err.Error())
	}
	logDir := FP.Join(path.Dir(exePath), "logs")
	if stat, err := os.Stat(logDir); err != nil {
		if os.IsNotExist(err) {
			os.Mkdir(logDir, 0644)
		}
	} else if !stat.IsDir() {
		res.Fatalf("%s exists but is not a directory", logDir)
	}
	mw := io.MultiWriter(&lumberjack.Logger{
		Filename:   FP.Join(logDir, "bot.log"),
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28, //days
	}, colorable.NewColorableStdout())
	res.SetOutput(mw)
	return res
}
