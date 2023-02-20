package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/wcharczuk/go-chart/v2"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
)

type PodMetricsList struct {
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Metadata   struct {
		SelfLink string `json:"selfLink"`
	} `json:"metadata"`
	Items []struct {
		Metadata struct {
			Name              string    `json:"name"`
			Namespace         string    `json:"namespace"`
			SelfLink          string    `json:"selfLink"`
			CreationTimestamp time.Time `json:"creationTimestamp"`
		} `json:"metadata"`
		Timestamp  time.Time `json:"timestamp"`
		Window     string    `json:"window"`
		Containers []struct {
			Name  string `json:"name"`
			Usage struct {
				CPU    string `json:"cpu"`
				Memory string `json:"memory"`
			} `json:"usage"`
		} `json:"containers"`
	} `json:"items"`
}

type podGraph struct {
	nodeName         string
	containerName    string
	podName          string
	containerCpu     float64
	cpuPercentage    float64
	containerMemory  float64
	memoryPercentage float64
}

type MyBox struct {
	Graph []podGraph
}
type nodeGraph struct {
	nodeName             string
	nodeCpuPercentage    float64
	nodeMemory           float64
	nodeMemoryPercentage float64
	noOfPods             int
}

type MyNode struct {
	Graph []nodeGraph
}

func clientSetup() *kubernetes.Clientset {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
	config, err := kubeconfig.ClientConfig()
	if err != nil {
		log.Printf("Error in new client config: %s\n", err)
	}
	clientset := kubernetes.NewForConfigOrDie(config)
	return clientset

}
func clientMetricsSetup() *metrics.Clientset {
	rules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(rules, &clientcmd.ConfigOverrides{})
	config, err := kubeconfig.ClientConfig()
	if err != nil {
		log.Printf("Error in new client config: %s\n", err)
	}
	clientsetMetrics, err := metrics.NewForConfig(config)
	if err != nil {
		log.Printf("Error in new client metrics config: %s\n", err)
	}
	return clientsetMetrics

}

func GetNodeCapacity(nodeName string) (int64, *resource.Quantity) {
	clientset := clientSetup()
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})

	if err != nil {
		fmt.Println(err)
	}
	for _, v := range nodes.Items {

		if v.Name == nodeName {
			return v.Status.Capacity.Cpu().MilliValue(), v.Status.Capacity.Memory()
		}
	}
	return -1, nil

}
func GetNodeName(podName string, namespace string) (nodeName string) {
	clientset := clientSetup()
	podList, err := clientset.CoreV1().Pods(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Printf("Error in getting pods list for namespace %s: %s\n", namespace, err)
	}
	for _, n := range podList.Items {
		if n.Name == podName {
			return n.Spec.NodeName
		}
	}
	return ""
}
func GetNodeMetrics() [][]string {
	var data [][]string
	clientset := clientSetup()
	clientsetMetrics := clientMetricsSetup()
	nodeMetricsList, err := clientsetMetrics.MetricsV1beta1().NodeMetricses().List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Printf("Error in getting nodes list : %s\n", err)
	}
	for _, n := range nodeMetricsList.Items {
		nodeName := n.Name

		cpu, memory := GetNodeCapacity(nodeName)
		cpuFloat := float64(cpu * 1000.0)
		memoryValue := n.Usage.Memory()
		memoryFloat := float64(memory.Value() / (1024 * 1024))
		nodeCpuUsage := float64(n.Usage.Cpu().MilliValue()) * 1000
		nodeMemoryUsage := float64(n.Usage.Memory().Value()) / (1024.0 * 1024)
		pods, _ := clientset.CoreV1().Pods("").List(context.Background(), metav1.ListOptions{
			FieldSelector: "spec.nodeName=" + nodeName,
		})
		numberPods := len(pods.Items)
		data = append(data, []string{nodeName, fmt.Sprintf("%v", cpu), fmt.Sprintf("%v", n.Usage.Cpu().MilliValue()), fmt.Sprintf("%v", (nodeCpuUsage/(cpuFloat))*100), fmt.Sprintf("%v", memory), fmt.Sprintf("%v", memoryValue), fmt.Sprintf("%v", nodeMemoryUsage/(memoryFloat)*100), fmt.Sprintf("%v", numberPods)})

	}
	return data

}
func GetPodMetrics(namespace string) ([][]string, MyBox) {
	clientsetMetrics := clientMetricsSetup()
	podMetricsList, err := clientsetMetrics.MetricsV1beta1().PodMetricses(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Printf("Error in getting pods list for namespace %s: %s\n", namespace, err)
	}

	graphAr := make([]podGraph, 0, 0)
	var data [][]string
	for _, n := range podMetricsList.Items {
		nodeName := GetNodeName(n.Name, namespace)
		cpu, memory := GetNodeCapacity(nodeName)
		cpuFloat := float64(cpu)
		memoryFloat := float64(memory.Value() / (1024 * 1024))
		for _, container := range n.Containers {
			var gr podGraph
			containerCpuFloat := float64(container.Usage.Cpu().MilliValue())
			containerMemoryFloat := float64(container.Usage.Memory().Value()) / (1024.0 * 1024.0)
			data = append(data, []string{nodeName, n.Name, container.Name, fmt.Sprintf("%v", containerCpuFloat), fmt.Sprintf("%v", (containerCpuFloat/cpuFloat)*100), fmt.Sprintf("%v", containerMemoryFloat), fmt.Sprintf("%v", (containerMemoryFloat/memoryFloat)*100)})

			//containerCPU, _ := strconv.ParseFloat(fmt.Sprintf("%v", containerCpuFloat), 32)
			containerMEM, _ := strconv.ParseFloat(fmt.Sprintf("%v", containerMemoryFloat), 32)
			//containerCPUPer, _ := strconv.ParseFloat(fmt.Sprintf("%v", (containerCpuFloat/cpuFloat)*100), 32)
			containerMEMPer, _ := strconv.ParseFloat(fmt.Sprintf("%v", (containerMemoryFloat/memoryFloat)*100), 32)

			gr = podGraph{
				nodeName:         nodeName,
				podName:          n.Name,
				containerName:    container.Name,
				containerCpu:     4,
				containerMemory:  containerMEM,
				cpuPercentage:    0.04,
				memoryPercentage: containerMEMPer,
			}
			graphAr = append(graphAr, gr)
		}
	}
	box := MyBox{Graph: graphAr}
	return data, box

}

