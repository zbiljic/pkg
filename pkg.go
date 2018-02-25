/*
Package pkg exists solely to aid consumers of the pkg library when using
dependency managers.
*/
package pkg

import (
	_ "github.com/zbiljic/pkg/logger"
	_ "github.com/zbiljic/pkg/metrics"
)
