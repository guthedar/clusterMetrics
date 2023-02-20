This is used to get the metrics of nodes and pods running in our cluster.

Instructions on how to use this:
1. Build the application : go build
2. Run the application: 
    To get node metrics: ./clusterMetrics -entity pods 
    To get pod metrics:  ./clusterMetrics -entity pods -namespace default


This code uses the "github.com/aquasecurity/table" library to print in table format in terminal.

This uses "github.com/wcharczuk/go-chart/v2" library to print the pie-chart

This uses "gonum.org/v1/plot" library to print the bar graph and histogram