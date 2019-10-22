package main

import (
	"math/rand"
	"runtime"
	"time"

	"github.com/refunc/refunc/pkg/utils/cmdutil"
	"github.com/refunc/refunc/pkg/utils/cmdutil/flagtools"
	"k8s.io/klog"
	"restis.dev/go-wind/pkg/windapi"
)

func main() {
	runtime.GOMAXPROCS(func() int {
		if runtime.NumCPU() > 128 {
			return runtime.NumCPU()
		}
		return 128
	}())
	rand.Seed(time.Now().UTC().UnixNano())

	flagtools.InitFlags()
	defer klog.Flush()

	err := windapi.Open()
	if err != nil {
		klog.Exit(err)
	}
	defer windapi.Close()

	subs, err := windapi.WSQ("600588.SH", "rt_date,rt_time,rt_pre_close,rt_open,rt_high,rt_low,rt_last,rt_pre_settle,rt_settle,rt_ask1,rt_bid1", "")
	if err != nil {
		klog.Exit(err)
	}

	for {
		select {
		case data := <-subs.C():
			klog.Info(data)
		case <-cmdutil.GetSysSig():
			return
		}
	}
}
