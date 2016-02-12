//   Copyright 2016 Wercker Holding BV
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package dockerlocal

import (
	"fmt"

	"github.com/wercker/wercker/core"
	"github.com/wercker/wercker/util"
)

// DockerDeploy is our basic wrapper for DockerDeploy operations
type DockerDeploy struct {
	*DockerPipeline
}

// ToDeploy grabs the build section from the config and configures all the
// instances necessary for the build
func NewDockerDeploy(name string, config *core.Config, options *core.PipelineOptions, dockerOptions *DockerOptions, builder Builder) (*DockerDeploy, error) {
	base, err := NewDockerPipeline(name, config, options, dockerOptions, builder)
	if err != nil {
		return nil, err
	}
	return &DockerDeploy{base}, nil
}

// InitEnv sets up the internal state of the environment for the build
func (d *DockerDeploy) InitEnv(hostEnv *util.Environment) {
	env := d.Env()

	a := [][]string{
		[]string{"DEPLOY", "true"},
		[]string{"WERCKER_DEPLOY_ID", d.options.DeployID},
		[]string{"WERCKER_DEPLOY_URL", fmt.Sprintf("%s/#deploy/%s", d.options.BaseURL, d.options.DeployID)},
		[]string{"WERCKER_GIT_DOMAIN", d.options.GitDomain},
		[]string{"WERCKER_GIT_OWNER", d.options.GitOwner},
		[]string{"WERCKER_GIT_REPOSITORY", d.options.GitRepository},
		[]string{"WERCKER_GIT_BRANCH", d.options.GitBranch},
		[]string{"WERCKER_GIT_COMMIT", d.options.GitCommit},
	}

	if d.options.DeployTarget != "" {
		a = append(a, []string{"WERCKER_DEPLOYTARGET_NAME", d.options.DeployTarget})
	}

	env.Update(d.CommonEnv())
	env.Update(a)
	env.Update(hostEnv.GetMirror())
	env.Update(hostEnv.GetPassthru().Ordered())
	env.Hidden.Update(hostEnv.GetHiddenPassthru().Ordered())
}

// DockerRepo returns the name where we might store this in docker
func (d *DockerDeploy) DockerRepo() string {
	if d.options.Repository != "" {
		return d.options.Repository
	}
	return fmt.Sprintf("%s/%s", d.options.ApplicationOwnerName, d.options.ApplicationName)
}

// DockerTag returns the tag where we might store this in docker
func (d *DockerDeploy) DockerTag() string {
	tag := d.options.Tag
	if tag == "" {
		tag = fmt.Sprintf("deploy-%s", d.options.DeployID)
	}
	return tag
}

// DockerMessage returns the message to store this with in docker
func (d *DockerDeploy) DockerMessage() string {
	message := d.options.Message
	if message == "" {
		message = fmt.Sprintf("Build %s", d.options.DeployID)
	}
	return message
}

// CollectArtifact copies the artifacts associated with the Deploy.
// Unlike a Build, this will only collect the output directory if we made
// a new one.
func (d *DockerDeploy) CollectArtifact(containerID string) (*core.Artifact, error) {
	artificer := NewArtificer(d.options, d.dockerOptions)

	artifact := &core.Artifact{
		ContainerID:   containerID,
		GuestPath:     d.options.GuestPath("output"),
		HostPath:      d.options.HostPath("output"),
		HostTarPath:   d.options.HostPath("output.tar"),
		ApplicationID: d.options.ApplicationID,
		DeployID:      d.options.DeployID,
		Bucket:        d.options.S3Bucket,
		ContentType:   "application/x-tar",
	}

	// Get the output dir, if it is empty grab the source dir.
	fullArtifact, err := artificer.Collect(artifact)
	if err != nil {
		return nil, err
	}

	return fullArtifact, nil
}
