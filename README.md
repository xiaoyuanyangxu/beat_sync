# BeatSync

Beat Sync is a simple GoLang application to generate a synchronized Beat signal. It is a synchronized metronome.

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
  -freqency int
    	Frequency in #Beats/Min. Default=200 (default 200)
  -max_error_to_skip int
    	Maximum error (in ms) that can tolerate in Adapt Mechanism. Default=5ms (default 5)
  -ntp_server string
    	NTP server to use. Default=0.es.pool.ntp.org (default "0.es.pool.ntp.org")
```