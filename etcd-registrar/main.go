package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/flaviostutz/etcd-registry/etcd-registry"
	"github.com/sirupsen/logrus"
)

func main() {
	logLevel := flag.String("loglevel", "debug", "debug, info, warning, error")
	etcdURL0 := flag.String("etcd-url", "", "ETCD URLs. ex: http://etcd0:2379")
	etcdBase0 := flag.String("etcd-base", "/services", "Base ETCD path. Defaults to '/services'")
	service0 := flag.String("service", "", "Service name. Ex: ServiceA")
	list0 := flag.Bool("list", false, "If true, will return a list of service nodes registered in ETCD")
	name0 := flag.String("name", "", "Node name. Maybe an IP or a custom name. Ex.: node345")
	info0 := flag.String("info", "", "Additional node info in json format")
	ttl0 := flag.Int("ttl", 60, "Time to Live. The daemon will keep updating the node's lease until it is killed")
	flag.Parse()

	etcdURL := *etcdURL0
	etcdBase := *etcdBase0
	list := *list0
	service := *service0
	name := *name0
	ttl := *ttl0
	info := *info0

	if etcdURL == "" {
		showUsage()
		panic("--etcd-url should be defined")
	}
	if service == "" {
		showUsage()
		panic("--service should be defined")
	}
	if !list && name == "" {
		showUsage()
		panic("--name should be defined")
	}

	logrus.Infof("Registering service node at %s/%s/%s [service=%s, name=%s, ttl=%d, info=%s]. etcdUrl=%s", etcdBase, service, name, service, name, ttl, info, etcdURL)

	switch *logLevel {
	case "debug":
		logrus.SetLevel(logrus.DebugLevel)
		break
	case "warning":
		logrus.SetLevel(logrus.WarnLevel)
		break
	case "error":
		logrus.SetLevel(logrus.ErrorLevel)
		break
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	etcdEndpoints := strings.Split(etcdURL, ",")
	reg, err := etcdregistry.NewEtcdRegistry(etcdEndpoints, etcdBase, 10*time.Second)
	if err != nil {
		panic(err)
	}

	if list {
		nodes, err := reg.GetServiceNodes(service)
		logrus.Debugf("%s %s", nodes, err)
		if err != nil {
			panic(err)
		}
		for _, n := range nodes {
			fmt.Sprintf("%s;%s;%s", service, n.Name, n.Info)
		}

	} else {
		node := etcdregistry.Node{}
		node.Name = name
		infom := make(map[string]string, 0)
		if info != "" {
			err = json.Unmarshal([]byte(info), &infom)
			if err != nil {
				logrus.Errorf("Could not parse 'info' as json content. err=%s", err)
				panic(err)
			}
		}
		node.Info = infom
		err := reg.RegisterNode(context.TODO(), service, node, 20*time.Second)
		if err != nil {
			panic(err)
		}
	}
}

func showUsage() {
	fmt.Printf("This utility maintains a TTL based service registry, so that service nodes can register themselves if they desapear, its registration will vanish. This daemon will keep the node alive on ETCD until it is killed")
	fmt.Printf("")
	fmt.Printf("For service node registration:")
	fmt.Printf("etcd-registrar --etcd-url=[ETCD URL] --etcd-base=[ETCD BASE] --service=[SERVICE NAME] --name=[NODE NAME] --ttl=[TTL IN SECONDS] --info=[NODE INFO JSON]")
	fmt.Printf(
		`Sample:
    etcd-registrar --etcd-url=http://etcd0:2375 --service=Service123 --name=node5 --ttl=60 --info='{address:172.17.1.23, weight:4}'
`)
}
