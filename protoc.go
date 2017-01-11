package blueprint

import (
	"fmt"
	"path/filepath"
)

func compileProtos(protoDir, protoFile, projectDir string) error {
	gosrc := "/go/src"

	calls := [][]string{
		{ // Create grpc stub
			"-I.",
			"-I" + protoDir,
			"-I" + gosrc,
			"-I" + filepath.Join(gosrc, "github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis"),
			"--go_out=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:" + projectDir,
		},
		{ // Create grpc-gateway code
			"-I.",
			"-I" + protoDir,
			"-I" + gosrc,
			"-I" + filepath.Join(gosrc, "github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis"),
			"--grpc-gateway_out=logtostderr=true:" + projectDir,
		},
		{ // Create grpc service struct
			"-I.",
			"-I" + protoDir,
			"-I" + gosrc,
			"-I" + filepath.Join(gosrc, "github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis"),
			"--grpc-goservice_out=" + filepath.Join(projectDir, "cmd", "grpcd"),
		},
	}

	for _, args := range calls {
		if err := sysExec("protoc", append(args, protoFile)...); err != nil {
			return fmt.Errorf("protoc failed: %s", err)
		}
	}

	return nil
}
