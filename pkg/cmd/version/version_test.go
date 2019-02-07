/*
Copyright 2019 the Heptio Ark contributors.

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

package version

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	velerov1 "github.com/heptio/velero/pkg/apis/velero/v1"
	"github.com/heptio/velero/pkg/buildinfo"
	"github.com/heptio/velero/pkg/generated/clientset/versioned/fake"
	v1 "github.com/heptio/velero/pkg/generated/clientset/versioned/typed/velero/v1"
	"github.com/heptio/velero/pkg/serverstatusrequest"
)

func TestPrintVersion(t *testing.T) {
	// set up some non-empty buildinfo values, but put them back to their
	// defaults at the end of the test
	var (
		origVersion      = buildinfo.Version
		origGitSHA       = buildinfo.GitSHA
		origGitTreeState = buildinfo.GitTreeState
	)
	defer func() {
		buildinfo.Version = origVersion
		buildinfo.GitSHA = origGitSHA
		buildinfo.GitTreeState = origGitTreeState
	}()
	buildinfo.Version = "v1.0.0"
	buildinfo.GitSHA = "somegitsha"
	buildinfo.GitTreeState = "dirty"

	clientVersion := fmt.Sprintf("Client:\n\tVersion: %s\n\tGit commit: %s\n", buildinfo.Version, buildinfo.FormattedGitSHA())

	tests := []struct {
		name                string
		clientOnly          bool
		serverStatusRequest *velerov1.ServerStatusRequest
		getterError         error
		want                string
	}{
		{
			name:       "client-only",
			clientOnly: true,
			want:       clientVersion,
		},
		{
			name:                "server status getter error",
			clientOnly:          false,
			serverStatusRequest: nil,
			getterError:         errors.New("an error"),
			want:                clientVersion + "<error getting server version: an error>\n",
		},
		{
			name:                "server status getter returns normally",
			clientOnly:          false,
			serverStatusRequest: serverstatusrequest.NewBuilder().ServerVersion("v1.0.1").Build(),
			getterError:         nil,
			want:                clientVersion + "Server:\n\tVersion: v1.0.1\n",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var (
				serverStatusGetter = new(mockServerStatusGetter)
				buf                = new(bytes.Buffer)
				client             = fake.NewSimpleClientset()
			)
			defer serverStatusGetter.AssertExpectations(t)

			// getServerStatus should only be called when clientOnly = false
			if !tc.clientOnly {
				serverStatusGetter.On("getServerStatus", client.VeleroV1()).Return(tc.serverStatusRequest, tc.getterError)
			}

			printVersion(buf, tc.clientOnly, client.VeleroV1(), serverStatusGetter)

			assert.Equal(t, tc.want, buf.String())
		})
	}
}

// serverStatusGetter is an autogenerated mock type for the serverStatusGetter type
type mockServerStatusGetter struct {
	mock.Mock
}

// getServerStatus provides a mock function with given fields: client
func (_m *mockServerStatusGetter) getServerStatus(client v1.ServerStatusRequestsGetter) (*velerov1.ServerStatusRequest, error) {
	ret := _m.Called(client)

	var r0 *velerov1.ServerStatusRequest
	if rf, ok := ret.Get(0).(func(v1.ServerStatusRequestsGetter) *velerov1.ServerStatusRequest); ok {
		r0 = rf(client)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*velerov1.ServerStatusRequest)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(v1.ServerStatusRequestsGetter) error); ok {
		r1 = rf(client)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}