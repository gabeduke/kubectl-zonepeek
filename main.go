package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/spf13/pflag"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type PodInfo struct {
	PodName     string
	NodeName    string
	NodeZone    string
	PVInfo      []PVDetails
	ZoneMatched bool
}

type PVDetails struct {
	PVCName string
	PVName  string
	PVZone  string
}

var (
	labelSelector string
	outputFormat  string
)

func main() {
	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	} else {
		kubeconfig = os.Getenv("KUBECONFIG")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	pflag.StringVarP(&labelSelector, "label", "l", "", "Label selector")
	pflag.StringVarP(&outputFormat, "output", "o", "table", "Output format: table, json, text")

	pflag.Parse()

	if labelSelector == "" {
		fmt.Println("Label selector is required")
		os.Exit(1)
	}

	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		LabelSelector: labelSelector,
	})
	if err != nil {
		panic(err.Error())
	}

	var podInfos []PodInfo
	for _, pod := range pods.Items {
		podInfo := PodInfo{
			PodName:  pod.Name,
			NodeName: pod.Spec.NodeName,
		}

		node, err := clientset.CoreV1().Nodes().Get(context.TODO(), pod.Spec.NodeName, metav1.GetOptions{})
		if err != nil {
			panic(err.Error())
		}
		podInfo.NodeZone = node.Labels["topology.kubernetes.io/zone"]

		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil {
				pvc, err := clientset.CoreV1().PersistentVolumeClaims(pod.Namespace).Get(context.TODO(), volume.PersistentVolumeClaim.ClaimName, metav1.GetOptions{})
				if err != nil {
					panic(err.Error())
				}

				pv, err := clientset.CoreV1().PersistentVolumes().Get(context.TODO(), pvc.Spec.VolumeName, metav1.GetOptions{})
				if err != nil {
					panic(err.Error())
				}

				pvDetails := PVDetails{
					PVCName: volume.PersistentVolumeClaim.ClaimName,
					PVName:  pvc.Spec.VolumeName,
					PVZone:  pv.Labels["topology.kubernetes.io/zone"],
				}
				podInfo.PVInfo = append(podInfo.PVInfo, pvDetails)

				if pvDetails.PVZone == podInfo.NodeZone {
					podInfo.ZoneMatched = true
				}
			}
		}

		podInfos = append(podInfos, podInfo)
	}

	switch outputFormat {
	case "json":
		printJSON(podInfos)
	case "text":
		printText(podInfos)
	default:
		printTable(podInfos)
	}
}

func printJSON(podInfos []PodInfo) {
	data, err := json.MarshalIndent(podInfos, "", "  ")
	if err != nil {
		panic(err.Error())
	}
	fmt.Println(string(data))
}

func printText(podInfos []PodInfo) {
	for _, info := range podInfos {
		fmt.Printf("Pod: %s, Node: %s, Node Zone: %s, Zone Matched: %t\n", info.PodName, info.NodeName, info.NodeZone, info.ZoneMatched)
	}
}

func printTable(podInfos []PodInfo) {
	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 0, 8, 0, '\t', 0)
	_, err := fmt.Fprintln(w, "POD\tNODE\tNODE ZONE\tZONE MATCHED\tPVC\tPV\tPV ZONE")
	if err != nil {
		return
	}

	for _, info := range podInfos {
		for _, pvInfo := range info.PVInfo {
			_, err := fmt.Fprintf(w, "%s\t%s\t%s\t%t\t%s\t%s\t%s\n", info.PodName, info.NodeName, info.NodeZone, info.ZoneMatched, pvInfo.PVCName, pvInfo.PVName, pvInfo.PVZone)
			if err != nil {
				return
			}
		}
	}
	w.Flush()
}
