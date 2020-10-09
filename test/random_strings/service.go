package random_strings

import (
	"github.com/brunvieira/lotus"
	"github.com/brunvieira/lotus/test/contract"
)

var (
	RandomStringService = lotus.Service{
		ServiceContract: &contract.RandomStringsServiceContract,
	}
)

func init() {
	RandomStringService.SetupRoute(contract.RandomStringsRouteContract.Label, generateRandomStrings, nil, nil)
	go RandomStringService.Start()
}
