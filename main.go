package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	buildTime  string
	validID    = regexp.MustCompile(`.*=vault:.*:.*`)
	httpClient = &http.Client{
		Timeout: *appConfig.httpTimeout,
	}
)

type appConfigType struct {
	Version           string
	debug             *bool
	httpTimeout       *time.Duration
	vaultAuthMode     *string
	vaultAuthLogin    *string
	vaultAuthPassword *string
	vaultAddr         *string
	vaultToken        *string
	vaultK8sRole      *string
	output            *string
}

var appConfig = appConfigType{
	Version: "1.0.1",
	debug: kingpin.Flag(
		"debug",
		"debug",
	).Default("false").Bool(),
	httpTimeout: kingpin.Flag(
		"http.timeout",
		"Vault http timeout",
	).Default("10s").Duration(),
	vaultAuthMode: kingpin.Flag(
		"vault.mode",
		"Vault auth mode",
	).Default("token").String(),
	vaultAuthLogin: kingpin.Flag(
		"vault.auth.login",
		"Vault auth login",
	).Default(os.Getenv("VAULT_LOGIN")).String(),
	vaultAuthPassword: kingpin.Flag(
		"vault.auth.password",
		"Vault auth login",
	).Default(os.Getenv("VAULT_PASSWORD")).String(),
	vaultAddr: kingpin.Flag(
		"vault.address",
		"Vault address",
	).Default(os.Getenv("VAULT_ADDR")).String(),
	vaultToken: kingpin.Flag(
		"vault.token",
		"Vault token",
	).Default(os.Getenv("VAULT_TOKEN")).String(),
	vaultK8sRole: kingpin.Flag(
		"vault.k8s.role",
		"Vault k8s role",
	).Default("default").String(),
	output: kingpin.Flag(
		"output",
		"Formated env",
	).Default("vault-env").String(),
}

func kubernetesLogin(client *api.Client) string {
	content, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/token")
	if err != nil {
		panic(err)
	}

	// to pass the password
	options := map[string]interface{}{
		"role": *appConfig.vaultK8sRole,
		"jwt":  strings.Trim(string(content), " "),
	}

	// PUT call to get a token
	secret, err := client.Logical().Write("auth/kubernetes/login", options)
	if err != nil {
		panic(err)
	}

	token := secret.Auth.ClientToken
	return token
}

func userpassLogin(client *api.Client) string {
	// to pass the password
	options := map[string]interface{}{
		"password": *appConfig.vaultAuthPassword,
	}
	path := fmt.Sprintf("auth/userpass/login/%s", *appConfig.vaultAuthLogin)

	// PUT call to get a token
	secret, err := client.Logical().Write(path, options)
	if err != nil {
		panic(err)
	}

	token := secret.Auth.ClientToken
	return token
}

func main() {
	kingpin.Version(appConfig.Version)
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

	client, err := api.NewClient(&api.Config{Address: *appConfig.vaultAddr, HttpClient: httpClient})
	if err != nil {
		panic(err)
	}

	vaultToken := ""

	if *appConfig.debug {
		fmt.Println("*appConfig.vaultAuthMode =", *appConfig.vaultAuthMode)
	}
	switch *appConfig.vaultAuthMode {
	case "token":
		vaultToken = *appConfig.vaultToken
		break
	case "login":
		vaultToken = userpassLogin(client)
		break
	case "k8s":
		vaultToken = kubernetesLogin(client)
		break
	default:
		panic("unknown auth mode")
	}

	if *appConfig.debug {
		fmt.Println("vaultToken =", vaultToken)
	}

	client.SetToken(vaultToken)

	if err != nil {
		panic(err)
	}

	var file *os.File

	if !*appConfig.debug {
		var err error
		file, err = os.OpenFile(*appConfig.output, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0777)
		defer file.Close()
		if err != nil {
			panic(err)
		}
	}

	for _, pair := range os.Environ() {
		if validID.MatchString(pair) {

			cmd := strings.Split(pair, "=")
			vaultParse := strings.Split(cmd[1], ":")

			data, err := client.Logical().Read(vaultParse[1])
			if err != nil {
				panic(err)
			}

			env := fmt.Sprintf("export %s=%s", cmd[0], data.Data["data"].(map[string]interface{})[vaultParse[2]])

			if !*appConfig.debug {
				file.WriteString(env)
				file.WriteString("\n")
			} else {
				fmt.Println(env)
			}
		}
	}

	if !*appConfig.debug {
		file.WriteString("$*")
	}

}
