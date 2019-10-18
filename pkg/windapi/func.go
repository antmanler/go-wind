// +build !windows

package windapi

import (
	"errors"

	ole "restis.dev/go-ole"
)

var errUnsupportedErr = errors.New("Unsupported system")

func newEventReceiver(C chan<- event) *eventReceiver {
	return &eventReceiver{}
}

func adviseEventReceiver(source *ole.IDispatch, r *eventReceiver) (err error) {
	return errUnsupportedErr
}

func unadviseEventReceiver(source *ole.IDispatch, r *eventReceiver) (err error) {
	return errUnsupportedErr
}

func getCurrentThreadID() uint32 {
	return 0
}

func postMessage0012(tid uint32) (res uintptr, err error) {
	return 1, errUnsupportedErr
}
