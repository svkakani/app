package packager

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/app/internal"
	"github.com/docker/app/internal/types"
	"github.com/docker/app/pkg/resto"
	"github.com/pkg/errors"
	yaml "gopkg.in/yaml.v2"
)

// Pull loads an app from a registry
func Pull(repotag string, outputDir string) error {
	payload, err := resto.PullConfigMulti(context.Background(), repotag, resto.RegistryOptions{Insecure: os.Getenv("DOCKERAPP_INSECURE_REGISTRY") != ""})
	if err != nil {
		return err
	}
	repoComps := strings.Split(repotag, ":")
	repo := repoComps[0]
	// handle the case where a port was specified in the domain part of repotag
	if len(repoComps) == 3 || (len(repoComps) == 2 && strings.Contains(repoComps[1], "/")) {
		repo = repoComps[1]
	}
	appDir := filepath.Join(outputDir, internal.DirNameFromAppName(filepath.Base(repo)))
	err = os.Mkdir(appDir, 0755)
	if err != nil {
		return errors.Wrap(err, "failed to create output application directory")
	}
	for k, v := range payload {
		// poor man's security
		if strings.Contains(k, "/") || strings.Contains(k, "\\") {
			continue
		}
		target := filepath.Join(appDir, k)
		err = ioutil.WriteFile(target, []byte(v), 0644)
		if err != nil {
			return errors.Wrap(err, "failed to write output file")
		}
	}
	return nil
}

// Push pushes an app to a registry
func Push(appname, namespace, tag string) error {
	app, err := Extract(appname)
	if err != nil {
		return err
	}
	defer app.Cleanup()
	payload := make(map[string]string)
	for _, n := range internal.FileNames {
		data, err := ioutil.ReadFile(filepath.Join(app.AppName, n))
		if err != nil {
			return err
		}
		if (namespace == "" || tag == "") && n == internal.MetadataFileName {
			var metadata types.AppMetadata
			err := yaml.Unmarshal(data, &metadata)
			if err != nil {
				return errors.Wrap(err, "failed to parse application metadata")
			}
			if namespace == "" {
				namespace = metadata.Namespace
			}
			if tag == "" {
				tag = metadata.Version
			}
		}
		payload[n] = string(data)
	}
	if namespace != "" && !strings.HasSuffix(namespace, "/") {
		namespace += "/"
	}
	imageName := namespace + internal.AppNameFromDir(app.AppName) + internal.AppExtension + ":" + tag
	_, err = resto.PushConfigMulti(context.Background(), payload, imageName, resto.RegistryOptions{Insecure: os.Getenv("DOCKERAPP_INSECURE_REGISTRY") != ""}, nil)
	return err
}
