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
    "math/rand"
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

//vertex type
const RAIL_SWITCH = 1
const STATION = 2

//repair vehicle parameters
const REPAIR_VEHICLE_SPEED = 150.0
const REPAIR_VEHICLE_NAME = "Repair Vehicle"
const STATION_VERTEX = 12
const RAIL_SWITCH_REPAIR_TIME_H = 2
const RAILWAY_REPAIR_TIME_H = 2 
const TRAIN_REPAIR_TIME_H = 2

//consts used by repair vehicle
const RAIL_SWITCH_REPAIR = 1
const RAILWAY_REPAIR = 2
const TRAIN_REPAIR = 3

//crash rate
//0.0 - 1.0
const CRASH_RATE = 0.2


/* Global variables */

//true if railway system is broken
var crash_active = false

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

type train struct {
    name            string
    people          int
    capacity        int 
    speed           float64 //max speed in kmh
    path            []int
    current_strech  []int
    broken          bool
    repaired        chan bool
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
    vertex_index    int
}

 //type of vertex
type rail_switch struct {
    wait_time       float64   //minutes to switch
    is_free         chan bool //true inside if no1 train using atm
    rotating        chan bool //
    rotate_done     chan bool //
    vertex_index    int
}

type repair_vehicle struct {
    name                string
    speed               float64 //max speed in kmh
    path                []int
    STATION_VERTEX      int
    rail_switch_crash   chan int
    railway_crash       chan int
    train_crash         chan int
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
            vertex_index,_ := strconv.Atoi(tokens[4])
            for i:=0; i<platforms; i++ {
                free_platforms <- true
            }
            for i:=0; i<depots; i++ {
                free_depots <- true
            }
            stations[j] = station{name:name, free_platforms:free_platforms, free_depots:free_depots, wait_time: wait_time, vertex_index:vertex_index}
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
            current_strech := make([]int, 2)
            repaired := make(chan bool, 1)
            for k:=0; k<len(path_string);k++{
                path_int[k],_ = strconv.Atoi(path_string[k])
            }
            trains[j] = train{name:name, capacity:capacity,speed:speed, path: path_int, current_strech:current_strech, repaired:repaired}
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

            if typ == RAIL_SWITCH {
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
            vertex_index,_ := strconv.Atoi(tokens[1])
            is_free := make(chan bool, 1)
            is_free <- true
            rotating := make(chan bool, 1)
            rotate_done := make(chan bool, 1)

            
            rail_switches[j] = rail_switch{wait_time: time, is_free:is_free, rotating:rotating, rotate_done:rotate_done, vertex_index:vertex_index}
            j++
        }
        i++
    }

    return system, stations, trains, vertex_set, rail_switches
}


// Dijkstra's algorithm to find shortest path from s to destin
func dijkstra(G [][]railway, src int, destin int) []int {
    n := len(G)
    dist := make([]int, n)
    pred := make([]int, n)  // preceeding node in path
    visited := make([]bool, n) // all false initially
  
    for i:=0; i<n; i++ {
        dist[i] = 30000 //very big value
    }
    dist[src] = 0
  
    for i:=0; i<n; i++ {
        next := minVertex(dist, visited)
        visited[next] = true

        neighbors := make([]int,0)
        for i:=0;i<n;i++{
            if G[next][i].length > 0{
                neighbors = append(neighbors,i)
            }
        }

        for j:=0; j<len(neighbors); j++ {
            v := neighbors[j];
            d := dist[next] + int(G[next][v].length);
            if dist[v] > d {
                dist[v] = d
                pred[v] = next
            }
        }
    }

    returnPath := make([]int,0)
    for destin != src {
        returnPath = append([]int{destin}, returnPath...)
        destin = pred[destin]
    }
    returnPath = append([]int{src},returnPath...)

    return returnPath
}

//Dijkstra helper function
func minVertex (dist []int, v []bool) int {
    x := 30000 //very big value
    y := -1   // graph not connected, or no unvisited vertices
    for  i:=0; i<len(dist); i++ {
        if !v[i] && dist[i]<x {
            y=i
            x=dist[i]
        }
    }
    return y
}

