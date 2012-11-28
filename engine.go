package tritium

import (
	"time"
	tp "tritium/proto"
)

type Engine interface {
	Run(transform *tp.Transform, rrules []*tp.RewriteRule, input interface{}, vars map[string]string, deadline time.Time) (output string, exports [][]string, logs []string)
	TransformRequest(transforms []*tp.Transform, rrules []*tp.RewriteRule, input interface{}, vars map[string]string, deadline time.Time) (output string, exports [][]string, logs []string)
	TransformResponse(transforms []*tp.Transform, rrules []*tp.RewriteRule, input interface{}, vars map[string]string, deadline time.Time) (output string, exports [][]string, logs []string)
	Free()
}
