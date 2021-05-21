package build

import (
	"bufio"
	"github.com/Justice-love/go-aspect/util"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"
)

type Inspect struct {
	Root         string
	IsMod        bool
	Dependencies []string // go mod verdor -v
}

func NewInspect(root string) *Inspect {
	fs, err := ioutil.ReadDir(root)
	if err != nil {
		log.Fatalf("root dir error, %v", err)
	}
	inspect := &Inspect{Root: root}
	for _, one := range fs {
		if one.Name() == "go.mod" {
			inspect.IsMod = true
		}
	}
	if !inspect.IsMod {
		return inspect
	}
	inspect.vendor(root)
	inspect.Dependencies = inspect.depend(root)
	return inspect
}

func (*Inspect) vendor(root string) {
	cmd := exec.Command("/bin/bash", "-c", "go mod vendor")
	cmd.Dir = root
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	_ = cmd.Start()
	_ = cmd.Wait()
}

func (i *Inspect) depend(root string) (depend []string) {
	f, err := os.Open(strings.Join([]string{root, "vendor", "modules.txt"}, "/"))
	if err != nil {
		return
	}
	defer func() {
		_ = f.Close()
	}()
	reader := bufio.NewReader(f)
	for {
		content, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		contentStr := string(content)
		if strings.HasPrefix(contentStr, "#") {
			continue
		}
		arr := util.SplitSpace(contentStr)
		depend = append(depend, arr[0])
	}
	return
}

func (i *Inspect) EndpointPath() (path []string) {
	if p := i.Find(i.Root); p != "" {
		path = append(path, p)
	}
	for _, one := range i.Dependencies {
		if p := i.Find(i.Root + "/" + "vendor" + "/" + one); p != "" {
			path = append(path, p)
		}
	}
	return
}

func (*Inspect) Find(path string) string {
	fs, err := ioutil.ReadDir(path)
	if err != nil {
		log.Fatalf("root dir error, %v", err)
	}
	for _, one := range fs {
		if one.Name() == "aspect.point" {
			return path + "/" + "aspect.point"
		}
	}
	return ""
}
