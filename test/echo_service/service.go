package echo_service

import (
	"github.com/brunvieira/lotus"
	"github.com/brunvieira/lotus/test/contract"
)

var EchoService = lotus.Service{ServiceContract: &contract.EchoServiceContract}

func init() {
	EchoService.SetupRoute(contract.SimpleEchoRouteContract.Label, echo, nil, nil)
	EchoService.SetupRoute(contract.PostEchoRouteContract.Label, randomString, nil, nil)
	go EchoService.Start()
}
