package workflow_test

import (
	"context"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/valet/cli/internal/ensure/cmd"
	mock_cmd "github.com/solo-io/valet/cli/internal/ensure/cmd/mocks"
	"github.com/solo-io/valet/cli/internal/ensure/resource/render"
	mock_render "github.com/solo-io/valet/cli/internal/ensure/resource/render/mocks"
	"github.com/solo-io/valet/cli/internal/ensure/resource/workflow"
	"k8s.io/client-go/tools/clientcmd"
)

var _ = Describe("patch", func() {
	const (
		name           = "test-name"
		namespace      = "test-namespace"
		path           = "test-path"
		registryName   = "test-registry"
		kubeType       = "test-kube-type"
		patchType      = "test-patch-type"
		patchContents  = "test-patch"
		otherNamespace = "other-test-namespace"
	)

	var (
		ctrl        *gomock.Controller
		runner      *mock_cmd.MockRunner
		registry    *mock_render.MockRegistry
		input       render.InputParams
		emptyErr    = errors.Errorf("")
		ctx         = context.TODO()
		expectedCmd = cmd.New().Kubectl().With("patch", kubeType, name, "-n", namespace,
			"--type", patchType, "--patch", patchContents).Cmd(clientcmd.RecommendedHomeFile)
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(T)
		runner = mock_cmd.NewMockRunner(ctrl)
		input = render.InputParams{
			CommandRunner: runner,
		}
		registry = mock_render.NewMockRegistry(ctrl)
		input.SetRegistry(registryName, registry)
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	Context("fully provided patch", func() {
		patch := &workflow.Patch{
			Name:         name,
			Namespace:    namespace,
			Path:         path,
			RegistryName: registryName,
			KubeType:     kubeType,
			PatchType:    patchType,
		}

		It("works for patch", func() {
			registry.EXPECT().LoadFile(ctx, path).Return(patchContents, nil).Times(1)
			runner.EXPECT().Run(ctx, expectedCmd).Return(nil).Times(1)
			err := patch.Ensure(ctx, input)
			Expect(err).To(BeNil())
		})

		It("returns error when patch file can't be loaded", func() {
			registry.EXPECT().LoadFile(ctx, path).Return("", emptyErr).Times(1)
			err := patch.Ensure(ctx, input)
			Expect(err).To(Equal(workflow.UnableToLoadPatchError(emptyErr)))
		})

	})

	Context("patch template rendering", func() {
		patch := &workflow.Patch{
			Name:         "{{ .Name }}",
			Namespace:    "{{ .Namespace }}",
			Path:         path,
			RegistryName: registryName,
			KubeType:     kubeType,
			PatchType:    patchType,
		}

		values := render.Values{
			render.NamespaceKey: namespace,
			render.NameKey:      name,
		}

		It("works for patch when values in input", func() {
			input.Values = values
			registry.EXPECT().LoadFile(ctx, path).Return(patchContents, nil).Times(1)
			runner.EXPECT().Run(ctx, expectedCmd).Return(nil).Times(1)
			err := patch.Ensure(ctx, input)
			Expect(err).To(BeNil())
		})

		It("works for patch when values on step", func() {
			patch.Values = values
			registry.EXPECT().LoadFile(ctx, path).Return(patchContents, nil).Times(1)
			runner.EXPECT().Run(ctx, expectedCmd).Return(nil).Times(1)
			err := patch.Ensure(ctx, input)
			Expect(err).To(BeNil())
		})

		It("works for patch when values on step and input, prefering input", func() {
			input.Values = values
			patch.Values = render.Values{
				render.NamespaceKey: otherNamespace,
			}
			registry.EXPECT().LoadFile(ctx, path).Return(patchContents, nil).Times(1)
			runner.EXPECT().Run(ctx, expectedCmd).Return(nil).Times(1)
			err := patch.Ensure(ctx, input)
			Expect(err).To(BeNil())
		})
	})

	Context("patch default rendering", func() {

		It("works", func() {
			patch := workflow.Patch{}
			err := input.RenderFields(&patch)
			Expect(err).To(BeNil())
			Expect(patch.RegistryName).To(Equal(render.DefaultRegistry))
		})
	})

})
