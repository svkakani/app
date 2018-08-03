package dockerapp

import (
	"archive/tar"
	"context"
	"fmt"
	"io"
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

// FromSingleFileContent splits a single-file docker application content
func FromSingleFileContent(input []byte) (DockerAppRaw, error) {
	in := string(input)
	parts := strings.Split(in, "\n---")
	if len(parts) != 3 {
		return DockerAppRaw{}, fmt.Errorf("malformed single-file application: expected 3 documents")
	}
	res := DockerAppRaw{}
	res.Metadata = []byte(parts[0])
	split := strings.SplitN(parts[1], "\n", 2)
	if len(split) > 1 {
		res.Compose = []byte(split[1])
	}
	split = strings.SplitN(parts[2], "\n", 2)
	if len(split) > 1 {
		res.Settings = []byte(split[2])
	}
	return res, nil
}

// FromSingleFile builds a DockerApp from a single-file path
func FromSingleFile(input string) (DockerApp, error) {
	data, err := ioutil.ReadFile(input)
	if err != nil {
		return DockerApp{}, errors.Wrap(err, "failed to read single-file application package")
	}
	raw, err := FromSingleFileContent(data)
	if err != nil {
		return DockerApp{}, err
	}
	return DockerApp {
		AppName: internal.AppNameFromDir(input),
		Origin:  input,
		Raw:     &raw,
	}, nil
}

// FromDirectory builds a DockerApp from a directory path
func FromDirectory(dir string) (DockerApp, error) {
	meta, err := ioutil.ReadFile(filepath.Join(dir, internal.MetadataFileName))
	if err != nil {
		return DockerApp{}, err
	}
	compose, err := ioutil.ReadFile(filepath.Join(dir, internal.ComposeFileName))
	if err != nil {
		return DockerApp{}, err
	}
	settings, err := ioutil.ReadFile(filepath.Join(dir, internal.SettingsFileName))
	if err != nil {
		return DockerApp{}, err
	}
	return DockerApp {
		AppName: internal.AppNameFromDir(dir),
		Origin:  dir,
		Raw : &DockerAppRaw {
			Metadata: meta,
			Compose:  compose,
			Settings: settings,
		},
	}, nil
}

// FromImage loads a docker app from an image name
func FromImage(repotag string) (DockerApp, error) {
	payload, err := resto.PullConfigMulti(context.Background(), repotag, resto.RegistryOptions{Insecure: os.Getenv("DOCKERAPP_INSECURE_REGISTRY") != ""})
	if err != nil {
		return DockerApp{}, err
	}
	repoComps := strings.Split(repotag, ":")
	repo := repoComps[0]
	// handle the case where a port was specified in the domain part of repotag
	if len(repoComps) == 3 || (len(repoComps) == 2 && strings.Contains(repoComps[1], "/")) {
		repo = repoComps[1]
	}
	return DockerApp {
		AppName: internal.AppNameFromDir(repo),
		Origin:  repotag,
		Raw:     &DockerAppRaw {
			Metadata: []byte(payload[internal.MetadataFileName]),
			Compose:  []byte(payload[internal.ComposeFileName]),
			Settings: []byte(payload[internal.SettingsFileName]),
		},
	}, nil
}

// FromTar loads a DockerApp from a tarball
func FromTar(path string) (DockerApp, error) {
	f, err := os.Open(path)
	if err != nil {
		return DockerApp{}, errors.Wrap(err, "failed to open application package")
	}
	defer f.Close()
	tarReader := tar.NewReader(f)
	raw := DockerAppRaw{}
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return DockerApp{}, errors.Wrap(err, "error reading from tar header")
		}
		if header.Typeflag == tar.TypeReg {
			data := make([]byte, header.Size)
			_, err := tarReader.Read(data)
			if err != nil && err != io.EOF {
				return DockerApp{}, errors.Wrap(err, "error reading from tar data")
			}
			switch header.Name {
			case internal.MetadataFileName:
				raw.Metadata = data
			case internal.ComposeFileName:
				raw.Compose = data
			case internal.SettingsFileName:
				raw.Settings = data
			}
		}
	}
	return DockerApp {
		AppName: internal.AppNameFromDir(path),
		Origin:  path,
		Raw: &raw,
	}, nil
}

// findApp looks for an app in CWD or subdirs
func findApp() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "cannot resolve current working directory")
	}
	if strings.HasSuffix(cwd, internal.AppExtension) {
		return cwd, nil
	}
	content, err := ioutil.ReadDir(cwd)
	if err != nil {
		return "", errors.Wrap(err, "failed to read current working directory")
	}
	hit := ""
	for _, c := range content {
		if strings.HasSuffix(c.Name(), internal.AppExtension) {
			if hit != "" {
				return "", fmt.Errorf("multiple applications found in current directory, specify the application name on the command line")
			}
			hit = c.Name()
		}
	}
	if hit == "" {
		return "", fmt.Errorf("no application found in current directory")
	}
	return filepath.Join(cwd, hit), nil
}

