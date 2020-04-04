package kubectl

import (
	"github.com/solo-io/go-utils/installutils/helmchart"
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/cmd"
	"github.com/solo-io/valet/pkg/render"
	"io/ioutil"
	"os"
	"strings"

	"github.com/solo-io/go-utils/osutils"

	"github.com/solo-io/go-utils/installutils/kuberesource"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	errors "github.com/rotisserie/eris"
)

const (
	encryptedSuffix = ".enc"
)

var (
	_ api.Step = new(CreateSecret)

	InvalidCiphertextFilenameError = errors.Errorf("Ciphertext files must end with '%s'.", encryptedSuffix)
	UnableToDecryptFileError       = func(err error) error {
		return errors.Wrapf(err, "Unable to decrypt file.")
	}
	MissingEnvVarError = func(envVar string) error {
		return errors.Errorf("Missing environment variable %s", envVar)
	}
)

type CreateSecret struct {
	// Currently, secrets cannot consist of values from multiple registries
	Name         string                 `json:"name,omitempty"`
	Namespace    string                 `json:"namespace,omitempty" valet:"key=Namespace"`
	Type         string                 `json:"typ,omitemptye" valet:"default=Opaque"`
	Entries      map[string]SecretValue `json:"entries,omitempty"`
}

type SecretValue struct {
	EnvVar                 string                  `json:"envVar,omitempty"`
	File                   string                  `json:"file,omitempty"`
	GcloudKmsEncryptedFile *GcloudKmsEncryptedFile `json:"gcloudKmsEncryptedFile,omitempty"`
}

type GcloudKmsEncryptedFile struct {
	CiphertextFile string `json:"ciphertextFile,omitempty"`
	GcloudProject  string `json:"gcloudProject,omitempty"`
	Keyring        string `json:"keyring,omitempty"`
	Key            string `json:"key,omitempty"`
}

func (s *CreateSecret) Run(ctx *api.WorkflowContext, values render.Values) error {
	if err := values.RenderFields(s, ctx.Runner); err != nil {
		return err

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
		k, err := render.LoadTemplate(k, values, ctx.Runner)
		if err != nil {
			return err
		}
		if v.File != "" {
			contents, err := ctx.FileStore.Load(v.File)
			if err != nil {
				return err
			}
			secret.Data[k] = []byte(contents)
		} else if v.EnvVar != "" {
			val := os.Getenv(v.EnvVar)
			if val == "" {
				return MissingEnvVarError(v.EnvVar)
			}
			secret.Data[k] = []byte(val)
		} else if v.GcloudKmsEncryptedFile != nil {
			if !strings.HasSuffix(v.GcloudKmsEncryptedFile.CiphertextFile, encryptedSuffix) {
				return InvalidCiphertextFilenameError
			}
			encContents, err := ctx.FileStore.Load(v.GcloudKmsEncryptedFile.CiphertextFile)
			if err != nil {
				return err
			}
			encrypted, err := ioutil.TempFile("", "valet-test-secret-enc-")
			if err != nil {
				return err
			}
			defer cleanupFile(encrypted.Name())
			if err := ioutil.WriteFile(encrypted.Name(), []byte(encContents), os.ModePerm); err != nil {
				return err
			}
			unencrypted, err := ioutil.TempFile("", "valet-test-secret-")
			if err != nil {
				return err
			}
			defer cleanupFile(unencrypted.Name())
			command := cmd.New().Gcloud().DecryptFile(
				encrypted.Name(),
				unencrypted.Name(),
				v.GcloudKmsEncryptedFile.GcloudProject,
				v.GcloudKmsEncryptedFile.Keyring,
				v.GcloudKmsEncryptedFile.Key).Cmd()
			err = ctx.Runner.Run(command)
			if err != nil {
				return UnableToDecryptFileError(err)
			}
			osClient := osutils.NewOsClient()
			contents, err := osClient.ReadFile(unencrypted.Name())
			if err != nil {
				return err
			}
			secret.Data[k] = contents
		}
	}
	resource, err := kuberesource.ConvertToUnstructured(&secret)
	if err != nil {
		return err
	}
	manifests, err := helmchart.ManifestsFromResources(kuberesource.UnstructuredResources{resource})
	if err != nil {
		return err
	}
	kubectlCmd := cmd.New().Kubectl().ApplyStdIn(manifests.CombinedString()).Cmd()
	return ctx.Runner.Run(kubectlCmd)
}

func cleanupFile(name string) {
	if err := os.Remove(name); err != nil {
		cmd.Stderr().Println("Error cleaning up file %s: %s", name, err.Error())
	}
}

func (s *CreateSecret) GetDescription(ctx *api.WorkflowContext, values render.Values) (string, error) {
	return "Creating secret", nil
}

func (s *CreateSecret) GetDocs(ctx *api.WorkflowContext, options api.DocsOptions) (string, error) {
	panic("implement me")
}
