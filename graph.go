package main

import (
	"path"

	"go.uber.org/zap"
	"net/http"
	"html/template"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"github.com/awalterschulze/gographviz"
	"strings"
)

var log *zap.SugaredLogger
var kubeClient *kubernetes.Clientset
const port = "8080"

type DotGraph struct {
	Name string
	Dot string
	DebugDot []string
}

func configureLogger() {
	zapper, _ := zap.NewDevelopment()
	defer zapper.Sync()
	// flushes buffer, if any
	log = zapper.Sugar()
}

func renderTemplate(w http.ResponseWriter, tmpl string, graph *DotGraph) {
	t, _ := template.ParseFiles(path.Join("templates", tmpl + ".html"))
	log.Infow("Serving page", "template", tmpl)
	replace := strings.Replace(graph.Dot, "\t", "  ", -1)
	replace = strings.Replace(replace, "\n\n", "\n", -1)
	graph.DebugDot = strings.Split(replace, "\n")
	t.Execute(w, graph)
}

func loadService(serviceName string) (*DotGraph, error) {
	g := gographviz.NewEscape()
	g.Name = serviceName
	g.SetDir(true)
	g.AddNode(serviceName, "0", nil)
	g.AddNode(serviceName, "1", nil)
	g.AddEdge("0", "1", true, nil)
	graphDot := g.String()
	log.Infow("Loaded service meta-data", "serviceName", serviceName)
	return &DotGraph{Name: serviceName, Dot: graphDot}, nil
}

func viewHandler(w http.ResponseWriter, r *http.Request) {
	serviceName := r.URL.Path[len("/view/"):]
	g, err := loadService(serviceName)
	if err != nil {
		log.Errorw("Could build graph", "error", err)
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "view", g)
}

func loadAllPods() (*DotGraph, error) {

	namespaces, err := kubeClient.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		log.Errorw("could not retrieve namespaces", "error", err)
	}

	graph := gographviz.NewEscape()
	graph.Name = "pods"
	attrs := make(map[string]string)
	attrs[string(gographviz.RankDir)] = "LR"
	ggvAttrs, _ := gographviz.NewAttrs(attrs)
	graph.Attrs.Extend(ggvAttrs)

	for _, ns := range namespaces.Items {
		pods, err := kubeClient.CoreV1().Pods(ns.Name).List(metav1.ListOptions{})
		if err != nil {
			log.Errorw("could not retrieve pods", "namespace", ns.Name, "error", err)
			continue
		}

		nsSubgraph := gographviz.NewSubGraph("cluster_" + ns.Name)
		graph.AddSubGraph("pods", nsSubgraph.Name, nil)
		if err = graph.AddAttr(nsSubgraph.Name, string(gographviz.Label), ns.Name); err != nil {
			log.Errorw("failed to set subgraph label", "error", err)
		}


		for _, pod := range pods.Items {
			graph.AddNode(nsSubgraph.Name, pod.Name, nil)
		}
		log.Debugw("Subgraph added", "subGraphName", nsSubgraph.Name)
	}

	graphDot := graph.String()
	log.Debugw("Marshalled all pods into dot format", "dot", string(graphDot[:]))
	return &DotGraph{Name: "Pods", Dot: graphDot}, nil
}

func podsHandler(w http.ResponseWriter, r *http.Request) {
	//namespace := r.URL.Path[len("/pods/"):]
	g, err := loadAllPods()
	if err != nil {
		log.Errorw("Could not load pods", "error", err)
		http.NotFound(w, r)
		return
	}
	renderTemplate(w, "pods", g)
}


func main() {
	configureLogger()
	kubeClient = KubernetesClient()
	http.HandleFunc("/view/", viewHandler)
	http.HandleFunc("/pods/", podsHandler)
	log.Fatal(http.ListenAndServe(":" + port, nil))
}

