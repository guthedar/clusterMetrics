package main

import (
	"flag"
	"fmt"

	"github.com/aquasecurity/table"

	//graph "github.com/guthedar/clusterMetrics/graph"
	"math"
	"os"
	"time"
)

func main() {

	namespace := flag.String("namespace", "", "Enter Namespace")
	entity := flag.String("entity", "", "Enter pods/ nodes (For nodes only use entity flag) (Required)")
	help := flag.String("h", "", "shows Usage of the command line arguments")

	flag.Parse()
	if *help == "" {
		flag.PrintDefaults()
	}
	if *namespace == "" {
		fmt.Println("Namespace not needed for nodes")
	}
	if *entity == "pods" {
		start := time.Now()
		fmt.Println("Generating metrics.............")
		//data, _ := graph.GetPodMetrics(*namespace)
		data, _ := GetPodMetrics(*namespace)
		end := time.Now()
		elapsed := end.Sub(start)
		fmt.Println("Time taken to get the metrics: ", math.Round(elapsed.Seconds()), "s")
		t := table.New(os.Stdout)

		t.SetHeaders("Node Name", "Pod Name", "Container Name", "CPU value", "CPU Percentage", "Memory Value", "Memory Percentage")
		for _, v := range data {
			t.AddRow(v[0], v[1], v[2], v[3], v[4], v[5], v[6])
		}
		t.Render()

	} else if *entity == "nodes" {
		start := time.Now()
		fmt.Println("Generating metrics.............")
		//data := graph.GetNodeMetrics()
		data := GetNodeMetrics()
		end := time.Now()
		elapsed := end.Sub(start)
		fmt.Println("Time taken to get the metrics: ", math.Round(elapsed.Seconds()), "s")
		t := table.New(os.Stdout)

		t.SetHeaders("Node Name", "CPU Value", "Node CPU usage", "CPU Percentage", "Memory Value", "Node Memory usage", "Memory Percentage", "No of pods")
		for _, v := range data {
			t.AddRow(v[0], v[1], v[2], v[3], v[4], v[5], v[6], v[7])
		}
		t.Render()

	}

	//Generate Pie chart
	GenerateGraphPods(*namespace)

	//Generate Bar Graph
	val := []float64{5, 7, 9, 10}
	BarPlot(val)
	//Generate Histogram
	HistPlot(val)

}
