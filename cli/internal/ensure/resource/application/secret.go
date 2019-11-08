package application

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/valet/cli/internal/ensure/resource"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

const (
	secret          = "secret"
	generic         = "generic"
	encryptedSuffix = ".enc"
)

var (
	_ resource.Resource = new(Secret)
	_ Renderable        = new(Secret)

	InvalidCiphertextFilenameError = errors.Errorf("Ciphertext files must end with '%s'.", encryptedSuffix)
	UnableToDecryptFileError       = func(err error) error {
		return errors.Wrapf(err, "Unable to decrypt file.")
	}
	UnableToCleanupPlaintextFileError = func(err error) error {
		return errors.Wrapf(err, "Unable to cleanup plaintext file.")
	}
)

type Secret struct {
	Name      string                 `yaml:"name"`
	Namespace string                 `yaml:"namespace" valet:"key=Namespace"`
	Entries   map[string]SecretValue `yaml:"entries"`
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

func (s *Secret) Ensure(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	if err := input.Values.RenderFields(s); err != nil {
		return err
	}
	cmd.Stdout().Println("Ensuring secret %s.%s with %d entries", s.Namespace, s.Name, len(s.Entries))
	toRun := command.Kubectl().Create(secret).With(generic).WithName(s.Name).Namespace(s.Namespace)
	var toCleanup []string
	for name, v := range s.Entries {
		if v.File != "" {
			fromFile := fmt.Sprintf("--from-file=%s=%s", name, v.File)
			toRun = toRun.With(fromFile)
		} else if v.EnvVar != "" {
			template := "--from-literal=%s=%s"
			fromLiteral := fmt.Sprintf(template, name, os.Getenv(v.EnvVar))
			fromLiteralRedacted := fmt.Sprintf(template, name, cmd.Redacted)
			toRun = toRun.With(fromLiteral).Redact(fromLiteral, fromLiteralRedacted)
		} else if v.GcloudKmsEncryptedFile != nil {
			if !strings.HasSuffix(v.GcloudKmsEncryptedFile.CiphertextFile, encryptedSuffix) {
				return InvalidCiphertextFilenameError
			}
			unencrypted := strings.TrimSuffix(v.GcloudKmsEncryptedFile.CiphertextFile, encryptedSuffix)
			err := command.Gcloud().DecryptFile(
				v.GcloudKmsEncryptedFile.CiphertextFile,
				unencrypted,
				v.GcloudKmsEncryptedFile.GcloudProject,
				v.GcloudKmsEncryptedFile.Keyring,
				v.GcloudKmsEncryptedFile.Key).Cmd().Run(ctx)
			if err != nil {
				return UnableToDecryptFileError(err)
			}
			toCleanup = append(toCleanup, unencrypted)
			fromFile := fmt.Sprintf("--from-file=%s=%s", name, unencrypted)
			toRun = toRun.With(fromFile)
		}
	}
	if err := toRun.DryRunAndApply(ctx, command); err != nil {
		return err
	}
	for _, fileToCleanup := range toCleanup {
		if err := os.Remove(fileToCleanup); err != nil {
			return UnableToCleanupPlaintextFileError(err)
		}
	}
	return nil
}

func (s *Secret) Teardown(ctx context.Context, input render.InputParams, command cmd.Factory) error {
	if err := input.Values.RenderFields(s); err != nil {
		return err
	}
	cmd.Stdout().Println("Tearing down secret %s.%s", s.Namespace, s.Name)
	return command.Kubectl().Delete(secret).Namespace(s.Namespace).WithName(s.Name).IgnoreNotFound().Cmd().Run(ctx)
}

func (s *Secret) Render(ctx context.Context, input render.InputParams, command cmd.Factory) (kuberesource.UnstructuredResources, error) {
	secret := v1.Secret{
		TypeMeta: v12.TypeMeta{
			Kind:       "Secret",
			APIVersion: "v1",
		},
		Type: v1.SecretTypeOpaque,
		ObjectMeta: v12.ObjectMeta{
			Name:      s.Name,
			Namespace: s.Namespace,
		},
		Data: make(map[string][]byte),
	}
	var toCleanup []string
	for k, v := range s.Entries {
		if v.File != "" {
			contents, err := render.LoadBytes(v.File)
			if err != nil {
				return nil, err
			}
			secret.Data[k] = contents
		} else if v.EnvVar != "" {
			val := os.Getenv(v.EnvVar)
			if val == "" {
				return nil, errors.Errorf("Missing environment variable %s", v.EnvVar)
			}
			secret.Data[k] = []byte(val)
		} else if v.GcloudKmsEncryptedFile != nil {
			if !strings.HasSuffix(v.GcloudKmsEncryptedFile.CiphertextFile, encryptedSuffix) {
				return nil, InvalidCiphertextFilenameError
			}
			unencrypted := strings.TrimSuffix(v.GcloudKmsEncryptedFile.CiphertextFile, encryptedSuffix)
			err := command.Gcloud().DecryptFile(
				v.GcloudKmsEncryptedFile.CiphertextFile,
				unencrypted,
				v.GcloudKmsEncryptedFile.GcloudProject,
				v.GcloudKmsEncryptedFile.Keyring,
				v.GcloudKmsEncryptedFile.Key).Cmd().Run(ctx)
			if err != nil {
				return nil, UnableToDecryptFileError(err)
			}
			toCleanup = append(toCleanup, unencrypted)
			contents, err := render.LoadBytes(v.File)
			if err != nil {
				return nil, err
			}
			secret.Data[k] = contents
		}
	}
	for _, fileToCleanup := range toCleanup {
		if err := os.Remove(fileToCleanup); err != nil {
			return nil, UnableToCleanupPlaintextFileError(err)
		}
	}
	resource, err := kuberesource.ConvertToUnstructured(&secret)
	if err != nil {
		return nil, err
	}
	return kuberesource.UnstructuredResources{resource}, nil
}
