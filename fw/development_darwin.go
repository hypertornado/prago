package fw

import (
	"fmt"
	"github.com/dshills/goauto"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func developmentCSS() {
	p := goauto.NewPipeline("CSS Pipeline", goauto.Verbose)
	defer p.Stop()

	wd := filepath.Join("public", "css")
	if err := p.WatchRecursive(wd, goauto.IgnoreHidden); err != nil {
		panic(err)
	}

	workflow := goauto.NewWorkflow(NewFuncCmd(compileCss))
	if err := workflow.WatchPattern(".*\\.less$"); err != nil {
		panic(err)
	}
	p.Add(workflow)
	p.Start()
}

//func task

type FuncCmd struct {
	f func() error
}

func NewFuncCmd(f func() error) goauto.Tasker {
	return FuncCmd{f}
}

func (fc FuncCmd) Run(info *goauto.TaskInfo) (err error) {
	return fc.f()
}

func NewTaskCmd(command string, args []string) goauto.Tasker {
	return TaskCmd{command, args}
}

type TaskCmd struct {
	cmd  string
	args []string
}

func (st TaskCmd) Run(info *goauto.TaskInfo) (err error) {
	fmt.Println("Running command", st.cmd, strings.Join(st.args, " "))
	cmd := exec.Command(st.cmd, st.args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	return nil
}
