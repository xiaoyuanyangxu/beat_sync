# BeatSync

BeatSync is a simple GoLang application to generate a synchronized Beat signal. It is a synchronized metronome.

The BestSync sync beats in minutes basis. The first beat of all running applications starts the begining of Minute period. Periodically, it also
perform a resync-up task (the Adapt Mechanim) to adapt the beat moment.

The Adapt Mechanim computes, periodically each 20 seconds, the time-shift of the local machine respect to the NTP server.  It perform a moving average with the previous value and generate the new time-shift. The new time-shift value will be used to adapt the beat moment. If the time-shift is larger than max_error_to_skip, it will consider it as a error in NTP. However, if the error persist for 5 period, the it will take it as new value and adapt. 

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
    	Frequency in #Beats/Min. (default 200)
  -max_error_to_skip int
    	Maximum error (in ms) that can tolerate in Adapt Mechanism. (default 5)
  -ntp_server string
    	NTP server to use. (default "0.es.pool.ntp.org")
```