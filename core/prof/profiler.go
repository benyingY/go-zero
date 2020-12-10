package prof

import "github.com/tal-tech/go-zero/core/utils"

type (
	ProfilePoint struct {
		*utils.ElapsedTimer
	}

	//接口
	Profiler interface {
		Start() ProfilePoint
		Report(name string, point ProfilePoint)
	}

	//实现了上述接口
	RealProfiler struct{}

	NullProfiler struct{}
)

//全局变量
var profiler = newNullProfiler()

//暴露出去的接口
func EnableProfiling() {
	profiler = newRealProfiler()
}

func Start() ProfilePoint {
	return profiler.Start()
}

func Report(name string, point ProfilePoint) {
	profiler.Report(name, point)
}

func newRealProfiler() Profiler {
	return &RealProfiler{}
}

func (rp *RealProfiler) Start() ProfilePoint {
	return ProfilePoint{
		ElapsedTimer: utils.NewElapsedTimer(),
	}
}

func (rp *RealProfiler) Report(name string, point ProfilePoint) {
	duration := point.Duration()
	report(name, duration)
}

func newNullProfiler() Profiler {
	return &NullProfiler{}
}

func (np *NullProfiler) Start() ProfilePoint {
	return ProfilePoint{}
}

func (np *NullProfiler) Report(string, ProfilePoint) {
}
