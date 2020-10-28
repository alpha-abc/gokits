package logv1_test

import (
	"fmt"
	"os"
	"testing"
	"time"

	logv1 "github.com/alpha-abc/gokits/log/logv1"
)

/*
export GOPATH=/root/devel/golang/go-libs/
go test go-log -v -test.run Test_Output
*/
func Test_Output(t *testing.T) {
	var fs = []*os.File{
		os.Stdout,
	}

	//var logger = NewLogger(fs, LevelDebug, "2006-01-02 15:04:05.000", "D", 0)
	var logger = &logv1.Logger{
		Outputs:    fs,
		Level:      logv1.LevelDebug,
		TimeFormat: "2006-01-02 15:04:05.000",
		BackupType: "D",
		CallDepth:  2,
		InitTime:   time.Now(),
	}

	logger.Debug("12", "asd")
	logger.Debugf("%s - %s", "12", "34")

	logger.Info("12", "asd")
	logger.Warn("12", "asd")
	logger.Error("12", "asd")

	fmt.Printf("%s", "abc")
	fmt.Printf("%s", "def")
	//logger.Fatal("12", "asd")
}

// func Test_FileSize(t *testing.T) {
// 	var f, _ = os.OpenFile("/root/devel/golang/go-libs/src/go-log/LICENSE", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)

// 	var fi, _ = f.Stat()
// 	t.Log(fi.Size())
// }

// /*
//  go test go-log -v -test.bench Benchmark_Mylog
//   200000	     11370 ns/op
// PASS
// ok  	go-log	2.383s
// */
// func Benchmark_Mylog(b *testing.B) {
// 	var f, _ = os.OpenFile("/tmp/testme.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)

// 	var fs = []*os.File{
// 		f,
// 	}

// 	//var logger = NewLogger(fs, LevelDebug, "2006-01-02 15:04:05.000", "D", 0)
// 	var logger = &Logger{
// 		Outputs:    fs,
// 		Level:      LevelDebug,
// 		TimeFormat: "2006-01-02 15:04:05.000",
// 		BackupType: "D",
// 		CallDepth:  2,
// 		InitTime:   time.Now(),
// 	}

// 	for i := 0; i < b.N; i++ {
// 		logger.Debug("benchmark test测试")
// 	}
// }

// /*
//   300000	      5071 ns/op
// PASS
// ok  	go-log	1.575s
// */
// func Benchmark_Syslog(b *testing.B) {
// 	var f, _ = os.OpenFile("/tmp/testsys.log", os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
// 	var l = log.New(f, "", log.LstdFlags|log.Lshortfile)

// 	for i := 0; i < b.N; i++ {
// 		l.Println("benchmark test测试")
// 	}
// }

