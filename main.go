/*
    Railway system v1
    Author: Paweł Okrutny 
    Index: 221478
*/


package main

import (
    "fmt"
    "time"
    "math"    
    "os"
    "bufio"
    "log"
    "strings"
    "strconv"
    //"sync"
)


/*  Constants set  */

//time multiplier
//3600 -> 1hour = 1sec
//60 -> 1hour = 1min
//1 -> 1hour = 1hour (real time speed)
const TIME_RATE = 1000

//silent mode (no terminal output)
const SILENT_MODE = false

//paths of input data files
const SYSTEM_PATH = "input_data/system.txt"
const RAILWAYS_PATH = "input_data/railways.txt"
const TRAINS_PATH = "input_data/trains.txt"
const STATIONS_PATH = "input_data/stations.txt"
const SWITCHES_PATH = "input_data/switches.txt"
const VERTEX_SET_PATH = "input_data/vertex_set.txt"

//start date and time
var start_time = time.Date(2017, 1, 1, 12, 0, 0, 0, time.UTC)
var init_start_time = time.Now()

/* Structure set */

//edge
type railway struct {
    max_speed   float64 //in kmh
    length      float64 //km
    is_free     chan bool
}

//vehicle
type train struct {
    name        string
    people      int
    capacity    int 
    speed       float64 //max speed in kmh
    path        []int
}

type vertex struct {
    vertex_type int //1=switch 2=station
    index       int //position in external array
}

 //type of vertex
type station struct {
    name            string
    free_platforms  chan bool
    free_depots     chan bool
    wait_time       float64 //in minutes
}

 //type of vertex
type rail_switch struct {
    wait_time   float64   //minutes to switch
    is_free     chan bool
}

/* Simulator time functions */

func get_current_simulator_time() time.Time {
    diff := time.Now().Sub(init_start_time)*TIME_RATE
    time_to_return := start_time
    return time_to_return.Add(diff)
}

func get_current_simulator_time_as_string() string {
    return get_current_simulator_time().Format("2006-01-02 15:04")
}

//return travel time in real-world miliseconds
func get_travel_time(km float64, train_kmh float64, rail_max_speed float64) float64{
    speed_in_kmh := math.Min(train_kmh, rail_max_speed)
    travel_time_in_h := km/speed_in_kmh
    return travel_time_in_h*3600000
}

/* Input data reading function */

//Get all data from files
func read_data(
    railways_path string,
    system_path string,
    trains_path string,
    stations_path string,
    vertex_set_path string,
    switches_path string) ([][]railway, []station, []train, []vertex, []rail_switch) {
    
    var railways []railway
    var system [][]railway
    var stations []station
    var trains []train
    var rail_switches []rail_switch
    var vertex_set []vertex

    //Get railways
    file, err := os.Open(railways_path)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    scanner := bufio.NewScanner(file)
    
    i,j := 0,0
    for scanner.Scan() {
        line := scanner.Text()
        switch i{
        case 0:
            n, _ := strconv.Atoi(line)
            railways = make([]railway, n)
        case 1:
        default:
            tokens := strings.Split(line, " ")
            max_speed,_ := strconv.ParseFloat(tokens[0],64)
            length,_ := strconv.ParseFloat(tokens[1],64)
            is_free := make(chan bool, 1)
            is_free <- true
            railways[j] = railway{max_speed: max_speed, length: length, is_free: is_free}
            j++
        }
        i++
    }

    //Get system
    file, err = os.Open(system_path)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    scanner = bufio.NewScanner(file)

    i,j = 0,0
    for scanner.Scan() {
        line := scanner.Text()
        switch i{
        case 0:
            n, _ := strconv.Atoi(line)
            system = make([][]railway, n)
            for i:=0; i<n; i++ {
                system[i] = make([]railway, n)
            }
        case 1:
        default:
            tokens := strings.Split(line, " ")
            vertex1,_ := strconv.Atoi(tokens[0])
            vertex2,_ := strconv.Atoi(tokens[1])

            system[vertex1][vertex2] = railways[j]
            j++
        }
        i++
    }

    //Get stations
    file, err = os.Open(stations_path)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    scanner = bufio.NewScanner(file)

    i,j = 0,0
    for scanner.Scan() {
        line := scanner.Text()
        switch i{
        case 0:
            n, _ := strconv.Atoi(line)
            stations = make([]station, n)
        case 1:
        default:
            tokens := strings.Split(line, " ")
            name := tokens[0]
            platforms,_ := strconv.Atoi(tokens[1])
            depots,_ := strconv.Atoi(tokens[2])
            wait_time,_ := strconv.ParseFloat(tokens[3], 64)
            free_platforms := make(chan bool, platforms)
            free_depots := make(chan bool, depots)
            for i:=0; i<platforms; i++ {
                free_platforms <- true
            }
            for i:=0; i<depots; i++ {
                free_depots <- true
            }
            stations[j] = station{name:name, free_platforms:free_platforms, free_depots:free_depots, wait_time: wait_time}
            j++
        }
        i++
    }

   //Get trains
    file, err = os.Open(trains_path)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    scanner = bufio.NewScanner(file)

    i,j = 0,0
    for scanner.Scan() {
        line := scanner.Text()
        switch i{
        case 0:
            n, _ := strconv.Atoi(line)
            trains = make([]train, n)
        case 1:
        default:
            tokens := strings.Split(line, " ")
            name := tokens[0]
            capacity,_ := strconv.Atoi(tokens[1])
            speed,_ := strconv.ParseFloat(tokens[2],64)
            path_string := strings.Split(tokens[3], "-")
            path_int := make([]int, len(path_string))
            for k:=0; k<len(path_string);k++{
                path_int[k],_ = strconv.Atoi(path_string[k])
            }
            trains[j] = train{name:name, capacity:capacity,speed:speed, path: path_int}
            j++
        }
        i++
    }

    //Get vertex set
    file, err = os.Open(vertex_set_path)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    scanner = bufio.NewScanner(file)

    i,j = 0,0
    switches_count, stations_count := 0,0
    for scanner.Scan() {
        line := scanner.Text()
        switch i{
        case 0:
            n, _ := strconv.Atoi(line)
            vertex_set = make([]vertex, n)
        case 1:
        default:
            tokens := strings.Split(line, " ")
            typ,_ := strconv.Atoi(tokens[0])

            if typ == 1{
                vertex_set[j] = vertex{vertex_type: typ, index:switches_count}
                switches_count++
            } else {
                vertex_set[j] = vertex{vertex_type: typ, index:stations_count}
                stations_count++
            }
            j++
        }
        i++
    }

    //Get rail switches
    file, err = os.Open(switches_path)
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()
    scanner = bufio.NewScanner(file)

    i,j = 0,0
    for scanner.Scan() {
        line := scanner.Text()
        switch i{
        case 0:
            n, _ := strconv.Atoi(line)
            rail_switches = make([]rail_switch, n)
        case 1:
        default:
            tokens := strings.Split(line, " ")
            time,_ := strconv.ParseFloat(tokens[0], 64)
            is_free := make(chan bool, 1)
            is_free <- true
            rail_switches[j] = rail_switch{wait_time: time, is_free:is_free}
            j++
        }
        i++
    }

    return system, stations, trains, vertex_set, rail_switches
}

