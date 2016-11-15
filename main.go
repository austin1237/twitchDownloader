package main

import (
	"os"
	"os/exec"
  "time"
	"flag"
  "log"
)

var (
	StreamName string
	Output string
	Token    string
	inProgress string
	finished string
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

func downloadStream (done chan string){
	commandFinished := make(chan error, 1)
  cmdString := "echo y |"
  cmdString = cmdString + " livestreamer https://www.twitch.tv/" + StreamName + " best"
  cmdString = cmdString + " --twitch-oauth-token=" + Token
  cmdString = cmdString + " -o=" + inProgress
  cmdString = cmdString + " --hls-segment-threads=3"
  cmd := exec.Command("bash", "-c", cmdString)
  cmd.Start()
	log.Println("command started")
	go func() {
    commandFinished <- cmd.Wait()
	}()
  select {
	// 30 seconds passed after commend ran
case <-time.After(30 * time.Second):
		log.Println("30 seconds have passed")
    if err := cmd.Process.Kill(); err != nil {
			done <- "process failed to be kill " + err.Error()
    }else{
			done <- ""
		}
	// command finished possible error sent through the channel commandFinished
	case err := <- commandFinished:
	log.Println("command died")
		if err != nil{
			done <- string(err.Error())
		}else{
			cmdOutput, _ := cmd.Output()
			done <- string(cmdOutput)
		}
  }
}

func renameVideo (done chan string){
	cmd := exec.Command("bash", "-c", "cp " + inProgress + " " + finished)
	cmd.Start()
	cmd.Wait()
	cmdOutput, _ := cmd.Output()
	done <- string(cmdOutput)
}

func recurisveDownload(done chan bool){
  downloaded := make(chan string, 1)
	renamed := make(chan string, 1)
  go downloadStream(downloaded)
  liveStreamerOutput := <- downloaded
	if liveStreamerOutput == "" {
		log.Println("clip downloaded successfully")
		go renameVideo(renamed)
		<- renamed
		log.Println("redownloading stream")
		recurisveDownload(done)
	} else {
		log.Println("An error occured waiting 1 minute to retry")
		<- time.After(1 * time.Minute)
		recurisveDownload(done)
	}
}

func main() {
  done := make(chan bool)
  recurisveDownload(done)
  <-done
  log.Println("program exited")
}
