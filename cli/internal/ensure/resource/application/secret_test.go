package application_test

import (
	"context"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/installutils/kuberesource"
	"github.com/solo-io/valet/cli/internal/ensure/resource/application"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
	v1 "k8s.io/api/core/v1"
)

var _ = Describe("Secret", func() {

	const (
		name              = "test-secret"
		namespace         = "test-namespace"
		secretValue       = "secret"
		secretValueBase64 = "c2VjcmV0"
		secretType        = "test"

		secretEnvVarEntry = "EnvVar"
		secretEnvVar      = "TEST_SECRET_ENV_VAR"

		secretFileEntry = "File"
		secretFilePath  = "test/files/secrets/secret.txt"

		secretGcloudFileEntry        = "GcloudFile"
		secretGcloudFilePath         = "test/files/secrets/secret.txt.enc"
		secretGcloudFileRegistryPath = "files/secrets/secret.txt.enc"

		registryName       = "test-registry"
		registryPath       = "test"
		secretRegistryPath = "files/secrets/secret.txt"
	)

	var (
		ctx          = context.TODO()
		emptyInput   = render.InputParams{}
		testRegistry = render.DirectoryRegistry{
			WorkingDirectory: registryPath,
		}
	)

	getUnstructured := func(key string) map[string]interface{} {
		return map[string]interface{}{
			key: secretValueBase64,
		}
	}

	getSecret := func(secretEntry string, secretValue application.SecretValue) application.Secret {
		return application.Secret{
			Name:      name,
			Namespace: namespace,
			Type:      secretType,
			Entries: map[string]application.SecretValue{
				secretEntry: secretValue,
			},
		}
	}

	expectSecret := func(resources kuberesource.UnstructuredResources, key string) {
		Expect(len(resources)).To(Equal(1))
		actual := resources[0]
		Expect(actual.GetName()).To(Equal(name))
		Expect(actual.GetNamespace()).To(Equal(namespace))
		Expect(actual.UnstructuredContent()["type"]).To(Equal(secretType))
		Expect(actual.UnstructuredContent()["data"]).To(Equal(getUnstructured(key)))
	}

	Context("render", func() {
		It("handles env var secret values", func() {
			err := os.Setenv(secretEnvVar, secretValue)
			Expect(err).To(BeNil())
			value := application.SecretValue{
				EnvVar: secretEnvVar,
			}
			secret := getSecret(secretEnvVarEntry, value)
			resources, err := secret.Render(ctx, emptyInput)
			Expect(err).To(BeNil())
			expectSecret(resources, secretEnvVarEntry)
		})

		It("returns err for missing env var secret value", func() {
			err := os.Setenv(secretEnvVar, "")
			Expect(err).To(BeNil())
			value := application.SecretValue{
				EnvVar: secretEnvVar,
			}
			secret := getSecret(secretEnvVarEntry, value)
			_, err = secret.Render(ctx, emptyInput)
			Expect(err).NotTo(BeNil())
		})

		It("handles file secret values with default registry", func() {
			value := application.SecretValue{
				File: secretFilePath,
			}
			secret := getSecret(secretFileEntry, value)
			resources, err := secret.Render(ctx, emptyInput)
			Expect(err).To(BeNil())
			expectSecret(resources, secretFileEntry)
		})

		It("returns error on bad path for secret value file", func() {
			value := application.SecretValue{
				File: "path/to/my/fake/secret.txt",
			}
			secret := getSecret(secretFileEntry, value)
			_, err := secret.Render(ctx, emptyInput)
			Expect(err).NotTo(BeNil())
		})

		It("handles file secret values with test registry", func() {
			value := application.SecretValue{
				File: secretRegistryPath,
			}
			secret := getSecret(secretFileEntry, value)
			secret.RegistryName = registryName
			input := render.InputParams{}
			input.SetRegistry(registryName, &testRegistry)
			resources, err := secret.Render(ctx, input)
			Expect(err).To(BeNil())
			expectSecret(resources, secretFileEntry)
		})

		It("handles gcloud secret values with default registry", func() {
			value := application.SecretValue{
				GcloudKmsEncryptedFile: &application.GcloudKmsEncryptedFile{
					GcloudProject:  "solo-public",
					Key:            "build-key",
					Keyring:        "build",
					CiphertextFile: secretGcloudFilePath,
				},
			}
			secret := getSecret(secretGcloudFileEntry, value)
			resources, err := secret.Render(ctx, emptyInput)
			Expect(err).To(BeNil())
			expectSecret(resources, secretGcloudFileEntry)
		})

		It("returns error for bad gcloud encrypted path", func() {
			value := application.SecretValue{
				GcloudKmsEncryptedFile: &application.GcloudKmsEncryptedFile{
					CiphertextFile: "path/to/my/fake/secret.txt.enc",
				},
			}
			secret := getSecret(secretGcloudFileEntry, value)
			_, err := secret.Render(ctx, emptyInput)
			Expect(err).NotTo(BeNil())
		})

		It("handles gcloud secret values with test registry", func() {
			value := application.SecretValue{
				GcloudKmsEncryptedFile: &application.GcloudKmsEncryptedFile{
					GcloudProject:  "solo-public",
					Key:            "build-key",
					Keyring:        "build",
					CiphertextFile: secretGcloudFileRegistryPath,
				},
			}
			secret := getSecret(secretGcloudFileEntry, value)
			secret.RegistryName = registryName
			input := render.InputParams{}
			input.SetRegistry(registryName, &testRegistry)
			resources, err := secret.Render(ctx, input)
			Expect(err).To(BeNil())
			expectSecret(resources, secretGcloudFileEntry)
		})

		It("uses a default secret type of Opaque", func() {
			err := os.Setenv(secretEnvVar, secretValue)
			Expect(err).To(BeNil())
			value := application.SecretValue{
				EnvVar: secretEnvVar,
			}
			secret := getSecret(secretEnvVarEntry, value)
			secret.Type = ""
			resources, err := secret.Render(ctx, emptyInput)
			Expect(err).To(BeNil())
			Expect(len(resources)).To(Equal(1))
			actual := resources[0]
			Expect(actual.UnstructuredContent()["type"]).To(Equal(string(v1.SecretTypeOpaque)))
		})

		It("uses namespace from values", func() {
			err := os.Setenv(secretEnvVar, secretValue)
			Expect(err).To(BeNil())
			value := application.SecretValue{
				EnvVar: secretEnvVar,
			}
			secret := getSecret(secretEnvVarEntry, value)
			secret.Namespace = ""
			input := render.InputParams{
				Values: render.Values{
					render.NamespaceKey: namespace,
				},
			}
			resources, err := secret.Render(ctx, input)
			Expect(err).To(BeNil())
			expectSecret(resources, secretEnvVarEntry)
		})
	})
})
