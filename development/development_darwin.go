package development

import (

	//"github.com/dshills/goauto"
	"fmt"
	"github.com/radovskyb/watcher"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func developmentLess(sourcePath, targetPath string) {
	w := watcher.New()
	w.SetMaxEvents(1)

	go func() {
		for {
			select {
			case event := <-w.Event:
				fmt.Println(event) // Print the event's info.
				compileLess(filepath.Join(sourcePath, "index.less"), targetPath)
			case err := <-w.Error:
				log.Fatalln(err)
			case <-w.Closed:
				return
			}
		}
	}()

	if err := w.AddRecursive(sourcePath); err != nil {
		log.Fatalln(err)
	}

	if err := w.Start(time.Millisecond * 100); err != nil {
		log.Fatalln(err)
	}

	compileLess(filepath.Join(sourcePath, "index.less"), targetPath)
}

/*func developmentLess(sourcePath, targetPath string) {

	p := goauto.NewPipeline("Less Pipeline", goauto.Verbose)
	defer p.Stop()

	if err := p.WatchRecursive(sourcePath, goauto.IgnoreHidden); err != nil {
		panic(err)
	}

	workflow := goauto.NewWorkflow(NewFuncCmd(func() error {
		return compileLess(filepath.Join(sourcePath, "index.less"), targetPath)
	}))

	if err := workflow.WatchPattern(".*\\.less$"); err != nil {
		panic(err)
	}
	p.Add(workflow)

	err := compileLess(filepath.Join(sourcePath, "index.less"), targetPath)
	if err != nil {
		panic(err)
	}

	p.Start()
}

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
}*/

/*func (st TaskCmd) Run(info *goauto.TaskInfo) (err error) {
	fmt.Println("Running command", st.cmd, strings.Join(st.args, " "))
	cmd := exec.Command(st.cmd, st.args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	return nil
}*/

func compileLess(from, to string) error {
	outfile, err := os.Create(to)
	if err != nil {
		return err
	}
	defer outfile.Close()

	return commandHelper(exec.Command("lessc", from), outfile)
}

func commandHelper(cmd *exec.Cmd, out io.Writer) error {
	var err error
	cmd.Stdout = out
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return err
	}
	err = cmd.Wait()
	if err != nil {
		return err
	}
	return nil
}
