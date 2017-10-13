package apt_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/cloudfoundry/libbuildpack/packager"
	"github.com/sclevine/spec"
	"github.com/sclevine/spec/report"
)

var bpLock = &sync.RWMutex{}
var bpHash map[string]bool = make(map[string]bool)

func getBP(dir string, info ...string) (string, error) {
	bpLock.Lock()
	defer bpLock.Unlock()

	info = append(info, "brats")
	bpName := strings.Join(info, "_")
	if _, present := bpHash[bpName]; present {
		return fmt.Sprintf("%s_buildpack", bpName), nil
	}

	buildpackVersion := time.Now().Format("20060102150405")
	file, err := packager.Package(dir, packager.CacheDir, buildpackVersion, true)
	if err != nil {
		return "", err
	}

	if err = cutlass.CreateOrUpdateBuildpack(bpName, file); err != nil {
		return "", err
	}

	bpHash[bpName] = true
	return fmt.Sprintf("%s_buildpack", bpName), nil
}

func getBpDir(lang string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if files, err := filepath.Glob(filepath.Join(dir, fmt.Sprintf("%s-buildpack", lang))); err != nil {
			return "", err
		} else if len(files) == 1 {
			return filepath.Join(dir, fmt.Sprintf("%s-buildpack", lang)), nil
		} else if dir == "/" {
			return "", fmt.Errorf("Could not find %s-buildpack", lang)
		} else {
			dir = filepath.Dir(dir)
		}
	}
}

func Test(t *testing.T) {
	spec.Run(t, "Apt Buildpack", func(t *testing.T, when spec.G, it spec.S) {
		var bpDir string
		var err error
		it.Before(func() {
			bpDir, err = getBpDir("apt")
			if err != nil {
				t.Error(err)
			}

			t.Log("bpDir: ", bpDir)
		})

		when("deploying an app with an updated version of the same buildpack", func() {
			var bp, bp2 string
			it.Before(func() {
				if bp, err = getBP(bpDir, "apt"); err != nil {
					t.Error(err)
				}
				if bp2, err = getBP(bpDir, "apt", "second"); err != nil {
					t.Error(err)
				}
				t.Log("bpName: ", bp, bp2)
			})

			// it.After(func() {
			// 	t.Log("after")
			// })

			it("prints useful warning message to stdout", func() {
				app := cutlass.New(filepath.Join(os.Getenv("GOPATH"), "fixtures", "apt"))
				t.Log(filepath.Join(os.Getenv("GOPATH"), "fixtures", "apt"))
				app.Buildpacks = []string{bp, "binary_buildpack"}
				if err := app.Push(); err != nil {
					time.Sleep(1 * time.Second)
					t.Log(app.Stdout.String())
					t.Fatal(err)
				}
				if body, err := app.GetBody("/"); err != nil {
					t.Fatal(err)
				} else if !strings.Contains(body, "Ascii: d") {
					t.Log(body)
					t.Fatal("App not launched correctly")
				}
				if strings.Contains(app.Stdout.String(), "WARNING: buildpack version changed from") {
					t.Fatal("Received warning on first push")
				}

				app.Buildpacks = []string{bp2, "binary_buildpack"}
				if err := app.Push(); err != nil {
					time.Sleep(1 * time.Second)
					t.Log(app.Stdout.String())
					t.Fatal(err)
				}
				if body, err := app.GetBody("/"); err != nil {
					t.Fatal(err)
				} else if !strings.Contains(body, "Ascii: d") {
					t.Log(body)
					t.Fatal("App not launched correctly")
				}
				if !strings.Contains(app.Stdout.String(), "WARNING: buildpack version changed from") {
					t.Log(app.Stdout.String())
					t.Fatal("Did not receive warning message: 'buildpack version changed'")
				}
			})
			// it("should do other thing", func() {
			// 	t.Log("second")
			// })
			// it("should do another thing", func() {
			// 	t.Log("third")
			// })
		}, spec.Parallel())
	}, spec.Report(report.Terminal{}))
}
