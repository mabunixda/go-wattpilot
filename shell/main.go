package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"

	api "github.com/mabunixda/wattpilot"
)

type InputFunc func(*api.Wattpilot, []string)

var inputs = map[string]InputFunc{
	"connect":    inConnect,
	"status":     inStatus,
	"get":        inGetValue,
	"set":        inSetValue,
	"disconnect": inDisconnect,
	"properties": inProperties,
	"dump":       dumpData,
	"log":        setLevel,
	"update":     inUpdateStatus,
}

func setLevel(w *api.Wattpilot, data []string) {
	if len(data) == 0 {
		return
	}
	if err := w.ParseLogLevel(data[0]); err != nil {
		fmt.Println("error on parsing: ", err)
	}
}

func inUpdateStatus(w *api.Wattpilot, data []string) {
	if err := w.RequestStatusUpdate(); err != nil {
		fmt.Println("error on update: ", err)
	}
}

func inStatus(w *api.Wattpilot, data []string) {
	w.StatusInfo()
	fmt.Println("")
}

func inGetValue(w *api.Wattpilot, data []string) {
	if len(data) == 0 {
		return
	}
	v, err := w.GetProperty(data[0])
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(v)
}
func inSetValue(w *api.Wattpilot, data []string) {
	if len(data) <= 1 {
		return
	}
	err := w.SetProperty(data[0], data[1])
	if err == nil {
		return
	}
	fmt.Println("error:", err)
}

func inProperties(w *api.Wattpilot, data []string) {
	keys := w.Alias()
	for idx := 0; idx < len(keys); idx += 1 {
		alias := keys[idx]
		raw := w.LookupAlias(alias)
		value, _ := w.GetProperty(alias)
		fmt.Printf("- %s: %s\n  %v\n", alias, raw, value)
	}
}

func remove[T comparable](l []T, item T) []T {
	for i, other := range l {
		if other == item {
			return append(l[:i], l[i+1:]...)
		}
	}
	return l
}

func dumpData(w *api.Wattpilot, data []string) {
	filename := "./wattpilot-data.csv"
	if len(data) > 0 {
		filename = data[0]
	}
	dumpHeader := false
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		dumpHeader = true
	}
	csvFile, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Could not create file: ", err)
		return
	}

	keys := remove(w.Properties(), "wsm")
	sort.Strings(keys)

	writer := csv.NewWriter(csvFile)
	if dumpHeader {
		if err := writer.Write(keys); err != nil {
			log.Fatalln("Could not create csv file dump")
			return
		}
	}
	dataSet := []string{}
	for idx := 0; idx < len(keys); idx += 1 {
		alias := keys[idx]
		value, _ := w.GetProperty(alias)
		dataSet = append(dataSet, fmt.Sprint(value))
	}
	if err := writer.Write(dataSet); err != nil {
		log.Fatalln("error writing csv-data:", err)
		return
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		log.Fatalln("error writing csv:", err)
		return
	}
	log.Println("export written to ", filename)
}

func inConnect(w *api.Wattpilot, data []string) {
	err := w.Connect()
	if err != nil {
		log.Println("Could not connect", err)
		return
	}
	log.Printf("Connected to WattPilot %s, Serial %s", w.GetName(), w.GetSerial())
}

func inDisconnect(w *api.Wattpilot, data []string) {
	w.Disconnect()
}

// func processUpdates(ups <-chan interface{}) {
// 	updates = ups
// }

var interrupt chan os.Signal

var updates <-chan interface{}

func main() {
	host := os.Getenv("WATTPILOT_HOST")
	pwd := os.Getenv("WATTPILOT_PASSWORD")
	level := os.Getenv("WATTPILOT_LOG")
	if host == "" || pwd == "" {
		return
	}
	if level == "" {
		level = "WARN"
	}
	w := api.New(host, pwd)
	if err := w.ParseLogLevel(level); err != nil {
		fmt.Println("Could not update loglevel to", level, err)
	}
	// just a sample to test notification updates
	// processUpdates(w.GetNotifications("fhz"))
	inConnect(w, nil)

	w.StatusInfo()

	for {
		select {

		case <-updates:
			fmt.Println(<-updates)
			break

		case <-interrupt:
			w.Disconnect()
			return

		default:
			fmt.Print("wattpilot> ")
			reader := bufio.NewReader(os.Stdin)
			str, err := reader.ReadString('\n')
			if err != nil {
				log.Fatal(err)
			}
			words := strings.Fields(str)
			if len(words) < 1 {
				continue
			}

			data := words[1:]
			cmd := words[:1]
			if _, fOk := inputs[cmd[0]]; !fOk {
				fmt.Println("Could not find command", cmd[0])
				continue
			}
			inputs[cmd[0]](w, data)
			fmt.Println("")
		}
	}
}
