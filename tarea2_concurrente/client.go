package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
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

var result []I_res

func main() {
	arguments := os.Args
	if len(arguments) == 1 {
		fmt.Println("Please provide host:port.")
		return
	}

	CONNECT := arguments[1]
	c, err := net.Dial("tcp", CONNECT)
	if err != nil {
		fmt.Println(err)
		return
	}

	reader := bufio.NewReader(os.Stdin)
	reader_con := bufio.NewReader(c)
	fmt.Print(">> ")
	text, _ := reader.ReadString('\n')
	fmt.Fprintf(c, text)

	data, _ := reader_con.ReadString('\n')
	var dataset []Instance
	json.Unmarshal([]byte(data), &dataset)

	n, _ := reader_con.ReadString('\n')
	k, _ := strconv.Atoi(strings.TrimSpace(n))

	for {
		it, _ := reader_con.ReadString('\n')
		var new_it Instance
		json.Unmarshal([]byte(it), &new_it)

		for i := 0; i < len(dataset); i++ {
			dist := math.Sqrt(math.Pow(dataset[i].Pelvic_incidence-new_it.Pelvic_incidence, 2) +
				math.Pow(dataset[i].Pelvic_tilt-new_it.Pelvic_tilt, 2) +
				math.Pow(dataset[i].Lumbar_lordosis_angle-new_it.Lumbar_lordosis_angle, 2) +
				math.Pow(dataset[i].Sacral_slope-new_it.Sacral_slope, 2) +
				math.Pow(dataset[i].Pelvic_radius-new_it.Pelvic_radius, 2) +
				math.Pow(dataset[i].Degree_spondylolisthesis-new_it.Degree_spondylolisthesis, 2))
			result = append(result, I_res{
				Distance: dist,
				Class:    dataset[i].Class,
			})
		}
		sort.SliceStable(result, func(i, j int) bool {
			return result[i].Distance < result[j].Distance
		})
		send_res, _ := json.Marshal(result[0:k])
		send_res_str := string(send_res) + "\n"
		fmt.Fprintf(c, send_res_str)
		result = nil
	}
}
