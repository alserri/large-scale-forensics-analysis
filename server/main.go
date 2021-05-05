// function to update last active by receiving post request
// body contains the endpoint id
// update based on the received endpoint id

//fix escape character in json string \t
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"

	"time"

	"github.com/gorilla/mux"
	"github.com/lithammer/shortuuid"
)

var mySigningKey = []byte("captainjacksparrowsayshi")

// type Endpoint struct {
// 	EndpointID      string `json:"EndpointID"`
// 	elasticsearchip string `json:"elasticsearchip"`
// 	elasticusername string `json:"elasticusername"`
// 	password        string `json:"password"`
// }

type Endpoints struct {
	EndpointID string `json:"endpointID"`
	Hostname   string `json:"hostname"`
	Ip         string `json:"ip"`
	Os         string `json:"os"`
	Lastactive string `json:"lastactive"`
}

type Task struct {
	Taskid     string            `json:"taskid"`
	EndpointID string            `json:"endpoint_id"`
	Timestamp  string            `json:"timestamp"`
	Ttype      string            `json:"ttype"`
	Values     map[string]string `json:"values"`

	Done string `json:"done"`
}

type Uitasks struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

type NewsAggPage struct {
	Title string
	News  string
}

/*func (t Endpoints) tasks() string {
	return
}*/

var uicommands []Uitasks
var eps []Endpoints

// add endpoint
func addEndPoint(w http.ResponseWriter, r *http.Request) {

	id := shortuuid.New()
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		//return nil, fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)

		fmt.Fprintf(w, "userip: %q is not IP:port", r.RemoteAddr)
	}

	file, err := os.OpenFile("endpoints.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println(err)
	}

	var ep Endpoints
	var eps []Endpoints
	err = json.NewDecoder(r.Body).Decode(&ep)
	if err != nil {
		fmt.Println(err)
	}
	now := time.Now()

	ep.EndpointID = id
	ep.Ip = ip
	ep.Lastactive = now.Format("2006-01-02 15:04:05")
	filecontent, _ := ioutil.ReadFile("endpoints.json")

	json.Unmarshal(filecontent, &eps)

	eps = append(eps, ep)
	fmt.Println("New endpoint added", ep.Ip)
	if err != nil {
		fmt.Println("Error marshal add endpiont response")
	}
	data, err := json.MarshalIndent(eps, "", "    ")

	//json.NewEncoder(w).Encode(&ep)
	fmt.Fprintln(w, id)
	//file.Write(data)
	ioutil.WriteFile("endpoints.json", data, 0644)
	defer file.Close()
}