//reverse array
func reverse(numbers []int) []int {
    for i, j := 0, len(numbers)-1; i < j; i, j = i+1, j-1 {
        numbers[i], numbers[j] = numbers[j], numbers[i]
    }
    return numbers
}


//display logs to terminal and save to file
func logs(file *os.File, line ... string) {
    output := ""
    time := get_current_simulator_time_as_string()
    for i:=0; i<len(line); i++{
        output += line[i] + " "
    }
    output += "\n"
    if !SILENT_MODE {
        fmt.Println(time,"\n   ", output)
    }
    file.WriteString(time + "   " + output)
}


//thread for every rail switch
func start_rail_switch(switch_unit rail_switch) {
    for {
            //wait until some train ask for rotating
            <- switch_unit.rotating
            time.Sleep(time.Millisecond * time.Duration(switch_unit.wait_time * 60000 / TIME_RATE))
            //rotate done, give train permission to continue
            switch_unit.rotate_done <- true
    }
}


//try to broke something sometimes
func crash(repair_vehicle_unit repair_vehicle, trains []train, system [][]railway, rail_switches []rail_switch) {
    for {
        time.Sleep(time.Millisecond * time.Duration(360000/TIME_RATE))

        if rand.Float32() < CRASH_RATE && !crash_active {
            choice := rand.Intn(3)
            switch choice {
                case 0: //crash railway
                    v1 := rand.Intn(len(system))
                    v2 := rand.Intn(len(system))
                    for system[v1][v2].is_free == nil{
                        v1 = rand.Intn(len(system))
                        v2 = rand.Intn(len(system))
                    }
                    logs(nil, "Railway crashed ", strconv.Itoa(v1),"====", strconv.Itoa(v2))
                    crash_active = true
                    //dont allow to use railway by other trains
                    <-system[v1][v2].is_free
                    //send information to repair vehicle about crashed railway
                    repair_vehicle_unit.railway_crash <- v1
                    repair_vehicle_unit.railway_crash <- v2

                case 1: //crash train
                    indx := rand.Intn(len(trains))
                    trains[indx].broken = true
                    crash_active = true
                    logs(nil, "Train",trains[indx].name,"has crashed")
                    repair_vehicle_unit.train_crash <- indx

                case 2: //crash switch
                    indx := rand.Intn(len(rail_switches))
                    crash_active = true
                    logs(nil, "Railswitch crashed at vertex", strconv.Itoa(rail_switches[indx].vertex_index))
                    //dont allow to use rail switch by other trains
                    <- rail_switches[indx].is_free 
                    repair_vehicle_unit.rail_switch_crash <- rail_switches[indx].vertex_index
            }       
        }
    }
}

//initialize the repair vehicle with paremeters
func init_repair_vehicle(repair_vehicle_unit repair_vehicle) repair_vehicle {
    //initialize channels in order to communicate

    repair_vehicle_unit.name = REPAIR_VEHICLE_NAME
    repair_vehicle_unit.speed = REPAIR_VEHICLE_SPEED
    repair_vehicle_unit.STATION_VERTEX = STATION_VERTEX
    repair_vehicle_unit.path = make([]int, 0)
    repair_vehicle_unit.rail_switch_crash = make(chan int,1)
    repair_vehicle_unit.train_crash = make(chan int, 1)
    repair_vehicle_unit.railway_crash = make(chan int, 2)

    return repair_vehicle_unit
}


func send_repair_vehicle(f *os.File, repair_type int, repair_vehicle_unit repair_vehicle, trains []train, system [][]railway, rail_switches []rail_switch, vertex_set []vertex, stations []station){
    for i:=0; i<len(repair_vehicle_unit.path)-1 ;i++{
        start := repair_vehicle_unit.path[i]
        end := repair_vehicle_unit.path[i+1]

        logs(f, "Repair vehicle is now on railway",strconv.Itoa(start),"->",strconv.Itoa(end))

        //count the needed time to travel and wait
        real_world_travel_time_in_ms := get_travel_time(system[start][end].length, repair_vehicle_unit.speed, system[start][end].max_speed)
        in_program_travel_time_in_ms := real_world_travel_time_in_ms / TIME_RATE //fasten time with TIME_RATE multiplier
        time.Sleep(time.Duration(in_program_travel_time_in_ms) * time.Millisecond)

        if vertex_set[end].vertex_type == RAIL_SWITCH {
            logs(f, "Repair vehicle is on railway switch at vertex", strconv.Itoa(end))
        } else {
            logs(f, "Repair vehicle is on station ", stations[vertex_set[end].index].name)
        }
    }
}


