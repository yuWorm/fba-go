package pool

import pond "github.com/alitto/pond/v2"

type Options struct {
	MaxWorkers int
	QueueSize  int
}

type Provider struct {
	pools map[string]pond.Pool
}

func NewProvider(config map[string]Options) *Provider {
	provider := &Provider{pools: make(map[string]pond.Pool, len(config))}
	for name, opts := range config {
		maxWorkers := opts.MaxWorkers
		if maxWorkers <= 0 {
			maxWorkers = 1
		}
		options := make([]pond.Option, 0, 1)
		if opts.QueueSize > 0 {
			options = append(options, pond.WithQueueSize(opts.QueueSize))
		}
		provider.pools[name] = pond.NewPool(maxWorkers, options...)
	}
	return provider
}

func (p *Provider) Pool(name string) (pond.Pool, bool) {
	workerPool, ok := p.pools[name]
	return workerPool, ok
}

func (p *Provider) Stop() {
	for _, workerPool := range p.pools {
		workerPool.StopAndWait()
	}
}
