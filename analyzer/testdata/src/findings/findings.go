// want +1 "custom build tag \"minimal\""
//go:build !minimal

package findings

import (
	_ "embed"
	htmltemplate "html/template" // want `import "html/template" adds runtime parsing`
	_ "image/png"                // want `side-effect import "image/png"`
	"net"                        // want `import "net" makes resolver behavior`
	"os/user"                    // want `import "os/user" makes user lookup behavior`
	"reflect"                    // want `import "reflect" may retain extra`
	texttemplate "text/template" // want `import "text/template" adds runtime parsing`
	_ "time/tzdata"              // want `import "time/tzdata" embeds the timezone database`
)

// want +2 `go:embed pattern "asset.txt"`
//
//go:embed asset.txt
var asset string

var (
	_ net.IP
	_ user.User
	_ reflect.Type
	_ *htmltemplate.Template
	_ *texttemplate.Template
)
