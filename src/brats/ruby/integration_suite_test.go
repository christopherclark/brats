package ruby_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/cloudfoundry/libbuildpack/packager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var bpData []string
var bpDir, buildpackVersion, buildpackCached, buildpackUncached string

const language = "ruby"

func init() {
	bpDir = os.Getenv("BP_DIR")
	flag.StringVar(&cutlass.DefaultMemory, "memory", "256M", "default memory for pushed apps")
	flag.StringVar(&cutlass.DefaultDisk, "disk", "384M", "default disk for pushed apps")
	flag.Parse()
}

var _ = SynchronizedBeforeSuite(func() []byte {
	// Run once
	if bpDir == "" {
		Fail("You must provide BP_DIR env")
	}

	buildpackVersion = fmt.Sprintf("brats_%s_%s_", language, time.Now().Format("20060102150405"))

	fmt.Println("Package cached")
	fileCached, err := packager.Package(bpDir, packager.CacheDir, buildpackVersion+"cached", true)
	Expect(err).NotTo(HaveOccurred())
	buildpackCached = buildpackVersion + "cached"
	command := exec.Command("cf", "create-buildpack", buildpackCached, fileCached, "100", "--enable")
	if output, err := command.CombinedOutput(); err != nil {
		fmt.Println(string(output))
		Fail("Could not create buildpack")
	}

	fmt.Println("Package uncached")
	fileUncached, err := packager.Package(bpDir, packager.CacheDir, buildpackVersion+"uncached", false)
	Expect(err).NotTo(HaveOccurred())
	buildpackUncached = buildpackVersion + "uncached"
	command = exec.Command("cf", "create-buildpack", buildpackUncached, fileUncached, "100", "--enable")
	if output, err := command.CombinedOutput(); err != nil {
		fmt.Println(string(output))
		Fail("Could not create buildpack")
	}

	data, err := json.Marshal([]string{
		bpDir, buildpackVersion,
		fileCached, buildpackCached,
		fileUncached, buildpackUncached,
	})
	Expect(err).NotTo(HaveOccurred())
	return data
}, func(data []byte) {
	// Run on all nodes
	var err error
	err = json.Unmarshal(data, &bpData)
	Expect(err).NotTo(HaveOccurred())

	bpDir = bpData[0]
	buildpackVersion = bpData[1]
	buildpackCached = bpData[3]
	buildpackUncached = bpData[5]

	cutlass.SeedRandom()
	cutlass.DefaultStdoutStderr = GinkgoWriter
})

var _ = SynchronizedAfterSuite(func() {
	// Run on all nodes
}, func() {
	// Run once
	if len(bpData) > 0 {
		Expect(os.Remove(bpData[2])).To(Succeed())
		Expect(exec.Command("cf", "delete-buildpack", buildpackCached).Run()).To(Succeed())
		Expect(os.Remove(bpData[4])).To(Succeed())
		Expect(exec.Command("cf", "delete-buildpack", buildpackUncached).Run()).To(Succeed())
	}

	Expect(cutlass.DeleteOrphanedRoutes()).To(Succeed())
})

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

func PushAppAndConfirm(app *cutlass.App) {
	Expect(app.Push()).To(Succeed())
	Eventually(func() ([]string, error) { return app.InstanceStates() }, 20*time.Second).Should(Equal([]string{"RUNNING"}))
	if app.Buildpack == "" {
		Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())
	}
}

// func Restart(app *cutlass.App) {
// 	Expect(app.Restart()).To(Succeed())
// 	Eventually(func() ([]string, error) { return app.InstanceStates() }, 20*time.Second).Should(Equal([]string{"RUNNING"}))
// }

// func ApiHasTask() bool {
// 	apiVersionString, err := cutlass.ApiVersion()
// 	Expect(err).To(BeNil())
// 	apiVersion, err := semver.Make(apiVersionString)
// 	Expect(err).To(BeNil())
// 	apiHasTask, err := semver.ParseRange(">= 2.75.0")
// 	Expect(err).To(BeNil())
// 	return apiHasTask(apiVersion)
// }

// func SkipUnlessUncached() {
// 	if cutlass.Cached {
// 		Skip("Running cached tests")
// 	}
// }

// func SkipUnlessCached() {
// 	if !cutlass.Cached {
// 		Skip("Running uncached tests")
// 	}
// }

func DestroyApp(app *cutlass.App) *cutlass.App {
	if app != nil {
		app.Destroy()
	}
	return nil
}

// func AssertUsesProxyDuringStagingIfPresent(fixtureName string) {
// 	Context("with an uncached buildpack", func() {
// 		BeforeEach(SkipUnlessUncached)

// 		It("uses a proxy during staging if present", func() {
// 			proxy, err := cutlass.NewProxy()
// 			Expect(err).To(BeNil())
// 			defer proxy.Close()

// 			bpFile := filepath.Join(bpDir, buildpackVersion+"tmp")
// 			cmd := exec.Command("cp", packagedBuildpack.File, bpFile)
// 			err = cmd.Run()
// 			Expect(err).To(BeNil())
// 			defer os.Remove(bpFile)

// 			traffic, err := cutlass.InternetTraffic(
// 				bpDir,
// 				filepath.Join("fixtures", fixtureName),
// 				bpFile,
// 				[]string{"HTTP_PROXY=" + proxy.URL, "HTTPS_PROXY=" + proxy.URL},
// 			)
// 			Expect(err).To(BeNil())

// 			destUrl, err := url.Parse(proxy.URL)
// 			Expect(err).To(BeNil())

// 			Expect(cutlass.UniqueDestination(
// 				traffic, fmt.Sprintf("%s.%s", destUrl.Hostname(), destUrl.Port()),
// 			)).To(BeNil())
// 		})
// 	})
// }

// func AssertNoInternetTraffic(fixtureName string) {
// 	It("has no traffic", func() {
// 		SkipUnlessCached()

// 		bpFile := filepath.Join(bpDir, buildpackVersion+"tmp")
// 		cmd := exec.Command("cp", packagedBuildpack.File, bpFile)
// 		err := cmd.Run()
// 		Expect(err).To(BeNil())
// 		defer os.Remove(bpFile)

// 		traffic, err := cutlass.InternetTraffic(
// 			bpDir,
// 			filepath.Join("fixtures", fixtureName),
// 			bpFile,
// 			[]string{},
// 		)
// 		Expect(err).To(BeNil())
// 		Expect(traffic).To(BeEmpty())
// 	})
// }
