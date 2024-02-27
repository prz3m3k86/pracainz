package identity

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/kfsoftware/hlf-operator/api/hlf.kungfusoftware.es/v1alpha1"
	"github.com/kfsoftware/hlf-operator/kubectl-hlf/cmd/helpers"
	"github.com/spf13/cobra"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type createIdentityCmd struct {
	name           string
	namespace      string
	caName         string
	caNamespace    string
	ca             string
	mspID          string
	enrollId       string
	enrollSecret   string
	caEnrollId     string
	caEnrollSecret string
	caType         string
}

func (c *createIdentityCmd) validate() error {
	if c.name == "" {
		return fmt.Errorf("--name is required")
	}
	if c.namespace == "" {
		return fmt.Errorf("--namespace is required")
	}
	if c.mspID == "" {
		return fmt.Errorf("--mspid is required")
	}
	if c.ca == "" {
		return fmt.Errorf("--ca is required")
	}
	if c.caName == "" {
		return fmt.Errorf("--ca-name is required")
	}
	if c.caNamespace == "" {
		return fmt.Errorf("--ca-namespace is required")
	}
	if c.enrollId == "" {
		return fmt.Errorf("--enroll-id is required")
	}
	if c.enrollSecret == "" {
		return fmt.Errorf("--enroll-secret is required")
	}
	return nil
}
func (c *createIdentityCmd) run() error {
	oclient, err := helpers.GetKubeOperatorClient()
	if err != nil {
		return err
	}
	ctx := context.Background()
	clientSet, err := helpers.GetKubeClient()
	if err != nil {
		return err
	}
	fabricCA, err := helpers.GetCertAuthByName(
		clientSet,
		oclient,
		c.caName,
		c.caNamespace,
	)
	if err != nil {
		return err
	}
	fabricIdentitySpec := v1alpha1.FabricIdentitySpec{
		Caname: c.ca,
		Cahost: fabricCA.Name,
		Caport: 7054,
		Catls: v1alpha1.Catls{
			Cacert: base64.StdEncoding.EncodeToString([]byte(fabricCA.Status.TlsCert)),
		},
		Enrollid:     c.enrollId,
		Enrollsecret: c.enrollSecret,
		MSPID:        c.mspID,
	}
	if c.caEnrollId != "" && c.caEnrollSecret != "" {
		fabricIdentitySpec.Register = &v1alpha1.FabricIdentityRegister{
			Enrollid:       c.caEnrollId,
			Enrollsecret:   c.caEnrollSecret,
			Type:           c.caType,
			Affiliation:    "",
			MaxEnrollments: -1,
			Attrs:          []string{},
		}
	}
	fabricIdentity := &v1alpha1.FabricIdentity{
		ObjectMeta: v1.ObjectMeta{
			Name:      c.name,
			Namespace: c.namespace,
		},
		Spec: fabricIdentitySpec,
	}
	fabricIdentity, err = oclient.HlfV1alpha1().FabricIdentities(c.namespace).Create(
		ctx,
		fabricIdentity,
		v1.CreateOptions{},
	)
	if err != nil {
		return err
	}
	fmt.Printf("Created hlf identity %s/%s\n", fabricIdentity.Name, fabricIdentity.Namespace)
	return nil
}
func newIdentityCreateCMD() *cobra.Command {
	c := &createIdentityCmd{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create HLF identity",
		Long:  `Create HLF identity`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := c.validate(); err != nil {
				return err
			}
			if err := c.run(); err != nil {
				return err
			}
			return nil
		},
	}
	f := cmd.Flags()
	f.StringVar(&c.name, "name", "", "Name of the external chaincode")
	f.StringVar(&c.namespace, "namespace", "", "Namespace of the external chaincode")
	f.StringVar(&c.caName, "ca-name", "", "Name of the CA")
	f.StringVar(&c.caNamespace, "ca-namespace", "", "Namespace of the CA")
	f.StringVar(&c.ca, "ca", "", "CA name")
	f.StringVar(&c.mspID, "mspid", "", "MSP ID")
	f.StringVar(&c.enrollId, "enroll-id", "", "Enroll ID")
	f.StringVar(&c.enrollSecret, "enroll-secret", "", "Enroll Secret")
	f.StringVar(&c.caEnrollId, "ca-enroll-id", "", "CA Enroll ID to register the user")
	f.StringVar(&c.caEnrollSecret, "ca-enroll-secret", "", "CA Enroll Secret to register the user")
	f.StringVar(&c.caType, "ca-type", "", "Type of the user to be registered in the CA")
	return cmd
}