func start_repair_vehicle(repair_vehicle_unit repair_vehicle, trains []train, system [][]railway, rail_switches []rail_switch, vertex_set []vertex, stations []station) {

    //log file
    f, _ := os.Create("logs/"+repair_vehicle_unit.name)
    defer f.Close()

    //keep waiting for crash
    for {
        select {
            case train_index := <-repair_vehicle_unit.train_crash:
                logs(f, "Repair vehicle has taken an order to repair train", trains[train_index].name, "at vertex", strconv.Itoa(trains[train_index].current_strech[1]))
                //find path to destination
                repair_vehicle_unit.path = dijkstra(system, repair_vehicle_unit.STATION_VERTEX, trains[train_index].current_strech[1])
                
                send_repair_vehicle(f, TRAIN_REPAIR ,repair_vehicle_unit, trains, system, rail_switches, vertex_set, stations)
                
                //repair
                time.Sleep(time.Duration(TRAIN_REPAIR_TIME_H*3600000/TIME_RATE) * time.Millisecond)
                trains[train_index].repaired <- true
                trains[train_index].broken = false
                logs(f, "Repair vehicle has repaired the train",trains[train_index].name)

                repair_vehicle_unit.path = reverse(repair_vehicle_unit.path)
                send_repair_vehicle(f, -1 ,repair_vehicle_unit, trains, system, rail_switches, vertex_set, stations)

                logs(f, "Repair vehicle has ended its job, returned to station at vertex", strconv.Itoa(repair_vehicle_unit.STATION_VERTEX))
                crash_active = false
                

            case rail_switch_vertex_index := <-repair_vehicle_unit.rail_switch_crash:
                logs(f, "Repair vehicle has taken an order to repair rail switch at vertex", strconv.Itoa(rail_switch_vertex_index))

                //find path to destination
                repair_vehicle_unit.path = dijkstra(system, repair_vehicle_unit.STATION_VERTEX, rail_switch_vertex_index)
                
                send_repair_vehicle(f, RAIL_SWITCH_REPAIR ,repair_vehicle_unit, trains, system, rail_switches, vertex_set, stations)
                
                //repair
                time.Sleep(time.Duration(RAIL_SWITCH_REPAIR_TIME_H*3600000/TIME_RATE) * time.Millisecond)
                rail_switches[vertex_set[rail_switch_vertex_index].index].is_free <- true
                logs(f, "Repair vehicle has repaired rail switch at vertex", strconv.Itoa(rail_switch_vertex_index))

                repair_vehicle_unit.path = reverse(repair_vehicle_unit.path)
                send_repair_vehicle(f, -1 ,repair_vehicle_unit, trains, system, rail_switches, vertex_set, stations)

                logs(f, "Repair vehicle has ended its job, returned to station at vertex", strconv.Itoa(repair_vehicle_unit.STATION_VERTEX))
                crash_active = false



            case railway_index_1 := <-repair_vehicle_unit.railway_crash:
                railway_index_2 := <-repair_vehicle_unit.railway_crash

                logs(f, "Repair vehicle has taken an order to repair railway", strconv.Itoa(railway_index_1),"====",strconv.Itoa(railway_index_2))

                //find path to destination
                repair_vehicle_unit.path = dijkstra(system, repair_vehicle_unit.STATION_VERTEX, railway_index_1)
                
                send_repair_vehicle(f, RAILWAY_REPAIR ,repair_vehicle_unit, trains, system, rail_switches, vertex_set, stations)
                
                //repair
                time.Sleep(time.Duration(RAILWAY_REPAIR_TIME_H*3600000/TIME_RATE) * time.Millisecond)
                system[railway_index_1][railway_index_2].is_free <- true
                logs(f, "Repair vehicle has repaired railway", strconv.Itoa(railway_index_1),"====",strconv.Itoa(railway_index_2))

                repair_vehicle_unit.path = reverse(repair_vehicle_unit.path)
                send_repair_vehicle(f, -1, repair_vehicle_unit, trains, system, rail_switches, vertex_set, stations)

                logs(f, "Repair vehicle has ended its job, returned to station at vertex", strconv.Itoa(repair_vehicle_unit.STATION_VERTEX))
                crash_active = false
        }
    }

}
  


