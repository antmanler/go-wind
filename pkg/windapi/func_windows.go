// +build windows

package windapi

import (
	"errors"
	"reflect"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	ole "restis.dev/go-ole"
)

func newEventReceiver(C chan<- event) *eventReceiver {
	evt := &eventReceiver{
		C: C,
	}

	evtAddRef := func(this *ole.IUnknown) uintptr {
		pthis := (*eventReceiver)(unsafe.Pointer(this))
		pthis.ref++
		return uintptr(pthis.ref)
	}

	evtRelease := func(this *ole.IUnknown) uintptr {
		pthis := (*eventReceiver)(unsafe.Pointer(this))
		pthis.ref--
		return uintptr(pthis.ref)
	}

	evt.vtbl = new(evtVtbl)
	evt.vtbl.pAddRef = syscall.NewCallback(evtAddRef)
	evt.vtbl.pRelease = syscall.NewCallback(evtRelease)
	evt.vtbl.pQueryInterface = syscall.NewCallback(func(this *ole.IUnknown, iid *ole.GUID, punk **ole.IUnknown /*output*/) uintptr {
		*punk = nil
		if ole.IsEqualGUID(iid, ole.IID_IUnknown) || ole.IsEqualGUID(iid, ole.IID_IDispatch) {
			evtAddRef(this)
			*punk = this
			return uintptr(ole.S_OK)
		}
		s, _ := ole.StringFromCLSID(iid)
		if s == WindDataCOMEventsGUID {
			evtAddRef(this)
			*punk = this
			return uintptr(ole.S_OK)
		}
		return uintptr(ole.E_NOINTERFACE)
	})
	evt.vtbl.pGetTypeInfoCount = syscall.NewCallback(func(pcount *int) uintptr {
		if pcount != nil {
			*pcount = 0
		}
		return uintptr(ole.S_OK)
	})
	evt.vtbl.pGetTypeInfo = syscall.NewCallback(func(ptypeif *uintptr) uintptr {
		return ole.E_NOTIMPL
	})
	evt.vtbl.pGetIDsOfNames = syscall.NewCallback(func(this *ole.IUnknown, iid *ole.GUID, wnames uintptr /*[]*uint16*/, namelen int, lcid int, pdisp uintptr /*int32*/) uintptr {
		return uintptr(ole.S_OK)
	})
	evt.vtbl.pInvoke = syscall.NewCallback(func(
		this *ole.IDispatch,
		dispid int,
		riid *ole.GUID,
		lcid int,
		flags int16,
		dispparams *ole.DISPPARAMS,
		result *ole.VARIANT,
		pexcepinfo *ole.EXCEPINFO,
		nerr *uint) uintptr {

		type DISPPARAMS struct {
			rgvarg            uintptr
			rgdispidNamedArgs uintptr
			cArgs             uint32
			cNamedArgs        uint32
		}

		if dispid == 2 {
			param := (*DISPPARAMS)(unsafe.Pointer(dispparams))

			length := int(param.cArgs)
			args := make([]ole.VARIANT, 0, 0)

			// copy slice haeder
			h := *(*reflect.SliceHeader)((unsafe.Pointer)(&args))
			h.Data, h.Len, h.Cap = param.rgvarg, length, length
			args = *(*[]ole.VARIANT)(unsafe.Pointer(&h))

			state := int32(args[2].Val)
			reqid := int64(args[1].Val)
			ecode := int32(args[0].Val)

			self := (*eventReceiver)(unsafe.Pointer(this))

			self.C <- event{
				State:     state,
				RequestID: reqid,
				ErrCode:   ecode,
			}
		}
		return ole.S_OK
	})

	return evt
}

func adviseEventReceiver(source *ole.IDispatch, r *eventReceiver) (err error) {
	cp, err := findWindConnectionPoint(source)
	if err != nil {
		return err
	}
	defer cp.Release()

	r.cookie, err = cp.Advise((*ole.IUnknown)(unsafe.Pointer(r)))
	if err != nil {
		r.cookie = 0
	}
	return
}

func unadviseEventReceiver(source *ole.IDispatch, r *eventReceiver) error {
	cp, err := findWindConnectionPoint(source)
	if err != nil {
		return err
	}
	defer cp.Release()

	return cp.Unadvise(r.cookie)

}

func findWindConnectionPoint(wind *ole.IDispatch) (*ole.IConnectionPoint, error) {
	unknown, err := wind.QueryInterface(ole.IID_IConnectionPointContainer)
	if err != nil {
		return nil, err
	}
	defer unknown.Release()

	// convert to container
	container := (*ole.IConnectionPointContainer)(unsafe.Pointer(unknown))

	var cp *ole.IConnectionPoint
	iid, _ := ole.CLSIDFromString(WindDataCOMEventsGUID)
	if err := container.FindConnectionPoint(iid, &cp); err != nil {
		return nil, err
	}
	if cp != nil {
		return cp, nil
	}

	return nil, errors.New("wind: got empty ConnectionPoint")
}

func getCurrentThreadID() uint32 {
	return windows.GetCurrentThreadId()
}

func postMessage0012(tid uint32) (res uintptr, err error) {
	moduser32, _ := syscall.LoadDLL("user32.dll")
	postMessage, _ := moduser32.FindProc("PostThreadMessageW")
	res, _, err = postMessage.Call(uintptr(tid), uintptr(0x0012), 0x0, 0x0)
	return
}
