package main

import (
	"bytes"
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type config struct {
	kubeconfig        string
	kubeconfigContent string
	kubeconfigContext string
}

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return &schema.Provider{
				Schema: map[string]*schema.Schema{
					"kubeconfig": &schema.Schema{
						Type:     schema.TypeString,
						Optional: true,
					},
					"kubeconfig_content": &schema.Schema{
						Type:     schema.TypeString,
						Optional: true,
					},
					"kubeconfig_context": &schema.Schema{
						Type:     schema.TypeString,
						Optional: true,
					},
				},
				ResourcesMap: map[string]*schema.Resource{
					"k8s_manifest": resourceManifest(),
				},
				ConfigureFunc: func(d *schema.ResourceData) (interface{}, error) {
					return &config{
						kubeconfig:        d.Get("kubeconfig").(string),
						kubeconfigContent: d.Get("kubeconfig_content").(string),
						kubeconfigContext: d.Get("kubeconfig_context").(string),
					}, nil
				},
			}
		},
	})
}

func resourceManifest() *schema.Resource {
	return &schema.Resource{
		Create: resourceManifestCreate,
		Read:   resourceManifestRead,
		Update: resourceManifestUpdate,
		Delete: resourceManifestDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: false,
				ForceNew: true,
			},
			"namespace": &schema.Schema{
				Type:      schema.TypeString,
				Optional: true,
				Sensitive: false,
				ForceNew: true,
			},
			"kind": &schema.Schema{
				Type:      schema.TypeString,
				Required: true,
				Sensitive: false,
				ForceNew: true,
			},
			"content": &schema.Schema{
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: false,
			},
		},
	}
}

func run(cmd *exec.Cmd) error {
	stderr := &bytes.Buffer{}
	cmd.Stderr = stderr
	if err := cmd.Run(); err != nil {
		cmdStr := cmd.Path + " " + strings.Join(cmd.Args, " ")
		if stderr.Len() == 0 {
			return fmt.Errorf("%s: %v", cmdStr, err)
		}
		return fmt.Errorf("%s %v: %s", cmdStr, err, stderr.Bytes())
	}
	return nil
}

func kubeconfigPath(m interface{}) (string, func(), error) {
	kubeconfig := m.(*config).kubeconfig
	kubeconfigContent := m.(*config).kubeconfigContent
	var cleanupFunc = func() {}

	if kubeconfig != "" && kubeconfigContent != "" {
		return kubeconfig, cleanupFunc, fmt.Errorf("both kubeconfig and kubeconfig_content are defined, " +
			"please use only one of the paramters")
	} else if kubeconfigContent != "" {
		tmpfile, err := ioutil.TempFile("", "kubeconfig_")
		if err != nil {
			defer cleanupFunc()
			return "", cleanupFunc, fmt.Errorf("creating a kubeconfig file: %v", err)
		}

		cleanupFunc = func() { os.Remove(tmpfile.Name()) }

		if _, err = io.WriteString(tmpfile, kubeconfigContent); err != nil {
			defer cleanupFunc()
			return "", cleanupFunc, fmt.Errorf("writing kubeconfig to file: %v", err)
		}
		if err = tmpfile.Close(); err != nil {
			defer cleanupFunc()
			return "", cleanupFunc, fmt.Errorf("completion of write to kubeconfig file: %v", err)
		}

		kubeconfig = tmpfile.Name()
	}

	if kubeconfig != "" {
		return kubeconfig, cleanupFunc, nil
	}

	return "", cleanupFunc, nil
}

func kubectl(m interface{}, kubeconfig string, args ...string) *exec.Cmd {
	if kubeconfig != "" {
		args = append([]string{"--kubeconfig", kubeconfig}, args...)
	}

	context := m.(*config).kubeconfigContext
	if context != "" {
		args = append([]string{"--context", context}, args...)
	}

	cmd := exec.Command("kubectl", args...)
	cmd.Env = os.Environ()
	return cmd
}

