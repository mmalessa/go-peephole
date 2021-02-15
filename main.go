package main

import (
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

	kubetools.FindPodsInService(namespace, serviceName)

}