//display logs to terminal and save to file
func logs(silent_mode bool, file *os.File, line ... string) {
    output := ""
    time := get_current_simulator_time_as_string()
    for i:=0; i<len(line); i++{
        output += line[i] + " "
    }
    output += "\n"
    if !silent_mode {
        fmt.Println(time,"\n   ", output)
    }
    file.WriteString(time + "   " + output)
}

//thread function for every train
func start_train(
    train_unit train,
    system [][]railway,
    stations []station,
    vertex_set []vertex,
    rail_switches []rail_switch,
    silent_mode bool) {

    //logs file for every train
    f, _ := os.Create("logs/"+train_unit.name)
    defer f.Close()

    //display logs
    logs(silent_mode, f, train_unit.name, "has started")

    i := 0 //actual path stage
    has_reservation := false

    //start traveling, endless loop
    for{

        //starting and ending vertex
        start := train_unit.path[i]
        end := train_unit.path[(i+1) % len(train_unit.path)]

        if !has_reservation{ //if train has reservated this railway before skip waiting for avalibility
            <-system[start][end].is_free 
        }
           
        logs(silent_mode, f, train_unit.name, "is now on railway",strconv.Itoa(start),"->",strconv.Itoa(end))

        //count the needed time to travel and wait
        real_world_travel_time_in_ms := get_travel_time(system[start][end].length, train_unit.speed, system[start][end].max_speed)
        in_program_travel_time_in_ms := real_world_travel_time_in_ms / TIME_RATE //fasten time with TIME_RATE multiplier
        time.Sleep(time.Duration(in_program_travel_time_in_ms) * time.Millisecond)

        if vertex_set[end].vertex_type == 1{ //arrived to rail switch

            //wait for switch avalibility
            <-rail_switches[vertex_set[end].index].is_free
            //now train can free used railway
            system[start][end].is_free <- true

            logs(silent_mode, f, train_unit.name, "is on railway switch at vertex", strconv.Itoa(end))

            //wait needed time to switch rails
            wait_time_in_ms := rail_switches[vertex_set[end].index].wait_time * 60000 / TIME_RATE
            time.Sleep(time.Millisecond * time.Duration(wait_time_in_ms))

            //check next railway avalibility before leaving switch
            next_start := end
            next_end := train_unit.path[(i+2) % len(train_unit.path)]
            <-system[next_start][next_end].is_free
            //next railway avalible, train has reservation now
            has_reservation = true 

            //now train can free used switch
            rail_switches[vertex_set[end].index].is_free <- true            

        } else { //arrived to station

            //wait for avalible platform
            <-stations[vertex_set[end].index].free_platforms
            //now train can free used railway
            system[start][end].is_free <- true
            
            logs(silent_mode, f, train_unit.name, "has arrived to station", stations[vertex_set[end].index].name)

            //count the needed time to wait at platform
            wait_time_in_ms := stations[vertex_set[end].index].wait_time * 60000 / TIME_RATE
            time.Sleep(time.Millisecond * time.Duration(wait_time_in_ms))

            //TODO: get people from platform

            logs(silent_mode, f, train_unit.name, "is ready to leave the station", stations[vertex_set[end].index].name)
            
            //check next railway before leaving station
            next_start := end
            next_end := train_unit.path[(i+2) % len(train_unit.path)]
            <-system[next_start][next_end].is_free
            has_reservation = true 

            stations[vertex_set[end].index].free_platforms <- true 

            logs(silent_mode, f, train_unit.name, "has left the station", stations[vertex_set[end].index].name)

        }

        //next stage
        i = (i+1) % len(train_unit.path)
    }
}


func main() {

    //get data from files
    system, stations, trains, vertex_set, rail_switches := read_data(RAILWAYS_PATH, SYSTEM_PATH, TRAINS_PATH, STATIONS_PATH, VERTEX_SET_PATH, SWITCHES_PATH)

    logs(SILENT_MODE, nil, "Simulator started")

    //I like trains
    for i:=0; i<len(trains);i++ {
        go start_train(trains[i], system, stations, vertex_set, rail_switches, SILENT_MODE)
    }

    //wait for user input to end simulator
    fmt.Scanln()
    logs(SILENT_MODE, nil, "Simulator end")
}






