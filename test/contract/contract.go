package contract

import "github.com/brunvieira/lotus"

var Contract = lotus.Contract{
	Services: []lotus.ServiceContract{
		EchoServiceContract,
	},
}
