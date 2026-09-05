package providers

import (
	"fmt"
	"repowipe/types"
)

var registry = map[types.Provider]Provider{}

func Register(p Provider) {
	registry[p.Name()] = p
}

func Get(name types.Provider) (Provider, error) {
	p, ok := registry[name]
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
	return p, nil
}

func All() []Provider {
	out := make([]Provider, 0, len(registry))
	for _, p := range registry {
		out = append(out, p)
	}
	return out
}
