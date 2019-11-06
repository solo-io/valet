package resource

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
)

const (
	secret          = "secret"
	generic         = "generic"
	encryptedSuffix = ".enc"
)

var (
	_ Resource = new(Secret)

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
	Namespace string                 `yaml:"namespace"`
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

func (s *Secret) updateWithValues(values Values) error {
	if s.Namespace == "" {
		if values.ContainsKey(NamespaceKey) {
			if val, err := values.GetValue(NamespaceKey); err != nil {
				return err
			} else {
				s.Namespace = val
			}
		}
	}
	return nil
}

func (s *Secret) Ensure(ctx context.Context, command cmd.Factory) error {
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

func (s *Secret) Teardown(ctx context.Context, command cmd.Factory) error {
	cmd.Stdout().Println("Tearing down secret %s.%s", s.Namespace, s.Name)
	return command.Kubectl().Delete(secret).Namespace(s.Namespace).WithName(s.Name).IgnoreNotFound().Cmd().Run(ctx)
}
