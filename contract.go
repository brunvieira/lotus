package lotus

type Contract struct {
	Services []ServiceContract
}

func (c *Contract) serviceContract(label string) *ServiceContract {
	for _, d := range c.Services {
		if d.Label == label {
			return &d
		}
	}
	return nil
}