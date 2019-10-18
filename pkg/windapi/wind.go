package windapi

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"k8s.io/klog"
	ole "restis.dev/go-ole"
	"restis.dev/go-wind/pkg/errs"
)

// Subscription is returned from wind's WSQ
type Subscription struct {
	c chan []*WindData

	err   error
	wind  *windObj
	reqid uint64
}

// C returns channel for receiving data
func (subs *Subscription) C() <-chan []*WindData {
	return subs.c
}

// Close unsubscribes and cleans up
func (subs *Subscription) Close() error {
	if subs.err != nil {
		return subs.err
	}
	// cancel first
	subs.err = subs.wind.cancel(subs.reqid)
	subs.wind.io.Lock()
	defer subs.wind.io.Unlock()
	subs.wind.io.ds[subs.reqid] = nil
	delete(subs.wind.io.ds, subs.reqid)
	// close receiving channel
	close(subs.c)
	return subs.err
}

// well known constants
const (
	ProgramID             = "WINDDATACOM.WindDataCOMCtrl.1"
	WindDataCOMEventsGUID = "{831F1C39-C657-4849-BE13-C52A9BC064ED}"
)

// ErrAPINotOpen the apiInst is nil
var (
	ErrAPINotOpen = errors.New("wind: api does not open")
	ErrClosing    = errors.New("wind: io is closing")
)

var (
	apiInst *windObj
	apiLock sync.RWMutex
)

// Open opens and starts wind's api,
// it internally creates a COM object and starts a message queue,
// if the object has been initialized, it reset the logger.
func Open() (err error) {
	apiLock.Lock()
	defer apiLock.Unlock()

	if apiInst == nil {
		apiInst, err = newAPI()
		return
	}
	return
}

// Close shuts donw the underlying COM object, and cleans up
func Close() (err error) {
	apiLock.Lock()
	defer apiLock.Unlock()
	if apiInst != nil {
		err = apiInst.close()
		apiInst = nil
		return
	}
	return ErrAPINotOpen
}

// WSQ subscribes realtime data using wind's api
func WSQ(codes, fields, options string) (*Subscription, error) {
	apiLock.RLock()
	defer apiLock.RUnlock()
	if apiInst != nil {
		return apiInst.WSQ(codes, fields, options)
	}
	return nil, ErrAPINotOpen
}

// WSS returns multidimensional data from wind
func WSS(codes, fields, options string) ([]*WindData, error) {
	apiLock.RLock()
	defer apiLock.RUnlock()
	if apiInst != nil {
		return apiInst.WSS(codes, fields, options)
	}
	return nil, ErrAPINotOpen
}

// IsConnected checks api connection status
func IsConnected() bool {
	apiLock.RLock()
	defer apiLock.RUnlock()
	if apiInst != nil {
		return apiInst.IsConnected()
	}
	return false
}

// windObj is wrapper of wind's COM
type windObj struct {
	wind *ole.IDispatch

	msgloop struct {
		running bool // flag if main loop is running
	}

	evtsink *eventReceiver

	C <-chan event

	ctx struct {
		context.Context
		cancel context.CancelFunc
		tid    uint32
	}

	wg sync.WaitGroup

	io struct {
		sync.Once
		sync.RWMutex
		ds map[uint64]*Subscription
	}

	err error
}

// newAPI creates a new windapi object
func newAPI() (*windObj, error) {
	w := &windObj{}
	w.ctx.Context, w.ctx.cancel = context.WithCancel(context.Background())

	// All initialisation should occur in the same OS thread,
	// for it's main message loop to reside in.
	// Here we use a error channel to coordinate this process
	waitErrC := make(chan error)

	w.wg.Add(1)
	go func() {
		defer w.wg.Done()

		// lock on current os thread
		runtime.LockOSThread()
		defer runtime.UnlockOSThread()

		eventC := make(chan event, 1)
		defer close(eventC)

		// init message loop
		if err := func() error {
			if err := ole.CoInitializeEx(0, 0x0); err != nil {
				return err
			}

			unknown, err := createObject(ProgramID)
			if err != nil {
				return err
			}
			defer unknown.Release()

			w.wind, err = unknown.QueryInterface(ole.IID_IDispatch)
			if err != nil {
				return err
			}

			// connect event source to sink
			r := newEventReceiver(eventC)
			if err := adviseEventReceiver(w.wind, r); err != nil {
				return errs.And(err, w.close())
			}
			w.evtsink = r
			w.C = eventC

			return w.start("", "", 5000)

		}(); err != nil {
			waitErrC <- err
			return
		}

		// init finished,
		// notify outside world, we r ready to work
		close(waitErrC)

		w.ctx.tid = getCurrentThreadID()

		w.msgloop.running = true
		klog.Infof("(wind) message loop started at tid: %d", w.ctx.tid)

		defer klog.Infof("(wind) message loop exited")
		defer ole.CoUninitialize()
		defer func() { w.msgloop.running = false }()

		var m ole.Msg
		for w.ctx.Err() == nil {
			rc, _ := ole.GetMessage(&m, 0, 0, 0)
			if rc == 0 {
				klog.Info("(wind) message loop with return code: 0")
				break
			}
			if rc != -1 {
				ole.DispatchMessage(&m)
			}
		}
	}()

	if err := <-waitErrC; err != nil {
		return nil, err
	}

	return w, nil
}

