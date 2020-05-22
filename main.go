package main

import (
	"flag"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/beevik/ntp"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
)

var sampleRate int64 = 44100
var nSamples int64 = 0
var beatsCount int64 = 0
var beatsPerMinute int64 = 120
var beatsSeparation int64 = 0
var beatWidth int64 = 1
var beatOn = false

var mux sync.Mutex
var globalTimeShift int64 = 0
var currentTimeShift int64 = 0
var errorEstimation float64 = 0.0

var syncPeriodInBeats int64 = 20
var syncdTime time.Time
var nano int64 = 1000 * 1000 * 1000
var skippedAdapPeriod = 0
var maxSkipPeriod = 5
var maxErrorToSkip = 5

var nTPRequestTimePeriodSec = 20
var nTPRequestPerPeriod = 1
var nTPServer = "0.es.pool.ntp.org"

func timeSync() (int64, int64, time.Time, error) {
	var acc time.Duration

	for i := 0; i < nTPRequestPerPeriod; i++ {
		response, err := ntp.Query(nTPServer)
		if err != nil {
			return 0, 0, time.Now(), err
		}
		time.Sleep(1 * time.Second)
		acc += response.ClockOffset
	}

	t := time.Now().Add(time.Duration(acc.Nanoseconds()/int64(nTPRequestPerPeriod)) * time.Nanosecond)

	beatWidth := float64(sampleRate) / float64(nano)
	offset := int64((float64(t.Second())*float64(nano) + float64(t.Nanosecond())) * beatWidth)

	return offset, acc.Nanoseconds() / int64(nTPRequestPerPeriod), t, nil
}

func syncupLoop() {
	for {
		_, timeShift, _, err := timeSync()
		sleepTime := nTPRequestTimePeriodSec
		if err != nil {
			fmt.Println("\tError in getting NTP clock")
			sleepTime = nTPRequestTimePeriodSec / 2
		} else {
			var currentTimeShift int64
			mux.Lock()
			currentTimeShift = globalTimeShift
			mux.Unlock()

			timeShift = int64(float64(currentTimeShift)*0.5 + float64(timeShift)*0.5)
			/*
				fmt.Println("\t------------------")
				fmt.Println("\tNew Time Shift:", timeShift)
				fmt.Println("\t------------------")
			*/
			mux.Lock()
			globalTimeShift = timeShift
			mux.Unlock()
		}
		time.Sleep(time.Duration(sleepTime) * time.Second)
	}
}

func BeatGenerator() beep.Streamer {
	return beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		for i := range samples {
			v := 0.0
			if (nSamples % beatsSeparation) < beatWidth {
				v = 1.0
				if beatOn == false {
					beatsCount++
					/*
						if beats_count%10 == 1 {
							fmt.Printf("Beat: #beats=%v #samples=%v offset=%v/%v time=%v\n", beats_count, n_samples, current_offset, global_offset, time.Now())
						}
					*/
					if beatsCount%syncPeriodInBeats == 0 {
						var cTimeShift int64 = 0
						mux.Lock()
						cTimeShift = globalTimeShift
						mux.Unlock()

						diff := float64(cTimeShift-currentTimeShift) / float64(nano)
						errorEstimatted := math.Abs(diff) * 1000.0

						if errorEstimatted < float64(maxErrorToSkip) || skippedAdapPeriod >= maxSkipPeriod {
							if skippedAdapPeriod > maxSkipPeriod {
								fmt.Printf(">>> Let force adapt after %v skipped periods\n", skippedAdapPeriod)
							}
							skippedAdapPeriod = 0
							if math.Abs(diff) > 1.0/float64(sampleRate) {
								shift := diff * float64(sampleRate)
								errorEstimation = errorEstimatted

								fmt.Printf(">> ERROR=~%0.1fms Change: value=%v #beats=%v #samples=%v time=%v\n",
									errorEstimation,
									int64(shift),
									beatsCount,
									nSamples,
									time.Now().Format("2006-01-02T15:04:05"))

								nSamples = nSamples + int64(shift)
								currentTimeShift = cTimeShift
							}
						} else {
							skippedAdapPeriod++
							fmt.Printf(">>> Let skip this adapt period, error is too high %0.1f/%dms #skipped=%v/%v\n",
								errorEstimatted,
								maxErrorToSkip,
								skippedAdapPeriod,
								maxSkipPeriod)
						}

					}
				}
				beatOn = true

			} else {
				beatOn = false

			}
			samples[i][0] = v
			samples[i][1] = v
			nSamples++
		}
		return len(samples), true
	})
}

func main() {

	flag.Int64Var(&beatsPerMinute, "frequency", 200, "Frequency in #Beats/Min.")
	flag.IntVar(&maxErrorToSkip, "max_error_to_skip", 5, "Maximum error (in ms) that can tolerate in Adapt Mechanism.")
	flag.StringVar(&nTPServer, "ntp_server", "0.es.pool.ntp.org", "NTP server to use.")
	flag.Parse()

	beatsSeparation = sampleRate * 60 / beatsPerMinute

	sr := beep.SampleRate(sampleRate)
	speaker.Init(sr, sr.N(time.Second))

	done := make(chan bool)
	fmt.Println("Frequency:", beatsPerMinute, " Max Error to Skip:", int64(maxErrorToSkip), "ms")
	fmt.Println("Sync-up the clock, it may take several seconds")

	var err error
	var timeShift int64
	nSamples, timeShift, _, err = timeSync()
	if err != nil {
		fmt.Println("Error in getting NTP clock")
		return
	}
	globalTimeShift = timeShift
	currentTimeShift = globalTimeShift

	go syncupLoop()

	speaker.Play(BeatGenerator())

	<-done // this will block the application
}
