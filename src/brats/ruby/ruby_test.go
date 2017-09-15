package ruby_test

import (
	"html/template"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("For the ruby buildpack", func() {
	var rootDir = "/home/pivotal/workspace/brats" // FIXME
	var app *cutlass.App
	AfterEach(func() { app = DestroyApp(app) })

	Describe("deploying an app with an updated version of the same buildpack", func() {
		PIt("prints useful warning message to stdout")
	})

	Describe("For all supported Ruby versions", func() {
		manifest, _ := libbuildpack.NewManifest(os.Getenv("BP_DIR"), libbuildpack.NewLogger(os.Stdout), time.Now())
		for _, version2 := range manifest.AllDependencyVersions("ruby") {
			version := version2
			It("with Ruby version "+version, func() {
				appDir, err := cutlass.CopyFixture(filepath.Join(rootDir, "fixtures", "ruby", "simple_brats"))
				Expect(err).ToNot(HaveOccurred())
				defer os.RemoveAll(appDir)

				gemfile, err := template.ParseFiles(filepath.Join(appDir, "Gemfile"))
				Expect(err).ToNot(HaveOccurred())

				fh, err := os.Create(filepath.Join(appDir, "Gemfile"))
				Expect(err).ToNot(HaveOccurred())
				defer fh.Close()

				Expect(gemfile.Execute(fh, map[string]string{"RubyVersion": version})).To(Succeed())

				app = cutlass.New(appDir)
				app.Buildpack = buildpackCached

				By("installs the correct version of Ruby", func() {
					PushAppAndConfirm(app)
					Expect(app.Stdout.String()).To(ContainSubstring("Installing ruby " + version))

					for i := 1; i <= 2; i++ {
						Expect(app.GetBody("/version")).To(ContainSubstring(version))
					}
				})

				By("runs a simple webserver", func() {
					for i := 1; i <= 2; i++ {
						Expect(app.GetBody("/")).To(ContainSubstring("Hello, World"))
					}
				})

				By("parses XML with nokogiri", func() {
					for i := 1; i <= 2; i++ {
						Expect(app.GetBody("/nokogiri")).To(ContainSubstring("Hello, World"))
					}
				})

				By("supports EventMachine", func() {
					for i := 1; i <= 2; i++ {
						Expect(app.GetBody("/em")).To(ContainSubstring("Hello, EventMachine"))
					}
				})

				By("encrypts with bcrypt", func() {
					for i := 1; i <= 2; i++ {
						// browser.visit_path("/bcrypt")
						// crypted_text = BCrypt::Password.new(browser.body)
						// expect(crypted_text).to eq "Hello, bcrypt"
						cryptedText, err := app.GetBody("/bcrypt")
						Expect(err).ToNot(HaveOccurred())
						Expect(bcrypt.CompareHashAndPassword([]byte(cryptedText), []byte("Hello, bcrypt"))).To(Succeed())
					}
				})

				By("supports bson", func() {
					for i := 1; i <= 2; i++ {
						Expect(app.GetBody("/bson")).To(ContainSubstring("00040000"))
					}
				})

				By("supports postgres", func() {
					for i := 1; i <= 2; i++ {
						Expect(app.GetBody("/pg")).To(ContainSubstring("could not connect to server: No such file or directory"))
					}
				})

				By("supports mysql2", func() {
					for i := 1; i <= 2; i++ {
						Expect(app.GetBody("/mysql2")).To(ContainSubstring("Unknown MySQL server host 'testing'"))
					}
				})
			})
		}
	})

	PIt("staging with ruby buildpack that sets EOL on dependency", func() {})

	// BeforeEach(func() {
	// 	app = cutlass.New(filepath.Join(bpDir, "fixtures", "rails51"))
	// 	app.SetEnv("BP_DEBUG", "1")
	// })

	// It("Installs node6 and runs", func() {
	// 	PushAppAndConfirm(app)
	// 	Expect(app.Stdout.String()).To(ContainSubstring("Installing node 6."))

	// 	Expect(app.GetBody("/")).To(ContainSubstring("Hello World"))
	// 	Eventually(func() string { return app.Stdout.String() }, 10*time.Second).Should(ContainSubstring(`Started GET "/" for`))

	// 	By("Make sure supply does not change BuildDir", func() {
	// 		Expect(app.Stdout.String()).To(ContainSubstring("BuildDir Checksum Before Supply: 5d823d48d154ee2622e8cf8c2fb21ff7"))
	// 		Expect(app.Stdout.String()).To(ContainSubstring("BuildDir Checksum After Supply: 5d823d48d154ee2622e8cf8c2fb21ff7"))
	// 	})
	// })
})
