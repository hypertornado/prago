package prago

import (
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
	app.AddCommand("server").
		Flag(
			NewCommandFlag("port", "port of server").Alias("p").Int(&port),
		).
		Flag(
			NewCommandFlag("development", "development mode").Alias("d").Bool(&developmentMode),
		).
		Callback(func() {
			app.developmentMode = developmentMode
			if port <= 0 {
				configPort, err := app.ConfigurationGetItem("port")
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
			must(app.ListenAndServe(port))
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
	commands []*Command
}

//Command represents command-line command
type Command struct {
	actions        []string
	description    string
	flags          map[string]*CommandFlag
	callback       func()
	stringArgument *string
}

//CommandFlag represents command-line flag
type CommandFlag struct {
	name        string
	description string
	typ         flagType
	aliases     []string
	value       interface{}
}

//AddCommand adds command to app
func (app *App) AddCommand(commands ...string) *Command {
	ret := &Command{
		actions: commands,
		flags:   map[string]*CommandFlag{},
	}
	app.commands.commands = append(app.commands.commands, ret)
	return ret
}

//Callback sets command callback
func (c *Command) Callback(callback func()) *Command {
	c.callback = callback
	return c
}

//Description sets description to command
func (c *Command) Description(description string) *Command {
	c.description = description
	return c
}

//Flag adds flag to command
func (c *Command) Flag(flag *CommandFlag) *Command {
	if flag.typ == noType {
		panic("no type of flag set")
	}
	c.flags[flag.name] = flag
	for _, v := range flag.aliases {
		c.flags[v] = flag
	}
	return c
}

//StringArgument sets command argument
func (c *Command) StringArgument(arg *string) *Command {
	c.stringArgument = arg
	return c
}

//NewCommandFlag creates new command flag
func NewCommandFlag(name, description string) *CommandFlag {
	return &CommandFlag{
		name:        name,
		description: description,
	}
}

//Alias sets flag alias
func (f *CommandFlag) Alias(alias string) *CommandFlag {
	f.aliases = append(f.aliases, alias)
	return f
}

//String sets flag type to string
func (f *CommandFlag) String(value *string) *CommandFlag {
	if f.typ != noType {
		panic("type of flag already set")
	}
	f.typ = stringFlag
	f.value = value
	return f
}

//Int sets flag type to int
func (f *CommandFlag) Int(value *int) *CommandFlag {
	if f.typ != noType {
		panic("type of flag already set")
	}
	f.typ = intFlag
	f.value = value
	return f
}

//Bool sets flag type to boolean
func (f *CommandFlag) Bool(value *bool) *CommandFlag {
	if f.typ != noType {
		panic("type of flag already set")
	}
	f.typ = boolFlag
	f.value = value
	return f
}

func (c *Command) match(fields []string) (bool, error) {
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

func parseFlags(flags map[string]*CommandFlag, fields []string) (argsX []string, err error) {
	var currentFlag *CommandFlag
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

func (f *CommandFlag) setValue(value string) error {
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
