package pragobleve

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type PragoBleve struct {
	path string
}

func New(path string) *PragoBleve {
	return &PragoBleve{
		path: path,
	}
}

func (pb *PragoBleve) indexPath(name string) string {
	return fmt.Sprintf("%s/%s.bleve", pb.path, name)
}

func (pb *PragoBleve) DeleteIndexByName(name string) error {
	name = strings.ToLower(name)
	if strings.Contains(name, "/") {
		return errors.New("bad index name")
	}
	return os.RemoveAll(pb.indexPath(name))
}
