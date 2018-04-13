package prago

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var defaultPort = 8585

func initCommands(app *App) {
	var port int
	var developmentMode bool
	app.AddCommand("server").
		Flag(
			NewFlag("port", "port of server").Alias("p").Int(&port),
		).
		Flag(
			NewFlag("development", "development mode").Alias("d").Bool(&developmentMode),
		).
		Callback(func() {
			if port <= 0 {
				configPort, err := app.Config.Get("port")
				switch configPort.(type) {
				case string:
					port, err = strconv.Atoi(configPort.(string))
					if err != nil {
						app.Log().Fatalf("wrong format of 'port' entry in config file, should be int")
					}
				case float64:
					port = int(configPort.(float64))
				default:
					port = defaultPort
				}
			}
			must(app.ListenAndServe(port, developmentMode))
		})
}

type flagType int

const (
	noType flagType = iota
	stringFlag
	boolFlag
	intFlag
)

type command struct {
	actions        []string
	description    string
	flags          map[string]*flag
	callback       func()
	stringArgument *string
}

type flag struct {
	name        string
	description string
	typ         flagType
	aliases     []string
	value       interface{}
}

func (app *App) AddCommand(commands ...string) *command {
	ret := &command{
		actions: commands,
		flags:   map[string]*flag{},
	}
	app.commands = append(app.commands, ret)
	return ret
}

func (c *command) Callback(callback func()) *command {
	c.callback = callback
	return c
}

func (c *command) Description(description string) *command {
	c.description = description
	return c
}

func (c *command) Flag(flag *flag) *command {
	if flag.typ == noType {
		panic("no type of flag set")
	}
	c.flags[flag.name] = flag
	for _, v := range flag.aliases {
		c.flags[v] = flag
	}
	return c
}

func (c *command) StringArgument(arg *string) *command {
	c.stringArgument = arg
	return c
}

func NewFlag(name, description string) *flag {
	return &flag{
		name:        name,
		description: description,
	}
}

func (f *flag) Alias(alias string) *flag {
	f.aliases = append(f.aliases, alias)
	return f
}

func (f *flag) String(value *string) *flag {
	if f.typ != noType {
		panic("type of flag already set")
	}
	f.typ = stringFlag
	f.value = value
	return f
}

func (f *flag) Int(value *int) *flag {
	if f.typ != noType {
		panic("type of flag already set")
	}
	f.typ = intFlag
	f.value = value
	return f
}

func (f *flag) Bool(value *bool) *flag {
	if f.typ != noType {
		panic("type of flag already set")
	}
	f.typ = boolFlag
	f.value = value
	return f
}

func (c *command) match(fields []string) (bool, error) {
	if len(c.actions) > len(fields) {
		return false, nil
	}

	for i := 0; i < len(c.actions); i++ {
		if c.actions[i] != fields[i] {
			return false, nil
		}
	}

	args, err := parseFlags(c.flags, fields[len(c.actions):])
	if c.stringArgument != nil && len(args) == 1 {
		reflect.ValueOf(c.stringArgument).Elem().SetString(args[0])
		return true, nil
	}

	if err != nil {
		return true, fmt.Errorf("error while parsing command '%s': %s", strings.Join(c.actions, " "), err)
	} else {
		return true, nil
	}
}

func parseFlags(flags map[string]*flag, fields []string) (argsX []string, err error) {
	var currentFlag *flag
	for k, flag := range fields {
		if currentFlag == nil {
			if !strings.HasPrefix(flag, "-") {
				return fields[k:], fmt.Errorf("unknown parameter '%s'", flag)
			}
			flag = strings.TrimPrefix(flag, "-")
			flag = strings.TrimPrefix(flag, "-")
			f := flags[flag]
			if f == nil {
				return fields[k:], fmt.Errorf("unknown flag '%s'", flag)
			}
			if f.typ == boolFlag {
				f.setValue("")
			} else {
				currentFlag = f
			}
		} else {
			err := currentFlag.setValue(flag)
			if err != nil {
				return fields[k-1:], err
			}
			currentFlag = nil
		}
	}
	return nil, nil
}

func (f *flag) setValue(value string) error {
	val := reflect.ValueOf(f.value).Elem()
	switch f.typ {
	case boolFlag:
		val.SetBool(true)
	case stringFlag:
		val.SetString(value)
	case intFlag:
		i, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("expected integer value for flag '%s', got '%s'", f.name, value)
		}
		val.SetInt(int64(i))
	default:
		return fmt.Errorf("unknown flag type")
	}
	return nil
}

func (app *App) parseCommands() {
	args := os.Args[1:]
	for _, command := range app.commands {
		matched, err := command.match(args)
		if matched {
			if err != nil {
				fmt.Println(err)
				app.usage()
			} else {
				command.callback()
			}
			return
		}
	}
	fmt.Println("no command found")
	app.usage()
}

func (app *App) usage() {
	fmt.Printf("%s, version %s, usage:\n", app.AppName, app.Version)
	for _, v := range app.commands {
		fmt.Print("  " + strings.Join(v.actions, " "))
		if len(v.flags) > 0 {
			fmt.Print(" <flags>")
		}
		if v.stringArgument != nil {
			fmt.Print(" <argument>")
		}
		fmt.Println()
		for k, flag := range v.flags {
			//is alias
			if k != flag.name {
				continue
			}

			fmt.Printf("    -%s", flag.name)
			for _, alias := range flag.aliases {
				fmt.Printf(" -%s", alias)
			}
			fmt.Println()
			fmt.Println("      " + flag.description)
		}
	}

}
