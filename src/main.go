package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/hajimehoshi/go-mp3"
	"github.com/hajimehoshi/oto/v2"
)

type Config struct {
	LatestLogPath string
	SoundPath     string
	TargetWords   []string
}

func main() {
	config := getConfig("./config.json")
	for {
		if checkFileExists(config.LatestLogPath) {
			break
		} else {
			time.Sleep(2 * time.Second)
		}
	}

	for {
		watchFile(config.LatestLogPath)
		if isPingRequired(getLatestLog(config.LatestLogPath), config.TargetWords) {
			playSound(config.SoundPath)
		}
	}
}

func getConfig(filePath string) Config {
	content, err := os.ReadFile(filePath)
	isErrorOccurred(err)
	var config Config
	err = json.Unmarshal(content, &config)
	isErrorOccurred(err)
	return Config{config.LatestLogPath, config.SoundPath, config.TargetWords}
	//return Config{"../logs/latest.log", "./sound.mp3", []string{"nickname", "ping"}, true}
}

func getLatestLog(filePath string) string { //find file latest.log if exists and returns last line
	logs, err := os.ReadFile(filePath)
	isErrorOccurred(err)
	logsLines := SplitLines(string(logs))
	return logsLines[len(logsLines)-1]
}

func isPingRequired(line string, targetWords []string) bool {
	for _, element := range strings.Fields(line) {
		for _, targetWord := range targetWords {
			if element == targetWord {
				return true
			}
		}
	}
	return false
}

func SplitLines(s string) []string {
	var lines []string
	sc := bufio.NewScanner(strings.NewReader(s))
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines
}

func playSound(filePath string) {
	f, err := os.Open(filePath)
	isErrorOccurred(err)
	defer f.Close()

	d, err := mp3.NewDecoder(f)
	isErrorOccurred(err)

	c, ready, err := oto.NewContext(d.SampleRate(), 2, 2)
	isErrorOccurred(err)
	<-ready

	p := c.NewPlayer(d)
	defer p.Close()
	p.Play()

	for {
		time.Sleep(time.Second)
		if !p.IsPlaying() {
			break
		}
	}

}

func isErrorOccurred(err error) {
	if err != nil {
		panic(err)
	}
}

func checkFileExists(filePath string) bool {
	_, error := os.Stat(filePath)
	return !errors.Is(error, os.ErrNotExist)
}

func watchFile(filePath string) error {
	initialStat, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	for {
		stat, err := os.Stat(filePath)
		if err != nil {
			return err
		}

		if stat.Size() != initialStat.Size() || stat.ModTime() != initialStat.ModTime() {
			break
		}

		time.Sleep(1 * time.Second)
	}

	return nil
}
