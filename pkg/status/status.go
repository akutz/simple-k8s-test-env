/*
simple-kubernetes-test-environment

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package status

import (
	"context"
	"io"
	"os"

	"github.com/sirupsen/logrus"

	"sigs.k8s.io/kind/pkg/log"
)

var statusContextKey struct{}

// Context returns a new Context with a status writer for stdout.
func Context() context.Context {
	status := New(os.Stdout)
	status.MaybeWrapLogrus(logrus.StandardLogger())
	return WithStatus(context.Background(), status)
}

// WithStatus returns a new Context with a status writer
func WithStatus(parent context.Context, status *log.Status) context.Context {
	return context.WithValue(parent, statusContextKey, status)
}

// New creates a new Status object.
func New(w io.Writer) *log.Status {
	return log.NewStatus(w)
}

// Start starts a status using the Status object in the provided
// context. Otherwise no status is started.
func Start(ctx context.Context, status string) {
	obj, ok := ctx.Value(statusContextKey).(*log.Status)
	if !ok {
		return
	}
	obj.Start(status)
}

// End ends a status using the Status object in the provided
// context. Otherwise no status is ended.
func End(ctx context.Context, success bool) {
	obj, ok := ctx.Value(statusContextKey).(*log.Status)
	if !ok {
		return
	}
	obj.End(success)
}
