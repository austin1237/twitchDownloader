package main

import (
	"errors"
	"flag"
	"log"
	"os"
	"strings"
	"time"
	"os/exec"
)

var (
	StreamName string
	Output     string
	Token      string
	inProgress string
	finished   string
)

func init() {
	flag.StringVar(&StreamName, "n", "", "Twitch Stream Name")
	flag.StringVar(&Output, "o", "", "Output destination for video files")
	flag.StringVar(&Token, "t", "", "Twitch OAuth Token")
	flag.Parse()

	if StreamName == "" {
		log.Println("Twitch stream name was not provided")
		os.Exit(1)
	}

	if Output == "" {
		log.Println("Desination for video files was not chosen")
		os.Exit(1)
	}

	if Token == "" {
		log.Println("Twitch OAuth Token was not provided")
		os.Exit(1)
	}

	inProgress = Output + "InProgress.mp4"
	finished = Output + ".mp4"
}

func downloadStream(done chan error) {
	commandFinished := make(chan string, 1)
	cmdString := "echo y |"
	cmdString = cmdString + " livestreamer https://www.twitch.tv/" + StreamName + " best"
	cmdString = cmdString + " --twitch-oauth-token=" + Token
	cmdString = cmdString + " -o=" + inProgress
	cmdString = cmdString + " --hls-segment-threads=3"
	cmd := exec.Command("bash", "-c", cmdString)
	// runs the command and waits for its output
	go func() {
		cmdOutputByteArray, _ := cmd.Output()
		commandFinished <- string(cmdOutputByteArray)
	}()
	select {
	// 30 seconds passed after command ran
	// livestreamer will not exit will downloading so we'll kill it
	case <-time.After(30 * time.Second):
		log.Println("30 seconds have passed")
		err := cmd.Process.Kill()
		done <- err
	// bash command exited before 30 seconds were up
	case cmdOutput := <-commandFinished:
		if strings.Contains(cmdOutput, "No streams found on this URL") {
			done <- errors.New("stream is offline")
		} else {
			done <- errors.New("unkown error " + cmdOutput)
		}
	}
}

func renameVideo(done chan string) {
	cmd := exec.Command("bash", "-c", "cp "+ inProgress +" "+ finished)
	cmdOutputByteArray, _ := cmd.Output()
	done <- string(cmdOutputByteArray)
}

func recurisveDownload(done chan bool) {
	downloaded := make(chan error, 1)
	renamed := make(chan string, 1)
	go downloadStream(downloaded)
	downloadErr := <-downloaded
	if downloadErr == nil {
		log.Println("clip downloaded successfully")
		go renameVideo(renamed)
		<-renamed
		log.Println("redownloading stream")
		recurisveDownload(done)
	} else if downloadErr.Error() == "stream is offline" {
		log.Println("stream is offline will retry in a minute")
		<-time.After(1 * time.Minute)
		recurisveDownload(done)
	} else {
		log.Println(downloadErr.Error())
		close(done)
	}
}

func main() {
	done := make(chan bool)
	recurisveDownload(done)
	<-done
	log.Println("program exited")
}
