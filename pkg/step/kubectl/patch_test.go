package kubectl_test

import (
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/solo-io/valet/pkg/api"
	"github.com/solo-io/valet/pkg/cmd"
	mock_cmd "github.com/solo-io/valet/pkg/cmd/mocks"
	"github.com/solo-io/valet/pkg/render"
	mock_render "github.com/solo-io/valet/pkg/render/mocks"
	"github.com/solo-io/valet/pkg/step/kubectl"
)

var _ = Describe("patch", func() {

	const (
		resourceName = "name"
		resourceNs   = "ns"
		patchType    = "patchType"
		kind         = "kind"
		patchFile    = "path"
	)

	var (
		ctrl        *gomock.Controller
		runner      *mock_cmd.MockRunner
		fileStore   *mock_render.MockFileStore
		ctx         *api.WorkflowContext
		expectedCmd = func(name, ns, patchString string) *cmd.Command {
			return cmd.New().Kubectl().With("patch", kind, name, "-n", ns, "--type", patchType, "--patch", patchString).Cmd()
		}
	)

	BeforeEach(func() {
		ctrl = gomock.NewController(T)
		runner = mock_cmd.NewMockRunner(ctrl)
		fileStore = mock_render.NewMockFileStore(ctrl)
		ctx = &api.WorkflowContext{
			Runner:    runner,
			FileStore: fileStore,
		}
	})

	AfterEach(func() {
		ctrl.Finish()
	})

	It("runs", func() {
		path := "foo"
		patch := kubectl.Patch{
			Name: resourceName,
			Namespace: resourceNs,
			Path: path,
			PatchType: patchType,
			KubeType: kind,
		}
		expected := expectedCmd(resourceName, resourceNs, "blah")
		fileStore.EXPECT().Load(path).Return("blah", nil).Times(1)
		runner.EXPECT().Run(expected).Return(nil).Times(1)
		err := patch.Run(ctx, nil)
		Expect(err).To(BeNil())
	})

	It("runs with rendered templates", func() {
		path := "foo"
		patch := kubectl.Patch{
			Name: "{{ .Name }}",
			Namespace: "{{ .Namespace }}",
			Path: path,
			PatchType: patchType,
			KubeType: kind,
		}
		values := render.Values{
			"Name": resourceName,
			"Namespace": resourceNs,
		}
		expected := expectedCmd(resourceName, resourceNs, "blah")
		fileStore.EXPECT().Load(path).Return("blah", nil).Times(1)
		runner.EXPECT().Run(expected).Return(nil).Times(1)
		err := patch.Run(ctx, values)
		Expect(err).To(BeNil())
	})

	It("gets the right description", func() {
		path := "foo"
		patch := kubectl.Patch{
			Name: "{{ .Name }}",
			Namespace: "{{ .Namespace }}",
			Path: path,
			PatchType: patchType,
			KubeType: kind,
		}
		values := render.Values{
			"Name": resourceName,
			"Namespace": resourceNs,
		}
		expected := expectedCmd(resourceName, resourceNs, "blah")
		fileStore.EXPECT().Load(path).Return("blah", nil).Times(1)
		desc, err := patch.GetDescription(ctx, values)
		Expect(err).To(BeNil())
		Expect(desc).To(Equal("Running command: " + expected.ToString()))
	})
})
