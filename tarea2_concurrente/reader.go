package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
)

type Instance struct {
	Pelvic_incidence         float64
	Pelvic_tilt              float64
	Lumbar_lordosis_angle    float64
	Sacral_slope             float64
	Pelvic_radius            float64
	Degree_spondylolisthesis float64
	Class                    int
}

type I_res struct {
	Distance float64
	Class    int
}

var dataset []Instance
var train []Instance
var test []Instance
var total_con int
var ready int
var acum int
var block bool
var k string
var done int
var result []I_res
var instance_received = false
var obj_str string
var actual int = 1

func readingData() {
	//fmt.Println(strconv.ParseFloat("63.0278175", 64))
	csvFile, _ := os.Open("column.csv")
	reader := csv.NewReader(bufio.NewReader(csvFile))
	for {
		line, error := reader.Read()
		//fmt.Println(line)
		if len(line) == 0 {
			break
		}
		line0, _ := strconv.ParseFloat(line[0], 64)
		line1, _ := strconv.ParseFloat(line[1], 64)
		line2, _ := strconv.ParseFloat(line[2], 64)
		line3, _ := strconv.ParseFloat(line[3], 64)
		line4, _ := strconv.ParseFloat(line[4], 64)
		line5, _ := strconv.ParseFloat(line[5], 64)
		line6, _ := strconv.Atoi(line[6])
		if error == io.EOF {
			break
		} else if error != nil {
			log.Fatal(error)
		}

		dataset = append(dataset, Instance{
			Pelvic_incidence:         line0,
			Pelvic_tilt:              line1,
			Lumbar_lordosis_angle:    line2,
			Sacral_slope:             line3,
			Pelvic_radius:            line4,
			Degree_spondylolisthesis: line5,
			Class:                    line6,
		})
	}
	/*rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(dataset), func(i, j int) {
		dataset[i], dataset[j] = dataset[j], dataset[i]
	})*/
	limit := 0.7 * float64(len(dataset))
	train = dataset[0:int(limit)]
	test = dataset[int(limit):len(dataset)]
	fmt.Println(test)
	fmt.Println("---------------------------")
}

func handle(con net.Conn, number_con int, stopper *sync.Mutex) {
	defer con.Close()
	fmt.Println(strconv.Itoa(number_con) + " cliente conectado...")
	reader := bufio.NewReader(con)
	message, _ := reader.ReadString('\n')
	if strings.TrimSpace(string(message)) == "ready" {
		ready++
		fmt.Println("Cliente " + strconv.Itoa(number_con) + " listo")
	}
	for ready != total_con {

	}
	for number_con != actual {

	}
	stopper.Lock()
	actual++
	tmp := math.Round(float64(number_con)/float64(total_con)*float64(len(train))) - float64(acum)
	arrjson, _ := json.Marshal(train[acum:(acum + int(tmp))])
	newArr := string(arrjson) + "\n"
	acum += int(tmp)
	stopper.Unlock()

	fmt.Fprintf(con, newArr)
	fmt.Fprintf(con, k)

	for {
		for !instance_received {

		}
		trim_obj_str := strings.TrimSpace(obj_str)
		arr_obj_str := strings.Split(trim_obj_str, ",")
		line0, _ := strconv.ParseFloat(arr_obj_str[0], 64)
		line1, _ := strconv.ParseFloat(arr_obj_str[1], 64)
		line2, _ := strconv.ParseFloat(arr_obj_str[2], 64)
		line3, _ := strconv.ParseFloat(arr_obj_str[3], 64)
		line4, _ := strconv.ParseFloat(arr_obj_str[4], 64)
		line5, _ := strconv.ParseFloat(arr_obj_str[5], 64)
		new_it := Instance{
			Pelvic_incidence:         line0,
			Pelvic_tilt:              line1,
			Lumbar_lordosis_angle:    line2,
			Sacral_slope:             line3,
			Pelvic_radius:            line4,
			Degree_spondylolisthesis: line5,
		}
		new_itjson, _ := json.Marshal(new_it)
		new_itsend := string(new_itjson) + "\n"
		fmt.Fprintf(con, new_itsend)
		knn, _ := reader.ReadString('\n')
		var result_obj []I_res
		json.Unmarshal([]byte(knn), &result_obj)
		result = append(result, result_obj...)
		done++
	}
}

func waitFinal() {
	for total_con == 0 || ready < total_con {

	}
	for {
		fmt.Print(">> ")
		it_reader, _ := bufio.NewReader(os.Stdin).ReadString('\n')
		obj_str = it_reader
		instance_received = true
		for done < ready {

		}
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].Distance < result[j].Distance
		})
		fmt.Println(result)
		clase0 := 0
		clase1 := 0
		kint, _ := strconv.Atoi(k)
		for i := 0; i < kint; i++ {
			if result[i].Class == 0 {
				clase0++
			} else {
				clase1++
			}
		}
		if clase0 >= clase1 {
			fmt.Println("Resultado: Columna vertebral en estado normal")
		} else {
			fmt.Println("Resultado: Columna vertebral en estado anormal")
		}
		done = 0
		result = nil
		instance_received = false
	}
}

func main() {
	readingData()
	go waitFinal()
	var stopper sync.Mutex
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide port number")
		return
	}
	hostname := arguments[1]
	k = arguments[2] + "\n"
	l, err := net.Listen("tcp", hostname)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer l.Close()
	for {
		con, _ := l.Accept()
		total_con++
		go handle(con, total_con, &stopper)
	}
}
