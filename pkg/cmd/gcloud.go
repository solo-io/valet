package cmd

import (
	"fmt"
)

const (
	Decrypt = "decrypt"
)

type Gcloud struct {
	cmd *Command
}

func (g *Gcloud) Cmd() *Command {
	return g.cmd
}

func (g *Gcloud) With(args ...string) *Gcloud {
	g.cmd = g.cmd.With(args...)
	return g
}

func (g *Gcloud) GetCredentials() *Gcloud {
	return g.With("container", "clusters", "get-credentials")
}

func (g *Gcloud) Project(project string) *Gcloud {
	return g.With(fmt.Sprintf("--project=%s", project))
}

func (g *Gcloud) Zone(zone string) *Gcloud {
	return g.With(fmt.Sprintf("--zone=%s", zone))
}

func (g *Gcloud) WithName(name string) *Gcloud {
	return g.With(name)
}

func (g *Gcloud) Kms(operation string) *Gcloud {
	return g.With("kms", operation)
}

func (g *Gcloud) Ciphertext(cipherText string) *Gcloud {
	return g.With(fmt.Sprintf("--ciphertext-file=%s", cipherText))
}

func (g *Gcloud) Plaintext(plainText string) *Gcloud {
	return g.With(fmt.Sprintf("--plaintext-file=%s", plainText))
}

func (g *Gcloud) Keyring(keyring string) *Gcloud {
	return g.With(fmt.Sprintf("--keyring=%s", keyring))
}

func (g *Gcloud) Key(key string) *Gcloud {
	return g.With(fmt.Sprintf("--key=%s", key))
}

func (g *Gcloud) Global() *Gcloud {
	return g.With("--location=global")
}

func (g *Gcloud) DecryptFile(cipherText, plainText, project, keyring, key string) *Gcloud {
	return g.Kms(Decrypt).Ciphertext(cipherText).Plaintext(plainText).Project(project).Keyring(keyring).Key(key).Global()
}
