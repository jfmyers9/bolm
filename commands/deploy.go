package commands

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	v1 "k8s.io/api/apps/v1"
	v1beta1 "k8s.io/api/apps/v1beta1"
	v1beta2 "k8s.io/api/apps/v1beta2"
	extv1beta1 "k8s.io/api/extensions/v1beta1"

	"k8s.io/client-go/kubernetes/scheme"
)

type Deployment struct {
	InstanceGroups []*InstanceGroup
}

type InstanceGroup struct {
	Name string
	Jobs []*Job
}

type Job struct {
	Name  string
	Image string
	Env   map[string]string
}

var RootCmd = &cobra.Command{
	Long:          "bolm yo",
	Short:         "bolm",
	SilenceErrors: true,
	Use:           "bolm",
}

var valuesFile string

func init() {
	deployCommand.Flags().StringVarP(&valuesFile, "values", "f", "", "values file")
	RootCmd.AddCommand(deployCommand)
}

var deployCommand = &cobra.Command{
	RunE:  deploy,
	Short: "bolm",
	Use:   "deploy <helm-chart> -f <values-file>",
}

func deploy(cmd *cobra.Command, args []string) error {
	chartName := args[0]
	productName := path.Base(chartName)
	tmpDir := path.Dir(chartName)

	fmt.Println(chartName)
	fmt.Println(tmpDir)

	if _, err := os.Stat(chartName); os.IsNotExist(err) {
		tmpDir, err := ioutil.TempDir("", "bolm")
		if err != nil {
			panic(err)
		}

		fmt.Println(tmpDir)

		fmt.Println("Fetching chart")
		fetchCmd := exec.Command("helm", "fetch", chartName, "--destination", tmpDir, "--untar")

		output, err := fetchCmd.CombinedOutput()
		if err != nil {
			panic(err)
		}

		fmt.Println(string(output))
	}

	fmt.Println("Templating the chart")
	templateArgs := []string{
		"template", filepath.Join(tmpDir, productName),
	}

	if valuesFile != "" {
		templateArgs = append(templateArgs, []string{"-f", valuesFile}...)
	}

	templateCommand := exec.Command("helm", templateArgs...)
	output, err := templateCommand.Output()
	if err != nil {
		log.Fatalf("%s: %s", err.Error(), string(output))
	}

	dep := parseObjects(output)
	data, err := yaml.Marshal(dep)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(data))

	return nil
}

func parseObjects(data []byte) *Deployment {
	deployment := &Deployment{}

	decoder := yaml.NewDecoder(bytes.NewReader(data))

	for {
		var object map[string]interface{}
		err := decoder.Decode(&object)
		if err == io.EOF {
			return deployment
		} else if err != nil {
			panic(err)
		}

		if len(object) == 0 {
			continue
		}

		f, err := yaml.Marshal(object)
		if err != nil {
			panic(err)
		}

		decode := scheme.Codecs.UniversalDeserializer().Decode
		obj, _, err := decode([]byte(f), nil, nil)
		if err == io.EOF {
			return deployment
		}
		if err != nil {
			panic(err)
		}

		// now use switch over the type of the object
		// and match each type-case
		switch o := obj.(type) {
		case *v1.Deployment:
		case *v1beta1.Deployment:
		case *v1beta2.Deployment:
		case *extv1beta1.Deployment:
			jobs := parseJobs(o)

			ig := &InstanceGroup{
				Name: igName(o.Spec.Template.ObjectMeta.Labels),
				Jobs: jobs,
			}

			deployment.InstanceGroups = append(deployment.InstanceGroups, ig)
			// o is a pod
		default:
			fmt.Printf("%T\n", o)
			//o is unknown for us
		}
	}
}

func parseJobs(d *extv1beta1.Deployment) []*Job {
	result := []*Job{}

	for _, c := range d.Spec.Template.Spec.Containers {
		result = append(result, &Job{
			Name:  c.Name,
			Image: c.Image,
		})
	}

	return result
}

func igName(labels map[string]string) string {
	result := []string{}
	for _, v := range labels {
		result = append(result, v)
	}

	return strings.Join(result, "-")
}
