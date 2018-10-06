package app

import (
	"net"
	"time"
)
import (
	log "github.com/sirupsen/logrus"
)
import (
	"github.com/zpeters/speedtest/internal/pkg/cmds"
	"github.com/zpeters/speedtest/internal/pkg/server"
)

// GetBestServer gets the first in the list
func GetBestServer() (bestserver server.Server, err error) {
	var bestspeed int64 = 999
	servers, err := server.GetAllServers()
	if err != nil {
		log.WithFields(log.Fields{
			"err":      err,
			"package":  "app",
			"function": "GetBestServer",
		}).Fatal()
	}
	for s := range servers {
		log.WithFields(log.Fields{
			"server":   servers[s].ID,
			"package":  "app",
			"function": "GetBestServer",
		}).Debug("GetBestServer")
		c := cmds.Connect(servers[s].Host)
		res := PingTest(c, 3)
		servers[s].BestTestPing = res
		log.WithFields(log.Fields{
			"speed":    res,
			"package":  "app",
			"function": "GetBestServer",
		}).Debug("GetBestServer")
		if res < bestspeed {
			bestspeed = res
			bestserver = servers[s]
		}
	}

	log.WithFields(log.Fields{
		"bestserver": bestserver,
		"package":    "app",
		"function":   "GetBestServer",
	}).Debug("GetBestServer")
	return bestserver, err
}

// DownloadTest runs download tests through numbytes until a download test takes more than numseconds
func DownloadTest(conn net.Conn, seedbytes int, numseconds int) (results float64) {
	var acc float64
	var numtests int = 1
	var sustained bool
	var numbytes int = seedbytes

	for {
		start := time.Now()
		res := cmds.Download(conn, numbytes)
		mbps := CalcMbps(res.Start, res.Finish, res.Bytes)
		acc = acc + mbps
		stop := time.Now()
		elapsed := stop.Sub(start)
		threshold := time.Second * time.Duration(numseconds)
		sustained = elapsed > threshold
		numtests++
		log.WithFields(log.Fields{
			"package":   "app",
			"function":  "DownloadTest",
			"testnum":   numtests,
			"elapsed":   elapsed,
			"threshold": threshold,
			"sustained": sustained,
		}).Info("sustained test")
		if sustained {
			break
		}
		numbytes = numbytes + (7 * seedbytes)
	}

	results = acc / float64(numtests)
	log.WithFields(log.Fields{
		"results":  results,
		"package":  "app",
		"function": "DownloadTest",
	}).Info("DownloadTest")
	return results
}

// UploadTest runs download tests through numbytes until a download test takes more than numseconds
func UploadTest(conn net.Conn, seedbytes int, numseconds int) (results float64) {
	var acc float64
	var numtests int = 1
	var sustained bool
	var numbytes int = seedbytes

	for {
		start := time.Now()
		res := cmds.Upload(conn, numbytes)
		mbps := CalcMbps(res.Start, res.Finish, res.Bytes)
		acc = acc + mbps
		stop := time.Now()
		elapsed := stop.Sub(start)
		threshold := time.Second * time.Duration(numseconds)
		sustained = elapsed > threshold
		numtests++
		log.WithFields(log.Fields{
			"package":   "app",
			"function":  "UploadTest",
			"testnum":   numtests,
			"elapsed":   elapsed,
			"threshold": threshold,
			"sustained": sustained,
		}).Info("sustained test")
		if sustained {
			break
		}
		numbytes = numbytes + (7 * seedbytes)
	}

	results = acc / float64(numtests)
	log.WithFields(log.Fields{
		"results":  results,
		"package":  "app",
		"function": "UploadTest",
	}).Info("UploadTest")
	return results
}

// PingTest gets roundtrip time to issue the "PING" command
func PingTest(conn net.Conn, numtests int) (results int64) {
	var acc int64

	for i := 1; i <= numtests; i++ {
		res := cmds.Ping(conn)
		acc = acc + res
	}

	results = acc / int64(numtests)
	log.WithFields(log.Fields{
		"results":  results,
		"package":  "app",
		"function": "PingTest",
	}).Info("PingTest")
	return results
}

func CalcMbps(start time.Time, finish time.Time, numbytes int) (mbps float64) {
	diff := finish.Sub(start)
	secs := float64(diff.Nanoseconds()) / float64(1000000000)
	megabits := float64(numbytes) / float64(125000)
	mbps = megabits / secs
	return mbps
}
