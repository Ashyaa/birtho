package bot

func (b *Bot) Debug(format string, args ...interface{}) {
	b.Log.Debugf(format, args...)
}

func (b *Bot) DebugE(err error, format string, args ...interface{}) {
	actualFormat := format + ": %s"
	actualArgs := append(args, err.Error())
	b.Log.Debugf(actualFormat, actualArgs...)
}

func (b *Bot) Warn(format string, args ...interface{}) {
	b.Log.Warnf(format, args...)
}

func (b *Bot) WarnE(err error, format string, args ...interface{}) {
	actualFormat := format + ": %s"
	actualArgs := append(args, err.Error())
	b.Log.Warnf(actualFormat, actualArgs...)
}

func (b *Bot) Error(format string, args ...interface{}) {
	b.Log.Errorf(format, args...)
}

func (b *Bot) ErrorE(err error, format string, args ...interface{}) {
	actualFormat := format + ": %s"
	actualArgs := append(args, err.Error())
	b.Log.Errorf(actualFormat, actualArgs...)
}

func (b *Bot) Info(format string, args ...interface{}) {
	b.Log.Infof(format, args...)
}

func (b *Bot) Fatal(format string, args ...interface{}) {
	b.Log.Fatalf(format, args...)
}

func (b *Bot) FatalE(err error, format string, args ...interface{}) {
	actualFormat := format + ": %s"
	actualArgs := append(args, err.Error())
	b.Log.Fatalf(actualFormat, actualArgs...)
}