func (wind *windObj) IsConnected() bool {
	if !wind.msgloop.running {
		return false
	}
	var state int32
	_, err := callMethod(wind.wind, "getConnectionState", &state)
	return err == nil && state == 0
}

// WSQ subscribes realtime data using wind's api
func (wind *windObj) WSQ(codes, fields, options string) (*Subscription, error) {
	// ensure starts ioloop lazily
	wind.io.Do(func() {
		wind.io.ds = make(map[uint64]*Subscription)
		wind.wg.Add(1)
		go wind.ioloop()
	})

	wind.io.Lock()
	defer wind.io.Unlock()
	reqid, err := wind.wsq(codes, fields, options)
	if err != nil {
		return nil, err
	}

	subs := &Subscription{
		c:     make(chan []*WindData, 1),
		reqid: reqid,
		wind:  wind,
	}
	wind.io.ds[reqid] = subs

	return subs, nil
}

// WSS returns multidimensional data from wind
func (wind *windObj) WSS(codes, fields, options string) ([]*WindData, error) {
	return wind.getWindData(func(codesOut, fieldsOut, timesOut *ole.VARIANT, ec *int32) (*ole.VARIANT, error) {
		return callMethod(wind.wind, "wss_syn", codes, fields, options, codesOut, fieldsOut, timesOut, ec)
	})
}

// close closes the wind api object and cleans up
func (wind *windObj) close() error {
	if wind.ctx.Err() != nil {
		return wind.err
	}

	if wind.wind != nil {
		if wind.evtsink != nil {
			wind.cancel(0) // nolint
			if err := unadviseEventReceiver(wind.wind, wind.evtsink); err != nil {
				wind.err = errs.And(wind.err, err)
			}
		}

		wind.ctx.cancel()
		wind.evtsink = nil

		wind.wind.Release()
		wind.wind = nil

		if wind.ctx.tid != 0 {
			if r0, err := postMessage0012(wind.ctx.tid); r0 != 0 {
				wind.wg.Wait()
			} else {
				klog.Warningf("(wind) post close message with error: %v(%v)", err, r0)
			}
			wind.ctx.tid = 0
		}
	}

	return wind.err
}

func (wind *windObj) ioloop() {
	klog.Info("(wind) ioloop started")
	defer wind.wg.Done()

	var (
		evtCnt int
		evt    event
		ok     bool
	)

	defer func() {
		klog.Infof("(wind) ioloop exited, deliveried #%d messages", evtCnt)
	}()

IOLOOP:
	for {
		select {
		case evt, ok = <-wind.C:
			if !ok {
				break IOLOOP
			}
			evtCnt++
		case <-wind.ctx.Done():
			break IOLOOP
		}

		if evt.State != 1 {
			klog.Warningf("(wind) io, unrecognized state code: %d", evt.State)
			continue
		}

		reqid := uint64(evt.RequestID)
		data, err := wind.readdata(reqid)
		if err != nil {
			klog.Errorf("(wind) io, failed to get updates for %d", reqid)
			continue
		}

		wind.io.RLock()
		if subs, ok := wind.io.ds[reqid]; ok {
			select {
			case subs.c <- data:
			case <-wind.ctx.Done():
				klog.Warningf("(wind) io, canceled when sending data to %d, may loose data", reqid)
				wind.io.RUnlock()
				break IOLOOP
			}
		}
		wind.io.RUnlock()
	}

	wind.io.Lock()
	defer wind.io.Unlock()
	for key := range wind.io.ds {
		subs := wind.io.ds[key]
		subs.err = ErrClosing
		close(subs.c)

		// clean
		wind.io.ds[key] = nil
		delete(wind.io.ds, key)
	}
}

