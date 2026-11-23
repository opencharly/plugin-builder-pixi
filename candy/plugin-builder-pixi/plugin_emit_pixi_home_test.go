package builderpixi

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/opencharly/sdk/proto/pb"
	"github.com/opencharly/spec/spec"
)

// TestResolveServesThePixiHomePins is THIS module's emitted-artifact check: it invokes the
// provider exactly as the host does (Op "resolve") and inspects the STAGE the reply carries —
// not the sdk helper behind it — asserting the environment pins that keep the pixi env inside
// the box's home and the guard that fails the build when it is not there.
//
// It fails if the dependency is reverted to a release whose template lacks these lines, which
// is the regression that reddened every consumer check reading <home>/.pixi
// (opencharly/pod-comfyui#51).
func TestResolveServesThePixiHomePins(t *testing.T) {
	for _, home := range []string{"/home/user", "/root"} {
		in := spec.BuilderResolveInput{
			Candy: "c", Builder: "pixi", BuilderRef: "builder:v1", StageName: "c-pixi-build",
			CopySrc: "candy/c", UID: 1000, GID: 1000, Home: home, User: "u", Manifest: "t",
		}
		params, _ := json.Marshal(in)
		reply, err := provider{}.Invoke(context.Background(), &pb.InvokeRequest{Op: "resolve", ParamsJson: params})
		if err != nil {
			t.Fatalf("Invoke(resolve): %v", err)
		}
		var served struct {
			Stage string `json:"stage"`
		}
		if err := json.Unmarshal(reply.GetResultJson(), &served); err != nil {
			t.Fatalf("decode reply: %v (%s)", err, reply.GetResultJson())
		}
		if !strings.Contains(served.Stage, "ENV HOME="+home+" PIXI_HOME="+home+"/.pixi") {
			t.Fatalf("served stage for home %q lacks the HOME/PIXI_HOME pins:\n%s", home, served.Stage)
		}
		if !strings.Contains(served.Stage, "test -d "+home+"/.pixi") {
			t.Fatalf("served stage for home %q lacks the env-root guard:\n%s", home, served.Stage)
		}
	}
}
