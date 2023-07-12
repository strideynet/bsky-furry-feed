package testing

import (
	"reflect"
	"unsafe"

	"github.com/bluesky-social/indigo/testing"
	"github.com/ipfs/go-log"
)

func init() {
	log.SetAllLoggers(log.LevelDebug)
}

// black magic to set an unexported field on the TestBGS
func SetTrialHostOnBGS(tbgs *testing.TestBGS, rawHost string) {
	hosts := []string{rawHost}

	trialHosts := reflect.ValueOf(tbgs).
		Elem().FieldByName("tr").
		Elem().FieldByName("TrialHosts")

	reflect.NewAt(trialHosts.Type(), unsafe.Pointer(trialHosts.UnsafeAddr())).Elem().
		Set(reflect.ValueOf(hosts))
}
