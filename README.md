# BeatSync

BeatSync is a simple GoLang application to generate a synchronized Beat signal. It is a synchronized metronome.

The BestSync sync beat in minutes basis. The first beat of all applicaitons start the begining of Minute period. Periodically, it also
perform a resync-up task (the Adapt Mechanims) to adapt the beat moment.

The Adapt Mechanim compute, periodically each 20, the time-shift of the local machine respect to the NTP server.  It perform a moving average with the previous value and generate the time-shift. The new time-shift value will be used to adapt the beat moment. If the time-shift is larger than max_error_to_skip, it will consider it as a error in NTP. If the error persist for 5 period, the it will also take it as new value and adapt. 

## Compile

You will need to install following Go packages

```
go get github.com/faiface/beep
go get github.com/beevik/ntp
go get github.com/faiface/beep
go get github.com/faiface/beep/speaker
```

and then, simply run

```
make
```

## Usage

```
Usage of ./beat_sync:
  -frequency int
    	Frequency in #Beats/Min. Default=200 (default 200)
  -max_error_to_skip int
    	Maximum error (in ms) that can tolerate in Adapt Mechanism. Default=5ms (default 5)
  -ntp_server string
    	NTP server to use. Default=0.es.pool.ntp.org (default "0.es.pool.ntp.org")
```