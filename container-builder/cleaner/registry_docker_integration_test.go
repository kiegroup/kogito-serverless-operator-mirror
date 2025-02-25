//go:build integration_docker

/*
 * Copyright 2022 Red Hat, Inc. and/or its affiliates.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cleaner

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"k8s.io/klog/v2"

	"github.com/kiegroup/kogito-serverless-operator/container-builder/common"
	"github.com/kiegroup/kogito-serverless-operator/container-builder/util/log"
)

func TestRegistryDockerIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(DockerTestSuite))
}

func (suite *DockerTestSuite) TestDockerRegistry() {
	klog.V(log.I).InfoS("TestPullTagPush ")
	assert.Truef(suite.T(), suite.RegistryID != "", "Registry not started")
	assert.Truef(suite.T(), suite.LocalRegistry.IsRegistryImagePresent(), "Registry image not present")
	assert.Truef(suite.T(), suite.LocalRegistry.GetRegistryRunningID() == suite.RegistryID, "Registry container not running")
	assert.True(suite.T(), suite.LocalRegistry.Connection.DaemonHost() == "unix:///var/run/docker.sock")
}

func (suite *DockerTestSuite) TestPullTagPush() {
	assert.Truef(suite.T(), suite.RegistryID != "", "Registry not started")
	registryContainer, err := common.GetRegistryContainer()
	assert.Nil(suite.T(), err)
	reposInitial, _ := registryContainer.GetRepositories()
	initialRepoSize := len(reposInitial)
	repos := CheckRepositoriesSize(suite.T(), initialRepoSize, registryContainer)

	result := dockerPullTagPushOnRegistryContainer(suite)
	assert.True(suite.T(), result)

	time.Sleep(2 * time.Second) // Needed on CI
	repos = CheckRepositoriesSize(suite.T(), initialRepoSize+1, registryContainer)
	klog.V(log.I).InfoS("Repo Size after pull image", "size", len(repos))
}

func dockerPullTagPushOnRegistryContainer(suite *DockerTestSuite) bool {
	dockerSocketConn := suite.Docker.Connection
	d := common.Docker{Connection: dockerSocketConn}

	err := d.PullImage(common.TEST_IMG_SECOND)
	time.Sleep(2 * time.Second) // needed on CI
	if err != nil {
		assert.Fail(suite.T(), "Pull Image Failed", err)
		return false
	}

	err = d.TagImage(common.TEST_IMG_SECOND_TAG, common.TEST_IMG_SECOND_LOCAL_TAG)
	if err != nil {
		assert.Fail(suite.T(), "Tag Image Failed", err)
		return false
	}

	err = d.PushImage(common.TEST_IMG_SECOND_LOCAL_TAG, common.REGISTRY_CONTAINER_URL_FROM_DOCKER_SOCKET, "", "")
	if err != nil {
		assert.Fail(suite.T(), "Push Image Failed", err)
		return false
	}
	return true
}
