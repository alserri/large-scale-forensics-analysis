package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

type Task struct {
	Taskid     string            `json:"taskid"`
	EndpointID string            `json:"endpoint_id"`
	Timestamp  string            `json:"timestamp"`
	Ttype      string            `json:"ttype"`
	Values     map[string]string `json:"values"`
	Done       string            `json:"done"`
	InProgress string            `json:"inprogress"`
}

var todo []Task

var ID []byte

type addselfRequest struct {
	Hostname string `json:"hostname"`
	Os       string `json:"os"`
}

func init() {
	_, err := os.Stat("config.yaml")
	if os.IsNotExist(err) {
		filehandler, err := os.OpenFile("config.yaml", os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			fmt.Println(err)
		}

		ostype := runtime.GOOS
		hostname, _ := os.Hostname()

		// if ostype != "darwin" && ostype != "linux" {

		// }
		requestBody := addselfRequest{}
		requestBody.Hostname = hostname
		requestBody.Os = ostype
		b, err := json.Marshal(requestBody)
		if err != nil {
			fmt.Printf("Error: %s", err)

		}
		res, err := http.Post(
			"http://forensicsserver:8000/api/addEndPoint",
			"application/json; charset=UTF-8",
			bytes.NewBuffer(b),
		)
		if err != nil {
			fmt.Printf("Error: %s", err)

		}
		defer res.Body.Close()
		resp, _ := ioutil.ReadAll(res.Body)
		filehandler.Write(resp)
	} else {
		fmt.Println("I am already added ")
		ID, _ = ioutil.ReadFile("config.yaml")
	}
}

func getNewTask() {

	for {
		var tasks []Task
		resp, err := http.Post("http://forensicsserver:8000/api/task", "text/plain", bytes.NewBuffer(ID))
		if err != nil {
			fmt.Printf("Error: %s", err)

		}
		res, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("Error: %s", err)

		}
		fmt.Println("checking for new tasks ")
		defer resp.Body.Close()
		json.Unmarshal(res, &tasks)
		fmt.Println(string(res))
		for _, item := range tasks {

			tid := item.Taskid
			found := 0
			for _, todoTask := range todo {
				if todoTask.Taskid == tid {
					found = 1
				}

			}
			if found == 0 {
				fmt.Printf("Found undone task %v", item)
				todo = append(todo, item)

			}

		}

		// for _, item := range todo {
		// 	fmt.Println(item)
		// }

		time.Sleep(time.Second * 5)

	}
}

func do() {

	for {
		//  loop through the received tasks from the server
		// compare each taks type and run it's specific function in a new thread
		for i, item := range todo {
			// wg := sync.WaitGroup{}

			// if item.Ttype == "searchbyname" {
			// 	go searhByName(item)
			// 	wg.Add(1)

			// }
			if item.Ttype == "kape" {
				if item.InProgress != "true" {
					go runKape(item)
					todo[i].InProgress = "true"
				}

			}
			if item.Ttype == "command" {
				//	go run executeCommand()
			}
			//wg.Wait()
			time.Sleep(time.Second * 1)

		}

		time.Sleep(time.Second * 6)

	}

}

func searhByName(t Task) {

	path := t.Values["path"]
	fielname := t.Values["name"]
	fmt.Println("found  searchbyname task ")
	fmt.Println("path>>", t.Values["path"])

	err := filepath.Walk(path,

		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			fmt.Println("file>>>>>>", info.Name())
			if info.Name() == fielname {
				fmt.Println(path, info.Size())
			}
			return nil
		})
	if err == nil {
		log.Println("file not found")
	}
	tellServerTaskFinished(t.Taskid)
	removeTaskFromTodo(t.Taskid)

}

func tellServerTaskFinished(taskid string) {

	fmt.Printf("Finishing task %s\n\n", taskid)

	_, err := http.Post("http://forensicsserver:8000/api/finishtask", "text/plain", bytes.NewBuffer([]byte(taskid)))
	if err != nil {
		fmt.Printf("Error: %s", err)

	}
	fmt.Println("Task ended successfully")

}

func removeTaskFromTodo(taskid string) {
	var pos int
	if len(todo) == 1 {
		todo = make([]Task, 0)
		return
	}
	for index, item := range todo {
		if item.Taskid == taskid {
			pos = index
		}
	}

	todo = append(todo[:pos], todo[pos+1:]...)

}

func runKape(t Task) {

	//--msource C:\\ --mdest C:\\kape-output\\prefetch --mef json --module RECmd
	fmt.Println("Received kape task")
	args := strings.Split(t.Values["command"], " ")
	fmt.Printf("command --> %s", args)
	cmd := exec.Command("C:\\KAPE\\kape.exe ", args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(out))

	tellServerTaskFinished(t.Taskid)
	removeTaskFromTodo(t.Taskid)
}

func finishtask(id string) {

	// call finishtask at the server using post and pass the id to be marked as done
	// resp, err := http.Post("http://forensicsserver:8000/api/finishtask", "text/plain", bytes.NewBuffer(ID))
	// if err != nil {
	// 	fmt.Printf("Error: %s", err)

	// }

}

func main() {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go getNewTask()
	go do()

	wg.Wait()

}

// fucntion to periodically check to do for new tasks
// if new tasks were found call the matching type function to do the task
// for example found searchbyname task, call searchbyname fucntction to do the tasks
// then call the finish task to mark it as done at the server

// function to search by name
// take name and path as values and search for the file
// if file exsits return true if not return flase

//function to perioidcally (for example 1 min) call Iamalive function
// send endpoint id in the body
// server side receives the id and update lastactive of the endpoint based on its id