//thread function for every train
func start_train(
    train_unit train,
    system [][]railway,
    stations []station,
    vertex_set []vertex,
    rail_switches []rail_switch) {

    //logs file for every train
    f, _ := os.Create("logs/"+train_unit.name)
    defer f.Close()

    //display logs
    logs(f, train_unit.name, "has started")

    i := 0 //actual path stage
    has_reservation := false

    //start traveling, endless loop
    for{

        //starting and ending vertex
        start := train_unit.path[i]
        end := train_unit.path[(i+1) % len(train_unit.path)]

        if(train_unit.broken){
            <-train_unit.repaired
        }

        train_unit.current_strech[0] = start
        train_unit.current_strech[1] = end

        if !has_reservation{ //if train has reservated this railway before skip waiting for avalibility
            <-system[start][end].is_free 
        }
           
        logs(f, train_unit.name, "is now on railway",strconv.Itoa(start),"->",strconv.Itoa(end))

        //count the needed time to travel and wait
        real_world_travel_time_in_ms := get_travel_time(system[start][end].length, train_unit.speed, system[start][end].max_speed)
        in_program_travel_time_in_ms := real_world_travel_time_in_ms / TIME_RATE //fasten time with TIME_RATE multiplier
        time.Sleep(time.Duration(in_program_travel_time_in_ms) * time.Millisecond)

        if vertex_set[end].vertex_type == RAIL_SWITCH { //arrived to rail switch

            //wait for switch avalibility
            <-rail_switches[vertex_set[end].index].is_free

            //now train can free used railway
            system[start][end].is_free <- true

            //start rotating switch
            rail_switches[vertex_set[end].index].rotating <- true

            logs(f, train_unit.name, "is on railway switch at vertex", strconv.Itoa(end))

            //wait for rotating over
            <- rail_switches[vertex_set[end].index].rotate_done

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
            
            logs(f, train_unit.name, "has arrived to station", stations[vertex_set[end].index].name)

            //count the needed time to wait at platform
            wait_time_in_ms := stations[vertex_set[end].index].wait_time * 60000 / TIME_RATE
            time.Sleep(time.Millisecond * time.Duration(wait_time_in_ms))

            //TODO: get people from platform

            logs(f, train_unit.name, "is ready to leave the station", stations[vertex_set[end].index].name)
            
            //check next railway before leaving station
            next_start := end
            next_end := train_unit.path[(i+2) % len(train_unit.path)]
            <-system[next_start][next_end].is_free
            has_reservation = true 

            stations[vertex_set[end].index].free_platforms <- true 

            logs(f, train_unit.name, "has left the station", stations[vertex_set[end].index].name)

        }

        //next stage
        i = (i+1) % len(train_unit.path)
    }
}


func main() {

    //seed for random values
    rand.Seed(time.Now().UTC().UnixNano())

    //get data from files
    system, stations, trains, vertex_set, rail_switches := read_data(RAILWAYS_PATH, SYSTEM_PATH, TRAINS_PATH, STATIONS_PATH, VERTEX_SET_PATH, SWITCHES_PATH)
    repair_vehicle_unit := init_repair_vehicle(repair_vehicle{})
    
    logs(nil, "Simulator started")

    //I like trains
    for i:=0; i<len(trains);i++ {
        go start_train(trains[i], system, stations, vertex_set, rail_switches)
    }

    //Start switches
    for i:=0; i<len(rail_switches);i++ {
        go start_rail_switch(rail_switches[i])
    }

 
    go start_repair_vehicle(repair_vehicle_unit, trains, system, rail_switches, vertex_set, stations)

    go crash(repair_vehicle_unit, trains, system, rail_switches)


    //wait for user input to end simulator
    fmt.Scanln()
    logs(nil, "Simulator end")
}