// Load loads a docker-app from any of the supported sources
func Load(input string) (DockerApp, error) {
	input = internal.DirNameFromAppName(input)
	if input == "" {
		var err error
		if input, err = findApp(); err != nil {
			return DockerApp{}, err
		}
	}
	if input == "." {
		var err error
		if input, err = os.Getwd(); err != nil {
			return DockerApp{}, errors.Wrap(err, "cannot resolve current working directory")
		}
	}
	s, err := os.Stat(input)
	if err != nil {
		res, err := FromImage(input)
		if err != nil {
			return DockerApp{}, errors.Wrap(err, "input is not a directory, file, or image")
		}
		return res, nil
	}
	if s.IsDir() {
		return FromDirectory(input)
	}
	res, err := FromTar(input)
	if err != nil {
		res, err = FromSingleFile(input)
		if err != nil {
			return DockerApp{}, errors.Wrap(err, "input is neither a single-file nor a tarball")
		}
	}
	return res, err
}

// ToSingleFile saves an application as a single file
func ToSingleFile(app DockerApp, target io.Writer) error {
	if _, err := target.Write(app.Raw.Metadata); err != nil {
		return err
	}
	if _, err := io.WriteString(target, "\n---\n"); err != nil {
		return err
	}
	if _, err := target.Write(app.Raw.Compose); err != nil {
		return err
	}
	if _, err := io.WriteString(target, "\n---\n"); err != nil {
		return err
	}
	if _, err := target.Write(app.Raw.Settings); err != nil {
		return err
	}
	return nil
}

// ToDirectory saves an application as a directory
func ToDirectory(app DockerApp, path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(path, internal.MetadataFileName), app.Raw.Metadata, 0644); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(path, internal.ComposeFileName), app.Raw.Compose, 0644); err != nil {
		return err
	}
	if err := ioutil.WriteFile(filepath.Join(path, internal.SettingsFileName), app.Raw.Settings, 0644); err != nil {
		return err
	}
	return nil
}

// ToImage saves an application to the given repotag
func ToImage(app DockerApp, repotag string) error {
	_, err := resto.PushConfigMulti(context.Background(),
		map[string] string {
			internal.MetadataFileName: string(app.Raw.Metadata),
			internal.ComposeFileName:  string(app.Raw.Compose),
			internal.SettingsFileName: string(app.Raw.Settings),
		},
		repotag,
		resto.RegistryOptions{Insecure: os.Getenv("DOCKERAPP_INSECURE_REGISTRY") != ""},
		nil)
	return err
}

// ToImageWithDefaults saves an aplication to a registry using
func ToImageWithDefaults(app DockerApp, namespace, tag string) error {
	if namespace == "" || tag == "" {
		var meta types.AppMetadata
		err := yaml.Unmarshal(app.Raw.Metadata, &meta)
		if err != nil {
			return err
		}
		if namespace == "" {
			namespace = meta.Namespace
		}
		if tag == "" {
			tag = meta.Version
		}
	}
	repotag := namespace + "/" + app.AppName + internal.AppExtension + ":" + tag
	return ToImage(app, repotag)
}

// ToTar saves an application as a tarball
func ToTar(app DockerApp, path string) error {
	target, err := os.Create(path)
	if err != nil {
		return err
	}
	defer target.Close()
	tarw := tar.NewWriter(target)
	if err := tarw.WriteHeader(&tar.Header{
			Name: internal.MetadataFileName,
			Size: int64(len(app.Raw.Metadata)),
			Mode: 0644,
			Typeflag: tar.TypeReg,
	}) ; err != nil {
	  return err
	}
	if _, err := tarw.Write(app.Raw.Metadata); err != nil {
		return err
	}
	if err := tarw.WriteHeader(&tar.Header{
			Name: internal.ComposeFileName,
			Size: int64(len(app.Raw.Compose)),
			Mode: 0644,
			Typeflag: tar.TypeReg,
	}) ; err != nil {
	  return err
	}
	if _, err := tarw.Write(app.Raw.Compose); err != nil {
		return err
	}
	if err := tarw.WriteHeader(&tar.Header{
			Name: internal.SettingsFileName,
			Size: int64(len(app.Raw.Settings)),
			Mode: 0644,
			Typeflag: tar.TypeReg,
	}) ; err != nil {
	  return err
	}
	if _, err := tarw.Write(app.Raw.Settings); err != nil {
		return err
	}
	return tarw.Close()
}