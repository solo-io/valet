package gen

import (
	"context"
	"fmt"
	"github.com/solo-io/valet/cli/internal/ensure/resource"
	"io"
	"strings"
)

func GenerateDoc(ctx context.Context, ref resource.WorkflowRef) (string, error) {
	output := strings.Builder{}
	output.WriteString(fmt.Sprintf("# %s\n\n", ref.Path))

}