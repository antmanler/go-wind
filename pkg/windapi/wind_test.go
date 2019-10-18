package windapi

import (
	"testing"
	"time"
)

func TestNewRelease(t *testing.T) {
	wind, err := newAPI()
	panicOnErr(err)
	panicOnErr(wind.close())
}

func TestWindwsq(t *testing.T) {
	wind, err := newAPI()
	panicOnErr(err)
	defer func() { panicOnErr(wind.close()) }()

	codes := "IF1703.CFE,USDCNY.IB,CU1701.SHF"
	fields := "rt_date,rt_time,rt_pre_close,rt_open,rt_high,rt_low,rt_last,rt_latest,rt_chg,rt_swing,rt_pre_settle,rt_settle"
	reqid, err := wind.wsq(codes, fields, "")
	panicOnErr(err)
	t.Logf("request id: %d", reqid)

	for {
		select {
		case ev, ok := <-wind.C:
			t.Logf("evt: %v, ok: %v", ev, ok)
			if ev.State == 1 {
				data, err := wind.readdata(uint64(ev.RequestID))
				t.Logf("data: %v, err: %v", data, err)
				return
			}
		case <-time.After(time.Second * 10):
			t.Error("timeout without message")
			return
		}
	}
}

func TestWSQ(t *testing.T) {
	wind, err := newAPI()
	panicOnErr(err)
	defer func() { panicOnErr(wind.close()) }()

	codes := "IF1703.CFE,USDCNY.IB,CU1701.SHF"
	fields := "rt_date,rt_time,rt_pre_close,rt_open,rt_high,rt_low,rt_last,rt_latest,rt_chg,rt_pct_chg,rt_swing,rt_settle"
	subs, err := wind.WSQ(codes, fields, "")
	panicOnErr(err)

	select {
	case data, ok := <-subs.C():
		if !ok {
			t.Error("channel closed")
		}
		t.Logf("data:%v", data)
	case <-time.After(time.Second * 10):
		t.Error("timeout without message")
		return
	}
}

func TestWSS(t *testing.T) {
	wind, err := newAPI()
	panicOnErr(err)
	defer func() { panicOnErr(wind.close()) }()

	codes := "IF1703.CFE,USDCNY.IB,CU1701.SHF"
	fields := "sec_name,sec_englishname,windcode,trade_code"
	data, err := wind.WSS(codes, fields, "")
	panicOnErr(err)
	t.Log(data)
}

func panicOnErr(err error) {
	if err != nil {
		panic(err)
	}
}
