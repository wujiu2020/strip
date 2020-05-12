package config

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

const (
	defaultSectionName = "default"
)

type Config struct {
	FileName   string
	FileReader io.Reader
}

type Configer interface {
	Find(name string) string
	ListSection(section string) map[string]string

	SetParent(Configer)
}

type config struct {
	Config

	data map[string]map[string]string // Section -> key : value

	parent Configer
}

var Global Configer = &config{
	data: make(map[string]map[string]string),
}

func NewConfiger(cfg Config) (Configer, error) {
	c := new(config)
	c.Config = cfg
	c.data = make(map[string]map[string]string)

	if c.FileName != "" {
		err := c.loadFile(c.FileName)
		if err != nil {
			return nil, err
		}
	}

	if c.FileReader != nil {
		err := c.loadFromReader(c.FileReader)
		if err != nil {
			return nil, err
		}
	}
	return c, nil
}

func LoadIniFile(fileName string) (Configer, error) {
	return NewConfiger(Config{
		FileName: fileName,
	})
}

func LoadIniFromReader(r io.Reader) (Configer, error) {
	return NewConfiger(Config{
		FileReader: r,
	})
}

func (c *config) SetParent(parent Configer) {
	c.parent = parent
}

func (c *config) ListSection(section string) map[string]string {
	data, ok := c.data[section]
	if !ok {
		return nil
	}
	values := make(map[string]string, len(data))
	for k, v := range data {
		values[k] = v
	}
	return values
}

func (c *config) Find(name string) (value string) {
	var section, key string
	parts := strings.SplitN(name, "::", 2)
	if len(parts) > 0 {
		key = parts[len(parts)-1]
		if len(parts) > 1 {
			section = parts[0]
		}
	}

	if key == "" {
		return
	}

	envKey := convertToEnvKey(section, key)
	value = os.Getenv(envKey)
	if value != "" {
		return
	}

	defer func() {
		// if current not exist fall back to parent
		if value == "" && c.parent != nil {
			value = c.parent.Find(name)
		}
	}()

	if section == "" {
		section = defaultSectionName
	}

	if _, ok := c.data[section]; !ok {
		// Section does not exist.
		return
	}

	value = c.data[section][key]
	return
}

func (c *config) setValue(section, key, value string) {
	if section == "" {
		section = defaultSectionName
	}

	// Check if section exists.
	if _, ok := c.data[section]; !ok {
		// Execute add operation.
		c.data[section] = make(map[string]string)
	}

	if key == "" {
		return
	}

	c.data[section][key] = value
}

func (c *config) loadFile(fileName string) error {
	f, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer f.Close()
	return c.loadFromReader(f)
}

func (c *config) loadFromReader(r io.Reader) error {
	buf := bufio.NewReader(r)

	// Handle BOM-UTF8.
	// http://en.wikipedia.org/wiki/Byte_order_mark#Representations_of_byte_order_marks_by_encoding
	mask, err := buf.Peek(3)
	if err == nil && len(mask) >= 3 &&
		mask[0] == 239 && mask[1] == 187 && mask[2] == 191 {
		buf.Read(mask)
	}

	count := 1 // Counter for auto increment.

	// Current section name.
	section := ""

	// Parse line-by-line
	for {
		line, err := buf.ReadString('\n')
		line = strings.TrimSpace(line)
		lineLengh := len(line) //[SWH|+]
		if err != nil {
			if err != io.EOF {
				return err
			}

			// Reached end of file, if nothing to read then break,
			// otherwise handle the last line.
			if lineLengh == 0 {
				break
			}
		}

		// switch written for readability (not performance)
		switch {
		case lineLengh == 0: // Empty line
			continue

		case line[0] == '#' || line[0] == ';': // Comment
			continue

		case line[0] == '[' && line[lineLengh-1] == ']': // New sction.
			// Get section name.
			section = strings.TrimSpace(line[1 : lineLengh-1])

			// No section defined so far
			if section == "" {
				return readError{ERR_BLANK_SECTION_NAME, line}
			}

			// Make section exist even though it does not have any key.
			c.setValue(section, "", "")
			// Reset counter.
			count = 1
			continue

		default: // Other alternatives
			var (
				i        int
				keyQuote string
				key      string
				valQuote string
				value    string
			)

			//[SWH|+]:支持引号包围起来的字串
			if line[0] == '"' {
				if lineLengh >= 6 && line[0:3] == `"""` {
					keyQuote = `"""`
				} else {
					keyQuote = `"`
				}
			} else if line[0] == '`' {
				keyQuote = "`"
			}

			if keyQuote != "" {
				qLen := len(keyQuote)
				pos := strings.Index(line[qLen:], keyQuote)
				if pos == -1 {
					return readError{ERR_COULD_NOT_PARSE, line}
				}

				pos = pos + qLen
				i = strings.IndexAny(line[pos:], "=:")
				if i <= 0 {
					return readError{ERR_COULD_NOT_PARSE, line}
				}

				i = i + pos
				key = line[qLen:pos] //保留引号内的两端的空格

			} else {
				i = strings.IndexAny(line, "=:")
				if i <= 0 {
					return readError{ERR_COULD_NOT_PARSE, line}
				}
				key = strings.TrimSpace(line[0:i])
			}
			//[SWH|+];

			// Check if it needs auto increment.
			if key == "-" {
				key = "#" + fmt.Sprint(count)
				count++
			}

			//[SWH|+]:支持引号包围起来的字串
			lineRight := strings.TrimSpace(line[i+1:])
			lineRightLength := len(lineRight)
			firstChar := ""
			if lineRightLength >= 2 {
				firstChar = lineRight[0:1]
			}

			if firstChar == "`" {
				valQuote = "`"
			} else if lineRightLength >= 6 && lineRight[0:3] == `"""` {
				valQuote = `"""`
			}

			if valQuote != "" {
				qLen := len(valQuote)
				pos := strings.LastIndex(lineRight[qLen:], valQuote)
				if pos == -1 {
					return readError{ERR_COULD_NOT_PARSE, line}
				}
				pos = pos + qLen
				value = lineRight[qLen:pos]
			} else {
				value = strings.TrimSpace(lineRight[0:])
			}
			//[SWH|+];

			c.setValue(section, key, value)
		}

		// Reached end of file.
		if err == io.EOF {
			break
		}
	}
	return nil
}

type ParseError int

const (
	ERR_SECTION_NOT_FOUND ParseError = iota + 1
	ERR_KEY_NOT_FOUND
	ERR_BLANK_SECTION_NAME
	ERR_COULD_NOT_PARSE
)

// readError occurs when read configuration file with wrong format.
type readError struct {
	Reason  ParseError
	Content string // Line content
}

// Error implement Error interface.
func (err readError) Error() string {
	switch err.Reason {
	case ERR_BLANK_SECTION_NAME:
		return "empty section name not allowed"
	case ERR_COULD_NOT_PARSE:
		return fmt.Sprintf("could not parse line: %s", string(err.Content))
	}
	return "invalid read error"
}

var envKeyRepalcer = strings.NewReplacer(
	".", "_",
	"-", "_",
	":", "_",
	" ", "_",
)

func convertToEnvKey(section, key string) string {
	if section != "" {
		section += "_"
	}
	return strings.ToUpper(envKeyRepalcer.Replace(section + key))
}
