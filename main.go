package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/beevik/ntp"
	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
)

var sample_rate int64 = 44100
var n_samples int64 = 0
var beats_count int64 = 0
var beats_per_minute int64 = 120
var beats_separation int64 = 0 //sample_rate * 60 / beats_per_minute
var beat_width int64 = 1
var beat_on = false

var mux sync.Mutex
var global_offset int64 = 0
var global_time_shift int64 = 0
var current_time_shift int64 = 0
var error_estimation float64 = 0.0

var sync_period_in_beats int64 = 20 //beats_per_minute
var syncd_time time.Time
var nano int64 = 1000 * 1000 * 1000
var skipped_adap_period = 0
var max_skip_period = 5
var max_error_to_skip = 5.0

func time_sync() (int64, int64, time.Time, error) {
	var acc time.Duration

	n := 1 // ntp sync period
	//fmt.Print("Progress: ")
	for i := 0; i < n; i++ {
		response, err := ntp.Query("0.es.pool.ntp.org")
		if err != nil {
			return 0, 0, time.Now(), err
		}
		time.Sleep(1 * time.Second)
		//fmt.Print(n-i+1, ".")
		acc += response.ClockOffset
	}

	t := time.Now().Add(time.Duration(acc.Nanoseconds()/int64(n)) * time.Nanosecond)

	beat_w := float64(sample_rate) / float64(nano)
	offset := int64((float64(t.Second())*float64(nano) + float64(t.Nanosecond())) * beat_w)

	return offset, acc.Nanoseconds(), t, nil
}

func syncup_loop() {
	for {
		_, timeShift, _, err := time_sync()
		sleep_time := 20 // Adapt period
		if err != nil {
			fmt.Println("\tError in getting NTP clock")
			sleep_time = 5
		} else {
			var currentTimeShift int64
			mux.Lock()
			currentTimeShift = global_time_shift
			mux.Unlock()

			timeShift = int64(float64(currentTimeShift)*0.5 + float64(timeShift)*0.5)
			/*
				fmt.Println("\t------------------")
				fmt.Println("\tOffset:", offset, float64(offset)/float64(sample_rate), "seconds")
				fmt.Println("\tNew Time Shift:", timeShift, " current shift:", global_time_shift)
				fmt.Println("\t------------------")
			*/
			mux.Lock()
			global_time_shift = timeShift
			mux.Unlock()
		}
		time.Sleep(time.Duration(sleep_time) * time.Second)
	}
}

func Noise() beep.Streamer {
	return beep.StreamerFunc(func(samples [][2]float64) (n int, ok bool) {
		for i := range samples {
			v := 0.0
			if (n_samples % beats_separation) < beat_width {
				v = 1.0
				if beat_on == false {
					beats_count++
					/*
						if beats_count%10 == 1 {
							fmt.Printf("Beat: #beats=%v #samples=%v offset=%v/%v time=%v\n", beats_count, n_samples, current_offset, global_offset, time.Now())
						}
					*/
					if beats_count%sync_period_in_beats == 0 {
						var currentTimeShift int64 = 0
						mux.Lock()
						currentTimeShift = global_time_shift
						mux.Unlock()

						diff := float64(currentTimeShift-current_time_shift) / float64(nano)
						errorEstimatted := math.Abs(diff) * 1000.0

						if errorEstimatted < max_error_to_skip || skipped_adap_period >= max_skip_period {
							if skipped_adap_period > max_skip_period {
								fmt.Printf(">>> Let force adapt after %v skipped periods\n", skipped_adap_period)
							}
							skipped_adap_period = 0
							if math.Abs(diff) > 1.0/float64(sample_rate) {
								shift := diff * float64(sample_rate)
								error_estimation = errorEstimatted

								fmt.Printf(">> ERROR=~%0.1fms Change: value=%v #beats=%v #samples=%v time=%v\n",
									error_estimation,
									int64(shift),
									beats_count,
									n_samples,
									time.Now().Format("2006-01-02T15:04:05"))

								n_samples = n_samples + int64(shift)
								current_time_shift = currentTimeShift
							}
						} else {
							skipped_adap_period++
							fmt.Printf(">>> Let skip this adapt period, error is too high %0.1f/%0.1fms #skipped=%v/%v\n",
								errorEstimatted,
								max_error_to_skip,
								skipped_adap_period,
								max_skip_period)
						}

					}
				}
				beat_on = true

			} else {
				beat_on = false

			}
			samples[i][0] = v
			samples[i][1] = v
			n_samples++
		}
		return len(samples), true
	})
}

func main() {

	if len(os.Args) >= 2 {
		i, err := strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Println("Error in parsing Frequency parameter")
			fmt.Printf("Usage: %v FREQUENCY(Beats/Min) MAX_ERROR_TO_SKIP(ms)\n", os.Args[0])
			return
		} else {
			beats_per_minute = int64(i)
		}

		if len(os.Args) > 2 {
			e, err := strconv.Atoi(os.Args[2])
			if err != nil {
				fmt.Println("Error in parsing Error parameter")
				fmt.Printf("Usage: %v FREQUENCY(Beats/Min) MAX_ERROR_TO_SKIP(ms)\n", os.Args[0])
				return

			}
			max_error_to_skip = float64(e)
		}
	}
	beats_separation = sample_rate * 60 / beats_per_minute

	sr := beep.SampleRate(sample_rate)
	speaker.Init(sr, sr.N(time.Second))

	done := make(chan bool)
	fmt.Println("Frequency:", beats_per_minute, " Max Error to Skip:", int64(max_error_to_skip), "ms")
	fmt.Println("Sync-up the clock, it may take several seconds")

	var err error
	var timeShift int64
	global_offset, timeShift, _, err = time_sync()
	if err != nil {
		fmt.Println("Error in getting NTP clock")
		return
	}
	n_samples = global_offset
	global_time_shift = timeShift
	current_time_shift = global_time_shift

	go syncup_loop()

	speaker.Play(Noise())

	/*speaker.Play(beep.Seq(beep.Take(sr.N(5*time.Second), Noise())), beep.Callback(func() {
		done <- true
	}))) */
	<-done
}
