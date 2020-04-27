package lib

import (
	"github.com/sirupsen/logrus"
)

var log *logrus.Logger

func init() {
	log = logrus.New()
	log.SetLevel(logrus.TraceLevel)
}
