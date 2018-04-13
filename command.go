package prago

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var defaultPort = 8585

func initKingpinCommand(app *App) {
	app.kingpin = kingpin.New("", "")
	app.commands = map[*kingpin.CmdClause]func(app *App){}

	var port int
	var name string
	var truth bool
	var argument string
	app.AddCommand2("server").
		Flag(NewFlag("port", "port of server").Alias("p").Int(&port)).
		Flag(NewFlag("name", "port of server").String(&name)).
		Flag(NewFlag("truth", "port of server").Bool(&truth)).
		StringArgument(&argument).
		Callback(func() {
			fmt.Printf("Data: %d %s %v, argument: %s", port, name, truth, argument)
			println()
		})

	serverCommand := app.CreateCommand("server", "Run server")
	portFlag := serverCommand.Flag("port", "server port").Short('p').Int()
	developmentMode := serverCommand.Flag("development", "Is in development mode").Default("false").Short('d').Bool()

	app.AddCommand(serverCommand, func(app *App) {
		var port = defaultPort
		if portFlag != nil && *portFlag > 0 {
			port = *portFlag
		} else {
			configPort, err := app.Config.Get("port")
			if err == nil {
				port, err = strconv.Atoi(configPort.(string))
				if err != nil {
					app.Log().Fatalf("wrong format of 'port' entry in config file, should be int")
				}
			}
		}
		err := app.ListenAndServe(port, *developmentMode)
		if err != nil {
			app.Log().Fatal(err)
		}
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

type cmdArgument struct {
}

func (app *App) AddCommand2(commands ...string) *command {
	ret := &command{
		actions: commands,
		flags:   map[string]*flag{},
	}
	app.commands2 = append(app.commands2, ret)
	return ret
}

func (c *command) Callback(callback func()) *command {
	c.callback = callback
	return c
}

func (c *command) Flag(flag *flag) *command {
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
				return fields[k:], fmt.Errorf("unknown parameter %s", flag)
			}
			flag = strings.TrimPrefix(flag, "-")
			flag = strings.TrimPrefix(flag, "-")
			f := flags[flag]
			if f == nil {
				return fields[k:], fmt.Errorf("unknown flag %s", flag)
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
			return fmt.Errorf("expected integer value for flag %s, got %s", f.name, value)
		}
		val.SetInt(int64(i))
	default:
		return fmt.Errorf("unknown flag type")
	}
	return nil
}

func (app *App) parseCommands() {
	args := os.Args[1:]
	for _, command := range app.commands2 {
		matched, err := command.match(args)
		if matched {
			if err != nil {
				app.Log().Fatalln(err)
			}
			command.callback()
		}
	}
}
