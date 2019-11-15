package application

import (
	"context"
	"io/ioutil"
	"os"
	"strings"

	"github.com/solo-io/go-utils/osutils"

	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

const (
	encryptedSuffix = ".enc"
)

var (
	_ Renderable = new(Secret)

	InvalidCiphertextFilenameError = errors.Errorf("Ciphertext files must end with '%s'.", encryptedSuffix)
	UnableToDecryptFileError       = func(err error) error {
		return errors.Wrapf(err, "Unable to decrypt file.")
	}
	MissingEnvVarError = func(envVar string) error {
		return errors.Errorf("Missing environment variable %s", envVar)
	}
)

type Secret struct {
	// Currently, secrets cannot consist of values from multiple registries
	RegistryName string                 `yaml:"registry" valet:"default=default"`
	Name         string                 `yaml:"name"`
	Namespace    string                 `yaml:"namespace" valet:"key=Namespace"`
	Type         string                 `yaml:"type" valet:"default=Opaque"`
	Entries      map[string]SecretValue `yaml:"entries"`
}

type SecretValue struct {
	EnvVar                 string                  `yaml:"envVar"`
	File                   string                  `yaml:"file"`
	GcloudKmsEncryptedFile *GcloudKmsEncryptedFile `yaml:"gcloudKmsEncryptedFile"`
}

type GcloudKmsEncryptedFile struct {
	CiphertextFile string `yaml:"ciphertextFile"`
	GcloudProject  string `yaml:"gcloudProject"`
	Keyring        string `yaml:"keyring"`
	Key            string `yaml:"key"`
}

func (s *Secret) Render(ctx context.Context, input render.InputParams) (kuberesource.UnstructuredResources, error) {
	if err := input.RenderFields(s); err != nil {
		return nil, err

	}
	cmd.Stdout().Println("Rendering secret %s.%s with type %s and %d entries", s.Namespace, s.Name, s.Type, len(s.Entries))
	secret := v1.Secret{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		Type: v1.SecretType(s.Type),
		ObjectMeta: metav1.ObjectMeta{
			Name:      s.Name,
			Namespace: s.Namespace,
		},
		Data: make(map[string][]byte),
	}
	for k, v := range s.Entries {
		k, err := render.LoadTemplate(k, input.Values, input.Runner())
		if err != nil {
			return nil, err
		}
		if v.File != "" {
			contents, err := input.LoadFile(s.RegistryName, v.File)
			if err != nil {
				return nil, err
			}
			secret.Data[k] = []byte(contents)
		} else if v.EnvVar != "" {
			val := os.Getenv(v.EnvVar)
			if val == "" {
				return nil, MissingEnvVarError(v.EnvVar)
			}
			secret.Data[k] = []byte(val)
		} else if v.GcloudKmsEncryptedFile != nil {
			if !strings.HasSuffix(v.GcloudKmsEncryptedFile.CiphertextFile, encryptedSuffix) {
				return nil, InvalidCiphertextFilenameError
			}
			encContents, err := input.LoadFile(s.RegistryName, v.GcloudKmsEncryptedFile.CiphertextFile)
			if err != nil {
				return nil, err
			}
			encrypted, err := ioutil.TempFile("", "valet-test-secret-enc-")
			if err != nil {
				return nil, err
			}
			defer cleanupFile(encrypted.Name())
			if err := ioutil.WriteFile(encrypted.Name(), []byte(encContents), os.ModePerm); err != nil {
				return nil, err
			}
			unencrypted, err := ioutil.TempFile("", "valet-test-secret-")
			if err != nil {
				return nil, err
			}
			defer cleanupFile(unencrypted.Name())
			command := cmd.New().Gcloud().DecryptFile(
				encrypted.Name(),
				unencrypted.Name(),
				v.GcloudKmsEncryptedFile.GcloudProject,
				v.GcloudKmsEncryptedFile.Keyring,
				v.GcloudKmsEncryptedFile.Key).Cmd()
			err = input.Runner().Run(ctx, command)
			if err != nil {
				return nil, UnableToDecryptFileError(err)
			}
			osClient := osutils.NewOsClient()
			contents, err := osClient.ReadFile(unencrypted.Name())
			if err != nil {
				return nil, err
			}
			secret.Data[k] = contents
		}
	}
	resource, err := kuberesource.ConvertToUnstructured(&secret)
	if err != nil {
		return nil, err
	}
	return kuberesource.UnstructuredResources{resource}, nil
}

func cleanupFile(name string) {
	if err := os.Remove(name); err != nil {
		cmd.Stderr().Println("Error cleaning up file %s: %s", name, err.Error())
	}
}
