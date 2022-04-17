package logrus

import (
	"testing"
)

func TestJWFormatting(t *testing.T) {
	//SetOutput(os.Stdout)
	SetLevel(DebugLevel)
	SetReportCaller(true)
	tf := &JWFormatter{}
	SetFormatter(tf)
	Infof("hello word!! %s, value=%d", "key", 1)
	Debugf("test1111")
	Warnf("warning!!")
	Error("error!!!")
}
