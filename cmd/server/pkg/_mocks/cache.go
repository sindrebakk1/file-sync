package _mocks

// MockCache is a mock implementation of the cache.Cache interface for testing purposes.
type MockCache struct{}

func (c *MockCache) Get(_ string) (interface{}, bool) {
	return nil, false
}

func (c *MockCache) Set(_ string, _ interface{}) {
}

func (c *MockCache) Delete(_ string) {
}
