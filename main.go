package main

import (
	"fmt"
	"time"

	"github.com/mmalessa/go-kube-test/kubetools"
	"github.com/spf13/viper"
)

func main() {

	viper.SetConfigName("config.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err.Error())
	}

	kubetools := kubetools.NewKubetools()
	namespace := viper.GetString("namespace")
	serviceName := viper.GetString("servicename")
	servicePort := viper.GetInt32("serviceport")
	localPort := viper.GetInt32("localport")

	kubetools.ForwardServicePort(namespace, serviceName, servicePort, localPort)
	time.Sleep(50 * time.Second)
	fmt.Println("Forward STOP")
}
