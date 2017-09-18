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

	generateApp := func(rubyVersion string) string {
		appDir, err := cutlass.CopyFixture(filepath.Join(rootDir, "fixtures", "ruby", "simple_brats"))
		Expect(err).ToNot(HaveOccurred())

		gemfile, err := template.ParseFiles(filepath.Join(appDir, "Gemfile"))
		Expect(err).ToNot(HaveOccurred())

		fh, err := os.Create(filepath.Join(appDir, "Gemfile"))
		Expect(err).ToNot(HaveOccurred())
		defer fh.Close()

		Expect(gemfile.Execute(fh, map[string]string{"RubyVersion": rubyVersion})).To(Succeed())

		return appDir
	}

	Describe("deploying an app with an updated version of the same buildpack", func() {
		PIt("prints useful warning message to stdout")
	})

	Describe("For all supported Ruby versions", func() {
		manifest, _ := libbuildpack.NewManifest(os.Getenv("BP_DIR"), libbuildpack.NewLogger(os.Stdout), time.Now())
		for _, version2 := range manifest.AllDependencyVersions("ruby") {
			version := version2
			It("with Ruby version "+version, func() {
				appDir := generateApp(version)
				app = cutlass.New(appDir)
				app.Buildpack = buildpackCached
				defer os.RemoveAll(appDir)

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

	Describe("staging with a version of ruby that is not the latest patch release in the manifest", func() {
		var appDir string
		BeforeEach(func() {
			versions, err := GetSortedDepVersions("ruby")
			Expect(err).ToNot(HaveOccurred())
			Expect(len(versions) > 0).To(BeTrue())
			appDir = generateApp(versions[0])
			app = cutlass.New(appDir)
			app.Buildpack = buildpackCached
			PushAppAndConfirm(app)
		})
		AfterEach(func() {
			if appDir != "" {
				_ = os.RemoveAll(appDir)
			}
		})

		It("logs a warning that tells the user to upgrade the dependency", func() {
			Expect(app.Stdout.String()).To(ContainSubstring("WARNING.*A newer version of ruby is available in this buildpack"))
		})
	})

	PIt("staging with custom buildpack that uses credentials in manifest dependency uris", func() {})

	It("deploying an app that has an executable .profile script", func() {
		versions, err := GetSortedDepVersions("ruby")
		Expect(err).ToNot(HaveOccurred())
		Expect(len(versions) > 0).To(BeTrue())
		appDir := generateApp(versions[len(versions)-1])
		defer os.RemoveAll(appDir)
		app = cutlass.New(appDir)
		app.Buildpack = buildpackCached
		Expect(AddDotProfileScriptToApp(app.Path)).To(Succeed())
		PushAppAndConfirm(app)

		By("executes the .profile script", func() {
			Expect(app.Stdout.String()).To(ContainSubstring("PROFILE_SCRIPT_IS_PRESENT_AND_RAN"))
		})

		By("does not let me view the .profile script", func() {
			body, err := app.GetBody("/.profile")
			Expect(err).ToNot(HaveOccurred())
			Expect(body).ToNot(ContainSubstring("Unknown MySQL server host 'testing'"))
			Expect(body).To(ContainSubstring("<h1>Not Found</h1>"))
		})
	})

	Describe("deploying an app that has sensitive environment variables", func() {
		var appDir string
		BeforeEach(func() {
			versions, err := GetSortedDepVersions("ruby")
			Expect(err).ToNot(HaveOccurred())
			Expect(len(versions) > 0).To(BeTrue())
			appDir = generateApp(versions[len(versions)-1])
			app = cutlass.New(appDir)
			app.Buildpack = buildpackCached
			Expect(AddDotProfileScriptToApp(app.Path)).To(Succeed())
			PushAppAndConfirm(app)
		})
		AfterEach(func() {
			if appDir != "" {
				_ = os.RemoveAll(appDir)
			}
		})

		PIt("will not write credentials to the app droplet", func() {
			// FIXME -- https://github.com/cloudfoundry/machete/blob/7f517ffa07d99362ab86ccab654dbed8583c6213/lib/machete/matchers/app_keeps_credentials_out_of_droplet.rb
			// expect(app.name).to keep_credentials_out_of_droplet
		})
	})
})
