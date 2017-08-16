package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/opalmer/dockertest"
	"github.com/opalmer/gerrittest"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

func getSSHKeys(cmd *cobra.Command) (ssh.PublicKey, ssh.Signer, string, error) {
	privateKeyPath, err := cmd.Flags().GetString("private-key")
	if err != nil {
		return nil, nil, "", err
	}

	// No private key given, generate one instead.
	if privateKeyPath == "" {
		private, err := gerrittest.GenerateRSAKey()
		if err != nil {
			return nil, nil, "", err
		}

		file, err := ioutil.TempFile("", "id_rsa-")
		if err != nil {
			return nil, nil, "", err
		}
		defer file.Close()
		if err := gerrittest.WriteRSAKey(private, file); err != nil {
			return nil, nil, "", err
		}
		signer, err := ssh.NewSignerFromKey(private)
		if err != nil {
			return nil, nil, "", err
		}

		return signer.PublicKey(), signer, file.Name(), err
	}

	public, private, err := gerrittest.ReadSSHKeys(privateKeyPath)
	return public, private, privateKeyPath, err
}

// NewConfigFromCommand converts a command to a config struct.
func NewConfigFromCommand(cmd *cobra.Command) (*gerrittest.Config, error) {
	image, err := cmd.Flags().GetString("image")
	if err != nil {
		return nil, err
	}

	portHTTP, err := cmd.Flags().GetUint16("port-http")
	if err != nil {
		return nil, err
	}

	portSSH, err := cmd.Flags().GetUint16("port-ssh")
	if err != nil {
		return nil, err
	}

	noCleanup, err := cmd.Flags().GetBool("no-cleanup")
	if err != nil {
		return nil, nil
	}

	return &gerrittest.Config{
		Image:            image,
		PortSSH:          portSSH,
		PortHTTP:         portHTTP,
		CleanupOnFailure: noCleanup == false,
	}, nil
}

func jsonOutput(cmd *cobra.Command, spec *gerrittest.ServiceSpec) error {
	data, err := json.MarshalIndent(spec, "", " ")
	if err != nil {
		return err
	}

	jsonPath, err := cmd.Flags().GetString("json")
	if err != nil {
		return err
	}
	if jsonPath != "" {
		if err := os.MkdirAll(filepath.Dir(jsonPath), 0700); err != nil {
			return err
		}
		return ioutil.WriteFile(jsonPath, data, 0600)
	}
	fmt.Println(string(data))
	return nil
}

// Start implements the `start` subcommand.
var Start = &cobra.Command{
	Use:   "start",
	Short: "Responsible for starting an instance of Gerrit.",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := NewConfigFromCommand(cmd)
		if err != nil {
			return err
		}
		// Setup timeout and Ctrl+C handling.
		timeout, err := cmd.Flags().GetDuration("timeout")
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		interrupts := make(chan os.Signal, 1)
		signal.Notify(interrupts, os.Interrupt)
		go func() {
			for range interrupts {
				cancel()
			}
		}()

		public, _, privateKeyPath, err := getSSHKeys(cmd)
		if err != nil {
			return err
		}
		if public == nil {
			return errors.New("internal error, failed to retrieve public key")
		}

		spec := &gerrittest.ServiceSpec{Admin: &gerrittest.User{}}
		service, err := gerrittest.Start(ctx, cfg)
		if err != nil {
			return err
		}

		spec.HTTP = service.HTTPPort
		spec.SSH = service.SSHPort
		spec.Container = service.Container.ID()

		startonly, err := cmd.Flags().GetBool("start-only")
		if startonly {
			return nil
		}

		client := service.HTTPClient()
		spec.URL = client.Prefix

		// Hitting /login/ will produce a cookie that can be used
		// for authenticated requests. Also, this first request
		// causes the first account to be created which happens
		// to be the admin account.
		if err := client.Login(); err != nil {
			return err
		}

		account, err := client.GetAccount()
		if err != nil {
			return err
		}
		spec.Admin.Login = account.Username

		// Password setup
		passwd, err := cmd.Flags().GetString("password")
		if err != nil {
			return err
		}
		if passwd == "" {
			password, err := client.GeneratePassword()
			if err != nil {
				return err
			}
			spec.Admin.Password = password

		} else {
			if err := client.SetPassword(passwd); err != nil {
				return err
			}
			spec.Admin.Password = passwd
		}

		// Insert ssh key and test
		if err := client.InsertPublicKey(public); err != nil {
			return err
		}
		spec.Admin.PrivateKey = privateKeyPath
		spec.SSHCommand = fmt.Sprintf(
			"ssh -p %d -i %s -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no "+
				"%s@%s", spec.SSH.Public, spec.Admin.PrivateKey, spec.Admin.Login, spec.SSH.Address)

		// Setup the SSH client and make sure we're able to connect.
		sshClient, err := gerrittest.NewSSHClient(
			spec.Admin.Login, spec.Admin.PrivateKey, spec.SSH)
		if err != nil {
			return err
		}
		if _, err := sshClient.Version(); err != nil {
			return err
		}

		return jsonOutput(cmd, spec)
	},
}

func init() {
	Start.Flags().Duration(
		"timeout", time.Minute*2,
		"The maximum amount of time to wait for the service to come up.")
	Start.Flags().BoolP(
		"no-cleanup", "n", false,
		"If provided then do not cleanup the container on failure. "+
			"Useful when debugging changes to the docker image.")
	Start.Flags().String(
		"json", "",
		"The location to write information about the service to. Any "+
			"existing content will be overwritten.")
	Start.Flags().String(
		"image", gerrittest.DefaultImage,
		"The Docker image to spin up Gerrit.")
	Start.Flags().Uint16(
		"port-http", dockertest.RandomPort,
		"The local port to map to Gerrit's REST API. Random by default.")
	Start.Flags().Uint16(
		"port-ssh", dockertest.RandomPort,
		"The local port to map to Gerrit's REST API. Random by default.")
	Start.Flags().StringP(
		"private-key", "i", "",
		"If provided then use this private key instead of generating one.")
	Start.Flags().Bool(
		"start-only", false,
		"If provided just start the container, don't setup anything else.")
	Start.Flags().StringP(
		"password", "p", "",
		"If provided then use this value for the admin password instead "+
			"of generating one.")
}