func processContent(content string, name string, namespace string, kind string)(string, error) {
	any := map[string]interface{}{}
	err := yaml.Unmarshal([]byte(content), &any)
	if err != nil {
		return "", fmt.Errorf("parsing yaml: %v", err)
	}
	_, ok := any["metadata"]
	if !ok {
		any["metadata"] = map[interface {}]interface {}{}
	}

	any["kind"] = kind
	metadata := any["metadata"].(map[interface {}]interface {})
	if namespace != "" {
		metadata["namespace"] = namespace
	} else {
		_, ok = metadata["namespace"]
		if ok {
			return "", fmt.Errorf("no namespace provided but yml has namespace")
		}
	}
	metadata["name"] = name
	any["metadata"] = metadata

	out, err := yaml.Marshal(any)
	if err != nil {
		return "", fmt.Errorf("generating yaml: %v", err)
	}
	return string(out), nil
}

func resourceManifestCreate(d *schema.ResourceData, m interface{}) error {
	name := d.Get("name").(string)
	kind := d.Get("kind").(string)
	namespace, ok := d.GetOk("namespace")
	if !ok {
		namespace = ""
	}

	kubeconfig, cleanup, err := kubeconfigPath(m)
	if err != nil {
		return fmt.Errorf("determining kubeconfig: %v", err)
	}
	defer cleanup()

	processed, err := processContent(d.Get("content").(string), name, namespace.(string), kind)
	if err != nil {
		return fmt.Errorf("processing content: %v", err)
	}

	cmd := kubectl(m, kubeconfig, "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(processed)
	if err := run(cmd); err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%s/%s/%s", kind, namespace.(string), name))
	return nil
}

func resourceManifestUpdate(d *schema.ResourceData, m interface{}) error {
	name := d.Get("name").(string)
	kind := d.Get("kind").(string)
	namespace, ok := d.GetOk("namespace")
	if !ok {
		namespace = ""
	}

	kubeconfig, cleanup, err := kubeconfigPath(m)
	if err != nil {
		return fmt.Errorf("determining kubeconfig: %v", err)
	}
	defer cleanup()

	processed, err := processContent(d.Get("content").(string), name, namespace.(string), kind)
	if err != nil {
		return fmt.Errorf("processing content: %v", err)
	}

	cmd := kubectl(m, kubeconfig, "apply", "-f", "-")
	cmd.Stdin = strings.NewReader(processed)
	return run(cmd)
}

func resourceManifestDelete(d *schema.ResourceData, m interface{}) error {
	name := d.Get("name").(string)
	kind := d.Get("kind").(string)
	args := []string{"delete", kind, name}

	namespace, ok := d.GetOk("namespace")
	if ok {
		args = append(args, "-n", namespace.(string))
	}

	kubeconfig, cleanup, err := kubeconfigPath(m)
	if err != nil {
		return fmt.Errorf("determining kubeconfig: %v", err)
	}
	defer cleanup()

	cmd := kubectl(m, kubeconfig, args...)
	return run(cmd)
}

func resourceManifestRead(d *schema.ResourceData, m interface{}) error {
	name := d.Get("name").(string)
	kind := d.Get("kind").(string)
	args := []string{"get", "--ignore-not-found", kind, name}

	// If importing a state without name or kind set
	if name == "" || kind == "" {
		d.SetId("")
		return nil
	}

	namespace, ok := d.GetOk("namespace")
	if ok {
		args = append(args, "-n", namespace.(string))
	}

	stdout := &bytes.Buffer{}
	kubeconfig, cleanup, err := kubeconfigPath(m)
	if err != nil {
		return fmt.Errorf("determining kubeconfig: %v", err)
	}
	defer cleanup()

	cmd := kubectl(m, kubeconfig, args...)
	cmd.Stdout = stdout
	if err := run(cmd); err != nil {
		return err
	}
	if strings.TrimSpace(stdout.String()) == "" {
		d.SetId("")
	}
	return nil
}
