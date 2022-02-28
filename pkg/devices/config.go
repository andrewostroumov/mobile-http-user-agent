package devices

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"os"
	"reflect"
)

type Android struct {
	API     uint `json:"api"`
	ID      uint `json:"id"`
	Version string `json:"version"`
}

type Chrome struct {
	ChromeRev string `json:"chrome_rev"`
	WebkitRev string `json:"webkit_rev"`
}

type Config struct {
	ChromeFile  string
	AndroidFile string
}

var and []Android
var chr []Chrome
var contextLogger *log.Entry

func Parse(config Config, logger *log.Logger) {
	contextLogger = logger.WithFields(log.Fields{"package": "devices"})
	parseChromes(config.ChromeFile)
	parseAndroids(config.AndroidFile)
}

func parseChromes(file string) {
	v := &struct {
		Chromes *[]Chrome `json:"chrome"`
	}{&chr}

	parseFile(file, v)
}

func parseAndroids(file string) {
	v := &struct {
		Androids *[]Android `json:"android"`
	}{&and}

	parseFile(file, v)
}

func parseFile(file string, v interface{}) {
	b, err := os.ReadFile(file)
	if err != nil {
		contextLogger.Errorf("%v read file error: %v", file, err)
		os.Exit(1)
	}

	err = json.Unmarshal(b, v)

	if err != nil {
		contextLogger.Errorf("%v integrity error: %v", file, err)
		os.Exit(1)
	}

	ref := reflect.ValueOf(v)
	f := ref.Elem().Field(0)
	contextLogger.Infof("%v read %d entries", file, f.Elem().Len())
}
