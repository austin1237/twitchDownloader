package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awsutil"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	StreamName string
	Token      string
	awsID      string
	awsSecret  string
	bucket     string
	filePath   string
)

func init() {
	flag.StringVar(&StreamName, "n", "", "Twitch Stream Name")
	flag.StringVar(&Token, "t", "", "Twitch OAuth Token")
	flag.StringVar(&awsID, "awsID", "", "AWS account with s3 access")
	flag.StringVar(&awsSecret, "awsSecret", "", "AWS secret with s3 access")
	flag.StringVar(&bucket, "bucket", "", "S3 bucketname")

	flag.Parse()

	if StreamName == "" {
		log.Println("Twitch stream name was not provided")
		os.Exit(1)
	}

	if Token == "" {
		log.Println("Twitch OAuth Token was not provided")
		os.Exit(1)
	}

	if awsID == "" {
		log.Println("awsID was not provided")
		os.Exit(1)
	}

	if awsSecret == "" {
		log.Println("awsSecret was not provided")
		os.Exit(1)
	}

	if bucket == "" {
		log.Println("bucket was not provided")
		os.Exit(1)
	}

	filePath = StreamName + strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
}

func downloadStream(done chan error) {
	commandFinished := make(chan string)
	// Get the home dir location

	// Invoke livestreamer
	cmdString := "echo y |"
	cmdString = cmdString + " livestreamer https://www.twitch.tv/" + StreamName + " best"
	cmdString = cmdString + " --twitch-oauth-token=" + Token
	cmdString = cmdString + " -o=" + filePath
	cmdString = cmdString + " --hls-segment-threads=3"
	cmdString = cmdString + " --yes-run-as-root"
	fmt.Println(cmdString)
	cmd := exec.Command("bash", "-c", cmdString)
	// runs the command and waits for its output
	go func() {
		log.Println("Starting livestreamer")
		cmdOutputByteArray, _ := cmd.Output()
		commandFinished <- string(cmdOutputByteArray)
	}()
	select {
	// 30 seconds passed after command ran
	// livestreamer will not exit will downloading so we'll kill it
	case <-time.After(30 * time.Second):
		log.Println("Stopping livestreamer")
		cmd.Process.Kill()
		cmd.Wait()
		close(done)
	// bash command exited before 30 seconds were up
	case cmdOutput := <-commandFinished:
		log.Println("error occured with livestreamer", cmdOutput)
		close(done)
	}
}

func uploadFile() {
	creds := credentials.NewStaticCredentials(awsID, awsSecret, "")
	_, err := creds.Get()
	if err != nil {
		fmt.Printf("bad credentials: %s", err)
	}
	cfg := aws.NewConfig().WithRegion("us-west-2").WithCredentials(creds)
	svc := s3.New(session.New(), cfg)

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("stream didnt download properly")
		panic(err)
	}
	defer file.Close()
	fileInfo, _ := file.Stat()
	size := fileInfo.Size()
	buffer := make([]byte, size) // read file content to buffer

	file.Read(buffer)
	fileBytes := bytes.NewReader(buffer)
	fileType := http.DetectContentType(buffer)
	path := file.Name()
	params := &s3.PutObjectInput{
		Bucket:        aws.String(bucket),
		Key:           aws.String(path),
		Body:          fileBytes,
		ContentLength: aws.Int64(size),
		ContentType:   aws.String(fileType),
	}
	resp, err := svc.PutObject(params)
	if err != nil {
		fmt.Printf("bad response: %s", err)
	}
	os.Remove(filePath)
	fmt.Printf("response %s", awsutil.StringValue(resp))
}

func main() {
	done := make(chan error)
	downloadStream(done)
	<-done
	uploadFile()
	log.Println("program exited")
}
