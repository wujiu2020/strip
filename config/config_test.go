package config

import (
	"os"
	"testing"

	"github.com/teapots/teapot"
)

func Test_Config(t *testing.T) {
	assert := &teapot.Assert{T: t}

	c, err := NewConfiger(Config{
		FileName: "testdata/conf.ini",
	})
	assert.NoError(err)

	assert.True(c.Find("app.name") == "google")

	assert.True(c.Find("key1") == "")

	assert.True(c.Find("Demo::key1") == "Let's us goconfig!! goconfig is good!!")
	assert.True(c.Find("Demo::key2") == "test data")
	assert.True(c.Find("Demo::quote") == "\"special case for quote")

	assert.True(c.Find("Demo::key:1") == "This is the value of \"key:1\"")
	assert.True(c.Find("Demo::中国") == "China")

	assert.True(c.Find("What's this?::name") == "try one more value ^-^")

	os.Setenv("APP_NAME", "apple")
	assert.True(c.Find("app.name") == "apple")
	os.Unsetenv("APP_NAME")

	os.Setenv("DEMO_KEY1", "apple")
	assert.True(c.Find("Demo::key1") == "apple")
	os.Unsetenv("DEMO_KEY1")
}

func Test_ListSection(t *testing.T) {
	assert := &teapot.Assert{T: t}

	c, err := LoadIniFile("testdata/conf.ini")
	assert.NoError(err)

	m := c.ListSection("section")
	assert.True(len(m) == 3)
	assert.True(m["a"] == "1")
	assert.True(m["b"] == "2")
	assert.True(m["c"] == "3")
}

func Test_ConfigParent(t *testing.T) {
	assert := &teapot.Assert{T: t}

	def, err := NewConfiger(Config{
		FileName: "testdata/conf.ini",
	})
	assert.NoError(err)

	c, err := LoadIniFile("testdata/conf_prod.ini")
	assert.NoError(err)

	c.SetParent(def)

	assert.True(c.Find("app.name") == "qiniu")

	assert.True(c.Find("Demo::quote") == "just no quote")
	assert.True(c.Find("What's this?::name") == "just a name")

	// global config
	assert.True(c.Find("key1") == "")

	assert.True(c.Find("Demo::key1") == "Let's us goconfig!! goconfig is good!!")
	assert.True(c.Find("Demo::key2") == "test data")

	assert.True(c.Find("Demo::key:1") == "This is the value of \"key:1\"")
	assert.True(c.Find("Demo::中国") == "China")

	os.Setenv("APP_NAME", "apple")
	assert.True(c.Find("app.name") == "apple")
	os.Unsetenv("APP_NAME")

	os.Setenv("DEMO_KEY1", "apple")
	assert.True(c.Find("Demo::key1") == "apple")
	os.Unsetenv("DEMO_KEY1")
}
