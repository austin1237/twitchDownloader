twitchDownloader
====
This is currently a go script that is using to [Livestreamer](http://docs.livestreamer.io/) to download 30 second clips from twitch.tv streams

### Dependencies
Since this script is using [Livestreamer](http://docs.livestreamer.io/install.html#dependencies) it has dependencies out side of golang. The outside dependencies are
1. python with at least version 2.6 or 3.3.
2. The Livestreamer package itself installed
3. Access to bash

### Setup
This assumes your in the project's directory

Build the binary

```sh
go install
```

### Start
The below example shows how to start recording clips of https://www.twitch.tv/drdisrespectlive and output the video files to the current directory labeled drdisrepect.mp4

```sh
$GOPATH/bin/twitchDownloader -t="TWITCH_OAUTH_TOKEN" -n="drdisrespectlive" -o="drdisrespect"
```

If the selected stream is offline the script will wait 1 minute and try downloading it again
