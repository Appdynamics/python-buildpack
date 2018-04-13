package integration_test

import (
	"os/exec"
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"fmt"
)

var _ = Describe("Appdynamics Integration", func() {
	var app, appdServiceBrokerApp *cutlass.App
	var sbUrl string
	const serviceName = "TestAppdynamics"

	RunCf := func(args ...string) error {
		command := exec.Command("cf", args...)
		command.Stdout = GinkgoWriter
		command.Stderr = GinkgoWriter
		return command.Run()
	}

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil

		command := exec.Command("cf", "purge-service-offering", "-f", serviceName)
		command.Stdout = GinkgoWriter
		command.Stderr = GinkgoWriter
		_ = command.Run()

		command = exec.Command("cf", "delete-service-broker", "-f", serviceName)
		command.Stdout = GinkgoWriter
		command.Stderr = GinkgoWriter
		_ = command.Run()

		if appdServiceBrokerApp != nil {
			appdServiceBrokerApp.Destroy()
		}
		appdServiceBrokerApp = nil
	})

	BeforeEach(func() {
		appdServiceBrokerApp = cutlass.New(filepath.Join(bpDir, "fixtures", "fake_appd_service_broker"))
		Expect(appdServiceBrokerApp.Push()).To(Succeed())
		Eventually(func() ([]string, error) { return appdServiceBrokerApp.InstanceStates() }, 20*time.Second).Should(Equal([]string{"RUNNING"}))

		var err error
		sbUrl, err = appdServiceBrokerApp.GetUrl("")
		Expect(err).ToNot(HaveOccurred())

		RunCf("create-service-broker", serviceName, "username", "password", sbUrl, "--space-scoped")
		RunCf("create-service", serviceName, "public", serviceName)

		app = cutlass.New(filepath.Join(bpDir, "fixtures", "with_appdynamics"))
		app.SetEnv("BP_DEBUG", "true")
		PushAppAndConfirm(app)
	})

	Context("bind a python app with appdynamics service", func() {
		BeforeEach(func() {
			RunCf("bind-service", app.Name, serviceName)

			app.Stdout.Reset()
			RunCf("restage", app.Name)
		})

		It("test if appdynamics was successfully bound", func() {
			//Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())
			fmt.Print(app.Stdout.String())
			//Expect(app.Stdout.String()).To(ContainSubstring("Snyk token was found"))
		})
	})
})