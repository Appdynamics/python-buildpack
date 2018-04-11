package appdynamics

import (
	"os"
	"encoding/json"
	"fmt"
	"strings"
	"io/ioutil"
	"errors"
)

type Plan struct {
	Credentials Credential `json:"credentials"`
}

type Credential struct {
	ControllerHost  string `json:"host-name"`
	ControllerPort  string `json:"port"`
	SslEnabled      bool   `json:"ssl-enabled"`
	AccountAccessKey string `json:"account-access-key"`
	AccountName     string `json:"account-name"`
}

type VcapApplication struct {
	ApplicationName string `json:"application_name"`
	ApplicationId   string `json:"application_id"`
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func GatherAppdynamicsInfo() map[string]string {
	vcapServices := os.Getenv("VCAP_SERVICES")
	vcapApplication := os.Getenv("VCAP_APPLICATION")

	services := make(map[string][]Plan)
	application := VcapApplication{}

	err := json.Unmarshal([]byte(vcapServices), &services)
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal([]byte(vcapApplication), &application)
	if err != nil {
		fmt.Println(err)
	}

	if val, ok := services["appdynamics"]; ok {
		appdynamicsPlan := val[0].Credentials
		AppdEnvMap := make(map[string]string)

		AppdEnvMap["APPD_APP_NAME"] = getEnv("APPD_APP_NAME", application.ApplicationName)
		AppdEnvMap["APPD_TIER_NAME"] = getEnv("APPD_TIER_NAME", application.ApplicationName)
		AppdEnvMap["APPD_NODE_NAME"] = getEnv("APPD_NODE_NAME", application.ApplicationName)
		AppdEnvMap["APPD_CONTROLLER_HOST"] = appdynamicsPlan.ControllerHost
		AppdEnvMap["APPD_CONTROLLER_PORT"] = appdynamicsPlan.ControllerPort
		AppdEnvMap["APPD_ACCOUNT_ACCESS_KEY"] = appdynamicsPlan.AccountAccessKey
		AppdEnvMap["APPD_ACCOUNT_NAME"] = appdynamicsPlan.AccountName
		AppdEnvMap["APPD_SSL_ENABLED"] = "off"

		return AppdEnvMap

	} else {
		return nil
	}

}

func GenerateAppdynamicsScript(envVars map[string]string) string {

	scriptContents := "# Autogenerated Appdynamics Script \n"

	for envKey, envVal := range envVars {
		envStr := fmt.Sprintf("export %s=%s", envKey, envVal)
		scriptContents += "\n" + envStr
	}

	return scriptContents
}

func GenerateStartUpCommand(startCommand string) (string, error) {
	webCommands := strings.SplitN(startCommand, ":", 2)
	if len(webCommands) != 2 {
		return "", errors.New("improper format found in Procfile")
	}
	return fmt.Sprintf("web: pyagent run -- %s", webCommands[1]), nil
}

func RewriteProcFile(procFilePath string) error {
	startCommand, err := ioutil.ReadFile(procFilePath)
	if err != nil {
		return err
	}
	if newCommand, err :=  GenerateStartUpCommand(string(startCommand)); err != nil {
		return err
	} else {
		if err := ioutil.WriteFile(procFilePath, []byte(newCommand), 0644); err != nil {
			return err
		}
	}
	return nil
}

/*func main() {
	vcapServicesStr := `{"appdynamics": [{"application-name": "app", "binding_name": null, "volume_mounts": [], "node-name": "node", "plan": "my-controller-config-name", "syslog_drain_url": null, "tier-name": "tier", "credentials": {"host-name": "442-saas-controller.e2e.appd-test.com", "plan-name": "my-controller-config-name", "guid": "c27156b4-34b0-4390-b0d8-1687fe9480c9", "plan-description": "my-controller-config-options", "ssl-enabled": false, "account-access-key": "c14b43ab-ceed-405e-8416-b028764f461b", "account-name": "customer1", "port": "8090"}, "tags": [], "label": "appdynamics", "instance_name": "1", "provider": null}]}`
	vcapApplicationStr := `{"cf_api": "https://api.sys.pie-21.cfplatformeng.com", "application_uris": ["sample-app.cfapps.pie-21.cfplatformeng.com"], "users": null, "application_id": "98d1e9d3-cea7-47d1-83a4-987661770d1a", "name": "sample-app", "limits": {"fds": 16384}, "application_name": "sample-app", "space_name": "appdynamics-space", "space_id": "a9ba301e-d4f3-41b7-a344-d381c461ae50", "uris": ["sample-app.cfapps.pie-21.cfplatformeng.com"]}`
	os.Setenv("VCAP_SERVICES", vcapServicesStr)
	os.Setenv("VCAP_APPLICATION", vcapApplicationStr)

	appd := GatherAppdynamicsInfo()
	script := GenerateAppdynamicsScript(appd)
	fmt.Println(script)

	fmt.Print(GenerateStartUpCommand("web: python hello.py:8080"))
}*/