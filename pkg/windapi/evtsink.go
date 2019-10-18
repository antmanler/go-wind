package windapi

type event struct {
	State     int32
	RequestID int64
	ErrCode   int32
}

// nolint
type eventReceiver struct {
	vtbl   *evtVtbl
	ref    int32
	cookie uint32

	C chan<- event
}

// nolint
type evtVtbl struct {
	pQueryInterface   uintptr
	pAddRef           uintptr
	pRelease          uintptr
	pGetTypeInfoCount uintptr
	pGetTypeInfo      uintptr
	pGetIDsOfNames    uintptr
	pInvoke           uintptr
}