// get all endpoints
func endpoints(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var eps []Endpoints
	endp, err := ioutil.ReadFile("endpoints.json")
	if err != nil {
		fmt.Println(err)
	}
	err = json.Unmarshal(endp, &eps)
	if err != nil {
		fmt.Println(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(&eps)
	//	fmt.Println(string(endp))

}

// get single endpoint
func endpoint(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	content, err := ioutil.ReadFile("endpoints.json")
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(content, &eps)
	if err != nil {
		log.Fatal(err)
	}
	params := mux.Vars(r)
	for _, item := range eps {

		if item.EndpointID == params["id"] {
			//json.NewEncoder(w).Encode(item)
			//fmt.Println(string(content))
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	//json.NewEncoder(w).Encode(&Endpoints{})

}

func addtask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	Tid := shortuuid.New()

	file, _ := os.OpenFile("tasks.json", os.O_CREATE|os.O_RDWR, 0644)
	filecontent, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
	}
	var task Task
	var tasks []Task
	resp, _ := ioutil.ReadAll(r.Body)
	errr := json.Unmarshal(resp, &task)
	if errr != nil {
		fmt.Println(errr)
	}
	task.Taskid = Tid
	task.Timestamp = time.Now().String()

	json.Unmarshal(filecontent, &tasks)
	//
	fmt.Println(task.Values)
	tasks = append(tasks, task)
	dataBytes, err := json.MarshalIndent(tasks, "", "    ")
	if err != nil {
		fmt.Println(err)
	}

	fmt.Fprintln(w, Tid)
	ioutil.WriteFile("tasks.json", dataBytes, 0644)
	defer file.Close()

}

func task(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var tasks []Task
	var return_tasks []Task
	file, err := ioutil.ReadFile("tasks.json")
	if err != nil {
		fmt.Println(err)
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		//return nil, fmt.Errorf("userip: %q is not IP:port", req.RemoteAddr)

		fmt.Fprintf(w, "userip: %q is not IP:port", r.RemoteAddr)
	}
	fmt.Printf("%s is checking for new task\n", ip)

	info, err := os.Stat("tasks.json")

	if err != nil {
		fmt.Print(err)
	}

	filesize := info.Size()
	//fmt.Print(filesize)

	if filesize == 0 {
		json.NewEncoder(w).Encode("no tasks found")
		return
	}

	resp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	defer r.Body.Close()

	id := string(resp)
	errr := json.Unmarshal([]byte(file), &tasks)
	if errr != nil {
		fmt.Println(errr)
	}
	fmt.Println(id)
	// to avoid running the task even if they are already done
	// tasks with done will not be returned to client
	for _, item := range tasks {
		if item.EndpointID == id {
			if item.Done != "true" {
				return_tasks = append(return_tasks, item)
			}
		}
	}
	json.NewEncoder(w).Encode(&return_tasks)
}

// get all tasks
func alltasks(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var tasks []Task

	task, err := ioutil.ReadFile("tasks.json")
	if err != nil {
		fmt.Println(err)
	}

	info, err := os.Stat("tasks.json")

	if err != nil {
		fmt.Print(err)
	}

	filesize := info.Size()
	fmt.Print(filesize)

	if filesize == 0 {

		return
	}

	err = json.Unmarshal(task, &tasks)
	if err != nil {
		fmt.Println(err)
	}
	if err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(&tasks)
	//	fmt.Println(string(endp))

}

func finishtask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var tasks []Task

	file, err := ioutil.ReadFile("tasks.json")
	if err != nil {
		fmt.Println(err)
	}

	resp, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Println(err)
	}
	defer r.Body.Close()

	taskid := string(resp)
	errr := json.Unmarshal([]byte(file), &tasks)
	if errr != nil {
		fmt.Println(errr)
	}
	var itemindex int
	for i, item := range tasks {
		if item.Taskid == taskid {
			itemindex = i
		}
	}
	tasks[itemindex].Done = "true"
	dataBytes, err := json.MarshalIndent(tasks, "", "    ")
	if err != nil {
		fmt.Println(err)
	}

	ioutil.WriteFile("tasks.json", dataBytes, 0644)
	fmt.Printf("Task terminated %s\n", taskid)
	fmt.Fprintln(w, "Success")

}

// get task by endpoint id
func getendptask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var tasks []Task

	file, err := ioutil.ReadFile("tasks.json")
	if err != nil {
		fmt.Println(err)
	}

	info, err := os.Stat("tasks.json")

	if err != nil {
		fmt.Print(err)
	}

	filesize := info.Size()
	fmt.Print(filesize)

	if filesize == 0 {
		return
	}

	err = json.Unmarshal(file, &tasks)
	if err != nil {
		log.Fatal(err)
	}
	var items []Task
	params := mux.Vars(r)
	for _, item := range tasks {

		if item.EndpointID == params["id"] {
			items = append(items, item)

		}

	}
	json.NewEncoder(w).Encode(items)
	return

}

func UItasks(w http.ResponseWriter, r *http.Request) {
	// w.Header().Set("Content-Type", "application/json")
	// var tasks string
	// // tasks =

	// // json.NewEncoder(w).Encode(tasks)

	// uitasks, err := ioutil.ReadFile("Uitasks.json")
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// err = json.Unmarshal(uitasks, &uicommands)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// json.NewEncoder(w).Encode(&uicommands)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(uicommands)

}