//this will generate the Pie charts
func GenerateGraphPods(namespace string) [][]string {
	start := time.Now()
	fmt.Println("Generating metrics.............")
	_, graphValues := GetPodMetrics(namespace)
	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Println("Time taken to get the metrics: ", elapsed)
	var cpuPie []chart.Value
	var memoryPie []chart.Value
	var data [][]string
	//colors := [...]string{"Blue", "Green", "Orange", "Gray", "Yellow", "Cyan", "White", "Black", "LightGray", "AlternateLightGray", "AlternateGreen", "AlternateBlue", "AlternateYellow", "AlternateGray"}
	for _, v := range graphValues.Graph {

		podContainerName := v.podName + "\n" + v.containerName
		cpuPie = append(cpuPie, chart.Value{Value: v.cpuPercentage, Label: podContainerName + fmt.Sprintf("%v", v.cpuPercentage)})
		memoryPie = append(memoryPie, chart.Value{Value: v.memoryPercentage, Label: podContainerName + fmt.Sprintf("%v", v.memoryPercentage)})
		data = append(data, []string{v.nodeName, v.podName, v.containerName, fmt.Sprintf("%v", v.cpuPercentage), fmt.Sprintf("%v", v.memoryPercentage)})
	}
	cpuPieChart := chart.PieChart{
		Width:  900,
		Height: 1024,
		Values: cpuPie,
		DPI:    40,
		Canvas: chart.Style{FontSize: 10},
	}
	memoryPieChart := chart.PieChart{
		Width:  900,
		Height: 1024,
		Values: memoryPie,
		DPI:    40,
		Canvas: chart.Style{FontSize: 10},
	}

	cpu, _ := os.Create("cpu_" + namespace + ".png")
	memory, _ := os.Create("memory_" + namespace + ".png")
	fmt.Println("Generated metrics")
	defer cpu.Close()
	defer memory.Close()
	cpuPieChart.Render(chart.PNG, cpu)
	memoryPieChart.Render(chart.PNG, memory)
	return data

}

func BarPlot(values plotter.Values) {
	p := plot.New()

	p.Title.Text = "bar plot"

	bar, err := plotter.NewBarChart(values, 15)
	if err != nil {
		panic(err)
	}
	p.Add(bar)

	if err := p.Save(3*vg.Inch, 3*vg.Inch, "bar.png"); err != nil {
		panic(err)
	}
}
func HistPlot(values plotter.Values) {
	p := plot.New()
	p.Title.Text = "histogram plot"

	hist, err := plotter.NewHist(values, 20)
	if err != nil {
		panic(err)
	}
	p.Add(hist)

	if err := p.Save(3*vg.Inch, 3*vg.Inch, "hist.png"); err != nil {
		panic(err)
	}
}
