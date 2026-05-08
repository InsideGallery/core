package fastlog

import (
	_ "github.com/InsideGallery/core/fastlog/handlers/stderr" // register stderr event-stream handler
	_ "github.com/InsideGallery/core/fastlog/handlers/stdout" // register stdout event-stream handler
)