func getuitask(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var uicommands []Uitasks

	resp, _ := ioutil.ReadAll(r.Body)
	errr := json.Unmarshal(resp, &uicommands)
	if errr != nil {
		fmt.Println(errr)
	}
}

func update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

}

func agent(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

}

// update
func changepassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

}

func beacon(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

}
func logout(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

}
func kaperequest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

}

func serveTemplate(w http.ResponseWriter, r *http.Request) {

	data := ""
	t2, err := template.ParseFiles("static/index.html")
	if err != nil {
		fmt.Println(err)
	}

	err = t2.Execute(w, data)
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	router := mux.NewRouter()

	// fs := http.FileServer(http.Dir("./static"))
	// http.Handle("/static/", http.StripPrefix("/static/", fs))

	uicommands = append(uicommands, Uitasks{Name: "NetStat", Command: "--mdest C:\\kape-output\\prefetch --module NetStat --mef json"})
	uicommands = append(uicommands, Uitasks{Name: "WxTCmd", Command: "--msource C:\\ --mdest C:\\kape-output\\live_response --mef json --module WxTCmd"})
	uicommands = append(uicommands, Uitasks{Name: "SBECmd", Command: "--msource C:\\ --mdest C:\\kape-output\\live_response --mef json --module SBECmd"})
	uicommands = append(uicommands, Uitasks{Name: "SystemInfo", Command: "--msource C:\\ --mdest C:\\kape-output\\live_response --mef json --module SystemInfo"})
	uicommands = append(uicommands, Uitasks{Name: "browsinghistoryview", Command: "--msource C:\\ --mdest C:\\kape-output\\live_response --mef json --module browsinghistoryview"})

	filesystem := http.FileServer(http.Dir("./static"))
	router.PathPrefix("/static").Handler(http.StripPrefix("/static", filesystem))

	router.HandleFunc("/", serveTemplate).Methods("GET")

	router.HandleFunc("/api/addEndPoint", addEndPoint).Methods("POST")

	router.HandleFunc("/api/endpoints", endpoints).Methods("GET")
	router.HandleFunc("/api/endpoint/{id}", endpoint).Methods("GET")
	router.HandleFunc("/api/getuitask/{", endpoint).Methods("GET")

	router.HandleFunc("/api/uitasks", UItasks).Methods("GET")

	router.HandleFunc("/api/task", task).Methods("POST")

	router.HandleFunc("/api/addtask", addtask).Methods("POST")
	router.HandleFunc("/api/getendptask/{id}", getendptask).Methods("GET")

	router.HandleFunc("/api/alltasks", alltasks).Methods("GET")

	router.HandleFunc("/api/finishtask", finishtask).Methods("POST")

	router.HandleFunc("/api/update", update).Methods("POST")

	router.HandleFunc("/windows", agent).Methods("GET")
	router.HandleFunc("/api/changepassword ", changepassword).Methods("PUT")
	router.HandleFunc("/api/task/addtask", addtask).Methods("POST")
	router.HandleFunc("/api/beacon", beacon).Methods("POST")
	router.HandleFunc("/logout", logout).Methods("POST")
	router.HandleFunc("/artifacts/kaperequest", kaperequest).Methods("POST")
	fmt.Println("server up and running")
	http.ListenAndServe(":8000", router)

}

//var auth []authen

/*func GenerateJWT() (string, error) {
	token := jwt.New(jwt.SigningMethodHS256)

	claims := token.Claims.(jwt.MapClaims)

	claims["authorized"] = true
	claims["client"] = "Elliot Forbes"
	claims["exp"] = time.Now().Add(time.Minute * 30).Unix()

	tokenString, err := token.SignedString(mySigningKey)

	if err != nil {
		fmt.Errorf("Something Went Wrong: %s", err.Error())
		return "", err
	}

	return tokenString, nil
}*/

/*func connect(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(eps)
}*/
