package oci

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/containers/image/copy"
	"github.com/containers/image/signature"
	"github.com/containers/image/transports"
	"github.com/containers/image/types"
)

func PullImage(imageDir, image string) ([]byte, error) {

	if err := os.MkdirAll(imageDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", imageDir, err)
	}
	ctx := context.Background()
	policyContext, err := getPolicyContext()
	if err != nil {
		return nil, fmt.Errorf("failed to get policy context: %w", err)
	}

	srcRef, err := ParseImageName(image, "docker")
	if err != nil {
		return nil, fmt.Errorf("invalid image name %s: %w", image, err)
	}

	destRef, err := ParseImageName(image, "oci")
	if err != nil {
		return nil, fmt.Errorf("failed to set destination name %s: %w", image, err)
	}
	return copy.Image(ctx, policyContext, destRef, srcRef, &copy.Options{})
}

func ParseImageName(imgName, transporttype string) (types.ImageReference, error) {

	transport := transports.Get(transporttype)
	if transport == nil {
		return nil, fmt.Errorf("failed to get image transport type")
	}

	if transporttype == "docker" && !strings.HasPrefix(imgName, "//") {
		imgName = fmt.Sprintf("//%s", imgName)
	}

	ref, err := transport.ParseReference(imgName)

	if err != nil {
		return nil, fmt.Errorf("failed to parse image name %s: %w", imgName, err)
	}

	return ref, nil
}

func getPolicyContext() (*signature.PolicyContext, error) {
	policy := &signature.Policy{
		Default: signature.PolicyRequirements{signature.NewPRInsecureAcceptAnything()},
	}

	return signature.NewPolicyContext(policy)

}
