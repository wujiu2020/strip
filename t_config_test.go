package strip

import (
	"testing"
)

type testConfig map[string]string

func (t testConfig) Find(name string) string {
	return t[name]
}

func Test_ConfigFind(t *testing.T) {
	assert := &Assert{T: t}

	test := testConfig{
		"run_mode": "test",
	}
	config := newConfig()
	config.setParent(test)

	assert.True(config.Find("run_mode") == "test")
	assert.True(config.Find("no_value") == "")
	assert.True(config.Find("no_value", "default") == "default")
}

func Test_ConfigBind(t *testing.T) {
	assert := &Assert{T: t}

	test := testConfig{
		"run_mode": "test",
		"year":     "2014",
		"toggle":
			"1",
		"array":    "1,2,",
	}
	config := newConfig()
	config.setParent(test)

	var (
		mode   Mode
		year   int
		toggle *bool
		no     *bool
		array  []string
	)
	config.Bind(&mode, "run_mode")
	assert.True(mode.IsTest())

	config.Bind(&year, "year")
	assert.True(year == 2014)

	config.Bind(&array, "array")
	assert.True(len(array) == 2)
	assert.True(array[0] == "1")
	assert.True(array[1] == "2")

	config.Bind(&toggle, "toggle")
	config.Bind(&no, "no")
	assert.NotNil(toggle)
	assert.True(*toggle)
	assert.Nil(no)
}

func Test_DefaultConfig(t *testing.T) {
	assert := &Assert{T: t}

	var (
		mode Mode
		addr string
		port string
	)
	tea := New().Routers()
	tea.Routers(Get(func(config *Config) {
		mode = config.RunMode
		addr = config.HttpAddr
		port = config.HttpPort
	}))

	assert.True(routeFound(tea, "GET", "/"))
	assert.True(mode == tea.Config.RunMode)
	assert.True(addr == tea.Config.HttpAddr)
	assert.True(port == tea.Config.HttpPort)
}

func Test_ImportConfig(t *testing.T) {
	assert := &Assert{T: t}

	var (
		mode Mode
		addr string
		port string
	)
	tea := New().Routers()
	tea.Routers(Get(func(config *Config) {
		mode = config.RunMode
		addr = config.HttpAddr
		port = config.HttpPort
	}))

	c := testConfig{
		"run_mode":  "test",
		"http_addr": "1.1.1.1",
		"http_port": "8080",
	}
	tea.ImportConfig(c)

	assert.True(routeFound(tea, "GET", "/"))
	assert.True(mode.IsTest())
	assert.True(addr == "1.1.1.1")
	assert.True(port == "8080")
}
