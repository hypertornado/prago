package prago

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

var defaultPort = 8585

func addServerCommand(app *App) {
	var port int
	var developmentMode bool
	app.addCommand("run").
		Description("runs app and blocks").
		flag(
			newCommandFlag("port", "port of server").Alias("p").Int(&port),
		).
		flag(
			newCommandFlag("development", "development mode").Alias("d").Bool(&developmentMode),
		).
		Callback(func() {
			app.developmentMode = developmentMode
			if port <= 0 {
				configPort := app.mustGetSetting(context.Background(), "port")
				var err error
				port, err = strconv.Atoi(configPort)
				if err != nil {
					panic("wrong format of 'port' entry in config file, should be int")
				}

			}
			must(app.listenAndServe(port))
		})
}

type flagType int

const (
	noType flagType = iota
	stringFlag
	boolFlag
	intFlag
)

type commands struct {
	commands []*command
}

type command struct {
	actions        []string
	description    string
	flags          map[string]*commandFlag
	callback       func()
	stringArgument *string
}

// CommandFlag represents command-line flag
type commandFlag struct {
	name        string
	description string
	typ         flagType
	aliases     []string
	value       interface{}
}

// AddCommand adds command to app
func (app *App) addCommand(commands ...string) *command {
	ret := &command{
		actions: commands,
		flags:   map[string]*commandFlag{},
	}
	app.commands.commands = append(app.commands.commands, ret)
	return ret
}

// Callback sets command callback
func (c *command) Callback(callback func()) *command {
	c.callback = callback
	return c
}

// Description sets description to command
func (c *command) Description(description string) *command {
	c.description = description
	return c
}

// Flag adds flag to command
func (c *command) flag(flag *commandFlag) *command {
	if flag.typ == noType {
		panic("no type of flag set")
	}
	c.flags[flag.name] = flag
	for _, v := range flag.aliases {
		c.flags[v] = flag
	}
	return c
}

// StringArgument sets command argument
func (c *command) StringArgument(arg *string) *command {
	c.stringArgument = arg
	return c
}

// NewCommandFlag creates new command flag
func newCommandFlag(name, description string) *commandFlag {
	return &commandFlag{
		name:        name,
		description: description,
	}
}

// Alias sets flag alias
func (f *commandFlag) Alias(alias string) *commandFlag {
	f.aliases = append(f.aliases, alias)
	return f
}

// String sets flag type to string
func (f *commandFlag) String(value *string) *commandFlag {
	if f.typ != noType {
		panic("type of flag already set")
	}
	f.typ = stringFlag
	f.value = value
	return f
}

// Int sets flag type to int
func (f *commandFlag) Int(value *int) *commandFlag {
	if f.typ != noType {
		panic("type of flag already set")
	}
	f.typ = intFlag
	f.value = value
	return f
}

// Bool sets flag type to boolean
func (f *commandFlag) Bool(value *bool) *commandFlag {
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
	}

	return true, nil
}

func parseFlags(flags map[string]*commandFlag, fields []string) (argsX []string, err error) {
	var currentFlag *commandFlag
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

func (f *commandFlag) setValue(value string) error {
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
	addServerCommand(app)
	args := os.Args[1:]
	for _, command := range app.commands.commands {
		matched, err := command.match(args)
		if matched {
			if err != nil {
				fmt.Println(err)
				app.usage()
			} else {
				app.commands = nil
				command.callback()
			}
			return
		}
	}
	fmt.Println("no command found")
	app.usage()
}

func (app *App) usage() {
	fmt.Printf("%s, version %s, usage:\n", app.codeName, app.version)
	for _, v := range app.commands.commands {
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
