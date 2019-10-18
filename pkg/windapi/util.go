package windapi

import ole "restis.dev/go-ole"

func createObject(programID string) (unknown *ole.IUnknown, err error) {
	classID, err := ole.ClassIDFrom(programID)
	if err != nil {
		return
	}

	unknown, err = ole.CreateInstance(classID, ole.IID_IUnknown)
	if err != nil {
		return
	}

	return
}

func callMethod(disp *ole.IDispatch, name string, params ...interface{}) (result *ole.VARIANT, err error) {
	return disp.InvokeWithOptionalArgs(name, ole.DISPATCH_METHOD, params)
}