func (wind *windObj) enableAsyn() (err error) {
	_, err = callMethod(wind.wind, "enableAsyn", 1)
	return
}

func (wind *windObj) stop() error {
	res, err := callMethod(wind.wind, "stop")
	if err != nil {
		return err
	}
	if err = parseErr(int32(res.Val)); err != nil {
		return err
	}
	return nil
}

func (wind *windObj) start(option1, option2 string, timeout int32) error {
	res, err := callMethod(wind.wind, "start_cpp", option1, option2, timeout)
	if err != nil {
		return err
	}
	if err = parseErr(int32(res.Val)); err != nil {
		return err
	}
	return nil
}

func (wind *windObj) cancel(reqid uint64) error {
	_, err := callMethod(wind.wind, "cancelRequest", reqid)
	if err != nil {
		return err
	}
	return nil
}

func (wind *windObj) wsq(codes, fields, options string) (reqid uint64, err error) {
	var (
		res     *ole.VARIANT
		errCode int32
	)
	options += ";REALTIME=Y"
	res, err = callMethod(wind.wind, "wsq", codes, fields, options, &errCode)
	if err != nil {
		return
	}
	if err = parseErr(errCode); err != nil {
		return 0, err
	}
	reqid = uint64(res.Val)
	return reqid, nil
}

type rawData struct {
	data      ole.VARIANT // 数据
	codes     ole.VARIANT // Code列表
	fields    ole.VARIANT // 指标列表
	times     ole.VARIANT // 时间列表
	stateCode int32       // 状态码
}

func (wind *windObj) readdata(reqid uint64) (data []*WindData, err error) {
	var rs int32
	return wind.getWindData(func(codes, fields, times *ole.VARIANT, ec *int32) (*ole.VARIANT, error) {
		return callMethod(wind.wind, "readdata", reqid, codes, fields, times, &rs, ec)
	})
}

func (wind *windObj) getWindData(fn func(codes, fields, times *ole.VARIANT, ec *int32) (*ole.VARIANT, error)) (data []*WindData, err error) {
	var (
		raw rawData
		rs  int32
		ec  int32
	)
	res, err := fn(&raw.codes, &raw.fields, &raw.times, &ec)
	raw.data = *res
	raw.stateCode = rs

	if err != nil {
		return nil, err
	}
	if err = parseErr(ec); err != nil {
		return nil, err
	}
	data, err = parseRawData(&raw)
	err = errs.And(err, raw.codes.Clear(), raw.fields.Clear(), raw.times.Clear(), raw.data.Clear())
	return
}

func checkSafeArray(name string, v *ole.VARIANT) (*ole.SafeArrayConversion, error) {
	arr := v.ToArray()
	if arr == nil {
		return nil, fmt.Errorf("wind: invlalid %s", name)
	}
	return arr, nil
}

func msTsToTime(msts float64) time.Time {
	val := msts - 693960
	day := time.Duration(int64(val))
	ns := time.Millisecond * time.Duration((val*float64(time.Hour*24)-float64(day*time.Hour*24))/float64(time.Millisecond)+0.5)
	return time.Date(1899, 12, 30, 0, 0, 0, 0, time.Local).Add(day*time.Hour*24 + ns)
}

func parseRawData(raw *rawData) ([]*WindData, error) {
	val, err := checkSafeArray("codes", &raw.codes)
	if val == nil {
		return nil, err
	}
	codes := val.ToStringArray()
	val, err = checkSafeArray("fields", &raw.fields)
	if val == nil {
		return nil, err
	}
	fields := val.ToStringArray()
	val, err = checkSafeArray("times", &raw.times)
	if val == nil {
		return nil, err
	}
	times := val.ToValueArray()
	val, err = checkSafeArray("data", &raw.data)
	if val == nil {
		return nil, err
	}
	data := val.ToValueArray()

	w := len(fields)
	out := make([]*WindData, len(times)*len(codes))
	for i, tval := range times {
		tm := msTsToTime(tval.(float64))
		for j, code := range codes {
			out[i+j] = &WindData{
				UpdateTime: tm,
				WindCode:   code,
				Fields:     fields[:],
				Values:     data[0:w],
				CreatedAt:  time.Now(),
			}
			data = data[w:]
		}
	}

	return out, nil
}
