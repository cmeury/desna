package main

import (
	"path"
	"go.uber.org/zap"
	"net/http"
	"html/template"
	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
	"gonum.org/v1/gonum/graph/encoding/dot"
)

var log *zap.SugaredLogger
const port = "8080"

type ServiceGraph struct {
	ServiceName string
	Graph graph.Graph
	Dot []byte
}

func configureLogger() {
	zapper, _ := zap.NewProduction()
	defer zapper.Sync()
	// flushes buffer, if any
	log = zapper.Sugar()
}

func loadService(serviceName string) (*ServiceGraph, error) {
	g := simple.NewDirectedGraph()
	node0 := simple.Node(0)
	node1 := simple.Node(1)
	g.AddNode(node0)
	g.AddNode(node1)
	g.HasEdgeBetween(0, 1)
	graphDot, err := dot.Marshal(g, "service", "", "  ", false)
	if err != nil {
		log.Errorw("could not render graph into dot format")
		return nil, err
	}
	log.Infow("Loaded service meta-data", "serviceName", serviceName)
	return &ServiceGraph{Graph: g, ServiceName: serviceName, Dot: graphDot}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, graph *ServiceGraph) {
	t, _ := template.ParseFiles(path.Join("templates", tmpl + ".html"))
	log.Infow("Serving page", "template", tmpl, "graph", graph)
	t.Execute(w, graph)
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	serviceName := r.URL.Path[len("/view/"):]
	g, err := loadService(serviceName)
	if err != nil {
		log.Errorw("Could not loadService .dot file", "error", err)
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "view", g)
}


func main() {
	configureLogger()
	http.HandleFunc("/view/", viewHandler)
	log.Fatal(http.ListenAndServe(":" + port, nil))
}

